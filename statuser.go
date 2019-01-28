package job

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jtacoma/uritemplates"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Statuser interface {
	Status(context.Context, Job, string, string) error
}

func NewStatuser(log logrus.FieldLogger) Statuser {
	return &httpStatuser{
		log: log,
	}
}

type httpStatuser struct {
	log logrus.FieldLogger
}

func (hs *httpStatuser) Status(ctx context.Context, job Job, curState, newState string) error {
	payload := hs.createStateUpdateBody(job, curState, newState)

	encodedPayload, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrap(err, "error encoding json")
	}

	template, err := uritemplates.Parse(job.JobStateURL())
	if err != nil {
		return errors.Wrap(err, "couldn't parse base URL template")
	}

	u, err := template.Expand(map[string]interface{}{
		"job_id": job.ID(),
	})
	if err != nil {
		return errors.Wrap(err, "couldn't expand base URL template")
	}

	req, err := http.NewRequest("PATCH", u, bytes.NewReader(encodedPayload))
	if err != nil {
		return errors.Wrap(err, "couldn't create request")
	}
	req = req.WithContext(ctx)

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", job.JWT()))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "error making state update request")
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("expected %d, but got %d", http.StatusOK, resp.StatusCode)
	}

	return nil
}

func (hs *httpStatuser) createStateUpdateBody(job Job, curState, newState string) map[string]interface{} {
	body := map[string]interface{}{
		"id":    job.ID(),
		"state": newState,
		"cur":   curState,
		"new":   newState,
		"meta": map[string]interface{}{
			// FIXME: track state_update_count really
			"state_update_count": 1,
		},
	}

	/* TODO: {
	if job.Payload().Job.QueuedAt != nil {
		body["queued_at"] = job.Payload().Job.QueuedAt.UTC().Format(time.RFC3339)
	}
	if !job.received.IsZero() {
		body["received_at"] = job.received.UTC().Format(time.RFC3339)
	}
	if !job.started.IsZero() {
		body["started_at"] = job.started.UTC().Format(time.RFC3339)
	}
	if !job.finished.IsZero() {
		body["finished_at"] = job.finished.UTC().Format(time.RFC3339)
	}

	if job.Payload().Trace {
		body["trace"] = true
	}
	} */

	return body
}
