package shellcli

import (
	"errors"
	"freyja/internal/shellcli"
	internalTest "freyja/test"
	"os"
	"path/filepath"
	"testing"
)

var TestNetworksDir string = filepath.Join(internalTest.FreyjaUnitTestDir, "/network-create")

// TestValidDefaultNetworkConfiguration is used to test the default values in a configuration
const testValidDefaultNetworkConfiguration string = "../configuration/static/default_conf.yaml"

// TestValidDefaultNetworkConfiguration is used to test the default values in a configuration
const testValidCompleteNetworkConfiguration string = "../configuration/static/network_complete_conf.yaml"

func TestGenerateLibvirtNetworksXMLDescriptions(t *testing.T) {
	validConfig := internalTest.BuildConfig(testValidCompleteNetworkConfiguration)
	// changing network names on the fly for testing purpose
	for i, network := range validConfig.Networks {
		network.Name = network.Name + "-test"
		validConfig.Networks[i] = network
	}

	xmlDescriptions, err := shellcli.GenerateLibvirtNetworksXMLDescriptions(validConfig, TestNetworksDir)
	if err != nil {
		t.Errorf("could not create xml descriptions from test config '%s': %v", testValidCompleteNetworkConfiguration, err)
	}

	if len(xmlDescriptions) != 2 {
		t.Errorf("expected 2 xml descriptions but got %d: '%v'", len(xmlDescriptions), xmlDescriptions)
		t.FailNow()
	}

	for _, network := range validConfig.Networks {
		networkDescriptionPath := shellcli.GetLibvirtNetworkDescriptionPath(network.Name)
		if _, err := os.Stat(networkDescriptionPath); errors.Is(err, os.ErrNotExist) {
			t.Errorf("expected xml description file in path '%s' but got none", networkDescriptionPath)
			t.Fail()
		}

		if err := os.RemoveAll(filepath.Dir(networkDescriptionPath)); err != nil {
			t.Fatalf("cannot remove test directory '%s'", networkDescriptionPath)
		}
	}

}
