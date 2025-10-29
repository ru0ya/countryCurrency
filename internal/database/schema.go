package database

const (
	CreateCountriesTable = `
		CREATE TABLE IF NOT EXISTS countries (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255) NOT NULL UNIQUE,
			capital VARCHAR(255),
			region VARCHAR(100),
			population BIGINT NOT NULL,
			currency_code VARCHAR(10),
			exchange_rate DOUBLE,
			estimated_gdp DOUBLE,
			flag_url TEXT,
			last_refreshed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_region (region),
			INDEX idx_currency (currency_code)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
		`

	CreateMetadataTable = `
		CREATE TABLE IF NOT EXISTS metadata (
			` + "`key`" + ` VARCHAR(255) PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
		`
)

const InitialMetadata = `
		INSERT IGNORE INTO metadata(` + "`key`" + `, value)
		VALUES ('last_refreshed_at', NOW());
		`
