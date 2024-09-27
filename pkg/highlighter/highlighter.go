package highlighter

import (
	"regexp"
	"strings"
	"github.com/charmbracelet/lipgloss"
)


func GetSyntaxDefiniton(fileEnd string) *SyntaxDefinition {
    switch (fileEnd){
        case ".py": {
            return NewPythonSyntaxDefinition(); 
        }
        case ".js": {
            return NewJavaScriptSyntaxDefinition(); 
        }
        case ".html": {
            return NewHTMLSyntaxDefinition(); 
        }
    }
    return NewDefaultSyntaxDefinition()
}

func (sd *SyntaxDefinition) LineLexer(line string) []Token {
	var tokens []Token
	remaining := line

	for len(remaining) > 0 {

		if len(remaining) == 0 {
			break
		}

		matched := false

        for _, rule := range sd.Rules {
            match := rule.Pattern.FindStringSubmatch(remaining)
            if len(match) > rule.Index{
                if rule.Pattern.FindIndex([]byte(remaining))[0] != 0 {continue}

                tokens = append(tokens, Token{rule.Color, match[rule.Index]})
				remaining = remaining[len(match[rule.Index]):]
                matched = true
                break
            }
        }
        if matched{ continue }

		for _, keyword := range sd.Keywords {
			if strings.HasPrefix(remaining, keyword + " ") {
				tokens = append(tokens, Token{sd.KeywordStyle, keyword})
				remaining = remaining[len(keyword):]
				matched = true
				break
			}
		}
        if matched{ continue }

        for _, op := range sd.Operators {
            if strings.HasPrefix(remaining,  op + " ") {
                tokens = append(tokens, Token{sd.OperatorStyle, op})
                remaining = remaining[len(op):]
                matched = true
                break
            }
        }
        if matched{ continue }

        for _, punct := range sd.Punctuation {
            if strings.HasPrefix(remaining, punct) {
                tokens = append(tokens, Token{sd.PunctuationStyle, punct})
                remaining = remaining[len(punct):]
                matched = true
                break
            }
        }
        if matched{ continue }

        remaining = remaining[1:]
	}
	return tokens
}

// input is the text without cursior and outout the text used to emite
func (sd *SyntaxDefinition) EmiteColorText(input string, output string) string {
	var result strings.Builder

    tokens := sd.LineLexer(input)
	remainingText := output

	for len(remainingText) > 0 {
        var color lipgloss.Style
		longestMatch := ""

		for _, token := range tokens {
            output_text, isToken :=  sd.findToken(remainingText, token)
            if (isToken){
                longestMatch = output_text
                color = token.Color
                break
			}
		}


		if longestMatch == "" {
			result.WriteByte(remainingText[0])
			remainingText = remainingText[1:]
            continue
        }
        result.WriteString(sd.colorizeText(longestMatch, color))
        remainingText = remainingText[len(longestMatch):]
	}

	return result.String()
}

func (sd *SyntaxDefinition) findToken(text string, token Token) (string, bool){
	cursorPattern := `█`

	re := regexp.MustCompile(cursorPattern)
	cleanedText := re.ReplaceAllString(text, "")

	if !strings.HasPrefix(cleanedText, token.Value) {
        return "", false
    }

    cursorPos := strings.Index(text, "█")
    if cursorPos == -1 {
        return token.Value, true
    }
    if (cursorPos >= len(token.Value)){
        return token.Value, true
    } 

    tokenWithCursor := token.Value[:cursorPos] + "█" + token.Value[cursorPos:]
    return tokenWithCursor, true
}

func (sd *SyntaxDefinition) colorizeText(text string, style lipgloss.Style) string {
	cursorPattern := `█`

	re := regexp.MustCompile(cursorPattern)

	parts := re.Split(text, -1)

	coloredText := ""
	for i, part := range parts {
		coloredText += style.Render(part)

		if i < len(parts)-1 {
			coloredText += "█"
		}
	}

	return coloredText
}

