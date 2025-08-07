package routes

import (
	"Forum_BE/controllers"
	"Forum_BE/repositories"
	"Forum_BE/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ChatbotRoutes(
	db *gorm.DB,
	authorized *gin.RouterGroup,
) {
	questionRepo := repositories.NewQuestionRepository(db)
	suggestionService := services.NewQuestionSuggestionService(questionRepo)
	controller := controllers.NewQuestionSuggestionController(suggestionService)

	group := authorized.Group("/suggestions")
	{
		group.POST("/chat", controller.SuggestSimilarQuestions)
	}
}
