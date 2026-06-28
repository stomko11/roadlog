package main

import (
	"embed"
	"net/http"
	"os"
	"roadlog/db"
	"roadlog/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

//go:embed static/index.html
var staticFS embed.FS

//go:embed static/favicon.png
var faviconData []byte

func main() {
	if os.Getenv("DATA_DIR") == "" {
		os.Setenv("DATA_DIR", ".")
	}
	db.Init()
	handlers.RunEVCCScheduler()

	r := gin.Default()
	r.Use(cors.Default())

	api := r.Group("/api")
	{
		api.POST("/register", handlers.Register)
		api.POST("/login", handlers.Login)
	}

	auth := api.Group("/", handlers.AuthMiddleware())
	{
		auth.GET("/dashboard", handlers.GetDashboard)
		auth.GET("/settings", handlers.GetSettings)
		auth.PUT("/settings", handlers.UpdateSettings)

		auth.GET("/users", handlers.ListUsers)
		auth.POST("/users", handlers.CreateUser)
		auth.PUT("/users/password", handlers.ChangePassword)
		auth.DELETE("/users/:id", handlers.DeleteUser)

		auth.GET("/backup", handlers.Backup)
		auth.POST("/restore", handlers.Restore)

		auth.GET("/vehicles", handlers.GetVehicles)
		auth.POST("/vehicles", handlers.CreateVehicle)
		auth.GET("/vehicles/:id", handlers.GetVehicle)
		auth.PUT("/vehicles/:id", handlers.UpdateVehicle)
		auth.DELETE("/vehicles/:id", handlers.DeleteVehicle)
		auth.GET("/vehicles/:id/export", handlers.ExportVehicleCSV)

		auth.GET("/vehicles/:id/fillups", handlers.GetFillups)
		auth.POST("/vehicles/:id/fillups", handlers.CreateFillup)
		auth.PUT("/fillups/:id", handlers.UpdateFillup)
		auth.DELETE("/fillups/:id", handlers.DeleteFillup)

		auth.GET("/vehicles/:id/expenses", handlers.GetExpenses)
		auth.POST("/vehicles/:id/expenses", handlers.CreateExpense)
		auth.PUT("/expenses/:id", handlers.UpdateExpense)
		auth.DELETE("/expenses/:id", handlers.DeleteExpense)
		auth.GET("/vehicles/:id/expenses/export", handlers.ExportExpenseCSV)

		auth.POST("/fillups/bulk-delete", handlers.BulkDeleteFillups)
		auth.POST("/expenses/bulk-delete", handlers.BulkDeleteExpenses)

		auth.GET("/vehicles/:id/fillups/prefill", handlers.GetFillupPrefill)
		auth.GET("/vehicles/:id/stats", handlers.GetVehicleStats)
		auth.GET("/vehicles/:id/chart", handlers.GetVehicleChartData)
		auth.GET("/vehicles/:id/ev-stats", handlers.GetEVStats)

		auth.GET("/stations", handlers.GetStations)
		auth.POST("/stations", handlers.CreateStation)
		auth.DELETE("/stations/:id", handlers.DeleteStation)

		auth.POST("/import/parse", handlers.ParseCSVHeaders)
		auth.POST("/import", handlers.ImportCSV)
		auth.POST("/import/expenses", handlers.ImportExpenses)

		auth.GET("/evcc/discover", handlers.EVCCDiscover)
		auth.GET("/vehicles/:id/evcc", handlers.GetVehicleEVCCSources)
		auth.POST("/vehicles/:id/evcc", handlers.CreateVehicleEVCCSource)
		auth.DELETE("/evcc/:id", handlers.DeleteVehicleEVCCSource)
		auth.PUT("/evcc/:id", handlers.UpdateVehicleEVCCSource)
		auth.POST("/vehicles/:id/evcc/sync", handlers.SyncVehicleEVCC)
	}

	r.GET("/favicon.png", func(c *gin.Context) {
		c.Data(http.StatusOK, "image/png", faviconData)
	})

	r.NoRoute(func(c *gin.Context) {
		data, _ := staticFS.ReadFile("static/index.html")
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})

	r.Run(":3000")
}
