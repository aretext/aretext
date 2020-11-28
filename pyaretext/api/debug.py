""" API for debugging the aretext editor. """

from pyaretext.api.rpcclient import default_client, ProfileMemoryMsg, OpResultMsg


def profile_memory(path: str) -> OpResultMsg:
    """Write a memory profile to the specified path. """
    req = ProfileMemoryMsg(path=path)
    return default_client().profile_memory(req)
