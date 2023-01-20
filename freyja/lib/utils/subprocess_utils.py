import subprocess
import sys
from typing import List

CMD_TIMEOUT_SECONDS = 10


def execute(command: List[str], user_input: str = None, stream_stdout: bool = False) -> "List[str]":
    """
    Run a subprocess from the provided command and log into a dedicated logfile
    Eg : execute(["echo", "hello world !"])

    :param command: List of params to launch the subprocess (just like a command line)
    :param user_input: Stdin value for user input to provide to the launched subprocess
    :param stream_stdout: If true, stream the stdout of the process in console
    :return: A list with process' [return_code, stdout]
    """
    process = subprocess.Popen(command, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    # user input
    if user_input is not None:
        process.communicate(str.encode(user_input), timeout=CMD_TIMEOUT_SECONDS)
    # process execution
    output = []
    while True:
        output_bytes = process.stdout.readline()
        if output_bytes == b'' and process.poll() is not None:
            break
        elif output_bytes:
            output_str = output_bytes.decode("utf-8").replace('\n', '')
            output.append(output_str)
            if stream_stdout:
                print(output_str)

    process.terminate()

    if process.returncode != 0:
        raise ChildProcessError(str(output))
    else:
        return output


def yes_no_question(question: str, default: bool):
    valid = {"yes": True, "Y": True, "no": False, "n": False}
    ask = True
    while ask:
        sys.stdout.write(question + " ")
        choice = input()
        if choice == "":
            return default
        elif choice in valid:
            return valid[choice]
        else:
            pass
