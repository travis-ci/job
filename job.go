package job

import (
	"encoding/json"
	"fmt"
)

type Job interface {
	ID() string
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

	var j Job
	switch vpj.Version {
	case 1:
		j = &jobV1{}
	default:
		return nil, fmt.Errorf("unknown job version %v", vpj.Version)
	}

	err = json.Unmarshal(b, j)
	return j, err
}
