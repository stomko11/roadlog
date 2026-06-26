package handlers

import (
	"encoding/csv"
	"net/http"
	"roadlog/db"
	"roadlog/models"
	"strconv"
	"strings"
	"time"
	"fmt"

	"github.com/gin-gonic/gin"
)

func ParseCSVHeaders(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no file uploaded"})
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	headers, err := reader.Read()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read CSV"})
		return
	}

	// Read up to 5 preview rows
	var preview [][]string
	for i := 0; i < 5; i++ {
		row, err := reader.Read()
		if err != nil {
			break
		}
		preview = append(preview, row)
	}

	c.JSON(http.StatusOK, gin.H{"headers": headers, "preview": preview})
}

type ImportMapping struct {
	VehicleID    uint              `json:"vehicleId"`
	DateFormat   string            `json:"dateFormat"`
	ClearFirst   bool              `json:"clearFirst"`
	Mapping      map[string]int    `json:"mapping"` // field name -> column index
	Data         [][]string        `json:"data"`
}

func ImportCSV(c *gin.Context) {
	var input ImportMapping
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.ClearFirst {
		db.DB.Where("vehicle_id = ?", input.VehicleID).Delete(&models.Fillup{})
	}

	dateFormat := input.DateFormat
	if dateFormat == "" {
		dateFormat = "2006-01-02"
	}

	imported := 0
	var errors []string

	for i, row := range input.Data {
		f := models.Fillup{VehicleID: input.VehicleID}

		// Date (required)
		if idx, ok := input.Mapping["date"]; ok && idx < len(row) {
			t, err := parseDate(row[idx], dateFormat)
			if err != nil {
				errors = append(errors, "row "+strconv.Itoa(i+1)+": invalid date '"+row[idx]+"'")
				continue
			}
			f.Date = t
		} else {
			continue
		}

		// Odometer (required)
		if idx, ok := input.Mapping["odometer"]; ok && idx < len(row) {
			f.Odometer = parseFloat(row[idx])
		}

		// Fuel amount (required)
		if idx, ok := input.Mapping["fuelAmount"]; ok && idx < len(row) {
			f.FuelAmount = parseFloat(row[idx])
			if f.FuelAmount == 0 {
				continue
			}
		} else {
			continue
		}

		// Optional fields
		if idx, ok := input.Mapping["pricePerUnit"]; ok && idx < len(row) {
			f.PricePerUnit = parseFloat(row[idx])
		}
		if idx, ok := input.Mapping["totalCost"]; ok && idx < len(row) {
			f.TotalCost = parseFloat(row[idx])
		}
		if idx, ok := input.Mapping["station"]; ok && idx < len(row) {
			f.Station = strings.TrimSpace(row[idx])
		}
		if idx, ok := input.Mapping["notes"]; ok && idx < len(row) {
			f.Notes = strings.TrimSpace(row[idx])
		}
		if idx, ok := input.Mapping["fullTank"]; ok && idx < len(row) {
			val := strings.ToLower(strings.TrimSpace(row[idx]))
			f.FullTank = val == "1" || val == "true" || val == "yes" || val == "full"
		}

		// Auto-calculate total if missing
		if f.TotalCost == 0 && f.FuelAmount > 0 && f.PricePerUnit > 0 {
			f.TotalCost = f.FuelAmount * f.PricePerUnit
		}

		// Skip duplicates (same vehicle + date + odometer)
		if !input.ClearFirst {
			var existing models.Fillup
			if db.DB.Where("vehicle_id = ? AND date = ? AND odometer = ?", f.VehicleID, f.Date, f.Odometer).First(&existing).Error == nil {
				continue
			}
		}

		db.DB.Create(&f)
		imported++
	}

	c.JSON(http.StatusOK, gin.H{"imported": imported, "errors": errors})
}

func parseFloat(s string) float64 {
	s = strings.TrimSpace(s)
	s = strings.NewReplacer("Lt", "", "lt", "", "L", "", "l", "", "kWh", "", "kwh", "", "kg", "", "Kg", "", "km", "", "Km", "", "KM", "", "mi", "", "€", "", "$", "", "£", "").Replace(s)
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, ",", ".")
	// Remove thousands separators (space or non-breaking space)
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "\u00a0", "")
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

func parseDate(s string, format string) (time.Time, error) {
	s = strings.TrimSpace(s)
	// Try provided format first
	if t, err := time.Parse(format, s); err == nil {
		return t, nil
	}
	// Try common formats
	formats := []string{
		"2006-01-02", "02/01/2006", "01/02/2006", "2006-01-02 15:04",
		"02.01.2006", "2.1.2006", "2006/01/02", "Jan 2, 2006",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unable to parse date: %s", s)
}
