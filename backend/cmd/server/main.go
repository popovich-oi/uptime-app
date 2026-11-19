package main

import (
	"log"
	"net/http"

	"uptime-app/backend/internal/auth"
	"uptime-app/backend/internal/config"
	authcontroller "uptime-app/backend/internal/controllers/auth"
	monitorcontroller "uptime-app/backend/internal/controllers/monitor"
	"uptime-app/backend/internal/database"
	"uptime-app/backend/internal/httpapi"
)

func main() {
	cfg := config.Load()
	db, err := database.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := database.Close(db); err != nil {
			log.Printf("close database: %v", err)
		}
	}()
	if err := database.Migrate(db); err != nil {
		log.Fatal(err)
	}
	service := auth.NewService(db, cfg.JWTSecret, cfg.AccessTTL, cfg.RefreshTTL)
	mux := http.NewServeMux()
	authcontroller.NewController(service, cfg).RegisterRoutes(mux)
	monitorcontroller.NewController(service, db).RegisterRoutes(mux)
	log.Printf("Server listening on %s", cfg.HTTPAddr)
	log.Fatal(http.ListenAndServe(cfg.HTTPAddr, httpapi.CORS(mux, cfg.CORSAllowedOrigin)))
}
