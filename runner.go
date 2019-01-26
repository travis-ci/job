package job

import (
	"context"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

type Runner interface {
	Run(context.Context, Job) error
}

func NewRunner(log logrus.FieldLogger) (Runner, error) {
	return &bashRunner{}, nil
}

type bashRunner struct{}

func (bjr *bashRunner) Run(ctx context.Context, job Job) error {
	fmt.Fprintf(os.Stderr, "---> not really running %v\n", job)
	return nil
}
