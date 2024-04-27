package device

import (
	"encoding/json"
	"fmt"
	"time"
)

type Id string

const IdEmpty = Id("")

type Kind string

type ProcessState string

const (
	Stopped  ProcessState = "stopped"
	Starting              = "starting"
	Running               = "running"
	Stopping              = "stopping"
)

type Metadata struct {
	Name string `json:"name"`
	Kind Kind   `json:"kind"`
}

type Config json.RawMessage

type StatusTimestamp time.Time

type ProcessStatus struct {
	StateChangedAt StatusTimestamp `json:"stateChangedAt"`
	State          ProcessState    `json:"state"`
	Id             int             `json:"id"`
}

type Status struct {
	CreatedAt StatusTimestamp `json:"createdAt"`
	Enabled   bool            `json:"enabled"`

	Process ProcessStatus `json:"process"`
}

func (metadata Metadata) Id() Id {
	return Id(fmt.Sprintf("%s.%s", metadata.Name, metadata.Kind))
}

func now() StatusTimestamp {
	return StatusTimestamp(time.Now())
}
