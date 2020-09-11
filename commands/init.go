package commands

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"runtime"

	"github.com/bravetools/bravetools/platform"
	"github.com/bravetools/bravetools/shared"
	"github.com/spf13/cobra"
)

var hostInit = &cobra.Command{
	Use:   "init",
	Short: "Create a new Bravetools host",
	Long:  ``,
	Run:   serverInit,
}

var hostConfigPath, storage, ram, network, backendType string

func init() {
	includeInitFlags(hostInit)
}

func includeInitFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&hostConfigPath, "config", "", "", "Path to the host configuration file [OPTIONAL]")
	cmd.PersistentFlags().StringVarP(&storage, "storage", "s", "", "Host storage size [OPTIONAL]")
	cmd.PersistentFlags().StringVarP(&ram, "memory", "m", "", "Host memory size [OPTIONAL]")
	cmd.PersistentFlags().StringVarP(&network, "network", "n", "", "Host network IP range [OPTIONAL]")
	cmd.PersistentFlags().StringVarP(&backendType, "backend", "b", "", "Backend type (multipass or lxd) [OPTIONAL]")
}

func serverInit(cmd *cobra.Command, args []string) {
	userHome, _ := os.UserHomeDir()

	params := make(map[string]string)

	if _, err := os.Stat(path.Join(userHome, ".bravetools")); !os.IsNotExist(err) {
		msg := errors.New("Bravetools is already initialised. Run \"brave configure\" if you'd like to tweak configuration")
		log.Fatal(msg.Error())
	}

	err := createLocalDirectories(userHome)
	if err != nil {
		log.Fatal(err.Error())
	}

	if storage == "" {
		storage = "12"
	}
	params["storage"] = storage
	if ram == "" {
		ram = "4GB"
	}
	params["ram"] = ram
	if network == "" {
		network = "10.0.0.1"
	}
	params["network"] = network
	if backendType == "" {
		hostOs := runtime.GOOS
		switch hostOs {
		case "linux":
			backendType = "lxd"
		case "darwin":
			backendType = "multipass"
		case "windows":
			backendType = "multipass"
		default:
			err := deleteLocalDirectories(userHome)
			if err != nil {
				log.Fatal(err.Error())
			}
			fmt.Println(runtime.GOOS)
			fmt.Println("Unsupported OS")
		}
	}
	params["backend"] = backendType

	if hostConfigPath != "" {
		// TODO: validate configuration. Now assume that path ends with config.yml
		err = shared.CopyFile(hostConfigPath, path.Join(userHome, ".bravetools", "config.yml"))
		if err != nil {
			log.Fatal(err)
		}
	} else {
		userHome, _ := os.UserHomeDir()
		platform.SetupHostConfiguration(params, userHome)
		loadConfig()
	}

	loadConfig()
	err = backend.BraveBackendInit()
	if err != nil {
		log.Fatal(err)
	}

	loadConfig()
	if backendType == "multipass" {
		info, err := backend.Info()
		if err != nil {
			log.Fatal(err)
		}
		settings := host.Settings

		settings.BackendSettings.Resources.IP = info.IPv4
		err = platform.UpdateBraveSettings(settings)
		if err != nil {
			log.Fatal(err)
		}

		loadConfig()
	}

	err = host.AddRemote()
	if err != nil {
		log.Fatal(err)
	}
}
