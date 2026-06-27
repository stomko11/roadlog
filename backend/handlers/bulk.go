package handlers

import (
	"net/http"
	"roadlog/db"
	"roadlog/models"

	"github.com/gin-gonic/gin"
)

func BulkDeleteFillups(c *gin.Context) {
	var input struct{ IDs []uint `json:"ids"` }
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Where("id IN ?", input.IDs).Delete(&models.Fillup{})
	c.JSON(http.StatusOK, gin.H{"deleted": len(input.IDs)})
}

func BulkDeleteExpenses(c *gin.Context) {
	var input struct{ IDs []uint `json:"ids"` }
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Where("id IN ?", input.IDs).Delete(&models.Expense{})
	c.JSON(http.StatusOK, gin.H{"deleted": len(input.IDs)})
}
