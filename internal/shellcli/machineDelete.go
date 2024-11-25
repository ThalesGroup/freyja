package shellcli

import (
	"errors"
	"freyja/internal"
	"github.com/digitalocean/go-libvirt"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var deleteDomainName string

// commands definitions
var machineDeleteCmd = &cobra.Command{
	Use:              "delete",
	Short:            "Machine deletion",
	Long:             "Machine deletion using handler",
	TraverseChildren: true, // ensure local flags do not spread to sub commands

	Run: func(cmd *cobra.Command, args []string) {
		// user confirmation
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
			// get domain by name
			domain, err := LibvirtConnexion.DomainLookupByName(deleteDomainName)
			if err != nil {
				Logger.Error("Cannot lookup domain from qemu connexion", "domain", deleteDomainName, "reason", err)
				os.Exit(1)
			}

			if err = LibvirtConnexion.DomainDestroyFlags(domain, libvirt.DomainDestroyDefault); err != nil {
				Logger.Error("Cannot stop the domain", "domain", deleteDomainName, "reason", err)
				os.Exit(1)
			}

			if err = LibvirtConnexion.DomainUndefine(domain); err != nil {
				Logger.Error("Cannot delete the domain", "domain", deleteDomainName, "reason", err)
				os.Exit(1)
			}

			// TODO if pool is still present, destroy and undefine it
			// TODO finally delete the freyja folder for this machine

		} else {
			Logger.Info("Canceled")
		}
	},
}

func init() {
	machineDeleteCmd.Flags().StringVarP(&deleteDomainName, "name", "n", "", "Name of the machine to delete.")
	if err := machineDeleteCmd.MarkFlagRequired("name"); err != nil {
		log.Panic(err.Error())
	}
}
