package process

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"sync"
	"time"
)

type RunnerOpts func(*Runner)

type Runner struct {
	executablePath string
	executableArgs []string

	isRunning bool
	startedAt time.Time
	exitedAt  time.Time

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

	return runner
}

func (runner *Runner) Start() error {
	if runner.isRunning {
		return ErrProcessIsAlreadyRunning
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

func (runner *Runner) Stop() error {
	if !runner.isRunning {
		return ErrProcessIsNotRunning
	}

	runner.logger.Debug("Killing process.", slog.Int("pid", runner.pid))
	err := runner.process.Process.Signal(os.Kill)
	if err != nil {
		return err
	}

	runner.wg.Wait()

	runner.logger.Info("Process killed.", slog.Int("pid", runner.pid))

	return nil
}

func (runner *Runner) IsRunning() bool {
	return runner.isRunning
}

func (runner *Runner) GetPID() (int, error) {
	if !runner.isRunning {
		return 0, ErrProcessIsNotRunning
	}

	return runner.pid, nil
}

func (runner *Runner) GetExitCode() (int, error) {
	if !runner.isRunning {
		return 0, ErrProcessIsNotRunning
	}

	return runner.exitCode, nil
}

func (runner *Runner) GetUptime() time.Duration {
	if runner.isRunning {
		return time.Since(runner.startedAt)
	} else {
		return runner.exitedAt.Sub(runner.startedAt)
	}
}

var ErrProcessIsAlreadyRunning = fmt.Errorf("process is already running")
var ErrProcessIsNotRunning = fmt.Errorf("process is not running")

func WithArgs(args ...string) RunnerOpts {
	return func(runner *Runner) {
		runner.executableArgs = args
	}
}

func WithStdin(stdin io.Writer) RunnerOpts {
	return func(runner *Runner) {

	}
}

func WithStdout(stdout io.Reader) RunnerOpts {
	return func(runner *Runner) {

	}
}

func WithStderr(stderr io.Reader) RunnerOpts {
	return func(runner *Runner) {

	}
}
