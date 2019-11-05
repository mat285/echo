package core

import (
	"bytes"

	"github.com/blend/go-sdk/exception"
	"github.com/blend/go-sdk/template"
)

var (
	// TemplateRootPath is the root directory where to find templates
	TemplateRootPath = "_templates"
)

// Template is a thin wrapper function around a common template use case
func Template(path string, vars map[string]interface{}) (*bytes.Buffer, error) {
	return TemplateWithDelimiters(path, vars, "", "")
}

// TemplateWithDelimiters is similar to `Template` but with possible custom delimiters
func TemplateWithDelimiters(path string, vars map[string]interface{}, left, right string) (*bytes.Buffer, error) {
	t, err := template.NewFromFile(path)
	if err != nil {
		return nil, exception.New(err)
	}

	b := bytes.NewBuffer(nil)
	if err := t.WithVars(vars).WithDelims(left, right).Process(b); err != nil {
		return nil, exception.New(err)
	}
	return b, nil
}
