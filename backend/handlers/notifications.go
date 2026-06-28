package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"roadlog/db"
	"roadlog/models"
	"time"

	"github.com/gin-gonic/gin"
)

func GetNotifications(c *gin.Context) {
	var items []models.NotificationConfig
	db.DB.Find(&items)
	c.JSON(http.StatusOK, items)
}

func CreateNotification(c *gin.Context) {
	var n models.NotificationConfig
	if err := c.ShouldBindJSON(&n); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Create(&n)
	c.JSON(http.StatusCreated, n)
}

func UpdateNotification(c *gin.Context) {
	var n models.NotificationConfig
	if err := db.DB.First(&n, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var input models.NotificationConfig
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Model(&n).Updates(map[string]interface{}{"type": input.Type, "enabled": input.Enabled, "config": input.Config})
	db.DB.First(&n, n.ID)
	c.JSON(http.StatusOK, n)
}

func DeleteNotification(c *gin.Context) {
	db.DB.Delete(&models.NotificationConfig{}, c.Param("id"))
	c.Status(http.StatusNoContent)
}

func SendNotification(title, message string) {
	var configs []models.NotificationConfig
	db.DB.Where("enabled = ?", true).Find(&configs)
	for _, cfg := range configs {
		switch cfg.Type {
		case "pushover":
			sendPushover(cfg.Config, title, message)
		case "webhook":
			sendWebhook(cfg.Config, title, message)
		}
	}
}

// POST /api/notifications/test
func TestNotification(c *gin.Context) {
	var input struct {
		Type   string `json:"type"`
		Config string `json:"config"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	switch input.Type {
	case "pushover":
		sendPushover(input.Config, "🚗 Roadlog Test", "Notifications are working!")
	case "webhook":
		sendWebhook(input.Config, "🚗 Roadlog Test", "Notifications are working!")
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func sendPushover(cfgJSON, title, message string) {
	var cfg struct {
		Token string `json:"token"`
		User  string `json:"user"`
	}
	json.Unmarshal([]byte(cfgJSON), &cfg)
	if cfg.Token == "" || cfg.User == "" {
		return
	}
	body, _ := json.Marshal(map[string]string{"token": cfg.Token, "user": cfg.User, "title": title, "message": message})
	http.Post("https://api.pushover.net/1/messages.json", "application/json", bytes.NewReader(body))
}

func sendWebhook(cfgJSON, title, message string) {
	var cfg struct {
		URL string `json:"url"`
	}
	json.Unmarshal([]byte(cfgJSON), &cfg)
	if cfg.URL == "" {
		return
	}
	body, _ := json.Marshal(map[string]string{"title": title, "message": message})
	http.Post(cfg.URL, "application/json", bytes.NewReader(body))
}

// RunReminderScheduler checks every 15 minutes for reminders that are due soon.
// Also runs immediately on startup.
func RunReminderScheduler() {
	go func() {
		checkReminders() // run immediately on startup
		for {
			time.Sleep(15 * time.Minute)
			checkReminders()
		}
	}()
}

func getTimezone() *time.Location {
	var pref models.UserPreference
	if db.DB.Where("`key` = ?", "timezone").First(&pref).Error == nil && pref.Value != "" {
		if loc, err := time.LoadLocation(pref.Value); err == nil {
			return loc
		}
	}
	return time.Local
}

func checkReminders() {
	loc := getTimezone()
	now := time.Now().In(loc)

	var reminders []models.Reminder
	db.DB.Where("done = ? AND notified_at IS NULL", false).Find(&reminders)

	for _, r := range reminders {
		notify := false
		msg := ""
		days := r.NotifyDaysBefore
		if days == 0 {
			days = 7
		}
		soon := now.AddDate(0, 0, days)

		if r.DueDate != nil && r.DueDate.Before(soon) {
			notify = true
			msg = fmt.Sprintf("Due: %s", r.DueDate.Format("2006-01-02"))
		}

		if r.DueOdometer != nil {
			var lastFillup models.Fillup
			if db.DB.Where("vehicle_id = ?", r.VehicleID).Order("odometer desc").First(&lastFillup).Error == nil {
				if *r.DueOdometer-lastFillup.Odometer <= 500 {
					notify = true
					msg = fmt.Sprintf("Due at %.0f km (current: %.0f)", *r.DueOdometer, lastFillup.Odometer)
				}
			}
		}

		if notify {
			var v models.Vehicle
			db.DB.First(&v, r.VehicleID)
			title := fmt.Sprintf("🔔 %s - %s", v.Name, r.Title)
			if msg == "" {
				msg = r.Notes
			}
			SendNotification(title, msg)
			db.DB.Model(&r).Update("notified_at", now)
		}
	}

	// Check recurring expenses
	var recurringExpenses []models.Expense
	db.DB.Where("recurring = ? AND recurring_active = ? AND notified_at IS NULL", true, true).Find(&recurringExpenses)
	for _, r := range recurringExpenses {
		days := r.NotifyDaysBefore
		if days == 0 {
			days = 7
		}
		threshold := now.AddDate(0, 0, days)
		if r.NextDue != nil && r.NextDue.Before(threshold) {
			var v models.Vehicle
			db.DB.First(&v, r.VehicleID)
			title := fmt.Sprintf("💰 %s - %s", v.Name, r.Category)
			msg := fmt.Sprintf("Due: %s · %s", r.NextDue.Format("2006-01-02"), r.Notes)
			SendNotification(title, msg)
			db.DB.Model(&r).Update("notified_at", now)
		}
	}
}
