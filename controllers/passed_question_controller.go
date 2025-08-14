package controllers

import (
	"Forum_BE/services"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type PassController struct {
	passService services.PassService
}

func NewPassController(s services.PassService) *PassController {
	return &PassController{passService: s}
}

func (pc *PassController) PassQuestion(c *gin.Context) {
	questionIDParam := c.Param("id")
	questionID, err := strconv.ParseUint(questionIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID câu hỏi không hợp lệ"})
		return
	}

	userID := c.GetUint("user_id")

	err = pc.passService.PassQuestion(userID, uint(questionID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể bỏ qua câu hỏi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Bỏ qua câu hỏi thành công"})
}

func (pc *PassController) GetPassedQuestionIDs(c *gin.Context) {
	userID := c.GetUint("user_id")

	ids, err := pc.passService.GetPassedIDs(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lấy danh sách câu hỏi đã bỏ qua"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"passed_ids": ids})
}
