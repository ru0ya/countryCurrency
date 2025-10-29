package database

import (
	"database/sql"
	"fmt"
	"time"

	"countryCurrency/internal/models"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) UpsertCountry(country *models.Country) error {
	query := `
		INSERT INTO countries (name, capital, region, population, currency_code, exchange_rate, estimated_gdp, flag_url, last_refreshed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			capital = VALUES(capital),
			region = VALUES(region),
			population = VALUES(population),
			currency_code = VALUES(currency_code),
			exchange_rate = VALUES(exchange_rate),
			estimated_gdp = VALUES(estimated_gdp),
			flag_url = VALUES(flag_url),
			last_refreshed_at = VALUES(last_refreshed_at)
	`

	result, err := r.db.Exec(
		query,
		country.Name,
		country.Capital,
		country.Region,
		country.Population,
		country.CurrencyCode,
		country.ExchangeRate,
		country.EstimatedGDP,
		country.FlagURL,
		country.LastRefreshedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to upsert country: %w", err)
	}

	// Get the last insert ID
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	country.ID = id

	return nil
}

func (r *Repository) GetAllCountries(region, currency, sort string) ([]models.Country, error) {

	query := "SELECT id, name, capital, region, population, currency_code, exchange_rate, estimated_gdp, flag_url, last_refreshed_at FROM countries WHERE 1=1"
	args := []interface{}{}

	// Add region filter if provided
	if region != "" {
		query += " AND LOWER(region) = LOWER(?)"
		args = append(args, region)
	}

	// Add currency filter if provided
	if currency != "" {
		query += " AND LOWER(currency_code) = LOWER(?)"
		args = append(args, currency)
	}

	switch sort {
	case "gdp_desc":
		query += " ORDER BY estimated_gdp IS NULL, estimated_gdp DESC"
	case "gdp_asc":
		query += " ORDER BY estimated_gdp IS NULL, estimated_gdp ASC"
	case "population_desc":
		query += " ORDER BY population DESC"
	case "population_asc":
		query += " ORDER BY population ASC"
	case "name_asc":
		query += " ORDER BY name ASC"
	case "name_desc":
		query += " ORDER BY name DESC"
	default:
		query += " ORDER BY name ASC" // Default sort
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query countries: %w", err)
	}
	defer rows.Close()

	countries := []models.Country{}
	for rows.Next() {
		var c models.Country
		err := rows.Scan(
			&c.ID,
			&c.Name,
			&c.Capital,
			&c.Region,
			&c.Population,
			&c.CurrencyCode,
			&c.ExchangeRate,
			&c.EstimatedGDP,
			&c.FlagURL,
			&c.LastRefreshedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan country: %w", err)
		}
		countries = append(countries, c)
	}

	// Check for errors during iteration
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return countries, nil
}

func (r *Repository) GetCountryByName(name string) (*models.Country, error) {
	query := `
		SELECT id, name, capital, region, population, currency_code, exchange_rate, estimated_gdp, flag_url, last_refreshed_at
		FROM countries
		WHERE LOWER(name) = LOWER(?)
	`

	var c models.Country
	err := r.db.QueryRow(query, name).Scan(
		&c.ID,
		&c.Name,
		&c.Capital,
		&c.Region,
		&c.Population,
		&c.CurrencyCode,
		&c.ExchangeRate,
		&c.EstimatedGDP,
		&c.FlagURL,
		&c.LastRefreshedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Not found, return nil (not an error)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get country: %w", err)
	}

	return &c, nil
}

// DeleteCountryByName deletes a country by name
func (r *Repository) DeleteCountryByName(name string) error {
	query := "DELETE FROM countries WHERE LOWER(name) = LOWER(?)"
	result, err := r.db.Exec(query, name)
	if err != nil {
		return fmt.Errorf("failed to delete country: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows // Not found
	}

	return nil
}

// GetTotalCountries returns the count of all countries
func (r *Repository) GetTotalCountries() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM countries").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count countries: %w", err)
	}
	return count, nil
}

// GetLastRefreshedAt retrieves the last refresh timestamp from metadata
func (r *Repository) GetLastRefreshedAt() (time.Time, error) {
	var timestamp string
	err := r.db.QueryRow("SELECT value FROM metadata WHERE `key` = 'last_refreshed_at'").Scan(&timestamp)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get last refresh time: %w", err)
	}

	t, err := time.Parse("2006-01-02 15:04:05", timestamp)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse timestamp: %w", err)
	}

	return t, nil
}

func (r *Repository) UpdateLastRefreshedAt() error {
	query := "UPDATE metadata SET value = NOW(), updated_at = NOW() WHERE `key` = 'last_refreshed_at'"
	_, err := r.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to update last refresh time: %w", err)
	}
	return nil
}

func (r *Repository) GetTopCountriesByGDP(limit int) ([]models.Country, error) {
	query := `
		SELECT id, name, capital, region, population, currency_code, exchange_rate, estimated_gdp, flag_url, last_refreshed_at
		FROM countries
		WHERE estimated_gdp IS NOT NULL
		ORDER BY estimated_gdp DESC
		LIMIT ?
	`

	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query top countries: %w", err)
	}
	defer rows.Close()

	countries := []models.Country{}
	for rows.Next() {
		var c models.Country
		err := rows.Scan(
			&c.ID,
			&c.Name,
			&c.Capital,
			&c.Region,
			&c.Population,
			&c.CurrencyCode,
			&c.ExchangeRate,
			&c.EstimatedGDP,
			&c.FlagURL,
			&c.LastRefreshedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan country: %w", err)
		}
		countries = append(countries, c)
	}

	// Check for row iteration errors
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return countries, nil
}
