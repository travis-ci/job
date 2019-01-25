package job

import (
	"context"

	"github.com/sirupsen/logrus"
)

type Runner interface {
	Run(context.Context, Job) error
	Validate() error
}

func NewRunner(log logrus.FieldLogger) (Runner, error) {
	runner := &bashRunner{}
	return runner, runner.Validate()
}

type bashRunner struct{}

func (bjr *bashRunner) Validate() error {
	return nil
}

func (bjr *bashRunner) Run(ctx context.Context, job Job) error {
	return nil
}
