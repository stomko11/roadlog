package handlers

import (
	"net/http"
	"roadlog/db"
	"roadlog/models"
	"time"

	"github.com/gin-gonic/gin"
)

type MonthSummary struct {
	Month        string  `json:"month"`
	TotalSpent   float64 `json:"totalSpent"`
	FuelSpent    float64 `json:"fuelSpent"`
	ExpenseSpent float64 `json:"expenseSpent"`
	Fillups      int     `json:"fillups"`
	Liters       float64 `json:"liters"`
}

type VehicleMonthly struct {
	VehicleID   uint           `json:"vehicleId"`
	VehicleName string         `json:"vehicleName"`
	Color       string         `json:"color"`
	Monthly     []MonthSummary `json:"monthly"`
}

type DashboardResponse struct {
	TotalVehicles  int              `json:"totalVehicles"`
	TotalFillups   int64            `json:"totalFillups"`
	TotalSpent     float64          `json:"totalSpent"`
	FuelSpent      float64          `json:"fuelSpent"`
	ExpenseSpent   float64          `json:"expenseSpent"`
	RecurringSpent float64          `json:"recurringSpent"`
	TotalDistance  float64          `json:"totalDistance"`
	Monthly        []MonthSummary   `json:"monthly"`
	PerVehicle     []VehicleMonthly `json:"perVehicle"`
}

