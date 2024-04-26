package device

import (
	"encoding/json"
	"fmt"
)

type Id string

const IdEmpty = Id("")

type Kind string

type ProcessStatus string

type Specification struct {
	Metadata Metadata `json:"metadata"`
	Config   Config   `json:"config"`
}

type Metadata struct {
	Name string `json:"name"`
	Kind Kind   `json:"kind"`
}

type Config json.RawMessage

type Status struct {
	CreatedAt string        `json:"createdAt,omitempty"`
	Process   ProcessStatus `json:"process"`
}

func (metadata Metadata) Id() Id {
	return Id(fmt.Sprintf("%s.%s", metadata.Name, metadata.Kind))
}
