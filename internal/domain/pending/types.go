package pending

import (
	"encoding/json"
)

type Status int

func (i Status) MarshalBinary() ([]byte, error) {
	return json.Marshal(i)
}

const (
	Undefined Status = iota
	Wait
	Ready
	Cancel
)

type Statuses map[string]Status
