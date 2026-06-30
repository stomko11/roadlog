package handlers

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"roadlog/db"
	"roadlog/models"

	"github.com/gin-gonic/gin"
)

func GetBackupConfig(c *gin.Context) {
	var cfg models.BackupConfig
	db.DB.FirstOrCreate(&cfg, models.BackupConfig{ID: 1})
	if cfg.Retain == 0 {
		cfg.Retain = 7
	}
	c.JSON(http.StatusOK, cfg)
}

func UpdateBackupConfig(c *gin.Context) {
	var input models.BackupConfig
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var cfg models.BackupConfig
	db.DB.FirstOrCreate(&cfg, models.BackupConfig{ID: 1})
	db.DB.Model(&cfg).Updates(map[string]interface{}{
		"enabled":      input.Enabled,
		"type":         input.Type,
		"schedule":     input.Schedule,
		"time":         input.Time,
		"weekday":      input.Weekday,
		"day_of_month": input.DayOfMonth,
		"retain":       input.Retain,
		"url":          input.URL,
		"username":     input.Username,
		"password":     input.Password,
		"path":         input.Path,
	})
	db.DB.First(&cfg, 1)
	c.JSON(http.StatusOK, cfg)
}

func TriggerBackupNow(c *gin.Context) {
	err := runAutoBackup()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func RunBackupScheduler() {
	go func() {
		for {
			time.Sleep(1 * time.Minute)
			var cfg models.BackupConfig
			if err := db.DB.First(&cfg, 1).Error; err != nil || !cfg.Enabled {
				continue
			}
			if !shouldRun(cfg) {
				continue
			}
			log.Println("[backup] Starting scheduled backup...")
			if err := runAutoBackup(); err != nil {
				log.Printf("[backup] Failed: %v", err)
			} else {
				log.Println("[backup] Completed successfully")
			}
		}
	}()
}

func shouldRun(cfg models.BackupConfig) bool {
	now := time.Now()

	// Parse configured time (default 03:00)
	hour, minute := 3, 0
	if cfg.Time != "" {
		fmt.Sscanf(cfg.Time, "%d:%d", &hour, &minute)
	}

	// Check if we're in the right minute window
	if now.Hour() != hour || now.Minute() != minute {
		return false
	}

	// Check day constraints for weekly/monthly
	switch cfg.Schedule {
	case "weekly":
		if int(now.Weekday()) != cfg.Weekday {
			return false
		}
	case "monthly":
		day := cfg.DayOfMonth
		if day == 0 {
			day = 1
		}
		if now.Day() != day {
			return false
		}
	}

	// Don't run if we already ran successfully today
	if cfg.LastRun != nil {
		lastDate := cfg.LastRun.Format("2006-01-02")
		todayDate := now.Format("2006-01-02")
		if lastDate == todayDate && cfg.LastStatus == "ok" {
			return false
		}
	}

	return true
}

func runAutoBackup() error {
	var cfg models.BackupConfig
	if err := db.DB.First(&cfg, 1).Error; err != nil {
		return err
	}

	var data BackupData
	data.ExportedAt = time.Now().Format(time.RFC3339)
	db.DB.Find(&data.Users)
	db.DB.Find(&data.Vehicles)
	db.DB.Find(&data.Fillups)
	db.DB.Find(&data.Expenses)
	db.DB.Find(&data.Settings)
	payload := backupToBytes(data)

	filename := fmt.Sprintf("roadlog-backup-%s.json", time.Now().Format("2006-01-02-150405"))

	var err error
	switch cfg.Type {
	case "webdav":
		err = uploadWebDAV(cfg, filename, payload)
	case "local":
		err = saveLocal(cfg, filename, payload)
	default:
		err = fmt.Errorf("unknown backup type: %s", cfg.Type)
	}

	now := time.Now()
	status := "ok"
	if err != nil {
		status = err.Error()
		db.DB.Model(&cfg).Update("last_status", status)
	} else {
		db.DB.Model(&cfg).Updates(map[string]interface{}{"last_run": now, "last_status": status})
	}

	if err == nil {
		cleanOldBackups(cfg)
	}
	return err
}

func uploadWebDAV(cfg models.BackupConfig, filename string, data []byte) error {
	url := strings.TrimRight(cfg.URL, "/") + "/" + filename
	req, err := http.NewRequest("PUT", url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if cfg.Username != "" {
		req.SetBasicAuth(cfg.Username, cfg.Password)
	}
	resp, err := (&http.Client{Timeout: 60 * time.Second}).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("WebDAV returned %d: %s", resp.StatusCode, string(body[:min(len(body), 200)]))
	}
	return nil
}

func saveLocal(cfg models.BackupConfig, filename string, data []byte) error {
	dir := cfg.Path
	if dir == "" {
		return fmt.Errorf("local path not configured")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, filename), data, 0o644)
}

func cleanOldBackups(cfg models.BackupConfig) {
	retain := cfg.Retain
	if retain <= 0 {
		retain = 7
	}
	switch cfg.Type {
	case "local":
		cleanLocalBackups(cfg.Path, retain)
	case "webdav":
		// WebDAV cleanup would need PROPFIND+DELETE; skip for now — users can set retain on Nextcloud side
	}
}

func cleanLocalBackups(dir string, retain int) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	var backups []string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "roadlog-backup-") && strings.HasSuffix(e.Name(), ".json") {
			backups = append(backups, e.Name())
		}
	}
	sort.Strings(backups)
	if len(backups) <= retain {
		return
	}
	for _, name := range backups[:len(backups)-retain] {
		os.Remove(filepath.Join(dir, name))
	}
}
