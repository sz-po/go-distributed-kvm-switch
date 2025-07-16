package utils

import (
	"fmt"
	"strings"
	"time"
)

type Time time.Time

type Duration time.Duration

func (t *Time) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", time.Time(*t).Format(time.RFC3339))), nil
}

func (t *Time) UnmarshalJSON(b []byte) error {
	tm, err := time.Parse(time.RFC3339, string(b))
	if err != nil {
		return err
	}
	*t = Time(tm)
	return nil
}

func (t *Duration) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", time.Duration(*t).String())), nil
}

func (t *Duration) UnmarshalJSON(b []byte) error {
	stripped := strings.Trim(string(b), "\"")
	d, err := time.ParseDuration(stripped)
	if err != nil {
		return err
	}
	*t = Duration(d)
	return nil
}
