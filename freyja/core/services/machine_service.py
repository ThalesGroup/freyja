import curses
import logging
import time
from pathlib import Path
from typing import Dict, List, Optional

from freyja.configuration import Configuration
from freyja.core.handlers.host_handler import HostHandler
from freyja.environment import FreyjaEnvironment
from freyja.lib.exceptions.machine_exceptions import MachineAlreadyExists
from freyja.lib.utils.bytes_utils import convert_size
from freyja.lib.utils.error_utils import check_message
from freyja.lib.utils.subprocess_utils import execute, execute_interactive
from freyja.lib.utils.virsh_utils import parse_info, parse_list, parse_stats
from freyja.logger import FreyjaLogger
from freyja.models import machine_info
from freyja.models.machine_info import Info

logger: logging.Logger = logging.getLogger(FreyjaLogger.name)


def launch_machines(hosts: List[HostHandler]):
    for host in hosts:
        logger.info(f"Create host {host.configuration.hostname}")
        try:
            host.launch()
        except ChildProcessError as e:
            logger.critical(f"{e.args}")


def create_machines(configuration: Path, dry: bool, foreground: bool):
    """
    For each host in configuration, build the needed files to create the related machines and
    networks.
    The configuration step and the launching step are separated to improve the error handling.
    :param configuration: hosts configuration file
    :param dry: If True, only create machines & networks configurations without launching them
    :param foreground: If True, create machines with interactive console mode
    :return: the list of the machines scripts required to create them
    """
    hosts: List[HostHandler] = []
    existing_hosts = list_machines(names=True, stdout=False)
    # configure
    configuration = Configuration.parse_file(configuration)
    for machine_conf in configuration.hosts:
        host = HostHandler(machine_conf)
        hostname = host.configuration.hostname
        if hostname in existing_hosts:
            logger.warning(f"Skip: {MachineAlreadyExists(hostname).message}")
        else:
            logger.info(f"Configure host {hostname}")
            host.configure(foreground)
            logger.debug(f"Host configured in {host.build_dir}")
            hosts.append(host)
    # launch
    if not dry:
        launch_machines(hosts)


def execute_for_domain(cmd: str, domain: str):
    try:
        execute(["virsh", cmd, domain])
    except ChildProcessError as e:
        if check_message(e, "failed to get domain"):
            logger.warning(f"Skip {domain}: Machine not found")
        elif check_message(e, "domain is not running"):
            logger.warning(f"{domain} is already stopped")
        else:
            raise e


def start_machines(names: List[str]):
    logger.info(f"Start machines: {names}")
    for domain in names:
        execute_for_domain("start", domain)


def stop_machines(names: List[str]):
    logger.info(f"Shutdown machines: {names}")
    for domain in names:
        execute_for_domain("shutdown", domain)


def restart_machines(names: List[str]):
    logger.info(f"Reboot machines: {names}")
    for domain in names:
        execute_for_domain("reboot", domain)


def delete_domain(name: str):
    logger.info(f"Delete {name}")
    # domain
    try:
        execute(["virsh", "destroy", name])
    except ChildProcessError as e:
        if check_message(e, "domain is not running"):
            pass
        elif check_message(e, "no storage pool"):
            pass
        else:
            raise e
    execute(["virsh", "undefine", name, "--remove-all-storage"])
    # pool
    execute(["virsh", "pool-destroy", name])
    execute(["virsh", "pool-undefine", name])
    # files
    execute(["rm", "-rf", f"{FreyjaEnvironment.BUILD_DIR}/{name}"])


def delete_machines(names: List[str]):
    for name in names:
        try:
            delete_domain(name)
        except ChildProcessError as e:
            if check_message(e, "failed to get domain"):
                logger.warning(f"Skip {name}: Machine not found")
            else:
                raise e


def list_machines(names: bool = False, stdout: bool = True) -> "List[str]":
    """
    List the machines in libvirt
    :param names: If true, only print the vm names
    :param stdout: If True, stream the subprocess stdout to this app stdout
    :return: The subprocess stdout lines
    """
    cmd = ["virsh", "list", "--all"]
    if names:
        cmd.append("--name")
    return execute(cmd, stream_stdout=stdout)


