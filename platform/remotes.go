package platform

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/bravetools/bravetools/shared"
)

// Remote represents a configuration of the remote
type Remote struct {
	Name       string `json:"name"`
	URL        string `json:"url"`
	Protocol   string `json:"protocol"`
	Public     bool   `json:"public"`
	Profile    string `json:"profile"`
	key        string
	cert       string
	servercert string
}

func NewBravehostRemote(settings BackendSettings, profileName string) Remote {
	var protocol string
	var url string

	switch settings.Type {
	case "lxd":
		protocol = "unix"
		url = "/var/snap/lxd/common/lxd/unix.socket"
	default:
		protocol = "lxd"
		url = "https://" + settings.Resources.IP + ":8443"
	}

	return Remote{
		Name:     shared.BravetoolsRemote,
		URL:      url,
		Protocol: protocol,
		Public:   false,
		Profile:  profileName,
	}
}

// ParseRemoteName unpacks remote and rest of image/service name and returns both
func ParseRemoteName(image string) (remote string, imageName string) {
	split := strings.SplitN(image, ":", 2)
	if len(split) == 1 {
		// Default remote
		return shared.BravetoolsRemote, split[0]
	}

	return split[0], split[1]
}

// loadRemoteConfig loads a saved bravetools remote config
func loadRemoteConfig(name string) (remote Remote, err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return remote, err
	}

	path := path.Join(home, shared.BraveRemoteStore, name+".json")

	var fileBytes bytes.Buffer
	f, err := os.Open(path)
	if err != nil {
		return remote, err
	}
	defer f.Close()

	_, err = io.Copy(&fileBytes, f)
	if err != nil {
		return remote, err
	}

	err = json.Unmarshal(fileBytes.Bytes(), &remote)
	return remote, err
}

// LoadRemoteSettings loads a saved bravetools remote with TLS auth certs/keys if present
func LoadRemoteSettings(remoteName string) (Remote, error) {

	userHome, err := os.UserHomeDir()
	if err != nil {
		return Remote{}, err
	}

	remote, err := loadRemoteConfig(remoteName)
	if err != nil {
		return Remote{}, err
	}

	// unix socket doesn't need any auth
	if remote.Protocol == "unix" {
		return remote, nil
	}

	// Load remote server cert for verification
	serverCertPath := path.Join(userHome, shared.BraveServerCertStore, remoteName+".crt")
	remote.servercert, _ = loadServerCert(serverCertPath)

	// Public Image server doesn't need client auth
	if remote.Public || remote.Protocol == "simplestreams" {
		return remote, nil
	}

	// Add client cert and key
	keyPath := path.Join(userHome, shared.BraveClientKey)
	certPath := path.Join(userHome, shared.BraveClientCert)

	remote.key, _ = loadKey(keyPath)
	remote.cert, _ = loadCert(certPath)

	return remote, nil
}

func SaveRemote(remote Remote) error {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to save remote %q: %s", remote.Name, err.Error())
	}

	path := path.Join(userHome, shared.BraveRemoteStore, remote.Name+".json")
	remoteJson, err := json.MarshalIndent(remote, "", "    ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, remoteJson, 0666)
}

func loadKey(path string) (string, error) {
	buf, err := shared.ReadFile(path)
	if err != nil {
		return "", errors.New("cannot load client key")
	}
	key := buf.String()
	return key, nil
}

func loadCert(path string) (string, error) {
	buf, err := shared.ReadFile(path)
	if err != nil {
		return "", errors.New("cannot load client certificate")
	}
	cert := buf.String()
	return cert, nil
}

func loadServerCert(path string) (string, error) {
	buf, err := shared.ReadFile(path)
	if err != nil {
		return "", errors.New("cannot load server certificate")
	}
	cert := buf.String()
	return cert, nil
}
