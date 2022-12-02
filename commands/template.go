package commands

import (
	"log"
	"os"

	"github.com/bravetools/bravetools/shared"
	"github.com/spf13/cobra"
)

var braveTemplateCmd = &cobra.Command{
	Use:   "template",
	Short: "Generate a template Bravefile",
	Long:  `Creates an empty template Bravefile in the current directory.`,
	Run:   generateTemplate,
}

func generateTemplate(cmd *cobra.Command, args []string) {
	destPath := "./Bravefile"

	if _, err := os.Stat(destPath); err == nil {
		log.Fatal("Bravefile already exists in the current directory")
	}

	f, err := os.Create(destPath)
	if err != nil {
		log.Fatal(err)
	}

	_, err = f.Write([]byte(shared.BravefileTemplate))
	if err != nil {
		log.Fatal(err)
	}
}
