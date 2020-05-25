package github

import (
	"bytes"
	"text/template"
)

func toMarkdown(templateString string, obj interface{}) (string, error) {
	template, err := getTemplate(templateString)
	if err != nil {
		return "", err
	}
	buf := bytes.NewBufferString("")
	template.Execute(buf, obj)
	return buf.String(), nil
}

func getTemplate(s string) (*template.Template, error) {
	return template.New("tmp").Parse(s)
}
