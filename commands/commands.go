package commands

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/bravetools/bravetools/platform"
	"github.com/bravetools/bravetools/shared"

	"github.com/spf13/cobra"
)

var host platform.BraveHost
var backend platform.Backend
var bravefile *shared.Bravefile

var (
	// BravetoolsCmd ..
	BravetoolsCmd = &cobra.Command{
		Use:           "brave",
		Short:         "A complete System Container management platform",
		Long:          ``,
		SilenceErrors: true,
		SilenceUsage:  true,
	}
)

func init() {
	BravetoolsCmd.AddCommand(configureHost)
	BravetoolsCmd.AddCommand(hostInit)
	BravetoolsCmd.AddCommand(hostInfo)
	BravetoolsCmd.AddCommand(braveBuild)
	BravetoolsCmd.AddCommand(braveRemove)
	BravetoolsCmd.AddCommand(braveListUnits)
	BravetoolsCmd.AddCommand(braveListImages)
	BravetoolsCmd.AddCommand(mountDir)
	BravetoolsCmd.AddCommand(umountDir)
	BravetoolsCmd.AddCommand(braveImportImage)
	BravetoolsCmd.AddCommand(braveDeploy)
	BravetoolsCmd.AddCommand(braveStart)
	BravetoolsCmd.AddCommand(braveStop)
	BravetoolsCmd.AddCommand(bravePublish)
	BravetoolsCmd.AddCommand(baseBuild)
	BravetoolsCmd.AddCommand(braveVersion)

	userHome, _ := os.UserHomeDir()
	exists, err := shared.CheckPath(path.Join(userHome, shared.PlatformConfig))
	if err != nil {
		log.Fatal(err.Error())
	}

	if exists {
		bravefile = shared.NewBravefile()
		loadConfig()
	}
}

func checkBackend() {
	if host.Settings.Name == "" {
		fmt.Println("brave host is not initialized. Run \"brave init\"")
		os.Exit(1)
	}
}

func setBackend(host platform.BraveHost) error {
	backendType := host.Settings.BackendSettings.Type

	switch backendType {
	case "multipass":
		backend = platform.NewMultipass(host.Settings)
	case "lxd":
		backend = platform.NewLxd(host.Settings)
	}

	return nil
}

func createBraveHome(userHome string) error {
	err := shared.CreateDirectory(path.Join(userHome, shared.BraveHome))
	if err != nil {
		return err
	}

	err = shared.CreateDirectory(path.Join(userHome, shared.BraveCertStore))
	if err != nil {
		return err
	}

	err = shared.CreateDirectory(path.Join(userHome, shared.ImageStore))
	if err != nil {
		return err
	}

	err = shared.CreateDirectory(path.Join(userHome, shared.BraveServerCertStore))
	if err != nil {
		return err
	}
	return nil
}

func deleteBraveHome(userHome string) error {
	exists, err := shared.CheckPath(path.Join(userHome, shared.BraveHome))
	if err != nil {
		return err
	}

	if exists {
		err = os.RemoveAll(path.Join(userHome, shared.BraveHome))
		if err != nil {
			return err
		}
	}
	return nil
}

func loadConfig() {
	host = *platform.NewBraveHost()
	err := setBackend(host)
	if err != nil {
		log.Fatal(err)
	}
}
