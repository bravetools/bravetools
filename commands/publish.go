package commands

import (
	"log"
	"os"

	"github.com/spf13/cobra"
)

var bravePublish = &cobra.Command{
	Use:   "publish NAME",
	Short: "Publish deployed Unit as image",
	Long:  `Published Unit will be saved in the current working directory as *.tar.gz file`,
	Run:   publish,
}

func publish(cmd *cobra.Command, args []string) {
	checkBackend()
	if len(args) == 0 {
		log.Fatal("missing name - please provide unit name")
	}

	name := args[0]

	err := host.PublishUnit(name, backend)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
