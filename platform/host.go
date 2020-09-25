package platform

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/bravetools/bravetools/shared"
	"gopkg.in/yaml.v2"
)

// Remote represents a configuration of the remote
type Remote struct {
	remoteURL string
	key       string
	cert      string
}

// HostSettings configuration data loaded from config.yaml
type HostSettings struct {
	Name            string          `yaml:"name"`
	Trust           string          `yaml:"trust"`
	Profile         string          `yaml:"profile"`
	StoragePool     Storage         `yaml:"storage"`
	Network         Network         `yaml:"network"`
	BackendSettings BackendSettings `yaml:"backendsettings"`
	Status          string          `yaml:"status"`
}

// Storage ..
type Storage struct {
	Type string `yaml:"type"`
	Name string `yaml:"name"`
	Size string `yaml:"size"`
}

// Network ..
type Network struct {
	Bridge string `yaml:"bridge"`
}

// BackendSettings ..
type BackendSettings struct {
	Type      string           `yaml:"type"`
	Resources BackendResources `yaml:"resources"`
}

// BackendResources ..
type BackendResources struct {
	Name string `yaml:"name"`
	OS   string `yaml:"os"`
	CPU  string `yaml:"cpu"`
	RAM  string `yaml:"ram"`
	HD   string `yaml:"hd"`
	IP   string `yaml:"ip"`
}

// BraveHost ..
type BraveHost struct {
	Settings HostSettings `yaml:"settings"`
	Remote   Remote
	Backend  Backend
}

// NewBraveHost returns Brave host
func NewBraveHost() *BraveHost {
	userHome, _ := os.UserHomeDir()

	settings, err := loadHostSettings(userHome)
	if err != nil {
		log.Fatal(err)
	}

	remote, err := loadRemoteSettings(userHome, settings.BackendSettings.Resources.IP)
	if err != nil {
		log.Fatal(err)
	}

	return &BraveHost{
		Settings: settings,
		Remote:   remote,
	}
}

// SetupHostConfiguration creates configuration file and saves it in bravetools directory
func SetupHostConfiguration(params map[string]string, userHome string) {
	var settings = HostSettings{}
	poolSizeInt, _ := strconv.Atoi(params["storage"])
	poolSizeInt = poolSizeInt - 2

	timestamp := time.Now()
	storagePoolName := "brave-" + timestamp.Format("20060102150405")

	settings = HostSettings{
		Name:    "brave",
		Trust:   "brave",
		Profile: "brave",
		StoragePool: Storage{
			Type: "zfs",
			Name: storagePoolName,
			Size: strconv.Itoa(poolSizeInt) + "GB",
		},
		Network: Network{
			Bridge: params["network"],
		},
		Status: "inactive",
	}

	if params["backend"] == "multipass" {
		backendSettings := BackendSettings{
			Type: "multipass",
			Resources: BackendResources{
				Name: "brave",
				OS:   "bionic",
				CPU:  "2",
				RAM:  params["ram"],
				HD:   params["storage"] + "GB",
				IP:   "",
			},
		}
		settings.BackendSettings = backendSettings
	}

	if params["backend"] == "lxd" {
		backendSettings := BackendSettings{
			Type: "lxd",
			Resources: BackendResources{
				RAM: "",
				HD:  "",
				IP:  "0.0.0.0",
			},
		}
		settings.BackendSettings = backendSettings
	}

	doc, err := yaml.Marshal(settings)
	if err != nil {
		log.Fatal(err.Error())
	}

	err = ioutil.WriteFile(path.Join(userHome, shared.PlatformConfig), doc, os.ModePerm)
	if err != nil {
		log.Fatal(err.Error())
	}
}

// UpdateBraveSettings configuration in place and write to config.yaml
func UpdateBraveSettings(settings HostSettings) error {
	config, err := yaml.Marshal(settings)
	if err != nil {
		return errors.New("Failed to update host settings file: " + err.Error())
	}

	userHome, _ := os.UserHomeDir()
	err = ioutil.WriteFile(path.Join(userHome, shared.PlatformConfig), config, os.ModePerm)
	if err != nil {
		return errors.New("Failed to write bravetools settings to file: " + err.Error())
	}

	return nil
}

// ConfigureHost configures local bravetools host and updates resources
func ConfigureHost(settings HostSettings, remote Remote) error {
	units, err := listHostUnits(remote)
	if err != nil {
		return errors.New("Failed to list units: " + err.Error())
	}

	if len(units) > 0 {
		return errors.New("One or more units rely on the existing storage pool. Delete all units and try again")
	}

	timestamp := time.Now()
	storagePoolName := "brave-" + timestamp.Format("20060102150405")
	storagePoolSize := settings.StoragePool.Size

	currentStoragePoolName := settings.StoragePool.Name

	err = CreateStoragePool(storagePoolName, storagePoolSize, remote)
	if err != nil {
		cleanUnusedStoragePool(storagePoolName, remote)
		return errors.New("Failed to create new storage pool: " + err.Error())
	}

	err = SetActiveStoragePool(storagePoolName, remote)
	if err != nil {
		cleanUnusedStoragePool(storagePoolName, remote)
		return errors.New("Failed to activate storage pool: " + err.Error())
	}

	settings.StoragePool.Name = storagePoolName
	UpdateBraveSettings(settings)

	err = DeleteStoragePool(currentStoragePoolName, remote)
	if err != nil {
		return errors.New("Failed to delete storage pool: " + err.Error())
	}

	return nil
}

func loadRemoteSettings(userHome string, remoteIP string) (Remote, error) {
	remote := Remote{}

	keyPath := userHome + shared.BraveClientKey
	certPath := userHome + shared.BraveClientCert
	key, _ := loadKey(keyPath)
	cert, _ := loadCert(certPath)

	remote.key = key
	remote.cert = cert
	remote.remoteURL = "https://" + remoteIP + ":8443"

	return remote, nil
}

// loadHostSettings reads config.yaml in /.bravetools directory
func loadHostSettings(userHome string) (HostSettings, error) {
	settings := HostSettings{}
	var buf bytes.Buffer

	f, err := os.Open(path.Join(userHome, shared.PlatformConfig))
	if err != nil {
		return settings, errors.New("Failed to load platform configuration: " + err.Error())
	}
	defer f.Close()
	_, err = io.Copy(&buf, f)
	if err != nil {
		return settings, err
	}

	err = yaml.Unmarshal(buf.Bytes(), &settings)
	if err != nil {
		return settings, errors.New("Failed to parse configuration yaml: " + err.Error())
	}

	return settings, nil
}

func loadKey(path string) (string, error) {
	buf, err := shared.ReadFile(path)
	if err != nil {
		return "", errors.New("Cannot load client key")
	}
	key := buf.String()
	return key, nil
}

func loadCert(path string) (string, error) {
	buf, err := shared.ReadFile(path)
	if err != nil {
		return "", errors.New("Cannot load client certificate")
	}
	cert := buf.String()
	return cert, nil
}
