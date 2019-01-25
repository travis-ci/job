package proc

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

func NewWaiter(log logrus.FieldLogger, interval time.Duration,
	src Source, runner Runner) Waiter {

	return &fetchRetryWaiter{
		src:      src,
		runner:   runner,
		interval: interval,
		log:      log,
	}
}

type Waiter interface {
	Wait(context.Context) error
}

type fetchRetryWaiter struct {
	log      logrus.FieldLogger
	interval time.Duration
	src      Source
	runner   Runner
}

func (w *fetchRetryWaiter) Wait(ctx context.Context) error {
	for {
		j, err := w.src.Fetch(ctx)
		if err != nil {
			w.log.WithFields(logrus.Fields{
				"err":      err,
				"interval": w.interval,
			}).Debug("waiting for job")
			time.Sleep(w.interval)
			continue
		}

		return w.runner.Run(ctx, j)
	}
}
