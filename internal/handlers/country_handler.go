package handlers

import (
	"database/sql"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"countryCurrency/internal/database"
	"countryCurrency/internal/models"
	"countryCurrency/internal/services"
)

type CountryHandler struct {
	repo           *database.Repository
	countryService *services.CountryService
	imageService   *services.ImageService
}

func NewCountryHandler(repo *database.Repository, countryService *services.CountryService, imageService *services.ImageService) *CountryHandler {
	return &CountryHandler{
		repo:           repo,
		countryService: countryService,
		imageService:   imageService,
	}
}

func (h *CountryHandler) RefreshCountries(c *gin.Context) {
	ctx := c.Request.Context()

	if err := h.countryService.RefreshCountries(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
			Error:   "External data source unavailable",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Countries refreshed successfully",
	})
	return
}

func (h *CountryHandler) GetAllCountries(c *gin.Context) {
	region := c.Query("region")
	currency := c.Query("currency")
	sort := c.Query("sort")

	countries, err := h.repo.GetAllCountries(region, currency, sort)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal server error",
			Details: err.Error(),
		})
		return
	}

	if countries == nil {
		countries = []models.Country{}
	}

	c.JSON(http.StatusOK, countries)
}

func (h *CountryHandler) GetCountryByName(c *gin.Context) {
	name := c.Param("name")

	if name == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Validation failed",
			Details: models.ValidationErrorDetails{
				"name": "is required",
			},
		})
		return
	}

	country, err := h.repo.GetCountryByName(name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal server error",
			Details: err.Error(),
		})
		return
	}

	if country == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "Country not found",
		})
		return
	}

	c.JSON(http.StatusOK, country)
}

func (h *CountryHandler) DeleteCountryByName(c *gin.Context) {
	name := c.Param("name")

	if name == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Validation failed",
			Details: models.ValidationErrorDetails{
				"name": "is required",
			},
		})
		return
	}

	err := h.repo.DeleteCountryByName(name)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "Country not found",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal server error",
			Details: err.Error(),
		})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *CountryHandler) GetStatus(c *gin.Context) {
	total, err := h.repo.GetTotalCountries()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal server error",
			Details: err.Error(),
		})
		return
	}

	lastRefresh, err := h.repo.GetLastRefreshedAt()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal server error",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.StatusResponse{
		TotalCountries:  total,
		LastRefreshedAt: lastRefresh,
	})
}

func (h *CountryHandler) GetSummaryImage(c *gin.Context) {
	imagePath := h.imageService.GetImagePath()

	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "Summary image not found",
		})
		return
	}

	c.File(imagePath)
}
