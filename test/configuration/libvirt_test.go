package configuration

import (
	"encoding/xml"
	"freyja/internal/configuration"
	internalTest "freyja/test"
	"testing"
)

// TestValidDefaultNetworkConfiguration is used to test the default values in a configuration
const testValidCompleteNetworkConfiguration string = "../configuration/static/network_complete_conf.yaml"

func TestCreateLibvirtNetworkXMLDescription(t *testing.T) {
	config := internalTest.BuildConfig(testValidCompleteNetworkConfiguration)
	// <network>
	//  <name>ctrlplane</name>
	//  <ip address="192.168.123.1" netmask="255.255.255.0">
	//    <dhcp>
	//      <range start="192.168.123.2" end="192.168.123.254"/>
	//    </dhcp>
	//  </ip>
	//</network>
	network := config.Networks[0]
	contentBytes, err := configuration.CreateLibvirtNetworkXMLDescription(network)
	if err != nil {
		t.Fatal(err)
	}
	var description configuration.XMLNetworkDescription
	if err := xml.Unmarshal(contentBytes, &description); err != nil {
		t.Errorf("cannot unmarshall xml description")
		t.FailNow()
	}

	if description.Name != network.Name {
		t.Errorf("expected network name '%s' but got '%s'", network.Name, description.Name)
		t.Fail()
	}

	expectedGateway := "192.168.123.1"
	if description.Ip.Address != expectedGateway {
		t.Errorf("expected network gateway '%s' but got '%s'", description.Ip.Address, expectedGateway)
		t.Fail()
	}

	expectedNetmask := "255.255.255.0"
	if description.Ip.Netmask != expectedNetmask {
		t.Errorf("expected netmask '%s' but got '%s'", description.Ip.Address, expectedNetmask)
		t.Fail()
	}

	expectedDhcpStart := "192.168.123.2"
	if description.Ip.Dhcp.Range.Start != expectedDhcpStart {
		t.Errorf("expected dhcp start '%s' but got '%s'", description.Ip.Dhcp.Range.Start, expectedDhcpStart)
		t.Fail()
	}

	expectedDhcpEnd := "192.168.123.254"
	if description.Ip.Dhcp.Range.End != expectedDhcpEnd {
		t.Errorf("expected dhcp start '%s' but got '%s'", description.Ip.Dhcp.Range.End, expectedDhcpEnd)
		t.Fail()
	}

}
