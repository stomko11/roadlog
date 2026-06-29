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
	ID               uint       `json:"id" gorm:"primaryKey"`
	VehicleID        uint       `json:"vehicleId"`
	Date             time.Time  `json:"date"`
	Amount           float64    `json:"amount"`
	Category         string     `json:"category"`
	Notes            string     `json:"notes"`
	Recurring        bool       `json:"recurring"`
	Interval         string     `json:"interval"`
	NextDue          *time.Time `json:"nextDue"`
	NotifyDaysBefore int        `json:"notifyDaysBefore" gorm:"default:7"`
	RecurringActive  *bool      `json:"recurringActive" gorm:"default:true"`
	NotifiedAt       *time.Time `json:"notifiedAt"`
	CreatedAt        time.Time  `json:"createdAt"`
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

type Reminder struct {
	ID               uint       `json:"id" gorm:"primaryKey"`
	VehicleID        uint       `json:"vehicleId"`
	Title            string     `json:"title"`
	DueDate          *time.Time `json:"dueDate"`
	DueOdometer      *float64   `json:"dueOdometer"`
	RepeatMonths     int        `json:"repeatMonths"`
	RepeatKm         float64    `json:"repeatKm"`
	NotifyDaysBefore int        `json:"notifyDaysBefore" gorm:"default:7"`
	Notes            string     `json:"notes"`
	Done             bool       `json:"done"`
	DoneDate         *time.Time `json:"doneDate"`
	NotifiedAt       *time.Time `json:"notifiedAt"`
	CreatedAt        time.Time  `json:"createdAt"`
}

type NotificationConfig struct {
	ID      uint   `json:"id" gorm:"primaryKey"`
	Type    string `json:"type"`
	Enabled bool   `json:"enabled"`
	Config  string `json:"config"`
}

type AuditEntry struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	UserID     uint      `json:"userId"`
	Action     string    `json:"action"`
	EntityType string    `json:"entityType"`
	EntityID   uint      `json:"entityId"`
	Details    string    `json:"details"`
	IP         string    `json:"ip"`
	CreatedAt  time.Time `json:"createdAt"`
}

type BackupConfig struct {
	ID         uint       `json:"id" gorm:"primaryKey"`
	Enabled    bool       `json:"enabled"`
	Type       string     `json:"type"`       // "webdav" or "local"
	Schedule   string     `json:"schedule"`   // "daily" or "weekly"
	Retain     int        `json:"retain"`     // number of backups to keep
	URL        string     `json:"url"`        // WebDAV URL
	Username   string     `json:"username"`
	Password   string     `json:"password"`
	Path       string     `json:"path"`       // local path
	LastRun    *time.Time `json:"lastRun"`
	LastStatus string     `json:"lastStatus"` // "ok" or error message
}
