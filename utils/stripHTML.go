package utils

import (
	"regexp"
	"strings"
)

func StripHTML(content string) string {
	re := regexp.MustCompile(`<[^>]+>`)
	clean := re.ReplaceAllString(content, "")
	clean = strings.TrimSpace(strings.Join(strings.Fields(clean), " "))
	return clean
}
