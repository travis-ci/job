package job

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/jtacoma/uritemplates"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Statuser interface {
	Status(context.Context, Job, StateUpdate) error
}

func NewStatuser(log logrus.FieldLogger) Statuser {
	return &urlStatuser{
		log: log.WithField("self", "http_statuser"),
	}
}

type urlStatuser struct {
	log logrus.FieldLogger
}

func (us *urlStatuser) Status(ctx context.Context, job Job, stateUpdate StateUpdate) error {
	template, err := uritemplates.Parse(job.JobStateURL())
	if err != nil {
		return errors.Wrap(err, "couldn't parse base URL template")
	}

	expanded, err := template.Expand(map[string]interface{}{
		"job_id": job.ID(),
	})
	if err != nil {
		return errors.Wrap(err, "couldn't expand base URL template")
	}

	u, err := url.Parse(expanded)
	if err != nil {
		return errors.Wrap(err, "couldn't parse expanded URL")
	}

	switch u.Scheme {
	case "file":
		return us.updateViaFile(ctx, job, u, stateUpdate)
	case "http", "https":
		return us.updateViaHTTP(ctx, job, u, stateUpdate)
	default:
		return fmt.Errorf("unknown scheme %v", u.Scheme)
	}
}

func (us *urlStatuser) updateViaFile(ctx context.Context, job Job, u *url.URL, stateUpdate StateUpdate) error {
	dest, err := filepath.Abs(u.Host + u.Path)
	if err != nil {
		return errors.Wrap(err, "failed to find absolute dest path")
	}
	return ioutil.WriteFile(dest, []byte(fmt.Sprintf("%v\n", stateUpdate.New())), os.FileMode(0644))
}

func (us *urlStatuser) updateViaHTTP(ctx context.Context, job Job, u *url.URL, stateUpdate StateUpdate) error {
	log := us.log.WithFields(logrus.Fields{
		"job_id":    job.ID(),
		"cur_state": stateUpdate.Cur(),
		"new_state": stateUpdate.New(),
	})

	log.Debug("serializing payload")
	encodedPayload, err := json.Marshal(stateUpdate)
	if err != nil {
		log.WithError(err).Debug("error encoding json")
		return errors.Wrap(err, "error encoding json")
	}

	req, err := http.NewRequest("PATCH", u.String(), bytes.NewReader(encodedPayload))
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
