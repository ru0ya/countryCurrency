package models

import "time"

type Country struct {
	ID              int64     `json:"id" db:"id"`
	Name            string    `json:"name" db:"name"`
	Capital         *string   `json:"capital" db:"capital"`
	Region          *string   `json:"region" db:"region"`
	Population      int64     `json:"population" db:"population"`
	CurrencyCode    *string   `json:"currency_code" db:"currency_code"`
	ExchangeRate    *float64  `json:"exchange_rate" db:"exchange_rate"`
	EstimatedGDP    *float64  `json:"estimated_gdp" db:"estimated_gdp"`
	FlagURL         *string   `json:"flag_url" db:"flag_url"`
	LastRefreshedAt time.Time `json:"last_refreshed_at" db:"last_refreshed_at"`
}

type CountryAPIResponse struct {
	Name       string `json:"name"`
	Capital    string `json:"capital"`
	Region     string `json:"region"`
	Population int64  `json:"population"`
	Flag       string `json:"flag"`
	Currencies []struct {
		Code   string `json:"code"`
		Name   string `json:"name"`
		Symbol string `json:"symbol"`
	} `json:"currencies"`
}

type ExchangeRateResponse struct {
	Result string             `json:"result"`
	Rates  map[string]float64 `json:"rates"`
}

type StatusResponse struct {
	TotalCountries  int       `json:"total_countries"`
	LastRefreshedAt time.Time `json:"last_refreshed_at"`
}

type ErrorResponse struct {
	Error   string      `json:"error"`
	Details interface{} `json:"details,omitempty"`
}

type ValidationErrorDetails map[string]string
