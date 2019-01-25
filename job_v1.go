package job

import (
	"fmt"
	"time"
)

type jobV1 struct {
	Data        *jobV1Data      `json:"data"`
	JobScript   *jobV1JobScript `json:"job_script"`
	JobStateURL string          `json:"job_state_url"`
	LogPartsURL string          `json:"log_parts_url"`
	JWT         string          `json:"jwt"`
	ImageName   string          `json:"image_name"`
}

type jobV1Data struct {
	Type       string                 `json:"type"`
	Job        *jobV1DataJob          `json:"job"`
	Build      *jobV1DataBuild        `json:"source"`
	Repository *jobV1DataRepository   `json:"repository"`
	UUID       string                 `json:"uuid"`
	Config     map[string]interface{} `json:"config"`
	Timeouts   *jobV1DataTimeouts     `json:"timeouts,omitempty"`
	VMType     string                 `json:"vm_type"`
	VMConfig   *jobV1DataVMConfig     `json:"vm_config"`
	Meta       *jobV1DataMeta         `json:"meta"`
	Queue      string                 `json:"queue"`
	Trace      bool                   `json:"trace"`
	Warmer     bool                   `json:"warmer"`
}

type jobV1DataJob struct {
	ID       uint64     `json:"id"`
	Number   string     `json:"number"`
	QueuedAt *time.Time `json:"queued_at"`
}

type jobV1JobScript struct {
	Name     string `json:"name"`
	Encoding string `json:"encoding"`
	Content  string `json:"content"`
}

type jobV1DataBuild struct {
	ID     uint64 `json:"id"`
	Number string `json:"number"`
}

type jobV1DataRepository struct {
	ID   uint64 `json:"id"`
	Slug string `json:"slug"`
}

type jobV1DataTimeouts struct {
	HardLimit  uint64 `json:"hard_limit"`
	LogSilence uint64 `json:"log_silence"`
}

type jobV1DataMeta struct {
	StateUpdateCount uint `json:"state_update_count"`
}

type jobV1DataVMConfig struct {
	GpuCount uint64 `json:"gpu_count"`
	GpuType  string `json:"gpu_type"`
	Zone     string `json:"zone"`
}

func (j *jobV1) ID() string {
	if j.Data != nil && j.Data.Job != nil {
		return fmt.Sprintf("%v", j.Data.Job.ID)
	}

	return ""
}
