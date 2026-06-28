package handlers

import (
	"net/http"
	"roadlog/db"
	"roadlog/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

func AuditLog(c *gin.Context, action, entityType string, entityID uint, details string) {
	userID, _ := c.Get("userID")
	uid, _ := userID.(uint)
	entry := models.AuditEntry{
		UserID: uid, Action: action, EntityType: entityType,
		EntityID: entityID, Details: details, IP: c.ClientIP(),
	}
	db.DB.Create(&entry)
}

func GetAuditLog(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	var entries []models.AuditEntry
	var total int64
	db.DB.Model(&models.AuditEntry{}).Count(&total)
	db.DB.Order("created_at desc").Limit(limit).Offset(offset).Find(&entries)

	// Enrich with user names and vehicle names
	var users []models.User
	db.DB.Find(&users)
	userMap := map[uint]string{}
	for _, u := range users {
		userMap[u.ID] = u.Name
	}
	var vehicles []models.Vehicle
	db.DB.Find(&vehicles)
	vehicleMap := map[uint]string{}
	for _, v := range vehicles {
		vehicleMap[v.ID] = v.Name
	}

	type EnrichedEntry struct {
		models.AuditEntry
		UserName    string `json:"userName"`
		VehicleName string `json:"vehicleName"`
	}
	enriched := make([]EnrichedEntry, len(entries))
	for i, e := range entries {
		enriched[i] = EnrichedEntry{AuditEntry: e, UserName: userMap[e.UserID]}
		// Look up vehicle from the entity
		if e.EntityType == "fillup" || e.EntityType == "expense" || e.EntityType == "reminder" {
			var vehicleID uint
			switch e.EntityType {
			case "fillup":
				var f models.Fillup
				if db.DB.First(&f, e.EntityID).Error == nil {
					vehicleID = f.VehicleID
				}
			case "expense":
				var ex models.Expense
				if db.DB.First(&ex, e.EntityID).Error == nil {
					vehicleID = ex.VehicleID
				}
			case "reminder":
				var r models.Reminder
				if db.DB.First(&r, e.EntityID).Error == nil {
					vehicleID = r.VehicleID
				}
			}
			enriched[i].VehicleName = vehicleMap[vehicleID]
		} else if e.EntityType == "vehicle" {
			enriched[i].VehicleName = vehicleMap[e.EntityID]
		}
	}

	c.JSON(http.StatusOK, gin.H{"entries": enriched, "total": total})
}
