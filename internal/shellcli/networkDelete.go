package shellcli

import (
	"errors"
	"freyja/internal"
	"github.com/spf13/cobra"
	"log"
	"os"
)

// commands definitions
var networkDeleteCmd = &cobra.Command{
	Use:              "delete",
	Short:            "Virtual network deletion",
	Long:             "Virtual network deletion using handler",
	TraverseChildren: true, // ensure local flags do not spread to sub commands

	Run: func(cmd *cobra.Command, args []string) {
		// TODO:
		//   - ask a confirmation
		agree, err := internal.AskUserYesNoConfirmation()
		if err != nil {
			if errors.Is(err, internal.ErrUserInput) {
				Logger.Error("wrong choice", "reason", err)
			} else {
				Logger.Error("cannot analyse user choice", "reason", err)
			}
		}

		if agree {
			// find
			network, err := LibvirtConnexion.NetworkLookupByName(networkName)
			if err != nil {
				Logger.Error("Cannot lookup network by name using Qemu connexion", "network", networkName, "reason", err)
				os.Exit(1)
			}

			// destroy
			if err := LibvirtConnexion.NetworkDestroy(network); err != nil {
				Logger.Error("Cannot destroy network", "network", networkName, "reason", err)
				os.Exit(1)
			}

			Logger.Info("Network deleted", "network", networkName)
		} else {
			Logger.Info("Canceled")
		}

	},
}

func init() {
	networkDeleteCmd.Flags().StringVarP(&networkName, "name", "n", "", "Name of the network to delete.")
	if err := networkDeleteCmd.MarkFlagRequired("name"); err != nil {
		log.Panic(err.Error())
	}
}
