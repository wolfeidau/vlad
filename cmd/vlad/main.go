package main

import (
	"fmt"
	"log"
	"path"
	"reflect"

	"github.com/alecthomas/kingpin"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/wolfeidau/vlad/pkg/runbook"
	"github.com/wolfeidau/vlad/pkg/tasks"
	"github.com/wolfeidau/vlad/pkg/vlad"
)

var (
	version       = "dev"
	verbose       = kingpin.Flag("verbose", "Verbose mode.").Short('v').Bool()
	runbookFile   = kingpin.Flag("runbook", "Runbook file to load.").Default("runbook.yml").String()
	varParameters = kingpin.Flag("params-prefix", "AWS Parameter store prefix to load vars from.").String()
)

func main() {
	kingpin.Version(version)
	kingpin.Parse()

	if *verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}

	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	logrus.Debugf("running version: %s", version)

	runbook, err := runbook.LoadFromFile(*runbookFile)
	if err != nil {
		log.Fatalf("failed to load runbook file: %+v", err)
	}

	err = tasks.Setup()
	if err != nil {
		logrus.Fatalf("failed to setup tasks: %+v", err)
	}

	// spew.Dump(config)

	tasksToExec := []tasks.Task{}

	for n, v := range runbook.Tasks {
		for k, v := range v {
			switch k {
			case "cloudformation":
				name := fmt.Sprintf("task[%d]cloudformation", n)

				cfnTask := tasks.Cloudformation(name)
				err = mapstructure.Decode(v, cfnTask)
				if err != nil {
					logrus.Errorf("failed to decode task: %+v", err)
				}

				val := reflect.ValueOf(cfnTask)

				logrus.Debugf("%s running templates to update vars", name)

				err = runbook.VisitRecursive(name, &val)
				if err != nil {
					logrus.Fatalf("%s failed to update template: %+v", name, err)
				}

				tasksToExec = append(tasksToExec, cfnTask)
			}
		}
	}

	ctx := &vlad.Context{
		BasePath: path.Dir(*runbookFile),
		Keys:     make(map[string]interface{}),
	}

	for _, v := range tasksToExec {
		err = v.Execute(ctx)
		if err != nil {
			logrus.Fatalf("%s failed to update template: %+v", v.GetName(), err)
		}
	}

}
