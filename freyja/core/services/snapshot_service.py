import logging

from freyja.lib.utils.error_utils import check_message
from freyja.lib.utils.subprocess_utils import execute
from freyja.logger import FreyjaLogger

logger: logging.Logger = logging.getLogger(FreyjaLogger.name)


def create_snapshot(domain: str, name: str):
    try:
        execute(["virsh", "snapshot-create-as", domain, "--name", name])
    except ChildProcessError as e:
        logger.warning(f"Skip {domain}: Machine not found")


def delete_snapshot(domain: str, name: str):
    try:
        execute(["virsh", "snapshot-delete", domain, "--snapshotname", name])
    except ChildProcessError as e:
        logger.warning(f"Skip {domain}: Machine not found")


def restore_snapshot(domain: str, name: str):
    try:
        execute(["virsh", "snapshot-revert", domain, "--snapshotname", name])
    except ChildProcessError as e:
        if check_message(e, "snapshot"):
            logger.warning(f"Skip {domain}: snapshot {name} not found")
        else:
            logger.warning(f"Skip {domain}: Machine not found")


def list_snapshot(domain: str):
    try:
        execute(["virsh", "snapshot-list", domain], stream_stdout=True)

    except ChildProcessError as e:
        logger.warning(f"Skip {domain}: Machine not found")
