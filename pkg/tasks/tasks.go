package tasks

import (
	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/wolfeidau/vlad/pkg/tasks/cfn"
	"github.com/wolfeidau/vlad/pkg/vlad"
)

// Task task which is run
type Task interface {
	Execute(ctx *vlad.Context) error
	GetName() string
	Validate() error
}

var (
	_ = Task(&cfn.CloudformationTask{}) // assert that CloudformationTask matches the Task interface
)

func Setup() error {
	sess := session.New(&aws.Config{
		Logger: aws.LoggerFunc(func(args ...interface{}) {
			logrus.Debug(args...)
		}),
		// LogLevel: aws.LogLevel(aws.LogDebug),
	},
	)

	cfn.SetCFNAPI(cloudformation.New(sess))

	return nil
}

func Cloudformation(name string, params *cfn.CloudformationParams) *cfn.CloudformationTask {
	return cfn.New(name, params)
}
