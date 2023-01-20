from pathlib import Path
from typing import Dict

import pytest
import yaml

from freyja.configuration import Configuration
from freyja.lib.exceptions.configuration_exceptions import ConfigurationContentError, \
    ConfigurationFileNotFoundException
from freyja.tests.common import RESOURCES_DIR

SIMPLE_CONF = f"{RESOURCES_DIR}/simple_conf.yaml"
DETAILED_CONF = f"{RESOURCES_DIR}/detailed_conf.yaml"
INVALID_CONF = f"{RESOURCES_DIR}/invalid_conf.yaml"
IGNITION_CONF = f"{RESOURCES_DIR}/ignition_conf.yaml"


def read_yaml(file: Path) -> "Dict":
    with open(file, "r") as stream:
        return yaml.safe_load(stream)


def compare(conf: Path, test_default: bool):
    """
    Compare a built model with a raw content, both from the same file
    :param test_default: provide if the input conf contains default values to test
    :param conf: conf path
    """
    input_conf = read_yaml(conf)  # raw dict from file
    output_model = Configuration.parse_file(conf)  # built model from file
    assert output_model

    # host values
    input_content = input_conf.get("hosts")
    output_hosts = output_model.hosts
    assert output_hosts
    assert len(output_hosts) == len(input_content)
    input_host = input_content[0]
    output_host = output_hosts[0]
    assert output_host.hostname == input_host.get("hostname")
    assert output_host.image == input_host.get("image")
    assert output_host.os == input_host.get("os")
    if test_default:
        assert output_host.disk == 30
        assert output_host.memory == 4096
        assert output_host.vcpus == 2
        assert not output_host.packages
    else:
        assert output_host.disk == input_host.get("disk")
        assert output_host.memory == input_host.get("memory")
        assert output_host.vcpus == input_host.get("vcpus")
        assert output_host.packages == input_host.get("packages")

    # network values
    input_networks = input_host.get("networks")
    output_networks = output_host.networks
    assert len(output_networks) == len(input_networks)

    input_network = input_networks[0]
    output_network = output_networks[0]
    assert output_network.name == input_network.get("name")
    assert output_network.address == input_network.get("address")

    # user values
    input_users = input_host.get("users")
    output_users = output_host.users
    assert output_users

    output_user = output_users[0]
    if test_default:
        assert output_user.username == "freyja"
        assert output_user.password == '$6$GM./aNJikL/g$AR2c35i1QIaimKo/zOC/1Qg2JO65ysPPjv/leWBcgBXaxNV3V8IcgJVeTzt4VHWzcja66zsBnR1iyYtO2DPme/'
        assert not output_user.keys
    else:
        input_user = input_users[0]
        assert output_user.username == input_user.get("username")
        assert output_user.password == input_user.get("password")
        assert output_user.keys == input_user.get("keys")


def test_configuration_parse_file(tmp_path):
    # missing
    with pytest.raises(ConfigurationFileNotFoundException):
        Configuration.parse_file(Path("/dev/null"))

    # invalid
    with pytest.raises(ConfigurationContentError):
        Configuration.parse_file(Path(INVALID_CONF))

    # proper
    compare(Path(SIMPLE_CONF), test_default=True)
    compare(Path(DETAILED_CONF), test_default=False)


def test_ignition_configuration():
    conf = Path(IGNITION_CONF)

    input_conf = read_yaml(conf)
    output_model = Configuration.parse_file(conf)

    input_host = input_conf.get("hosts")[0]
    output_host = output_model.hosts[0]

    assert output_host.ignition
    assert output_host.ignition.version == input_host.get("ignition").get("version")
    assert output_host.ignition.file == Path(input_host.get("ignition").get("file"))
