package commands

import (
	"log"

	"github.com/spf13/cobra"
)

var braveListImages = &cobra.Command{
	Use:   "images",
	Short: "List images",
	Long:  ``,
	Run:   listImages,
}

func listImages(cmd *cobra.Command, args []string) {
	checkBackend()
	err := host.PrintLocalImages()
	if err != nil {
		log.Fatal(err)
	}
}
