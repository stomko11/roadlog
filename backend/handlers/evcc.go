package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"roadlog/db"
	"roadlog/models"
	"time"

	"github.com/gin-gonic/gin"
)

type evccState struct {
	Loadpoints []struct {
		Title        string `json:"title"`
		Name         string `json:"name"`
		VehicleName  string `json:"vehicleName"`
		VehicleTitle string `json:"vehicleTitle"`
	} `json:"loadpoints"`
	Vehicles map[string]struct {
		Title string `json:"title"`
	} `json:"vehicles"`
}

type evccSession struct {
	ID               int      `json:"id"`
	Created          string   `json:"created"`
	Finished         string   `json:"finished"`
	Loadpoint        string   `json:"loadpoint"`
	Vehicle          string   `json:"vehicle"`
	Odometer         float64  `json:"odometer"`
	ChargedEnergy    float64  `json:"chargedEnergy"`
	SolarPercentage  float64  `json:"solarPercentage"`
	Price            *float64 `json:"price"`
	PricePerKWh      *float64 `json:"pricePerKWh"`
	ReferencePricePerKWh *float64 `json:"referencePricePerKWh"`
}

// GET /api/evcc/discover?url=...
func EVCCDiscover(c *gin.Context) {
	url := c.Query("url")
	if url == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "url required"})
		return
	}
	resp, err := http.Get(url + "/api/state")
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "cannot reach EVCC: " + err.Error()})
		return
	}
	defer resp.Body.Close()
	var state evccState
	if err := json.NewDecoder(resp.Body).Decode(&state); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "invalid response"})
		return
	}
	type LP struct {
		Name         string `json:"name"`
		Title        string `json:"title"`
		VehicleName  string `json:"vehicleName"`
		VehicleTitle string `json:"vehicleTitle"`
	}
	type V struct {
		Name  string `json:"name"`
		Title string `json:"title"`
	}
	lps := []LP{}
	for _, lp := range state.Loadpoints {
		lps = append(lps, LP{Name: lp.Name, Title: lp.Title, VehicleName: lp.VehicleName, VehicleTitle: lp.VehicleTitle})
	}
	vehicles := []V{}
	for name, v := range state.Vehicles {
		vehicles = append(vehicles, V{Name: name, Title: v.Title})
	}
	c.JSON(http.StatusOK, gin.H{"loadpoints": lps, "vehicles": vehicles})
}

// GET /api/vehicles/:id/evcc
func GetVehicleEVCCSources(c *gin.Context) {
	var sources []models.VehicleEVCC
	db.DB.Where("vehicle_id = ?", c.Param("id")).Find(&sources)
	c.JSON(http.StatusOK, sources)
}

// POST /api/vehicles/:id/evcc
func CreateVehicleEVCCSource(c *gin.Context) {
	var input models.VehicleEVCC
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.VehicleID = parseUint(c.Param("id"))
	db.DB.Create(&input)
	c.JSON(http.StatusCreated, input)
}

// DELETE /api/evcc/:id
func DeleteVehicleEVCCSource(c *gin.Context) {
	db.DB.Delete(&models.VehicleEVCC{}, c.Param("id"))
	c.Status(http.StatusNoContent)
}

// PUT /api/evcc/:id
func UpdateVehicleEVCCSource(c *gin.Context) {
	var src models.VehicleEVCC
	if err := db.DB.First(&src, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var input models.VehicleEVCC
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Model(&src).Updates(map[string]interface{}{
		"url": input.URL, "evcc_vehicle": input.EVCCVehicle, "evcc_loadpoint": input.EVCCLoadpoint,
		"label": input.Label, "sync_since": input.SyncSince, "auto_sync": input.AutoSync, "sync_time": input.SyncTime,
		"fallback_price": input.FallbackPrice,
	})
	db.DB.First(&src, src.ID)
	c.JSON(http.StatusOK, src)
}

// POST /api/vehicles/:id/evcc/sync
func SyncVehicleEVCC(c *gin.Context) {
	vehicleID := parseUint(c.Param("id"))
	var sources []models.VehicleEVCC
	db.DB.Where("vehicle_id = ?", vehicleID).Find(&sources)

	totalImported := 0
	for i := range sources {
		n, err := syncSource(&sources[i], vehicleID)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("%s: %v", sources[i].Label, err)})
			return
		}
		totalImported += n
	}
	c.JSON(http.StatusOK, gin.H{"imported": totalImported})
}

