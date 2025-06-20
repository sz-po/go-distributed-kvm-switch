// process_local_test.go
package process

import (
	"context"
	"github.com/coder/quartz"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/sz-po/go-distributed-kvm-switch/internal/pkg/api/utils"
	"go.openly.dev/pointy"
	"os"
	"testing"
	"time"
)

func TestProcess_Specification_AutoEnable(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	process, err := NewLocalProcess(Name("foo"), Specification{
		ExecutablePath: "/bin/sleep",
		Arguments: []string{
			"5",
		},
		AutoEnable: true,
	})
	assert.NoError(t, err)

	timeoutCtx, timeoutCancel := context.WithTimeout(ctx, 1*time.Second)
	defer timeoutCancel()
	assert.NoError(t, process.Wait(timeoutCtx, Running))
}

func TestProcess_Enable(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	process, err := NewLocalProcess(Name("foo"), Specification{
		ExecutablePath: "/bin/sleep",
		Arguments: []string{
			"5",
		},
		AutoEnable: false,
	})

	timeoutCtx, timeoutCancel := context.WithTimeout(ctx, 1*time.Second)
	defer timeoutCancel()
	assert.NoError(t, process.Wait(timeoutCtx, Idle))

	assert.Equal(t, Idle, process.GetStatus().Phase)
	assert.NoError(t, err)
	assert.NoError(t, process.Enable(ctx))

	assert.ErrorIs(t, process.Enable(ctx), ErrProcessAlreadyEnabled)

	timeoutCtx, timeoutCancel = context.WithTimeout(ctx, 1*time.Second)
	defer timeoutCancel()
	assert.NoError(t, process.Wait(timeoutCtx, Running))
}

func TestProcess_Disable(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	process, err := NewLocalProcess(Name("foo"), Specification{
		ExecutablePath: "/bin/sleep",
		Arguments: []string{
			"5",
		},
		AutoEnable: true,
	})
	assert.NoError(t, err)

	timeoutCtx, timeoutCancel := context.WithTimeout(ctx, 1*time.Second)
	defer timeoutCancel()
	assert.NoError(t, process.Wait(timeoutCtx, Running))

	assert.NoError(t, process.Disable(ctx))
	assert.ErrorIs(t, process.Disable(ctx), ErrProcessAlreadyDisabled)

	timeoutCtx, timeoutCancel = context.WithTimeout(ctx, 1*time.Second)
	defer timeoutCancel()
	assert.NoError(t, process.Wait(timeoutCtx, Idle))
}

func TestProcess_buildWorkingDirectory_ProvidedInSpecification(t *testing.T) {
	process, err := NewLocalProcess(Name("foo"), Specification{
		ExecutablePath:       "/bin/sleep",
		WorkingDirectoryPath: pointy.String("/tmp"),
	})
	assert.NoError(t, err)

	workingDirectory, err := process.buildWorkingDirectory()
	assert.NoError(t, err)
	assert.Equal(t, "/tmp", workingDirectory)
}

func TestProcess_buildWorkingDirectory_ProvidedInSpecification_File(t *testing.T) {
	process, err := NewLocalProcess(Name("foo"), Specification{
		ExecutablePath:       "/bin/sleep",
		WorkingDirectoryPath: pointy.String("/bin/sleep"),
	})
	assert.ErrorAs(t, err, &validator.ValidationErrors{})
	assert.Nil(t, process)

}

func TestProcess_buildWorkingDirectory_ProvidedInSpecification_NotExists(t *testing.T) {
	process, err := NewLocalProcess(Name("foo"), Specification{
		ExecutablePath:       "/bin/sleep",
		WorkingDirectoryPath: pointy.String("/tmp/non-existent-directory"),
	})
	assert.ErrorAs(t, err, &validator.ValidationErrors{})
	assert.Nil(t, process)
}

func TestProcess_buildWorkingDirectory_NotProvidedInSpecification(t *testing.T) {
	process, err := NewLocalProcess(Name("foo"), Specification{
		ExecutablePath:       "/bin/sleep",
		WorkingDirectoryPath: nil,
	})
	assert.NoError(t, err)

	currentWorkingDirectory, err := os.Getwd()
	assert.NoError(t, err)
	workingDirectory, err := process.buildWorkingDirectory()
	assert.NoError(t, err)
	assert.Equal(t, currentWorkingDirectory, workingDirectory)
}

func TestProcess_GetSpecification(t *testing.T) {
	specification := Specification{
		ExecutablePath:       "/bin/sleep",
		WorkingDirectoryPath: nil,
		EnvironmentVariables: map[string]string{
			"FOO": "BAR",
		},
		Arguments:   []string{"foo", "bar"},
		RestartMode: Always,
		AutoEnable:  true,
		KillTimeout: utils.Duration(time.Second),
	}

	process, err := NewLocalProcess(Name("foo"), specification)
	assert.NoError(t, err)
	assert.Equal(t, specification, process.GetSpecification())
}

func TestProcess_Restart(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	process, err := NewLocalProcess(Name("foo"), Specification{
		ExecutablePath: "/bin/sleep",
		Arguments: []string{
			"5",
		},
		AutoEnable: false,
	})
	assert.NoError(t, err)

	assert.Equal(t, 0, process.GetStatus().RestartCount)
	err = process.Enable(ctx)
	assert.NoError(t, err)

	timeoutCtx, timeoutCancel := context.WithTimeout(ctx, 1*time.Second)
	defer timeoutCancel()
	assert.NoError(t, process.Wait(timeoutCtx, Running))
	assert.Equal(t, 0, process.GetStatus().RestartCount)

	timeoutCtx, timeoutCancel = context.WithTimeout(ctx, 1*time.Second)
	defer timeoutCancel()
	assert.NoError(t, process.Restart(timeoutCtx))

	timeoutCtx, timeoutCancel = context.WithTimeout(ctx, 1*time.Second)
	defer timeoutCancel()
	assert.NoError(t, process.Wait(timeoutCtx, Running))
	assert.Equal(t, 1, process.GetStatus().RestartCount)
}

func TestCreateLocalProcessFactory(t *testing.T) {
	processFactory := CreateLocalProcessFactory()
	assert.NotNil(t, processFactory)

	process, err := processFactory(Name("foo"), Specification{
		ExecutablePath: "/bin/sleep",
		Arguments: []string{
			"5",
		},
		AutoEnable: false,
	})
	assert.NoError(t, err)
	assert.NotNil(t, process)
	assert.Equal(t, "/bin/sleep", process.GetSpecification().ExecutablePath)

	process, err = processFactory(Name("foo"), Specification{
		ExecutablePath: "/bin/sleep",
		Arguments: []string{
			"5",
		},
		AutoEnable: false,
	}, WithClock(quartz.NewMock(t)))
	assert.NoError(t, err)
	assert.NotNil(t, process)

	process, err = processFactory(Name("foo"), Specification{
		ExecutablePath: "/bin/sleep",
		Arguments: []string{
			"5",
		},
		AutoEnable: false,
	}, string("invalid-opt"))
	assert.ErrorIs(t, err, ErrInvalidLocalProcessOpt)
	assert.Nil(t, process)
}
