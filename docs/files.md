Files
=====

Aretext is designed to integrate seamlessly with a terminal-based workflow. This strongly influences aretext's approach to managing files:

-	It delegates window management to your terminal multiplexer or emulator. Each instance of aretext opens a single document at a time; to edit multiple documents simultaneously, you can use [tmux](https://wiki.archlinux.org/title/Tmux) to run multiple instances of aretext in the same terminal.

-	It provides no commands within the editor to move, rename, create files, or change the working directory. You can use your shell (outside the editor) for these functions.

-	It automatically reloads files that change on disk (unless there are unsaved changes). For example, if you run a code formatting tool that changes a file, aretext will automatically reload it.

Aretext currently supports only UTF-8 encoded documents with Unix-style (LF) line endings.

Fuzzy file search
-----------------

Aretext has built-in fuzzy search for files. This allows you to quickly find and open a file without leaving the editor:

1.	In normal mode, type ":" to open the command menu.
2.	In the menu search bar, type "f" to select the "find and open" command, then press enter. (If there are unsaved changes in the current document, you will need to either save them first or force-reload to discard the changes.)
3.	Type in the search bar to filter the file paths. Use arrow keys or tab to choose a file path to open.
4.	Press enter to open the selected file.

Aretext always searches within the current working directory.

Opening a file from the command line
------------------------------------

To have aretext open a document immediately, pass the path as a positional argument like this: `aretext path/to/file`.

If you do not provide a path argument, aretext will start an empty document called something like "untitled-1621625423.txt" (the number is a Unix timestamp). You can either insert text and save this document (useful for writing quick notes) or use fuzzy file search to open another document.

Previous and next document
--------------------------

Aretext remembers which documents you have opened in the editor. To return to the previous document:

1.	In normal mode, type ":" to open the command menu.
2.	In the menu search bar, type "p" then select "open previous document".

Once you have opened a previous document, you can return to next document using the "open next document" menu command.

Unsaved changes
---------------

Aretext will warn you if a command would discard unsaved changes or overwrite changes made by another program to the file on disk. You must then decide to either force-save, force-reload, or force-quit.

-	To force-save, select the "force save document" menu command. This will overwrite the changes on disk.
-	To force-reload, select the "force reload" menu command. This will discard unsaved changes and reload the document from disk.
-	To force-quit, select the "force quit" menu command. This will discard unsaved changes and exit the program.

Using grep to search files
--------------------------

To search for a term in multiple files, you can create a custom menu command that calls `grep`. See [Custom Menu Commands](custom-menu-commands.md) for instructions.
