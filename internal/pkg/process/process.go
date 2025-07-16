package process

import (
	"context"
	"fmt"
	"github.com/sz-po/go-distributed-kvm-switch/pkg/utils"
	"regexp"
)

type ProcessOpt any

// Name represents a unique identifier for a instance.
type Name string

// RestartMode defines the behavior for restarting a instance when it exits.
type RestartMode string

const (
	// Always indicates that the instance should be restarted immediately after it exits, regardless of exit status.
	Always RestartMode = "always"

	// Never indicates that the instance should not be restarted after it exits.
	Never RestartMode = "never"

	// OnError indicates that the instance should only be restarted if it exits with a non-zero (error) status.
	OnError RestartMode = "on-error"
)

// Phase represents the current lifecycle stage of a running instance.
type Phase string

const (
	// Idle indicates that the instance is not yet started or has been explicitly stopped.
	Idle Phase = "idle"

	// Starting indicates that the instance is in the phase of being launched.
	Starting Phase = "starting"

	// Running indicates that the instance is currently executing.
	Running Phase = "running"

	// Terminating indicates that the instance is in the phase of being terminated (by receiving the SIGTERM signal).
	Terminating Phase = "terminating"

	// Killing indicates that the instance is in the phase of being killed (by receiving the SIGKILL signal).
	Killing Phase = "killing"
)

// Specification defines how to launch a new instance.
type Specification struct {
	// ExecutablePath is the file system path to the binary or script to be executed.
	ExecutablePath string `json:"executablePath" validate:"required,filepath"`

	// WorkingDirectoryPath is an optional custom working directory for the instance.
	// If nil, the instance inherits the caller's working directory.
	WorkingDirectoryPath *string `json:"workingDirectoryPath,omitempty" validate:"omitempty,dirpath"`

	// EnvironmentVariables is an optional map of environment-variable key–value pairs
	// that will be injected into the new instance.
	EnvironmentVariables map[string]string `json:"environmentVariables,omitempty" validate:"omitempty,dive,keys,required,endkeys,required" default:"{}"`

	// Arguments is an optional list of command-line arguments passed to the executable.
	Arguments []string `json:"arguments,omitempty" validate:"omitempty,dive,required"`

	// InheritEnvironmentVariables indicates whether the instance should inherit the caller's environment variables.
	InheritEnvironmentVariables bool `json:"inheritEnvironmentVariables" default:"false"`

	// RestartMode defines the behavior for restarting a process instance when it exits.
	RestartMode RestartMode `json:"restartMode" default:"on-error"`

	// AutoEnable indicates whether the process instance should be enabled automatically when it is created.
	AutoEnable bool `json:"autoEnable"`

	// KillTimeout is the maximum amount of time that the process should wait for the instance to be killed after
	// unsuccessful termination.
	KillTimeout utils.Duration `json:"killTimeout" default:"5s"`
}

// Status represents the current runtime state of a process instance.
type Status struct {
	// Enabled indicates whether the process instance should be started or not.
	Enabled bool `json:"enabled"`

	// Phase indicates the lifecycle stage of the process instance (e.g., Idle, Running).
	Phase Phase `json:"phase"`

	// ProcessId is the numeric process ID assigned by the operating system, if the instance has been started.
	// It is omitted if the instance has not yet been assigned an ID.
	ProcessId *int `json:"id,omitempty"`

	// ExitCode is the exit code of the process instance, if it has exited.
	ExitCode *int `json:"exitCode,omitempty"`

	// ErrorMessage is the error message of the process instance, if it has terminated with an error.
	ErrorMessage *string `json:"errorMessage,omitempty"`

	// CreatedAt is the time when the process instance was initially created.
	CreatedAt utils.Time `json:"createdAt"`

	// RestartCount is the number of times the process instance has been restarted.
	RestartCount int `json:"restartCount"`

	// FailureCount is the number of restarts triggered when the process terminated
	// with a non-zero exit code (i.e., a failure).
	FailureCount int `json:"failureCount"`

	// FailureBackoff is the current backoff duration after a failure.
	FailureBackoff utils.Duration `json:"failureBackoff"`

	// EnvironmentVariables is a map of environment variable key-value pairs that were injected into the new process
	// instance. It may be different from the original specification, if InheritEnvironmentVariables is true.
	EnvironmentVariables map[string]string `json:"environmentVariables"`

	// WorkingDirectoryPath is the file system path to the process instance's working directory. It may be different from
	// the original specification, if WorkingDirectoryPath is not nil (inherit the caller's working directory).
	WorkingDirectoryPath *string `json:"workingDirectoryPath,omitempty"`
}

type ProcessFactory func(name Name, specification Specification, opts ...ProcessOpt) (Process, error)

type Process interface {
	GetStatus() Status
	GetSpecification() Specification
	Enable(ctx context.Context) error
	Disable(ctx context.Context) error
	Restart(ctx context.Context) error
	Wait(ctx context.Context, phase Phase) error
}

func (name Name) Validate() error {
	regex := regexp.MustCompile("^[a-z0-9-]+$")

	if !regex.MatchString(string(name)) {
		return fmt.Errorf("name must only contain small letters and hyphens")
	}
	return nil
}

var ErrProcessAlreadyRunning = fmt.Errorf("process already running")
var ErrProcessNotRunning = fmt.Errorf("process not running")
var ErrProcessAlreadyEnabled = fmt.Errorf("process already enabled")
var ErrProcessAlreadyDisabled = fmt.Errorf("process already disabled")
var ErrInvalidProcessName = fmt.Errorf("invalid process name")
