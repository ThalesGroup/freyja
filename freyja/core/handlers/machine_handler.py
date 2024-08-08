import base64
import logging
import os
import shutil
from pathlib import Path
from typing import Dict

from jinja2 import Environment, FileSystemLoader

from freyja.core.provisioning.provisioning import Provisioning
from freyja.environment import FreyjaEnvironment
from freyja.lib.exceptions.configuration_exceptions import ConfigurationContentError
from freyja.lib.utils.file_utils import read_files, set_exec_permission, write_str
from freyja.lib.utils.subprocess_utils import execute
from freyja.logger import FreyjaLogger
from freyja.models.machine_configuration import MachineConfiguration

logger = logging.getLogger(FreyjaLogger.name)


class MachineHandler:
    IGNITION_KEY = "ignition"
    configuration: MachineConfiguration
    build_dir: Path
    provisioning: Provisioning
    hostname: str

    def __init__(self, configuration: MachineConfiguration, build_dir: Path):
        self.configuration = configuration
        self.hostname = configuration.hostname
        self.build_dir = build_dir
        self.launch_script = self.build_dir / FreyjaEnvironment.CREATE_VM_FILENAME
        self.init_provision()

    def init_provision(self):
        if self.configuration.ignition:
            ignition_file = self.configuration.ignition.file
            ignition_path = os.path.expandvars(ignition_file) if ignition_file else ignition_file
            self.provisioning = \
                Provisioning(template=FreyjaEnvironment.IGNITION_TEMPLATE_NAME,
                             output=self.build_dir / FreyjaEnvironment.IGNITION_FILENAME,
                             user_input=ignition_path)
        else:
            self.provisioning = Provisioning(template=FreyjaEnvironment.CLOUD_INIT_TEMPLATE_NAME,
                                             output=self.build_dir / FreyjaEnvironment.CLOUD_INIT_FILENAME)

    @staticmethod
    def _enrich_keys(configuration_dict: Dict):
        """
        Enrich the configuration with the given ssh public key paths
        :param configuration_dict: the dictionary of the machine configuration
        """
        for user_conf in configuration_dict.get('users'):
            if user_conf.get('keys'):
                user_conf["ssh_keys_contents"] = read_files(user_conf.get('keys'))

        return configuration_dict

    @staticmethod
    def _enrich_write_files(configuration_dict: Dict):
        """
        Enrich the configuration with the given files to write on the machine filesystem
        :param configuration_dict: the dictionary of the machine configuration
        """
        for file in configuration_dict.get("write_files"):
            file_path = Path(file.get("source"))
            try:
                with open(os.path.expandvars(file_path), "rb") as f:
                    content_b64 = base64.standard_b64encode(f.read())
                    file["content"] = content_b64.decode('ascii')
            except FileNotFoundError:
                raise ConfigurationContentError(f"write-files: file not found : "
                                                f"{str(file.get('source'))}")

        return configuration_dict

    def _enrich(self, foreground: bool):
        """
        Enrich a host configuration with various information for the cloud init conf file and the
        installation script
        :param foreground: enable console during machine creation
        :return: the enriched configuration
        """
        # USING DICT ON PURPOSE
        # We want to enrich the configuration with dynamic values
        configuration_dict = self.configuration.dict()
        # enrich with images root dir
        configuration_dict['image_dir'] = self.build_dir
        # enrich with provisioning
        configuration_dict['provisioning_file'] = self.provisioning.output
        # enrich with foreground vm startup mode
        configuration_dict['foreground'] = True if foreground else False
        # enrich with each key content
        if self.configuration.users:
            configuration_dict = MachineHandler._enrich_keys(configuration_dict)
        # write files content
        if self.configuration.write_files:
            configuration_dict = MachineHandler._enrich_write_files(configuration_dict)

        return configuration_dict

    def _render(self, enriched_configuration: Dict):
        """
        For each machine configuration, render the templates needed to create the VM in a
        dedicated directory.
        """
        jinja_env = Environment(loader=FileSystemLoader(FreyjaEnvironment.TEMPLATES_DIR), trim_blocks=True,
                                lstrip_blocks=True)

        # render installation script for handler
        install_script: str = jinja_env \
            .get_template(FreyjaEnvironment.CREATE_VM_TEMPLATE_NAME) \
            .render(enriched_configuration)
        write_str(install_script, self.launch_script)
        set_exec_permission(self.launch_script)

        # render the init file : cloud-init or ignition file
        if not self.provisioning.user_input:
            provision: str = jinja_env \
                .get_template(self.provisioning.template) \
                .render(enriched_configuration)
            # write files
            write_str(provision, self.provisioning.output)
        else:
            shutil.copyfile(self.provisioning.user_input, self.provisioning.output)

    def configure(self, foreground: bool):
        logger.debug(f"Configure machine {self.hostname} launch script and cloud init files in "
                     f"{self.build_dir}")
        enriched_configuration = self._enrich(foreground)
        self._render(enriched_configuration)

    def create(self):
        logger.debug(f"Create machine {self.hostname}")
        logger.debug(f"Execute {self.launch_script}")
        hosts = execute(["virsh", "list", "--all", "--name"])
        if self.hostname not in hosts:
            execute(["bash", self.launch_script.absolute()], stream_stdout=True)
        else:
            logger.debug(f"Skip: host {self.hostname} already exists")
