# from pathlib import Path
#
# from freyja.lib.utils.file_utils import read_file
# from freyja.lib.utils.virsh_utils import parse_info, parse_list, parse_stats
# from freyja.tests.common import RESOURCES_DIR
#
#
# def test_parse_info():
#     input_file = f"{RESOURCES_DIR}/virsh_info.txt"
#     virsh_output = read_file(Path(input_file)).split("\n")
#     keys_whitelist = ["Id", "Name"]
#
#     output_dict = parse_info(virsh_output)
#
#     assert output_dict
#     assert len(output_dict) == 10
#     assert output_dict.get("Id")
#     assert output_dict.get("Id") == "2"
#
#     output_dict_filtered = parse_info(virsh_output, keys_whitelist)
#     assert output_dict_filtered
#     assert len(output_dict_filtered) == 2
#     assert "State" not in output_dict_filtered.keys()
#     assert output_dict_filtered.get("Name") == "vm-test"
#
#
# def test_parse_list():
#     input_file = f"{RESOURCES_DIR}/virsh_list.txt"
#     virsh_output = read_file(Path(input_file)).split("\n")
#
#     output_list = parse_list(virsh_output)
#
#     assert output_list
#     assert len(output_list) == 2
#     assert output_list[0] and output_list[1]
#     assert output_list[0].get("Interface") == "vnet0"
#     assert output_list[0].get("Type") == 'bridge'
#     assert output_list[0].get("Source") == 'ctrl-plane'
#     assert output_list[0].get("Model") == 'virtio'
#     assert output_list[0].get("MAC") == '52:54:02:aa:bb:dd'
#
#
# def test_parse_cpu_stats():
#     input_file = f"{RESOURCES_DIR}/virsh_cpustats.txt"
#     virsh_output = read_file(Path(input_file)).split("\n")
#
#     output_dict = parse_stats(virsh_output)
#
#     assert output_dict
#     assert len(output_dict) == 3
#     assert output_dict.get("cpu_time") == "182.223063220"
#     assert output_dict.get("user_time") == "3.530000000"
#     assert output_dict.get("system_time") == "25.640000000"
