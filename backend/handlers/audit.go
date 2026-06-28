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
	c.JSON(http.StatusOK, gin.H{"entries": entries, "total": total})
}
