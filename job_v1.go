package job

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
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
	Streams    map[string]string      `json:"streams"`
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

type jobV1Wrap struct {
	J *jobV1
}

func (j *jobV1Wrap) data() *jobV1Data {
	if j.J == nil {
		return nil
	}
	return j.J.Data
}

func (j *jobV1Wrap) ID() string {
	data := j.data()
	if data != nil && data.Job != nil {
		return fmt.Sprintf("%v", data.Job.ID)
	}

	return ""
}

func (j *jobV1Wrap) JobStateURL() string {
	if j.J != nil {
		return j.J.JobStateURL
	}

	return ""
}

func (j *jobV1Wrap) LogPartsURL() string {
	if j.J != nil {
		return j.J.LogPartsURL
	}

	return ""
}

func (j *jobV1Wrap) JWT() string {
	if j.J != nil {
		return j.J.JWT
	}

	return ""
}

func (j *jobV1Wrap) Raw() interface{} {
	if j.J != nil {
		return j.J
	}

	return nil
}

func (j *jobV1Wrap) Script() (string, error) {
	if j.J == nil || j.J.JobScript == nil {
		return "", nil
	}

	script := j.J.JobScript
	if script.Content == "" {
		return "", nil
	}

	if script.Encoding != "base64" {
		return "", fmt.Errorf("unknown job script encoding: %s", script.Encoding)
	}

	decoded, err := base64.StdEncoding.DecodeString(script.Content)
	return string(decoded), err
}

func (j *jobV1Wrap) Streams() map[string]string {
	streams := map[string]string{}
	data := j.data()
	if data != nil && data.Streams != nil {
		streams = data.Streams
	}

	if _, ok := streams["stdouterr"]; !ok {
		streams["stdouterr"] = "-"
	}

	return streams
}

func (j *jobV1Wrap) MarshalJSON() ([]byte, error) {
	outJSON := &bytes.Buffer{}
	err := json.NewEncoder(outJSON).Encode(j.Raw())
	return outJSON.Bytes(), err
}
