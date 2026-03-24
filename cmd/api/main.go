/* 
This file is responsible for wiring the peices of the applicaiton together 
- Starting the server
- Initalizing dependencies
- Handing the above off  

This file does not implement any business logic
*/
package main

import (
	"fmt"
	"log"
	"net/http" 
	"url_shortener/internal/api"
	"url_shortener/internal/cache"
	"url_shortener/internal/config"
	"url_shortener/internal/db"
	"github.com/joho/godotenv"
)

func healthHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(w, "ok")
}

func main() {

	// This is always the order it should be done in 
	// Connect to Database

	godotenv.Load()
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Error loading config: ", err)
	} 
	
	if err := db.Migrate(cfg.DatabaseURL); err != nil {
		log.Fatal("Error migrating database: ", err)
	}

	db, err := db.NewConnection(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Error connecting to database: ", err)
	}
	fmt.Println("Database connection established") 

	// Connect to Redis
	redisClient, err := cache.NewClient(cfg.RedisURL)
	if err != nil {
		log.Fatal("Error connecting to redis: ", err)
	}
	fmt.Println("Redis connection established") 

	// Register Routes
	http.HandleFunc("/health", healthHandler)

	http.HandleFunc("/shorten", func(w http.ResponseWriter, req *http.Request) { 
		api.ShortenHandler(w, req, db, redisClient)
	}) 

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		api.RedirectHandler(w, req, db, redisClient)
	})

	// Start Server
	fmt.Println("Server starting on" + cfg.Port)
	err = http.ListenAndServe(cfg.Port, nil)
	if err != nil {
		log.Fatal("Error starting server: ", err)
	}
	
}
