Command Reference
=================

This document lists every command in aretext.

All commands are compatible with vim keybindings, but not all vim keybindings are implemented. If you want to use a command that is not yet available, please consider contributing to the project!

Normal Mode Commands
--------------------

Some commands may be prefixed with a number *count* to repeat the command *count* times. For example "5x" deletes the next five characters.

Commands that interact with the clipboard accept a *clipboard page* prefix of the form `"[a-z]`, where the letter is the name of the page. If not provided, a default (unnamed) page is used.

| Name                                                            | Key Binding       | Options               |
|-----------------------------------------------------------------|-------------------|-----------------------|
| cursor left                                                     | left arrow        | count                 |
| cursor right                                                    | right arrow       | count                 |
| cursor up                                                       | up arrow          | count                 |
| cursor down                                                     | down arrow        | count                 |
| cursor left                                                     | h                 | count                 |
| cursor right                                                    | l                 | count                 |
| cursor up                                                       | k                 | count                 |
| cursor down                                                     | j                 | count                 |
| cursor forward                                                  | space             | count                 |
| cursor back                                                     | backspace         | count                 |
| cursor start of next line after indentation                     | enter             |                       |
| cursor to next matching character in line                       | f\{char\}         | count                 |
| cursor to prev matching character in line                       | F\{char\}         | count                 |
| cursor till next matching character in line                     | t\{char\}         | count                 |
| cursor till prev matching character in line                     | T\{char\}         | count                 |
| cursor next word start                                          | w                 | count                 |
| cursor next word start, including punctuation                   | W                 | count                 |
| cursor prev word start                                          | b                 | count                 |
| cursor prev word start, including punctuation                   | B                 | count                 |
| cursor next word end                                            | e                 | count                 |
| cursor next word end, including punctuation                     | E                 | count                 |
| cursor prev paragraph                                           | \{                |                       |
| cursor next paragraph                                           | \}                |                       |
| cursor line start                                               | 0                 |                       |
| cursor line start after indentation                             | ^                 |                       |
| cursor line end                                                 | $                 |                       |
| cursor start of first line                                      | gg                |                       |
| cursor start of line number                                     | \{count\}gg       |                       |
| cursor start of last line                                       | G                 |                       |
| cursor matching code block delimiter (paren, brace, or bracket) | %                 |                       |
| cursor prev unmatched open brace                                | [{                |                       |
| cursor next unmatched close brace                               | ]}                |                       |
| cursor prev unmatched open paren                                | [(                |                       |
| cursor next unmatched close paren                               | ])                |                       |
| scroll up (full page)                                           | ctrl-f            |                       |
| scroll down (full page)                                         | ctrl-b            |                       |
| scroll up (half page)                                           | ctrl-u            |                       |
| scroll down (half page)                                         | ctrl-d            |                       |
| insert                                                          | i                 |                       |
| insert at start of line                                         | I                 |                       |
| append                                                          | a                 |                       |
| append at end of line                                           | A                 |                       |
| new line below                                                  | o                 |                       |
| new line above                                                  | O                 |                       |
| join lines                                                      | J                 |                       |
| delete next character in line                                   | x                 | count, clipboard page |
| delete line                                                     | dd                | count, clipboard page |
| delete previous character in line                               | dh                | clipboard page        |
| delete lines below                                              | dj                | clipboard page        |
| delete lines above                                              | dk                | clipboard page        |
| delete next character in line                                   | dl                | count, clipboard page |
| delete to end of line                                           | d$                | clipboard page        |
| delete to start of line                                         | d0                | clipboard page        |
| delete to start of line after indentation                       | d^                | clipboard page        |
| delete to end of line                                           | D                 | clipboard page        |
| delete to start of next word                                    | dw                | count, clipboard page |
| delete to start of next word, including punctuation             | dW                | count, clipboard page |
| delete a word                                                   | daw               | count, clipboard page |
| delete inner word                                               | diw               | count, clipboard page |
| delete to next matching character in line                       | df\{char\}        | count, clipboard page |
| delete to prev matching character in line                       | dF\{char\}        | count, clipboard page |
| delete till next matching character in line                     | dt\{char\}        | count, clipboard page |
| delete till prev matching character in line                     | dT\{char\}        | count, clipboard page |
| delete inner paren block                                        | dib / di\( / di\) | clipboard page        |
| delete a paren block                                            | dab / da\( / da\) | clipboard page        |
| delete inner brace block                                        | diB / di\{ / di\} | clipboard page        |
| delete a brace block                                            | daB / da\{ / da\} | clipboard page        |
| delete inner angle block                                        | di&lt; / di&gt;   | clipboard page        |
| delete an angle block                                           | da&lt; / da&gt;   | clipboard page        |
| change word                                                     | cw                | count, clipboard page |
| change a word                                                   | caw               | count, clipboard page |
| change inner word                                               | ciw               | count, clipboard page |
| change to next matching character in line                       | cf\{char\}        | count, clipboard page |
| change to prev matching character in line                       | cF\{char\}        | count, clipboard page |
| change till next matching character in line                     | ct\{char\}        | count, clipboard page |
| change till prev matching character in line                     | cT\{char\}        | count, clipboard page |
| change inner paren block                                        | cib / ci\( / ci\) | clipboard page        |
| change a paren block                                            | cab / ca\( / ca\) | clipboard page        |
| change inner brace block                                        | ciB / ci\{ / ci\} | clipboard page        |
| change a brace block                                            | caB / ca\{ / ca\} | clipboard page        |
| change inner angle block                                        | ci&lt; / ci&gt;   | clipboard page        |
| change an angle block                                           | ca&lt; / ca&gt;   | clipboard page        |
| replace character                                               | r                 |                       |
| toggle case                                                     | ~                 |                       |
| indent line                                                     | &gt;&gt;          |                       |
| outdent line                                                    | &lt;&lt;          |                       |
| yank to start of next word                                      | yw                | count, clipboard page |
| yank to start of next word, including punctuation               | yW                | count, clipboard page |
| yank a word                                                     | yaw               | count, clipboard page |
| yank inner word                                                 | yiw               | count, clipboard page |
| yank line                                                       | yy                | clipboard page        |
| put after cursor                                                | p                 | clipboard page        |
| put before cursor                                               | P                 | clipboard page        |
| show command menu                                               | :                 |                       |
| start forward search                                            | /                 |                       |
| start backward search                                           | ?                 |                       |
| find next match                                                 | n                 |                       |
| find previous match                                             | N                 |                       |
| search forward for word under cursor                            | \*                | count                 |
| search backward for word under cursor                           | \#                | count                 |
| undo                                                            | u                 |                       |
| redo                                                            | ctrl-r            |                       |
| visual mode charwise                                            | v                 |                       |
| visual mode linewise                                            | V                 |                       |
| repeat last action                                              | .                 |                       |

