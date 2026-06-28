package handlers

import (
	"net/http"
	"roadlog/db"
	"roadlog/models"

	"github.com/gin-gonic/gin"
)

func GetFillups(c *gin.Context) {
	var fillups []models.Fillup
	db.DB.Where("vehicle_id = ?", c.Param("id")).Order("date desc").Find(&fillups)
	c.JSON(http.StatusOK, fillups)
}

func CreateFillup(c *gin.Context) {
	var f models.Fillup
	if err := c.ShouldBindJSON(&f); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	f.VehicleID = parseUint(c.Param("id"))
	if f.TotalCost == 0 && f.FuelAmount > 0 && f.PricePerUnit > 0 {
		f.TotalCost = f.FuelAmount * f.PricePerUnit
	}
	db.DB.Create(&f)
	c.JSON(http.StatusCreated, f)
}

func DeleteFillup(c *gin.Context) {
	db.DB.Delete(&models.Fillup{}, c.Param("id"))
	c.Status(http.StatusNoContent)
}

func UpdateFillup(c *gin.Context) {
	var f models.Fillup
	if err := db.DB.First(&f, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var input models.Fillup
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Model(&f).Updates(map[string]interface{}{
		"date": input.Date, "odometer": input.Odometer, "fuel_amount": input.FuelAmount,
		"price_per_unit": input.PricePerUnit, "total_cost": input.TotalCost,
		"station": input.Station, "full_tank": input.FullTank, "missed_previous": input.MissedPrevious, "notes": input.Notes,
	})
	db.DB.First(&f, f.ID)
	c.JSON(http.StatusOK, f)
}

// GetFillupPrefill returns smart defaults for the fill-up form.
// Respects global settings for which fields to pre-fill.
func GetFillupPrefill(c *gin.Context) {
	vehicleID := c.Param("id")

	// Load settings
	settings := map[string]bool{"prefill_price": true, "prefill_station": true, "prefill_odometer": true}
	var prefs []models.UserPreference
	db.DB.Where("`key` IN ?", []string{"prefill_price", "prefill_station", "prefill_odometer"}).Find(&prefs)
	for _, p := range prefs {
		settings[p.Key] = p.Value != "false"
	}

	prefill := models.FillupPrefill{}

	// Odometer: use the most recent fillup (including EVCC)
	var latest models.Fillup
	if db.DB.Where("vehicle_id = ?", vehicleID).Order("date desc").First(&latest).Error == nil {
		if settings["prefill_odometer"] {
			prefill.Odometer = &latest.Odometer
		}
	}

	// Price and station: skip EVCC entries
	var lastManual models.Fillup
	if db.DB.Where("vehicle_id = ? AND (notes NOT LIKE 'evcc#%' OR notes IS NULL OR notes = '')", vehicleID).Order("date desc").First(&lastManual).Error == nil {
		if settings["prefill_price"] {
			prefill.PricePerUnit = &lastManual.PricePerUnit
		}
		if settings["prefill_station"] && lastManual.Station != "" {
			prefill.Station = &lastManual.Station
		}
		prefill.FullTank = &lastManual.FullTank
	}
	c.JSON(http.StatusOK, prefill)
}

func GetVehicleStats(c *gin.Context) {
	vehicleID := c.Param("id")
	from := c.Query("from")
	to := c.Query("to")

	var stats models.VehicleStats
	var fillups []models.Fillup

	q := db.DB.Where("vehicle_id = ?", vehicleID).Order("odometer asc")
	if from != "" && to != "" {
		q = q.Where("date >= ? AND date <= ?", from+"T00:00:00Z", to+"T23:59:59Z")
	} else if from != "" {
		q = q.Where("date >= ?", from+"T00:00:00Z")
	}
	q.Where("odometer > 0").Find(&fillups)

	stats.TotalFillups = int64(len(fillups))
	for _, f := range fillups {
		stats.TotalSpent += f.TotalCost
	}

	if len(fillups) > 1 {
		stats.TotalDistance = fillups[len(fillups)-1].Odometer - fillups[0].Odometer
		var totalFuel float64
		for i := 1; i < len(fillups); i++ {
			totalFuel += fillups[i].FuelAmount
		}
		if totalFuel > 0 && stats.TotalDistance > 0 {
			stats.AvgConsumption = (totalFuel / stats.TotalDistance) * 100
		}
	}
	c.JSON(http.StatusOK, stats)
}

type ChartPoint struct {
	Date        string  `json:"date"`
	Consumption float64 `json:"consumption"`
	PricePerUnit float64 `json:"pricePerUnit"`
	TotalCost   float64 `json:"totalCost"`
}

func GetVehicleChartData(c *gin.Context) {
	vehicleID := c.Param("id")
	from := c.Query("from")
	to := c.Query("to")

	var fillups []models.Fillup
	q := db.DB.Where("vehicle_id = ?", vehicleID).Order("date asc")
	if from != "" && to != "" {
		q = q.Where("date >= ? AND date <= ?", from+"T00:00:00Z", to+"T23:59:59Z")
	} else if from != "" {
		q = q.Where("date >= ?", from+"T00:00:00Z")
	}
	q.Find(&fillups)

	var points []ChartPoint
	for i, f := range fillups {
		p := ChartPoint{
			Date:         f.Date.Format("2006-01-02"),
			PricePerUnit: f.PricePerUnit,
			TotalCost:    f.TotalCost,
		}
		if i > 0 && f.FullTank && !f.MissedPrevious {
			dist := f.Odometer - fillups[i-1].Odometer
			if dist > 0 {
				p.Consumption = (f.FuelAmount / dist) * 100
			}
		}
		points = append(points, p)
	}
	c.JSON(http.StatusOK, points)
}

func parseUint(s string) uint {
	var n uint
	for _, c := range s {
		n = n*10 + uint(c-'0')
	}
	return n
}
