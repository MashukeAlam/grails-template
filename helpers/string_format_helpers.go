package helpers

import "strings"

const (
	Reset     = "\033[0m"
	Red       = "\033[31m"
	Green     = "\033[32m"
	Yellow    = "\033[33m"
	Blue      = "\033[34m"
	Magenta   = "\033[35m"
	Cyan      = "\033[36m"
	Bold      = "\033[1m"
	Underline = "\033[4m"
	Blink     = "\033[5m"
)

func CapitalizeFirstLetter(str string) string {
	if len(str) == 0 {
		return str
	}
	return strings.ToUpper(string(str[0])) + str[1:]
}

// Converts a snake_case string to CamelCase.
func ToCamelCase(str string) string {
	parts := strings.Split(str, "_")
	for i := range parts {
		parts[i] = CapitalizeFirstLetter(parts[i])
	}
	return strings.Join(parts, "")
}
