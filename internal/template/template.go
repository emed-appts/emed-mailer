//go:generate retool do packr

package template

import (
	"html/template"
	"io"
	"time"

	"github.com/gobuffalo/packr"
	"github.com/pkg/errors"
)

var (
	box = packr.NewBox("../../templates")

	funcMap = template.FuncMap{
		"DateFmt": func(t time.Time) string {
			return t.Format("02.01.2006 15:04")
		},
	}
)

// Execute executes the named template
func Execute(wr io.Writer, name string, data interface{}) error {
	msgTemplate, err := box.MustString(name)
	if err != nil {
		return errors.Wrap(err, "could not open template file")
	}

	t, err := template.New("message").Funcs(funcMap).Parse(string(msgTemplate))
	if err != nil {
		return errors.Wrap(err, "could not parse template")
	}

	err = t.Execute(wr, data)
	return errors.Wrap(err, "could not execute template")
}
