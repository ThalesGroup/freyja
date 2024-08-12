package shellcli

import (
	"embed"
	"freyja/internal"
	"github.com/digitalocean/go-libvirt"
	"github.com/spf13/cobra"
	"log"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
)

// Flags
var verbose bool

// Logger Freyja logger
var Logger *slog.Logger

// LibvirtConnexion qemu connexion
var LibvirtConnexion *libvirt.Libvirt

var FreyjaWorkspaceDir = filepath.Join(os.Getenv("HOME"), "freyja-workspace")

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
		Logger = internal.InitLogger()
	},
	PostRun: func(cmd *cobra.Command, args []string) {
		// Will be call at the end of any subcommand
		// Do your processing here

	},
}

func initLibvirtConnexion() *libvirt.Libvirt {
	// qemu connexion initialization
	uri, err := url.Parse(string(libvirt.QEMUSystem))
	if err != nil {
		log.Panic("Cannot parse Qemu system URI: ", err)
	}
	connexion, err := libvirt.ConnectToURI(uri)
	if err != nil {
		log.Panic("Could not open Qemu connexion: ", err)
	}
	return connexion
}

func initWorkspace() {
	// check the freyja workspace dir
	// create if it does not exist
	if err := os.MkdirAll(FreyjaWorkspaceDir, os.ModePerm); err != nil {
		log.Panic("Could not create freyja workspace dir in home user: ", err)
	}
}

// this function is called before all

func init() {
	// Commands
	// machine management
	rootCmd.AddCommand(MachineCmd)
	// network management
	rootCmd.AddCommand(NetworkCmd)

	// Flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Activate debug mode") // version flag only for root command

	// Init libvirt connexion before any commands or flags
	LibvirtConnexion = initLibvirtConnexion()

	// Init freyja workspace dir in home user
	initWorkspace()
}

func finalize() {
	// Finalize libvirt connexion
	if err := LibvirtConnexion.Disconnect(); err != nil {
		log.Panic("Could not close Qemu connexion: ", err)
	}
}

// Execute is the entry point of the cli
// You can call it from external packages
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Println(os.Stderr, err)
		os.Exit(1)
	}
	finalize()
}
