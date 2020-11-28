""" API for controlling the aretext editor. """
from typing import Optional
from pyaretext.api.rpcclient import default_client, EmptyMsg, SetSyntaxMsg, OpResultMsg


def set_syntax(language: Optional[str]) -> OpResultMsg:
    """Set the syntax of the current document. """
    if language is None:
        language = "undefined"

    msg = SetSyntaxMsg(language=language)
    return default_client().set_syntax(msg)


def quit() -> OpResultMsg:
    """Quit the aretext editor. """
    # TODO: prompt for confirmation if there are unsaved changes.
    return default_client().quit(EmptyMsg())
