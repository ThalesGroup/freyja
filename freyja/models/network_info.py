from pydantic import BaseModel


class Network(BaseModel):
    name: str
    ip: str
    mac: str
    type: str
    interface: str
