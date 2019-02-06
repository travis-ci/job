package job

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/cenk/backoff"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	remoteSourceNoJobErr = fmt.Errorf("no jobs available")
)

type Source interface {
	Fetch(context.Context) (Job, error)
}

func NewRemoteSource(log logrus.FieldLogger, jobURL, processorID string) Source {
	return &remoteSource{
		log:         log,
		jobURL:      jobURL,
		processorID: processorID,
	}
}

func NewLocalSource(log logrus.FieldLogger, jobPath, processorID string) Source {
	return &localSource{
		log:         log,
		jobPath:     jobPath,
		processorID: processorID,
	}
}

type remoteSource struct {
	log         logrus.FieldLogger
	jobURL      string
	processorID string
}

func (rs *remoteSource) Fetch(ctx context.Context) (Job, error) {
	u, err := url.Parse(rs.jobURL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse job URL")
	}

	popURL, err := url.Parse(u.String())
	if err != nil {
		return nil, errors.Wrap(err, "failed to copy job pop URL")
	}

	popURL.Path = "/jobs/pop"
	client := &http.Client{}

	req, err := http.NewRequest("POST", popURL.String(), nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create job-board job pop request")
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Travis-Site", "com")
	req.Header.Add("From", rs.processorID)
	req = req.WithContext(ctx)

	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to make job-board job pop request")
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return nil, remoteSourceNoJobErr
	}

	fetchResponsePayload := map[string]string{"job_id": ""}
	err = json.NewDecoder(resp.Body).Decode(&fetchResponsePayload)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode job-board job pop response")
	}

	fetchedJobID, err := strconv.ParseUint(fetchResponsePayload["job_id"], 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse job ID")
	}

	jobURL, err := url.Parse(u.String())
	if err != nil {
		return nil, errors.Wrap(err, "failed to copy job URL")
	}

	jobURL.Path = fmt.Sprintf("/jobs/%d", fetchedJobID)

	req, err = http.NewRequest("GET", jobURL.String(), nil)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't make job-board job request")
	}

	req.Header.Add("Travis-Infrastructure", "detached")
	req.Header.Add("Travis-Site", "com")
	req.Header.Add("From", rs.processorID)
	req = req.WithContext(ctx)

	bo := backoff.NewExponentialBackOff()
	bo.MaxInterval = 10 * time.Second
	bo.MaxElapsedTime = 1 * time.Minute

	respCapture := map[string]*http.Response{"resp": nil}
	err = backoff.Retry(func() (err error) {
		resp, err = client.Do(req)
		if resp != nil && resp.StatusCode != http.StatusOK {
			rs.log.WithFields(logrus.Fields{
				"expected_status": http.StatusOK,
				"actual_status":   resp.StatusCode,
			}).Debug("job fetch failed")

			if resp.Body != nil {
				resp.Body.Close()
			}

			return errors.Errorf("expected %d but got %d", http.StatusOK, resp.StatusCode)
		}
		respCapture["resp"] = resp
		return
	}, bo)

	if err != nil {
		return nil, errors.Wrap(err, "error making job-board job request")
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "error reading body from job-board job request")
	}

	return newJobFromBytes(body)
}

type localSource struct {
	log         logrus.FieldLogger
	jobPath     string
	processorID string
}

func (ls *localSource) Fetch(ctx context.Context) (Job, error) {
	if ls.jobPath == "-" {
		jobBytes, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return nil, err
		}

		return newJobFromBytes(jobBytes)
	}

	abspath, err := filepath.Abs(ls.jobPath)
	if err != nil {
		return nil, err
	}

	jobBytes, err := ioutil.ReadFile(abspath)
	if err != nil {
		return nil, err
	}

	return newJobFromBytes(jobBytes)
}
