from pathlib import Path
from typing import Optional


class Provisioning:
    # template name
    template: str
    # rendering output
    output: Path
    user_input: Optional[Path]

    def __init__(self, template: str, output: Path, user_input: Optional[Path] = None):
        self.template = template
        self.output = output
        self.user_input = user_input


