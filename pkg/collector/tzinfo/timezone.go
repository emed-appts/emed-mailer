// +build !windows

package tzinfo

import "time"

// LoadLocation returns the location for unix platforms
func LoadLocation(name string) (*time.Location, error) {
	return time.LoadLocation(name)
}