""" API for controlling the aretext editor. """

from pyaretext.api.rpcclient import default_client, EmptyMsg


def quit():
    """Quit the aretext editor. """
    # TODO: prompt for confirmation if there are unsaved changes.
    result = default_client().quit(EmptyMsg())
    print(result)
