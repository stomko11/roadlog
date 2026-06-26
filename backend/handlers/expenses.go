package handlers

import (
	"net/http"
	"roadlog/db"
	"roadlog/models"

	"github.com/gin-gonic/gin"
)

func GetExpenses(c *gin.Context) {
	var expenses []models.Expense
	db.DB.Where("vehicle_id = ?", c.Param("id")).Order("date desc").Find(&expenses)
	c.JSON(http.StatusOK, expenses)
}

func CreateExpense(c *gin.Context) {
	var e models.Expense
	if err := c.ShouldBindJSON(&e); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	e.VehicleID = parseUint(c.Param("id"))
	db.DB.Create(&e)
	c.JSON(http.StatusCreated, e)
}

func UpdateExpense(c *gin.Context) {
	var e models.Expense
	if err := db.DB.First(&e, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var input models.Expense
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Model(&e).Updates(map[string]interface{}{
		"date": input.Date, "amount": input.Amount, "category": input.Category, "notes": input.Notes,
	})
	db.DB.First(&e, e.ID)
	c.JSON(http.StatusOK, e)
}

func DeleteExpense(c *gin.Context) {
	db.DB.Delete(&models.Expense{}, c.Param("id"))
	c.Status(http.StatusNoContent)
}
