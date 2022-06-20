package commands

import (
	"log"
	"os"

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
exist in the current working directory, Bravetools expects an image name as the first agrument.`,
	Run: deploy,
}
var unitConfig, unitIP, unitCPU, unitRAM, name string
var unitPort []string

func init() {
	includeDeployFlags(braveDeploy)
}

func includeDeployFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&unitConfig, "config", "", "", "Path to Unit configuration file [OPTIONAL]")
	cmd.Flags().StringVarP(&unitIP, "ip", "i", "", "IPv4 address (e.g., 10.0.0.20) [OPTIONAL]")
	cmd.Flags().StringVarP(&unitCPU, "cpu", "c", "", "Number of allocated CPUs (e.g., 2) [OPTIONAL]")
	cmd.Flags().StringVarP(&unitRAM, "ram", "r", "", "Number of allocated CPUs (e.g., 2GB) [OPTIONAL]")
	cmd.Flags().StringSliceVarP(&unitPort, "port", "p", []string{}, "Publish Unit port to host [OPTIONAL]")
	cmd.Flags().StringVarP(&name, "name", "n", "", "Assign name to deployed Unit")
}

func deploy(cmd *cobra.Command, args []string) {
	checkBackend()

	var useBravefile = false
	var bravefilePath string
	var err error

	_, err = os.Stat("Bravefile")
	// if Bravefile is in current directory continue with parameters set there
	if err == nil {
		useBravefile = true
		bravefilePath = "Bravefile"
	}
	if unitConfig != "" {
		useBravefile = true
		bravefilePath = unitConfig
	}

	if useBravefile {
		err = bravefile.Load(bravefilePath)
		if err != nil {
			log.Fatal(err)
		}

	}
	if (len(args) == 0) && (!useBravefile) {
		log.Fatal("missing image name")
	}

	if len(args) > 0 {
		bravefile.PlatformService.Image = args[0]
		if unitCPU == "" {
			bravefile.PlatformService.Resources.CPU = "2"
		}

		if unitRAM == "" {
			bravefile.PlatformService.Resources.RAM = "2GB"
		}
	}

	if name != "" {
		bravefile.PlatformService.Name = name
	}

	if unitCPU != "" {
		bravefile.PlatformService.Resources.CPU = unitCPU
	}

	if unitRAM != "" {
		bravefile.PlatformService.Resources.RAM = unitRAM
	}

	if unitIP != "" {
		bravefile.PlatformService.IP = unitIP
	}

	if len(unitPort) != 0 {
		bravefile.PlatformService.Ports = unitPort
	}

	err = host.InitUnit(backend, bravefile)
	if err != nil {
		log.Fatal(err)
	}
}
