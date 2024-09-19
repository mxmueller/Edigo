package highlighter

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// TokenType repräsentiert verschiedene Arten von Tokens in der Sprache
type TokenType int

const (
	TokenKeyword TokenType = iota
	TokenIdentifier
	TokenString
	TokenNumber
	TokenComment
	TokenOperator
	TokenPunctuation
)

// Token repräsentiert ein einzelnes Token im Code
type Token struct {
	Type  TokenType
	Value string
}

// SyntaxDefinition definiert die Regeln für das Syntax-Highlighting
type SyntaxDefinition struct {
	Keywords    []string
	Operators   []string
	Punctuation []string
	Rules       []Rule
}

// Rule repräsentiert eine einzelne Regel für das Matching von Tokens
type Rule struct {
	Pattern *regexp.Regexp
	Type    TokenType
    Index   int
}

// NewSyntaxDefinition erstellt eine neue SyntaxDefinition
func NewSyntaxDefinition() *SyntaxDefinition {
	return &SyntaxDefinition{
		Keywords:    []string{"if", "else", "for", "func", "return", "var", "const"},
		Operators:   []string{"+", "-", "*", "/", "==", "!=", "<", ">", "<=", ">="},
		Punctuation: []string{"(", ")", "{", "}", "[", "]", ",", ";", "."},
		Rules: []Rule{
			{regexp.MustCompile(`\b[0-9]+(\.[0-9]+)?\b`), TokenNumber, 0},
			{regexp.MustCompile(`"[^"]*"`), TokenString, 0},
			{regexp.MustCompile(`//.*$`), TokenComment, 0},
			{regexp.MustCompile(`func (\w*)\(`), TokenNumber, 1},
		},
	}
}

func (sd *SyntaxDefinition) TokenizeLine(line string) []Token {
	var tokens []Token
	remaining := line

	for len(remaining) > 0 {
        remaining = strings.TrimLeft(remaining, " \t")
		if len(remaining) == 0 {
			break
		}

		matched := false

        for _, rule := range sd.Rules {
            match := rule.Pattern.FindStringSubmatch(remaining) 
            if len(match) > rule.Index{
                tokens = append(tokens, Token{rule.Type, match[rule.Index]})
                break
            }
        }

		// Prüfe zuerst auf Keywords, Operatoren und Interpunktion
		for _, keyword := range sd.Keywords {
			if strings.HasPrefix(remaining, keyword) {
				tokens = append(tokens, Token{TokenKeyword, keyword})
				remaining = remaining[len(keyword):]
				matched = true
				break
			}
		}

		if !matched {
			for _, op := range sd.Operators {
				if strings.HasPrefix(remaining, op) {
					tokens = append(tokens, Token{TokenOperator, op})
					remaining = remaining[len(op):]
					matched = true
					break
				}
			}
		}

		if !matched {
			for _, punct := range sd.Punctuation {
				if strings.HasPrefix(remaining, punct) {
					tokens = append(tokens, Token{TokenPunctuation, punct})
					remaining = remaining[len(punct):]
					matched = true
					break
				}
			}
		}

		if !matched {
			tokens = append(tokens, Token{TokenIdentifier, string(remaining[0])})
			remaining = remaining[1:]
		}
	}
	return tokens
}

func stripColorCodes(text string) string {
    // Compile the regular expression for ANSI escape codes
    re := regexp.MustCompile(`\x1B(?:[@-Z\\-_]|\[[0-?]*[ -/]*[@-~])`)
    
    // Replace all occurrences of ANSI escape codes with an empty string
    return re.ReplaceAllString(text, "")
}

func (sd *SyntaxDefinition) ColorText(text string) string {
	var result strings.Builder

    cleanText := stripColorCodes(text)
    tokens := sd.TokenizeLine(cleanText)
	remainingText := text

	keywordStyle :=     lipgloss.NewStyle().Foreground(lipgloss.Color("#6C48C5"))
	stringStyle :=      lipgloss.NewStyle().Foreground(lipgloss.Color("#41B3A2"))
	numberStyle :=      lipgloss.NewStyle().Foreground(lipgloss.Color("#0000FF"))
	commentStyle :=     lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00"))
	operatorStyle :=    lipgloss.NewStyle().Foreground(lipgloss.Color("#FF00FF"))
	punctuationStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF"))

	for len(remainingText) > 0 {
		longestMatch := ""
        var tokenType TokenType

		for _, token := range tokens {
			if strings.HasPrefix(remainingText, token.Value) {
                longestMatch = token.Value
                tokenType = token.Type
                break
			}
		}

		if longestMatch != "" {
            switch tokenType {
                case TokenKeyword:
                    result.WriteString(keywordStyle.Render(longestMatch))
                case TokenString:
                    result.WriteString(stringStyle.Render(longestMatch))
                case TokenNumber:
                    result.WriteString(numberStyle.Render(longestMatch))
                case TokenComment:
                    result.WriteString(commentStyle.Render(longestMatch))
                case TokenOperator:
                    result.WriteString(operatorStyle.Render(longestMatch))
                case TokenPunctuation:
                    result.WriteString(punctuationStyle.Render(longestMatch))
                default:
                    result.WriteString(longestMatch)
		}
			remainingText = remainingText[len(longestMatch):]
		} else {
			result.WriteByte(remainingText[0])
			remainingText = remainingText[1:]
		}
	}

	return result.String()
}
