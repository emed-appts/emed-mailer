package template

import (
	_ "embed"
	"html/template"
	"io"
	"time"

	"github.com/pkg/errors"
)

var (
	//go:embed changedappts.tmpl
	msgTemplate string

	funcMap = template.FuncMap{
		"DateFmt": func(t time.Time) string {
			return t.Format("02.01.2006 15:04")
		},
	}
)

// Execute executes the named template
func Execute(wr io.Writer, name string, data interface{}) error {
	t, err := template.New("message").Funcs(funcMap).Parse(msgTemplate)
	if err != nil {
		return errors.Wrap(err, "could not parse template")
	}

	err = t.Execute(wr, data)
	return errors.Wrap(err, "could not execute template")
}
