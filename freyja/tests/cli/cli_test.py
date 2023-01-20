from typer.testing import CliRunner

from freyja.cli.cli import app

runner = CliRunner()


def test_cli():
    result = runner.invoke(app, ["--version"])
    assert result.exit_code == 0
    assert result.stdout
