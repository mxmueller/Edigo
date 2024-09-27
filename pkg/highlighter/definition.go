package highlighter

import (
	"regexp"

	"github.com/charmbracelet/lipgloss"
)


type Token struct {
	Color  lipgloss.Style
	Value string
}

type SyntaxDefinition struct {
	Keywords    []string
	Operators   []string
	Punctuation []string
	Rules       []Rule
	KeywordStyle lipgloss.Style
	StringStyle lipgloss.Style
	NumberStyle lipgloss.Style
	OperatorStyle lipgloss.Style
	PunctuationStyle lipgloss.Style
	DefaultStyle lipgloss.Style
}

type Rule struct {
	Pattern *regexp.Regexp
	Color  lipgloss.Style
    Index   int
}

func NewPythonSyntaxDefinition() *SyntaxDefinition {
	return &SyntaxDefinition{
		Keywords:    []string{"def", "return", "if", "else", "for", "while", "import", "from", "class", "try", "except", "lambda"},
		Operators:   []string{"+", "-", "*", "/", "%", "**", "==", "!=", "<", ">", "<=", ">=", "and", "or", "not", "in", "is"},
		Punctuation: []string{"(", ")", "{", "}", "[", "]", ",", ":", "."},
		Rules: []Rule{
			{regexp.MustCompile(`\b[0-9]+(\.[0-9]+)?\b`), lipgloss.NewStyle().Foreground(lipgloss.Color("#0000FF")), 0},
			{regexp.MustCompile(`"[^"]*"`), lipgloss.NewStyle().Foreground(lipgloss.Color("#41B3A2")), 0},
			{regexp.MustCompile(`'[^']*'`), lipgloss.NewStyle().Foreground(lipgloss.Color("#41B3A2")), 0},
			{regexp.MustCompile(`#.*$`), lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA")), 0},
			{regexp.MustCompile(`def (\w*)\(`), lipgloss.NewStyle().Foreground(lipgloss.Color("#6C48C5")), 1},
			{regexp.MustCompile("print"), lipgloss.NewStyle().Foreground(lipgloss.Color("#6C48C5")), 0},
		},
        KeywordStyle :     lipgloss.NewStyle().Foreground(lipgloss.Color("#6C48C5")),
        StringStyle :      lipgloss.NewStyle().Foreground(lipgloss.Color("#41B3A2")),
        NumberStyle :      lipgloss.NewStyle().Foreground(lipgloss.Color("#0000FF")),
        OperatorStyle :    lipgloss.NewStyle().Foreground(lipgloss.Color("#FF00FF")),
        PunctuationStyle : lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF")),
	}
}

func NewJavaScriptSyntaxDefinition() *SyntaxDefinition {
	return &SyntaxDefinition{
		Keywords:    []string{"function", "return", "if", "else", "for", "while", "let", "const", "var", "class", "import", "export"},
		Operators:   []string{"+", "-", "*", "/", "%", "===", "==", "!==", "!=", "<", ">", "<=", ">=", "&&", "||", "!"},
		Punctuation: []string{"(", ")", "{", "}", "[", "]", ",", ";", "."},
		Rules: []Rule{
			{regexp.MustCompile(`\b[0-9]+(\.[0-9]+)?\b`), lipgloss.NewStyle().Foreground(lipgloss.Color("#0000FF")), 0},
			{regexp.MustCompile(`"[^"]*"`), lipgloss.NewStyle().Foreground(lipgloss.Color("#41B3A2")), 0},
			{regexp.MustCompile(`'[^']*'`), lipgloss.NewStyle().Foreground(lipgloss.Color("#41B3A2")), 0},
			{regexp.MustCompile(`//.*$`), lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA")), 0},
			{regexp.MustCompile(`/\*[\s\S]*?\*/`), lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA")), 0},
			{regexp.MustCompile(`function (\w*)\(`), lipgloss.NewStyle().Foreground(lipgloss.Color("#6C48C5")), 1},
		},
        KeywordStyle :     lipgloss.NewStyle().Foreground(lipgloss.Color("#6C48C5")),
        StringStyle :      lipgloss.NewStyle().Foreground(lipgloss.Color("#41B3A2")),
        NumberStyle :      lipgloss.NewStyle().Foreground(lipgloss.Color("#0000FF")),
        OperatorStyle :    lipgloss.NewStyle().Foreground(lipgloss.Color("#FF00FF")),
        PunctuationStyle : lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF")),
	}
}

func NewHTMLSyntaxDefinition() *SyntaxDefinition {
	return &SyntaxDefinition{
		Keywords:    []string{"html", "head", "title", "body", "div", "span", "a", "p", "h1", "h2", "h3", "h4", "h5", "h6", "ul", "li", "table", "tr", "td", "img", "script", "style"},
		Operators:   []string{"=", "/"},
		Punctuation: []string{"<", ">", "/", "\"", "'", "="},
		Rules: []Rule{
			{regexp.MustCompile(`<!DOCTYPE[^>]+>`), lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA")), 0},
			{regexp.MustCompile(`<([^>]+)>`), lipgloss.NewStyle().Foreground(lipgloss.Color("#6C48C5")), 1},
			{regexp.MustCompile(`"[^"]*"`), lipgloss.NewStyle().Foreground(lipgloss.Color("#41B3A2")), 0},
			{regexp.MustCompile(`(<!--[\s\S]*?-->)`), lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA")), 1},
		},
        KeywordStyle :     lipgloss.NewStyle().Foreground(lipgloss.Color("#6C48C5")),
        StringStyle :      lipgloss.NewStyle().Foreground(lipgloss.Color("#41B3A2")),
        OperatorStyle :    lipgloss.NewStyle().Foreground(lipgloss.Color("#FF00FF")),
        PunctuationStyle : lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF")),
	}
}

func NewDefaultSyntaxDefinition() *SyntaxDefinition {
	return &SyntaxDefinition{
		Keywords:    []string{},
		Operators:   []string{},
		Punctuation: []string{},
		Rules:       []Rule{},
        KeywordStyle :     lipgloss.NewStyle(),
        StringStyle :      lipgloss.NewStyle(),
        NumberStyle :      lipgloss.NewStyle(),
        OperatorStyle :    lipgloss.NewStyle(),
        PunctuationStyle : lipgloss.NewStyle(),
	}
}
