package expander

import (
	"bytes"
	"log"
	"text/template"
)

func ExpandArguments(args string, lookup map[string]string) string {
	tmpl, err := template.New("args").Parse(args)
	if err != nil {
		log.Println("Error parsing args for template:", err)
		return args
	}

	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, lookup); err != nil {
		log.Println("Error executing template:", err)
		return args
	}

	return buf.String()
}
