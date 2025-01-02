Client-Server Redesign
======================

Aretext is currently a single binary. To open multiple documents concurrently, users must launch multiple copies of the editor. This architecture, while simple to understand and implement, has some serious limitations:

1.	There is no synchronization between unsaved changes of the same document opened in multiple editor instances, leading to conflicts on save.
2.	It duplicates document and syntax parser state in the common case of viewing/editing the same document in multiple editor instances.
3.	It greatly complicates language server protocol (LSP) integration, since LSP assumes a single editor per LSP server.

(These problems aren't unique to aretext: vim and its descendants have the same limitations.)

The solution is a *client-server* architecture. Aretext will run as a daemon process, serving clients connected over a unix domain socket. All editor state will reside in the server, which can then act as an LSP client.

An important design goal is to adopt this architecture with minimal changes to the user experience. From a user perspective, aretext should continue to work as it does today, just with better synchronization of unsaved changes and LSP support.

Out-of-scope: clients connecting from other machines, multiple client implementation, client/server version skew.

Architecture
------------

The **server** is responsible for:

-	listening on a unix domain socket (UDS) and accepting client connections.
-	receiving and processing input from clients.
-	rendering to the client terminal, using the tty delegation mechanism described below.
-	loading documents from disk, including watching and reloading on changes.
-	executing shell commands.
-	managing LSP lifecycle and communication.

The **clients** are responsible for:

-	proxying terminal input/output to the server, using the tty delegation mechanism described below.
-	(optionally) starting and daemonizing server process.

To avoid conflict between different user accounts, the UDS path will be `$XDG_RUNTIME_DIR/aretext.socket`, with fallback to `/var/run/aretext.socket` for execution within an OCI container.

Client TTY Delegation
---------------------

The client delegates all interactions with its TTY to the server using the following procedure:

1.	Set client tty to raw mode.
2.	Create a socketpair, `s1` and `s2`.
3.	Send `s2` to the server over UDS using SCM_RIGHTS out-of-band data.
4.	Send a "start session" message to the server encoding `$TERM`, the current working directory, and filepath to open.
5.	Copy stdin -> `s1` and `s1` -> stdout until `s2` is closed by the server.

The server then controls the client tty by reading/writing `s2`. This is achieved through a custom `tcell.Tty` implementation that reads and writes `s2`.

For executing shell with `CmdModeTerminal`, the server maintains a psuedoterminal pair (`ptmx` and `pts`). Before executing the command, the server suspends the `tcell.Screen`, then begins copying `ptmx` <-> `s2`. The `pts` side of the pty is used by the subcommand as stdin/stdout/stderr. When the command completes, the server interrupts the copy (by setting read deadline to `time.Now()`) and resumes the `tcell.Screen`.

When tty dimensions change, the client will receive a SIGWINCH signal. To propagate the change, the client sends a `TerminalResizeMsg` to the server. The server updates the `tcell.Tty` dimensions and the `ptmx` (which triggers SIGWINCH to any subcommands using `pts`).

Client-Server Messages
----------------------

Messages will be serialized as JSON, with a uint32 header indicating msg length.

-	`StartSessionMsg`:
	-	sent from client to server after connect
	-	fields:
		-	current working directory
		-	filepath to open
		-	`$TERM` and other env vars used by tcell and other TUI programs
		-	initial width and height of the terminal.
	-	out-of-band data:
		-	SCM_RIGHTS with a file descriptor for proxying the client tty.
-	`TerminalResizeMsg`
	-	sent from client to server on receipt of SIGWINCH signal.
	-	fields:
		-	width
		-	height

The server will close the connection if it receives a message larger than some limit.

Client-Server Lifecycle
-----------------------

The server process may be run by the init system. However, by default the server will be started by a client on-demand and terminate after all clients have disconnected (optionally after some "linger" delay).

This follows a procedure similar to tmux:

1.	Attempt to connect to the server over the unix domain socket.
2.	If connect fails, try to acquire a lock.
3.	If lock acquisition fails, retry from (1).
4.	Otherwise, retry connecting to the server (since another client may have started it before this client acquired the lock).
5.	If connection fails again, start the server process as a daemon, then retry from (4).

To ensure at most one server instance at a time, the server attempts to acquires a file lock on startup. If it cannot acquire the lock, it exits with an error.

The client and server detect that the other side has terminated when the UDS socket is closed.

Editor state
------------

Editor state (owned by the server) must now represent both **shared** data (available to all clients) and **per-session** data. Per-session data is indexed by a unique session ID, which the server assigns on client connection and removes when the client disconnects. Additionally, editor state must include multiple documents loaded at the same time, with each document opened in one or more sessions.

Shared data:

-	Document contents (textTree)
-	File watchers (path/hash/timestamp of each open document)
-	Syntax parse trees
-	Undo/redo log
-	Clipboard
-	Configuration

Per-session data:

-	View (window size, text origin)
-	Currently opened document
-	Cursor position
-	Selection
-	Input mode
-	Working directory
-	Menu/search/textfield state
-	Status msg

Input Coordination
------------------

When one client modifies a document, the server must update the per-session state of all other sessions editing the same document:

1.	All other clients must transition to normal mode (same as if the client pressed "escape"). This implicitly resets most per-session state, including committing any staged undo log operations, clearing selections, etc.
2.	The server then executes the client's action, possibly modifying the document. For each insertion/deletion that occurs before another client's cursor position, increment/decrement the cursor position accordingly.

Shell commands
--------------

All shell commands are executed by the server on behalf of a client. The user is responsible for configuring env vars needed by shell commands the server executes. The shell command's working directory will be set to the current working directory of the client. Likewise, env vars affecting TUI programs like `$TERM` will be set to match the client.

Shell commands always execute asynchronously to avoid blocking the server.

By default, the client and server processes run as the same user/group. The unix domain socket and config files are writable only by this user. This avoids any risk of privilege escalation / confused deputy. To protect against misconfiguration, aretext will require and verify that configuration and unix domain socket files have only user write permission (not group or other).

Configuration
-------------

The current configuration format derives the effective configuration by matching rules to the filepath of the current document. The server will manage multiple documents, so the current format provides no way to specify global configuration.

Since the server loads the configuration once, it is no longer necessary to store all configuration in a single YAML file. New configuration options will be added for client and server settings.

-	`config.yaml` for top-level client/server config.
	-	config version (to allow future migrations)
	-	server timeout after all clients disconnect
	-	whether the client should try to start the server
	-	debug logging
	-	(later) LSP configuration
-	`rules/*.yaml` for document configuration rules
	-	total order by filename

The user can reload configuration using a menu command. On reload, the server will validate the config and display error status on failure. The new configuration applies immediately to all documents.

As this is a breaking change, aretext will provide a built-in migration tool `aretext --migrate-config` that:

1.	Detects old config.yaml format.
2.	Prompts for confirmation.
3.	Moves `config.yaml` to `rules/01-config.yaml`
4.	Creates default global config.yaml.

If the server detects an old configuration, it will exit with an error recommending that the user run `aretext --migrate-config`. The same process will be used moving forward to automate changes to the configuration format.

File Watching
-------------

The server will watch all open files to detect changes. The initial implementation will poll the filesystem (as today), from a single goroutine checking all open files. This could be optimized later by registering for notifications from the operating system.

LSP Integration
---------------

Initial support for the following commands, accessible via menu items:

-	Go to definition
-	Find references
-	Find implementations

The server synchronizes document state with the LSP servers, including edits (insert/delete), document saves, and reloads.

Workspace directories are inferred from the file path by traversing up directories until a file matching a "root pattern" is found (for example, the git repository, Go module, ...). If no match is found, fallback to the client's working directory. The server is responsible for adding/removing workspace directories as client's open files and change their working directories.

Versioning
----------

This is a significant and breaking change, necessitating a major version increment to `2.0`. Users will need to migrate their configuration to the new format.

Development Milestones
----------------------

1.	client/server tty delegation, proving tcell integration and subprocess execution.
2.	client daemonizes server and server exits when all clients disconnect.
3.	multi-client input system -- process input for each client and output the command (without executing anything).
4.	multi-client editor state and actions, with input coordination (without file watch).
5.	file watch and reload.
6.	shellcmd in client tty (run async).
7.	config file refactor, hot reload.
8.	configuration migration tool
9.	LSP integration.

Alternatives Considered
-----------------------

### Synchronize editor state across multiple independent editor instances

Rejected due to complexity.

### Send input/display explicitly through UDS messages

Rejected since any custom protocol would either be tightly coupled with tcell or introduce unnecessary translation to/from tcell types.

### Send pty file descriptor over UDS (instead of using a socketpair)

Rejected due to a subtle bug in macOS terminal driver: when the `pts` is closed by a subcommand, macOS closes the pty even if the server maintains an open FD for the `pts`. Using a socketpair to proxy tty input/output avoids any interactions with the OS terminal drivers.

### LSP daemonization

`gopls` has a "daemon" mode for vim that shares state across multiple LSP instances. However, other LSPs do not provide similar functionality, and there's no straightforward way to adapt them to this architecture.
