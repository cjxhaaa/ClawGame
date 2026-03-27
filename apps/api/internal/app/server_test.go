package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"testing"

	"clawgame/apps/api/internal/platform/config"
)

func TestPublicWorldRoutes(t *testing.T) {
	server := NewServer(config.API{Port: "8080"})

	t.Run("lists seeded regions", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/api/v1/world/regions", nil)

		server.httpServer.Handler.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", recorder.Code)
		}

		var payload struct {
			RequestID string `json:"request_id"`
			Data      struct {
				Regions []struct {
					RegionID string `json:"region_id"`
				} `json:"regions"`
			} `json:"data"`
		}
		decodeJSON(t, recorder, &payload)

		if payload.RequestID == "" {
			t.Fatal("expected request_id to be populated")
		}
		if len(payload.Data.Regions) != 6 {
			t.Fatalf("expected 6 regions, got %d", len(payload.Data.Regions))
		}
		if payload.Data.Regions[0].RegionID != "main_city" {
			t.Fatalf("expected first region to be main_city, got %q", payload.Data.Regions[0].RegionID)
		}
	})

	t.Run("returns region detail", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/api/v1/regions/main_city", nil)

		server.httpServer.Handler.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", recorder.Code)
		}

		var payload struct {
			Data struct {
				Region struct {
					RegionID string `json:"region_id"`
				} `json:"region"`
				Buildings []struct {
					BuildingID string `json:"building_id"`
				} `json:"buildings"`
			} `json:"data"`
		}
		decodeJSON(t, recorder, &payload)

		if payload.Data.Region.RegionID != "main_city" {
			t.Fatalf("expected main_city, got %q", payload.Data.Region.RegionID)
		}
		if len(payload.Data.Buildings) != 7 {
			t.Fatalf("expected 7 buildings in main city, got %d", len(payload.Data.Buildings))
		}
	})

	t.Run("returns 404 for unknown region", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/api/v1/regions/not_real", nil)

		server.httpServer.Handler.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", recorder.Code)
		}

		var payload struct {
			Error struct {
				Code string `json:"code"`
			} `json:"error"`
		}
		decodeJSON(t, recorder, &payload)

		if payload.Error.Code != "REGION_NOT_FOUND" {
			t.Fatalf("expected REGION_NOT_FOUND, got %q", payload.Error.Code)
		}
	})
}

