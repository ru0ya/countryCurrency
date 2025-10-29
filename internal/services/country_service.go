package services

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"countryCurrency/internal/database"
	"countryCurrency/internal/models"
)

type CountryService struct {
	repo       *database.Repository
	apiClient  *APIClient
	imgService *ImageService
}

func NewCountryService(repo *database.Repository, apiClient *APIClient, imgService *ImageService) *CountryService {
	return &CountryService{
		repo:       repo,
		apiClient:  apiClient,
		imgService: imgService,
	}
}

func (s *CountryService) RefreshCountries(ctx context.Context) error {
	countriesData, err := s.apiClient.FetchCountries(ctx)
	if err != nil {
		return fmt.Errorf("could not fetch data from countries API: %w", err)
	}

	exchangeRates, err := s.apiClient.FetchExchangeRates(ctx)
	if err != nil {
		return fmt.Errorf("could not fetch data from exchange rate API: %w", err)
	}

	now := time.Now()
	for _, apiCountry := range countriesData {
		country := s.transformCountry(apiCountry, exchangeRates, now)

		if err := s.repo.UpsertCountry(&country); err != nil {
			fmt.Printf("Warning: failed to upsert country %s: %v\n", country.Name, err)
			continue
		}
	}

	if err := s.repo.UpdateLastRefreshedAt(); err != nil {
		return fmt.Errorf("failed to update refresh timestamp: %w", err)
	}

	if err := s.imgService.GenerateSummaryImage(); err != nil {
		fmt.Printf("Warning: failed to generate summary image: %v\n", err)
	}

	return nil
}

func (s *CountryService) transformCountry(
	apiCountry models.CountryAPIResponse,
	exchangeRates map[string]float64,
	refreshTime time.Time,
) models.Country {
	country := models.Country{
		Name:            apiCountry.Name,
		Population:      apiCountry.Population,
		LastRefreshedAt: refreshTime,
	}

	if apiCountry.Capital != "" {
		country.Capital = &apiCountry.Capital
	}
	if apiCountry.Region != "" {
		country.Region = &apiCountry.Region
	}
	if apiCountry.Flag != "" {
		country.FlagURL = &apiCountry.Flag
	}

	if len(apiCountry.Currencies) > 0 {
		currencyCode := apiCountry.Currencies[0].Code
		if currencyCode != "" {
			country.CurrencyCode = &currencyCode

			if rate, exists := exchangeRates[currencyCode]; exists {
				country.ExchangeRate = &rate

				gdp := s.calculateEstimatedGDP(apiCountry.Population, rate)
				country.EstimatedGDP = &gdp
			} else {
				country.ExchangeRate = nil
				country.EstimatedGDP = nil
			}
		}
	}

	return country
}

func (s *CountryService) calculateEstimatedGDP(population int64, exchangeRate float64) float64 {
	// Note: Using random multiplier for GDP estimation is not ideal
	// Consider using a deterministic calculation based on real economic data
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	multiplier := 1000 + r.Float64()*1000

	gdp := float64(population) * multiplier / exchangeRate

	return gdp
}
