package shellcmd

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// FileLocation represents a location within a file.
type FileLocation struct {
	Path    string // Path to the file.
	LineNum uint64 // Line number in the file.
	Snippet string // Snippet of text from the location.
}

// FileLocationsFromLines parses each non-empty line as a file location.
// The supported file location formats are:
//
//	<file>:<line>:<col>:<snippet>
//	<file>:<line>:<snippet>
//
// which correspond to the outputs to `grep -n` and `ripgrep --vimgrep`
// If any line cannot be parsed, this function aborts and returns an error.
func FileLocationsFromLines(r io.Reader) ([]FileLocation, error) {
	var fileLocations []FileLocation
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		loc, err := parseFileLocation(line)
		if err != nil {
			return nil, err
		}
		fileLocations = append(fileLocations, loc)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanner.Scan: %w", err)
	}

	return fileLocations, nil
}

func parseFileLocation(s string) (FileLocation, error) {
	parts := strings.SplitN(s, ":", 4)
	var filePart, lineNumPart, snippetPart string
	switch len(parts) {
	case 4: // <file>:<line>:<col>:<snippet>
		filePart, lineNumPart, snippetPart = parts[0], parts[1], parts[3]
	case 3: // <file>:<line>:<snippet>
		filePart, lineNumPart, snippetPart = parts[0], parts[1], parts[2]
	default:
		return FileLocation{}, fmt.Errorf("Unsupported format for file location: %q", s)
	}

	lineNum, err := parseLineNum(lineNumPart)
	if err != nil {
		return FileLocation{}, err
	}

	loc := FileLocation{
		Path:    filePart,
		LineNum: lineNum,
		Snippet: strings.TrimSpace(snippetPart),
	}
	return loc, nil
}

func parseLineNum(s string) (uint64, error) {
	lineNum, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("Invalid line number in file location %q", s)
	}

	return uint64(lineNum), nil
}
