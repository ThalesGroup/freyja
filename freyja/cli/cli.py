import logging

import typer

from freyja.cli import machine, network, snapshot
from freyja.environment import FreyjaEnvironment
from freyja.logger import FreyjaLogger

app = typer.Typer(help=f"Manage virtual machine and network using QEMU and KVM",
                  no_args_is_help=True)
app.add_typer(machine.app, name="machine")
app.add_typer(network.app, name="network")
app.add_typer(snapshot.app, name="snapshot")

logger = logging.getLogger(FreyjaLogger.name)


@app.callback(invoke_without_command=True)
def main_callback(
        version: bool = typer.Option(False, "--version",
                                     help="Print version",
                                     is_eager=True),
        environment: bool = typer.Option(False, "--env",
                                         help="Print the application's environment",
                                         is_eager=True)):
    if version:
        typer.echo(FreyjaEnvironment.get_version())
        # raise exit to avoid further mandatory options
        raise typer.Exit()


def cli():
    app()
