package handlers

import (
	"net/http"
	"roadlog/db"
	"roadlog/models"
	"time"

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
	if e.Recurring && e.NotifyDaysBefore == 0 {
		e.NotifyDaysBefore = 7
	}
	db.DB.Create(&e)
	AuditLog(c, "create", "expense", e.ID, e.Category)
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
		"interval": input.Interval, "next_due": input.NextDue, "notify_days_before": input.NotifyDaysBefore,
	})
	db.DB.First(&e, e.ID)
	AuditLog(c, "update", "expense", e.ID, e.Category)
	c.JSON(http.StatusOK, e)
}

func DeleteExpense(c *gin.Context) {
	id := parseUint(c.Param("id"))
	db.DB.Delete(&models.Expense{}, id)
	AuditLog(c, "delete", "expense", id, "")
	c.Status(http.StatusNoContent)
}

func ConfirmRecurringExpense(c *gin.Context) {
	var tmpl models.Expense
	if err := db.DB.First(&tmpl, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var input struct {
		Amount *float64 `json:"amount"`
	}
	c.ShouldBindJSON(&input)
	amount := tmpl.Amount
	if input.Amount != nil {
		amount = *input.Amount
	}
	expense := models.Expense{VehicleID: tmpl.VehicleID, Date: time.Now(), Amount: amount, Category: tmpl.Category, Notes: tmpl.Notes, Recurring: false}
	db.DB.Create(&expense)
	switch tmpl.Interval {
	case "monthly":
		next := tmpl.NextDue.AddDate(0, 1, 0)
		tmpl.NextDue = &next
	case "quarterly":
		next := tmpl.NextDue.AddDate(0, 3, 0)
		tmpl.NextDue = &next
	case "yearly":
		next := tmpl.NextDue.AddDate(1, 0, 0)
		tmpl.NextDue = &next
	}
	db.DB.Model(&tmpl).Updates(map[string]interface{}{"next_due": tmpl.NextDue, "notified_at": nil})
	AuditLog(c, "confirm", "expense", tmpl.ID, tmpl.Category)
	c.JSON(http.StatusOK, tmpl)
}

func EndRecurringExpense(c *gin.Context) {
	var e models.Expense
	if err := db.DB.First(&e, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	active := false
	db.DB.Model(&e).Update("recurring_active", &active)
	AuditLog(c, "end", "expense", e.ID, e.Category)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
