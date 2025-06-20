package process

import (
	"context"
	"errors"
	"fmt"
	"github.com/coder/quartz"
	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"
	"github.com/sz-po/go-distributed-kvm-switch/internal/pkg/api/utils"
	"go.openly.dev/pointy"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
)

type LocalProcessOpt func(*LocalProcess)

type LocalProcess struct {
	name          Name
	specification Specification

	statusMutex *sync.Mutex
	status      Status

	instanceMutex *sync.Mutex
	instance      *exec.Cmd

	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer

	controlLoopStop     context.CancelFunc
	controlLoopInterval time.Duration

	logger *slog.Logger
	clock  quartz.Clock
}

func WithLogger(logger *slog.Logger) LocalProcessOpt {
	return func(p *LocalProcess) { p.logger = logger }
}

func WithClock(clock quartz.Clock) LocalProcessOpt {
	return func(p *LocalProcess) { p.clock = clock }
}

func WithStdin(stdin io.Reader) LocalProcessOpt {
	return func(p *LocalProcess) {
		p.stdin = stdin
	}
}

func WithStdout(stdout io.Writer) LocalProcessOpt {
	return func(p *LocalProcess) {
		p.stdout = stdout
	}
}

func WithStderr(stderr io.Writer) LocalProcessOpt {
	return func(p *LocalProcess) {
		p.stderr = stderr
	}
}

func CreateLocalProcessFactory() ProcessFactory {
	return func(name Name, specification Specification, opts ...ProcessOpt) (Process, error) {
		var localProcessOpts []LocalProcessOpt

		for _, opt := range opts {
			if localOpt, ok := opt.(LocalProcessOpt); ok {
				localProcessOpts = append(localProcessOpts, localOpt)
			} else {
				return nil, fmt.Errorf("%w: %T", ErrInvalidLocalProcessOpt, opt)
			}
		}

		return NewLocalProcess(name, specification, localProcessOpts...)
	}
}

func NewLocalProcess(name Name, specification Specification, opts ...LocalProcessOpt) (*LocalProcess, error) {
	if err := name.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidProcessName, err)
	}

	if err := defaults.Set(&specification); err != nil {
		return nil, fmt.Errorf("failed to set defaults: %w", err)
	}

	validator := validator.New(validator.WithRequiredStructEnabled())

	if err := validator.Struct(specification); err != nil {
		return nil, fmt.Errorf("failed to validate specification: %w", err)
	}

	process := &LocalProcess{
		name:          name,
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

		controlLoopInterval: 100 * time.Millisecond,

		logger: slog.Default(),
		clock:  quartz.NewReal(),
	}

	for _, opt := range opts {
		opt(process)
	}

	process.logger = process.logger.With(
		slog.String("componentName", "LocalProcess"),
		slog.String("processName", string(process.name)),
	)

	ctx := context.Background()

	controlLoopCtx, controlLoopStop := context.WithCancel(ctx)
	process.controlLoopStop = controlLoopStop
	go process.controlLoop(controlLoopCtx)

	return process, nil
}

func (process *LocalProcess) GetStatus() Status {
	process.statusMutex.Lock()
	defer process.statusMutex.Unlock()
	return process.status
}

func (process *LocalProcess) GetSpecification() Specification { return process.specification }

func (process *LocalProcess) Enable(ctx context.Context) error {
	process.statusMutex.Lock()
	defer process.statusMutex.Unlock()

	if process.status.Enabled {
		return ErrProcessAlreadyEnabled
	}
	process.status.Enabled = true
	process.logger.Info("Local process enabled.")
	return nil
}

func (process *LocalProcess) Disable(ctx context.Context) error {
	process.statusMutex.Lock()
	defer process.statusMutex.Unlock()

	if !process.status.Enabled {
		return ErrProcessAlreadyDisabled
	}
	process.status.Enabled = false
	process.logger.Info("Local process disabled.")
	return nil
}

func (process *LocalProcess) Restart(ctx context.Context) error {
	if err := process.Disable(ctx); err != nil {
		return fmt.Errorf("failed to disable process: %w", err)
	}
	if err := process.Wait(ctx, Idle); err != nil {
		return fmt.Errorf("failed to wait for process to stop: %w", err)
	}
	if err := process.Enable(ctx); err != nil {
		return fmt.Errorf("failed to enable process: %w", err)
	}
	return nil
}

func (process *LocalProcess) Wait(ctx context.Context, phase Phase) error {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			process.statusMutex.Lock()
			currentPhase := process.status.Phase
			process.statusMutex.Unlock()
			if currentPhase == phase {
				return nil
			}
		}
	}
}

func (process *LocalProcess) startInstance() error {
	var err error

	process.instanceMutex.Lock()
	defer process.instanceMutex.Unlock()

	process.statusMutex.Lock()
	defer process.statusMutex.Unlock()

	if process.instance != nil {
		return ErrProcessAlreadyRunning
	}

	if process.status.ProcessId != nil {
		process.status.RestartCount++
	}

	// reset status fields from any previous run
	process.status.ProcessId = nil
	process.status.ExitCode = nil
	process.status.ErrorMessage = nil
	process.status.Phase = Idle

	cmd := exec.Command(
		process.specification.ExecutablePath,
		process.specification.Arguments...,
	)

	if process.stdin != nil {
		cmd.Stdin = process.stdin
	}

	if process.stdout != nil {
		cmd.Stdout = process.stdout
	}

	if process.stderr != nil {
		cmd.Stderr = process.stderr
	}

	if cmd.Dir, err = process.buildWorkingDirectory(); err != nil {
		process.status.FailureCount++
		process.status.ErrorMessage = pointy.String(err.Error())
		return fmt.Errorf("failed to build working directory: %w", err)
	}
	process.status.WorkingDirectoryPath = pointy.String(cmd.Dir)

	// environment
	if cmd.Env, process.status.EnvironmentVariables, err = process.buildEnvironmentVariables(); err != nil {
		process.status.FailureCount++
		process.status.ErrorMessage = pointy.String(err.Error())
		return fmt.Errorf("failed to build environment variables: %w", err)
	}

	// actually start
	if err = cmd.Start(); err != nil {
		process.status.FailureCount++
		process.status.ErrorMessage = pointy.String(err.Error())
		return fmt.Errorf("failed to start process: %w", err)
	}

	process.status.ProcessId = pointy.Int(cmd.Process.Pid)
	process.status.Phase = Running
	process.instance = cmd
	process.logger.Info("Local process started.", slog.Int("processId", cmd.Process.Pid))

	go process.watchInstance()
	return nil
}

