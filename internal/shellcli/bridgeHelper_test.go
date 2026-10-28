package shellcli

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeACL writes an ACL file in a temp dir and returns its path
func writeACL(t *testing.T, dir string, name string, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestAuditBridgeHelperSetup(t *testing.T) {
	tests := []struct {
		name      string
		acl       string
		bridges   []string
		wantError string // empty means no error expected
	}{
		{
			name:    "freyja recommended setup",
			acl:     "allow virbr0\nallow freyja0\nallow freyja1\n",
			bridges: []string{"virbr0", "freyja1"},
		},
		{
			name:    "comments and blank lines",
			acl:     "# freyja bridges\n\n   # indented comment\nallow virbr0\n",
			bridges: []string{"virbr0"},
		},
		{
			name:    "allow all",
			acl:     "allow all\n",
			bridges: []string{"virbr0", "freyja3"},
		},
		{
			name:      "bridge not allowed",
			acl:       "allow virbr0\n",
			bridges:   []string{"virbr0", "freyja0"},
			wantError: "[freyja0]",
		},
		{
			name:      "empty file",
			acl:       "",
			bridges:   []string{"virbr0"},
			wantError: "[virbr0]",
		},
		{
			name:      "deny wins over allow whatever the order",
			acl:       "deny freyja0\nallow all\n",
			bridges:   []string{"freyja0"},
			wantError: "[freyja0]",
		},
		{
			name:      "deny all",
			acl:       "allow virbr0\ndeny all\n",
			bridges:   []string{"virbr0"},
			wantError: "[virbr0]",
		},
		{
			name:    "deny of another bridge",
			acl:     "allow all\ndeny br0\n",
			bridges: []string{"virbr0"},
		},
		{
			name:    "duplicated and unresolved bridges",
			acl:     "allow virbr0\n",
			bridges: []string{"virbr0", "virbr0", unresolvedBridgeName},
		},
		{
			name:      "duplicated bridge reported once",
			acl:       "allow virbr0\n",
			bridges:   []string{"freyja0", "freyja0"},
			wantError: "[freyja0]",
		},
		{
			name:      "unknown rule",
			acl:       "permit virbr0\n",
			bridges:   []string{"virbr0"},
			wantError: "unknown rule 'permit'",
		},
		{
			name:      "malformed line",
			acl:       "allow\n",
			bridges:   []string{"virbr0"},
			wantError: "line 1",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			aclPath := writeACL(t, t.TempDir(), "bridge.conf", test.acl)
			err := auditBridgeHelperSetup(aclPath, test.bridges)
			if test.wantError == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected an error containing %q, got none", test.wantError)
			}
			if !strings.Contains(err.Error(), test.wantError) {
				t.Errorf("error %q does not contain %q", err.Error(), test.wantError)
			}
		})
	}
}

func TestAuditBridgeHelperSetupMissingFile(t *testing.T) {
	err := auditBridgeHelperSetup(filepath.Join(t.TempDir(), "bridge.conf"), []string{"virbr0"})
	if err == nil {
		t.Fatal("expected an error for a missing ACL file, got none")
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("error %q should wrap os.ErrNotExist", err.Error())
	}
}

func TestAuditBridgeHelperSetupUnreadableFile(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root can read any file")
	}
	aclPath := writeACL(t, t.TempDir(), "bridge.conf", "allow virbr0\n")
	if err := os.Chmod(aclPath, 0); err != nil {
		t.Fatal(err)
	}
	// the setuid helper may read it while the user cannot: freyja cannot verify, so it only warns
	if err := auditBridgeHelperSetup(aclPath, []string{"freyja0"}); err != nil {
		t.Fatalf("an unreadable ACL file should only be warned about, got error: %v", err)
	}
}

func TestAuditBridgeHelperSetupInclude(t *testing.T) {
	dir := t.TempDir()
	includedPath := writeACL(t, dir, "freyja.conf", "allow freyja0\n")
	aclPath := writeACL(t, dir, "bridge.conf", "allow virbr0\ninclude "+includedPath+"\n")
	if err := auditBridgeHelperSetup(aclPath, []string{"virbr0", "freyja0"}); err != nil {
		t.Fatalf("bridge allowed by an included file was rejected: %v", err)
	}

	// a missing included file is an error for qemu-bridge-helper too
	brokenPath := writeACL(t, dir, "broken.conf", "include "+filepath.Join(dir, "missing.conf")+"\n")
	if err := auditBridgeHelperSetup(brokenPath, []string{"virbr0"}); err == nil {
		t.Fatal("expected an error for a missing included file, got none")
	}
}

func TestAuditBridgeHelperSetupIncludeLoop(t *testing.T) {
	dir := t.TempDir()
	aclPath := filepath.Join(dir, "bridge.conf")
	writeACL(t, dir, "bridge.conf", "allow virbr0\ninclude "+aclPath+"\n")
	err := auditBridgeHelperSetup(aclPath, []string{"virbr0"})
	if err == nil || !strings.Contains(err.Error(), "nested 'include'") {
		t.Fatalf("expected an include loop error, got: %v", err)
	}
}