def parse_machines_network_info(domain: str, domiflist: List[Dict[str, str]]) \
        -> "List[machine_info.Network]":
    """
    Parse domain's domiflist and domifaddr and network information into a list of Networks info
    models
    :param domain: domain name
    :param domiflist: raw domiflist dictionary output from libvirt
    :return: the resulting Info model
    """
    result = []
    for if_info in domiflist:
        name = if_info.get(machine_info.mapping.get("net_name"))
        mac = if_info.get(machine_info.mapping.get("net_mac"))
        # get interface
        net_info = parse_info(execute(["virsh", "net-info", name]))
        net_info_lowercase = {k.lower(): v for k, v in net_info.items()}
        interface_type = if_info.get(machine_info.mapping.get("net_type"))
        interface = net_info_lowercase.get(interface_type)
        # get IP using DHCP info
        ip = None
        dhcp = parse_list(execute(["virsh", "net-dhcp-leases", name]))
        for entry in dhcp:
            if mac == entry.get('MAC address'):
                ip = entry.get('IP address').replace('/0', '')

        # build model
        default_info = "unknown"
        result.append(machine_info.Network(name=name,
                                           mac=mac,
                                           type=interface_type,
                                           interface=interface if interface else default_info,
                                           ip=ip if ip else default_info))

    return result


def parse_machines_info(domain: str, dominfo: Dict[str, str], domiflist: List[Dict[str, str]]) \
        -> "Info":
    """
    Parse domain's dominfo and network information to the Info model
    :param domain: domain name
    :param dominfo: raw dominfo dictionary output from libvirt
    :param domiflist: raw domiflist dictionary output from libvirt
    :return: the resulting Info model
    """
    # parse networks information domiflist and domifaddr for this domain
    networks = parse_machines_network_info(domain, domiflist)
    # parse memory information for this domain
    # convert max memory display '4194304 KiB' into a human-readable format in GB
    kib_mem = str(dominfo.get(machine_info.mapping.get("memory")))
    mem = convert_size(int(kib_mem.split(" ")[0]) * 1024)
    # parse into the model
    return machine_info.Info(state=dominfo.get(machine_info.mapping.get("state")),
                             networks=networks,
                             vcpus=int(dominfo.get(machine_info.mapping.get("vcpus"))),
                             memory=mem)


def get_domain_info(domain: str) -> "Info":
    """
    Collect various information about a libvirt domain and parse it into a model
    :param domain: the domain concerned by the information query
    :return: the information model
    """
    # get vm general info
    dominfo_raw = execute(["virsh", "dominfo", domain])
    dominfo: Dict = parse_info(dominfo_raw, machine_info.mapping.values())

    # get domain interfaces list
    domiflist_raw = execute(["virsh", "domiflist", domain])
    domiflist: List[Dict[str, str]] = parse_list(domiflist_raw)

    return parse_machines_info(domain, dominfo, domiflist)


def info_machines(names: List[str]) -> "Dict":
    """
    List the machines in libvirt and query all the needed information for each machine
    :param names: If true, only print the vm names
    :return: The subprocess stdout lines
    """
    result = {}
    domains: List[str] = names if names else \
        list(filter(None, list_machines(names=True, stdout=False)))
    for domain in domains:
        try:
            result[domain] = get_domain_info(domain).dict()
        except ChildProcessError as e:
            if check_message(e, "failed to get domain"):
                logger.warning(f"Skip {domain}: Machine not found")
            elif check_message(e, "domain is not running"):
                logger.warning(f"Skip {domain}: not running. Start it to get info.")
            else:
                raise e

    return result


