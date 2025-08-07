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
	if interstatus, ok := filters["interstatus"]; ok {
		key += fmt.Sprintf("interstatus:%v:", interstatus)
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
	if answerID, ok := filters["answer_id"]; ok {
		key += fmt.Sprintf("answer:%v:", answerID)
	}
	if postID, ok := filters["post_id"]; ok {
		key += fmt.Sprintf("post:%v:", postID)
	}
	if topicID, ok := filters["topic_id"]; ok {
		key += fmt.Sprintf("topic_id:%v:", topicID)
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
	if tagfilter, ok := filters["tagfilter"]; ok {
		key += fmt.Sprintf("tagfilter:%v:", tagfilter)
	}
	if typefilter, ok := filters["typefilter"]; ok {
		key += fmt.Sprintf("typefilter:%v:", typefilter)
	}
	if file_type, ok := filters["file_type"]; ok {
		key += fmt.Sprintf("file_type:%v:", file_type)
	}
	if author, ok := filters["author"]; ok {
		key += fmt.Sprintf("author:%v:", author)
	}
	return key
}
