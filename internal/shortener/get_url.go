package shortener

import "database/sql"

func GetURL(db *sql.DB, shortCode string) (string, error) {
	var originalURL string
	err := db.QueryRow("SELECT original_url FROM urls WHERE short_code = $1", shortCode).Scan(&originalURL)
	if err != nil {
		return "", err
	}
	return originalURL, nil
}