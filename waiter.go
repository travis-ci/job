package job

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

func NewWaiter(log logrus.FieldLogger, interval, max time.Duration,
	src Source, runner Runner) Waiter {

	return &fetchRetryWaiter{
		log:      log,
		interval: interval,
		max:      max,
		src:      src,
		runner:   runner,
	}
}

type Waiter interface {
	Wait(context.Context) error
}

type fetchRetryWaiter struct {
	log           logrus.FieldLogger
	interval, max time.Duration
	src           Source
	runner        Runner
}

func (w *fetchRetryWaiter) Wait(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, w.max)
	defer cancel()

	for {
		j, err := w.src.Fetch(ctx)
		if err != nil {
			w.log.WithFields(logrus.Fields{
				"err":      err,
				"interval": w.interval,
			}).Debug("waiting for job")

			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				time.Sleep(w.interval)
				continue
			}
		}

		return w.runner.Run(ctx, j)
	}
}
