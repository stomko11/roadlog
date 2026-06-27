package models

import "time"

type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name"`
	Email     string    `json:"email" gorm:"uniqueIndex"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"createdAt"`
}

type Vehicle struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name"`
	Make        string    `json:"make"`
	Model       string    `json:"model"`
	Year        int       `json:"year"`
	Plate       string    `json:"plate"`
	FuelType    string    `json:"fuelType"`
	Color       string    `json:"color"`
	Active      *bool     `json:"active" gorm:"default:true"`
	ShowInStats *bool     `json:"showInStats" gorm:"default:true"`
	UserID      uint      `json:"userId"`
	CreatedAt   time.Time `json:"createdAt"`
}

type Fillup struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	VehicleID      uint      `json:"vehicleId"`
	Date           time.Time `json:"date"`
	Odometer       float64   `json:"odometer"`
	FuelAmount     float64   `json:"fuelAmount"`
	PricePerUnit   float64   `json:"pricePerUnit"`
	TotalCost      float64   `json:"totalCost"`
	FullTank       bool      `json:"fullTank"`
	MissedPrevious bool      `json:"missedPrevious"`
	Station        string    `json:"station"`
	Notes          string    `json:"notes"`
	CreatedAt      time.Time `json:"createdAt"`
}

type Expense struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	VehicleID uint      `json:"vehicleId"`
	Date      time.Time `json:"date"`
	Amount    float64   `json:"amount"`
	Category  string    `json:"category"`
	Notes     string    `json:"notes"`
	CreatedAt time.Time `json:"createdAt"`
}

type UserPreference struct {
	ID    uint   `json:"id" gorm:"primaryKey"`
	Key   string `json:"key" gorm:"uniqueIndex"`
	Value string `json:"value"`
}

type Station struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Name     string `json:"name"`
	FuelType string `json:"fuelType"`
}

type FillupPrefill struct {
	Odometer     *float64 `json:"odometer"`
	PricePerUnit *float64 `json:"pricePerUnit"`
	Station      *string  `json:"station"`
	FullTank     *bool    `json:"fullTank"`
}

type VehicleEVCC struct {
	ID                  uint      `json:"id" gorm:"primaryKey"`
	VehicleID           uint      `json:"vehicleId"`
	URL                 string    `json:"url"`
	EVCCVehicle         string    `json:"evccVehicle"`
	EVCCLoadpoint       string    `json:"evccLoadpoint"`
	Label               string    `json:"label"`
	SyncSince           string    `json:"syncSince"`
	LastSyncedSessionID int       `json:"lastSyncedSessionId"`
	AutoSync            bool      `json:"autoSync"`
	SyncTime            string    `json:"syncTime"`
	FallbackPrice       float64   `json:"fallbackPrice"`
	CreatedAt           time.Time `json:"createdAt"`
}

type VehicleStats struct {
	TotalFillups   int64   `json:"totalFillups"`
	TotalSpent     float64 `json:"totalSpent"`
	AvgConsumption float64 `json:"avgConsumption"`
	TotalDistance  float64 `json:"totalDistance"`
}
