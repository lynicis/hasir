package sanitize

import (
	"html"
	"strings"
)

func StripHTML(s string) string {
	var b strings.Builder
	inTag := false
	for _, r := range s {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			continue
		}
		if !inTag {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func EscapeHTML(s string) string {
	return html.EscapeString(s)
}
