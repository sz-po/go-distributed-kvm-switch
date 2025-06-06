package process

import (
	"context"
	"errors"
	"fmt"
	"github.com/creasty/defaults"
	"github.com/sz-po/go-distributed-kvm-switch/internal/pkg/api/utils"
	"go.openly.dev/pointy"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
)

type ProcessOpt func(*Process)

type Process struct {
	specification Specification

	statusMutex *sync.Mutex
	status      Status

	instanceMutex *sync.Mutex
	instance      *exec.Cmd

	logger          *slog.Logger
	controlLoopStop context.CancelFunc
	controlLoopCtx  context.Context
}

func WithLogger(logger *slog.Logger) ProcessOpt {
	return func(p *Process) { p.logger = logger }
}

func WithContext(ctx context.Context) ProcessOpt {
	return func(p *Process) {
		controlLoopCtx, controlLoopStop := context.WithCancel(ctx)
		p.controlLoopCtx = controlLoopCtx
		p.controlLoopStop = controlLoopStop

	}
}

func NewProcess(specification Specification, opts ...ProcessOpt) (*Process, error) {
	if err := defaults.Set(&specification); err != nil {
		return nil, fmt.Errorf("failed to set defaults: %w", err)
	}

	p := &Process{
		specification: specification,
		statusMutex:   &sync.Mutex{},
		status: Status{
			Phase:                Idle,
			CreatedAt:            utils.Time(time.Now()),
			Enabled:              specification.AutoEnable,
			EnvironmentVariables: map[string]string{},
		},
		instanceMutex: &sync.Mutex{},
		instance:      nil,

		logger: slog.Default(),
	}

	// apply user‑supplied options
	for _, opt := range opts {
		opt(p)
	}

	p.logger = p.logger.With(slog.String("componentName", "Process"))

	// start control loop in its own cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	p.controlLoopStop = cancel
	go p.controlLoop(ctx)

	return p, nil
}

func (p *Process) GetStatus() Status {
	p.statusMutex.Lock()
	defer p.statusMutex.Unlock()
	return p.status
}

func (p *Process) GetSpecification() Specification { return p.specification }

func (p *Process) Enable(ctx context.Context) error {
	p.statusMutex.Lock()
	defer p.statusMutex.Unlock()

	if p.status.Enabled {
		return ErrProcessAlreadyEnabled
	}
	p.status.Enabled = true
	p.logger.Info("Process enabled.")
	return nil
}

func (p *Process) Disable(ctx context.Context) error {
	p.statusMutex.Lock()
	defer p.statusMutex.Unlock()

	if !p.status.Enabled {
		return ErrProcessAlreadyDisabled
	}
	p.status.Enabled = false
	p.logger.Info("Process disabled.")
	return nil
}

func (p *Process) Restart(ctx context.Context) error {
	if err := p.Disable(ctx); err != nil {
		return fmt.Errorf("failed to disable process: %w", err)
	}
	if err := p.Wait(ctx, Idle); err != nil {
		return fmt.Errorf("failed to wait for process to stop: %w", err)
	}
	if err := p.Enable(ctx); err != nil {
		return fmt.Errorf("failed to enable process: %w", err)
	}
	return nil
}

func (p *Process) Wait(ctx context.Context, phase Phase) error {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			p.statusMutex.Lock()
			currentPhase := p.status.Phase
			p.statusMutex.Unlock()
			if currentPhase == phase {
				return nil
			}
		}
	}
}

// -----------------------------------------------------------------------------
//  Internals with injected factories
// -----------------------------------------------------------------------------

func (p *Process) startInstance() error {
	var err error

	p.instanceMutex.Lock()
	defer p.instanceMutex.Unlock()

	p.statusMutex.Lock()
	defer p.statusMutex.Unlock()

	if p.instance != nil {
		return ErrProcessAlreadyRunning
	}

	if p.status.ProcessId != nil {
		p.status.RestartCount++
	}

	// reset status fields from any previous run
	p.status.ProcessId = nil
	p.status.ExitCode = nil
	p.status.ErrorMessage = nil
	p.status.Phase = Idle

	cmd := exec.Command(
		p.specification.ExecutablePath,
		p.specification.Arguments...,
	)

	if cmd.Dir, err = p.buildWorkingDirectory(); err != nil {
		p.status.FailureCount++
		p.status.ErrorMessage = pointy.String(err.Error())
		return fmt.Errorf("failed to build working directory: %w", err)
	}
	p.status.WorkingDirectoryPath = cmd.Dir

	// environment
	if cmd.Env, p.status.EnvironmentVariables, err = p.buildEnvironmentVariables(); err != nil {
		p.status.FailureCount++
		p.status.ErrorMessage = pointy.String(err.Error())
		return fmt.Errorf("failed to build environment variables: %w", err)
	}

	// actually start
	if err = cmd.Start(); err != nil {
		p.status.FailureCount++
		p.status.ErrorMessage = pointy.String(err.Error())
		return fmt.Errorf("failed to start process: %w", err)
	}

	p.status.ProcessId = pointy.Int(cmd.Process.Pid)
	p.status.Phase = Running
	p.instance = cmd
	p.logger.Info("Process started.", slog.Int("processId", cmd.Process.Pid))

	go p.watchInstance()
	return nil
}

