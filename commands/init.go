package commands

import (
	"fmt"
	"log"
	"os"
	"path"
	"runtime"

	"github.com/bravetools/bravetools/db"
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
var remoteBackend bool

func init() {
	includeInitFlags(hostInit)
}

func includeInitFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&hostConfigPath, "config", "c", "", "Path to the host configuration file [OPTIONAL]")
	cmd.PersistentFlags().StringVarP(&storage, "storage", "s", "12", "Host storage size in GB[OPTIONAL]. default: 12")
	cmd.PersistentFlags().StringVarP(&ram, "memory", "m", "4GB", "Host memory size [OPTIONAL]. default 4GB")
	cmd.PersistentFlags().StringVarP(&network, "network", "n", "", "Host network IP range [OPTIONAL]. default: randomly generate RFC1918 address")

	cmd.PersistentFlags().BoolVar(&remoteBackend, "remote", false, "whether backend is remote (will not be initialized)")
}

func serverInit(cmd *cobra.Command, args []string) {
	userHome, _ := os.UserHomeDir()

	if _, err := os.Stat(path.Join(userHome, shared.BraveHome)); !os.IsNotExist(err) {
		log.Fatal("$HOME/.bravetools directory exists. Run rm -r $HOME/.bravetools to create a fresh install")
	}

	hostOs := runtime.GOOS

	if !remoteBackend {
		switch hostOs {
		case "linux":
			backendType = "lxd"
		case "darwin":
			backendType = "multipass"
		case "windows":
			backendType = "multipass"
		default:
			err := deleteBraveHome(userHome)
			if err != nil {
				fmt.Println(err.Error())
			}
			log.Fatal("unsupported host OS: ", hostOs)
		}

		if network == "" {
			ip, err := shared.GenerateRandomRFC1919()
			if err != nil {
				log.Fatal(err.Error())
			}
			network = ip
		}
	} else {
		backendType = "remote"
	}

	// Create $HOME/.bravetools
	err := createBraveHome(userHome)
	if err != nil {
		log.Fatal(err.Error())
	}

	dbPath := path.Join(userHome, shared.BraveDB)

	log.Println("Initialising Bravetools unit database")
	_, err = os.Stat(dbPath)
	if os.IsNotExist(err) {
		err = db.InitDB(dbPath)

		if err != nil {
			if err := deleteBraveHome(userHome); err != nil {
				fmt.Println(err.Error())
			}
			log.Fatal("failed to initialize database: ", err)
		}
	}

	params := platform.HostConfig{
		Storage: storage,
		Ram:     ram,
		Network: network,
		Backend: backendType,
	}

	if hostConfigPath != "" {
		err = shared.CopyFile(hostConfigPath, path.Join(userHome, shared.PlatformConfig))
		if err != nil {
			if err := deleteBraveHome(userHome); err != nil {
				fmt.Println(err.Error())
			}
			log.Fatal(err)
		}
		loadConfig()
	} else {
		userHome, _ := os.UserHomeDir()
		platform.SetupHostConfiguration(params, userHome)
		loadConfig()
	}

	if remoteBackend {
		fmt.Println("bravetools initialized - add a remote with `brave remote add local ...`")
		return
	}

	log.Println("Initialising Bravetools backend")
	err = backend.BraveBackendInit()
	if err != nil {
		if err := deleteBraveHome(userHome); err != nil {
			fmt.Println(err.Error())
		}

		log.Fatal("error initializing Bravetools backend: ", err)
	}

	loadConfig()

	if backendType == "multipass" {
		info, err := backend.Info()

		if err != nil {
			if err := deleteBraveHome(userHome); err != nil {
				fmt.Println(err.Error())
			}
			log.Fatal(err)
		}

		settings := host.Settings
		if hostOs == "windows" {
			settings.BackendSettings.Resources.IP = info.Name + ".mshome.net"
		} else {
			settings.BackendSettings.Resources.IP = info.IPv4
		}
		err = platform.UpdateBraveSettings(settings)

		if err != nil {
			if err := deleteBraveHome(userHome); err != nil {
				fmt.Println(err.Error())
			}
			log.Fatal(err)
		}

		loadConfig()
	}

	log.Println("Registering a Remote")
	host.Remote = platform.NewBravehostRemote(host.Settings)
	err = platform.SaveRemote(host.Remote)
	if err != nil {
		if err := deleteBraveHome(userHome); err != nil {
			fmt.Println(err.Error())
		}
		log.Fatal("failed to save default bravetools remote: ", err)
	}
	err = host.AddRemote()
	if err != nil {
		if err := deleteBraveHome(userHome); err != nil {
			fmt.Println(err.Error())
		}
		log.Fatal(err)
	}
}
