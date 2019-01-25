package main

import (
	"os"

	"github.com/travis-ci/proc"
	cli "gopkg.in/urfave/cli.v2"
)

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "debug",
				Value: true,
				Usage: "enable debug logging and ensure TRAVIS_DEBUG is set in all subshells",
			},
		},
		Commands: []*cli.Command{
			{
				Name:   "wait",
				Usage:  "wait for and then run a job",
				Action: proc.Wait,
			},
		},
	}
	app.Run(os.Args)
}
