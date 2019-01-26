package job

import (
	"context"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

type Source interface {
	Fetch(context.Context) (Job, error)
}

func NewSource(log logrus.FieldLogger, jobURL string) (Source, error) {
	p := &urlSource{
		log:    log,
		jobURL: jobURL,
	}

	return p, nil
}

type urlSource struct {
	log    logrus.FieldLogger
	jobURL string
}

func (us *urlSource) Fetch(ctx context.Context) (Job, error) {
	u, err := url.Parse(us.jobURL)
	if err != nil {
		return nil, err
	}

	switch u.Scheme {
	case "file":
		return us.jobFromFile(u.Host + u.Path)
	case "http", "https":
		return us.jobFromRemote(u)
	}

	return nil, nil
}

func (us *urlSource) jobFromFile(path string) (Job, error) {
	if path == "-" {
		jobBytes, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return nil, err
		}

		return newJobFromBytes(jobBytes)
	}

	abspath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	jobBytes, err := ioutil.ReadFile(abspath)
	if err != nil {
		return nil, err
	}

	return newJobFromBytes(jobBytes)
}

func (us *urlSource) jobFromRemote(u *url.URL) (Job, error) {
	return nil, nil
}
