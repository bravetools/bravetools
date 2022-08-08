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
func NewBraveHost() (*BraveHost, error) {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	host := BraveHost{}

	host.Settings, err = loadHostSettings(userHome)
	if err != nil {
		return nil, err
	}

	// Load host remote if initialized
	host.Remote, _ = LoadRemoteSettings(shared.BravetoolsRemote)

	host.Backend, err = NewHostBackend(host.Settings)
	if err != nil {
		return nil, err
	}

	return &host, nil
}

type HostConfig struct {
	Ram     string
	Network string
	Storage string
	Backend string
}

// SetupHostConfiguration creates configuration file and saves it in bravetools directory
func SetupHostConfiguration(params HostConfig, userHome string) (settings HostSettings) {
	poolSizeInt, _ := strconv.Atoi(params.Storage)
	poolSizeInt = poolSizeInt - 2

	hostName, err := getCurrentUsername()
	if err != nil {
		log.Fatalf(err.Error())
	}

	timestamp := time.Now()
	storagePoolName := hostName + "-" + timestamp.Format("20060102150405")

	settings = HostSettings{
		Name:    hostName,
		Trust:   hostName,
		Profile: hostName,
		StoragePool: Storage{
			Type: "zfs",
			Name: storagePoolName,
			Size: strconv.Itoa(poolSizeInt) + "GB",
		},
		Network: Network{
			Bridge: params.Network,
		},
		Status: "inactive",
	}

	if params.Backend == "multipass" {

		backendSettings := BackendSettings{
			Type: "multipass",
			Resources: BackendResources{
				Name: hostName,
				OS:   "bionic",
				CPU:  "2",
				RAM:  params.Ram,
				HD:   params.Storage + "GB",
				IP:   "",
			},
		}
		settings.BackendSettings = backendSettings
	}

	if params.Backend == "lxd" {
		backendSettings := BackendSettings{
			Type: "lxd",
			Resources: BackendResources{
				RAM: "",
				HD:  "",
				IP:  "127.0.0.1",
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

	return settings
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

	lxdServer, err := GetLXDInstanceServer(remote)
	if err != nil {
		return err
	}

	units, err := GetUnits(lxdServer)
	if err != nil {
		return errors.New("failed to list units: " + err.Error())
	}

	if len(units) > 0 {
		return errors.New("one or more units rely on the existing storage pool. Delete all units and try again")
	}

	timestamp := time.Now()
	storagePoolName := settings.Name + "-" + timestamp.Format("20060102150405")
	storagePoolSize := settings.StoragePool.Size

	currentStoragePoolName := settings.StoragePool.Name

	err = CreateStoragePool(lxdServer, storagePoolName, storagePoolSize)
	if err != nil {
		cleanUnusedStoragePool(lxdServer, storagePoolName)
		return errors.New("failed to create new storage pool: " + err.Error())
	}

	err = SetActiveStoragePool(lxdServer, storagePoolName)
	if err != nil {
		cleanUnusedStoragePool(lxdServer, storagePoolName)
		return errors.New("failed to activate storage pool: " + err.Error())
	}

	settings.StoragePool.Name = storagePoolName
	UpdateBraveSettings(settings)

	err = DeleteStoragePool(lxdServer, currentStoragePoolName)
	if err != nil {
		return errors.New("failed to delete storage pool: " + err.Error())
	}

	return nil
}

// loadHostSettings reads config.yaml in /.bravetools directory
func loadHostSettings(userHome string) (HostSettings, error) {
	settings := HostSettings{}
	var buf bytes.Buffer

	f, err := os.Open(path.Join(userHome, shared.PlatformConfig))
	if err != nil {
		return settings, errors.New("failed to load platform configuration: " + err.Error())
	}
	defer f.Close()
	_, err = io.Copy(&buf, f)
	if err != nil {
		return settings, err
	}

	err = yaml.Unmarshal(buf.Bytes(), &settings)
	if err != nil {
		return settings, errors.New("failed to parse configuration yaml: " + err.Error())
	}

	return settings, nil
}
