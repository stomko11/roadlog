package handlers

import (
	"net/http"
	"roadlog/db"
	"roadlog/models"

	"github.com/gin-gonic/gin"
)

func GetRecurring(c *gin.Context) {
	var items []models.RecurringExpense
	db.DB.Where("vehicle_id = ?", c.Param("id")).Order("start_date desc").Find(&items)
	c.JSON(http.StatusOK, items)
}

func CreateRecurring(c *gin.Context) {
	var r models.RecurringExpense
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	r.VehicleID = parseUint(c.Param("id"))
	db.DB.Create(&r)
	AuditLog(c, "create", "recurring", r.ID, r.Category)
	c.JSON(http.StatusCreated, r)
}

func UpdateRecurring(c *gin.Context) {
	var r models.RecurringExpense
	if err := db.DB.First(&r, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var input models.RecurringExpense
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Model(&r).Updates(map[string]interface{}{
		"amount": input.Amount, "category": input.Category, "interval": input.Interval,
		"start_date": input.StartDate, "end_date": input.EndDate, "notes": input.Notes, "active": input.Active,
	})
	db.DB.First(&r, r.ID)
	AuditLog(c, "update", "recurring", r.ID, r.Category)
	c.JSON(http.StatusOK, r)
}

func DeleteRecurring(c *gin.Context) {
	id := parseUint(c.Param("id"))
	db.DB.Delete(&models.RecurringExpense{}, id)
	AuditLog(c, "delete", "recurring", id, "")
	c.Status(http.StatusNoContent)
}
