package proc

import (
	"context"
	"fmt"
	"net/url"

	"github.com/sirupsen/logrus"
)

type Source interface {
	Validate() error
	Fetch(context.Context) (Job, error)
}

func NewSource(log logrus.FieldLogger, jobURL string) (Source, error) {
	p := &httpSource{
		log:    log,
		jobURL: jobURL,
	}

	if err := p.Validate(); err != nil {
		return nil, err
	}

	return p, nil
}

type httpSource struct {
	log    logrus.FieldLogger
	jobURL string
}

func (hjs *httpSource) Fetch(ctx context.Context) (Job, error) {
	return nil, nil
}

func (hp *httpSource) Validate() error {
	u, err := url.Parse(hp.jobURL)

	if err != nil {
		return err
	}

	if u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("invalid url %q", u)
	}

	return nil
}
