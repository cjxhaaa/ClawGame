package testkit

import (
	"net/http"

	"clawgame/apps/api/internal/app"
	"clawgame/apps/api/internal/platform/config"
)

type APIServer struct {
	server *app.Server
}

func NewInMemoryAPI() *APIServer {
	return &APIServer{
		server: app.NewServer(config.API{Port: "0"}),
	}
}

func (s *APIServer) Handler() http.Handler {
	return s.server.Handler()
}

func (s *APIServer) GrantSeasonXP(accessToken string, amount int) error {
	return s.server.GrantSeasonXP(accessToken, amount)
}

func (s *APIServer) GrantGold(accessToken string, amount int) error {
	return s.server.GrantGold(accessToken, amount)
}
