package utils

import "strings"

func DetermineFileType(contentType string) string {
	if strings.HasPrefix(contentType, "image/") {
		return "image"
	} else if strings.HasPrefix(contentType, "video/") {
		return "video"
	} else if strings.HasPrefix(contentType, "audio/") {
		return "audio"
	} else if contentType == "application/pdf" || strings.HasPrefix(contentType, "application/") {
		return "document"
	}
	return "other"
}
