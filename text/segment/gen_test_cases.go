//go:build ignore

package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/template"
)

var (
	prefix     string
	dataPath   string
	outputPath string
)

func main() {
	flag.StringVar(&prefix, "prefix", "", "Prefix for test case names")
	flag.StringVar(&dataPath, "dataPath", "", "Input data file with unicode test cases")
	flag.StringVar(&outputPath, "outputPath", "", "Output path for generated go file")
	flag.Parse()

	if len(prefix) < 1 {
		log.Fatalf("Must specify a prefix")
	}

	if len(dataPath) < 1 {
		log.Fatalf("Must specify input data path")
	}

	if len(outputPath) < 1 {
		log.Fatalf("Must specify output path")
	}

	fmt.Printf("Generating %s from %s\n", outputPath, dataPath)

	testCases, err := parseDataFile(dataPath)
	if err != nil {
		log.Fatalf("error loading test cases from data file at %s\n: %v", dataPath, err)
	}

	testCaseGroups := groupTestCases(testCases)

	if err := writeOutputFile(prefix, outputPath, testCaseGroups); err != nil {
		log.Fatalf("error generating output file %s: %v", outputPath, err)
	}
}

type TestCase struct {
	InputString string
	Segments    [][]rune
	Description string
}

func parseDataFile(path string) ([]TestCase, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	testCases := make([]TestCase, 0)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		hasTestCase, tc, err := parseLine(line)
		if err != nil {
			return nil, err
		}

		if hasTestCase {
			testCases = append(testCases, tc)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return testCases, nil
}

var DESCRIPTION_RE = regexp.MustCompile(`^.+#\s*(.+)$`)
var BREAKPOINT_RE = regexp.MustCompile(`([÷×])\s*([0-9A-Z]+)`)

func parseLine(line string) (bool, TestCase, error) {
	if len(line) == 0 || line[0] == '#' {
		return false, TestCase{}, nil
	}

	descriptionMatch := DESCRIPTION_RE.FindStringSubmatch(line)
	if descriptionMatch == nil {
		return false, TestCase{}, nil
	}
	description := descriptionMatch[1]

	breakpointMatches := BREAKPOINT_RE.FindAllStringSubmatch(line, -1)

	segments := make([][]rune, 0)
	var sb strings.Builder
	var prevPos int
	for pos, m := range breakpointMatches {
		breakResult, hexCode := m[1], m[2]

		if breakResult == "÷" && pos > 0 {
			s := []rune(sb.String())
			seg := make([]rune, len(s)-prevPos)
			copy(seg, s[prevPos:])
			segments = append(segments, seg)
			prevPos = pos
		}

		codePoint, err := strconv.ParseUint(hexCode, 16, 32)
		if err != nil {
			return false, TestCase{}, err
		}
		sb.WriteRune(rune(codePoint))
	}

	// Always break at the end of text
	s := []rune(sb.String())
	seg := make([]rune, len(s)-prevPos)
	copy(seg, s[prevPos:])
	segments = append(segments, seg)

	tc := TestCase{
		InputString: sb.String(),
		Segments:    segments,
		Description: description,
	}

	return true, tc, nil
}

func groupTestCases(testCases []TestCase) [][]TestCase {
	const maxGroupSize = 256
	var groups [][]TestCase
	for start := 0; start < len(testCases); start += maxGroupSize {
		end := len(testCases)
		if end-start > maxGroupSize {
			end = start + maxGroupSize
		}
		groups = append(groups, testCases[start:end])
	}
	return groups
}

func writeOutputFile(prefix string, path string, testCaseGroups [][]TestCase) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	tmpl, err := template.New("output").Parse(`
	// This file is generated by gen_test_cases.go. DO NOT EDIT.

	package segment

	{{ $prefix := .Prefix }}

	type {{ $prefix }}TestCase struct {
		inputString string
		segments    [][]rune
		description string
	}

	func {{ $prefix }}TestCases() []{{ $prefix }}TestCase {
		// Split test cases into groups as a workaround for
		// https://github.com/golang/go/issues/33437
		var testCases []{{ $prefix }}TestCase
		{{ range $i, $testCases := .TestCaseGroups -}}
		testCases = append(testCases, {{ $prefix }}testCaseGroup{{ $i }}()...)
		{{ end -}}
		return testCases
	}

	{{ range $i, $testCases := .TestCaseGroups }}
	func {{ $prefix }}testCaseGroup{{ $i }}() []{{ $prefix }}TestCase {
		return []{{ $prefix }}TestCase{
			{{ range $testCases -}}
			{
				inputString: {{ printf "%#v" .InputString }},
				segments: {{ printf "%#v" .Segments }},
				description: "{{ .Description }}",
			},
			{{ end }}
		}
	}
	{{ end }}
	`)

	if err != nil {
		return err
	}

	return tmpl.Execute(file, map[string]any{
		"Prefix":         prefix,
		"TestCaseGroups": testCaseGroups,
	})
}
