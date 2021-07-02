Parser Redesign
---------------

Overview
========

Aretext tokenizes documents using a stateless discrete finite automata (DFA). This approach provides fast and accurate syntax highlighting for programming languages whose tokens can be recognized by a regular language. Unfortunately, not all languages meet this criteria: in particular, the YAML grammar uses context-sensitive rules to identify tokens such as keys, strings, and numbers.

This proposal attempts to increase the expressive power of aretext's syntax highlighting while preserving the simplicity and performance of the current implementation.

Design
======

Each tokenizer is implemented using a function to produce the next token starting from a position in the text. For stateful tokenizers, the function will accept and return an opaque `state` value.

The tokenizer results are cached in a search tree. Each subtree represents an ordered set of tokens, augmented with a "lookahead range" representing the text positions examined while producing the tokens in the subtree. The search tree is immutable; all insert and merge operations are implemented by creating new nodes that may point to nodes in the original tree.

The retokenization algorithm repeatedly queries the tree to find the largest subtree starting at the current position and tokenizer state. If the position occurs on or after the edit range, it is adjusted to match the positions from before the edit. There are four cases to consider:

1.	No matching subtree is found, so we need to rerun the tokenizer.
2.	The subtree's lookahead range intersects the edit range, so it may have been invalidated by the edit.
3.	The subtree's lookahead range ends before the edit range, so it can be reused.
4.	The subtree's lookahead range starts after the edit range, so it can be reused once its token positions are shifted to account for the edit.

This algorithm ensures that the tokenizer is not rerun for parts of the text unaffected by the edit. If a subtree can be reused, the tokens from the subtree are added to the new result, and the current position and state are updated accordingly. Once the new tree has been constructed, the Go garbage collector will reclaim unused nodes from the old tree.

The search tree must support insertions and merges while remaining (approximately) balanced. The "join" algorithms described in Blelloch, et. al (2016) should be sufficient.

Additionally, the search tree must be able to shift all token positions in constant time. This can be achieved by storing an "offset" at each inner node and the calculating the effective position when retrieving the subtree (if the offset is nonzero, this may require creating a copy of the subtree root with an adjusted offset).

TODO: more stuff to think about:
	How to efficiently walk the leaves?  Need a walker that keeps track of the path so it can traverse the leaves?  Basically just the stack.
	How to clean up the stateful reader stuff, rewinding is really awkward an error-prone.
	No more subtokenizers, just implement it in code.
	Lexer pre- (and maybe post-) patterns?
	No more word boundaries based on syntax tokens.
