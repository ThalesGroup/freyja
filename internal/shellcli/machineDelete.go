package shellcli

import (
	"errors"
	"freyja/internal"
	"github.com/digitalocean/go-libvirt"
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
			// get domain by name
			domain, err := LibvirtConnexion.DomainLookupByName(deleteDomainName)
			if err != nil {
				if strings.Contains(err.Error(), "not found") {
					Logger.Warn("canceled : machines not found", "machines", deleteDomainName)
					os.Exit(0)
				} else {
					Logger.Error("cannot lookup domain from qemu connexion", "domain", deleteDomainName, "reason", err)
					os.Exit(1)
				}
			}

			if err = LibvirtConnexion.DomainDestroyFlags(domain, libvirt.DomainDestroyDefault); err != nil {
				Logger.Error("cannot stop the machines", "machines", deleteDomainName, "reason", err)
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
