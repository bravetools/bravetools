package commands

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/bravetools/bravetools/platform"
	"github.com/bravetools/bravetools/shared"
	"github.com/spf13/cobra"
)

var remoteCmd = &cobra.Command{
	Use:   "remote",
	Short: "Manage remotes",
	Long:  ``,
}

var remoteAddCmd = &cobra.Command{
	Use:   "add [NAME] [URL]",
	Short: "Add a remote",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			cmd.Usage()
			return fmt.Errorf("`brave remote add` expects two positional args, [NAME] and [URL], got %d", len(args))
		}
		return nil
	},
	Long: `Create a remote called [NAME] available at [URL].
Example: brave remote add test https://localhost:8443`,
	Run: remoteAdd,
}

var remoteRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a remote",
	Long:  ``,
	Args:  cobra.MinimumNArgs(1),
	Run:   remoteRemove,
}

var remoteGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a remote",
	Long:  "Returns a JSON string listing configuration of selected remote",
	Args:  cobra.ExactArgs(1),
	Run:   remoteGet,
}

var remoteListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all remotes",
	Long:  "",
	Args:  cobra.NoArgs,
	Run:   remoteList,
}

var remoteArgs = &platform.Remote{}
var remotePassword = ""

func init() {
	remoteCmd.AddCommand(remoteAddCmd)
	remoteCmd.AddCommand(remoteRemoveCmd)
	remoteCmd.AddCommand(remoteGetCmd)
	remoteCmd.AddCommand(remoteListCmd)
	includeRemoteAddFlags(remoteAddCmd)
}

func includeRemoteAddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&remoteArgs.Protocol, "protocol", "lxd", "LXD server protocol to connect with (e.g. 'lxd', 'simplestreams')")
	cmd.Flags().BoolVar(&remoteArgs.Public, "public", false, "Publicly available server with no authentication")
	cmd.Flags().StringVar(&remoteArgs.Profile, "profile", "default", "Name of LXD profile to use with this remote.")
	cmd.Flags().StringVar(&remoteArgs.Network, "network", "lxdbr0", "LXD-managed bridge to use for networking containers")
	cmd.Flags().StringVar(&remoteArgs.Storage, "storage", "", "Name of LXD storage pool to use for container")
	cmd.Flags().StringVar(&remotePassword, "password", "", "Trusted password to use when communicating with remote")
}

func remoteAdd(cmd *cobra.Command, args []string) {

	remoteArgs.Name = args[0]
	remoteArgs.URL = args[1]

	err := platform.SaveRemote(*remoteArgs)
	if err != nil {
		log.Fatal(err)
	}

	// Need to generate certs for non-public remotes
	if !remoteArgs.Public && !(remoteArgs.Protocol == "unix") {

		if remotePassword == "" {
			log.Println("adding a non-public remote without providing a trusted password with the --password flag only works with remotes that already trust bravetools")
		}

		err = platform.AddRemote(*remoteArgs, remotePassword)
		if err != nil {
			platform.RemoveRemote(remoteArgs.Name)
			log.Fatal(err)
		}
	} else {
		// Validate public/unix connection
		if remoteArgs.Protocol == "unix" {
			_, err = platform.GetLXDInstanceServer(*remoteArgs)
		} else {
			_, err = platform.GetLXDImageSever(*remoteArgs)
		}

		if err != nil {
			platform.RemoveRemote(remoteArgs.Name)
			log.Fatal(err)
		}
	}
}

func remoteRemove(cmd *cobra.Command, args []string) {
	for _, arg := range args {
		if arg == shared.BravetoolsRemote {
			log.Printf("remote %q cannot be removed, skipping", arg)
			continue
		}

		err := platform.RemoveRemote(arg)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func remoteGet(cmd *cobra.Command, args []string) {
	remote, err := platform.LoadRemoteSettings(args[0])
	if err != nil {
		log.Fatal(err)
	}
	remoteJson, err := json.MarshalIndent(remote, "", "    ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(remoteJson))
}

func remoteList(cmd *cobra.Command, args []string) {
	remoteNames, err := platform.ListRemotes()
	if err != nil {
		log.Fatal(err)
	}
	for _, name := range remoteNames {
		fmt.Println(name)
	}
}
