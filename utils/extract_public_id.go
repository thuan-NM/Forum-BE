package utils

import (
	"path/filepath"
	"strings"
)

func ExtractPublicID(url string) string {
	// URL Cloudinary có dạng: https://res.cloudinary.com/<cloud_name>/<asset_type>/<delivery_type>/<public_id>.<format>
	parts := strings.Split(url, "/")
	for i, part := range parts {
		if part == "upload" && i+1 < len(parts) {
			return strings.TrimSuffix(parts[i+1], filepath.Ext(parts[i+1]))
		}
	}
	return ""
}
