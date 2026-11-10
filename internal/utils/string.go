package utils

import "unicode"

func MixCase(value string) string {
	runes := []rune(value)
	for index, char := range runes {
		if index%2 == 0 {
			runes[index] = unicode.ToUpper(char)
			continue
		}

		runes[index] = unicode.ToLower(char)
	}

	return string(runes)
}
