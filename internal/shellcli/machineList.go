package shellcli

import (
	"github.com/aquasecurity/table"
	"github.com/digitalocean/go-libvirt"
	"github.com/spf13/cobra"
	"log"
	"os"
)

// commands definitions
var machineListCmd = &cobra.Command{
	Use:              "list",
	Short:            "List machines",
	Long:             "List machines using handler",
	TraverseChildren: true, // ensure local flags do not spread to sub commands

	Run: func(cmd *cobra.Command, args []string) {
		// domains list
		domains, err := getDomainsList()
		if err != nil {
			log.Panic("Could not list the machines", "error", err)
		}

		// print list in stdout
		printMachinesList(domains)
	},
}

func getDomainsList() ([]libvirt.Domain, error) {
	flags := libvirt.ConnectListDomainsActive | libvirt.ConnectListDomainsInactive
	domains, _, err := LibvirtConnexion.ConnectListAllDomains(1, flags)
	return domains, err
}

func printMachinesList(domains []libvirt.Domain) {
	// init table
	t := table.New(os.Stdout)
	t.SetRowLines(false)
	t.SetBorders(false)
	t.SetHeaders("Name", "Status")
	// for each domain
	for _, d := range domains {
		state, _, _ := LibvirtConnexion.DomainGetState(d, 0)
		t.AddRow(d.Name, getMachineState(state))
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
func getMachineState(state int32) string {
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
