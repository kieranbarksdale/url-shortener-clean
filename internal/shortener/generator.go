package shortener

import (
	"database/sql"
	"fmt"
	"math/rand"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GenerateCode(db *sql.DB) (string, error) {

	for range 10 {

		b := make([]byte, 6)
		for i := range 6 {
			b[i] = charset[rand.Intn(len(charset))]
		}

		var exists bool 
		if err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM urls WHERE short_code = $1)", string(b)).Scan(&exists); err != nil {
			return "", fmt.Errorf("error checking database: %w", err)
		}
		if !exists {
			return string(b), nil
		}
	}

	return "", fmt.Errorf("error generating code after 10 attempts")

}
