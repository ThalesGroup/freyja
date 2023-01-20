from typer.testing import CliRunner

from freyja.cli.cli import app

runner = CliRunner()


def test_network_list():
    result = runner.invoke(app, ["network", "list"])
    assert result.exit_code == 0


def test_network_info():
    # handles not found
    result = runner.invoke(app, ["network", "info", "devnull"])
    assert result.exit_code == 0

    # ok
    result = runner.invoke(app, ["network", "info"])
    assert result.exit_code == 0
