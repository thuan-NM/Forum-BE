package utils

import (
	"fmt"
)

func GenerateCacheKey(prefix string, id uint, filters map[string]interface{}) string {
	key := fmt.Sprintf("%s:%d:", prefix, id)
	if page, ok := filters["page"]; ok {
		key += fmt.Sprintf("page:%v:", page)
	}
	if limit, ok := filters["limit"]; ok {
		key += fmt.Sprintf("limit:%v:", limit)
	}
	if status, ok := filters["status"]; ok {
		key += fmt.Sprintf("status:%v:", status)
	}
	if search, ok := filters["title_search"]; ok {
		key += fmt.Sprintf("search:%v:", search)
	}
	if userID, ok := filters["user_id"]; ok {
		key += fmt.Sprintf("user:%v:", userID)
	}
	if tagID, ok := filters["tag_id"]; ok {
		key += fmt.Sprintf("tag:%v:", tagID)
	}
	if sort, ok := filters["sort"]; ok {
		key += fmt.Sprintf("sort:%v:", sort)
	}
	if questiontitle, ok := filters["questiontitle"]; ok {
		key += fmt.Sprintf("questiontitle:%v:", questiontitle)
	}
	if search, ok := filters["search"]; ok {
		key += fmt.Sprintf("search:%v:", search)
	}
	if typefilter, ok := filters["typefilter"]; ok {
		key += fmt.Sprintf("typefilter:%v:", typefilter)

	}
	return key
}
