package handlers

import (
	"net/http"
	"roadlog/db"
	"roadlog/models"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
)

// HistoryItem is a unified fill-up / expense row for the global History page.
// The frontend does all filtering (car, station, price, date, type) client-side against this list,
// which keeps the filter UI instant and flexible; the dataset for a personal tracker is small.
type HistoryItem struct {
	Type         string    `json:"type"` // "fillup" | "expense"
	ID           uint      `json:"id"`
	VehicleID    uint      `json:"vehicleId"`
	VehicleName  string    `json:"vehicleName"`
	VehicleColor string    `json:"vehicleColor"`
	Date         time.Time `json:"date"`
	Amount       float64   `json:"amount"`       // fill-up total cost, or expense amount
	Station      string    `json:"station"`      // fill-ups only
	Category     string    `json:"category"`     // expenses only
	PricePerUnit float64   `json:"pricePerUnit"` // fill-ups only
	FuelAmount   float64   `json:"fuelAmount"`   // fill-ups only
	Notes        string    `json:"notes"`
}

func GetHistory(c *gin.Context) {
	var vehicles []models.Vehicle
	db.DB.Find(&vehicles)
	vName := map[uint]string{}
	vColor := map[uint]string{}
	for _, v := range vehicles {
		vName[v.ID] = v.Name
		vColor[v.ID] = v.Color
	}

	items := []HistoryItem{}

	var fillups []models.Fillup
	db.DB.Order("date desc").Find(&fillups)
	for _, f := range fillups {
		items = append(items, HistoryItem{
			Type: "fillup", ID: f.ID, VehicleID: f.VehicleID,
			VehicleName: vName[f.VehicleID], VehicleColor: vColor[f.VehicleID],
			Date: f.Date, Amount: f.TotalCost, Station: f.Station,
			PricePerUnit: f.PricePerUnit, FuelAmount: f.FuelAmount, Notes: f.Notes,
		})
	}

	// Exclude recurring templates (the same rule the dashboard uses) - only real, dated expenses.
	var expenses []models.Expense
	db.DB.Where("recurring = ?", false).Order("date desc").Find(&expenses)
	for _, e := range expenses {
		items = append(items, HistoryItem{
			Type: "expense", ID: e.ID, VehicleID: e.VehicleID,
			VehicleName: vName[e.VehicleID], VehicleColor: vColor[e.VehicleID],
			Date: e.Date, Amount: e.Amount, Category: e.Category, Notes: e.Notes,
		})
	}

	sort.Slice(items, func(i, j int) bool { return items[i].Date.After(items[j].Date) })
	c.JSON(http.StatusOK, items)
}
