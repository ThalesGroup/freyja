

def check_message(e: Exception, sub_message: str):
    """
    Look for a substring in an error messages
    :param e: Exception concerned by the sub message to search
    :param sub_message: sub string to search in error messages
    :return: True if the sub_message is found in error messages, or False
    """
    for msg in e.args:
        if sub_message in msg:
            return True

    return False
