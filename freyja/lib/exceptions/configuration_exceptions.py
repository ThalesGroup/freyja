from pathlib import Path


class ConfigurationException(Exception):
    pass


class ConfigurationFileNotFoundException(ConfigurationException):
    def __init__(self, message: str = "Configuration file not found", *args):
        self.message = message
        super().__init__(self.message, args)


class ConfigurationFormatError(ConfigurationException):
    def __init__(self, message: str = "Unexpected configuration format", *args):
        self.message = message
        super().__init__(self.message, args)


class ConfigurationContentError(ConfigurationException):
    def __init__(self, message: str = "Erroneous configuration content", *args):
        self.message = message
        super().__init__(self.message, args)


class ConfigurationVersionError(ConfigurationContentError):
    def __init__(self, message: str = "Wrong configuration version", *args):
        self.message = message
        super().__init__(self.message, args)


class ConfigurationImageFileError(ConfigurationContentError):
    def __init__(self, image: Path, message: str = "Wrong configuration for image file", *args):
        self.message = message
        self.image = image
        super().__init__(self.message, args)
