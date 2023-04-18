import logging
from pathlib import Path
from typing import Dict, List, Optional

import typer
import yaml

from freyja.core.services.machine_service import create_machines, delete_machines, info_machines, \
    list_machines, \
    restart_machines, start_machines, \
    stop_machines, usage_machine, open_console_machine
from freyja.lib.exceptions.configuration_exceptions import ConfigurationContentError, \
    ConfigurationFileNotFoundException, \
    ConfigurationFormatError
from freyja.lib.utils.subprocess_utils import yes_no_question
from freyja.logger import FreyjaLogger

app = typer.Typer(help="Manage virtual machines")

logger: logging.Logger = logging.getLogger(FreyjaLogger.name)


def log_debug(verbose: bool):
    """
    Logger callback for more verbosity
    """
    if verbose:
        logger.setLevel(logging.DEBUG)


@app.command()
def create(configuration: Path = typer.Option(..., "-c", "--config",
                                              exists=True,
                                              file_okay=True,
                                              readable=True,
                                              resolve_path=True,
                                              help="Configuration file to create the virtual "
                                                   "machine"),
           foreground: bool = typer.Option(False, "--foreground",
                                           help="If enabled, start VMs in foreground and open "
                                                "consoles during creation"),
           dry: bool = typer.Option(False, "--dry-run",
                                    help="If enabled, skip startup"),
           verbose: bool = typer.Option(False, "-v", help="More verbosity for debug",
                                        callback=log_debug)):
    """
    Create one or more virtual machines based on a configuration file
    """
    logger.info(f"Create hosts")
    logger.debug(f"Using configuration {configuration}")
    try:
        create_machines(configuration, dry, foreground)
    except (ConfigurationFileNotFoundException, ConfigurationFormatError,
            ConfigurationContentError) as e:
        logger.fatal(f"Error in configuration: {str(e)}")
        raise typer.Exit(1)


@app.command()
def start(names: List[str] = typer.Argument(..., help="VM names list to start")):
    """
    Start virtual machines
    """
    start_machines(names)
    logger.info("OK")


@app.command()
def stop(names: List[str] = typer.Argument(..., help="VM names list to stop")):
    """
    Stop virtual machines
    """
    stop_machines(names)
    logger.info("OK")


@app.command()
def restart(names: List[str] = typer.Argument(..., help="VM names list to reboot")):
    """
    Reboot virtual machines
    """
    restart_machines(names)
    logger.info("OK")


@app.command()
def delete(names: List[str] = typer.Argument(None, help="VM names to delete"),
           del_all: bool = typer.Option(False, "-a", "--all", help="Delete all machines")):
    """
    Destroy and undefine one or more virtual machines based on names, or all
    """
    if del_all:
        names = list(filter(None, list_machines(names=True, stdout=False)))
    logger.warning(f"The following machines will be destroyed: {names}")
    if yes_no_question("Are you sure ? (Y/n)[default: n]", False):
        delete_machines(names)
        logger.info("OK")
    else:
        logger.info("Aborted")


@app.command(name="list")
def list_():
    """
    List the installed virtual machines
    """
    list_machines()


@app.command()
def info(names: Optional[List[str]] = typer.Argument(None, help="VM names list to describe. "
                                                                "Provide none to describe all.")):
    """
    Describe the virtual machines with various information. If no name is provided, display all.
    """
    info_model: Dict = info_machines(names)
    print(yaml.dump(info_model))


@app.command()
def usage(names: Optional[List[str]] = typer.Argument(None, help="VM names list to display their "
                                                                 "usage. Provide none to display "
                                                                 "all."),
          watch: bool = typer.Option(False, "-w", "--watch", help="Enable real time usage")):
    """
    Display the virtual machines cpu and memory usage.
    """
    usage_machine(names, watch)

@app.command()
def console(name: str = typer.Argument(..., help="VM name in which a console should be opened")):
    """
    Opens a console in the specified machine
    """
    open_console_machine(name)