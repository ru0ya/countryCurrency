package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"

	"countryCurrency/internal/config"
	"countryCurrency/internal/database"
	"countryCurrency/internal/handlers"
	"countryCurrency/internal/services"
)

func main() {
	// Note: rand.Seed is deprecated in Go 1.20+, but kept for Go 1.24 compatibility
	// Each service that needs random numbers creates its own source

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	db, err := database.Connect(cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Database connected successfully")

	repo := database.NewRepository(db)

	apiClient := services.NewAPIClient(cfg.CountriesAPIURL, cfg.ExchangeAPIURL)

	imageService := services.NewImageService(repo, "./cache/summary.png")

	countryService := services.NewCountryService(repo, apiClient, imageService)

	countryHandler := handlers.NewCountryHandler(repo, countryService, imageService)

	router := setupRouter(countryHandler)

	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("Server starting on %s", addr)

	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func setupRouter(handler *handlers.CountryHandler) *gin.Engine {
	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	countryRoutes := router.Group("/countries")
	{
		countryRoutes.POST("/refresh", handler.RefreshCountries)
		countryRoutes.GET("", handler.GetAllCountries)
		countryRoutes.GET("/image", handler.GetSummaryImage)
		countryRoutes.GET("/:name", handler.GetCountryByName)
		countryRoutes.DELETE("/:name", handler.DeleteCountryByName)
	}

	router.GET("/status", handler.GetStatus)

	return router
}
