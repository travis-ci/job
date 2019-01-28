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
	script, err := job.Script()
	if err != nil {
		return errors.Wrap(err, "failed to extract job script")
	}

	dest := "build.sh"
	err = ioutil.WriteFile(dest, []byte(script), os.FileMode(0755))
	if err != nil {
		return errors.Wrap(err, "failed to write job script")
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	pr, pw := io.Pipe()
	go br.streamer.Stream(ctx, job, "stdouterr", pr)

	cmd := exec.CommandContext(ctx, "bash", dest)
	cmd.Stdout = pw
	cmd.Stderr = pw

	err = cmd.Start()
	if err != nil {
		statusErr := br.statuser.Status(ctx, job, "started", "failed")
		if statusErr != nil {
			err = errors.Wrap(err, statusErr.Error())
		}
		return errors.Wrap(err, "failed to start command")
	}

	err = cmd.Wait()
	if err != nil {
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

	return nil
}
