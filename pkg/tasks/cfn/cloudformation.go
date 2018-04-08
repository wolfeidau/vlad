package cfn

import (
	"io/ioutil"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/wolfeidau/vlad/pkg/vlad"
)

var (
	cfnAPI cloudformationiface.CloudFormationAPI

	defaultCapabilities = []string{"CAPABILITY_IAM", "CAPABILITY_NAMED_IAM"}
)

func SetCFNAPI(api cloudformationiface.CloudFormationAPI) {
	cfnAPI = api
}

// CloudformationTask cloudformation task
type CloudformationTask struct {
	Name   string
	Params *CloudformationParams

	// results
	action  string
	results map[string]interface{}
}

// CloudformationParams parameters which were extracted from the runbook
type CloudformationParams struct {
	StackName          string             `mapstructure:"stack_name"`
	Template           string             `mapstructure:"template"`
	NotificationArns   []string           `mapstructure:"notification_arns"`
	DisableRollback    bool               `mapstructure:"disable_rollback"`
	TemplateParameters map[string]*string `mapstructure:"template_parameters"`
	Tags               map[string]*string `mapstructure:"tags"`
}

// New cloudformation task
func New(name string, params *CloudformationParams) *CloudformationTask {
	return &CloudformationTask{
		Name:    name,
		Params:  params,
		results: make(map[string]interface{}),
	}
}

// Execute execute template
func (ct *CloudformationTask) Execute(ctx *vlad.Context) error {

	templatePath := path.Join(ctx.BasePath, ct.Params.Template)

	templateBody, err := ioutil.ReadFile(templatePath)
	if err != nil {
		return errors.Wrapf(err, "failed to load template %s", templatePath)
	}

	created, err := ct.createStack(templateBody)
	if err != nil {
		return errors.Wrap(err, "failed to create stack")
	}

	if !created {

		changed, err := ct.updateStack(templateBody)
		if err != nil {
			return errors.Wrap(err, "failed to update stack")
		}
		// no error and it wasn't changed
		if !changed {
			logrus.Infof("%s stack is already up to date", ct.Name)
			return nil
		}

	}

	ct.emitResults()

	return nil
}

// Validate validate
func (ct *CloudformationTask) Validate() error {
	return nil
}

// GetName get name
func (ct *CloudformationTask) GetName() string {
	return ct.Name
}

func (ct *CloudformationTask) createStack(templateBody []byte) (bool, error) {
	res, err := cfnAPI.CreateStack(&cloudformation.CreateStackInput{
		StackName:        aws.String(ct.Params.StackName),
		TemplateBody:     aws.String(string(templateBody)),
		NotificationARNs: aws.StringSlice(ct.Params.NotificationArns),
		Parameters:       awsParameters(ct.Params.TemplateParameters),
		DisableRollback:  aws.Bool(ct.Params.DisableRollback),
		Capabilities:     aws.StringSlice(defaultCapabilities),
		Tags:             awsTags(ct.Params.Tags),
	})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "AlreadyExistsException" {
				return false, nil
			}

			return false, errors.Wrap(err, "failed to create stack")
		}
	}

	ct.action = "created"
	ct.results["StackID"] = aws.StringValue(res.StackId)

	return true, nil
}

func (ct *CloudformationTask) updateStack(templateBody []byte) (bool, error) {
	res, err := cfnAPI.UpdateStack(&cloudformation.UpdateStackInput{
		StackName:        aws.String(ct.Params.StackName),
		TemplateBody:     aws.String(string(templateBody)),
		NotificationARNs: aws.StringSlice(ct.Params.NotificationArns),
		Parameters:       awsParameters(ct.Params.TemplateParameters),
		Capabilities:     aws.StringSlice(defaultCapabilities),
		Tags:             awsTags(ct.Params.Tags),
	})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "ValidationError" && strings.Contains(awsErr.Message(), "No updates are to be performed.") {
				return false, nil
			}
		}
		return false, errors.Wrap(err, "failed to update stack")
	}

	ct.action = "updated"
	ct.results["StackID"] = aws.StringValue(res.StackId)

	return true, nil
}

func (ct *CloudformationTask) emitResults() {
	logrus.WithFields(logrus.Fields{
		"StackName": ct.Params.StackName,
		"StackID":   ct.results["StackId"],
	}).Infof("%s stack %s", ct.Name, ct.action)
}

func awsTags(tags map[string]*string) []*cloudformation.Tag {

	cfntags := []*cloudformation.Tag{}

	for k, v := range tags {
		tag := &cloudformation.Tag{
			Key:   aws.String(k),
			Value: v,
		}

		cfntags = append(cfntags, tag)
	}

	return cfntags
}

func awsParameters(tags map[string]*string) []*cloudformation.Parameter {

	cfnParams := []*cloudformation.Parameter{}

	for k, v := range tags {
		tag := &cloudformation.Parameter{
			ParameterKey:   aws.String(k),
			ParameterValue: v,
		}

		cfnParams = append(cfnParams, tag)
	}

	return cfnParams
}
