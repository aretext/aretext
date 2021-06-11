Command Reference
=================

This document lists every command in aretext.

All commands are compatible with vim keybindings, but not all vim keybindings are implemented. If you want to use a command that is not yet available, please consider contributing to the project!

Normal Mode Commands
--------------------

| Name                                                     | Key Binding         |
|----------------------------------------------------------|---------------------|
| cursor left                                              | left arrow          |
| cursor right                                             | right arrow         |
| cursor up                                                | up arrow            |
| cursor down                                              | down arrow          |
| cursor left                                              | h                   |
| cursor right                                             | l                   |
| cursor up                                                | k                   |
| cursor down                                              | j                   |
| cursor back                                              | backspace           |
| cursor to next matching character in line                | f\{char\}           |
| cursor to \{count\}'th next matching character in line   | \{count\}f\{char\}  |
| cursor to prev matching character in line                | F\{char\}           |
| cursor to \{count\}'th prev matching character in line   | \{count\}F\{char\}  |
| cursor till next matching character in line              | t\{char\}           |
| cursor till \{count\}'th next matching character in line | \{count\}t\{char\}  |
| cursor till prev matching character in line              | T\{char\}           |
| cursor till \{count\}'th prev matching character in line | \{count\}T\{char\}  |
| cursor next word start                                   | w                   |
| cursor prev word start                                   | b                   |
| cursor next word end                                     | e                   |
| cursor prev paragraph                                    | \{                  |
| cursor next paragraph                                    | \}                  |
| cursor line start                                        | 0                   |
| cursor line start after indentation                      | ^                   |
| cursor line end                                          | $                   |
| cursor start of first line                               | gg                  |
| cursor start of line number                              | \{count\}gg         |
| cursor start of last line                                | G                   |
| scroll up                                                | ctrl-u              |
| scroll down                                              | ctrl-d              |
| delete next character in line                            | x                   |
| delete next \{count\} characters in line                 | \{count\}x          |
| delete to next matching character in line                | df\{char\}          |
| delete to \{count\}'th next matching character in line   | \{count\}df\{char\} |
| delete to prev matching character in line                | dF\{char\}          |
| delete to \{count\}'th prev matching character in line   | \{count\}dF\{char\} |
| delete till next matching character in line              | dt\{char\}          |
| delete till \{count\}'th next matching character in line | \{count\}dt\{char\} |
| delete till prev matching character in line              | dT\{char\}          |
| delete till \{count\}'th prev matching character in line | \{count\}dT\{char\} |
| insert                                                   | i                   |
| insert at start of line                                  | I                   |
| append                                                   | a                   |
| append at end of line                                    | A                   |
| new line below                                           | o                   |
| new line above                                           | O                   |
| join lines                                               | J                   |
| delete line                                              | dd                  |
| delete {count} lines down                                | \{count\}dd         |
| delete previous character in line                        | dh                  |
| delete lines below                                       | dj                  |
| delete lines above                                       | dk                  |
| delete next character in line                            | dl                  |
| delete next \{count\} characters in line                 | \{count\}dl         |
| delete to end of line                                    | d$                  |
| delete to start of line                                  | d0                  |
| delete to start of line after indentation                | d^                  |
| delete to end of line                                    | D                   |
| delete to start of next word                             | dw                  |
| delete a word                                            | daw                 |
| delete inner word                                        | diw                 |
| change to start of next word                             | cw                  |
| change a word                                            | caw                 |
| change inner word                                        | ciw                 |
| replace character                                        | r                   |
| toggle case                                              | ~                   |
| indent line                                              | >>                  |
| outdent line                                             | \<\<                |
| yank to start of next word                               | yw                  |
| yank a word                                              | yaw                 |
| yank inner word                                          | yiw                 |
| yank line                                                | yy                  |
| put after cursor                                         | p                   |
| put before cursor                                        | P                   |
| show command menu                                        | :                   |
| start forward search                                     | /                   |
| start backward search                                    | ?                   |
| find next match                                          | n                   |
| find previous match                                      | N                   |
| undo                                                     | u                   |
| redo                                                     | ctrl-r              |
| visual mode charwise                                     | v                   |
| visual mode linewise                                     | V                   |
| repeat last action                                       | .                   |

Visual Mode Commands
--------------------

| Name                        | Key Binding |
|-----------------------------|-------------|
| toggle visual mode charwise | v           |
| toggle visual mode linewise | V           |
| return to normal mode       | escape      |
| show command menu           | :           |
| delete selection            | x           |
| delete selection            | d           |
| change selection            | c           |
| toggle case for selection   | ~           |
| indent selection            | \>          |
| outdent selection           | \<          |
| yank selection              | y           |

Menu Commands
-------------

| Name                         | Aliases  |
|------------------------------|----------|
| quit                         | q        |
| force quit                   | q!       |
| save document                | s, w     |
| save document and quit       | sq, wq   |
| force save document          | s!, w!   |
| force save document and quit | sq!, wq! |
| force reload                 | r!       |
| find and open                | f        |
| open previous document       | p        |
| open next document           | n        |
| toggle show tabs             | ta       |
| toggle line numbers          | nu       |
| toggle auto-indent           | ai       |
| set syntax plaintext         |          |
| set syntax json              |          |
| set syntax yaml              |          |
| set syntax go                |          |
| set syntax gitcommit         |          |
| set syntax gitrebase         |          |
