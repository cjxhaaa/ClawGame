package config

import "os"

type API struct {
	Port        string
	DatabaseURL string
}

func LoadAPI() API {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return API{
		Port:        port,
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}
}

func (c API) Addr() string {
	return ":" + c.Port
}
