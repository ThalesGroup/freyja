package shellcli

import (
	"errors"
	"freyja/internal"
	"github.com/spf13/cobra"
	"log"
	"os"
	"strings"
)

var deleteDomainName string

// commands definitions
var machineDeleteCmd = &cobra.Command{
	Use:              "delete",
	Short:            "Machine deletion",
	Long:             "Machine deletion using handler",
	TraverseChildren: true, // ensure local flags do not spread to sub commands

	Run: func(cmd *cobra.Command, args []string) {
		// get domain by name
		domain, err := LibvirtConnexion.DomainLookupByName(deleteDomainName)
		// cancel if the machine is not found
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				Logger.Warn("canceled", "reason", "machine not found", "machine", deleteDomainName)
				os.Exit(0)
			} else {
				Logger.Error("cannot find the machine", "machine", deleteDomainName, "reason", err)
				os.Exit(1)
			}
		}

		// user confirmation
		Logger.Info("delete", "machines", deleteDomainName)
		agree, err := internal.AskUserYesNoConfirmation()
		if err != nil {
			if errors.Is(err, internal.ErrUserInput) {
				Logger.Error("wrong choice", "reason", err)
			} else {
				Logger.Error("cannot analyse user choice", "reason", err)
			}
		}

		// exec
		if agree {
			// stop libvirt domain
			if err = LibvirtConnexion.DomainDestroy(domain); err != nil {
				Logger.Error("cannot stop the machines", "machines", deleteDomainName, "reason", err)
				os.Exit(1)
			}
			// undefine libvirt domain
			if err = LibvirtConnexion.DomainUndefine(domain); err != nil {
				Logger.Error("cannot undefine the machines", "machines", deleteDomainName, "reason", err)
				os.Exit(1)
			}
			// delete machine directory in filesystem
			machineDirPath := getMachineDirByName(deleteDomainName)
			if err = os.RemoveAll(machineDirPath); err != nil {
				Logger.Error("cannot remove machine directory", "machine", deleteDomainName, "dir", machineDirPath, "error", err)
				os.Exit(1)
			}
			Logger.Info("deleted", "machines", deleteDomainName)

		} else {
			Logger.Info("canceled")
		}
	},
}

func init() {
	machineDeleteCmd.Flags().StringVarP(&deleteDomainName, "name", "n", "", "Name of the machine to delete.")
	if err := machineDeleteCmd.MarkFlagRequired("name"); err != nil {
		log.Panic(err.Error())
	}
}
