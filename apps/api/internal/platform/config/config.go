package config

import "os"

type API struct {
	Port string
}

func LoadAPI() API {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return API{
		Port: port,
	}
}

func (c API) Addr() string {
	return ":" + c.Port
}

