package util

import (
	"strings"
	"unicode"
)

// HolderHintFromName делает инициализацию вида "И.И." из полного имени/псевдонима
func HolderHintFromName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	// split by spaces, take first letters
	parts := strings.Fields(name)
	if len(parts) == 0 {
		return ""
	}
	// take up to 2 initials
	initials := make([]rune, 0, 2)
	for _, p := range parts {
		for _, r := range p {
			if unicode.IsLetter(r) {
				initials = append(initials, unicode.ToUpper(r))
				break
			}
		}
		if len(initials) == 2 {
			break
		}
	}
	if len(initials) == 0 {
		return ""
	}
	if len(initials) == 1 {
		return string(initials[0]) + "."
	}
	return string(initials[0]) + "." + string(initials[1]) + "."
}
