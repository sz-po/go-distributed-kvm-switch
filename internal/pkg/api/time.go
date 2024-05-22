package api

import "time"

const TimestampFormat = time.RFC3339Nano

type Duration string
type Timestamp string

func Now() Timestamp {
	return NewTimestamp(time.Now())
}

func NewTimestamp(t time.Time) Timestamp {
	return Timestamp(t.Format(TimestampFormat))
}

func (timestamp Timestamp) Time() *time.Time {
	t, err := time.Parse(TimestampFormat, string(timestamp))
	if err != nil {
		return nil
	}

	return &t
}

func (timestamp Timestamp) IsEmpty() bool {
	return timestamp == Timestamp("")
}
