package job

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Runner interface {
	Run(context.Context, Job) error
}

func NewRunner(log logrus.FieldLogger, statuser Statuser, streamer Streamer) (Runner, error) {
	return &bashRunner{
		log:      log,
		statuser: statuser,
		streamer: streamer,
	}, nil
}

type bashRunner struct {
	log      logrus.FieldLogger
	statuser Statuser
	streamer Streamer
}

func (br *bashRunner) Run(ctx context.Context, job Job) error {
	log := br.log.WithField("job_id", job.ID())

	log.Debug("extracting script")
	script, err := job.Script()
	if err != nil {
		log.WithError(err).Error("failed to extract job script")
		return errors.Wrap(err, "failed to extract job script")
	}

	dest := "build.sh"
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
	go br.streamer.Stream(ctx, job, "stdouterr", pr)

	cmd := exec.CommandContext(ctx, "bash", dest)
	cmd.Stdout = pw
	cmd.Stderr = pw

	log.Debug("staring command")
	err = cmd.Start()
	if err != nil {
		log.WithError(err).Error("failed to start command")
		_ = br.statuser.Status(ctx, job, "started", "failed")
		return errors.Wrap(err, "failed to start command")
	}

	err = cmd.Wait()
	if err != nil {
		log.WithError(err).Error("command wait errored")

		if cmd.ProcessState == nil {
			_ = br.statuser.Status(ctx, job, "running", "errored")
			return errors.Wrap(err, "no process state found")
		}

		if !cmd.ProcessState.Exited() {
			_ = br.statuser.Status(ctx, job, "running", "errored")
			return errors.Wrap(err, "process did not exit")
		}

		if !cmd.ProcessState.Success() {
			_ = br.statuser.Status(ctx, job, "running", "failed")
			return errors.Wrap(err, "process exited without success")
		}

		return err
	}

	log.Debug("command completed")
	return nil
}
