package shellcli

import (
	"bufio"
	"freyja/internal"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
	"text/template"
)

var configurationPath string

const networkTemplatePath string = "templates/network.xml.tmpl"

type NetworkData struct {
	Name      string
	UUID      string
	Interface string
}

var networksDir string

// commands definitions
var networkCreateCmd = &cobra.Command{
	Use:              "create",
	Short:            "Network creation",
	Long:             "Network creation using libvirt",
	TraverseChildren: true, // ensure local flags do not spread to sub commands

	Run: func(cmd *cobra.Command, args []string) {
		// logger
		internal.InitPrettyLogger()

		// init the network workspace directory in home user
		initNetworksWorkspace()

		// execute
		if err := networkCreate("test-dev"); err != nil {
			log.Panic("Cannot create network: ", err)
		}
	},
}

// initNetworksWorkspace creates the networks dir
func initNetworksWorkspace() {
	networksDir = filepath.Join(FreyjaWorkspaceDir, "networks")
	if err := os.MkdirAll(networksDir, os.ModePerm); err != nil {
		log.Panic("Could not create networks directory: ", err)
	}
}

func init() {
	networkCreateCmd.Flags().StringVarP(&configurationPath, "config", "c", "", "Path to the configuration file to create the machines and the networks.")
	if err := networkCreateCmd.MarkFlagRequired("config"); err != nil {
		log.Panic(err.Error())
	}

}

// networkCreate
// we choose to implement network creation for machine creation only
func networkCreate(networkName string) error {
	// TODO Check if the uuid does not already used by an existing network
	uuid := internal.GenerateUUID()
	// template input data
	data := &NetworkData{
		Name: networkName,
		UUID: uuid,
	}

	// template file loading
	networkTemplateContent, err := Templates.ReadFile(networkTemplatePath)
	if err != nil {
		log.Panic("Impossible to load internal network template file: ", err)
	}

	// template rendering in output file
	outputPath := filepath.Join(networksDir, networkName+".xml")
	outputIoFile, err := os.Create(outputPath)
	if err != nil {
		Logger.Error("Cannot open path to generate network config file", "path", outputPath)
		return err
	}
	defer outputIoFile.Close()
	outputWriter := bufio.NewWriter(outputIoFile)

	t := template.Must(template.New("networkTemplate").Parse(string(networkTemplateContent)))
	if err = t.Execute(outputWriter, data); err != nil {
		Logger.Error("Cannot generate the network output config", "output", outputPath, "input", data)
		return err
	}
	if err = outputWriter.Flush(); err != nil {
		log.Panic("Cannot flush buffer writer to the generated network config file: ", err)
	}
	return nil
}
