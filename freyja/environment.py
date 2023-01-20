from pathlib import Path


class FreyjaEnvironment:
    """
    Use this class to :
      - initialize environment in main
      - call this environment to get environment values
    """
    APP_NAME = "freyja"
    VERSION = "0.1.0-beta"
    TEMPLATES_DIR = Path(__file__).parent / "templates"
    BUILD_DIR = Path.home() / f"{APP_NAME}-workspace/build"
    CREATE_VM_FILENAME = "create.sh"
    CLOUD_INIT_FILENAME = "vm.clinit"
    IGNITION_FILENAME = "provisioning.ign"
    NETWORK_FILENAME = "network.xml"
    CREATE_VM_TEMPLATE_NAME = CREATE_VM_FILENAME + ".j2"
    CLOUD_INIT_TEMPLATE_NAME = CLOUD_INIT_FILENAME + ".j2"
    IGNITION_TEMPLATE_NAME = IGNITION_FILENAME + ".j2"
    NETWORK_TEMPLATE_NAME = NETWORK_FILENAME + ".j2"

    @classmethod
    def get_version(cls):
        return f"{cls.APP_NAME} v{cls.VERSION}"

    @classmethod
    def init(cls):
        if not cls.BUILD_DIR.exists():
            cls.BUILD_DIR.mkdir(parents=True)
