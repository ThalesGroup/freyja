package shellcli

import (
	"errors"
	"freyja/internal"
	"github.com/spf13/cobra"
)

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
			Logger.Info("Deleted")
		} else {
			Logger.Info("Canceled")
		}
	},
}
