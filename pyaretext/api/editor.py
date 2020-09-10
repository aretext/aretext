""" API for controlling the aretext editor. """

from pyaretext.api.rpcclient import DEFAULT_CLIENT, EmptyMsg


def quit():
    """Quit the aretext editor. """
    # TODO: prompt for confirmation if there are unsaved changes.
    result = DEFAULT_CLIENT.quit(EmptyMsg())
    print(result)
