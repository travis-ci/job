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
	er.status(ctx, job, "queued", "received")

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

	log.Debug("starting stdouterr streamer")
	pr, pw := io.Pipe()
	go er.streamer.Stream(ctx, job, "stdouterr", pr)

	cmd := exec.CommandContext(ctx, er.interpreter, dest)
	cmd.Stdout = pw
	cmd.Stderr = pw

	er.status(ctx, job, "received", "started")
	log.Debug("starting command")
	err = cmd.Start()
	if err != nil {
		log.WithError(err).Error("failed to start command")
		er.status(ctx, job, "started", "failed")
		return errors.Wrap(err, "failed to start command")
	}

	err = cmd.Wait()
	if err != nil {
		log.WithError(err).Error("command wait errored")

		if cmd.ProcessState == nil {
			er.status(ctx, job, "started", "errored")
			return errors.Wrap(err, "no process state found")
		}

		if !cmd.ProcessState.Exited() {
			er.status(ctx, job, "started", "errored")
			return errors.Wrap(err, "process did not exit")
		}

		if !cmd.ProcessState.Success() {
			er.status(ctx, job, "started", "failed")
			return errors.Wrap(err, "process exited without success")
		}

		er.status(ctx, job, "started", "failed")
		return err
	}

	er.status(ctx, job, "started", "passed")
	log.Debug("command completed")
	return nil
}

func (er *execRunner) status(ctx context.Context, job Job, curState, newState string) {
	log := er.log.WithFields(logrus.Fields{
		"job_id": job.ID(),
	})

	statusErr := er.statuser.Status(ctx, job, curState, newState)
	if statusErr != nil {
		log.WithError(statusErr).Error("failed to set job status")
	}
}
