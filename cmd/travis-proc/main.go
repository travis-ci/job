package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/travis-ci/job"
	cli "gopkg.in/urfave/cli.v2"
)

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "debug",
				Value:   false,
				Usage:   "enable debug logging and ensure TRAVIS_DEBUG is set in all subshells",
				EnvVars: []string{"DEBUG", "TRAVIS_PROC_DEBUG"},
			},
			&cli.StringFlag{
				Name:    "health-url",
				Usage:   "url to poll during runtime to report health of job",
				EnvVars: []string{"HEALTH_URL", "TRAVIS_PROC_HEALTH_URL"},
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "wait",
				Usage: "wait for and then run a job",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "job-url",
						Usage:   "url to poll for a job",
						EnvVars: []string{"JOB_URL", "TRAVIS_PROC_JOB_URL"},
					},
					&cli.DurationFlag{
						Name:    "max-poll-time",
						Value:   30 * time.Minute,
						Usage:   "max amount of time to poll before imploding",
						EnvVars: []string{"MAX_POLL_TIME", "TRAVIS_PROC_MAX_POLL_TIME"},
					},
					&cli.DurationFlag{
						Name:    "wait-interval",
						Value:   3 * time.Second,
						Usage:   "interval to sleep between waiting polls",
						EnvVars: []string{"WAIT_INTERVAL", "TRAVIS_PROC_WAIT_INTERVAL"},
					},
				},
				Action: func(c *cli.Context) error {
					ctx, cancel := context.WithTimeout(
						context.Background(), c.Duration("max-poll-time"))

					defer cancel()

					log := setupLogger(c.Bool("debug"))

					src, err := job.NewSource(log, c.String("job-url"))
					if err != nil {
						return cli.Exit(fmt.Sprintf("failed to create job source: %v", err), 2)
					}

					runner, err := job.NewRunner(log)
					if err != nil {
						return cli.Exit(fmt.Sprintf("failed to create job runner: %v", err), 3)
					}

					w := job.NewWaiter(log, c.Duration("wait-interval"), src, runner)

					err = w.Wait(ctx)

					if err != nil {
						return cli.Exit(fmt.Sprintf("failed to wait for job: %v", err), 4)
					}

					return nil
				},
			},
		},
	}
	app.Run(os.Args)
}

func setupLogger(debug bool) logrus.FieldLogger {
	log := logrus.New()
	if debug {
		log.Level = logrus.DebugLevel
	}

	return log
}
