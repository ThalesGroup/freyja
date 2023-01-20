import logging
import uuid
from pathlib import Path
from typing import Dict

from jinja2 import Environment, FileSystemLoader

from freyja.environment import FreyjaEnvironment
from freyja.lib.utils.error_utils import check_message
from freyja.lib.utils.file_utils import write_str
from freyja.lib.utils.subprocess_utils import execute
from freyja.logger import FreyjaLogger
from freyja.models.network_configuration import NetworkConfiguration

logger = logging.getLogger(FreyjaLogger.name)


class NetworkHandler:
    NETWORK_TEMPLATE_NAME = "network.xml.j2"
    ADDRESS_TEMPLATE_NAME = "address.xml.j2"
    environment: FreyjaEnvironment
    name: str
    configuration: Dict
    configuration_file: Path

    def __init__(self, configuration: NetworkConfiguration, build_dir: Path):
        self.name = configuration.name
        self.configuration = configuration.dict()
        self.configuration_file = build_dir / f"{configuration.name}_{configuration.address}_" \
                                              f"{FreyjaEnvironment.NETWORK_FILENAME}"

    def _enrich(self):
        """
        Get the user's networks configurations and for each one of them, add needed information
        to create them in libvirt
        :return: the enriched networks configuration
        """
        self.configuration['uuid'] = uuid.uuid4()

    def _render(self):
        """
        For each network configuration, render the templates needed to create the network in libvirt
        """
        jinja_env = Environment(loader=FileSystemLoader(FreyjaEnvironment.TEMPLATES_DIR), trim_blocks=True,
                                lstrip_blocks=True)
        # network
        rendered_network_conf: str = jinja_env \
            .get_template(self.NETWORK_TEMPLATE_NAME) \
            .render(self.configuration)
        write_str(rendered_network_conf, self.configuration_file)

    def configure(self):
        logger.debug(f"Configure network {self.name} in {self.configuration_file}")
        self._enrich()
        self._render()

    def create(self):
        logger.debug(f"Create network {self.name}")
        try:
            execute(["virsh", "net-define", self.configuration_file.absolute()])
            execute(["virsh", "net-autostart", self.name])
            execute(["virsh", "net-start", self.name])
        except ChildProcessError as e:
            if check_message(e, "network 'ctrl-plane' already exists"):
                logger.debug(f"Skip: network {self.name} already exists")
