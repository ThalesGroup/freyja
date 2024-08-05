package shellcli

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

// rootCmd is the root command definitions
// define here the helper and the root command flags behavior
var rootCmd = &cobra.Command{
	Use:              "freyja",
	Long:             "Freyja shell client",
	TraverseChildren: true, // ensure local flags do not spread to sub commands

	Run: func(cmd *cobra.Command, args []string) {},
}

func init() {
	// machine management
	rootCmd.AddCommand(machineCmd)

	// network management
	rootCmd.AddCommand(networkCmd)
}

// Execute is the entry point of the cli
// You can call it from external packages
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
