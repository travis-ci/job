package job

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	cli "gopkg.in/urfave/cli.v2"
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

	src, err := NewSource(log, c.String("url"))
	if err != nil {
		return cli.Exit(fmt.Sprintf("failed to create job source: %v", err), 2)
	}

	runner, err := NewRunner(log)
	if err != nil {
		return cli.Exit(fmt.Sprintf("failed to create job runner: %v", err), 3)
	}

	w := NewWaiter(log, c.Duration("wait-interval"),
		c.Duration("max-wait-time"), src, runner)

	err = w.Wait(ctx)

	if err != nil {
		return cli.Exit(fmt.Sprintf("failed during wait for job: %v", err), 4)
	}

	return nil
}

func runCommandAction(c *cli.Context) error {
	ctx, cancel := context.WithTimeout(
		context.Background(), c.Duration("max-lifetime"))

	defer cancel()

	log := setupLogger(c.Bool("debug"))

	src, err := NewSource(log, fmt.Sprintf("file://%s", c.String("json")))
	if err != nil {
		return cli.Exit(fmt.Sprintf("failed to create job source: %v", err), 2)
	}

	runner, err := NewRunner(log)
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
