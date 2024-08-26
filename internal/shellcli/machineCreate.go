package shellcli

import (
	"freyja/internal"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var configurationPath string

// commands definitions
var machineCreateCmd = &cobra.Command{
	Use:              "create",
	Short:            "Machine creation",
	Long:             "Machine creation using handler and cloud-init or ignition",
	TraverseChildren: true, // ensure local flags do not spread to sub commands

	Run: func(cmd *cobra.Command, args []string) {
		// build config from path
		var confData internal.Configuration
		if err := confData.BuildFromFile(configurationPath); err != nil {
			Logger.Error("cannot parse configuration", "configuration", configurationPath, "reason", err)
			os.Exit(1)
		}
		log.Print(confData)
		// TODO :
		//   - build and create the cloud-init from model, using the go module for cloud init
		//   - build and create the vm configuration from model
		//   - create the domain using libvirt with the cloud init iso file for boot provisioning
		//cloudconfig.New()
	},
}

func init() {
	machineCreateCmd.Flags().StringVarP(&configurationPath, "config", "c", "", "Path to the configuration file to create the machines and the networks.")
	if err := machineCreateCmd.MarkFlagRequired("config"); err != nil {
		log.Panic(err.Error())
	}
}

func machineCreate(configurationPath string) (*internal.XMLDomainDescription, error) {
	//config, err := internal.BuildConfiguration(configurationPath)
	//if err != nil {
	//	return nil, err
	//}
	//
	return nil, nil

	// TODO create a metadata to know which domain uses what network
}
