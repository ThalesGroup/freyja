import logging

import typer

from freyja.core.services.snapshot_service import restore_snapshot, create_snapshot, list_snapshot, delete_snapshot
from freyja.lib.utils.subprocess_utils import yes_no_question
from freyja.logger import FreyjaLogger

app = typer.Typer(help="Manage snapshots of virtual machines")

logger: logging.Logger = logging.getLogger(FreyjaLogger.name)


@app.command()
def restore(name: str = typer.Argument(..., help="VM name to restore"),
            snapshot: str = typer.Argument(..., help="Name of the snapshot")):
    """
    Restore a VM from a snapshot
    """
    restore_snapshot(name, snapshot)
    logger.warning(f"The machine {name} will be restore to snapshot {snapshot}")
    if yes_no_question("Are you sure ? (Y/n)[default: n]", False):
        restore_snapshot(name, snapshot)
        logger.info("OK")
    else:
        logger.info("Aborted")


@app.command()
def create(name: str = typer.Argument(..., help="VM name to snapshot"),
           snapshot: str = typer.Argument(..., help="Name of the snapshot")):
    """
    Create a snapshot of a VM
    """
    create_snapshot(name, snapshot)
    logger.info(f"Created snapshot {snapshot}")


@app.command()
def delete(name: str = typer.Argument(..., help="VM name concerned by the snapshot deletion"),
           snapshot: str = typer.Argument(..., help="Name of the snapshot to delete")):
    """
    Delete a snapshot of a VM
    """
    if yes_no_question(f"Are you sure to delete snapshot {snapshot} ? (Y/n)[default: n]", False):
        delete_snapshot(name, snapshot)
        logger.info(f"Deleted snapshot {snapshot}")
    else:
        logger.info("Aborted")


@app.command(name="list")
def list_(name: str = typer.Argument(..., help="The name of the VM concerned by the snapshots")):
    """
    List snapshots of a VM
    """
    list_snapshot(name)
