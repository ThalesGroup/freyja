package shellcli

import (
	"bytes"
	"crypto/ed25519"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/crypto/ssh"
)

func TestMarshalED25519KeyPair(t *testing.T) {
	comment := "freyja@freyja-ubuntu"
	privateKeyPEM, publicKeyLine, err := marshalED25519KeyPair(comment)
	if err != nil {
		t.Fatalf("marshalED25519KeyPair failed: %v", err)
	}

	// public key must be a valid authorized_keys line
	if !strings.HasPrefix(string(publicKeyLine), "ssh-ed25519 ") {
		t.Errorf("public key does not start with 'ssh-ed25519 ': %q", publicKeyLine)
	}
	if !strings.HasSuffix(string(publicKeyLine), "\n") {
		t.Errorf("public key line must end with a newline: %q", publicKeyLine)
	}
	publicKey, parsedComment, _, rest, err := ssh.ParseAuthorizedKey(publicKeyLine)
	if err != nil {
		t.Fatalf("public key is not a valid authorized_keys line: %v", err)
	}
	if len(rest) != 0 {
		t.Errorf("unexpected trailing data after public key: %q", rest)
	}
	if publicKey.Type() != ssh.KeyAlgoED25519 {
		t.Errorf("public key type = %q, want %q", publicKey.Type(), ssh.KeyAlgoED25519)
	}
	if parsedComment != comment {
		t.Errorf("public key comment = %q, want %q", parsedComment, comment)
	}

	// private key must be an OpenSSH PEM block
	if !strings.HasPrefix(string(privateKeyPEM), "-----BEGIN OPENSSH PRIVATE KEY-----") {
		t.Errorf("private key is not in OpenSSH PEM format: %q", privateKeyPEM)
	}
	rawPrivateKey, err := ssh.ParseRawPrivateKey(privateKeyPEM)
	if err != nil {
		t.Fatalf("private key cannot be parsed: %v", err)
	}
	if _, ok := rawPrivateKey.(*ed25519.PrivateKey); !ok {
		t.Fatalf("private key type = %T, want *ed25519.PrivateKey", rawPrivateKey)
	}

	// both files must belong to the same key pair
	signer, err := ssh.NewSignerFromKey(rawPrivateKey)
	if err != nil {
		t.Fatalf("cannot create signer from private key: %v", err)
	}
	if !bytes.Equal(signer.PublicKey().Marshal(), publicKey.Marshal()) {
		t.Error("public key does not match the private key")
	}
}

func TestMarshalED25519KeyPairIsRandom(t *testing.T) {
	_, firstPublicKey, err := marshalED25519KeyPair("a@b")
	if err != nil {
		t.Fatal(err)
	}
	_, secondPublicKey, err := marshalED25519KeyPair("a@b")
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(firstPublicKey, secondPublicKey) {
		t.Error("two generated key pairs are identical")
	}
}

func TestWriteFileWithModeFixesExistingFileMode(t *testing.T) {
	path := filepath.Join(t.TempDir(), "key")
	// simulate a leftover key file with too open permissions
	if err := os.WriteFile(path, []byte("old"), 0666); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0666); err != nil {
		t.Fatal(err)
	}

	if err := writeFileWithMode(path, []byte("new"), SSHPrivateKeyFileMode); err != nil {
		t.Fatalf("writeFileWithMode failed: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != SSHPrivateKeyFileMode {
		t.Errorf("file mode = %o, want %o", info.Mode().Perm(), SSHPrivateKeyFileMode)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "new" {
		t.Errorf("file content = %q, want %q", content, "new")
	}
}
