package handlers

import (
	"net/http"
	"regexp"
	"roadlog/db"
	"roadlog/models"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

var solarRe = regexp.MustCompile(`☀️\s*(\d+)%`)
var savedRe = regexp.MustCompile(`Saved €([\d.]+) vs grid`)

type EVStats struct {
	TotalEnergy     float64         `json:"totalEnergy"`
	TotalCost       float64         `json:"totalCost"`
	TotalSaved      float64         `json:"totalSaved"`
	SolarEnergy     float64         `json:"solarEnergy"`
	GridEnergy      float64         `json:"gridEnergy"`
	AvgSolarPercent float64         `json:"avgSolarPercent"`
	CostPerKm      float64         `json:"costPerKm"`
	TotalDistance   float64         `json:"totalDistance"`
	Sessions        int             `json:"sessions"`
	Monthly         []EVMonthStats  `json:"monthly"`
}

type EVMonthStats struct {
	Month       string  `json:"month"`
	Energy      float64 `json:"energy"`
	Cost        float64 `json:"cost"`
	Saved       float64 `json:"saved"`
	SolarPct    float64 `json:"solarPct"`
}

// GET /api/vehicles/:id/ev-stats
func GetEVStats(c *gin.Context) {
	vehicleID := c.Param("id")

	var fillups []models.Fillup
	db.DB.Where("vehicle_id = ? AND notes LIKE 'evcc#%'", vehicleID).Order("date asc").Find(&fillups)

	var allFillups []models.Fillup
	db.DB.Where("vehicle_id = ? AND odometer > 0", vehicleID).Order("odometer asc").Find(&allFillups)

	stats := EVStats{}
	monthly := map[string]*EVMonthStats{}

	for _, f := range fillups {
		stats.Sessions++
		stats.TotalEnergy += f.FuelAmount
		stats.TotalCost += f.TotalCost

		// Parse solar % from station
		solarPct := 0.0
		if m := solarRe.FindStringSubmatch(f.Station); len(m) > 1 {
			solarPct, _ = strconv.ParseFloat(m[1], 64)
		}
		stats.SolarEnergy += f.FuelAmount * solarPct / 100
		stats.GridEnergy += f.FuelAmount * (1 - solarPct/100)

		// Parse savings from notes
		if m := savedRe.FindStringSubmatch(f.Notes); len(m) > 1 {
			saved, _ := strconv.ParseFloat(m[1], 64)
			stats.TotalSaved += saved
		}

		// Monthly aggregation
		key := f.Date.Format("2006-01")
		if _, ok := monthly[key]; !ok {
			monthly[key] = &EVMonthStats{Month: key}
		}
		monthly[key].Energy += f.FuelAmount
		monthly[key].Cost += f.TotalCost
		if m := savedRe.FindStringSubmatch(f.Notes); len(m) > 1 {
			saved, _ := strconv.ParseFloat(m[1], 64)
			monthly[key].Saved += saved
		}
	}

	// Calculate avg solar %
	if stats.TotalEnergy > 0 {
		stats.AvgSolarPercent = (stats.SolarEnergy / stats.TotalEnergy) * 100
	}

	// Monthly solar %
	for _, f := range fillups {
		key := f.Date.Format("2006-01")
		solarPct := 0.0
		if m := solarRe.FindStringSubmatch(f.Station); len(m) > 1 {
			solarPct, _ = strconv.ParseFloat(m[1], 64)
		}
		if monthly[key] != nil && monthly[key].Energy > 0 {
			monthly[key].SolarPct += f.FuelAmount * solarPct / monthly[key].Energy
		}
	}

	// Total distance (from all fillups, not just EVCC)
	if len(allFillups) > 1 {
		stats.TotalDistance = allFillups[len(allFillups)-1].Odometer - allFillups[0].Odometer
	}

	// Cost per km (all charging costs / total distance)
	var totalAllCost float64
	for _, f := range allFillups {
		if strings.HasPrefix(f.Notes, "evcc#") || f.TotalCost > 0 {
			totalAllCost += f.TotalCost
		}
	}
	if stats.TotalDistance > 0 {
		stats.CostPerKm = totalAllCost / stats.TotalDistance
	}

	// Sort monthly
	for _, m := range monthly {
		stats.Monthly = append(stats.Monthly, *m)
	}
	// Sort ascending
	for i := 0; i < len(stats.Monthly); i++ {
		for j := i + 1; j < len(stats.Monthly); j++ {
			if stats.Monthly[j].Month < stats.Monthly[i].Month {
				stats.Monthly[i], stats.Monthly[j] = stats.Monthly[j], stats.Monthly[i]
			}
		}
	}

	c.JSON(http.StatusOK, stats)
}
