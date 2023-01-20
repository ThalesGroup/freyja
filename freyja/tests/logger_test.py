import logging
from enum import Enum
from pathlib import Path

from freyja.logger import FreyjaLogger
from freyja.tests.common import TESTS_OUTPUT_DIR, read_file


class Level(str, Enum):
    INFO = "[INFO]"
    DEBUG = "[DEBUG]"
    ERROR = "[ERROR]"


def init_logger(logfile: Path, debug: bool) -> "logging.Logger":
    """
    init a logger dedicated to one test
    :param logfile: path of the logfile, will be created in TESTS_OUTPUT_DIR
    :param debug: enable debug level
    """
    logger_init = FreyjaLogger(logfile=logfile, debug=debug)
    assert logger_init
    assert logger_init.formatter
    assert logger_init.name == "freyja"
    return logging.getLogger(logger_init.name)


def write_message(logger: logging.Logger, message: str, level: Level):
    if level == Level.INFO:
        logger.info(message)
    elif level == Level.DEBUG:
        logger.debug(message)
    else:
        logger.error(message)


def check_message(message: str, logfile: Path, level: Level):
    info_content = read_file(logfile)
    assert str(level.value) in info_content
    assert message in info_content


def run(logfile_name: str, message: str, level: Level):
    """
    Common run testing for a logfile and a log level
    :param logfile_name: name of the logfile
    :param message: message included in the log
    :param level: level of the log
    """
    logfile = Path(TESTS_OUTPUT_DIR) / logfile_name
    debug = True if level == level.DEBUG else False
    logger = init_logger(logfile, debug=debug)
    assert logfile.exists()

    write_message(logger, message, level)
    check_message(message, logfile, level)


def test_info():
    run("test_info.log", "info message", Level.INFO)


def test_debug():
    run("test_debug.log", "debug message", Level.DEBUG)


def test_error():
    run("test_error.log", "error message", Level.ERROR)



