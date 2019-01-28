package job

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

type Streamer interface {
	Stream(context.Context, Job, string, io.Reader) error
}

func NewStreamer(log logrus.FieldLogger) Streamer {
	return &httpStreamer{log: log}
}

type httpStreamer struct {
	log logrus.FieldLogger
}

func (hs *httpStreamer) Stream(ctx context.Context, job Job, name string, r io.Reader) error {
	log := hs.log.WithFields(logrus.Fields{
		"self":   "http_streamer",
		"job_id": job.ID(),
		"stream": name,
	})

	for {
		// { TODO: do http stuff
		log.Debug("copying to stdout")
		_, _ = io.Copy(os.Stdout, r)
		// }
		select {
		case <-ctx.Done():
			log.WithError(ctx.Err()).Debug("done streaming")
			return ctx.Err()
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}
