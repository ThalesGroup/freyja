package shellcli

import (
	"log"

	"github.com/spf13/cobra"
)

var sshDomainName string

var sshUserName string

var sshPassword bool

// commands definitions
var machineSSHCmd = &cobra.Command{
	Use:              "ssh",
	Short:            "Open an SSH terminal for a machine",
	Long:             "Open an SSH terminal for a machine with a configured user and SSH keys generated in the freyja machine's directory.",
	TraverseChildren: true, // ensure local flags do not spread to sub commands

	Run: func(cmd *cobra.Command, args []string) {
		// domains list
		Logger.Info("Opening machine SSH session", "machine", sshDomainName)

		// todo implement ssh session with golang
		//   https://www.pixelstech.net/article/1699714722-guide-to-implement-an-ssh-client-using-golang

		//// configure key authentication
		//key, err := os.ReadFile("/tmp/mycert")
		//if err != nil {
		//	log.Fatalf("unable to read private key: %v", err)
		//}
		//
		//// Create the Signer for this private key.
		//signer, err := ssh.ParsePrivateKey(key)
		//if err != nil {
		//	log.Fatalf("unable to parse private key: %v", err)
		//}
		//
		//// Load the certificate
		//cert, err := ioutil.ReadFile("/tmp/mycert-cert.pub")
		//if err != nil {
		//	log.Fatalf("unable to read certificate file: %v", err)
		//}
		//
		//pk, _, _, _, err := ssh.ParseAuthorizedKey(cert)
		//if err != nil {
		//	log.Fatalf("unable to parse public key: %v", err)
		//}
		//
		//certSigner, err := ssh.NewCertSigner(pk.(*ssh.Certificate), signer)
		//if err != nil {
		//	log.Fatalf("failed to create cert signer: %v", err)
		//}
		//
		//sshAuth := []ssh.AuthMethod{
		//	ssh.PublicKeys(certSigner),
		//}

		// create the client
		//sshConfig := &ssh.ClientConfig{
		//	Config: ssh.Config{},
		//	User:   sshUserName,
		//	// sshAuth is nil if no password was provided
		//	Auth:              sshAuth,
		//	HostKeyCallback:   ssh.FixedHostKey(),
		//	BannerCallback:    nil,
		//	ClientVersion:     "",
		//	HostKeyAlgorithms: nil,
		//	Timeout:           0,
		//}
	},
}

func init() {
	// --name,-n [machine name]
	machineSSHCmd.Flags().StringVarP(&sshDomainName, "name", "n", "", "Name of the target machine for the SSH session.")
	if err := machineSSHCmd.MarkFlagRequired("name"); err != nil {
		log.Panic(err.Error())
	}
	// --user,-u [user name]
	machineSSHCmd.Flags().StringVarP(&sshUserName, "user", "u", "freyja", "Name of the user for the SSH session.")
	// --password, -p
	machineSSHCmd.Flags().BoolVarP(&sshPassword, "password", "p", false, "Enable password input from terminal. Otherwise, no password is used for authentication.")
}
