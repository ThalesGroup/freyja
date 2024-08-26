package shellcli

import (
	"freyja/internal/shellcli"
	"regexp"
	"testing"
)

func TestNetworkCreate2(t *testing.T) {
	// generate conf
	networkName := "network-test"
	config := shellcli.NetworkCreateConfig2("network-test")

	if config.Name != networkName {
		t.Logf("expected network name '%s' but got '%s'", networkName, config.Name)
		t.Fail()
	}

	uuidRegex := "^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$"
	match, _ := regexp.MatchString(uuidRegex, config.UUID)
	if !match {
		t.Logf("expected uuid pattern '%s' but got '%s'", uuidRegex, config.UUID)
		t.Fail()
	}

	if config.Forward.Mode != "bridge" {
		t.Logf("expected forward mode 'bridge' but got '%s'", config.Forward.Mode)
		t.Fail()
	}

	if config.Bridge.Name != "virbr0" {
		t.Logf("expected forward mode 'virbr0' but got '%s'", config.Bridge.Name)
		t.Fail()
	}

}
