import sys
from pathlib import Path

from freyja.cli.cli import cli
from freyja.environment import FreyjaEnvironment
from freyja.logger import FreyjaLogger


def main():
    # init logger
    FreyjaLogger()
    # env
    FreyjaEnvironment.init()
    # init shellcli
    cli()


if __name__ == '__main__':
    """
    For debug purpose
    """
    sys.argv[0] = str(Path(sys.argv[0]).parent.name)
    main()
