class MachineException(Exception):
    pass


class MachineAlreadyExists(MachineException):
    def __init__(self, hostname: str, *args):
        self.message = f"Machine already exists: {hostname}"
        super().__init__(self.message, args)
