package handlers

import (
	"net/http"
	"roadlog/db"
	"roadlog/models"
	"time"

	"github.com/gin-gonic/gin"
)

func GetReminders(c *gin.Context) {
	var items []models.Reminder
	db.DB.Where("vehicle_id = ?", c.Param("id")).Order("done asc, due_date asc").Find(&items)
	c.JSON(http.StatusOK, items)
}

func CreateReminder(c *gin.Context) {
	var r models.Reminder
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	r.VehicleID = parseUint(c.Param("id"))
	db.DB.Create(&r)
	AuditLog(c, "create", "reminder", r.ID, r.Title)
	c.JSON(http.StatusCreated, r)
}

func UpdateReminder(c *gin.Context) {
	var r models.Reminder
	if err := db.DB.First(&r, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var input models.Reminder
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Model(&r).Updates(map[string]interface{}{
		"title": input.Title, "due_date": input.DueDate, "due_odometer": input.DueOdometer,
		"repeat_months": input.RepeatMonths, "repeat_km": input.RepeatKm, "notes": input.Notes,
	})
	db.DB.First(&r, r.ID)
	AuditLog(c, "update", "reminder", r.ID, r.Title)
	c.JSON(http.StatusOK, r)
}

func DeleteReminder(c *gin.Context) {
	id := parseUint(c.Param("id"))
	db.DB.Delete(&models.Reminder{}, id)
	AuditLog(c, "delete", "reminder", id, "")
	c.Status(http.StatusNoContent)
}

func MarkReminderDone(c *gin.Context) {
	var r models.Reminder
	if err := db.DB.First(&r, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	now := time.Now()
	db.DB.Model(&r).Updates(map[string]interface{}{"done": true, "done_date": now})

	// Auto-create next occurrence if repeat values set
	if r.RepeatMonths > 0 || r.RepeatKm > 0 {
		next := models.Reminder{
			VehicleID: r.VehicleID, Title: r.Title, Notes: r.Notes,
			RepeatMonths: r.RepeatMonths, RepeatKm: r.RepeatKm,
		}
		if r.RepeatMonths > 0 && r.DueDate != nil {
			nd := r.DueDate.AddDate(0, r.RepeatMonths, 0)
			next.DueDate = &nd
		}
		if r.RepeatKm > 0 && r.DueOdometer != nil {
			no := *r.DueOdometer + r.RepeatKm
			next.DueOdometer = &no
		}
		db.DB.Create(&next)
	}
	AuditLog(c, "done", "reminder", r.ID, r.Title)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
