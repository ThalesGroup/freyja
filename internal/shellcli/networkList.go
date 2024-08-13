package shellcli

import (
	"freyja/internal"
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
		// logger
		Logger = internal.InitPrettyLogger()

		// networks list
		networks, err := getNetworksList()
		if err != nil {
			log.Panic("Could not list the networks", "error", err)
		}

		// interfaces list
		interfaces, err := getInterfacesList()
		if err != nil {
			log.Panic("Could not list the Interfaces", "error", err)
		}

		// print list in stdout
		printNetworksList(networks)
		printInterfacesList(interfaces)
	},
}

func getNetworksList() ([]libvirt.Network, error) {
	flags := libvirt.ConnectListNetworksActive | libvirt.ConnectListNetworksInactive
	networks, _, err := LibvirtConnexion.ConnectListAllNetworks(1, flags)
	return networks, err
}

func getInterfacesList() ([]libvirt.Interface, error) {
	flags := libvirt.ConnectListInterfacesActive | libvirt.ConnectListInterfacesInactive
	interfaces, _, err := LibvirtConnexion.ConnectListAllInterfaces(1, flags)
	return interfaces, err
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

// DomainNostate     DomainState = iota
// DomainRunning     DomainState = 1
// DomainBlocked     DomainState = 2
// DomainPaused      DomainState = 3
// DomainShutdown    DomainState = 4
// DomainShutoff     DomainState = 5
// DomainCrashed     DomainState = 6
// DomainPmsuspended DomainState = 7
func getNetworkState(state int32) string {
	switch state {
	case int32(libvirt.DomainRunning):
		return "running"
	case int32(libvirt.DomainBlocked):
		return "blocked"
	case int32(libvirt.DomainPaused):
		return "paused"
	case int32(libvirt.DomainShutdown):
		return "shutdown"
	case int32(libvirt.DomainShutoff):
		return "shutoff"
	case int32(libvirt.DomainCrashed):
		return "crashed"
	case int32(libvirt.DomainPmsuspended):
		return "suspended"
	}
	return "unknown"
}
