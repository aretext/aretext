package syntax

import (
	"github.com/aretext/aretext/syntax/languages"
	"github.com/aretext/aretext/syntax/parser"
)

// Language is an enum of languages that we can parse.
type Language string

// AllLanguages lists every available language.
var AllLanguages []Language

const (
	LanguagePlaintext    = Language("plaintext")
	LanguageJson         = Language("json")
	LanguageYaml         = Language("yaml")
	LanguageGo           = Language("go")
	LanguageGoTemplate   = Language("gotemplate")
	LanguagePython       = Language("python")
	LanguageRust         = Language("rust")
	LanguageC            = Language("c")
	LanguageBash         = Language("bash")
	LanguageXml          = Language("xml")
	LanguageGitCommit    = Language("gitcommit")
	LanguageGitRebase    = Language("gitrebase")
	LanguageProtobuf     = Language("protobuf")
	LanguageTodoTxt      = Language("todotxt")
	LanguageMarkdown     = Language("markdown")
	LanguageCriticMarkup = Language("criticmarkup")
	LanguageMakefile     = Language("makefile")
	LanguageP4           = Language("p4")
	LanguageSQL          = Language("sql")
)

// languageToParseFunc maps each language to its parse func.
var languageToParseFunc map[Language]parser.Func

func init() {
	languageToParseFunc = map[Language]parser.Func{
		LanguagePlaintext:    nil,
		LanguageJson:         languages.JsonParseFunc(),
		LanguageYaml:         languages.YamlParseFunc(),
		LanguageGo:           languages.GolangParseFunc(),
		LanguageGoTemplate:   languages.GoTemplateParseFunc(),
		LanguagePython:       languages.PythonParseFunc(),
		LanguageRust:         languages.RustParseFunc(),
		LanguageC:            languages.CParseFunc(),
		LanguageBash:         languages.BashParseFunc(),
		LanguageXml:          languages.XmlParseFunc(),
		LanguageGitCommit:    languages.GitCommitParseFunc(),
		LanguageGitRebase:    languages.GitRebaseParseFunc(),
		LanguageProtobuf:     languages.ProtobufParseFunc(),
		LanguageTodoTxt:      languages.TodoTxtParseFunc(),
		LanguageMarkdown:     languages.MarkdownParseFunc(),
		LanguageCriticMarkup: languages.CriticMarkupParseFunc(),
		LanguageMakefile:     languages.MakefileParseFunc(),
		LanguageP4:           languages.P4ParseFunc(),
		LanguageSQL:          languages.SQLParseFunc(),
	}

	for language := range languageToParseFunc {
		AllLanguages = append(AllLanguages, language)
	}
}

// ParseForLanguage creates a parser for a syntax language.
// If no parser is available (e.g. for LanguagePlaintext) this returns nil.
func ParserForLanguage(language Language) *parser.P {
	parseFunc := languageToParseFunc[language]
	if parseFunc == nil {
		return nil
	}
	return parser.New(parseFunc)
}
