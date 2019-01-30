package job

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Runner interface {
	Run(context.Context, Job) error
}

func NewRunner(log logrus.FieldLogger, statuser Statuser, streamer Streamer) (Runner, error) {
	return &execRunner{
		interpreter: "bash",
		log:         log.WithField("self", "exec_runner"),
		statuser:    statuser,
		streamer:    streamer,
	}, nil
}

type execRunner struct {
	interpreter string
	log         logrus.FieldLogger
	statuser    Statuser
	streamer    Streamer
}

func (er *execRunner) Run(ctx context.Context, job Job) error {
	log := er.log.WithFields(logrus.Fields{
		"job_id": job.ID(),
	})
	er.status(ctx, job, QueuedState, ReceivedState)

	log.Debug("extracting script")
	script, err := job.Script()
	if err != nil {
		log.WithError(err).Error("failed to extract job script")
		return errors.Wrap(err, "failed to extract job script")
	}

	dest := path.Join(os.TempDir(), fmt.Sprintf("travis-job-%v.%s", job.ID(), er.interpreter))
	log.WithFields(logrus.Fields{
		"dest": dest,
		"len":  len(script),
	}).Debug("writing script")
	err = ioutil.WriteFile(dest, []byte(script), os.FileMode(0755))
	if err != nil {
		return errors.Wrap(err, "failed to write job script")
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if _, ok := job.Streams()[stdOutErrName]; !ok {
		log.Error("job is missing stdouterr stream")
		er.status(ctx, job, ReceivedState, ErroredState)
		return fmt.Errorf("missing stdouterr stream")
	}

	stdOutErr := job.Streams()[stdOutErrName]

	log.Debug("starting stdouterr streamer")

	pr, pw := io.Pipe()
	stdOutErr.Source().SetReader(pr)
	stdOutErr.Dest().SetWriter(pw)

	go func() {
		err := er.streamer.Stream(ctx, job, stdOutErr)
		if err != nil {
			log.WithError(err).Error("failure during streaming")
		}
	}()

	cmd := exec.CommandContext(ctx, er.interpreter, dest)
	cmd.Stdout = pw
	cmd.Stderr = pw

	er.status(ctx, job, ReceivedState, StartedState)
	log.Debug("starting command")
	err = cmd.Start()
	if err != nil {
		log.WithError(err).Error("failed to start command")
		er.status(ctx, job, StartedState, FailedState)
		return errors.Wrap(err, "failed to start command")
	}

	err = cmd.Wait()
	if err != nil {
		log.WithError(err).Error("command wait errored")

		if cmd.ProcessState == nil {
			er.status(ctx, job, StartedState, ErroredState)
			return errors.Wrap(err, "no process state found")
		}

		if !cmd.ProcessState.Exited() {
			er.status(ctx, job, StartedState, ErroredState)
			return errors.Wrap(err, "process did not exit")
		}

		if !cmd.ProcessState.Success() {
			er.status(ctx, job, StartedState, FailedState)
			return errors.Wrap(err, "process exited without success")
		}

		er.status(ctx, job, StartedState, FailedState)
		return err
	}

	er.status(ctx, job, StartedState, PassedState)
	log.Debug("command completed")
	return nil
}

func (er *execRunner) status(ctx context.Context, job Job, curState, newState State) {
	log := er.log.WithFields(logrus.Fields{
		"job_id": job.ID(),
	})

	statusErr := er.statuser.Status(ctx, job, NewStateUpdate(job.ID(), curState, newState))
	if statusErr != nil {
		log.WithError(statusErr).Error("failed to set job status")
	}
}
