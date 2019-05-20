package tzinfo

import (
	"time"

	"4d63.com/tz"
)

// LoadLocation returns the location for windows platform
// https://github.com/golang/go/issues/21881
func LoadLocation(name string) (*time.Location, error) {
	return tz.LoadLocation(name)
}