func TestAuthCharacterFlow(t *testing.T) {
	server := NewServer(config.API{Port: "8080"})

	var registerResponse struct {
		Data struct {
			Account struct {
				AccountID string `json:"account_id"`
				BotName   string `json:"bot_name"`
			} `json:"account"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/register", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-alpha",
		"password": "verysecure",
	}), "", http.StatusOK, &registerResponse)

	if registerResponse.Data.Account.AccountID == "" {
		t.Fatal("expected account_id to be returned")
	}
	if registerResponse.Data.Account.BotName != "bot-alpha" {
		t.Fatalf("expected bot-alpha, got %q", registerResponse.Data.Account.BotName)
	}

	var loginResponse struct {
		Data struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/login", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-alpha",
		"password": "verysecure",
	}), "", http.StatusOK, &loginResponse)

	if loginResponse.Data.AccessToken == "" || loginResponse.Data.RefreshToken == "" {
		t.Fatal("expected both access and refresh tokens")
	}

	var createCharacterResponse struct {
		Data struct {
			Character struct {
				CharacterID string `json:"character_id"`
				Name        string `json:"name"`
				Class       string `json:"class"`
				WeaponStyle string `json:"weapon_style"`
				Gold        int    `json:"gold"`
			} `json:"character"`
			RecentEvents []struct {
				EventType string `json:"event_type"`
			} `json:"recent_events"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/characters", map[string]any{
		"name":         "Aster",
		"class":        "mage",
		"weapon_style": "staff",
	}, loginResponse.Data.AccessToken, http.StatusOK, &createCharacterResponse)

	if createCharacterResponse.Data.Character.CharacterID == "" {
		t.Fatal("expected character_id to be returned")
	}
	if createCharacterResponse.Data.Character.Gold != 100 {
		t.Fatalf("expected starter gold 100, got %d", createCharacterResponse.Data.Character.Gold)
	}
	if len(createCharacterResponse.Data.RecentEvents) == 0 || createCharacterResponse.Data.RecentEvents[0].EventType != "character.created" {
		t.Fatal("expected first recent event to be character.created")
	}

	var meResponse struct {
		Data struct {
			Account struct {
				BotName string `json:"bot_name"`
			} `json:"account"`
			Character struct {
				Name string `json:"name"`
			} `json:"character"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me", nil, loginResponse.Data.AccessToken, http.StatusOK, &meResponse)

	if meResponse.Data.Account.BotName != "bot-alpha" {
		t.Fatalf("expected bot-alpha account, got %q", meResponse.Data.Account.BotName)
	}
	if meResponse.Data.Character.Name != "Aster" {
		t.Fatalf("expected character Aster, got %q", meResponse.Data.Character.Name)
	}

	var stateResponse struct {
		Data struct {
			Character struct {
				LocationRegionID string `json:"location_region_id"`
			} `json:"character"`
			Objectives []struct {
				QuestID string `json:"quest_id"`
			} `json:"objectives"`
			ValidActions []struct {
				ActionType string `json:"action_type"`
			} `json:"valid_actions"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/state", nil, loginResponse.Data.AccessToken, http.StatusOK, &stateResponse)

	if stateResponse.Data.Character.LocationRegionID != "main_city" {
		t.Fatalf("expected main_city, got %q", stateResponse.Data.Character.LocationRegionID)
	}
	if len(stateResponse.Data.ValidActions) == 0 {
		t.Fatal("expected valid actions to be populated")
	}
	if len(stateResponse.Data.Objectives) != 0 {
		t.Fatalf("expected no active objectives yet, got %d", len(stateResponse.Data.Objectives))
	}

	var travelResponse struct {
		Data struct {
			ActionResult struct {
				ToRegionID string `json:"to_region_id"`
			} `json:"action_result"`
			State struct {
				Character struct {
					LocationRegionID string `json:"location_region_id"`
					Gold             int    `json:"gold"`
				} `json:"character"`
				RecentEvents []struct {
					EventType string `json:"event_type"`
				} `json:"recent_events"`
			} `json:"state"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/travel", map[string]any{
		"region_id": "whispering_forest",
	}, loginResponse.Data.AccessToken, http.StatusOK, &travelResponse)

	if travelResponse.Data.ActionResult.ToRegionID != "whispering_forest" {
		t.Fatalf("expected whispering_forest, got %q", travelResponse.Data.ActionResult.ToRegionID)
	}
	if travelResponse.Data.State.Character.LocationRegionID != "whispering_forest" {
		t.Fatalf("expected location to update, got %q", travelResponse.Data.State.Character.LocationRegionID)
	}
	if travelResponse.Data.State.Character.Gold != 90 {
		t.Fatalf("expected gold to drop to 90, got %d", travelResponse.Data.State.Character.Gold)
	}
	if len(travelResponse.Data.State.RecentEvents) == 0 || travelResponse.Data.State.RecentEvents[0].EventType != "travel.completed" {
		t.Fatal("expected first recent event to be travel.completed")
	}

	var refreshResponse struct {
		Data struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/refresh", map[string]any{
		"refresh_token": loginResponse.Data.RefreshToken,
	}, "", http.StatusOK, &refreshResponse)

	if refreshResponse.Data.AccessToken == "" || refreshResponse.Data.RefreshToken == "" {
		t.Fatal("expected refresh endpoint to issue replacement tokens")
	}
}

func TestCharacterValidationAndAuthErrors(t *testing.T) {
	server := NewServer(config.API{Port: "8080"})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	server.httpServer.Handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for missing token, got %d", recorder.Code)
	}

	var loginResponse struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/register", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-beta",
		"password": "verysecure",
	}), "", http.StatusOK, nil)
	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/login", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-beta",
		"password": "verysecure",
	}), "", http.StatusOK, &loginResponse)

	var errorResponse struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/characters", map[string]any{
		"name":         "Iris",
		"class":        "mage",
		"weapon_style": "great_axe",
	}, loginResponse.Data.AccessToken, http.StatusBadRequest, &errorResponse)

	if errorResponse.Error.Code != "CHARACTER_INVALID_WEAPON_STYLE" {
		t.Fatalf("expected CHARACTER_INVALID_WEAPON_STYLE, got %q", errorResponse.Error.Code)
	}
}

func TestAuthChallengeRequiredAndInvalid(t *testing.T) {
	server := NewServer(config.API{Port: "8080"})

	var requiredResponse struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/register", map[string]any{
		"bot_name": "bot-missing-challenge",
		"password": "verysecure",
	}, "", http.StatusBadRequest, &requiredResponse)

	if requiredResponse.Error.Code != "AUTH_CHALLENGE_REQUIRED" {
		t.Fatalf("expected AUTH_CHALLENGE_REQUIRED, got %q", requiredResponse.Error.Code)
	}

	challenge := issueAuthChallenge(t, server)

	var invalidResponse struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/register", map[string]any{
		"bot_name":         "bot-invalid-challenge",
		"password":         "verysecure",
		"challenge_id":     challenge.ChallengeID,
		"challenge_answer": "999999",
	}, "", http.StatusUnauthorized, &invalidResponse)

	if invalidResponse.Error.Code != "AUTH_CHALLENGE_INVALID" {
		t.Fatalf("expected AUTH_CHALLENGE_INVALID, got %q", invalidResponse.Error.Code)
	}
}

func TestQuestBoardAndSubmissionFlow(t *testing.T) {
	server := NewServer(config.API{Port: "8080"})

	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/register", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-quester",
		"password": "verysecure",
	}), "", http.StatusOK, nil)

	var loginResponse struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/login", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-quester",
		"password": "verysecure",
	}), "", http.StatusOK, &loginResponse)

	doJSONRequest(t, server, http.MethodPost, "/api/v1/characters", map[string]any{
		"name":         "Courier",
		"class":        "warrior",
		"weapon_style": "sword_shield",
	}, loginResponse.Data.AccessToken, http.StatusOK, nil)

	var questsResponse struct {
		Data struct {
			BoardID string `json:"board_id"`
			Quests  []struct {
				QuestID        string `json:"quest_id"`
				TemplateType   string `json:"template_type"`
				TargetRegionID string `json:"target_region_id"`
				Status         string `json:"status"`
			} `json:"quests"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/quests", nil, loginResponse.Data.AccessToken, http.StatusOK, &questsResponse)

	if questsResponse.Data.BoardID == "" {
		t.Fatal("expected board_id to be populated")
	}
	if len(questsResponse.Data.Quests) != 6 {
		t.Fatalf("expected 6 quests, got %d", len(questsResponse.Data.Quests))
	}

	deliveryQuestID := ""
	for _, quest := range questsResponse.Data.Quests {
		if quest.TemplateType == "deliver_supplies" && quest.TargetRegionID == "greenfield_village" {
			deliveryQuestID = quest.QuestID
			break
		}
	}
	if deliveryQuestID == "" {
		t.Fatal("expected a deliver_supplies quest for greenfield_village")
	}

	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/quests/"+deliveryQuestID+"/accept", nil, loginResponse.Data.AccessToken, http.StatusOK, nil)

	var acceptedState struct {
		Data struct {
			Objectives []struct {
				QuestID string `json:"quest_id"`
				Status  string `json:"status"`
			} `json:"objectives"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/state", nil, loginResponse.Data.AccessToken, http.StatusOK, &acceptedState)

	if len(acceptedState.Data.Objectives) == 0 || acceptedState.Data.Objectives[0].QuestID != deliveryQuestID {
		t.Fatal("expected accepted delivery quest to appear in objectives")
	}

	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/travel", map[string]any{
		"region_id": "greenfield_village",
	}, loginResponse.Data.AccessToken, http.StatusOK, nil)

	var completedBoard struct {
		Data struct {
			Quests []struct {
				QuestID string `json:"quest_id"`
				Status  string `json:"status"`
			} `json:"quests"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/quests", nil, loginResponse.Data.AccessToken, http.StatusOK, &completedBoard)

	foundCompleted := false
	for _, quest := range completedBoard.Data.Quests {
		if quest.QuestID == deliveryQuestID && quest.Status == "completed" {
			foundCompleted = true
		}
	}
	if !foundCompleted {
		t.Fatal("expected delivery quest to become completed after travel")
	}

	var submitResponse struct {
		Data struct {
			State struct {
				Character struct {
					Reputation int    `json:"reputation"`
					Gold       int    `json:"gold"`
					Rank       string `json:"rank"`
				} `json:"character"`
				Limits struct {
					QuestCompletionUsed int `json:"quest_completion_used"`
				} `json:"limits"`
				RecentEvents []struct {
					EventType string `json:"event_type"`
				} `json:"recent_events"`
			} `json:"state"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/quests/"+deliveryQuestID+"/submit", nil, loginResponse.Data.AccessToken, http.StatusOK, &submitResponse)

	if submitResponse.Data.State.Character.Reputation <= 0 {
		t.Fatal("expected quest submission to increase reputation")
	}
	if submitResponse.Data.State.Limits.QuestCompletionUsed != 1 {
		t.Fatalf("expected quest completion used to be 1, got %d", submitResponse.Data.State.Limits.QuestCompletionUsed)
	}
	if len(submitResponse.Data.State.RecentEvents) == 0 || submitResponse.Data.State.RecentEvents[0].EventType != "quest.submitted" {
		t.Fatal("expected quest.submitted to be the latest recent event")
	}

	var rerollResponse struct {
		Data struct {
			RerollCount int `json:"reroll_count"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/quests/reroll", map[string]any{
		"confirm_cost": true,
	}, loginResponse.Data.AccessToken, http.StatusOK, &rerollResponse)

	if rerollResponse.Data.RerollCount != 1 {
		t.Fatalf("expected reroll_count to be 1, got %d", rerollResponse.Data.RerollCount)
	}
}

func TestPublicRoutesReflectRuntimeData(t *testing.T) {
	server := NewServer(config.API{Port: "8080"})

	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/register", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-public",
		"password": "verysecure",
	}), "", http.StatusOK, nil)

	var loginResponse struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/login", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-public",
		"password": "verysecure",
	}), "", http.StatusOK, &loginResponse)

	doJSONRequest(t, server, http.MethodPost, "/api/v1/characters", map[string]any{
		"name":         "PublicRunner",
		"class":        "priest",
		"weapon_style": "holy_tome",
	}, loginResponse.Data.AccessToken, http.StatusOK, nil)

	var boardResponse struct {
		Data struct {
			Quests []struct {
				QuestID        string `json:"quest_id"`
				TemplateType   string `json:"template_type"`
				TargetRegionID string `json:"target_region_id"`
			} `json:"quests"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/quests", nil, loginResponse.Data.AccessToken, http.StatusOK, &boardResponse)

	deliveryQuestID := ""
	for _, quest := range boardResponse.Data.Quests {
		if quest.TemplateType == "deliver_supplies" && quest.TargetRegionID == "greenfield_village" {
			deliveryQuestID = quest.QuestID
			break
		}
	}
	if deliveryQuestID == "" {
		t.Fatal("expected a greenfield delivery quest for runtime public data test")
	}

	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/quests/"+deliveryQuestID+"/accept", nil, loginResponse.Data.AccessToken, http.StatusOK, nil)
	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/travel", map[string]any{
		"region_id": "greenfield_village",
	}, loginResponse.Data.AccessToken, http.StatusOK, nil)
	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/quests/"+deliveryQuestID+"/submit", nil, loginResponse.Data.AccessToken, http.StatusOK, nil)

	var worldStateResponse struct {
		Data struct {
			ActiveBotCount       int `json:"active_bot_count"`
			QuestsCompletedToday int `json:"quests_completed_today"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/public/world-state", nil, "", http.StatusOK, &worldStateResponse)

	if worldStateResponse.Data.ActiveBotCount != 1 {
		t.Fatalf("expected active bot count 1, got %d", worldStateResponse.Data.ActiveBotCount)
	}
	if worldStateResponse.Data.QuestsCompletedToday != 1 {
		t.Fatalf("expected quests completed today 1, got %d", worldStateResponse.Data.QuestsCompletedToday)
	}

	var eventsResponse struct {
		Data struct {
			Items []struct {
				EventType string `json:"event_type"`
				ActorName string `json:"actor_name"`
			} `json:"items"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/public/events?limit=10", nil, "", http.StatusOK, &eventsResponse)

	if len(eventsResponse.Data.Items) == 0 {
		t.Fatal("expected runtime public events to be present")
	}
	if eventsResponse.Data.Items[0].ActorName != "PublicRunner" {
		t.Fatalf("expected PublicRunner as latest public actor, got %q", eventsResponse.Data.Items[0].ActorName)
	}

	var leaderboardsResponse struct {
		Data struct {
			Reputation []struct {
				Name  string `json:"name"`
				Score int    `json:"score"`
			} `json:"reputation"`
			Gold []struct {
				Name string `json:"name"`
			} `json:"gold"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/public/leaderboards", nil, "", http.StatusOK, &leaderboardsResponse)

	if len(leaderboardsResponse.Data.Reputation) == 0 || leaderboardsResponse.Data.Reputation[0].Name != "PublicRunner" {
		t.Fatal("expected runtime reputation leaderboard to include PublicRunner")
	}
	if len(leaderboardsResponse.Data.Gold) == 0 || leaderboardsResponse.Data.Gold[0].Name != "PublicRunner" {
		t.Fatal("expected runtime gold leaderboard to include PublicRunner")
	}
}

func TestDungeonAutoResolveAndClaimFlow(t *testing.T) {
	server := NewServer(config.API{Port: "8080"})

	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/register", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-dungeon",
		"password": "verysecure",
	}), "", http.StatusOK, nil)

	var loginResponse struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/login", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-dungeon",
		"password": "verysecure",
	}), "", http.StatusOK, &loginResponse)

	doJSONRequest(t, server, http.MethodPost, "/api/v1/characters", map[string]any{
		"name":         "DungeonRunner",
		"class":        "warrior",
		"weapon_style": "sword_shield",
	}, loginResponse.Data.AccessToken, http.StatusOK, nil)

	var preClaimState struct {
		Data struct {
			Character struct {
				Gold int `json:"gold"`
			} `json:"character"`
			Limits struct {
				DungeonEntryUsed int `json:"dungeon_entry_used"`
			} `json:"limits"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/state", nil, loginResponse.Data.AccessToken, http.StatusOK, &preClaimState)

	var enterResponse struct {
		Data struct {
			RunID            string `json:"run_id"`
			DungeonID        string `json:"dungeon_id"`
			RunStatus        string `json:"run_status"`
			RuntimePhase     string `json:"runtime_phase"`
			RewardClaimable  bool   `json:"reward_claimable"`
			CurrentRoomIndex int    `json:"current_room_index"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/dungeons/ancient_catacomb_v1/enter", nil, loginResponse.Data.AccessToken, http.StatusOK, &enterResponse)

	if enterResponse.Data.RunID == "" {
		t.Fatal("expected run_id to be returned")
	}
	if enterResponse.Data.DungeonID != "ancient_catacomb_v1" {
		t.Fatalf("expected ancient_catacomb_v1, got %q", enterResponse.Data.DungeonID)
	}
	if enterResponse.Data.RunStatus != "cleared" {
		t.Fatalf("expected run_status cleared, got %q", enterResponse.Data.RunStatus)
	}
	if enterResponse.Data.RuntimePhase != "result_ready" {
		t.Fatalf("expected runtime_phase result_ready, got %q", enterResponse.Data.RuntimePhase)
	}
	if !enterResponse.Data.RewardClaimable {
		t.Fatal("expected reward_claimable true after auto resolve")
	}
	if enterResponse.Data.CurrentRoomIndex != 6 {
		t.Fatalf("expected current_room_index 6, got %d", enterResponse.Data.CurrentRoomIndex)
	}

	var activeRunResponse struct {
		Data any `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/runs/active", nil, loginResponse.Data.AccessToken, http.StatusOK, &activeRunResponse)
	if activeRunResponse.Data != nil {
		t.Fatalf("expected no active run after auto-resolve, got %#v", activeRunResponse.Data)
	}

	var getRunResponse struct {
		Data struct {
			RunID string `json:"run_id"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/runs/"+enterResponse.Data.RunID, nil, loginResponse.Data.AccessToken, http.StatusOK, &getRunResponse)
	if getRunResponse.Data.RunID != enterResponse.Data.RunID {
		t.Fatalf("expected run_id %q, got %q", enterResponse.Data.RunID, getRunResponse.Data.RunID)
	}

	var claimResponse struct {
		Data struct {
			RunID           string  `json:"run_id"`
			RuntimePhase    string  `json:"runtime_phase"`
			RewardClaimable bool    `json:"reward_claimable"`
			RewardClaimedAt *string `json:"reward_claimed_at"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/runs/"+enterResponse.Data.RunID+"/claim", nil, loginResponse.Data.AccessToken, http.StatusOK, &claimResponse)

	if claimResponse.Data.RuntimePhase != "claim_settled" {
		t.Fatalf("expected claim_settled phase, got %q", claimResponse.Data.RuntimePhase)
	}
	if claimResponse.Data.RewardClaimable {
		t.Fatal("expected reward_claimable false after claim")
	}
	if claimResponse.Data.RewardClaimedAt == nil || *claimResponse.Data.RewardClaimedAt == "" {
		t.Fatal("expected reward_claimed_at to be populated")
	}

	var postClaimState struct {
		Data struct {
			Character struct {
				Gold int `json:"gold"`
			} `json:"character"`
			Limits struct {
				DungeonEntryUsed int `json:"dungeon_entry_used"`
			} `json:"limits"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/state", nil, loginResponse.Data.AccessToken, http.StatusOK, &postClaimState)

	if postClaimState.Data.Character.Gold <= preClaimState.Data.Character.Gold {
		t.Fatalf("expected gold to increase after claim, before=%d after=%d", preClaimState.Data.Character.Gold, postClaimState.Data.Character.Gold)
	}
	if postClaimState.Data.Limits.DungeonEntryUsed != preClaimState.Data.Limits.DungeonEntryUsed+1 {
		t.Fatalf("expected dungeon_entry_used increment by 1, before=%d after=%d", preClaimState.Data.Limits.DungeonEntryUsed, postClaimState.Data.Limits.DungeonEntryUsed)
	}

	var secondClaimError struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/runs/"+enterResponse.Data.RunID+"/claim", nil, loginResponse.Data.AccessToken, http.StatusBadRequest, &secondClaimError)
	if secondClaimError.Error.Code != "DUNGEON_REWARD_NOT_CLAIMABLE" {
		t.Fatalf("expected DUNGEON_REWARD_NOT_CLAIMABLE, got %q", secondClaimError.Error.Code)
	}
}

func TestActionAliasClaimRunRewards(t *testing.T) {
	server := NewServer(config.API{Port: "8080"})

	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/register", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-claim-alias",
		"password": "verysecure",
	}), "", http.StatusOK, nil)

	var loginResponse struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/login", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-claim-alias",
		"password": "verysecure",
	}), "", http.StatusOK, &loginResponse)

	doJSONRequest(t, server, http.MethodPost, "/api/v1/characters", map[string]any{
		"name":         "AliasRunner",
		"class":        "warrior",
		"weapon_style": "sword_shield",
	}, loginResponse.Data.AccessToken, http.StatusOK, nil)

	var enterResponse struct {
		Data struct {
			RunID string `json:"run_id"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/dungeons/ancient_catacomb_v1/enter", nil, loginResponse.Data.AccessToken, http.StatusOK, &enterResponse)

	if enterResponse.Data.RunID == "" {
		t.Fatal("expected run_id from dungeon enter")
	}

	var actionResponse struct {
		Data struct {
			ActionResult struct {
				ActionType string `json:"action_type"`
				RunID      string `json:"run_id"`
			} `json:"action_result"`
			State struct {
				Run struct {
					RuntimePhase string `json:"runtime_phase"`
				} `json:"run"`
			} `json:"state"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/actions", map[string]any{
		"action_type": "claim_run_rewards",
		"action_args": map[string]any{
			"run_id": enterResponse.Data.RunID,
		},
	}, loginResponse.Data.AccessToken, http.StatusOK, &actionResponse)

	if actionResponse.Data.ActionResult.ActionType != "claim_dungeon_rewards" {
		t.Fatalf("expected normalized action_type claim_dungeon_rewards, got %q", actionResponse.Data.ActionResult.ActionType)
	}
	if actionResponse.Data.ActionResult.RunID != enterResponse.Data.RunID {
		t.Fatalf("expected run_id %q, got %q", enterResponse.Data.RunID, actionResponse.Data.ActionResult.RunID)
	}
	if actionResponse.Data.State.Run.RuntimePhase != "claim_settled" {
		t.Fatalf("expected claim_settled runtime phase, got %q", actionResponse.Data.State.Run.RuntimePhase)
	}
}

