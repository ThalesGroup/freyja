package shellcli

import (
	"embed"
	"fmt"
	"github.com/digitalocean/go-libvirt"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"net/url"
	"os"
)

// Flags
var verbose bool

// Logger Freyja logger
var Logger *zap.Logger

// LibvirtConnexion qemu connexion
var LibvirtConnexion *libvirt.Libvirt

//go:embed templates
var templates embed.FS

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
		setLogger()
	},
	PostRun: func(cmd *cobra.Command, args []string) {
		// Will be call at the end of any subcommand
		// Do your processing here

	},
}

func setLogger() {
	level := zap.InfoLevel
	if verbose {
		level = zap.DebugLevel
	}
	config := zap.Config{
		Level:             zap.NewAtomicLevelAt(level),
		Development:       false,
		DisableCaller:     true,
		DisableStacktrace: true,
		Sampling:          nil,
		Encoding:          "console",
		EncoderConfig:     zap.NewDevelopmentEncoderConfig(),
		OutputPaths:       []string{"stdout", "/tmp/golang_starter-tutorial-logger-zap.log"},
		ErrorOutputPaths:  []string{"stderr", "/tmp/golang_starter-tutorial-logger-zap-error.log"},
		InitialFields:     nil,
	}
	// build the new custom logger
	Logger = zap.Must(config.Build())

}

func initLibvirtConnexion() *libvirt.Libvirt {
	// qemu connexion initialization
	uri, err := url.Parse(string(libvirt.QEMUSystem))
	if err != nil {
		Logger.Panic("Cannot parse Qemu system URI",
			zap.Error(err))
	}
	connexion, err := libvirt.ConnectToURI(uri)
	if err != nil {
		Logger.Panic("Could not open Qemu connexion",
			zap.Error(err))
	}
	return connexion
}

// this function is called before all

func init() {
	// Commands
	// machine management
	rootCmd.AddCommand(machineCmd)
	// network management
	rootCmd.AddCommand(networkCmd)

	// Flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Activate debug mode") // version flag only for root command

	// Init libvirt connexion before any commands or flags
	LibvirtConnexion = initLibvirtConnexion()
}

func finalize() {
	// Finalize libvirt connexion
	if err := LibvirtConnexion.Disconnect(); err != nil {
		Logger.Error("Could not close Qemu connexion",
			zap.Error(err))
	}

	// Finalize logger
	if Logger != nil {
		Logger.Sync()
	}
}

// Execute is the entry point of the cli
// You can call it from external packages
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	finalize()
}
