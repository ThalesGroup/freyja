package shellcli

import (
	"github.com/aquasecurity/table"
	"github.com/digitalocean/go-libvirt"
	"github.com/spf13/cobra"
	"log"
	"os"
)

// commands definitions
var networkListCmd = &cobra.Command{
	Use:              "list",
	Short:            "List virtual network",
	Long:             "List virtual network using libvirt",
	TraverseChildren: true, // ensure local flags do not spread to sub commands

	Run: func(cmd *cobra.Command, args []string) {
		networks, err := getNetworksList()
		if err != nil {
			log.Panic("Could not list the networks", "error", err)
		}

		// print list in stdout
		printNetworksList(networks)
	},
}

func getNetworksList() ([]libvirt.Network, error) {
	flags := libvirt.ConnectListNetworksActive | libvirt.ConnectListNetworksInactive
	networks, _, err := LibvirtConnexion.ConnectListAllNetworks(1, flags)
	return networks, err
}

func printNetworksList(domains []libvirt.Network) {
	// init table
	t := table.New(os.Stdout)
	t.SetRowLines(false)
	t.SetBorders(false)
	t.SetHeaders("Name")
	// for each domain
	for _, n := range domains {
		//state, _, _ := LibvirtConnexion.DomainGetState(d, 0)
		t.AddRow(n.Name)

	}
	t.Render()
}

func printInterfacesList(domains []libvirt.Interface) {
	// init table
	t := table.New(os.Stdout)
	t.SetRowLines(false)
	t.SetBorders(false)
	t.SetHeaders("Name", "Mac")
	// for each domain
	for _, i := range domains {
		//state, _, _ := LibvirtConnexion.DomainGetState(d, 0)
		t.AddRow(i.Name, i.Mac)

	}
	t.Render()
}
