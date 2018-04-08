package main

import (
	"log"
	"os"
	"path"

	"github.com/alecthomas/kingpin"
	"github.com/sirupsen/logrus"
	"github.com/wolfeidau/vlad/pkg/runbook"
	"github.com/wolfeidau/vlad/pkg/tasks"
	"github.com/wolfeidau/vlad/pkg/vlad"
)

var (
	version = "dev"

	app = kingpin.New("vlad", "A command-line deployment application.")

	verbose = app.Flag("verbose", "Verbose mode.").Short('v').Bool()

	exec        = app.Command("exec", "Execute a runbook.")
	runbookFile = exec.Flag("runbook", "Runbook file to load.").Default("runbook.yml").String()
	// varParameters = exec.Flag("params-prefix", "AWS Parameter store prefix to load vars from.").String()
)

func main() {
	kingpin.Version(version)

	command := kingpin.MustParse(app.Parse(os.Args[1:]))

	if *verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}

	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	logrus.Debugf("running version: %s", version)

	// kingpin.Parse()
	switch command {
	case exec.FullCommand():
		err := execute(*runbookFile)
		if err != nil {
			log.Fatalf("failed to execute runbook: %+v", err)
		}
	}

}

func execute(runbookPath string) error {
	runbook, err := runbook.LoadFromFile(runbookPath)
	if err != nil {
		// log.Fatalf("failed to load runbook file: %+v", err)
		return err
	}

	engine := tasks.NewEngine()

	// parse and load runbook
	err = engine.Build(runbook)
	if err != nil {
		// log.Fatalf("failed to build tasks: %+v", err)
		return err
	}

	ctx := &vlad.Context{
		BasePath: path.Dir(*runbookFile),
		Keys:     make(map[string]interface{}),
	}

	// exec tasks
	err = engine.Run(ctx)
	if err != nil {
		// log.Fatalf("failed to run tasks: %+v", err)
		return err
	}

	return nil
}