Visual Mode Commands
--------------------

| Name                        | Key Binding    | Options        |
|-----------------------------|----------------|----------------|
| toggle visual mode charwise | v              |                |
| toggle visual mode linewise | V              |                |
| return to normal mode       | escape         |                |
| show command menu           | :              |                |
| delete selection            | x              | clipboard page |
| delete selection            | d              | clipboard page |
| change selection            | c              | clipboard page |
| toggle case for selection   | ~              |                |
| indent selection            | &gt;           |                |
| outdent selection           | &lt;           |                |
| yank selection              | y              | clipboard page |
| select inner word           | iw             | count          |
| select a word               | aw             | count          |
| select inner paren block    | ib / i\( / i\) |                |
| select a paren block        | ab / a\( / a\) |                |
| select inner brace block    | iB / i\{ / i\} |                |
| select a brace block        | aB / a\{ / a\} |                |
| select inner angle block    | i&lt; / i&gt;  |                |
| select an angle block       | a&lt; / a&gt;  |                |

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
| child directory              | cd       |
| parent directory             | pd       |
| toggle show tabs             | ta       |
| toggle tab expand            | te       |
| toggle line numbers          | nu       |
| toggle auto-indent           | ai       |
| start/stop recording macro   | m        |
| replay macro                 | r        |
