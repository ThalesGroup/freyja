import logging
import os
import stat
from pathlib import Path
from typing import List

logger = logging.getLogger(__name__)


def read_file(file: Path) -> "str":
    with open(os.path.expandvars(file), 'r') as f:
        return f.read().replace('\n', '')


def read_files(files: List[Path]) -> "List[str]":
    return [read_file(file) for file in files]


def write_str(content: str, output: Path):
    with open(os.path.expandvars(output), 'w') as f:
        f.write(content)


def set_exec_permission(file: Path):
    st = os.stat(file)
    os.chmod(file, st.st_mode | stat.S_IEXEC)
