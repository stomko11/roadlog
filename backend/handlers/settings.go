package handlers

import (
	"net/http"
	"roadlog/db"
	"roadlog/models"

	"github.com/gin-gonic/gin"
)

type SettingsResponse struct {
	PrefillPrice          bool   `json:"prefillPrice"`
	PrefillStation        bool   `json:"prefillStation"`
	PrefillOdometer       bool   `json:"prefillOdometer"`
	PrefillPricePerStation bool  `json:"prefillPricePerStation"`
	Currency              string `json:"currency"`
	VolumeUnit            string `json:"volumeUnit"`
	DistanceUnit          string `json:"distanceUnit"`
	DateFormat            string `json:"dateFormat"`
	Timezone              string `json:"timezone"`
}

func GetSettings(c *gin.Context) {
	s := SettingsResponse{PrefillPrice: true, PrefillStation: true, PrefillOdometer: true, PrefillPricePerStation: true, Currency: "EUR", VolumeUnit: "L", DistanceUnit: "km", DateFormat: "DD/MM/YYYY"}
	var prefs []models.UserPreference
	db.DB.Where("`key` IN ?", []string{"prefill_price", "prefill_station", "prefill_odometer", "prefill_price_per_station", "currency", "volume_unit", "distance_unit", "date_format", "timezone"}).Find(&prefs)
	for _, p := range prefs {
		switch p.Key {
		case "prefill_price":
			if p.Value == "false" { s.PrefillPrice = false }
		case "prefill_station":
			if p.Value == "false" { s.PrefillStation = false }
		case "prefill_odometer":
			if p.Value == "false" { s.PrefillOdometer = false }
		case "prefill_price_per_station":
			if p.Value == "false" { s.PrefillPricePerStation = false }
		case "currency":
			if p.Value != "" { s.Currency = p.Value }
		case "volume_unit":
			if p.Value != "" { s.VolumeUnit = p.Value }
		case "distance_unit":
			if p.Value != "" { s.DistanceUnit = p.Value }
		case "date_format":
			if p.Value != "" { s.DateFormat = p.Value }
		case "timezone":
			if p.Value != "" { s.Timezone = p.Value }
		}
	}
	c.JSON(http.StatusOK, s)
}

func UpdateSettings(c *gin.Context) {
	var input SettingsResponse
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	upsert := func(key string, val string) {
		db.DB.Where(models.UserPreference{Key: key}).Assign(models.UserPreference{Value: val}).FirstOrCreate(&models.UserPreference{})
	}
	boolStr := func(b bool) string { if b { return "true" }; return "false" }
	upsert("prefill_price", boolStr(input.PrefillPrice))
	upsert("prefill_station", boolStr(input.PrefillStation))
	upsert("prefill_odometer", boolStr(input.PrefillOdometer))
	upsert("prefill_price_per_station", boolStr(input.PrefillPricePerStation))
	upsert("currency", input.Currency)
	upsert("volume_unit", input.VolumeUnit)
	upsert("distance_unit", input.DistanceUnit)
	upsert("date_format", input.DateFormat)
	upsert("timezone", input.Timezone)
	c.JSON(http.StatusOK, input)
}