func syncSource(src *models.VehicleEVCC, vehicleID uint) (int, error) {
	resp, err := http.Get(src.URL + "/api/sessions")
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	var sessions []evccSession
	if err := json.NewDecoder(resp.Body).Decode(&sessions); err != nil {
		return 0, err
	}

	var syncSince time.Time
	if src.SyncSince != "" {
		syncSince, _ = time.Parse("2006-01-02", src.SyncSince)
	}

	imported := 0
	maxID := src.LastSyncedSessionID
	for _, s := range sessions {
		// Skip already synced
		if s.ID <= src.LastSyncedSessionID {
			continue
		}
		// Filter by vehicle
		if src.EVCCVehicle != "" && s.Vehicle != src.EVCCVehicle {
			continue
		}
		// Filter by loadpoint
		if src.EVCCLoadpoint != "" && s.Loadpoint != src.EVCCLoadpoint {
			continue
		}
		// Parse date and filter by syncSince
		created, err := time.Parse(time.RFC3339Nano, s.Created)
		if err != nil {
			continue
		}
		if !syncSince.IsZero() && created.Before(syncSince) {
			continue
		}
		// Skip sessions without price data (unless fallback configured)
		if s.Price == nil && src.FallbackPrice <= 0 {
			continue
		}
		// Dedup: check if fillup with same date+vehicleId+notes containing session ID exists
		evccNote := fmt.Sprintf("evcc#%d", s.ID)
		var existing models.Fillup
		if db.DB.Where("vehicle_id = ? AND notes LIKE ?", vehicleID, "%"+evccNote+"%").First(&existing).Error == nil {
			continue
		}

		station := src.Label
		if s.SolarPercentage > 0 {
			station = fmt.Sprintf("%s (☀️ %.0f%%)", src.Label, s.SolarPercentage)
		}

		var pricePerUnit float64
		var totalCost float64
		estimated := false
		if s.Price != nil {
			totalCost = *s.Price
			if s.PricePerKWh != nil {
				pricePerUnit = *s.PricePerKWh
			}
		} else {
			// Estimate using fallback price and solar percentage
			gridFraction := 1.0 - s.SolarPercentage/100.0
			totalCost = s.ChargedEnergy * gridFraction * src.FallbackPrice
			pricePerUnit = gridFraction * src.FallbackPrice
			estimated = true
		}

		var saved string
		if s.ReferencePricePerKWh != nil && s.ChargedEnergy > 0 {
			gridCost := *s.ReferencePricePerKWh * s.ChargedEnergy
			saved = fmt.Sprintf(" | Saved €%.2f vs grid", gridCost-totalCost)
		}
		if estimated {
			saved += " | ~estimated"
		}

		fillup := models.Fillup{
			VehicleID:    vehicleID,
			Date:         created,
			Odometer:     s.Odometer,
			FuelAmount:   s.ChargedEnergy,
			PricePerUnit: pricePerUnit,
			TotalCost:    totalCost,
			FullTank:     false,
			Station:      station,
			Notes:        fmt.Sprintf("%s%s", evccNote, saved),
		}
		db.DB.Create(&fillup)
		imported++

		if s.ID > maxID {
			maxID = s.ID
		}
	}

	if maxID > src.LastSyncedSessionID {
		src.LastSyncedSessionID = maxID
		db.DB.Model(src).Update("last_synced_session_id", maxID)
	}
	return imported, nil
}

// RunEVCCScheduler starts a goroutine that checks every minute for sources needing sync
func RunEVCCScheduler() {
	go func() {
		for {
			time.Sleep(60 * time.Second)
			now := time.Now()
			nowTime := now.Format("15:04")

			var sources []models.VehicleEVCC
			db.DB.Where("auto_sync = ? AND sync_time = ?", true, nowTime).Find(&sources)
			for i := range sources {
				syncSource(&sources[i], sources[i].VehicleID)
			}
		}
	}()
}
