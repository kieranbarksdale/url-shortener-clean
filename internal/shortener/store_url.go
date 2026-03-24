package shortener

import "database/sql"

func StoreURL(db *sql.DB, shortCode string, originalURL string) error {
	_, err := db.Exec("INSERT INTO urls (short_code, original_url) VALUES ($1, $2)", shortCode, originalURL)
	return err
}