func (process *LocalProcess) stopInstance(timeout time.Duration) error {
	process.instanceMutex.Lock()
	cmd := process.instance
	process.instanceMutex.Unlock()

	if cmd == nil {
		return ErrProcessNotRunning
	}

	logger := process.logger.With(slog.Int("processId", cmd.Process.Pid))

	process.statusMutex.Lock()
	process.status.Phase = Terminating
	process.statusMutex.Unlock()

	logger.Debug("Sending SIGTERM to the process.")
	if termErr := cmd.Process.Signal(syscall.SIGTERM); termErr != nil {
		logger.Warn("Failed to send SIGTERM.", slog.String("error", termErr.Error()))
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := process.Wait(ctx, Idle); err == nil {
		logger.Info("Local process terminated gracefully.")
		return nil
	}

	process.statusMutex.Lock()
	process.status.Phase = Killing
	process.statusMutex.Unlock()

	logger.Debug("Sending SIGKILL to the process.")
	if killErr := cmd.Process.Kill(); killErr != nil {
		logger.Warn("Failed to send SIGKILL.", slog.String("error", killErr.Error()))
	}

	ctx, cancel = context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := process.Wait(ctx, Idle); err != nil {
		logger.Info("Local process not killed within timeout.")
		return nil
	}

	logger.Info("Local process killed.")
	return nil
}

func (process *LocalProcess) watchInstance() {
	process.instanceMutex.Lock()
	cmd := process.instance
	process.instanceMutex.Unlock()

	logger := process.logger.With(slog.Int("processId", cmd.Process.Pid))
	logger.Debug("Watching process.")

	err := cmd.Wait()

	process.statusMutex.Lock()
	defer process.statusMutex.Unlock()

	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			process.status.ExitCode = pointy.Int(exitErr.ExitCode())
			logger = logger.With(slog.Int("exitCode", exitErr.ExitCode()))
		}
		process.status.ErrorMessage = pointy.String(err.Error())
		process.status.FailureCount++

		if process.status.FailureBackoff <= 0 {
			process.status.FailureBackoff = utils.Duration(time.Second)
		}
		if process.status.FailureBackoff >= utils.Duration(time.Second*30) {
			process.status.FailureBackoff = utils.Duration(time.Second * 30)
		}

		process.status.FailureBackoff *= 2

		logger.Warn("Local process finished with an error.", slog.String("error", err.Error()), slog.Duration("backoff", time.Duration(process.status.FailureBackoff)))

		time.Sleep(time.Duration(process.status.FailureBackoff))
	} else {
		process.status.FailureBackoff = 0
		process.status.ExitCode = pointy.Int(0)
		logger.Info("Local process finished successfully.")
	}

	process.status.Phase = Idle
	process.status.WorkingDirectoryPath = nil

	process.instanceMutex.Lock()
	process.instance = nil
	process.instanceMutex.Unlock()
}

func (process *LocalProcess) buildWorkingDirectory() (string, error) {
	if process.specification.WorkingDirectoryPath != nil {
		return *process.specification.WorkingDirectoryPath, nil
	}
	return os.Getwd()
}

func (process *LocalProcess) buildEnvironmentVariables() ([]string, map[string]string, error) {
	envMap := map[string]string{}

	if process.specification.InheritEnvironmentVariables {
		for _, variable := range os.Environ() {
			k, v, ok := strings.Cut(variable, "=")
			if !ok {
				return nil, nil, fmt.Errorf("failed to parse environment variable: %s", variable)
			}
			envMap[k] = v
		}
	}

	for k, v := range process.specification.EnvironmentVariables {
		envMap[k] = v
	}

	var list []string
	for k, v := range envMap {
		list = append(list, fmt.Sprintf("%s=%s", k, v))
	}
	return list, envMap, nil
}

func (process *LocalProcess) controlLoop(ctx context.Context) {
	process.logger.Debug("Starting the control loop.")

	ticker := process.clock.NewTicker(process.controlLoopInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			process.logger.Debug("The control loop has been stopped.")
			return
		case <-ticker.C:
			if err := process.controlFn(); err != nil {
				process.logger.Warn("Failed to control the process.", slog.String("error", err.Error()))
			}
		}
	}
}

func (process *LocalProcess) controlFn() error {
	process.statusMutex.Lock()
	phase := process.status.Phase
	enabled := process.status.Enabled
	process.statusMutex.Unlock()

	switch {
	case enabled && phase == Idle:
		return process.startInstance()
	case !enabled && phase == Running:
		return process.stopInstance(time.Duration(process.specification.KillTimeout))
	default:
		return nil
	}
}

var ErrInvalidLocalProcessOpt = errors.New("invalid local process opt")
