Quickstart
==========

This guide helps you get started using aretext. It assumes you are comfortable using a terminal.

Introduction
------------

Aretext is a terminal-based editor with vim-compatible key bindings. If you have used vim before, aretext should feel very familiar. If not, this guide will help you get started.

Installing the editor
---------------------

The first step is to [install aretext](install.md). You can check whether aretext is installed by running:

```
aretext -version
```

If you see a version string like "1.2.3 @ 9955832c3e2036b762e59238fe39f648a3cb1199" then aretext is installed!

Starting the editor
-------------------

To start the editor, run `aretext`. This will start a new, empty document called something like "untitled-1621605673.txt" (the number is a Unix timestamp).

Many users set an alias so they can launch `aretext` quickly. If you are using bash, you can add this line to your `~/.bashrc` or `~/.bash_profile`:

```
# alias "at" to start aretext
alias at="aretext"
```

Inserting text
--------------

To insert text, first press "i" to enter insert mode. You can tell you are in insert mode because the bottom of the screen will display:

```
-- INSERT --
```

While in insert mode, every character you type will be inserted in the document.

When you are done typing, press the escape key to return to normal mode.

Navigating a document
---------------------

In normal mode you can move the cursor using the "h", "j", "k", and "l" keys:

-	"h" moves the cursor left.
-	"j" moves the cursor down.
-	"k" moves the cursor up.
-	"l" moves the cursor right.

If you prefer, you can also use the arrow keys.

Saving and quitting
-------------------

Return to normal mode by pressing the escape key. Then press ":" to open the command menu.

Type "s" to search for the "save" command. The first item should be "save". Press enter to save the document and return to normal mode.

In normal mode, type ":" to open the command menu again. This time, type "q" to search for the "quit" command. Press enter to quit the editor.

NOTE: if you are familiar with vim's "ex" mode commands, you can use these too! For example, "w" (write) always matches the save command, and "wq!" matches "force save document and quit".

Next steps
----------

Congratulations! You can now edit a document in aretext!

The next sections explain how to use aretext effectively:

-	[Files](files.md): How to find, open, and save documents.
-	[Navigation](navigation.md): How to navigate within a document.
-	[Edit](edit.md): How to edit a document.
-	[Configuration](configuration.md): How to configure settings.
-	[Custom Menu Commands](custom-menu-commands.md): How to extend the editor with custom menu commands.

For a list of common commands, please see the [Cheat Sheet](cheat-sheet.html).
