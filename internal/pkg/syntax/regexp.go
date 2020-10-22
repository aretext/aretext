package syntax

import "errors"

// Regexp represents a regular expression.
type Regexp interface{}

// regexpConcat represents the concatenation operation.
type regexpConcat struct {
	left  Regexp
	right Regexp
}

// regexpUnion represents the union operation.
type regexpUnion struct {
	left  Regexp
	right Regexp
}

// regexpStar represents the Kleene star operation.
type regexpStar struct {
	child Regexp
}

// regexpParenExpr represents an expression in parentheses.
type regexpParenExpr struct {
	child Regexp
}

// regexpChar represents a character match in the regular expression.
type regexpChar struct {
	char byte
}

// ParseRegexp parses a regular expression string.
func ParseRegexp(s string) (Regexp, error) {
	regexp, _, err := parseRegexp(s, 0, false)
	return regexp, err
}

func parseRegexp(s string, pos int, inParen bool) (Regexp, int, error) {
	if pos >= len(s) {
		return nil, 0, errors.New("Unexpected end of regular expression")
	}

	var regexp Regexp
	for pos < len(s) {
		switch s[pos] {
		case '(':
			nextRegexp, newPos, err := parseRegexp(s, pos+1, true)
			if err != nil {
				return nil, 0, err
			}

			if newPos >= len(s) || s[newPos] != ')' {
				return nil, 0, errors.New("Expected closing paren")
			}

			if regexp == nil {
				regexp = regexpParenExpr{child: nextRegexp}
			} else {
				regexp = regexpConcat{
					left: regexp,
					right: regexpParenExpr{
						child: nextRegexp,
					},
				}
			}

			pos = newPos + 1

		case ')':
			if regexp == nil || !inParen {
				return nil, 0, errors.New("Unexpected closing paren")
			}
			return regexp, pos, nil

		case '|':
			nextRegexp, newPos, err := parseRegexp(s, pos+1, inParen)
			if err != nil {
				return nil, 0, err
			}
			if regexp == nil {
				return nil, 0, errors.New("Expected characters before union")
			}
			regexp = regexpUnion{left: regexp, right: nextRegexp}
			pos = newPos

		case '*':
			if regexp == nil {
				return nil, 0, errors.New("Expected characters before star")
			}

			if concat, ok := regexp.(regexpConcat); ok {
				regexp = regexpConcat{
					left:  concat.left,
					right: regexpStar{child: concat.right},
				}
			} else {
				regexp = regexpStar{child: regexp}
			}
			pos++

		case '+':
			if regexp == nil {
				return nil, 0, errors.New("Expected characters before plus")
			}

			if concat, ok := regexp.(regexpConcat); ok {
				regexp = regexpConcat{
					left: concat.left,
					right: regexpConcat{
						left:  concat.right,
						right: regexpStar{child: concat.right},
					},
				}
			} else {
				regexp = regexpConcat{
					left:  regexp,
					right: regexpStar{child: regexp},
				}
			}
			pos++

		case '\\':
			nextRegexp, newPos, err := parseEscapeSequence(s, pos)
			if err != nil {
				return nil, 0, err
			}

			if regexp == nil {
				regexp = nextRegexp
			} else {
				regexp = regexpConcat{left: regexp, right: nextRegexp}
			}

			pos = newPos

		default:
			nextRegexp := regexpChar{char: s[pos]}
			if regexp == nil {
				regexp = nextRegexp
			} else {
				regexp = regexpConcat{left: regexp, right: nextRegexp}
			}
			pos++
		}
	}
	return regexp, pos, nil
}

func parseEscapeSequence(s string, pos int) (Regexp, int, error) {
	if pos+1 >= len(s) {
		return nil, 0, errors.New("Invalid escape sequence")
	}

	if c := s[pos+1]; c == '*' || c == '(' || c == ')' || c == '\\' || c == '|' || c == '+' {
		return regexpChar{char: c}, pos + 2, nil
	}

	return nil, 0, errors.New("Unrecognized escape sequence")
}
