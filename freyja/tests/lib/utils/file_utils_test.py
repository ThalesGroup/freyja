import os
from pathlib import Path

from freyja.lib.utils.file_utils import read_files, set_exec_permission, write_str
from freyja.tests.common import RESOURCES_DIR


def test_read_files():
    resources = ["detailed_conf.yaml", "net_control_plane.xml", "simple_conf.yaml"]
    files_to_read = [Path(f"{RESOURCES_DIR}/{resource}") for resource in resources]
    contents = read_files(files_to_read)

    assert contents
    assert len(contents) == 3

    for content in contents:
        assert content


def test_write_str(tmp_path):
    content = "test_write_str"
    file = tmp_path / "test_write_str.txt"
    tmp_path.mkdir(exist_ok=True)

    write_str(content, file)
    assert file.read_text() == content
    assert len(list(tmp_path.iterdir())) == 1


def test_set_exec_permission(tmp_path):
    tmp_path.mkdir(exist_ok=True)
    file = tmp_path / "test_set_exec_permission.txt"
    file.touch(mode=660)
    assert not os.access(file, os.X_OK)

    set_exec_permission(file)
    assert os.access(file, os.X_OK)
