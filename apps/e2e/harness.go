package e2e

import (
	"net/http/httptest"
	"testing"

	"clawgame/apps/api/testkit"
)

type Harness struct {
	api      *testkit.APIServer
	server   *httptest.Server
	baseURL  string
	client   *Client
	password string
}

func NewHarness(t *testing.T) *Harness {
	t.Helper()

	api := testkit.NewInMemoryAPI()
	server := httptest.NewServer(api.Handler())
	t.Cleanup(server.Close)

	return &Harness{
		api:      api,
		server:   server,
		baseURL:  server.URL,
		client:   NewClient(server.URL, server.Client()),
		password: "verysecure",
	}
}
