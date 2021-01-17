package config

import (
	"os"
	"strings"
)

// GlobMatch checks if a path matches a glob pattern.
// A "*" in the pattern is a wildcard that matches part of a component in a path.
// A "**" in a component of the pattern is a wildcard that matches components of the path.
// For example, the pattern "**/*.go" matches "foo/bar/baz.go"
// The algorithm is based on the backtracking approach described in https://research.swtch.com/glob
func GlobMatch(pattern, name string) bool {
	patternComponents := splitPathComponents(pattern)
	nameComponents := splitPathComponents(name)
	i, j := 0, 0
	bti, btj := 0, 0 // backtrack indices

	for i < len(patternComponents) || j < len(nameComponents) {
		if i < len(patternComponents) {
			pc := patternComponents[i]
			if pc == "**" {
				// Wildcard pattern can either consume nothing or the current name component.
				// Assume for now that the wildcard matches nothing.
				// If this assumption fails to produce a match, backtrack to this wildcard
				// and try consuming the current name component.
				bti = i
				btj = j + 1
				i++
				continue
			}

			if j < len(nameComponents) {
				nc := nameComponents[j]
				if componentsMatch(pc, nc) {
					// The current pattern component matches the current name component.
					// Consume both components and continue.
					i++
					j++
					continue
				}
			}
		}

		if 0 < btj && btj <= len(nameComponents) {
			// Found a mismatched component, so backtrack and try again.
			i = bti
			j = btj
			continue
		}

		// There are mismatched components and we can't backtrack, so this is a mismatch.
		return false
	}

	// All the components match.
	return true
}

// splitPathComponent splits a path into components using the OS-specific separator (e.g. "/" for unix).
func splitPathComponents(path string) []string {
	return strings.Split(path, string(os.PathSeparator))
}

// componentsMatch checks if a component in the pattern matches a component in the path.
func componentsMatch(pc, nc string) bool {
	i, j := 0, 0
	bti, btj := 0, 0 // backtrack indices

	for i < len(pc) || j < len(nc) {
		if i < len(pc) {
			p := pc[i]
			if p == '*' {
				// Wildcard pattern can either consume nothing or the current character.
				// Assume for now that the wildcard matches nothing.
				// If this assumption fails to produce a match, backtrack to this wildcard
				// and try consuming the current character.
				bti = i
				btj = j + 1
				i++
				continue
			}

			if j < len(nc) {
				n := nc[j]
				if p == n {
					// Current character is an exact match.
					// Consume the character from both the pattern and name, then continue.
					i++
					j++
					continue
				}
			}
		}

		if 0 < btj && btj <= len(nc) {
			// Found a mismatched component, so backtrack and try again.
			i = bti
			j = btj
			continue
		}

		// Name does not match the pattern.
		return false
	}

	// Name matches the pattern.
	return true
}
