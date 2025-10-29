package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"countryCurrency/internal/models"
)

type APIClient struct {
	httpClient      *http.Client
	countriesAPIURL string
	exchangeAPIURL  string
}

func NewAPIClient(countriesURL, exchangeURL string) *APIClient {
	return &APIClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		countriesAPIURL: countriesURL,
		exchangeAPIURL:  exchangeURL,
	}
}

func (c *APIClient) FetchCountries(ctx context.Context) ([]models.CountryAPIResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.countriesAPIURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch countries: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("countries API returned status %d", resp.StatusCode)
	}

	var countries []models.CountryAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&countries); err != nil {
		return nil, fmt.Errorf("failed to decode countries response: %w", err)
	}

	return countries, nil
}

func (c *APIClient) FetchExchangeRates(ctx context.Context) (map[string]float64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.exchangeAPIURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch exchange rates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("exchange rate API returned status %d", resp.StatusCode)
	}

	var result models.ExchangeRateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode exchange rates response: %w", err)
	}

	// Return just the rates map
	// Why? Handler only cares about rates, not other metadata
	return result.Rates, nil
}
