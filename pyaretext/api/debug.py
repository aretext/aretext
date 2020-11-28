""" API for debugging the aretext editor. """

from pyaretext.api.rpcclient import (
    default_client,
    ProfileMemoryRequestMsg,
    ProfileMemoryResponseMsg,
)


def profile_memory(path: str):
    """Write a memory profile to the specified path. """
    req = ProfileMemoryRequestMsg(path=path)
    resp = default_client().profile_memory(req)
    if not resp.succeeded:
        print("Error writing memory profile: {}".format(resp.error))
