package platform

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
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
	Network    string `JSON:"network"`
	Storage    string `json:"storage"`
	key        string
	cert       string
	servercert string
}

func NewBravehostRemote(settings HostSettings) Remote {
	var protocol string
	var url string

	switch settings.BackendSettings.Type {
	case "lxd":
		protocol = "unix"
	
		//Check which LXC binary is present, and set the url accordingly - JVB
		_, whichLxc, _ := lxdCheck(Lxd{&settings})

		if strings.Contains("/snap/",whichLxc) {
			url = "/var/snap/lxd/common/lxd/unix.socket"
		} else {
			url = "/var/lib/lxd/unix.socket"
		}
	default:
		protocol = "lxd"
		url = "https://" + settings.BackendSettings.Resources.IP + ":8443"
	}

	return Remote{
		Name:     shared.BravetoolsRemote,
		URL:      url,
		Protocol: protocol,
		Public:   false,
		Profile:  settings.Profile,
		Network:  settings.Network.Name,
		Storage:  settings.StoragePool.Name,
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
		return remote, errors.New("unrecognised remote " + name)
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
	remoteNames, err := ListRemotes()
	if err != nil {
		return err
	}

	if shared.StringInSlice(remote.Name, remoteNames) {
		return errors.New("remote " + remote.Name + " already exists")
	}

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

func ListRemotes() (names []string, err error) {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return names, errors.New("failed to list remotes: " + err.Error())
	}

	dir, err := os.Open(path.Join(userHome, shared.BraveRemoteStore))
	if err != nil {
		return names, errors.New("failed to list remotes: " + err.Error())
	}

	names, err = dir.Readdirnames(-1)
	for i := range names {
		names[i] = strings.TrimSuffix(names[i], filepath.Ext(names[i]))
	}

	return names, err
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
