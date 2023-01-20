import os
from pathlib import Path

TEST_DIR = os.path.dirname(os.path.abspath(__file__))
RESOURCES_DIR = f"{TEST_DIR}/resources"
APPLICATION_NAME = "freyja"
TESTS_OUTPUT_DIR = "/tmp/freyja/test"


def read_file(path: Path) -> "str":
    with open(path, 'r') as f:
        return str(f.read())
