package utils

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTime_MarshalJSON(t *testing.T) {
	base := time.Date(2021, 12, 31, 23, 59, 59, 0, time.UTC)
	tVal := Time(base)

	bytes, err := tVal.MarshalJSON()
	assert.NoError(t, err)

	expected := []byte(fmt.Sprintf("%q", time.Time(base).Format(time.RFC3339)))
	assert.Equal(t, expected, bytes)
}

func TestTime_UnmarshalJSON_Success(t *testing.T) {
	inputStr := "2021-12-31T23:59:59Z"
	var tVal Time

	err := tVal.UnmarshalJSON([]byte(inputStr))
	assert.NoError(t, err)

	expected := Time(time.Date(2021, 12, 31, 23, 59, 59, 0, time.UTC))
	assert.Equal(t, expected, tVal)
}

func TestTime_UnmarshalJSON_Error(t *testing.T) {
	invalidInput := "niepoprawna-data"
	var tVal Time

	err := tVal.UnmarshalJSON([]byte(invalidInput))
	assert.Error(t, err)
}

func TestDuration_MarshalJSON(t *testing.T) {
	dur := Duration(2*time.Hour + 30*time.Minute)

	bytes, err := dur.MarshalJSON()
	assert.NoError(t, err)

	expected := []byte(fmt.Sprintf("%q", (time.Duration)(dur).String()))
	assert.Equal(t, expected, bytes)
}

func TestDuration_UnmarshalJSON_Success(t *testing.T) {
	inputStr := "2h30m0s"
	var dur Duration

	err := dur.UnmarshalJSON([]byte(inputStr))
	assert.NoError(t, err)

	expected := Duration(2*time.Hour + 30*time.Minute)
	assert.Equal(t, expected, dur)
}

func TestDuration_UnmarshalJSON_Error(t *testing.T) {
	invalidInput := "invalid-duration"
	var dur Duration

	err := dur.UnmarshalJSON([]byte(invalidInput))
	assert.Error(t, err)
}
