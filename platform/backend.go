package platform

import "fmt"

// Backend ..
type Backend interface {
	BraveBackendInit() error
	Info() (Info, error)
}

// Info describes Brave Platform
type Info struct {
	ImageStorage  string
	VolumeStorage string
	Name          string
	State         string
	IPv4          string
	Release       string
	ImageHash     string
	Load          string
	Disk          []string
	Memory        []string
	CPU           string
}

// NewHostBackend returns a new Backend from provided host Settings
func NewHostBackend(host BraveHost) (backend Backend, err error) {
	backendType := host.Settings.BackendSettings.Type

	switch backendType {
	case "multipass":
		backend = NewMultipass(host.Settings)
	case "lxd":
		backend = NewLxd(host.Settings)
	default:
		err = fmt.Errorf("backend type %q not supported", backendType)
	}
	return backend, err
}
