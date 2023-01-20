import logging
from typing import List, Optional

import typer

from freyja.core.services.network_service import delete_networks, info_networks, list_networks
from freyja.lib.utils.subprocess_utils import yes_no_question
from freyja.logger import FreyjaLogger

app = typer.Typer(help="Gives information about virtual networks")

logger = logging.getLogger(FreyjaLogger.name)


def log_debug(verbose: bool):
    """
    Logger callback for more verbosity
    """
    if verbose:
        logger.setLevel(logging.DEBUG)


@app.command(name="list")
def list_():
    """
    List the created networks in libvirt
    """
    list_networks()


@app.command()
def info(names: Optional[List[str]] = typer.Argument(None, help="Networks names list to describe. "
                                                                "Provide none to describe all.")):
    """
    Describe networks for virtual machines
    """
    info_networks(names)


@app.command()
def delete(names: List[str] = typer.Argument(None, help="Network names to delete"),
           del_all: bool = typer.Option(False, "-a", "--all", help="Delete all machines")):
    """
    Destroy and undefine one or more virtual machines based on names, or all
    """
    if del_all:
        names = list(filter(None, list_networks(names=True, stdout=False)))
    logger.warning(f"The following networks will be destroyed: {names}")
    if yes_no_question("Are you sure ? (Y/n)[default: n]", False):
        delete_networks(names)
        logger.info("OK")
    else:
        logger.info("Aborted")
