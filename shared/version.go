package shared

import (
	"fmt"
)

var Version string

// VersionString prints Bravetools version
func VersionString() string {
	return fmt.Sprintf("Bravetools Version: %s", Version)
}
