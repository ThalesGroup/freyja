import re
from typing import Dict, List, Optional


def trim_columns(columns: str) -> "List[str]":
    """
    split a string that describes columns into a list of strings
    """
    columns = columns.lstrip()
    return re.split(r'\s{2,}', columns)


def trim_lines(output: List[str]) -> "List[str]":
    """
    remove duplicated whitespaces and empty members in list of lines
    """
    return list(filter(None, [" ".join(line.split()) for line in output]))


def parse_info(output: List[str], whitelist: Optional[List[str]] = None) \
        -> "Dict":
    """
    Get raw virsh output containing lines which contain 'key:   value' information.
    This list is parsed into a dictionary.
    If a whitelist is provided, the output is filtered according to the information's keys in the
    whitelist.
    example of info output:
    ```
    Id:             2
    Name:           vm-test
    OS Type:        hvm
    ```
    :param output: Raw string output from a virsh info command
    :param whitelist: List of keys to filter the output information
    :return: a dictionary with the information
    """
    # remove duplicated whitespaces and empty members
    trimmed: List[str] = trim_lines(output)
    # parse list of strings 'k:v' to a dictionary
    if whitelist:
        return {kv[0]: kv[1] for kv in [entry.split(": ") for entry in trimmed]
                if kv[0] in whitelist}
    else:
        return {kv[0]: kv[1] for kv in [entry.split(": ") for entry in trimmed]}


def parse_list(output: List[str]) -> "List[Dict[str, str]]":
    """
    Get raw virsh output containing columns names and information.
    The column names and values are parsed into a dictionary.
    example of info output:
    ```
     Interface   Type     Source       Model    MAC
    ---------------------------------------------------------------
     vnet0       bridge   ctrl-plane   virtio   52:54:02:aa:bb:dd
     vnet1       bridge   data-plane   virtio   52:54:02:aa:bb:ee
    ```
    :param output: Raw string output from a virsh info command
    :return: a dictionary with the information
    """
    result: List[Dict[str, str]] = []
    # remove duplicated whitespaces and empty members
    trimmed_output = list(filter(None, output))
    columns: List[str] = trim_columns(trimmed_output[0])
    for line in trimmed_output[2:]:
        values = trim_columns(line)
        result.append(dict(zip(columns, values)))

    return result


def parse_stats(output: List[str], header: bool = True) -> "Dict":
    """
    Get raw virsh statistics output containing lines which contain 'key   value' information.
    These stats are parsed into a dictionary.
    example of virsh stats :
    ```
    actual 4194304
    swap_in 0
    swap_out 0
    ```
    :param output: Raw string output from a virsh stats commands
    :param header: If True, skip the first line of stats
    :return: a dictionary with the stats
    """
    trimmed: List[str] = trim_lines(output)[1:] if header else trim_lines(output)
    return {kv[0]: kv[1] for kv in [entry.split(" ") for entry in trimmed]}
