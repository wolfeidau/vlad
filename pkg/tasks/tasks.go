package tasks

import (
	"fmt"
	"reflect"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/wolfeidau/vlad/pkg/runbook"
	"github.com/wolfeidau/vlad/pkg/tasks/cfn"
	"github.com/wolfeidau/vlad/pkg/vlad"
)

// Task task which is run
type Task interface {
	Execute(ctx *vlad.Context) error
	GetName() string
	Validate() error
}

// Engine task creation and executing is done in here
type Engine struct {
	sess        *session.Session
	tasksToExec []Task
}

// NewEngine create a new engine and configure the AWS session
func NewEngine() *Engine {
	sess := session.New(&aws.Config{
		Logger: aws.LoggerFunc(func(args ...interface{}) {
			logrus.Debug(args...)
		}),
		// LogLevel: aws.LogLevel(aws.LogDebug),
	})

	return &Engine{
		sess:        sess,
		tasksToExec: []Task{},
	}
}

// Build build the tasks from the runbook
func (e *Engine) Build(runbook *runbook.RunBook) error {

	var err error

	for n, v := range runbook.Tasks {
		for k, v := range v {
			switch k {
			case "cloudformation":
				name := fmt.Sprintf("task[%d]cloudformation", n)

				cfnParams := new(cfn.Params)

				err = mapstructure.Decode(v, cfnParams)
				if err != nil {
					// logrus.Fatalf("failed to decode task: %+v", err)
					return errors.Wrapf(err, "failed to decode task: %s", name)
				}

				val := reflect.ValueOf(cfnParams)

				logrus.Debugf("%s running templates to update vars", name)

				err = runbook.VisitRecursive(name, &val)
				if err != nil {
					return errors.Wrapf(err, "failed to update template: %s", name)
				}

				err := e.cloudformation(name, cfnParams)
				if err != nil {
					return errors.Wrapf(err, "failed to create cfn task: %s", name)
				}
			}
		}
	}

	return nil
}

// Run run the current list of tasks
func (e *Engine) Run(ctx *vlad.Context) error {
	var err error

	for _, v := range e.tasksToExec {
		err = v.Execute(ctx)
		if err != nil {
			// logrus.Fatalf("%s failed to update template: %+v", v.GetName(), err)
			return errors.Wrapf(err, "failed to execute task: %s", v.GetName())
		}
	}

	return nil
}

// cloudformation create a cloudformation task
func (e *Engine) cloudformation(name string, params *cfn.Params) error {
	e.tasksToExec = append(e.tasksToExec,
		cfn.New(name, params, e.sess))
	return nil
}
