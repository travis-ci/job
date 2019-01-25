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

var (
	maxLifetimeFlag = &cli.DurationFlag{
		Name:    "max-lifetime",
		Value:   5 * 60 * time.Minute,
		Usage:   "max amount of time to wait before imploding",
		EnvVars: []string{"MAX_LIFETIME", "TRAVIS_JOB_MAX_LIFETIME"},
	}
)

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "debug",
				Value:   false,
				Usage:   "enable debug logging and ensure TRAVIS_DEBUG is set in all subshells",
				EnvVars: []string{"DEBUG", "TRAVIS_JOB_DEBUG"},
			},
			&cli.StringFlag{
				Name:    "health-url",
				Usage:   "url for runtime health",
				EnvVars: []string{"HEALTH_URL", "TRAVIS_JOB_HEALTH_URL"},
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
						EnvVars: []string{"JOB_URL", "TRAVIS_JOB_URL"},
					},
					maxLifetimeFlag,
					&cli.DurationFlag{
						Name:    "max-wait-time",
						Value:   30 * time.Minute,
						Usage:   "max amount of time to wait before imploding",
						EnvVars: []string{"MAX_WAIT_TIME", "TRAVIS_JOB_MAX_WAIT_TIME"},
					},
					&cli.DurationFlag{
						Name:    "wait-interval",
						Value:   3 * time.Second,
						Usage:   "interval to sleep between attempts",
						EnvVars: []string{"WAIT_INTERVAL", "TRAVIS_JOB_WAIT_INTERVAL"},
					},
				},
				Action: func(c *cli.Context) error {
					ctx, cancel := context.WithTimeout(
						context.Background(), c.Duration("max-lifetime"))

					defer cancel()

					log := setupLogger(c.Bool("debug"))

					src, err := job.NewSource(log, c.String("url"))
					if err != nil {
						return cli.Exit(fmt.Sprintf("failed to create job source: %v", err), 2)
					}

					runner, err := job.NewRunner(log)
					if err != nil {
						return cli.Exit(fmt.Sprintf("failed to create job runner: %v", err), 3)
					}

					w := job.NewWaiter(log, c.Duration("wait-interval"),
						c.Duration("max-wait-time"), src, runner)

					err = w.Wait(ctx)

					if err != nil {
						return cli.Exit(fmt.Sprintf("failed during wait for job: %v", err), 4)
					}

					return nil
				},
			},
			{
				Name:  "run",
				Usage: "run a job via json input",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "json",
						Usage:   "json input file to run",
						EnvVars: []string{"JSON", "TRAVIS_JOB_JSON"},
					},
					maxLifetimeFlag,
				},
				Action: func(c *cli.Context) error {
					ctx, cancel := context.WithTimeout(
						context.Background(), c.Duration("max-lifetime"))

					defer cancel()

					log := setupLogger(c.Bool("debug"))

					src, err := job.NewSource(log, fmt.Sprintf("file://%s", c.String("json")))
					if err != nil {
						return cli.Exit(fmt.Sprintf("failed to create job source: %v", err), 2)
					}

					runner, err := job.NewRunner(log)
					if err != nil {
						return cli.Exit(fmt.Sprintf("failed to create job runner: %v", err), 3)
					}

					job, err := src.Fetch(ctx)
					if err != nil {
						return cli.Exit(fmt.Sprintf("failed to fetch job: %v", err), 4)
					}

					err = runner.Run(ctx, job)
					if err != nil {
						return cli.Exit(fmt.Sprintf("failed to run job: %v", err), 5)
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
