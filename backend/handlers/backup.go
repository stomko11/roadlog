package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"roadlog/db"
	"roadlog/models"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type BackupData struct {
	ExportedAt string                 `json:"exportedAt"`
	Users      []models.User          `json:"users"`
	Vehicles   []models.Vehicle       `json:"vehicles"`
	Fillups    []models.Fillup        `json:"fillups"`
	Expenses   []models.Expense       `json:"expenses"`
	Settings   []models.UserPreference `json:"settings"`
}

func Backup(c *gin.Context) {
	var data BackupData
	data.ExportedAt = time.Now().Format(time.RFC3339)
	db.DB.Find(&data.Users)
	db.DB.Find(&data.Vehicles)
	db.DB.Find(&data.Fillups)
	db.DB.Find(&data.Expenses)
	db.DB.Find(&data.Settings)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=roadlog-backup-%s.json", time.Now().Format("2006-01-02")))
	c.JSON(http.StatusOK, data)
}

func Restore(c *gin.Context) {
	var data BackupData
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get DB path and close
	path := "/data/roadlog.db"
	if p := os.Getenv("DATA_DIR"); p != "" && p != "/" {
		path = p + "/roadlog.db"
	}

	// Clear and reimport
	db.DB.Exec("DELETE FROM fillups")
	db.DB.Exec("DELETE FROM expenses")
	db.DB.Exec("DELETE FROM vehicles")
	db.DB.Exec("DELETE FROM user_preferences")
	db.DB.Exec("DELETE FROM users")

	for _, u := range data.Users {
		db.DB.Create(&u)
	}
	for _, v := range data.Vehicles {
		db.DB.Create(&v)
	}
	for _, f := range data.Fillups {
		db.DB.Create(&f)
	}
	for _, e := range data.Expenses {
		db.DB.Create(&e)
	}
	for _, s := range data.Settings {
		db.DB.Create(&s)
	}

	_ = path // used for potential file-level backup in future
	c.JSON(http.StatusOK, gin.H{"restored": true, "users": len(data.Users), "vehicles": len(data.Vehicles), "fillups": len(data.Fillups), "expenses": len(data.Expenses)})
}

func ExportVehicleCSV(c *gin.Context) {
	vehicleID := c.Param("id")
	var vehicle models.Vehicle
	if err := db.DB.First(&vehicle, vehicleID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var fillups []models.Fillup
	db.DB.Where("vehicle_id = ?", vehicleID).Order("date asc").Find(&fillups)

	var sb strings.Builder
	sb.WriteString("Date,Odometer (km),Fuel Amount (L),Price per Unit,Total Cost,Station,Full Tank,Notes\n")
	for _, f := range fillups {
		sb.WriteString(fmt.Sprintf("%s,%.0f,%.2f,%.3f,%.2f,%s,%t,%s\n",
			f.Date.Format("2006-01-02"), f.Odometer, f.FuelAmount, f.PricePerUnit, f.TotalCost,
			csvEscape(f.Station), f.FullTank, csvEscape(f.Notes)))
	}

	filename := fmt.Sprintf("roadlog-%s-%s.csv", slugify(vehicle.Name), time.Now().Format("2006-01-02"))
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, "text/csv", []byte(sb.String()))
}

func ExportExpenseCSV(c *gin.Context) {
	vehicleID := c.Param("id")
	var vehicle models.Vehicle
	if err := db.DB.First(&vehicle, vehicleID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var expenses []models.Expense
	db.DB.Where("vehicle_id = ?", vehicleID).Order("date asc").Find(&expenses)

	var sb strings.Builder
	sb.WriteString("Date,Amount,Category,Notes\n")
	for _, e := range expenses {
		sb.WriteString(fmt.Sprintf("%s,%.2f,%s,%s\n",
			e.Date.Format("2006-01-02"), e.Amount, csvEscape(e.Category), csvEscape(e.Notes)))
	}

	filename := fmt.Sprintf("roadlog-expenses-%s-%s.csv", slugify(vehicle.Name), time.Now().Format("2006-01-02"))
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, "text/csv", []byte(sb.String()))
}

func ExportAllJSON(c *gin.Context) {
	Backup(c)
}

func csvEscape(s string) string {
	if strings.ContainsAny(s, ",\"\n") {
		return "\"" + strings.ReplaceAll(s, "\"", "\"\"") + "\""
	}
	return s
}

func slugify(s string) string {
	r := strings.ToLower(s)
	r = strings.ReplaceAll(r, " ", "-")
	return r
}

// Helper to pretty-print backup for file storage
func backupToBytes(data BackupData) []byte {
	b, _ := json.MarshalIndent(data, "", "  ")
	return b
}
