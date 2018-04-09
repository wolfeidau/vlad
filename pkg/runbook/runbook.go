package runbook

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

// name of the tags used in local structs
const tagName = "runbook"

// RunBook the configuration which vars and a list of tasks to execute
type RunBook struct {
	Vars  map[string]interface{}   `yaml:"vars"`
	Tasks []map[string]interface{} `yaml:"tasks"`
}

// LoadFromFile load a runbook from the supplied file path
func LoadFromFile(path string) (*RunBook, error) {

	logrus.Debugf("loading runbook from path: %s", path)

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read file")
	}

	runbook := new(RunBook)

	err = yaml.Unmarshal(data, runbook)
	if err != nil {
		return nil, errors.Wrap(err, "failed to Unmarshal file")
	}

	return runbook, nil
}

// VisitRecursive visit all the properties of the runbook and update any template strings
func (rb *RunBook) VisitRecursive(name string, obj *reflect.Value) error {

	switch obj.Kind() {

	case reflect.Ptr:
		originalValue := obj.Elem()
		// Check if the pointer is nil
		if !originalValue.IsValid() {
			return nil
		}
		rb.VisitRecursive(name, &originalValue)
	case reflect.Interface:
		rb.VisitRecursive(name, obj)
	case reflect.Struct:

		structType := obj.Type()

		for i := 0; i < obj.NumField(); i++ {
			fieldName := structType.Field(i).Name

			// load the tag value and skip visiting this structure
			tagValue := structType.Field(i).Tag.Get(tagName)
			tagValue = strings.SplitN(tagValue, ",", 2)[0]
			if tagValue == "-" {
				logrus.Debugf("%s skipping struct: %s", name, fieldName)
				return nil
			}

			if name != "" {
				fieldName = fmt.Sprintf("%s.%s", name, fieldName)
			}
			field := obj.Field(i)
			rb.VisitRecursive(fieldName, &field)
		}
	case reflect.Slice:
		for i := 0; i < obj.Len(); i++ {
			idx := obj.Index(i)
			rb.VisitRecursive(name, &idx)
		}
	case reflect.Map:
		for _, key := range obj.MapKeys() {
			objValue := obj.MapIndex(key)
			fieldName := fmt.Sprintf("%s[%s]", name, key)
			rb.VisitRecursive(fieldName, &objValue)
		}
	case reflect.String:
		objString := obj.Interface().(string)

		// does it contain golang template characters
		if !(strings.Contains(objString, "{{") && strings.Contains(objString, "}}")) {
			return nil
		}

		logrus.Debugf("%s running template: %s", name, objString)

		t, err := template.New(name).Parse(objString)
		if err != nil {
			return errors.Wrap(err, "failed to Parse template")
		}

		buf := &bytes.Buffer{}
		err = t.Execute(buf, rb.Vars)
		if err != nil {
			return errors.Wrap(err, "failed to Execute template")
		}

		logrus.Debugf("%s, %s", name, buf.String())
		if obj.CanSet() {
			obj.SetString(buf.String())
		}

	}

	return nil
}
