package handlers

import (
	"net/http"
	"roadlog/db"
	"roadlog/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func ListUsers(c *gin.Context) {
	var users []models.User
	db.DB.Order("created_at asc").Find(&users)
	c.JSON(http.StatusOK, users)
}

func CreateUser(c *gin.Context) {
	var input struct {
		Name     string `json:"name" binding:"required"`
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	user := models.User{Name: input.Name, Email: input.Email, Password: string(hash)}
	if err := db.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
		return
	}
	c.JSON(http.StatusCreated, user)
}

func ChangePassword(c *gin.Context) {
	var input struct {
		UserID      uint   `json:"userId" binding:"required"`
		NewPassword string `json:"newPassword" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	db.DB.Model(&models.User{}).Where("id = ?", input.UserID).Update("password", string(hash))
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteUser(c *gin.Context) {
	db.DB.Delete(&models.User{}, c.Param("id"))
	c.Status(http.StatusNoContent)
}
