import logging
from pathlib import Path
from typing import List

from freyja.core.handlers.machine_handler import MachineHandler
from freyja.core.handlers.network_handler import NetworkHandler
from freyja.environment import FreyjaEnvironment
from freyja.logger import FreyjaLogger
from freyja.models.machine_configuration import MachineConfiguration

logger: logging.Logger = logging.getLogger(FreyjaLogger.name)


class HostHandler:
    """
    Host handler has:
     - One machine handler
     - Multiple networks handlers
    """
    build_dir: Path
    configuration: MachineConfiguration
    machine_handler: MachineHandler
    network_handlers: List[NetworkHandler]

    def __init__(self, configuration: MachineConfiguration):
        self.build_dir = Path(FreyjaEnvironment.BUILD_DIR) / configuration.hostname
        self.configuration = configuration

    def configure_machine(self, foreground: bool):
        self.machine_handler = MachineHandler(self.configuration, self.build_dir)
        self.machine_handler.configure(foreground)

    def configure_networks(self):
        handlers: List[NetworkHandler] = []
        if self.configuration.networks:
            for network_conf in self.configuration.networks:
                handler = NetworkHandler(network_conf, self.build_dir)
                handler.configure()
                handlers.append(handler)

        self.network_handlers = handlers

    def configure(self, foreground: bool):
        self.build_dir.mkdir(parents=True, exist_ok=True)
        self.configure_networks()
        self.configure_machine(foreground)

    def create_networks(self):
        for handler in self.network_handlers:
            handler.create()

    def create_machines(self):
        self.machine_handler.create()

    def launch(self):
        self.create_networks()
        self.create_machines()
