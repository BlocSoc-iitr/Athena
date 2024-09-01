package athena

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"unicode"
)

// randomAddress generates a random 20-byte ChecksumAddress in Go.
func randomAddress() (string, error) {
	bytes := make([]byte, 20)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "0x" + hex.EncodeToString(bytes), nil
}

// uintOverUnderFlow handles uint overflow and underflow based on the specified precision.
func uintOverUnderFlow(value int64, precision int) int64 {
	mod := int64(1 << precision)
	if value < 0 {
		return value + mod
	}
	if value >= mod {
		return value - mod
	}
	return value
}

// camelToSnake converts a CamelCase string to snake_case.
func camelToSnake(name string) string {
	var outString strings.Builder
	lastChar := rune(name[0])

	for _, char := range name[1:] {
		if unicode.IsUpper(char) {
			if unicode.IsLower(lastChar) || unicode.IsDigit(lastChar) {
				outString.WriteRune(lastChar)
				outString.WriteRune('_')
			} else {
				outString.WriteRune(unicode.ToLower(lastChar))
			}
		} else if unicode.IsDigit(char) {
			if unicode.IsLetter(lastChar) {
				outString.WriteRune(unicode.ToLower(lastChar))
				outString.WriteRune('_')
			} else {
				outString.WriteRune(unicode.ToLower(lastChar))
			}
		} else {
			outString.WriteRune(unicode.ToLower(lastChar))
		}
		lastChar = char
	}
	return outString.String()
}

// pprintList formats a list of strings to fit within a terminal width, wrapping lines as needed.
func pprintList(writeArray []string, termWidth int) []string {
	currentLine := ""
	output := []string{}

	for _, writeVal := range writeArray {
		if len(currentLine)+len(writeVal)+1 > termWidth {
			output = append(output, currentLine)
			currentLine = ""
		}
		currentLine += "'" + writeVal + "', "
	}

	output = append(output, currentLine)

	return output
}
