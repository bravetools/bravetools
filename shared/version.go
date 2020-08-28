package shared

import (
	"fmt"
	"path/filepath"
	"runtime"
)

var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(b)
)

// VersionString prints Bravetools version
func VersionString() string {
	buf, _ := ReadFile(filepath.Dir(basepath) + "/VERSION")

	return fmt.Sprintf("Bravetools Version: %s", buf.String())
}
