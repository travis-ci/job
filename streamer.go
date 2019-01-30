package job

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

type Streamer interface {
	Stream(context.Context, Job, Stream) error
}

func NewStreamer(log logrus.FieldLogger) Streamer {
	return &urlStreamer{log: log.WithField("self", "http_streamer")}
}

type urlStreamer struct {
	log logrus.FieldLogger
}

func (hs *urlStreamer) Stream(ctx context.Context, job Job, str Stream) error {
	log := hs.log.WithFields(logrus.Fields{
		"job_id": job.ID(),
	})

	if str.Source() == nil || str.Dest() == nil {
		err := fmt.Errorf("stream missing source/dest")
		log.WithError(err).Error("cannot stream")
		return err
	}

	for {
		// { TODO: do http stuff
		log.Debug("copying to stdout")
		_, _ = io.Copy(os.Stdout, str.Source().Reader())
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