func TestInventoryArenaAndPublicBotsRoutes(t *testing.T) {
	server := NewServer(config.API{Port: "8080"})

	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/register", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-routing",
		"password": "verysecure",
	}), "", http.StatusOK, nil)

	var loginResponse struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/login", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-routing",
		"password": "verysecure",
	}), "", http.StatusOK, &loginResponse)

	doJSONRequest(t, server, http.MethodPost, "/api/v1/characters", map[string]any{
		"name":         "RouteRunner",
		"class":        "mage",
		"weapon_style": "staff",
	}, loginResponse.Data.AccessToken, http.StatusOK, nil)

	var inventoryResponse struct {
		Data struct {
			EquipmentScore int `json:"equipment_score"`
			Equipped       []struct {
				ItemID string `json:"item_id"`
				Slot   string `json:"slot"`
			} `json:"equipped"`
			Inventory []struct {
				ItemID string `json:"item_id"`
				Slot   string `json:"slot"`
			} `json:"inventory"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/inventory", nil, loginResponse.Data.AccessToken, http.StatusOK, &inventoryResponse)
	if len(inventoryResponse.Data.Equipped) == 0 {
		t.Fatal("expected starter equipped item")
	}
	if len(inventoryResponse.Data.Inventory) == 0 {
		t.Fatal("expected inventory items")
	}

	var equipResponse struct {
		Data struct {
			Equipped []struct {
				ItemID string `json:"item_id"`
				Slot   string `json:"slot"`
			} `json:"equipped"`
		} `json:"data"`
	}
	var targetItemID string
	for _, item := range inventoryResponse.Data.Inventory {
		if item.Slot == "weapon" {
			targetItemID = item.ItemID
			break
		}
	}
	if targetItemID == "" {
		targetItemID = inventoryResponse.Data.Inventory[0].ItemID
	}

	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/equipment/equip", map[string]any{
		"item_id": targetItemID,
	}, loginResponse.Data.AccessToken, http.StatusOK, &equipResponse)
	if len(equipResponse.Data.Equipped) == 0 {
		t.Fatal("expected equipped items after equip")
	}

	doJSONRequest(t, server, http.MethodGet, "/api/v1/buildings/guild_main_city", nil, "", http.StatusOK, nil)

	var arenaCurrent struct {
		Data struct {
			TournamentID string `json:"tournament_id"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/arena/current", nil, "", http.StatusOK, &arenaCurrent)
	if arenaCurrent.Data.TournamentID == "" {
		t.Fatal("expected tournament_id in arena current")
	}

	var botsResponse struct {
		Data struct {
			Items []struct {
				CharacterSummary struct {
					CharacterID string `json:"character_id"`
					Name        string `json:"name"`
				} `json:"character_summary"`
			} `json:"items"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/public/bots", nil, "", http.StatusOK, &botsResponse)
	if len(botsResponse.Data.Items) == 0 {
		t.Fatal("expected public bots items")
	}

	botID := botsResponse.Data.Items[0].CharacterSummary.CharacterID
	if botID == "" {
		t.Fatal("expected public bot character_id")
	}

	var botDetail struct {
		Data struct {
			CharacterSummary struct {
				CharacterID string `json:"character_id"`
			} `json:"character_summary"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/public/bots/"+botID, nil, "", http.StatusOK, &botDetail)
	if botDetail.Data.CharacterSummary.CharacterID != botID {
		t.Fatalf("expected bot detail %q, got %q", botID, botDetail.Data.CharacterSummary.CharacterID)
	}
}

func TestBuildingActionsApplyEconomyEffects(t *testing.T) {
	server := NewServer(config.API{Port: "8080"})

	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/register", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-building",
		"password": "verysecure",
	}), "", http.StatusOK, nil)

	var loginResponse struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/login", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-building",
		"password": "verysecure",
	}), "", http.StatusOK, &loginResponse)

	doJSONRequest(t, server, http.MethodPost, "/api/v1/characters", map[string]any{
		"name":         "BuilderOne",
		"class":        "warrior",
		"weapon_style": "sword_shield",
	}, loginResponse.Data.AccessToken, http.StatusOK, nil)

	var shopInventory struct {
		Data struct {
			Items []struct {
				CatalogID string `json:"catalog_id"`
				PriceGold int    `json:"price_gold"`
			} `json:"items"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/buildings/weapon_shop_main_city/shop-inventory", nil, loginResponse.Data.AccessToken, http.StatusOK, &shopInventory)
	if len(shopInventory.Data.Items) == 0 {
		t.Fatal("expected non-empty weapon shop inventory")
	}

	purchaseCatalogID := shopInventory.Data.Items[0].CatalogID
	purchasePrice := shopInventory.Data.Items[0].PriceGold

	var purchaseResponse struct {
		Data struct {
			Result struct {
				PriceGold int `json:"price_gold"`
				Item      struct {
					ItemID string `json:"item_id"`
				} `json:"item"`
			} `json:"result"`
			State struct {
				Character struct {
					Gold int `json:"gold"`
				} `json:"character"`
			} `json:"state"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/buildings/weapon_shop_main_city/purchase", map[string]any{
		"catalog_id": purchaseCatalogID,
	}, loginResponse.Data.AccessToken, http.StatusOK, &purchaseResponse)

	if purchaseResponse.Data.Result.Item.ItemID == "" {
		t.Fatal("expected purchased item_id")
	}
	if purchaseResponse.Data.Result.PriceGold != purchasePrice {
		t.Fatalf("expected purchase price %d, got %d", purchasePrice, purchaseResponse.Data.Result.PriceGold)
	}
	if purchaseResponse.Data.State.Character.Gold != 100-purchasePrice {
		t.Fatalf("expected gold %d after purchase, got %d", 100-purchasePrice, purchaseResponse.Data.State.Character.Gold)
	}

	var sellResponse struct {
		Data struct {
			Result struct {
				GainGold int `json:"gain_gold"`
			} `json:"result"`
			State struct {
				Character struct {
					Gold int `json:"gold"`
				} `json:"character"`
			} `json:"state"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/buildings/weapon_shop_main_city/sell", map[string]any{
		"item_id": purchaseResponse.Data.Result.Item.ItemID,
	}, loginResponse.Data.AccessToken, http.StatusOK, &sellResponse)
	if sellResponse.Data.Result.GainGold <= 0 {
		t.Fatalf("expected gain_gold > 0, got %d", sellResponse.Data.Result.GainGold)
	}
	if sellResponse.Data.State.Character.Gold <= purchaseResponse.Data.State.Character.Gold {
		t.Fatal("expected gold to increase after sell")
	}

	var healResponse struct {
		Data struct {
			ActionResult struct {
				Status string `json:"status"`
			} `json:"action_result"`
			State struct {
				Character struct {
					Gold int `json:"gold"`
				} `json:"character"`
			} `json:"state"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/buildings/temple_main_city/heal", nil, loginResponse.Data.AccessToken, http.StatusOK, &healResponse)
	if healResponse.Data.ActionResult.Status != "success" {
		t.Fatalf("expected success status, got %q", healResponse.Data.ActionResult.Status)
	}

	var sellEquippedErr struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}

	var inventoryView struct {
		Data struct {
			Equipped []struct {
				ItemID string `json:"item_id"`
			} `json:"equipped"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/inventory", nil, loginResponse.Data.AccessToken, http.StatusOK, &inventoryView)
	if len(inventoryView.Data.Equipped) == 0 {
		t.Fatal("expected at least one equipped item")
	}

	doJSONRequest(t, server, http.MethodPost, "/api/v1/buildings/weapon_shop_main_city/sell", map[string]any{
		"item_id": inventoryView.Data.Equipped[0].ItemID,
	}, loginResponse.Data.AccessToken, http.StatusBadRequest, &sellEquippedErr)
	if sellEquippedErr.Error.Code != "INVALID_ACTION_STATE" {
		t.Fatalf("expected INVALID_ACTION_STATE, got %q", sellEquippedErr.Error.Code)
	}
}

func decodeJSON(t *testing.T, recorder *httptest.ResponseRecorder, target any) {
	t.Helper()

	if err := json.Unmarshal(recorder.Body.Bytes(), target); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
}

func doJSONRequest(t *testing.T, server *Server, method, path string, body any, bearerToken string, expectedStatus int, target any) {
	t.Helper()

	var payload []byte
	var err error
	if body != nil {
		payload, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}
	}

	request := httptest.NewRequest(method, path, bytes.NewReader(payload))
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if bearerToken != "" {
		request.Header.Set("Authorization", "Bearer "+bearerToken)
	}

	recorder := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(recorder, request)

	if recorder.Code != expectedStatus {
		t.Fatalf("expected %d for %s %s, got %d with body %s", expectedStatus, method, path, recorder.Code, recorder.Body.String())
	}

	if target != nil {
		decodeJSON(t, recorder, target)
	}
}

func withAuthChallenge(t *testing.T, server *Server, payload map[string]any) map[string]any {
	t.Helper()

	challenge := issueAuthChallenge(t, server)
	enriched := make(map[string]any, len(payload)+2)
	for key, value := range payload {
		enriched[key] = value
	}
	enriched["challenge_id"] = challenge.ChallengeID
	enriched["challenge_answer"] = solveChallengePrompt(t, challenge.PromptText)
	return enriched
}

func issueAuthChallenge(t *testing.T, server *Server) struct {
	ChallengeID string `json:"challenge_id"`
	PromptText  string `json:"prompt_text"`
} {
	t.Helper()

	var challengeResponse struct {
		Data struct {
			Challenge struct {
				ChallengeID string `json:"challenge_id"`
				PromptText  string `json:"prompt_text"`
			} `json:"challenge"`
		} `json:"data"`
	}

	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/challenge", map[string]any{}, "", http.StatusOK, &challengeResponse)
	return challengeResponse.Data.Challenge
}

func solveChallengePrompt(t *testing.T, prompt string) string {
	t.Helper()

	matcher := regexp.MustCompile(`ember=(\d+).+frost=(\d+).+moss=(\d+).+factor=(\d+)`)
	matches := matcher.FindStringSubmatch(prompt)
	if len(matches) != 5 {
		t.Fatalf("unexpected challenge prompt format: %q", prompt)
	}

	ember, _ := strconv.Atoi(matches[1])
	frost, _ := strconv.Atoi(matches[2])
	moss, _ := strconv.Atoi(matches[3])
	factor, _ := strconv.Atoi(matches[4])

	return strconv.Itoa(((ember + frost) - moss) * factor)
}
