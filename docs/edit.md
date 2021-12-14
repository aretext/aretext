Edit
====

This guide describes commands you can use to edit a document efficiently using vim key bindings. It assumes you are familiar with the commands described in the [Quickstart](quickstart.md).

Insert and append
-----------------

In normal mode, type "i" to enter insert mode at the current cursor position. To insert at the position *after* the cursor, type "a" (short for "append") instead.

To insert at the *start* of the line, type "I". To append after the *end* of the line, type "A".

From insert mode, you can return to normal mode by pressing the escape key.

Delete
------

To delete a single character in normal mode, type "x". (In insert mode, you can use the backspace key instead.)

To delete a line, type "dd" in normal mode. To delete from the cursor to the end of the line, type "D".

There are many delete commands of the form "d<motion>", where <motion> is one of the cursor movement commands described in [Navigation](navigation.md):

-	"d$" deletes from the cursor to the end of the line, because "$" means "move the cursor to the end of the line".
-	"diw" (which means "delete inner word") deletes the current word under the cursor.
-	"daw" (which means "delete a word") deletes the current word under the cursor and any trailing whitespace.
-	"dh" deletes one character to the left, and "dl" deletes one character to the right.
-	"dj" deletes the current and previous lines, and "dk" deletes the current and next line.
-	"dt\{char\}" deletes up to, but not including, the next matching character on the current line.

Replace
-------

To replace the character under the cursor, type "r" in normal mode, then type the new character.

Change
------

In aretext, "change" means to delete some text then enter insert mode. This is useful for quickly changing a word or line.

To change the current word under the cursor in normal mode, type "ciw" ("change inner word").

To change the current word under the cursor and trailing whitespace in normal mode, type "caw" ("change a word").

Put and yank
------------

When you delete text, aretext copies it to a hidden buffer. You can then insert the deleted text after the cursor by typing "p" (short for "put") in normal mode. Alternatively, you can type "P" to insert the text before the cursor.

You can copy a line into the buffer by typing "yy" (short for "yank") in normal mode.

If you want to copy/paste using your system's clipboard, you will need to add custom menu commands (see [Customization](customization.md) for instructions).

Inserting and joining lines
---------------------------

To start a new line below the cursor and enter insert mode, type "o".

To start a new line above the cursor and enter insert mode, type "O".

To join the current line with the line below, type "J".

Indenting and outdenting
------------------------

To indent the current line, type ">>".

To outdent the current line, type "\<<".

Toggle case
-----------

To change the character under the cursor from uppercase to lowercase, or vice versa, type "~".

Selection (visual mode)
-----------------------

To start a selection, type "v" (short for "visual mode"). You can use the same cursor motions as in normal mode to move the end of the selection.

To select entire lines (instead of individual characters), type "V".

You can use the following commands to modify the selected text:

-	"x" and "d" both delete the selection.
-	"c" (short for "change") deletes the selection and enters insert mode.
-	"~" toggles the case of the selection.
-	">" indents the selection.
-	"<" outdents the selection.
-	"y" (short for "yank") copies the selection.

To clear the selection and return to normal mode, press the escape key.

Undo and redo
-------------

To undo the last edit, type "u" (short for "undo") in normal mode.

To redo the last edit, press Ctrl-r (short for "redo") in normal mode.

Aretext clears the undo history whenever a document is loaded or reloaded.

Repeat last action
------------------

To repeat the last action, type "." in normal mode. This is useful for avoiding repetitive typing.

Record and replay a macro
-------------------------

To repeat a sequence of commands, you can record a macro.

1.	In normal mode, type ":" to open the command menu.
2.	Search for and select "start/stop recording macro" to begin recording a macro.
3.	Edit the document. Any changes you make will be recorded in the macro.
4.	In the command menu, select "start/stop recording macro" again to stop recording the macro.

To replay the recorded macro, select "replay macro" in the command menu.

Once you have replayed a macro, you can repeat it using the "." (repeat last action) command in normal mode.
