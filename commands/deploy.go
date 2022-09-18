package commands

import (
	"log"

	"github.com/bravetools/bravetools/shared"
	"github.com/spf13/cobra"
)

var braveDeploy = &cobra.Command{
	Use:   "deploy IMAGE",
	Short: "Deploy Unit from image",
	Long: `Bravetools supports Unit deployment using either command line arguments or a configuration file.
In cases where IPv4 address is not provided, a random ephemeral IP address will be assigned. More detailed
deployment options e.g. CPU and RAM should be configured through a configuration file.

Deployment loads Unit specifications from a Bravefile, which is expected to be in the working directory.
Parameters specificed in the Bravefile can be overridden using command line options. If Bravefile does not
exist in the current working directory, Bravetools expects an image name as the first agrument.

Deployment can be both local or to a remote Bravetools host. To specificify a remote host using command line argument
set --name flag to <remote>:<instance>`,
	Run: deploy,
}
var unitConfig string
var deployArgs = &shared.Service{}

func init() {
	includeDeployFlags(braveDeploy)
}

func includeDeployFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&unitConfig, "config", "", "", "Path to Unit configuration file [OPTIONAL]")
	cmd.Flags().StringVarP(&deployArgs.IP, "ip", "i", "", "IPv4 address (e.g., 10.0.0.20) [OPTIONAL]")
	cmd.Flags().StringVarP(&deployArgs.Resources.CPU, "cpu", "c", "", "Number of allocated CPUs (e.g., 2) [OPTIONAL]")
	cmd.Flags().StringVarP(&deployArgs.Resources.RAM, "ram", "r", "", "Number of allocated CPUs (e.g., 2GB) [OPTIONAL]")
	cmd.Flags().StringVarP(&deployArgs.Profile, "profile", "", "", "LXD profile to deploy to. Defaults to bravetools local profile [OPTIONAL]")
	cmd.Flags().StringSliceVarP(&deployArgs.Ports, "port", "p", []string{}, "Publish Unit port to host [OPTIONAL]")
	cmd.Flags().StringVarP(&deployArgs.Name, "name", "n", "", "Assign name to deployed Unit")
	cmd.Flags().StringVar(&deployArgs.Network, "network", "", "LXD-managed bridge to use for networking containers (e.g. lxdbr0)")
	cmd.Flags().StringVar(&deployArgs.Storage, "storage", "", "Name of LXD storage pool to use for container")
}

func deploy(cmd *cobra.Command, args []string) {
	checkBackend()

	var useBravefile = true
	var bravefilePath = "Bravefile"
	var err error

	// If args provided, use CLI, not Bravefile
	if len(args) > 0 {
		useBravefile = false
		bravefile.PlatformService.Image = args[0]
		if deployArgs.Name == "" {
			log.Fatal("unit must have a name: pass one using the '--name' flag")
		}
	}

	// Use Bravefile if no CLI args
	if useBravefile {
		if unitConfig != "" {
			bravefilePath = unitConfig
		}
		if !shared.FileExists(bravefilePath) {
			log.Fatalf("Bravefile not found at %q", bravefilePath)
		}

		err = bravefile.Load(bravefilePath)
		if err != nil {
			log.Fatal(err)
		}
	}

	deployArgs.Merge(&bravefile.PlatformService)
	bravefile.PlatformService = *deployArgs

	err = host.InitUnit(backend, &bravefile.PlatformService)
	if err != nil {
		log.Fatal(err)
	}
}
