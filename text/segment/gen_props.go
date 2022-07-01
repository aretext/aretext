//go:build ignore

package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"
)

var prefix string
var dataPaths stringArrayFlags
var propertyNames stringArrayFlags
var outputPath string

func main() {
	flag.StringVar(&prefix, "prefix", "", "Prefix for symbols in the generated go file")
	flag.Var(&dataPaths, "dataPath", "Input data file with unicode property definitions")
	flag.Var(&propertyNames, "propertyName", "Unicode properties to include in the generated go file")
	flag.StringVar(&outputPath, "outputPath", "", "Output path for the generated go file")
	flag.Parse()

	if len(prefix) < 1 {
		log.Fatalf("Must specify prefix")
	}

	if len(dataPaths) < 1 {
		log.Fatalf("Must specify at least one input data path")
	}

	if len(propertyNames) < 1 {
		log.Fatalf("Must specify at least one property name")
	}

	if len(outputPath) < 1 {
		log.Fatalf("Must specify output path")
	}

	propFilter := NewPropFilter(propertyNames)

	fmt.Printf("Generating %s from %s\n", outputPath, dataPaths.String())

	ranges := make([]propRange, 0)
	for _, path := range dataPaths {
		var err error
		ranges, err = parseDataFile(path, propFilter, ranges)
		if err != nil {
			log.Fatalf("error parsing data file %s: %v", path, err)
		}
	}

	sort.Slice(ranges, func(i, j int) bool {
		return ranges[i].Start < ranges[j].Start
	})

	checkNonOverlapping(ranges)
	ranges = fillGaps(ranges)
	ranges = coalesce(ranges)
	propNames := uniquePropNames(ranges)

	if err := writeOutputFile(prefix, outputPath, propNames, ranges); err != nil {
		log.Fatalf("error generating output file %s: %v", outputPath, err)
	}
}

type stringArrayFlags []string

func (f *stringArrayFlags) String() string {
	return fmt.Sprintf("[%s]", strings.Join(*f, ", "))
}

func (f *stringArrayFlags) Set(s string) error {
	*f = append(*f, s)
	return nil
}

type propRange struct {
	Start    uint64
	End      uint64
	PropName string
}

type propFilter struct {
	filterProps map[string]struct{}
}

func NewPropFilter(includeProps []string) *propFilter {
	filterProps := make(map[string]struct{}, 0)
	for _, p := range includeProps {
		filterProps[p] = struct{}{}
	}
	return &propFilter{filterProps}
}

func (f *propFilter) CheckAllowed(prop string) bool {
	_, ok := f.filterProps[prop]
	return ok
}

func parseDataFile(dataPath string, propFilter *propFilter, ranges []propRange) ([]propRange, error) {
	file, err := os.Open(dataPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		hasRange, rng, err := parseLine(line, propFilter)
		if err != nil {
			return nil, err
		}

		if hasRange {
			ranges = append(ranges, rng)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return ranges, nil
}

var LINE_RE = regexp.MustCompile(`^([A-Z0-9]+)(..[A-Z0-9]+)?\s*;\s*([A-z0-9]+)`)

func parseLine(line string, propFilter *propFilter) (bool, propRange, error) {
	match := LINE_RE.FindStringSubmatch(line)
	if match == nil {
		return false, propRange{}, nil
	}

	propName := match[3]
	if !propFilter.CheckAllowed(propName) {
		return false, propRange{}, nil
	}

	startCodepoint, err := strconv.ParseUint(match[1], 16, 32)
	if err != nil {
		return false, propRange{}, err
	}

	var endCodepoint uint64
	endMatch := match[2]
	if len(endMatch) == 0 {
		endCodepoint = startCodepoint
	} else {
		endHex := endMatch[2:] // remove ".." prefix
		endCodepoint, err = strconv.ParseUint(endHex, 16, 32)
		if err != nil {
			return false, propRange{}, err
		}
	}

	rng := propRange{
		Start:    startCodepoint,
		End:      endCodepoint,
		PropName: propName,
	}

	return true, rng, nil
}

func checkNonOverlapping(ranges []propRange) {
	// Assume that ranges are sorted.
	for i := 1; i < len(ranges); i++ {
		if ranges[i-1].End >= ranges[i].Start {
			log.Fatalf("Overlapping range detected between %v and %v\n", ranges[i-1], ranges[i])
		}
	}
}

func fillGaps(ranges []propRange) []propRange {
	// Assume that ranges are sorted by start and non-overlapping.
	var lastRng propRange
	result := make([]propRange, 0, len(ranges))
	for i := 0; i < len(ranges); i++ {
		rng := ranges[i]
		if i == 0 {
			if rng.Start > 0 {
				result = append(result, propRange{
					Start:    0,
					End:      rng.Start - 1,
					PropName: "Other",
				})
			}

			result = append(result, rng)
			lastRng = rng
			continue
		}

		if lastRng.End+1 < rng.Start {
			result = append(result, propRange{
				Start:    lastRng.End + 1,
				End:      rng.Start - 1,
				PropName: "Other",
			})
		}

		result = append(result, rng)
		lastRng = rng
	}
	return result
}

func coalesce(ranges []propRange) []propRange {
	// Assume that ranges are sorted by start and non-overlapping.
	var lastRng propRange
	result := make([]propRange, 0, len(ranges))
	for i := 0; i < len(ranges); i++ {
		rng := ranges[i]
		if i > 0 && lastRng.End+1 == rng.Start && rng.PropName == lastRng.PropName {
			// Current range has the same property as the previous range,
			// so extend the previous range.
			result[len(result)-1].End = rng.End
		} else {
			result = append(result, rng)
		}
		lastRng = rng
	}
	return result
}

func uniquePropNames(ranges []propRange) []string {
	set := make(map[string]struct{}, 0)
	for _, rng := range ranges {
		set[rng.PropName] = struct{}{}
	}

	result := make([]string, 0, len(set))
	for propName := range set {
		result = append(result, propName)
	}

	sort.Strings(result)
	return result
}

func writeOutputFile(prefix string, path string, propNames []string, ranges []propRange) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	tmpl, err := template.New("output").Parse(`
	// This file is generated by gen_props.go. DO NOT EDIT.

	package segment

	{{ $input := . }}

	type {{ $input.Prefix }}Prop byte

	const (
		{{ range $propName := $input.PropNames -}}
		{{ $input.Prefix }}Prop{{ $propName }} = {{ $input.Prefix }}Prop(iota)
		{{ end  }}
	)

	func {{ $input.Prefix }}PropForRune(r rune) {{ $input.Prefix }}Prop {
		switch {
		{{ range $rng := $input.PropRanges }}
		case r <= {{ $rng.End }}:
		return {{ $input.Prefix }}Prop{{ $rng.PropName }}
		{{ end }}

		default:
		return {{ $input.Prefix }}PropOther
		}
	}
	`)
	if err != nil {
		return err
	}

	return tmpl.Execute(file, map[string]interface{}{
		"Prefix":     prefix,
		"PropNames":  propNames,
		"PropRanges": ranges,
	})
}
