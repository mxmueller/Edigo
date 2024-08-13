package file

import (
	"io/ioutil"
	"strings"
)

func Load(filePath string) ([]string, error) {
	fileContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return splitLines(string(fileContent)), nil
}

func splitLines(text string) []string {
	return strings.Split(text, "\n")
}

func Save(filePath string, lines []string) error {
	content := ""
	for _, line := range lines {
		content += line + "\n"
	}
	return ioutil.WriteFile(filePath, []byte(content), 0644)
}
