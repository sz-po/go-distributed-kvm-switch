// process_test.go
package process

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestProcess_Enable(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	process, err := NewProcess(Specification{
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
