package main

import (
	"os"

	"github.com/travis-ci/job"
)

func main() {
	app := job.NewApp()
	app.Run(os.Args)
}
