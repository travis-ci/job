package job

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

type Job interface {
	ID() string
	JWT() string
	JobStateURL() string
	LogPartsURL() string
	Raw() interface{}
	Script() (string, error)
	Streams() map[string]Stream
}

func newJobFromBytes(b []byte) (Job, error) {
	j := &jobWrapper{J: &job{}}
	err := json.Unmarshal(b, j.J)
	return j, err
}

type job struct {
	Data        *jobData      `json:"data"`
	JobScript   *jobJobScript `json:"job_script"`
	JobStateURL string        `json:"job_state_url"`
	LogPartsURL string        `json:"log_parts_url"`
	JWT         string        `json:"jwt"`
	ImageName   string        `json:"image_name"`
}

type jobData struct {
	Type       string                 `json:"type"`
	Job        *jobDataJob            `json:"job"`
	Build      *jobDataBuild          `json:"source"`
	Repository *jobDataRepository     `json:"repository"`
	UUID       string                 `json:"uuid"`
	Config     map[string]interface{} `json:"config"`
	Timeouts   *jobDataTimeouts       `json:"timeouts,omitempty"`
	VMType     string                 `json:"vm_type"`
	VMConfig   *jobDataVMConfig       `json:"vm_config"`
	Meta       *jobDataMeta           `json:"meta"`
	Queue      string                 `json:"queue"`
	Trace      bool                   `json:"trace"`
	Warmer     bool                   `json:"warmer"`
	Streams    map[string]Stream      `json:"streams"`
}

type jobDataJob struct {
	ID       uint64     `json:"id"`
	Number   string     `json:"number"`
	QueuedAt *time.Time `json:"queued_at"`
}

type jobJobScript struct {
	Name     string `json:"name"`
	Encoding string `json:"encoding"`
	Content  string `json:"content"`
}

type jobDataBuild struct {
	ID     uint64 `json:"id"`
	Number string `json:"number"`
}

type jobDataRepository struct {
	ID   uint64 `json:"id"`
	Slug string `json:"slug"`
}

type jobDataTimeouts struct {
	HardLimit  uint64 `json:"hard_limit"`
	LogSilence uint64 `json:"log_silence"`
}

type jobDataMeta struct {
	StateUpdateCount uint `json:"state_update_count"`
}

type jobDataVMConfig struct {
	GpuCount uint64 `json:"gpu_count"`
	GpuType  string `json:"gpu_type"`
	Zone     string `json:"zone"`
}

type jobWrapper struct {
	J *job
}

func (j *jobWrapper) data() *jobData {
	if j.J == nil {
		return nil
	}
	return j.J.Data
}

func (j *jobWrapper) ID() string {
	data := j.data()
	if data != nil && data.Job != nil {
		return fmt.Sprintf("%v", data.Job.ID)
	}

	return ""
}

func (j *jobWrapper) JobStateURL() string {
	if j.J != nil {
		return j.J.JobStateURL
	}

	return ""
}

func (j *jobWrapper) LogPartsURL() string {
	if j.J != nil {
		return j.J.LogPartsURL
	}

	return ""
}

func (j *jobWrapper) JWT() string {
	if j.J != nil {
		return j.J.JWT
	}

	return ""
}

func (j *jobWrapper) Raw() interface{} {
	if j.J != nil {
		return j.J
	}

	return nil
}

func (j *jobWrapper) Script() (string, error) {
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

func (j *jobWrapper) Streams() map[string]Stream {
	streams := map[string]Stream{}
	data := j.data()
	if data != nil && data.Streams != nil {
		streams = data.Streams
	}

	if _, ok := streams[stdOutErrName]; !ok {
		streams[stdOutErrName] = NewStdOutErrStream()
	}

	return streams
}

func (j *jobWrapper) MarshalJSON() ([]byte, error) {
	outJSON := &bytes.Buffer{}
	err := json.NewEncoder(outJSON).Encode(j.Raw())
	return outJSON.Bytes(), err
}
