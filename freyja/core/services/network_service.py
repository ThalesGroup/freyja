import logging
import xml.etree.ElementTree as ET
from pathlib import Path
from typing import List

import yaml

from freyja.lib.utils.error_utils import check_message
from freyja.lib.utils.subprocess_utils import execute
from freyja.lib.utils.virsh_utils import parse_info
from freyja.logger import FreyjaLogger

logger = logging.getLogger(FreyjaLogger.name)


def get_network_name(config: Path) -> str:
    """
    :param config: XML configuration file path
    :return: the name of the configured network
    """
    tree = ET.parse(config)
    root = tree.getroot()
    return root.find('name').text


def create_network(config: Path):
    network_name = get_network_name(config)
    logger.info(f"Create network '{network_name}'")
    execute(["virsh", "net-define", str(config.absolute())])
    execute(["virsh", "net-autostart", network_name])
    execute(["virsh", "net-start", network_name])


def delete_networks(names: List[str]):
    for name in names:
        execute(["virsh", "net-destroy", name])
        execute(["virsh", "net-undefine", name])


def list_networks(names: bool = False, stdout: bool = True) -> "List[str]":
    """
    :param names: if True, print only names
    :param stdout: if True, stream the stdout output
    """
    cmd = ["virsh", "net-list", "--all"]
    if names:
        cmd.append("--name")
    return execute(cmd, stream_stdout=stdout)


def get_net_info(net: str):
    output = execute(["virsh", "net-info", net], stream_stdout=False)
    return parse_info(output)


def info_networks(names: List[str]):
    net_list: List[str] = names if names else \
        list(filter(None, list_networks(names=True, stdout=False)))
    result = {}
    for net in net_list:
        try:
            result[net] = get_net_info(net)
        except ChildProcessError as e:
            if check_message(e, "no network with matching name"):
                logger.warning(f"Skip {net}: network not found")
            else:
                raise e

    print(yaml.dump(result))

