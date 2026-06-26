package handlers

import (
	"net/http"
	"roadlog/db"
	"roadlog/models"

	"github.com/gin-gonic/gin"
)

func GetStations(c *gin.Context) {
	fuelType := c.Query("fuelType")
	var stations []models.Station
	if fuelType != "" {
		db.DB.Where("fuel_type = ?", fuelType).Order("name asc").Find(&stations)
	} else {
		db.DB.Order("name asc").Find(&stations)
	}
	c.JSON(http.StatusOK, stations)
}

func CreateStation(c *gin.Context) {
	var input models.Station
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Don't duplicate
	var existing models.Station
	if db.DB.Where("name = ? AND fuel_type = ?", input.Name, input.FuelType).First(&existing).Error == nil {
		c.JSON(http.StatusOK, existing)
		return
	}
	db.DB.Create(&input)
	c.JSON(http.StatusCreated, input)
}

func DeleteStation(c *gin.Context) {
	db.DB.Delete(&models.Station{}, c.Param("id"))
	c.Status(http.StatusNoContent)
}
