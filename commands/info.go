package commands

import (
	"log"

	"github.com/spf13/cobra"
)

var hostInfo = &cobra.Command{
	Use:   "info",
	Short: "Display workspace information",
	Long:  ``,
	Run:   hostInfoList,
}

var short bool

func init() {
	includeInfoFlags(hostInfo)
}

func includeInfoFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&short, "short", false, "Returns host IP address")
}

func hostInfoList(cmd *cobra.Command, args []string) {
	checkBackend()
	err := host.HostInfo(short)
	if err != nil {
		log.Fatal(err)
	}
}
