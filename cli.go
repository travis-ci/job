package job

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	cli "gopkg.in/urfave/cli.v2"
)

var (
	osHostname, _ = os.Hostname()
)

func NewApp() *cli.App {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "debug",
				Value:   false,
				Usage:   "enable debug logging and ensure TRAVIS_DEBUG is set in all subshells",
				EnvVars: envVars("DEBUG"),
			},
			&cli.StringFlag{
				Name:    "health-url",
				Usage:   "url for runtime health",
				EnvVars: envVars("HEALTH_URL"),
			},
			&cli.DurationFlag{
				Name:    "max-lifetime",
				Value:   5 * 60 * time.Minute,
				Usage:   "max amount of time to wait before imploding",
				EnvVars: envVars("MAX_LIFETIME"),
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "wait",
				Usage: "wait for an available job and then run it",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "url",
						Usage:   "url for a job",
						EnvVars: envVars("JOB_URL"),
					},
					&cli.DurationFlag{
						Name:    "max-wait-time",
						Value:   30 * time.Minute,
						Usage:   "max amount of time to wait before imploding",
						EnvVars: envVars("MAX_WAIT_TIME"),
					},
					&cli.DurationFlag{
						Name:    "wait-interval",
						Value:   3 * time.Second,
						Usage:   "interval to sleep between attempts",
						EnvVars: envVars("WAIT_INTERVAL"),
					},
				},
				Action: waitCommandAction,
			},
			{
				Name:  "run",
				Usage: "run a job via json input",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "json",
						Usage:   "json input file to run",
						EnvVars: envVars("JSON"),
					},
				},
				Action: runCommandAction,
			},
		},
	}
	return app
}

func waitCommandAction(c *cli.Context) error {
	ctx, cancel := context.WithTimeout(
		context.Background(), c.Duration("max-lifetime"))

	defer cancel()

	log := setupLogger(c.Bool("debug"))
	processorID, err := buildProcessorID()
	if err != nil {
		return cli.Exit(fmt.Sprintf("failed to build processor ID: %v", err), 2)
	}

	src, err := NewSource(log, c.String("url"), processorID)
	if err != nil {
		return cli.Exit(fmt.Sprintf("failed to create job source: %v", err), 2)
	}

	runner, err := NewRunner(log, NewStatuser(log), NewStreamer(log))
	if err != nil {
		return cli.Exit(fmt.Sprintf("failed to create job runner: %v", err), 2)
	}

	w := NewWaiter(log, c.Duration("wait-interval"),
		c.Duration("max-wait-time"), src, runner)

	err = w.Wait(ctx)

	if err != nil {
		return cli.Exit(fmt.Sprintf("failed during wait for job: %v", err), 2)
	}

	return nil
}

func runCommandAction(c *cli.Context) error {
	ctx, cancel := context.WithTimeout(
		context.Background(), c.Duration("max-lifetime"))

	defer cancel()

	log := setupLogger(c.Bool("debug"))
	processorID, err := buildProcessorID()
	if err != nil {
		return cli.Exit(fmt.Sprintf("failed to build processor ID: %v", err), 2)
	}

	src, err := NewSource(log, fmt.Sprintf("file://%s", c.String("json")), processorID)
	if err != nil {
		return cli.Exit(fmt.Sprintf("failed to create job source: %v", err), 2)
	}

	log.Debug("creating job runner")
	runner, err := NewRunner(log, NewStatuser(log), NewStreamer(log))
	if err != nil {
		return cli.Exit(fmt.Sprintf("failed to create job runner: %v", err), 2)
	}

	log.Debug("fetching job")
	job, err := src.Fetch(ctx)
	if err != nil {
		return cli.Exit(fmt.Sprintf("failed to fetch job: %v", err), 2)
	}

	log.WithField("job_id", job.ID()).Debug("running job")
	err = runner.Run(ctx, job)
	if err != nil {
		return cli.Exit(fmt.Sprintf("failed to run job: %v", err), 2)
	}

	return nil
}

func setupLogger(debug bool) logrus.FieldLogger {
	log := logrus.New()
	if debug {
		log.Level = logrus.DebugLevel
	}

	return log
}

func envVars(key string) []string {
	key = strings.ToUpper(key)
	return []string{
		key,
		fmt.Sprintf("TRAVIS_JOB_%s", key),
	}
}

func buildProcessorID() (string, error) {
	hostname := osHostname
	if hostname == "" {
		hostname = "localhost"
	}
	procUUID, err := uuid.NewRandom()
	if err != nil {
		return "", errors.Wrap(err, "failed to generate uuid")
	}
	return fmt.Sprintf("%s@%d.%s", procUUID.String(), os.Getpid(), hostname), nil
}
