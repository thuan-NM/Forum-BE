package controllers

import (
	"Forum_BE/responses"
	"Forum_BE/services"
	"github.com/gin-gonic/gin"
	"net/http"
)

type QuestionSuggestionController struct {
	suggestionService services.QuestionSuggestionService
}

func NewQuestionSuggestionController(ss services.QuestionSuggestionService) *QuestionSuggestionController {
	return &QuestionSuggestionController{suggestionService: ss}
}

func (qsc *QuestionSuggestionController) SuggestSimilarQuestions(c *gin.Context) {
	var req struct {
		Text string `json:"text" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	questions, err := qsc.suggestionService.FindSimilarQuestions(req.Text)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var result []responses.QuestionResponse
	for _, q := range questions {
		result = append(result, responses.ToQuestionResponse(&q))
	}

	c.JSON(http.StatusOK, gin.H{
		"related_questions": result,
	})
}