def get_usages(usages: Dict[str, Dict], sleep_seconds: int):
    """
    Get cpus times and return the live cpu usage in percents during the measurement's seconds
    :param usages: the domains' information with max memory and vcpus amounts for each domain
    :param sleep_seconds: The measurement's time in seconds. The cpu usage is taken at T0 and
                          T0+sleep_seconds.
    :return: the cpu usage in percents
    """
    # get cpu time and ram usage at T0
    cpu_times = {}
    for domain, static in usages.items():
        # memory
        try:
            mem_report = execute(["virsh", "dommemstat", domain, "--live"])
            parsed_report = parse_stats(mem_report, header=False)
            used_mem_kib = int(parsed_report.get("rss"))
            static['used_mem_bytes'] = convert_size(used_mem_kib * 1000)
            static['used_mem'] = round(float(used_mem_kib * 100 / int(parsed_report.get("actual"))))
            # cpu
            cpu_report_0 = execute(["virsh", "cpu-stats", domain, "--total"])
            cpu_times[domain] = float(parse_stats(cpu_report_0).get("cpu_time"))
            static['running'] = True
        except ChildProcessError as e:
            if check_message(e, "domain is not running"):
                static['running'] = False

    # get cpu times at T0+sleep_seconds
    time.sleep(sleep_seconds)
    for domain, static in usages.items():
        try:
            cpu_report_1 = execute(["virsh", "cpu-stats", domain, "--total"])
            cpu_time_1 = float(parse_stats(cpu_report_1).get("cpu_time"))
            static['used_cpu'] = round((cpu_time_1 - cpu_times.get(domain))
                                        / (sleep_seconds * static.get("vcpus")) * 100)
            static['running'] = True
        except ChildProcessError as e:
            if check_message(e, "domain is not running"):
                static['running'] = False

    return usages


def display_report(static: bool, usages: Dict[str, Dict], stdscr: Optional = None):
    """
    Display a report about cpu usage, mem usage in percent and in Bytes.
    The report may be static (usages are computed only once) and displayed once in the user's
    terminal
    Or the report may be dynamic (usages are computed in real time) and the user's terminal is
    cleared at each refresh.
    :param static: If true, enable the static report.
    :param usages: dictionary per domain, containing all the necessary information to display
    :param stdscr: the stdout screen instance from 'curses' module. Used only if the report is not
                   static
    """
    i = 0
    for domain, usage in usages.items():
        running: bool = bool(usage.get("running"))
        state = "\trunning: " + str(usage.get("running"))
        if running:
            cpu = "\tcpu: " + str(usage.get("used_cpu")) + "%"
            mem = "\tmemory: " + str(usage.get("used_mem")) + "%"
            mem_bytes = "\tused memory: " + str(usage.get("used_mem_bytes"))
        if static:
            print(domain + ":")
            print(state)
            if running:
                print(cpu)
                print(mem)
                print(mem_bytes)
        else:
            stdscr.addstr(i, 0, domain + ":")
            stdscr.addstr(i + 1, 0, cpu + "\t")
            stdscr.addstr(i + 2, 0, mem + "\t")
            stdscr.addstr(i + 3, 0, mem_bytes + "\t")
            stdscr.refresh()
            i += 4


def usage_machine(names: List[str], watch: bool = False):
    domains: List[str] = names if names else \
        list(filter(None, list_machines(names=True, stdout=False)))
    # display
    if watch:
        stdscr = curses.initscr()
        curses.noecho()
        curses.cbreak()
    # collect static info
    static_usages: Dict[str, Dict] = {}
    for domain in domains:
        try:
            domain_info = get_domain_info(domain)
            static_usages[domain] = {}
            static_usages[domain]['vcpus'] = int(domain_info.vcpus)
            static_usages[domain]['max_mem'] = domain_info.memory
        except ChildProcessError as e:
            if check_message(e, "failed to get domain"):
                logger.warning(f"Skip {domain}: Machine not found")
            else:
                raise e

    # collect and display live info
    while True:
        usages = get_usages(static_usages, 1)
        if watch:
            display_report(False, usages, stdscr)
        else:
            display_report(True, usages)
            break

    if watch:
        curses.nocbreak()
        curses.echo()
        curses.endwin()

def open_console_machine(domain: str):
    """
    Opens a console for the provided machine
    :param name: name of the machine in which the console will be opened
    """
    try:
        execute_interactive(["virsh", "console", domain])
    except ChildProcessError as e:
        logger.warning(f"Skip {domain}: Machine not found")