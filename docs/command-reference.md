Command Reference
=================

This document lists every command in aretext.

All commands are compatible with vim keybindings, but not all vim keybindings are implemented. If you want to use a command that is not yet available, please consider contributing to the project!

Normal Mode Commands
--------------------

| Name                                      | Key Binding |
|-------------------------------------------|-------------|
| cursor left                               | left arrow  |
| cursor right                              | right arrow |
| cursor up                                 | up arrow    |
| cursor down                               | down arrow  |
| cursor left                               | h           |
| cursor right                              | l           |
| cursor up                                 | k           |
| cursor down                               | j           |
| cursor back                               | backspace   |
| cursor next word start                    | w           |
| cursor prev word start                    | b           |
| cursor next word end                      | e           |
| cursor prev paragraph                     | \{          |
| cursor next paragraph                     | \}          |
| cursor line start                         | 0           |
| cursor line start after indentation       | ^           |
| cursor line end                           | $           |
| cursor start of first line                | gg          |
| cursor start of line number               | {count}gg   |
| cursor start of last line                 | G           |
| scroll up                                 | ctrl-u      |
| scroll down                               | ctrl-d      |
| delete next character in line             | x           |
| insert                                    | i           |
| insert at start of line                   | I           |
| append                                    | a           |
| append at end of line                     | A           |
| new line below                            | o           |
| new line above                            | O           |
| join lines                                | J           |
| delete line                               | dd          |
| delete previous character in line         | dh          |
| delete lines below                        | dj          |
| delete lines above                        | dk          |
| delete next characater in line            | dl          |
| delete to end of line                     | d$          |
| delete to start of line                   | d0          |
| delete to start of line after indentation | d^          |
| delete to end of line                     | D           |
| delete to start of next word              | dw          |
| delete a word                             | daw         |
| delete inner word                         | diw         |
| change a word                             | caw         |
| change inner word                         | ciw         |
| replace character                         | r           |
| toggle case                               | ~           |
| indent line                               | >>          |
| outdent line                              | \<\<        |
| yank line                                 | yy          |
| put after cursor                          | p           |
| put before cursor                         | P           |
| show command menu                         | :           |
| start forward search                      | /           |
| start backward search                     | ?           |
| find next match                           | n           |
| find previous match                       | N           |
| undo                                      | u           |
| redo                                      | ctrl-r      |
| visual mode charwise                      | v           |
| visual mode linewise                      | V           |
| repeat last action                        | .           |

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
| toggle line numbers          | nu       |
| set syntax plaintext         |          |
| set syntax json              |          |
| set syntax yaml              |          |
| set syntax go                |          |