func (p *Process) stopInstance(timeout time.Duration) error {
	p.instanceMutex.Lock()
	cmd := p.instance
	p.instanceMutex.Unlock()

	if cmd == nil {
		return ErrProcessNotRunning
	}

	logger := p.logger.With(slog.Int("processId", cmd.Process.Pid))

	p.statusMutex.Lock()
	p.status.Phase = Terminating
	p.statusMutex.Unlock()

	logger.Debug("Sending SIGTERM to the process.")
	if termErr := cmd.Process.Signal(syscall.SIGTERM); termErr != nil {
		logger.Warn("Failed to send SIGTERM.", slog.String("error", termErr.Error()))
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := p.Wait(ctx, Idle); err == nil {
		logger.Info("Process terminated gracefully.")
		return nil
	}

	p.statusMutex.Lock()
	p.status.Phase = Killing
	p.statusMutex.Unlock()

	logger.Debug("Sending SIGKILL to the process.")
	if killErr := cmd.Process.Kill(); killErr != nil {
		logger.Warn("Failed to send SIGKILL.", slog.String("error", killErr.Error()))
	}

	ctx, cancel = context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := p.Wait(ctx, Idle); err != nil {
		logger.Info("Process not killed within timeout.")
		return nil
	}

	logger.Info("Process killed.")
	return nil
}

func (p *Process) watchInstance() {
	p.instanceMutex.Lock()
	cmd := p.instance
	p.instanceMutex.Unlock()

	logger := p.logger.With(slog.Int("processId", cmd.Process.Pid))
	logger.Debug("Watching process.")

	err := cmd.Wait()

	p.statusMutex.Lock()
	defer p.statusMutex.Unlock()

	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			p.status.ExitCode = pointy.Int(exitErr.ExitCode())
			logger = logger.With(slog.Int("exitCode", exitErr.ExitCode()))
		}
		p.status.ErrorMessage = pointy.String(err.Error())
		p.status.FailureCount++
		logger.Warn("Process finished with an error.", slog.String("error", err.Error()))
	} else {
		p.status.ExitCode = pointy.Int(0)
		logger.Info("Process finished successfully.")
	}

	p.status.Phase = Idle

	p.instanceMutex.Lock()
	p.instance = nil
	p.instanceMutex.Unlock()
}

func (p *Process) buildWorkingDirectory() (string, error) {
	if p.specification.WorkingDirectoryPath != nil {
		stat, err := os.Stat(*p.specification.WorkingDirectoryPath)
		if err != nil {
			return "", fmt.Errorf("%w: %w", ErrInvalidWorkingDirectory, err)
		}
		if !stat.IsDir() {
			return "", fmt.Errorf("%w: %s", ErrWorkingDirectoryIsNotDirectory, *p.specification.WorkingDirectoryPath)
		}
		return *p.specification.WorkingDirectoryPath, nil
	}
	return os.Getwd()
}

func (p *Process) buildEnvironmentVariables() ([]string, map[string]string, error) {
	envMap := map[string]string{}

	if p.specification.InheritEnvironmentVariables {
		for _, variable := range os.Environ() {
			k, v, ok := strings.Cut(variable, "=")
			if !ok {
				return nil, nil, fmt.Errorf("failed to parse environment variable: %s", variable)
			}
			envMap[k] = v
		}
	}

	for k, v := range p.specification.EnvironmentVariables {
		envMap[k] = v
	}

	var list []string
	for k, v := range envMap {
		list = append(list, fmt.Sprintf("%s=%s", k, v))
	}
	return list, envMap, nil
}

func (p *Process) controlLoop(ctx context.Context) {
	p.logger.Debug("Starting the control loop.")

	ticker := time.NewTicker(1 * time.Duration(p.specification.PollInterval))
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			p.logger.Debug("The control loop has been stopped.")
			return
		case <-ticker.C:
			if err := p.controlFn(); err != nil {
				p.logger.Warn("Failed to control the process.", slog.String("error", err.Error()))
			}
		}
	}
}

func (p *Process) controlFn() error {
	p.statusMutex.Lock()
	phase := p.status.Phase
	enabled := p.status.Enabled
	p.statusMutex.Unlock()

	switch {
	case enabled && phase == Idle:
		return p.startInstance()
	case !enabled && phase == Running:
		return p.stopInstance(time.Duration(p.specification.KillTimeout))
	default:
		return nil
	}
}
