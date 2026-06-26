package handlers

import (
	"net/http"
	"roadlog/db"
	"roadlog/models"

	"github.com/gin-gonic/gin"
)

func GetVehicles(c *gin.Context) {
	var vehicles []models.Vehicle
	db.DB.Order("created_at desc").Find(&vehicles)
	c.JSON(http.StatusOK, vehicles)
}

func CreateVehicle(c *gin.Context) {
	var v models.Vehicle
	if err := c.ShouldBindJSON(&v); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Create(&v)
	c.JSON(http.StatusCreated, v)
}

func GetVehicle(c *gin.Context) {
	var v models.Vehicle
	if err := db.DB.First(&v, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, v)
}

func DeleteVehicle(c *gin.Context) {
	db.DB.Delete(&models.Vehicle{}, c.Param("id"))
	c.Status(http.StatusNoContent)
}

func UpdateVehicle(c *gin.Context) {
	var v models.Vehicle
	if err := db.DB.First(&v, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var input models.Vehicle
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Model(&v).Updates(models.Vehicle{Name: input.Name, Make: input.Make, Model: input.Model, Year: input.Year, Plate: input.Plate, FuelType: input.FuelType, Color: input.Color})
	db.DB.First(&v, v.ID)
	c.JSON(http.StatusOK, v)
}
