package job

import (
	"encoding/json"
	"fmt"
)

type Job interface {
	ID() string
	JWT() string
	JobStateURL() string
	LogPartsURL() string
	Raw() interface{}
	Script() (string, error)
	Streams() map[string]string
}

type versionPeekJob struct {
	Version uint `json:"version"`
}

func newJobFromBytes(b []byte) (Job, error) {
	vpj := &versionPeekJob{Version: 1}
	err := json.Unmarshal(b, vpj)
	if err != nil {
		return nil, err
	}

	switch vpj.Version {
	case 1:
		j := &jobV1Wrap{J: &jobV1{}}
		err = json.Unmarshal(b, j.J)
		return j, err
	default:
		return nil, fmt.Errorf("unknown job version %v", vpj.Version)
	}
}
