from typing import Optional

from pydantic import BaseModel


class NetworkConfiguration(BaseModel):
    name: str
    address: str
    bridge: Optional[str]
    interface: Optional[str]
