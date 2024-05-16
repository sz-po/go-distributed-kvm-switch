package process

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"sync"
	"time"
)

type RunnerOpts func(*Runner)

// Runner is a wrapper around exec.Cmd. It allows to run a command and wait for it to finish. It also provides
// methods to get information about the process.
type Runner struct {
	executablePath string
	executableArgs []string

	isRunning bool
	startedAt time.Time
	exitedAt  time.Time

	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer

	pid      int
	exitCode int

	process *exec.Cmd
	logger  *slog.Logger
	wg      *sync.WaitGroup
}

func NewRunner(executablePath string, opts ...RunnerOpts) *Runner {
	runner := &Runner{
		executablePath: executablePath,
		executableArgs: []string{},

		isRunning: false,
		wg:        &sync.WaitGroup{},

		stdin:  os.Stdin,
		stdout: os.Stdout,
		stderr: os.Stderr,

		logger: slog.With(),
	}

	for _, opt := range opts {
		opt(runner)
	}

	runner.logger = runner.logger.With(
		slog.String("module", "process.runner"),
		slog.String("executablePath", executablePath),
	)

	runner.process = exec.Command(runner.executablePath, runner.executableArgs...)
	runner.process.Stdin = runner.stdin
	runner.process.Stdout = runner.stdout
	runner.process.Stderr = runner.stderr

	return runner
}

func (runner *Runner) Start() error {
	if runner.isRunning {
		return ErrProcessIsRunning
	}

	runner.exitCode = 0
	runner.pid = 0
	runner.exitedAt = time.Time{}
	runner.startedAt = time.Time{}

	runner.logger.Debug("Starting process.")
	err := runner.process.Start()
	if err != nil {
		runner.logger.Warn("Process start error.", slog.String("error", err.Error()))
		return err
	}

	runner.wg.Add(1)

	runner.pid = runner.process.Process.Pid
	runner.isRunning = true
	runner.startedAt = time.Now()

	runner.logger.Info("Process started.", slog.Int("pid", runner.pid))

	go func() {
		runner.logger.Debug("Waiting for process finish.", slog.Int("pid", runner.pid))
		err = runner.process.Wait()
		if err != nil {
			runner.logger.Warn("Process error.", slog.String("error", err.Error()), slog.Int("pid", runner.pid))
		}

		runner.exitedAt = time.Now()
		runner.isRunning = false
		runner.exitCode = runner.process.ProcessState.ExitCode()

		runner.wg.Done()

		runner.logger.Info("Process finished.", slog.Int("pid", runner.pid), slog.Duration("duration", runner.GetUptime()))
	}()

	return nil
}

// Wait for the process to finish. If the process is not running, it returns an error.
func (runner *Runner) Wait() error {
	if !runner.isRunning {
		return ErrProcessIsNotRunning
	}

	runner.wg.Wait()

	return nil
}

func (runner *Runner) Stop(ctx context.Context) error {
	if !runner.isRunning {
		return ErrProcessIsNotRunning
	}

	wait := make(chan struct{})
	go func() {
		runner.wg.Wait()
		wait <- struct{}{}
	}()

	runner.logger.Debug("Stopping process.", slog.Int("pid", runner.pid))
	err := runner.process.Process.Signal(os.Interrupt)
	if err != nil {
		return err
	}

	select {
	case <-wait:
		runner.logger.Info("Process stopped.", slog.Int("pid", runner.pid))
		return nil
	case <-ctx.Done():
	}

	runner.logger.Warn("Stopping process fails. Killing it.", slog.Int("pid", runner.pid))
	err = runner.process.Process.Signal(os.Kill)
	if err != nil {
		return err
	}

	<-wait

	runner.logger.Info("Process killed.", slog.Int("pid", runner.pid))

	return nil
}

// IsRunning returns true if the process is running.
func (runner *Runner) IsRunning() bool {
	return runner.isRunning
}

// GetPID returns the PID of the process. If the process is not running, it returns an error.
func (runner *Runner) GetPID() (int, error) {
	if !runner.isRunning {
		return 0, ErrProcessIsNotRunning
	}

	return runner.pid, nil
}

// GetExitCode returns the exit code of the process. If the process is running, it returns an error.
func (runner *Runner) GetExitCode() (int, error) {
	if runner.isRunning {
		return 0, ErrProcessIsRunning
	}

	return runner.exitCode, nil
}

// GetUptime returns the uptime of the process. If the process is not running, it returns duration between start and
// exit.
func (runner *Runner) GetUptime() time.Duration {
	if runner.isRunning {
		return time.Since(runner.startedAt)
	} else {
		return runner.exitedAt.Sub(runner.startedAt)
	}
}

var ErrProcessIsRunning = fmt.Errorf("process is running")
var ErrProcessIsNotRunning = fmt.Errorf("process is not running")

// WithArgs sets the arguments of the process.
func WithArgs(args ...string) RunnerOpts {
	return func(runner *Runner) {
		runner.executableArgs = args
	}
}

func WithStdin(stdin io.Reader) RunnerOpts {
	return func(runner *Runner) {
		runner.stdin = stdin
	}
}

func WithStdout(stdout io.Writer) RunnerOpts {
	return func(runner *Runner) {
		runner.stdout = stdout
	}
}

func WithStderr(stderr io.Writer) RunnerOpts {
	return func(runner *Runner) {
		runner.stderr = stderr
	}
}
