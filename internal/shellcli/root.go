package shellcli

import (
	"embed"
	"fmt"
	"freyja/internal"
	"freyja/internal/configuration"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/digitalocean/go-libvirt"
	"github.com/digitalocean/go-libvirt/socket/dialers"
	"github.com/spf13/cobra"
)

var Logger = internal.Logger

// Flags
var verbose bool

// LibvirtDomainConnexion qemu session connexion, used for the domains: the machines run
// unprivileged, as the user who launched freyja, so their storage can stay in $HOME
var LibvirtDomainConnexion *libvirt.Libvirt

// LibvirtNetworkConnexion qemu system connexion, used for the virtual networks only.
// A virtual network cannot run in the unprivileged session daemon (creating the
// bridge, the dnsmasq process and the firewall rules requires root), so all the
// networks are managed by the single privileged system daemon and shared between
// the users. This also rules out IP conflicts between users: each network has one
// DHCP server delivering unique leases, and libvirt refuses to start a network
// whose subnet collides with an existing route.
// The user must be a member of the 'libvirt' group to open this connexion.
var LibvirtNetworkConnexion *libvirt.Libvirt

//go:embed templates
var Templates embed.FS

const networkTemplate string = "templates/network.xml.tmpl"

// rootCmd is the root command definitions
// define here the helper and the root command flags behavior
var rootCmd = &cobra.Command{
	Use:              "freyja",
	Long:             "Freyja shell client",
	TraverseChildren: true, // ensure local flags do not spread to sub commands

	Run: func(cmd *cobra.Command, args []string) {
		// Use this function for code logic after commands and flags initialization
		// Only called if the root command is called only.
		// Is overridden by 'Run' function of subcommands calls.
		// Do you processing here
		// Like command annotations
	},
	PostRun: func(cmd *cobra.Command, args []string) {
		// Will be call at the end of any subcommand
		// Do your processing here

	},
}

// sessionSocketPath returns the path of the current user's libvirt session socket.
// The socket, not the URI name, is what decides between qemu:///session and
// qemu:///system: go-libvirt's default local dialer targets the privileged
// /var/run/libvirt/libvirt-sock, so we must explicitly dial the per-user one.
func sessionSocketPath() string {
	runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
	if runtimeDir == "" {
		runtimeDir = fmt.Sprintf("/run/user/%d", os.Getuid())
	}
	socketDir := filepath.Join(runtimeDir, "libvirt")
	// modular daemon socket first, then monolithic daemon socket
	for _, name := range []string{"virtqemud-sock", "libvirt-sock"} {
		socketPath := filepath.Join(socketDir, name)
		if _, err := os.Stat(socketPath); err == nil {
			return socketPath
		}
	}
	return filepath.Join(socketDir, "libvirt-sock")
}

// startSessionDaemon spawns the user's libvirt session daemon if its socket does
// not exist yet. Virsh auto-spawns it on 'virsh -c qemu:///session' but go-libvirt
// only dials sockets, so freyja has to launch it the same way virsh would.
func startSessionDaemon() (string, error) {
	socketPath := sessionSocketPath()
	if _, err := os.Stat(socketPath); err == nil {
		return socketPath, nil
	}
	var daemon string
	for _, candidate := range []string{"virtqemud", "libvirtd"} {
		if path, err := exec.LookPath(candidate); err == nil {
			daemon = path
			break
		}
	}
	if daemon == "" {
		return "", fmt.Errorf("no session socket at '%s' and neither 'virtqemud' nor 'libvirtd' found in PATH", socketPath)
	}
	Logger.Debug("starting libvirt session daemon", "daemon", daemon)
	cmd := exec.Command(daemon, "--daemon", "--timeout=120")
	// a failed spawn is not fatal by itself: a concurrent freyja process may have
	// won the race and started the daemon first, in which case the socket shows up
	spawnErr := cmd.Run()
	// the daemon forks: wait for the session socket to show up
	socketPath = sessionSocketPath()
	for i := 0; i < 50; i++ {
		if _, err := os.Stat(socketPath); err == nil {
			return socketPath, nil
		}
		time.Sleep(100 * time.Millisecond)
		socketPath = sessionSocketPath()
	}
	if spawnErr != nil {
		return "", fmt.Errorf("cannot start libvirt session daemon '%s': %w", daemon, spawnErr)
	}
	return "", fmt.Errorf("libvirt session daemon started but socket '%s' did not appear", socketPath)
}

func initLibvirtConnexion() *libvirt.Libvirt {
	// qemu session connexion initialization
	socketPath, err := startSessionDaemon()
	if err != nil {
		log.Panic("Could not reach the libvirt session daemon: ", err)
	}
	connexion := libvirt.NewWithDialer(dialers.NewLocal(dialers.WithSocket(socketPath)))
	if err := connexion.ConnectToURI(libvirt.QEMUSession); err != nil {
		log.Panic("Could not open Qemu session connexion: ", err)
	}
	return connexion
}

func initLibvirtNetworkConnexion() *libvirt.Libvirt {
	// the default local dialer targets the privileged system socket
	// /var/run/libvirt/libvirt-sock, which the members of the 'libvirt' group are
	// allowed to reach
	connexion := libvirt.NewWithDialer(dialers.NewLocal())
	if err := connexion.ConnectToURI(libvirt.QEMUSystem); err != nil {
		log.Panic("Could not open Qemu system connexion for the networks (the user must be a member of the 'libvirt' group): ", err)
	}
	return connexion
}

func initWorkspace() {
	// check the freyja workspace dir
	// create if it does not exist
	if err := os.MkdirAll(configuration.FreyjaWorkspaceDir, os.ModePerm); err != nil {
		log.Panic("Could not create freyja workspace dir in home user: ", err)
	}
}

// this function is called before all

func init() {
	// Commands
	// machine management
	rootCmd.AddCommand(MachineCmd)
	// network management
	rootCmd.AddCommand(networkCmd)

	// Flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Activate debug mode") // version flag only for root command

	// Init libvirt connexions before any commands or flags
	LibvirtDomainConnexion = initLibvirtConnexion()
	LibvirtNetworkConnexion = initLibvirtNetworkConnexion()

	// Init freyja workspace dir in home user
	initWorkspace()
}

func finalize() {
	// Finalize libvirt connexions
	if err := LibvirtDomainConnexion.Disconnect(); err != nil {
		log.Panic("Could not close Qemu session connexion: ", err)
	}
	if err := LibvirtNetworkConnexion.Disconnect(); err != nil {
		log.Panic("Could not close Qemu system connexion: ", err)
	}
}

// Execute is the entry point of the cli
// You can call it from external packages
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
	finalize()
}
