package pending

import (
	"encoding/json"
)

type Status int

func (i Status) MarshalBinary() ([]byte, error) {
	return json.Marshal(i)
}

const (
	Wait Status = iota
	Ready
	Cancel
)

type Statuses map[string]Status