func GetDashboard(c *gin.Context) {
	var d DashboardResponse
	var vehicles []models.Vehicle
	db.DB.Find(&vehicles)

	// Build set of vehicle IDs to include in stats
	statsVehicleIDs := []uint{}
	for _, v := range vehicles {
		if v.ShowInStats == nil || *v.ShowInStats {
			statsVehicleIDs = append(statsVehicleIDs, v.ID)
		}
	}
	d.TotalVehicles = len(statsVehicleIDs)

	from := c.Query("from")
	to := c.Query("to")

	var fillups []models.Fillup
	q := db.DB.Where("vehicle_id IN ?", statsVehicleIDs).Order("date desc")
	if from != "" && to != "" {
		q = q.Where("date >= ? AND date <= ?", from+"T00:00:00Z", to+"T23:59:59Z")
	} else if from != "" {
		q = q.Where("date >= ?", from+"T00:00:00Z")
	}
	q.Find(&fillups)
	d.TotalFillups = int64(len(fillups))

	monthly := map[string]*MonthSummary{}
	vOdoMin := map[uint]float64{}
	vOdoMax := map[uint]float64{}
	for _, f := range fillups {
		d.TotalSpent += f.TotalCost
		d.FuelSpent += f.TotalCost
		key := f.Date.Format("2006-01")
		if _, ok := monthly[key]; !ok {
			monthly[key] = &MonthSummary{Month: key}
		}
		monthly[key].TotalSpent += f.TotalCost
		monthly[key].FuelSpent += f.TotalCost
		monthly[key].Fillups++
		monthly[key].Liters += f.FuelAmount
		if f.Odometer > 0 {
			if _, ok := vOdoMin[f.VehicleID]; !ok || f.Odometer < vOdoMin[f.VehicleID] {
				vOdoMin[f.VehicleID] = f.Odometer
			}
			if f.Odometer > vOdoMax[f.VehicleID] {
				vOdoMax[f.VehicleID] = f.Odometer
			}
		}
	}
	for vid, max := range vOdoMax {
		d.TotalDistance += max - vOdoMin[vid]
	}

	// Add expenses (exclude recurring templates)
	var expenses []models.Expense
	qe := db.DB.Where("vehicle_id IN ? AND recurring = ?", statsVehicleIDs, false).Order("date desc")
	if from != "" && to != "" {
		qe = qe.Where("date >= ? AND date <= ?", from+"T00:00:00Z", to+"T23:59:59Z")
	} else if from != "" {
		qe = qe.Where("date >= ?", from+"T00:00:00Z")
	}
	qe.Find(&expenses)
	for _, e := range expenses {
		d.TotalSpent += e.Amount
		d.ExpenseSpent += e.Amount
		key := e.Date.Format("2006-01")
		if _, ok := monthly[key]; !ok {
			monthly[key] = &MonthSummary{Month: key}
		}
		monthly[key].TotalSpent += e.Amount
		monthly[key].ExpenseSpent += e.Amount
	}

	for _, m := range monthly {
		d.Monthly = append(d.Monthly, *m)
	}
	for i := 0; i < len(d.Monthly); i++ {
		for j := i + 1; j < len(d.Monthly); j++ {
			if d.Monthly[j].Month > d.Monthly[i].Month {
				d.Monthly[i], d.Monthly[j] = d.Monthly[j], d.Monthly[i]
			}
		}
	}
	if len(d.Monthly) > 24 {
		d.Monthly = d.Monthly[:24]
	}

	// Per-vehicle breakdown
	for _, v := range vehicles {
		if v.ShowInStats != nil && !*v.ShowInStats {
			continue
		}
		vm := VehicleMonthly{VehicleID: v.ID, VehicleName: v.Name, Color: v.Color}
		vMonthly := map[string]*MonthSummary{}
		for _, f := range fillups {
			if f.VehicleID != v.ID {
				continue
			}
			key := f.Date.Format("2006-01")
			if _, ok := vMonthly[key]; !ok {
				vMonthly[key] = &MonthSummary{Month: key}
			}
			vMonthly[key].TotalSpent += f.TotalCost
			vMonthly[key].FuelSpent += f.TotalCost
			vMonthly[key].Fillups++
		}
		for _, e := range expenses {
			if e.VehicleID != v.ID {
				continue
			}
			key := e.Date.Format("2006-01")
			if _, ok := vMonthly[key]; !ok {
				vMonthly[key] = &MonthSummary{Month: key}
			}
			vMonthly[key].TotalSpent += e.Amount
			vMonthly[key].ExpenseSpent += e.Amount
		}
		for _, m := range vMonthly {
			vm.Monthly = append(vm.Monthly, *m)
		}
		// Sort descending
		for i := 0; i < len(vm.Monthly); i++ {
			for j := i + 1; j < len(vm.Monthly); j++ {
				if vm.Monthly[j].Month > vm.Monthly[i].Month {
					vm.Monthly[i], vm.Monthly[j] = vm.Monthly[j], vm.Monthly[i]
				}
			}
		}
		d.PerVehicle = append(d.PerVehicle, vm)
	}

	c.JSON(http.StatusOK, d)
}

func GetDueSoon(c *gin.Context) {
	now := time.Now()
	soon := now.AddDate(0, 0, 30)

	var reminders []models.Reminder
	db.DB.Where("done = ? AND due_date IS NOT NULL AND due_date <= ?", false, soon).Find(&reminders)

	type ReminderItem struct {
		models.Reminder
		VehicleName string `json:"vehicleName"`
	}
	var reminderItems []ReminderItem
	for _, r := range reminders {
		var v models.Vehicle
		db.DB.First(&v, r.VehicleID)
		reminderItems = append(reminderItems, ReminderItem{Reminder: r, VehicleName: v.Name})
	}

	var expenses []models.Expense
	db.DB.Where("recurring = ? AND recurring_active = ? AND next_due <= ?", true, true, soon).Find(&expenses)

	type ExpenseItem struct {
		models.Expense
		VehicleName string `json:"vehicleName"`
	}
	var expenseItems []ExpenseItem
	for _, e := range expenses {
		var v models.Vehicle
		db.DB.First(&v, e.VehicleID)
		expenseItems = append(expenseItems, ExpenseItem{Expense: e, VehicleName: v.Name})
	}

	c.JSON(http.StatusOK, gin.H{"reminders": reminderItems, "expenses": expenseItems})
}
