package commands

import (
	"log"

	"github.com/spf13/cobra"
)

var bravePublish = &cobra.Command{
	Use:   "publish <instance> [<instance>...]",
	Short: "Publish deployed Units as images",
	Long:  `Published Units will be saved in the current working directory as *.tar.gz file`,
	Run:   publish,
}

func publish(cmd *cobra.Command, args []string) {
	checkBackend()
	if len(args) == 0 {
		log.Fatal("missing name - please provide unit name")
	}

	for _, name := range args {
		err := host.PublishUnit(name, backend)
		if err != nil {
			log.Fatal(err)
		}
	}
}
