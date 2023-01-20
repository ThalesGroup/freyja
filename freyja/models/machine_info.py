from typing import List

from pydantic import BaseModel

# used to map input with output models
from freyja.models.network_info import Network

mapping = {"id": "Id",
           "vcpus": "CPU(s)",
           "state": "State",
           "memory": "Max memory",
           "disk": "Allocation",
           "net_name": "Source",
           "net_ip": "Address",
           "net_mac": "MAC",
           "net_type": "Type",
           "net_interface": "Interface"}


class Info(BaseModel):
    state: str
    networks: List[Network]
    vcpus: int
    memory: str
