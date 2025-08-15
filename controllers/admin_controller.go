package controllers

import (
	"Forum_BE/services"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type AdminController struct {
	AdminService services.AdminService
}

func NewAdminController(aService services.AdminService) *AdminController {
	return &AdminController{AdminService: aService}
}

// ApproveQuestionHandler phê duyệt câu hỏi
func (c *AdminController) ApproveQuestionHandler(ctx *gin.Context) {
	idParam := ctx.Param("id")
	questionID, err := strconv.Atoi(idParam)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ID câu hỏi không hợp lệ"})
		return
	}

	err = c.AdminService.ApproveQuestion(uint(questionID))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Phê duyệt câu hỏi thành công"})
}

// RejectQuestionHandler từ chối câu hỏi
func (c *AdminController) RejectQuestionHandler(ctx *gin.Context) {
	idParam := ctx.Param("id")
	questionID, err := strconv.Atoi(idParam)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ID câu hỏi không hợp lệ"})
		return
	}

	err = c.AdminService.RejectQuestion(uint(questionID))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Từ chối câu hỏi thành công"})
}
