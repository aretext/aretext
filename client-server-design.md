Client-Server Redesign
======================

Aretext is currently a single binary. To open multiple documents concurrently, users must launch multiple copies of the editor. This architecture, while simple to understand and implement, has some serious limitations:

1.	There is no synchronization between unsaved changes of the same document opened in multiple editor instances, leading to conflicts on save.
2.	It duplicates document and syntax parser state in the common case of viewing/editing the same document in multiple editor instances.
3.	It greatly complicates language server protocol (LSP) integration, since LSP assumes a single editor per LSP server. parts of a single document.

(These problems aren't unique to aretext: vim and its descendants have the same limitations.)

The solution is a *client-server* architecture. Aretext will run as a daemon process, serving clients connected over a unix domain socket. All editor state will reside in the server, which can then act as an LSP client.

An important design goal is to adopt this architecture with minimal changes to the user experience. From a user perspective, aretext should continue to work as it does today, just with better synchronization of unsaved changes and LSP support.

Out-of-scope: clients connecting from other machines, multiple client implementation, client/server version skew.

Architecture
------------

The **server** is responsible for:

-	listening on a unix domain socket (UDS) and accepting client connections.
-	loading documents from disk, including watching and reloading on changes.
-	receiving and processing input from clients.
-	calculating terminal contents to display (but delegate rendering to the terminal).
-	detecting on-disk changes to all open documents and reloading contents.
-	executing shell commands on behalf of the client.
-	managing LSP lifecycle and communication.

The **clients** are responsible for:

-	proxying terminal input/output to the server, using a mechanism described below.
-	(optionally) starting and daemonizing server process.

To avoid conflict between different user accounts, the UDS path will be `$XDG_RUNTIME_DIR/aretext.socket`.

Client TTY Delegation
---------------------

The client delegates all interactions with its TTY to the server using the following procedure:

1.	Create a pseudoterminal (pty) pair: ptmx (primary) and pts (secondary)
2.	Send pts to the server over UDS using SCM_RIGHTS out-of-band data.
3.	Send a "hello" message to the server encoding the arguments to the client (e.g. the filepath to open).
4.	Copy stdin -> ptmx and ptmx -> stdout until the pty is closed.

The server receives a file descriptor for pts over UDS and uses it to control the client's terminal:

1.	Initialize a tcell `Screen` for the client using the pts file descriptor. This is used for all input/output to/from the editor.
2.	When executing shell commands with `CmdModeTerminal`, use the pts file descriptor as stdin, stdout, stderr, and the controlling terminal for the subprocess.

When tty dimensions change, the client will receive a SIGWINCH signal. To propagate the change, the client must:

1.	Resize ptmx to match the client's tty. This causes the kernel to signal SIGWINCH to all server subprocesses whose controlling terminal is pts.
2.	Send a `WindowResizeMsg` to the server. This allows the server to update screen dimensions for the client's editor.

Client-Server Messages
----------------------

Messages will be serialized as JSON, with a uint32 header indicating msg length.

-	`ClientHelloMsg`:
	-	sent from client to server after connect
	-	fields:
		-	current working directory
		-	filepath to open
	-	out-of-band data:
		-	SCM_RIGHTS with the pty file descriptor.
-	`ServerHelloMsg`:
	-	sent from server to client after connect
	-	fields:
		-	client ID (used for debug logging)
-	`GoodbyeMsg`
	-	sent from either client or server to gracefully terminate
	-	fields:
		-	code: used to differentiate user-initiated quit from an error.
-	`WindowResizeMsg`
	-	sent from client to server on receipt of SIGWINCH signal.
	-	fields:
		-	width
		-	height

The server will close the connection if it receives a message larger than some limit.

Client-Server Lifecycle
-----------------------

The server process may be run by the init system. However, by default the server will be started by a client on-demand and terminate some time after all clients have disconnected.

This follows a procedure similar to tmux:

1.	Attempt to connect to the server over the unix domain socket.
2.	If connect fails, try to acquire a lock.
3.	If lock acquisition fails, retry from (1).
4.	Otherwise, retry connecting to the server (since another client may have started it before this client acquired the lock).
5.	If connection fails again, start the server process as a daemon, then retry from (4).

To ensure at most one server instance at a time, the server attempts to acquires a file lock on startup. If it cannot acquire the lock, it exits with an error.

The client and server can detect that the other side has terminated if writes to/from the pty fail. Under normal conditions, however, the client and server will attempt to send a "Goodbye" message to terminate gracefully.

-	If a client detects the server has terminated, the client should exit immediately.
-	If a server detects a client has terminated, the server should remove all per-client state for that client.

When the user quits from the aretext menu, the server will remove all per-client state, send the client a `GoodbyeMsg`, then close the client's pty.

Editor state
------------

Editor state (owned by the server) must now represent both **shared** data (available to all clients) and **per-client** data. Per-client data is indexed by a unique client ID, which the server assigns on client connection and removes when the client disconnects. Additionally, editor state must include multiple documents loaded at the same time, with each document opened by one or more clients.

Shared data:

-	Document contents (textTree)
-	File watchers (path/hash/timestamp of each open document)
-	Syntax parse trees
-	Undo/redo log
-	Clipboard
-	Configuration

Per-client data:

-	Pty file descriptor
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

When one client modifies a document, the server must update the per-client state of all other clients editing the same document:

1.	All other clients must transition to normal mode (same as if the client pressed "escape"). This implicitly resets most per-client state, including committing any staged undo log operations, clearing selections, etc.
2.	The server then executes the client's action, possibly modifying the document. For each insertion/deletion that occurs before another client's cursor position, increment/decrement the cursor position accordingly.

Shell commands
--------------

All shell commands are executed by the server on behalf of a client. The user is responsible for configuring env vars needed by shell commands the server executes. The shell command's working directory will be set to the current working directory of the client. Shell commands always execute asynchronously to avoid blocking the server.

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
