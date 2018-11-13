package util

import (
	"bytes"
	"text/template"
	"encoding/json"
)

var templateFuncs = template.FuncMap{
	"newline": func() (string) {
		return "\n"
	},
	"json": func(d interface{}) (string) {
		payload, err := json.Marshal(d)
		if err != nil {
			return ""
		}

		return string(payload[:])
	},
	"json_indent": func(d interface{}) (string) {
		payload, err := json.MarshalIndent(d, "", "  ")
		if err != nil {
			return ""
		}

		return string(payload[:])
	},
}

func NewTemplate() (tmpl *template.Template) {
	return template.New("root").Funcs(templateFuncs)
}

func TemplateToBytes(tmpl *template.Template, v interface{}) (text []byte, err error) {
	buffer := bytes.NewBuffer(nil)
	defer buffer.Reset()

	err = tmpl.Execute(buffer, v)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func TemplateToString(tmpl *template.Template, v interface{}) (text string, err error) {
	data, err := TemplateToBytes(tmpl, v)
	if err != nil {
		return "", err
	}

	return string(data[:]), nil
}
