package db

import (
	"os"
	"roadlog/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init() {
	path := "/data/roadlog.db"
	if p := os.Getenv("DATA_DIR"); p != "" && p != "/" {
		path = p + "/roadlog.db"
	}
	var err error
	DB, err = gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	DB.AutoMigrate(&models.User{}, &models.Vehicle{}, &models.Fillup{}, &models.Expense{}, &models.UserPreference{}, &models.Station{})
	seedAdmin()
}

func seedAdmin() {
	var count int64
	DB.Model(&models.User{}).Count(&count)
	if count == 0 {
		hash, _ := bcrypt.GenerateFromPassword([]byte("roadlog"), bcrypt.DefaultCost)
		DB.Create(&models.User{Name: "Admin", Email: "admin@roadlog.local", Password: string(hash)})
	}
}
