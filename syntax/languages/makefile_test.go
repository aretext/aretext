package languages

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aretext/aretext/syntax/parser"
)

func TestMakefileParseFunc(t *testing.T) {
	testCases := []struct {
		name     string
		text     string
		expected []TokenWithText
	}{
		{
			name: "comment",
			text: `# comment`,
			expected: []TokenWithText{
				{Text: `# comment`, Role: parser.TokenRoleComment},
			},
		},
		{
			name: "comment with colon",
			text: `# notatarget: xyz`,
			expected: []TokenWithText{
				{Text: `# notatarget: xyz`, Role: parser.TokenRoleComment},
			},
		},
		{
			name: "rule",
			text: `
foo: bar.c
	cc bar.c -o foo

bar.c:
	echo "test" > bar.c
`,
			expected: []TokenWithText{},
		},
		{
			name: "variable assignment",
			text: `
CC := gcc
x ?= maybe
y = foobar
out != echo "hello"
`,
			expected: []TokenWithText{
				{Text: `:=`, Role: parser.TokenRoleOperator},
				{Text: `?=`, Role: parser.TokenRoleOperator},
				{Text: `=`, Role: parser.TokenRoleOperator},
				{Text: `!=`, Role: parser.TokenRoleOperator},
			},
		},
		{
			name: "variable assignment with %",
			text: "BUILD_DATE   := `date +%FT%T%z`",
			expected: []TokenWithText{
				{Text: `:=`, Role: parser.TokenRoleOperator},
			},
		},
		{
			name: "variable assignment with backslash line continuation",
			text: `
VAR = abc \
% \
xyz

%.o: %.c
	$(CC) -o $@
`,
			expected: []TokenWithText{
				{Text: "=", Role: parser.TokenRoleOperator},
				{Text: "%", Role: makefileTokenRolePattern},
				{Text: "%", Role: makefileTokenRolePattern},
				{Text: "$(CC)", Role: makefileTokenRoleVariable},
				{Text: "$@", Role: makefileTokenRoleVariable},
			},
		},
		{
			name: "command with colon",
			text: `
test:
	echo "foo:bar"
`,
			expected: []TokenWithText{},
		},
		{
			name: "variable assignment followed by colon",
			text: `
ARETEXT_URL ?= https://aretext.org
`,
			expected: []TokenWithText{
				{Text: `?=`, Role: parser.TokenRoleOperator},
			},
		},
		{
			name: "variable assignment with expansion",
			text: `COMMIT       := $(shell git rev-parse --short HEAD)`,
			expected: []TokenWithText{
				{Text: `:=`, Role: parser.TokenRoleOperator},
				{Text: `$(shell git rev-parse --short HEAD)`, Role: makefileTokenRoleVariable},
			},
		},
		{
			name: "variable with nested expansions",
			text: `$(${$(VAR)} abc)`,
			expected: []TokenWithText{
				{Text: `$(${$(VAR)} abc)`, Role: makefileTokenRoleVariable},
			},
		},
		{
			name: "rule with variables",
			text: `
objects = program.o foo.o utils.o
program : $(objects)
        cc -o program $(objects)

$(objects) : defs.h
`,
			expected: []TokenWithText{
				{Text: "=", Role: parser.TokenRoleOperator},
				{Text: "$(objects)", Role: makefileTokenRoleVariable},
				{Text: "$(objects)", Role: makefileTokenRoleVariable},
				{Text: "$(objects)", Role: makefileTokenRoleVariable},
			},
		},
		{
			name: "rule with automatic variables",
			text: `
%.o: %.c
        $(CC) -c $(CFLAGS) $(CPPFLAGS) $< -o $@
`,
			expected: []TokenWithText{
				{Text: "%", Role: makefileTokenRolePattern},
				{Text: "%", Role: makefileTokenRolePattern},
				{Text: "$(CC)", Role: makefileTokenRoleVariable},
				{Text: "$(CFLAGS)", Role: makefileTokenRoleVariable},
				{Text: "$(CPPFLAGS)", Role: makefileTokenRoleVariable},
				{Text: "$<", Role: makefileTokenRoleVariable},
				{Text: "$@", Role: makefileTokenRoleVariable},
			},
		},
		{
			name: "escaped dollar sign",
			text: `
foo: foo.c
	echo "$${SHELL}"
`,
			expected: []TokenWithText{},
		},
		{
			name: "@ suppress output in command",
			text: `
foo: foo.c
	@ echo "test"
	@echo "test2"
`,
			expected: []TokenWithText{
				{Text: "@", Role: parser.TokenRoleOperator},
				{Text: "@", Role: parser.TokenRoleOperator},
			},
		},
		{
			name: "conditional",
			text: `
ifeq ($(CC),gcc)
  libs=$(libs_for_gcc)
else
  libs=$(normal_libs)
endif
`,
			expected: []TokenWithText{
				{Text: "ifeq", Role: parser.TokenRoleKeyword},
				{Text: "$(CC)", Role: makefileTokenRoleVariable},
				{Text: "=", Role: parser.TokenRoleOperator},
				{Text: "$(libs_for_gcc)", Role: makefileTokenRoleVariable},
				{Text: "else", Role: parser.TokenRoleKeyword},
				{Text: "=", Role: parser.TokenRoleOperator},
				{Text: "$(normal_libs)", Role: makefileTokenRoleVariable},
				{Text: "endif", Role: parser.TokenRoleKeyword},
			},
		},
		{
			name: "one-line recipe with semicolon",
			text: `all : ; @echo 'hello'`,
			expected: []TokenWithText{
				{Text: "@", Role: parser.TokenRoleOperator},
			},
		},
		{
			name: "define",
			text: `
define run-yacc =
yacc $(firstword $^)
mv y.tab.c $@
endef
`,
			expected: []TokenWithText{
				{Text: "define", Role: parser.TokenRoleKeyword},
				{Text: "=", Role: parser.TokenRoleOperator},
				{Text: "$(firstword $^)", Role: makefileTokenRoleVariable},
				{Text: "$@", Role: makefileTokenRoleVariable},
				{Text: "endef", Role: parser.TokenRoleKeyword},
			},
		},
		{
			name: "include",
			text: `include ../Makefile.defs`,
			expected: []TokenWithText{
				{Text: "include", Role: parser.TokenRoleKeyword},
			},
		},
		{
			name: "include with error suppression",
			text: `-include $(DEPS)`,
			expected: []TokenWithText{
				{Text: "-include", Role: parser.TokenRoleKeyword},
				{Text: "$(DEPS)", Role: makefileTokenRoleVariable},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokens := ParseTokensWithText(MakefileParseFunc(), tc.text)
			assert.Equal(t, tc.expected, tokens)
		})
	}
}

func FuzzMakefileParseFunc(f *testing.F) {
	seeds := LoadFuzzTestSeeds(f, "./testdata/makefile/*")
	FuzzParser(f, MakefileParseFunc(), seeds)
}
