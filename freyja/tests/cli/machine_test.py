from typer.testing import CliRunner

from freyja.cli.cli import app
from freyja.tests.common import RESOURCES_DIR

runner = CliRunner()


def test_machine_create():
    # testing mandatory configuration
    result = runner.invoke(app, ["machine", "create"])
    assert result.exit_code == 2

    # testing missing conf
    result = runner.invoke(app, ["machine", "create", "-c", "/dev/null"])
    assert result.exit_code == 1

    # testing erroneous conf
    result = runner.invoke(app, ["machine", "create", "-c", f"{RESOURCES_DIR}/invalid_conf.yaml"])
    assert result.exit_code == 1

    # testing ok
    result = runner.invoke(app, ["machine", "create", "-c", f"{RESOURCES_DIR}/simple_conf.yaml",
                                 "--dry-run", "-v"])
    assert result.exit_code == 0


def test_machine_list():
    result = runner.invoke(app, ["machine", "list"])
    assert result.exit_code == 0


def test_machine_info():
    # handles not found
    result = runner.invoke(app, ["machine", "info", "devnull"])
    assert result.exit_code == 0

    # ok
    result = runner.invoke(app, ["machine", "info"])
    assert result.exit_code == 0


def test_machine_delete():
    # handles not found
    result = runner.invoke(app, ["machine", "delete", "devnull"], input="Y")
    assert result.exit_code == 0

    # handles input
    result = runner.invoke(app, ["machine", "delete", "devnull"], input="n")
    assert result.exit_code == 0


def test_machine_start():
    # handles not found
    result = runner.invoke(app, ["machine", "start", "devnull"])
    assert result.exit_code == 0


def test_machine_stop():
    # handles not found
    result = runner.invoke(app, ["machine", "stop", "devnull"])
    assert result.exit_code == 0


def test_machine_usage():
    # all by default
    result = runner.invoke(app, ["machine", "usage"])
    assert result.exit_code == 0

    # handles not found
    result = runner.invoke(app, ["machine", "usage", "devnull"])
    assert result.exit_code == 0
