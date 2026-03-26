package main

import (
	"log"

	"clawgame/apps/api/internal/app"
	"clawgame/apps/api/internal/platform/config"
)

func main() {
	cfg := config.LoadAPI()

	server := app.NewServer(cfg)
	log.Printf("api starting on %s", cfg.Addr())

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

