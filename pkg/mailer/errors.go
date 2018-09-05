package mailer

import "github.com/pkg/errors"

type notRunning interface {
	NotRunning() bool
}

func IsNotRunning(err error) bool {
	nr, ok := errors.Cause(err).(notRunning)
	return ok && nr.NotRunning()
}

type notRunningError struct{}

func newNotRunningError() error {
	return &notRunningError{}
}

func (err *notRunningError) Error() string {
	return "mailer daemon is not running"
}

func (err *notRunningError) NotRunning() bool {
	return true
}

type alreadyRunning interface {
	AlreadyRunning() bool
}

func IsAlreadyRunning(err error) bool {
	ar, ok := errors.Cause(err).(alreadyRunning)
	return ok && ar.AlreadyRunning()
}

type alreadyRunningError struct{}

func newAlreadyRunningError() error {
	return &alreadyRunningError{}
}

func (err *alreadyRunningError) Error() string {
	return "mailer daemon is already running"
}

func (err *alreadyRunningError) AlreadyRunning() bool {
	return true
}
