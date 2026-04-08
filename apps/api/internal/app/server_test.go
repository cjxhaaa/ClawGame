package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"clawgame/apps/api/internal/characters"
	"clawgame/apps/api/internal/dungeons"
	"clawgame/apps/api/internal/inventory"
	"clawgame/apps/api/internal/platform/config"
	"clawgame/apps/api/internal/world"
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
		if len(payload.Data.Regions) != 10 {
			t.Fatalf("expected 10 regions, got %d", len(payload.Data.Regions))
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
				InteractionLayer       string   `json:"interaction_layer"`
				RiskLevel              string   `json:"risk_level"`
				FacilityFocus          string   `json:"facility_focus"`
				EncounterFamily        string   `json:"encounter_family"`
				CurioStatus            string   `json:"curio_status"`
				CurioHint              string   `json:"curio_hint"`
				HostileEncounters      bool     `json:"hostile_encounters"`
				AvailableRegionActions []string `json:"available_region_actions"`
				Buildings              []struct {
					BuildingID string `json:"building_id"`
				} `json:"buildings"`
			} `json:"data"`
		}
		decodeJSON(t, recorder, &payload)

		if payload.Data.Region.RegionID != "main_city" {
			t.Fatalf("expected main_city, got %q", payload.Data.Region.RegionID)
		}
		if len(payload.Data.Buildings) != 6 {
			t.Fatalf("expected 6 buildings in main city, got %d", len(payload.Data.Buildings))
		}
		if payload.Data.InteractionLayer != "safe_hub" {
			t.Fatalf("expected main_city interaction layer safe_hub, got %q", payload.Data.InteractionLayer)
		}
		if payload.Data.RiskLevel != "low" {
			t.Fatalf("expected main_city risk level low, got %q", payload.Data.RiskLevel)
		}
		if payload.Data.FacilityFocus != "guild_services" {
			t.Fatalf("expected main_city facility focus guild_services, got %q", payload.Data.FacilityFocus)
		}
		if payload.Data.EncounterFamily != "non_combat_hub" {
			t.Fatalf("expected main_city encounter family non_combat_hub, got %q", payload.Data.EncounterFamily)
		}
		if payload.Data.CurioStatus != "dormant" {
			t.Fatalf("expected main_city curio status dormant, got %q", payload.Data.CurioStatus)
		}
		if payload.Data.CurioHint == "" {
			t.Fatal("expected main_city curio hint to be populated")
		}
		if payload.Data.HostileEncounters {
			t.Fatal("expected main_city hostile encounters to be false")
		}
		if len(payload.Data.AvailableRegionActions) == 0 || payload.Data.AvailableRegionActions[0] != "enter_building" {
			t.Fatalf("expected main_city available_region_actions to start with enter_building, got %#v", payload.Data.AvailableRegionActions)
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

func TestPublicWorldStateIncludesRegionGameplay(t *testing.T) {
	server := NewServer(config.API{Port: "8080"})

	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/register", withAuthChallenge(t, server, map[string]any{
		"bot_name": "map-bot",
		"password": "verysecure",
	}), "", http.StatusOK, nil)

	var loginResponse struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/login", withAuthChallenge(t, server, map[string]any{
		"bot_name": "map-bot",
		"password": "verysecure",
	}), "", http.StatusOK, &loginResponse)

	doJSONRequest(t, server, http.MethodPost, "/api/v1/characters", map[string]any{
		"name":         "MapRunner",
		"class":        "warrior",
		"weapon_style": "sword_shield",
	}, loginResponse.Data.AccessToken, http.StatusOK, nil)

	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/travel", map[string]any{
		"region_id": "whispering_forest",
	}, loginResponse.Data.AccessToken, http.StatusOK, nil)

	var worldStateResponse struct {
		Data struct {
			Regions []struct {
				RegionID               string   `json:"region_id"`
				InteractionLayer       string   `json:"interaction_layer"`
				RiskLevel              string   `json:"risk_level"`
				FacilityFocus          string   `json:"facility_focus"`
				EncounterFamily        string   `json:"encounter_family"`
				CurioStatus            string   `json:"curio_status"`
				CurioHint              string   `json:"curio_hint"`
				LinkedDungeon          string   `json:"linked_dungeon"`
				HostileEncounters      bool     `json:"hostile_encounters"`
				AvailableRegionActions []string `json:"available_region_actions"`
			} `json:"regions"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/public/world-state", nil, "", http.StatusOK, &worldStateResponse)

	found := false
	for _, region := range worldStateResponse.Data.Regions {
		if region.RegionID != "whispering_forest" {
			continue
		}
		found = true
		if region.InteractionLayer != "field" {
			t.Fatalf("expected whispering_forest interaction layer field, got %q", region.InteractionLayer)
		}
		if region.RiskLevel != "low" {
			t.Fatalf("expected whispering_forest risk level low, got %q", region.RiskLevel)
		}
		if region.FacilityFocus != "hunt_camp" {
			t.Fatalf("expected whispering_forest facility focus hunt_camp, got %q", region.FacilityFocus)
		}
		if region.EncounterFamily != "forest_hunt" {
			t.Fatalf("expected whispering_forest encounter family forest_hunt, got %q", region.EncounterFamily)
		}
		if region.CurioStatus != "active" {
			t.Fatalf("expected whispering_forest curio status active after travel, got %q", region.CurioStatus)
		}
		if region.LinkedDungeon != "ancient_catacomb" {
			t.Fatalf("expected whispering_forest linked dungeon ancient_catacomb, got %q", region.LinkedDungeon)
		}
		if region.CurioHint == "" {
			t.Fatal("expected whispering_forest curio hint to be populated")
		}
		if !region.HostileEncounters {
			t.Fatal("expected whispering_forest hostile encounters to be true")
		}
		expectedActions := []string{
			"resolve_field_encounter:hunt",
			"resolve_field_encounter:gather",
			"resolve_field_encounter:curio",
			"enter_dungeon",
		}
		for _, expectedAction := range expectedActions {
			foundAction := false
			for _, action := range region.AvailableRegionActions {
				if action == expectedAction {
					foundAction = true
					break
				}
			}
			if !foundAction {
				t.Fatalf("expected whispering_forest available_region_actions to include %q, got %#v", expectedAction, region.AvailableRegionActions)
			}
		}
	}

	if !found {
		t.Fatal("expected whispering_forest to be present in public world state")
	}
}

func TestArenaRatingBoardAndChallengeEndpoints(t *testing.T) {
	server := NewServer(config.API{Port: "8080"})
	now := time.Date(2026, time.April, 6, 10, 0, 0, 0, time.FixedZone("CST", 8*60*60))
	server.arenaService.SetClock(func() time.Time { return now })
	server.worldService.SetClock(func() time.Time { return now })

	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/register", withAuthChallenge(t, server, map[string]any{
		"bot_name": "arena-a",
		"password": "verysecure",
	}), "", http.StatusOK, nil)
	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/register", withAuthChallenge(t, server, map[string]any{
		"bot_name": "arena-b",
		"password": "verysecure",
	}), "", http.StatusOK, nil)

	var loginA struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/login", withAuthChallenge(t, server, map[string]any{
		"bot_name": "arena-a",
		"password": "verysecure",
	}), "", http.StatusOK, &loginA)

	var loginB struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/login", withAuthChallenge(t, server, map[string]any{
		"bot_name": "arena-b",
		"password": "verysecure",
	}), "", http.StatusOK, &loginB)

	doJSONRequest(t, server, http.MethodPost, "/api/v1/characters", map[string]any{
		"name":         "ArenaAlpha",
		"class":        "warrior",
		"weapon_style": "great_axe",
	}, loginA.Data.AccessToken, http.StatusOK, nil)
	doJSONRequest(t, server, http.MethodPost, "/api/v1/characters", map[string]any{
		"name":         "ArenaBeta",
		"class":        "mage",
		"weapon_style": "spellbook",
	}, loginB.Data.AccessToken, http.StatusOK, nil)

	var boardResponse struct {
		Data struct {
			WeekKey               string `json:"week_key"`
			FreeAttemptsRemaining int    `json:"free_attempts_remaining"`
			Candidates            []struct {
				CharacterID string `json:"character_id"`
			} `json:"candidates"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/arena/rating-board", nil, loginA.Data.AccessToken, http.StatusOK, &boardResponse)
	if boardResponse.Data.WeekKey == "" {
		t.Fatal("expected week_key in rating board")
	}
	if boardResponse.Data.FreeAttemptsRemaining != 3 {
		t.Fatalf("expected 3 free attempts, got %d", boardResponse.Data.FreeAttemptsRemaining)
	}
	if len(boardResponse.Data.Candidates) == 0 {
		t.Fatal("expected at least one challenge candidate")
	}

	var challengeResponse struct {
		Data struct {
			Result string `json:"result"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/arena/rating-challenges", map[string]any{
		"target_character_id": boardResponse.Data.Candidates[0].CharacterID,
	}, loginA.Data.AccessToken, http.StatusOK, &challengeResponse)
	if challengeResponse.Data.Result == "" {
		t.Fatal("expected challenge result")
	}

	var purchasedResponse struct {
		Data struct {
			PriceGold int `json:"price_gold"`
			Board     struct {
				PurchasedAttemptsUsed int `json:"purchased_attempts_used"`
			} `json:"board"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/arena/rating-challenges/purchase", map[string]any{}, loginA.Data.AccessToken, http.StatusOK, &purchasedResponse)
	if purchasedResponse.Data.PriceGold <= 0 {
		t.Fatalf("expected positive purchase price, got %d", purchasedResponse.Data.PriceGold)
	}
	if purchasedResponse.Data.Board.PurchasedAttemptsUsed != 1 {
		t.Fatalf("expected one purchased attempt recorded, got %d", purchasedResponse.Data.Board.PurchasedAttemptsUsed)
	}
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
			CombatPower struct {
				PanelPowerScore int `json:"panel_power_score"`
			} `json:"combat_power"`
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
	if stateResponse.Data.CombatPower.PanelPowerScore <= 0 {
		t.Fatal("expected combat_power.panel_power_score in /me/state")
	}
	if len(stateResponse.Data.Objectives) != characters.DailyQuestBoardSize {
		t.Fatalf("expected %d active objectives on first login, got %d", characters.DailyQuestBoardSize, len(stateResponse.Data.Objectives))
	}

	var rawStateResponse map[string]any
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/state", nil, loginResponse.Data.AccessToken, http.StatusOK, &rawStateResponse)

	data, ok := rawStateResponse["data"].(map[string]any)
	if !ok {
		t.Fatal("expected data object in /me/state response")
	}
	stats, ok := data["stats"].(map[string]any)
	if !ok {
		t.Fatal("expected stats object in /me/state response")
	}
	if _, exists := stats["max_mp"]; exists {
		t.Fatal("expected /me/state stats to omit max_mp")
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

	if errorResponse.Error.Code != "INVALID_REQUEST" {
		t.Fatalf("expected INVALID_REQUEST, got %q", errorResponse.Error.Code)
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

func TestCivilianSkillsStartReadyAndCanUpgradeAndLoadout(t *testing.T) {
	server := NewServer(config.API{Port: "8080"})

	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/register", withAuthChallenge(t, server, map[string]any{
		"bot_name": "skill-civilian",
		"password": "verysecure",
	}), "", http.StatusOK, nil)

	var loginResponse struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/login", withAuthChallenge(t, server, map[string]any{
		"bot_name": "skill-civilian",
		"password": "verysecure",
	}), "", http.StatusOK, &loginResponse)

	var createResponse struct {
		Data struct {
			Character struct {
				Class string `json:"class"`
				Gold  int    `json:"gold"`
			} `json:"character"`
			Skills struct {
				BasicAttack struct {
					SkillID string `json:"skill_id"`
				} `json:"basic_attack"`
				CivilianSkills []struct {
					SkillID    string `json:"skill_id"`
					IsUnlocked bool   `json:"is_unlocked"`
					Level      int    `json:"level"`
				} `json:"civilian_skills"`
				ClassCommonSkills []struct {
					SkillID    string `json:"skill_id"`
					IsUnlocked bool   `json:"is_unlocked"`
					Level      int    `json:"level"`
				} `json:"class_common_skills"`
			} `json:"skills"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/characters", map[string]any{
		"name": "SkillStarter",
	}, loginResponse.Data.AccessToken, http.StatusOK, &createResponse)
	account, err := server.authService.Authenticate(loginResponse.Data.AccessToken)
	if err != nil {
		t.Fatalf("failed to authenticate skill test account: %v", err)
	}
	summary, ok := server.characterService.GetCharacterByAccount(account)
	if !ok {
		t.Fatal("expected civilian character to exist")
	}
	if _, err := server.characterService.GrantGold(summary.CharacterID, 200); err != nil {
		t.Fatalf("failed to grant gold for civilian skill test: %v", err)
	}
	createResponse.Data.Character.Gold += 200

	if createResponse.Data.Character.Class != "civilian" {
		t.Fatalf("expected civilian class, got %q", createResponse.Data.Character.Class)
	}
	if createResponse.Data.Skills.BasicAttack.SkillID != "Strike" {
		t.Fatalf("expected civilian basic attack Strike, got %q", createResponse.Data.Skills.BasicAttack.SkillID)
	}

	readyCivilianSkills := 0
	for _, skill := range createResponse.Data.Skills.CivilianSkills {
		if skill.IsUnlocked && skill.Level == 1 {
			readyCivilianSkills++
		}
	}
	if readyCivilianSkills != len(createResponse.Data.Skills.CivilianSkills) {
		t.Fatalf("expected all civilian skills to start unlocked at level 1, got %d ready entries out of %d", readyCivilianSkills, len(createResponse.Data.Skills.CivilianSkills))
	}
	if len(createResponse.Data.Skills.ClassCommonSkills) == 0 {
		t.Fatal("expected civilian character to access profession-common skills")
	}

	var upgradeResponse struct {
		Data struct {
			Character struct {
				Gold int `json:"gold"`
			} `json:"character"`
			Skills struct {
				CivilianSkills []struct {
					SkillID    string `json:"skill_id"`
					IsUnlocked bool   `json:"is_unlocked"`
					Level      int    `json:"level"`
				} `json:"civilian_skills"`
				ClassCommonSkills []struct {
					SkillID    string `json:"skill_id"`
					IsUnlocked bool   `json:"is_unlocked"`
					Level      int    `json:"level"`
				} `json:"class_common_skills"`
			} `json:"skills"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/skills/Quickstep/upgrade", nil, loginResponse.Data.AccessToken, http.StatusOK, &upgradeResponse)

	if upgradeResponse.Data.Character.Gold != createResponse.Data.Character.Gold-180 {
		t.Fatalf("expected gold to drop by 180, got %d -> %d", createResponse.Data.Character.Gold, upgradeResponse.Data.Character.Gold)
	}

	foundQuickstep := false
	for _, skill := range upgradeResponse.Data.Skills.CivilianSkills {
		if skill.SkillID != "Quickstep" {
			continue
		}
		foundQuickstep = true
		if !skill.IsUnlocked || skill.Level != 2 {
			t.Fatalf("expected Quickstep upgraded to level 2, got unlocked=%v level=%d", skill.IsUnlocked, skill.Level)
		}
	}
	if !foundQuickstep {
		t.Fatal("expected Quickstep to be present in civilian skill list")
	}
	foundFocusPulse := false
	for _, skill := range upgradeResponse.Data.Skills.ClassCommonSkills {
		if skill.SkillID != "Focus Pulse" {
			continue
		}
		foundFocusPulse = true
		if !skill.IsUnlocked || skill.Level != 1 {
			t.Fatalf("expected Focus Pulse ready for civilian loadouts, got unlocked=%v level=%d", skill.IsUnlocked, skill.Level)
		}
	}
	if !foundFocusPulse {
		t.Fatal("expected Focus Pulse to be present in civilian-accessible class-common skills")
	}

	var loadoutResponse struct {
		Data struct {
			ActiveLoadout []string `json:"active_loadout"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/skills/loadout", map[string]any{
		"skill_ids": []string{"Quickstep", "Focus Pulse"},
	}, loginResponse.Data.AccessToken, http.StatusOK, &loadoutResponse)

	if len(loadoutResponse.Data.ActiveLoadout) != 2 || loadoutResponse.Data.ActiveLoadout[0] != "Quickstep" || loadoutResponse.Data.ActiveLoadout[1] != "Focus Pulse" {
		t.Fatalf("expected civilian mixed loadout, got %#v", loadoutResponse.Data.ActiveLoadout)
	}
}

func TestGuildSkillUpgradeForProfessionCharacter(t *testing.T) {
	server := NewServer(config.API{Port: "8080"})

	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/register", withAuthChallenge(t, server, map[string]any{
		"bot_name": "skill-guild",
		"password": "verysecure",
	}), "", http.StatusOK, nil)

	var loginResponse struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/login", withAuthChallenge(t, server, map[string]any{
		"bot_name": "skill-guild",
		"password": "verysecure",
	}), "", http.StatusOK, &loginResponse)

	doJSONRequest(t, server, http.MethodPost, "/api/v1/characters", map[string]any{
		"name":         "GuildCaster",
		"class":        "mage",
		"weapon_style": "staff",
	}, loginResponse.Data.AccessToken, http.StatusOK, nil)
	account, err := server.authService.Authenticate(loginResponse.Data.AccessToken)
	if err != nil {
		t.Fatalf("failed to authenticate guild skill test account: %v", err)
	}
	summary, ok := server.characterService.GetCharacterByAccount(account)
	if !ok {
		t.Fatal("expected guild skill character to exist")
	}
	if _, err := server.characterService.GrantGold(summary.CharacterID, 200); err != nil {
		t.Fatalf("failed to grant gold for guild skill test: %v", err)
	}

	var guildSkills struct {
		Data struct {
			BuildingID string `json:"building_id"`
			Skills     struct {
				ClassCommonSkills []struct {
					SkillID    string `json:"skill_id"`
					IsUnlocked bool   `json:"is_unlocked"`
					Level      int    `json:"level"`
				} `json:"class_common_skills"`
				ClassSkills []struct {
					SkillID    string `json:"skill_id"`
					IsUnlocked bool   `json:"is_unlocked"`
					Level      int    `json:"level"`
				} `json:"class_skills"`
			} `json:"skills"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/buildings/guild_main_city/skills", nil, loginResponse.Data.AccessToken, http.StatusOK, &guildSkills)

	if guildSkills.Data.BuildingID != "guild_main_city" {
		t.Fatalf("expected guild_main_city, got %q", guildSkills.Data.BuildingID)
	}
	if len(guildSkills.Data.Skills.ClassCommonSkills) == 0 {
		t.Fatal("expected mage class-common skills in guild response")
	}

	foundFlameBurst := false
	for _, skill := range guildSkills.Data.Skills.ClassSkills {
		if skill.SkillID != "Flame Burst" {
			continue
		}
		foundFlameBurst = true
		if !skill.IsUnlocked || skill.Level != 1 {
			t.Fatalf("expected Flame Burst ready at level 1 before upgrade, got unlocked=%v level=%d", skill.IsUnlocked, skill.Level)
		}
	}
	if !foundFlameBurst {
		t.Fatal("expected Flame Burst in mage class skill list")
	}

	var upgradeResponse struct {
		Data struct {
			BuildingID string `json:"building_id"`
			Skills     struct {
				ClassCommonSkills []struct {
					SkillID    string `json:"skill_id"`
					IsUnlocked bool   `json:"is_unlocked"`
					Level      int    `json:"level"`
				} `json:"class_common_skills"`
				ClassSkills []struct {
					SkillID    string `json:"skill_id"`
					IsUnlocked bool   `json:"is_unlocked"`
					Level      int    `json:"level"`
				} `json:"class_skills"`
			} `json:"skills"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/buildings/guild_main_city/skills/Flame%20Burst/upgrade", nil, loginResponse.Data.AccessToken, http.StatusOK, &upgradeResponse)

	upgraded := false
	for _, skill := range upgradeResponse.Data.Skills.ClassSkills {
		if skill.SkillID != "Flame Burst" {
			continue
		}
		upgraded = true
		if !skill.IsUnlocked || skill.Level != 2 {
			t.Fatalf("expected Flame Burst upgraded to level 2, got unlocked=%v level=%d", skill.IsUnlocked, skill.Level)
		}
	}
	if !upgraded {
		t.Fatal("expected upgraded Flame Burst in guild response")
	}

	var loadoutResponse struct {
		Data struct {
			Skills struct {
				ActiveLoadout []string `json:"active_loadout"`
			} `json:"skills"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/buildings/guild_main_city/skill-loadout", map[string]any{
		"skill_ids": []string{"Flame Burst"},
	}, loginResponse.Data.AccessToken, http.StatusOK, &loadoutResponse)

	if len(loadoutResponse.Data.Skills.ActiveLoadout) != 1 || loadoutResponse.Data.Skills.ActiveLoadout[0] != "Flame Burst" {
		t.Fatalf("expected Flame Burst guild loadout, got %#v", loadoutResponse.Data.Skills.ActiveLoadout)
	}
}

func TestWorldBossQueueAndReforgeFlow(t *testing.T) {
	server := NewServer(config.API{Port: "8080"})

	type accountFixture struct {
		token       string
		characterID string
	}
	fixtures := make([]accountFixture, 0, 6)

	for i := 0; i < 6; i++ {
		botName := "boss-bot-" + strconv.Itoa(i)
		charName := "BossRunner" + strconv.Itoa(i)
		doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/register", withAuthChallenge(t, server, map[string]any{
			"bot_name": botName,
			"password": "verysecure",
		}), "", http.StatusOK, nil)

		var loginResponse struct {
			Data struct {
				AccessToken string `json:"access_token"`
			} `json:"data"`
		}
		doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/login", withAuthChallenge(t, server, map[string]any{
			"bot_name": botName,
			"password": "verysecure",
		}), "", http.StatusOK, &loginResponse)

		var createResponse struct {
			Data struct {
				Character struct {
					CharacterID string `json:"character_id"`
				} `json:"character"`
			} `json:"data"`
		}
		doJSONRequest(t, server, http.MethodPost, "/api/v1/characters", map[string]any{
			"name": charName,
		}, loginResponse.Data.AccessToken, http.StatusOK, &createResponse)

		fixtures = append(fixtures, accountFixture{
			token:       loginResponse.Data.AccessToken,
			characterID: createResponse.Data.Character.CharacterID,
		})
	}

	var queueResponse struct {
		Data struct {
			Status struct {
				CurrentQueuedCount int    `json:"current_queued_count"`
				LastRaidID         string `json:"last_raid_id"`
			} `json:"status"`
			ResolvedRaid *struct {
				RaidID      string `json:"raid_id"`
				RewardTier  string `json:"reward_tier"`
				TotalDamage int    `json:"total_damage"`
				Members     []struct {
					CharacterID string `json:"character_id"`
					DamageDealt int    `json:"damage_dealt"`
				} `json:"members"`
			} `json:"resolved_raid"`
		} `json:"data"`
	}
	for i, fx := range fixtures {
		target := &queueResponse
		if i < len(fixtures)-1 {
			var ignored map[string]any
			doJSONRequest(t, server, http.MethodPost, "/api/v1/world-boss/queue", nil, fx.token, http.StatusOK, &ignored)
			continue
		}
		doJSONRequest(t, server, http.MethodPost, "/api/v1/world-boss/queue", nil, fx.token, http.StatusOK, target)
	}

	if queueResponse.Data.ResolvedRaid == nil {
		t.Fatal("expected sixth world boss queue join to resolve a raid")
	}
	if queueResponse.Data.ResolvedRaid.RaidID == "" {
		t.Fatal("expected resolved raid id")
	}
	if queueResponse.Data.ResolvedRaid.RewardTier == "" {
		t.Fatal("expected resolved raid reward tier")
	}
	if len(queueResponse.Data.ResolvedRaid.Members) != 6 {
		t.Fatalf("expected 6 raid members, got %d", len(queueResponse.Data.ResolvedRaid.Members))
	}

	firstAccount, err := server.authService.Authenticate(fixtures[0].token)
	if err != nil {
		t.Fatalf("failed to authenticate first account: %v", err)
	}
	firstCharacter, ok := server.characterService.GetCharacterByAccount(firstAccount)
	if !ok {
		t.Fatal("expected first character to exist")
	}
	if _, _, err := server.inventoryService.GrantItemFromCatalog(firstCharacter, "gravewake_bastion_chest_red"); err != nil {
		t.Fatalf("failed to grant reforge item: %v", err)
	}

	var inventoryResponse struct {
		Data struct {
			Inventory []struct {
				ItemID       string `json:"item_id"`
				CatalogID    string `json:"catalog_id"`
				ExtraAffixes []struct {
					AffixKey string `json:"affix_key"`
					Value    int    `json:"value"`
				} `json:"extra_affixes"`
			} `json:"inventory"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/inventory", nil, fixtures[0].token, http.StatusOK, &inventoryResponse)

	var targetItemID string
	var originalAffixCount int
	for _, item := range inventoryResponse.Data.Inventory {
		if item.CatalogID == "gravewake_bastion_chest_red" {
			targetItemID = item.ItemID
			originalAffixCount = len(item.ExtraAffixes)
			break
		}
	}
	if targetItemID == "" {
		t.Fatal("expected granted red item in inventory")
	}
	if originalAffixCount != 4 {
		t.Fatalf("expected red item to have 4 extra affixes, got %d", originalAffixCount)
	}

	var stateResponse struct {
		Data struct {
			Materials []struct {
				MaterialKey string `json:"material_key"`
				Quantity    int    `json:"quantity"`
			} `json:"materials"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/state", nil, fixtures[0].token, http.StatusOK, &stateResponse)
	foundStone := false
	totalStones := 0
	for _, material := range stateResponse.Data.Materials {
		if material.MaterialKey == inventory.ReforgeMaterialKey && material.Quantity > 0 {
			foundStone = true
			totalStones = material.Quantity
			break
		}
	}
	if !foundStone {
		t.Fatal("expected world boss reward to grant reforge stone")
	}
	if totalStones < 5 {
		if _, err := server.characterService.GrantMaterials(firstCharacter.CharacterID, []map[string]any{{
			"material_key": inventory.ReforgeMaterialKey,
			"quantity":     5 - totalStones,
		}}); err != nil {
			t.Fatalf("top up reforge stones for red-item test: %v", err)
		}
	}

	var reforgeResponse struct {
		Data struct {
			ReforgeCost struct {
				MaterialKey string `json:"material_key"`
				Quantity    int    `json:"quantity"`
			} `json:"reforge_cost"`
			Item struct {
				ItemID         string `json:"item_id"`
				PendingReforge struct {
					AttemptID        string `json:"attempt_id"`
					MaterialKey      string `json:"material_key"`
					MaterialQuantity int    `json:"material_quantity"`
				} `json:"pending_reforge"`
			} `json:"item"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/items/"+targetItemID+"/reforge", nil, fixtures[0].token, http.StatusOK, &reforgeResponse)
	if reforgeResponse.Data.Item.PendingReforge.AttemptID == "" {
		t.Fatal("expected pending reforge result after using reforge stone")
	}
	if reforgeResponse.Data.ReforgeCost.MaterialKey != inventory.ReforgeMaterialKey || reforgeResponse.Data.ReforgeCost.Quantity != 5 {
		t.Fatalf("expected red item reforge cost to be 5 stones, got %#v", reforgeResponse.Data.ReforgeCost)
	}
	if reforgeResponse.Data.Item.PendingReforge.MaterialQuantity != 5 {
		t.Fatalf("expected pending reforge to record 5-stone cost, got %d", reforgeResponse.Data.Item.PendingReforge.MaterialQuantity)
	}

	var discardResponse struct {
		Data struct {
			Item struct {
				PendingReforge any `json:"pending_reforge"`
			} `json:"item"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/items/"+targetItemID+"/reforge/discard", nil, fixtures[0].token, http.StatusOK, &discardResponse)
	if discardResponse.Data.Item.PendingReforge != nil {
		t.Fatal("expected discard to clear pending reforge result")
	}
}

func TestQuestBoardAndSubmissionFlow(t *testing.T) {
	server := NewServer(config.API{Port: "8080"})
	token, questsResponse := createCharacterWithBoard(t, server, "bot-quester", "Courier", func(quests []boardQuestView) bool {
		for _, quest := range quests {
			if quest.TemplateType == "deliver_supplies" && quest.TargetRegionID == "greenfield_village" {
				return true
			}
		}
		return false
	})
	if len(questsResponse) != 4 {
		t.Fatalf("expected 4 quests, got %d", len(questsResponse))
	}

	deliveryQuestID := ""
	for _, quest := range questsResponse {
		if quest.TemplateType == "deliver_supplies" && quest.TargetRegionID == "greenfield_village" {
			deliveryQuestID = quest.QuestID
			break
		}
	}
	if deliveryQuestID == "" {
		t.Fatal("expected a deliver_supplies quest for greenfield_village")
	}

	var acceptedState struct {
		Data struct {
			Objectives []struct {
				QuestID string `json:"quest_id"`
				Status  string `json:"status"`
			} `json:"objectives"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/state", nil, token, http.StatusOK, &acceptedState)

	foundObjective := false
	for _, objective := range acceptedState.Data.Objectives {
		if objective.QuestID == deliveryQuestID {
			foundObjective = true
			break
		}
	}
	if !foundObjective {
		t.Fatal("expected accepted delivery quest to appear in objectives")
	}

	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/travel", map[string]any{
		"region_id": "greenfield_village",
	}, token, http.StatusOK, nil)

	var completedBoard struct {
		Data struct {
			Quests []struct {
				QuestID string `json:"quest_id"`
				Status  string `json:"status"`
			} `json:"quests"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/quests", nil, token, http.StatusOK, &completedBoard)

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
					Reputation int `json:"reputation"`
					Gold       int `json:"gold"`
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
	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/quests/"+deliveryQuestID+"/submit", nil, token, http.StatusOK, &submitResponse)

	if submitResponse.Data.State.Character.Reputation <= 0 {
		t.Fatal("expected quest submission to increase reputation")
	}
	if submitResponse.Data.State.Limits.QuestCompletionUsed != 1 {
		t.Fatalf("expected quest completion used to be 1, got %d", submitResponse.Data.State.Limits.QuestCompletionUsed)
	}
	if len(submitResponse.Data.State.RecentEvents) == 0 || submitResponse.Data.State.RecentEvents[0].EventType != "quest.submitted" {
		t.Fatal("expected quest.submitted to be the latest recent event")
	}

}

func TestQuestDetailExposesRuntimeFramework(t *testing.T) {
	server := NewServer(config.API{Port: "8080"})
	token, boardResponse := createCharacterWithBoard(t, server, "bot-quest-detail", "QuestDetailer", func(quests []boardQuestView) bool {
		for _, quest := range quests {
			if quest.TemplateType == "deliver_supplies" {
				return true
			}
		}
		return false
	})

	var targetQuestID string
	var targetDifficulty string
	for _, quest := range boardResponse {
		if quest.TemplateType != "deliver_supplies" {
			continue
		}
		targetQuestID = quest.QuestID
		targetDifficulty = quest.Difficulty
		if quest.Difficulty == "" {
			t.Fatal("expected board quest difficulty to be populated")
		}
		if quest.FlowKind != "delivery" {
			t.Fatalf("expected deliver_supplies flow_kind delivery, got %q", quest.FlowKind)
		}
		break
	}
	if targetQuestID == "" {
		t.Fatal("expected to find a deliver_supplies quest on the board")
	}

	var questDetail struct {
		Data struct {
			Quest struct {
				QuestID    string `json:"quest_id"`
				Difficulty string `json:"difficulty"`
				FlowKind   string `json:"flow_kind"`
			} `json:"quest"`
			Runtime struct {
				QuestID             string         `json:"quest_id"`
				CurrentStepKey      string         `json:"current_step_key"`
				CurrentStepLabel    string         `json:"current_step_label"`
				CurrentStepHint     string         `json:"current_step_hint"`
				SuggestedActionType string         `json:"suggested_action_type"`
				CompletedStepKeys   []string       `json:"completed_step_keys"`
				AvailableChoices    []any          `json:"available_choices"`
				Clues               []any          `json:"clues"`
				State               map[string]any `json:"state_json"`
			} `json:"runtime"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/quests/"+targetQuestID, nil, token, http.StatusOK, &questDetail)

	if questDetail.Data.Quest.QuestID != targetQuestID {
		t.Fatalf("expected quest detail %q, got %q", targetQuestID, questDetail.Data.Quest.QuestID)
	}
	if questDetail.Data.Quest.Difficulty == "" {
		t.Fatal("expected quest detail difficulty to be populated")
	}
	if questDetail.Data.Quest.FlowKind != "delivery" {
		t.Fatalf("expected detail flow_kind delivery, got %q", questDetail.Data.Quest.FlowKind)
	}
	if questDetail.Data.Runtime.QuestID != targetQuestID {
		t.Fatalf("expected runtime quest id %q, got %q", targetQuestID, questDetail.Data.Runtime.QuestID)
	}
	expectedStepKey := "reach_target_region"
	if targetDifficulty == "hard" {
		expectedStepKey = "confirm_route"
	}
	if questDetail.Data.Runtime.CurrentStepKey != expectedStepKey {
		t.Fatalf("expected delivery quest current_step_key %s, got %q", expectedStepKey, questDetail.Data.Runtime.CurrentStepKey)
	}
	if questDetail.Data.Runtime.CurrentStepLabel == "" || questDetail.Data.Runtime.CurrentStepHint == "" {
		t.Fatal("expected quest runtime step label and hint to be populated")
	}
	if targetDifficulty == "hard" {
		if questDetail.Data.Runtime.SuggestedActionType != "quest_interact" {
			t.Fatalf("expected hard delivery to suggest quest_interact, got %q", questDetail.Data.Runtime.SuggestedActionType)
		}
	} else if questDetail.Data.Runtime.SuggestedActionType != "" {
		t.Fatalf("expected plain delivery step to not force suggested action, got %q", questDetail.Data.Runtime.SuggestedActionType)
	}
	if questDetail.Data.Runtime.State == nil {
		t.Fatal("expected quest runtime state_json to be populated")
	}
}

func TestQuestRuntimeChoiceAndInteractionEndpoints(t *testing.T) {
	server := NewServer(config.API{Port: "8080"})
	token, boardResponse := createCharacterWithBoard(t, server, "bot-quest-runtime", "RuntimeQuester", func(quests []boardQuestView) bool {
		for _, quest := range quests {
			if quest.Difficulty == "nightmare" {
				return true
			}
		}
		return false
	})

	var nightmareQuestID string
	for _, quest := range boardResponse {
		if quest.Difficulty == "nightmare" {
			nightmareQuestID = quest.QuestID
			break
		}
	}
	if nightmareQuestID == "" {
		t.Fatal("expected a nightmare quest on the daily board")
	}

	var detailBefore struct {
		Data struct {
			Runtime struct {
				CurrentStepKey      string `json:"current_step_key"`
				CurrentStepLabel    string `json:"current_step_label"`
				CurrentStepHint     string `json:"current_step_hint"`
				SuggestedActionType string `json:"suggested_action_type"`
				AvailableChoices    []struct {
					ChoiceKey string `json:"choice_key"`
				} `json:"available_choices"`
			} `json:"runtime"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/quests/"+nightmareQuestID, nil, token, http.StatusOK, &detailBefore)

	if detailBefore.Data.Runtime.CurrentStepKey != "inspect_clue" {
		t.Fatalf("expected nightmare quest to start at inspect_clue, got %q", detailBefore.Data.Runtime.CurrentStepKey)
	}
	if detailBefore.Data.Runtime.CurrentStepLabel == "" || detailBefore.Data.Runtime.CurrentStepHint == "" {
		t.Fatal("expected nightmare runtime step metadata before interaction")
	}
	if detailBefore.Data.Runtime.SuggestedActionType != "quest_interact" {
		t.Fatalf("expected nightmare runtime suggested action quest_interact, got %q", detailBefore.Data.Runtime.SuggestedActionType)
	}
	if len(detailBefore.Data.Runtime.AvailableChoices) != 0 {
		t.Fatal("expected nightmare quest to require clue inspection before exposing choices")
	}

	var interactFirst struct {
		Data struct {
			Runtime struct {
				CurrentStepKey      string `json:"current_step_key"`
				CurrentStepLabel    string `json:"current_step_label"`
				CurrentStepHint     string `json:"current_step_hint"`
				SuggestedActionType string `json:"suggested_action_type"`
				AvailableChoices    []struct {
					ChoiceKey string `json:"choice_key"`
				} `json:"available_choices"`
				CompletedSteps []string `json:"completed_step_keys"`
				Clues          []struct {
					ClueKey string `json:"clue_key"`
				} `json:"clues"`
				State map[string]any `json:"state_json"`
			} `json:"runtime"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/quests/"+nightmareQuestID+"/interact", map[string]any{
		"interaction": "inspect_clue",
	}, token, http.StatusOK, &interactFirst)

	if interactFirst.Data.Runtime.CurrentStepKey != "submit_choice" {
		t.Fatalf("expected inspect_clue to advance nightmare runtime to submit_choice, got %q", interactFirst.Data.Runtime.CurrentStepKey)
	}
	if interactFirst.Data.Runtime.SuggestedActionType != "quest_choice" {
		t.Fatalf("expected submit_choice step to suggest quest_choice, got %q", interactFirst.Data.Runtime.SuggestedActionType)
	}
	if len(interactFirst.Data.Runtime.AvailableChoices) == 0 {
		t.Fatal("expected clue inspection to expose nightmare quest choices")
	}

	var choiceResponse struct {
		Data struct {
			Runtime struct {
				CurrentStepKey      string         `json:"current_step_key"`
				CurrentStepLabel    string         `json:"current_step_label"`
				CurrentStepHint     string         `json:"current_step_hint"`
				SuggestedActionType string         `json:"suggested_action_type"`
				AvailableChoices    []any          `json:"available_choices"`
				CompletedSteps      []string       `json:"completed_step_keys"`
				State               map[string]any `json:"state_json"`
				Clues               []struct {
					ClueKey string `json:"clue_key"`
				} `json:"clues"`
			} `json:"runtime"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/quests/"+nightmareQuestID+"/choice", map[string]any{
		"choice_key": interactFirst.Data.Runtime.AvailableChoices[0].ChoiceKey,
	}, token, http.StatusOK, &choiceResponse)

	if choiceResponse.Data.Runtime.CurrentStepKey != "clear_target_dungeon" {
		t.Fatalf("expected choice to advance nightmare runtime to clear_target_dungeon, got %q", choiceResponse.Data.Runtime.CurrentStepKey)
	}
	if choiceResponse.Data.Runtime.CurrentStepLabel == "" || choiceResponse.Data.Runtime.CurrentStepHint == "" {
		t.Fatal("expected post-choice runtime step metadata to be populated")
	}
	if len(choiceResponse.Data.Runtime.AvailableChoices) != 0 {
		t.Fatal("expected quest choices to be consumed after choice submission")
	}
	if choiceResponse.Data.Runtime.State["selected_choice_key"] == nil {
		t.Fatal("expected selected_choice_key to be recorded in runtime state")
	}
	if choiceResponse.Data.Runtime.State["selected_choice_label"] == nil {
		t.Fatal("expected selected_choice_label to be recorded in runtime state")
	}
	if len(choiceResponse.Data.Runtime.Clues) <= len(interactFirst.Data.Runtime.Clues) {
		t.Fatal("expected choice resolution to append an outcome clue")
	}

	foundStep := false
	for _, step := range interactFirst.Data.Runtime.CompletedSteps {
		if step == "inspect_clue" {
			foundStep = true
			break
		}
	}
	if !foundStep {
		t.Fatal("expected interaction to be recorded as a completed runtime step")
	}
	if len(interactFirst.Data.Runtime.Clues) == 0 {
		t.Fatal("expected interaction to append a runtime clue")
	}
	if interactFirst.Data.Runtime.State["last_interaction"] != "inspect_clue" {
		t.Fatalf("expected last_interaction inspect_clue, got %#v", interactFirst.Data.Runtime.State["last_interaction"])
	}
}

func TestNightmareDungeonQuestRequiresChoiceBeforeCompletion(t *testing.T) {
	server := NewServer(config.API{Port: "8080"})
	token, boardResponse := createCharacterWithBoard(t, server, "bot-nightmare-gate", "NightmareGate", func(quests []boardQuestView) bool {
		for _, quest := range quests {
			if quest.TemplateType == "clear_dungeon" && quest.Difficulty == "nightmare" {
				return true
			}
		}
		return false
	})

	var nightmareQuestID string
	for _, quest := range boardResponse {
		if quest.TemplateType == "clear_dungeon" && quest.Difficulty == "nightmare" {
			nightmareQuestID = quest.QuestID
			break
		}
	}
	if nightmareQuestID == "" {
		t.Fatal("expected nightmare clear_dungeon quest on board")
	}

	doJSONRequest(t, server, http.MethodPost, "/api/v1/dungeons/ancient_catacomb_v1/enter", dungeonEnterPayload(""), token, http.StatusOK, nil)

	var questsAfterFirstRun struct {
		Data struct {
			Quests []struct {
				QuestID string `json:"quest_id"`
				Status  string `json:"status"`
			} `json:"quests"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/quests", nil, token, http.StatusOK, &questsAfterFirstRun)

	for _, quest := range questsAfterFirstRun.Data.Quests {
		if quest.QuestID == nightmareQuestID && quest.Status == "completed" {
			t.Fatal("expected nightmare clear_dungeon quest to stay incomplete before clue inspection and choice")
		}
	}

	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/quests/"+nightmareQuestID+"/interact", map[string]any{
		"interaction": "inspect_clue",
	}, token, http.StatusOK, nil)
	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/quests/"+nightmareQuestID+"/choice", map[string]any{
		"choice_key": "follow_standard_brief",
	}, token, http.StatusOK, nil)
	doJSONRequest(t, server, http.MethodPost, "/api/v1/dungeons/ancient_catacomb_v1/enter", dungeonEnterPayload(""), token, http.StatusOK, nil)

	var questsAfterSecondRun struct {
		Data struct {
			Quests []struct {
				QuestID string `json:"quest_id"`
				Status  string `json:"status"`
			} `json:"quests"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/quests", nil, token, http.StatusOK, &questsAfterSecondRun)

	completed := false
	for _, quest := range questsAfterSecondRun.Data.Quests {
		if quest.QuestID == nightmareQuestID && quest.Status == "completed" {
			completed = true
			break
		}
	}
	if !completed {
		t.Fatal("expected nightmare clear_dungeon quest to complete after inspect_clue, choice, and dungeon clear")
	}
}

func TestHardDeliveryQuestRequiresRouteConfirmation(t *testing.T) {
	server := NewServer(config.API{Port: "8080"})
	token, boardResponse := createCharacterWithBoard(t, server, "bot-hard-delivery", "HardCourier", func(quests []boardQuestView) bool {
		for _, quest := range quests {
			if quest.TemplateType == "deliver_supplies" && quest.Difficulty == "hard" && quest.TargetRegionID == "whispering_forest" {
				return true
			}
		}
		return false
	})

	var hardQuestID string
	for _, quest := range boardResponse {
		if quest.TemplateType == "deliver_supplies" && quest.Difficulty == "hard" && quest.TargetRegionID == "whispering_forest" {
			hardQuestID = quest.QuestID
			break
		}
	}
	if hardQuestID == "" {
		t.Fatal("expected hard deliver_supplies quest targeting whispering_forest")
	}

	var detailBefore struct {
		Data struct {
			Runtime struct {
				CurrentStepKey string `json:"current_step_key"`
			} `json:"runtime"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/quests/"+hardQuestID, nil, token, http.StatusOK, &detailBefore)
	if detailBefore.Data.Runtime.CurrentStepKey != "confirm_route" {
		t.Fatalf("expected hard delivery quest to start at confirm_route, got %q", detailBefore.Data.Runtime.CurrentStepKey)
	}

	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/travel", map[string]any{
		"region_id": "whispering_forest",
	}, token, http.StatusOK, nil)

	var afterFirstTravel struct {
		Data struct {
			Quests []struct {
				QuestID string `json:"quest_id"`
				Status  string `json:"status"`
			} `json:"quests"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/quests", nil, token, http.StatusOK, &afterFirstTravel)
	for _, quest := range afterFirstTravel.Data.Quests {
		if quest.QuestID == hardQuestID && quest.Status == "completed" {
			t.Fatal("expected hard delivery quest to remain incomplete before route confirmation")
		}
	}

	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/quests/"+hardQuestID+"/interact", map[string]any{
		"interaction": "confirm_route",
	}, token, http.StatusOK, nil)
	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/travel", map[string]any{
		"region_id": "main_city",
	}, token, http.StatusOK, nil)
	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/travel", map[string]any{
		"region_id": "whispering_forest",
	}, token, http.StatusOK, nil)

	var afterConfirmedTravel struct {
		Data struct {
			Quests []struct {
				QuestID string `json:"quest_id"`
				Status  string `json:"status"`
			} `json:"quests"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/quests", nil, token, http.StatusOK, &afterConfirmedTravel)

	completed := false
	for _, quest := range afterConfirmedTravel.Data.Quests {
		if quest.QuestID == hardQuestID && quest.Status == "completed" {
			completed = true
			break
		}
	}
	if !completed {
		t.Fatal("expected hard delivery quest to complete after route confirmation and travel")
	}
}

func TestInvestigationQuestFlowAndPlannerHints(t *testing.T) {
	server := NewServer(config.API{Port: "8080"})
	token, boardResponse := createCharacterWithBoard(t, server, "bot-investigation", "Investigator", func(quests []boardQuestView) bool {
		for _, quest := range quests {
			if quest.TemplateType == "investigate_anomaly" && quest.Difficulty == "hard" {
				return true
			}
		}
		return false
	})

	var investigationQuestID string
	for _, quest := range boardResponse {
		if quest.TemplateType == "investigate_anomaly" && quest.Difficulty == "hard" {
			investigationQuestID = quest.QuestID
			break
		}
	}
	if investigationQuestID == "" {
		t.Fatal("expected hard investigate_anomaly quest on daily board")
	}

	var plannerResponse struct {
		Data struct {
			SuggestedActions []string `json:"suggested_actions"`
			RuntimeHints     []struct {
				QuestID             string         `json:"quest_id"`
				CurrentStepKey      string         `json:"current_step_key"`
				CurrentStepLabel    string         `json:"current_step_label"`
				CurrentStepHint     string         `json:"current_step_hint"`
				SuggestedActionType string         `json:"suggested_action_type"`
				SuggestedActionArgs map[string]any `json:"suggested_action_args"`
				TargetRegionID      string         `json:"target_region_id"`
			} `json:"quest_runtime_hints"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/planner?region_id=main_city", nil, token, http.StatusOK, &plannerResponse)

	foundInteractSuggestion := false
	foundInspectHint := false
	for _, action := range plannerResponse.Data.SuggestedActions {
		if action == "quest_interact" {
			foundInteractSuggestion = true
			break
		}
	}
	for _, hint := range plannerResponse.Data.RuntimeHints {
		if hint.QuestID == investigationQuestID && hint.CurrentStepKey == "inspect_clue" {
			if hint.CurrentStepLabel == "" || hint.CurrentStepHint == "" {
				t.Fatal("expected planner runtime hint to expose step label and hint")
			}
			if hint.SuggestedActionType != "quest_interact" {
				t.Fatalf("expected planner inspect step suggested action quest_interact, got %q", hint.SuggestedActionType)
			}
			if hint.SuggestedActionArgs["interaction"] != "inspect_clue" {
				t.Fatalf("expected planner suggested interaction inspect_clue, got %#v", hint.SuggestedActionArgs["interaction"])
			}
			foundInspectHint = true
			break
		}
	}
	if !foundInteractSuggestion || !foundInspectHint {
		t.Fatalf("expected planner to suggest quest_interact and expose inspect_clue runtime hint, got actions=%#v hints=%#v", plannerResponse.Data.SuggestedActions, plannerResponse.Data.RuntimeHints)
	}

	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/actions", map[string]any{
		"action_type": "quest_interact",
		"action_args": map[string]any{
			"quest_id":    investigationQuestID,
			"interaction": "inspect_clue",
		},
	}, token, http.StatusOK, nil)

	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/planner?region_id=main_city", nil, token, http.StatusOK, &plannerResponse)
	foundChoiceSuggestion := false
	foundChoiceHint := false
	for _, action := range plannerResponse.Data.SuggestedActions {
		if action == "quest_choice" {
			foundChoiceSuggestion = true
			break
		}
	}
	for _, hint := range plannerResponse.Data.RuntimeHints {
		if hint.QuestID == investigationQuestID && hint.CurrentStepKey == "submit_choice" {
			if hint.SuggestedActionType != "quest_choice" {
				t.Fatalf("expected planner choice step suggested action quest_choice, got %q", hint.SuggestedActionType)
			}
			foundChoiceHint = true
			break
		}
	}
	if !foundChoiceSuggestion || !foundChoiceHint {
		t.Fatalf("expected planner to suggest quest_choice after clue inspection, got actions=%#v hints=%#v", plannerResponse.Data.SuggestedActions, plannerResponse.Data.RuntimeHints)
	}

	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/actions", map[string]any{
		"action_type": "quest_choice",
		"action_args": map[string]any{
			"quest_id":   investigationQuestID,
			"choice_key": "handoff_to_guild",
		},
	}, token, http.StatusOK, nil)

	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/travel", map[string]any{
		"region_id": "greenfield_village",
	}, token, http.StatusOK, nil)

	var questsAfter struct {
		Data struct {
			Quests []struct {
				QuestID string `json:"quest_id"`
				Status  string `json:"status"`
			} `json:"quests"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/quests", nil, token, http.StatusOK, &questsAfter)

	completed := false
	for _, quest := range questsAfter.Data.Quests {
		if quest.QuestID == investigationQuestID && quest.Status == "completed" {
			completed = true
			break
		}
	}
	if !completed {
		t.Fatal("expected investigation quest to complete after inspect_clue, choice, and delivery travel")
	}
}

func TestQuestRuntimeActionsAppearInActionAndStateViews(t *testing.T) {
	server := NewServer(config.API{Port: "8080"})
	token, boardResponse := createCharacterWithBoard(t, server, "bot-quest-actions", "ActionSeer", func(quests []boardQuestView) bool {
		for _, quest := range quests {
			if quest.TemplateType == "investigate_anomaly" {
				return true
			}
		}
		return false
	})

	var investigationQuestID string
	for _, quest := range boardResponse {
		if quest.TemplateType == "investigate_anomaly" {
			investigationQuestID = quest.QuestID
			break
		}
	}
	if investigationQuestID == "" {
		t.Fatal("expected investigate_anomaly quest")
	}

	var actionsResponse struct {
		Data struct {
			Actions []struct {
				ActionType string         `json:"action_type"`
				Label      string         `json:"label"`
				ArgsSchema map[string]any `json:"args_schema"`
			} `json:"actions"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/actions", nil, token, http.StatusOK, &actionsResponse)

	foundInteract := false
	for _, action := range actionsResponse.Data.Actions {
		if action.ActionType == "quest_interact" {
			if action.Label == "" {
				t.Fatal("expected quest_interact label to be populated")
			}
			suggestedInteraction, _ := action.ArgsSchema["suggested_interaction"].(string)
			if suggestedInteraction != "confirm_route" && suggestedInteraction != "inspect_clue" {
				t.Fatalf("expected suggested_interaction to match investigation runtime, got %#v", action.ArgsSchema["suggested_interaction"])
			}
			foundInteract = true
			break
		}
	}
	if !foundInteract {
		t.Fatalf("expected /me/actions to include quest_interact, got %#v", actionsResponse.Data.Actions)
	}

	var stateResponse struct {
		Data struct {
			ValidActions []struct {
				ActionType string `json:"action_type"`
			} `json:"valid_actions"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/state", nil, token, http.StatusOK, &stateResponse)

	foundStateInteract := false
	for _, action := range stateResponse.Data.ValidActions {
		if action.ActionType == "quest_interact" {
			foundStateInteract = true
			break
		}
	}
	if !foundStateInteract {
		t.Fatalf("expected /me/state valid_actions to include quest_interact, got %#v", stateResponse.Data.ValidActions)
	}

	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/actions", map[string]any{
		"action_type": "quest_interact",
		"action_args": map[string]any{
			"quest_id":    investigationQuestID,
			"interaction": "inspect_clue",
		},
	}, token, http.StatusOK, nil)

	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/actions", nil, token, http.StatusOK, &actionsResponse)
	foundChoice := false
	for _, action := range actionsResponse.Data.Actions {
		if action.ActionType == "quest_choice" {
			if action.ArgsSchema["choice_options"] == nil {
				t.Fatal("expected quest_choice args_schema to include choice_options")
			}
			foundChoice = true
			break
		}
	}
	if !foundChoice {
		t.Fatalf("expected /me/actions to include quest_choice after clue inspection, got %#v", actionsResponse.Data.Actions)
	}
}

func TestQuestE2EInvestigationFlowUsesAPIHints(t *testing.T) {
	server := NewServer(config.API{Port: "8080"})
	token, boardResponse := createCharacterWithBoard(t, server, "bot-e2e-investigation", "E2EInvestigator", func(quests []boardQuestView) bool {
		for _, quest := range quests {
			if quest.TemplateType == "investigate_anomaly" && quest.Difficulty == "hard" {
				return true
			}
		}
		return false
	})

	var investigationQuestID string
	for _, quest := range boardResponse {
		if quest.TemplateType == "investigate_anomaly" && quest.Difficulty == "hard" {
			investigationQuestID = quest.QuestID
			break
		}
	}
	if investigationQuestID == "" {
		t.Fatal("expected hard investigate_anomaly quest on board")
	}

	planner := getPlannerE2EView(t, server, token, "main_city")
	hint := findQuestRuntimeHint(t, planner.Data.RuntimeHints, investigationQuestID)
	if hint.SuggestedActionType != "quest_interact" {
		t.Fatalf("expected first investigation hint action quest_interact, got %q", hint.SuggestedActionType)
	}
	interaction := asStringValue(hint.SuggestedActionArgs["interaction"])
	if interaction == "" {
		t.Fatal("expected planner hint to provide interaction name")
	}

	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/actions", map[string]any{
		"action_type": hint.SuggestedActionType,
		"action_args": map[string]any{
			"quest_id":    investigationQuestID,
			"interaction": interaction,
		},
	}, token, http.StatusOK, nil)

	var actionsResponse struct {
		Data struct {
			Actions []actionSchemaView `json:"actions"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/actions", nil, token, http.StatusOK, &actionsResponse)

	choiceAction := findActionByType(t, actionsResponse.Data.Actions, "quest_choice")
	choiceOptions := decodeChoiceOptions(t, choiceAction.ArgsSchema["choice_options"])
	if len(choiceOptions) == 0 {
		t.Fatal("expected quest_choice action to expose choice options")
	}
	selectedChoiceKey := choiceOptions[0].ChoiceKey

	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/actions", map[string]any{
		"action_type": "quest_choice",
		"action_args": map[string]any{
			"quest_id":   investigationQuestID,
			"choice_key": selectedChoiceKey,
		},
	}, token, http.StatusOK, nil)

	planner = getPlannerE2EView(t, server, token, "main_city")
	hint = findQuestRuntimeHint(t, planner.Data.RuntimeHints, investigationQuestID)
	if hint.TargetRegionID == "" {
		t.Fatal("expected planner hint to expose target region after choice")
	}

	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/actions", map[string]any{
		"action_type": "travel",
		"action_args": map[string]any{
			"region_id": hint.TargetRegionID,
		},
	}, token, http.StatusOK, nil)

	assertQuestStatus(t, server, token, investigationQuestID, "completed")
}

func TestQuestE2EHardDeliveryFlowUsesAPIHints(t *testing.T) {
	server := NewServer(config.API{Port: "8080"})
	token, boardResponse := createCharacterWithBoard(t, server, "bot-e2e-delivery", "E2ECourier", func(quests []boardQuestView) bool {
		for _, quest := range quests {
			if quest.TemplateType == "deliver_supplies" && quest.Difficulty == "hard" {
				return true
			}
		}
		return false
	})

	var hardQuestID string
	for _, quest := range boardResponse {
		if quest.TemplateType == "deliver_supplies" && quest.Difficulty == "hard" {
			hardQuestID = quest.QuestID
			break
		}
	}
	if hardQuestID == "" {
		t.Fatal("expected hard delivery quest on board")
	}

	var questDetail struct {
		Data struct {
			Quest struct {
				TargetRegionID string `json:"target_region_id"`
			} `json:"quest"`
			Runtime struct {
				SuggestedActionType string         `json:"suggested_action_type"`
				SuggestedActionArgs map[string]any `json:"suggested_action_args"`
			} `json:"runtime"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/quests/"+hardQuestID, nil, token, http.StatusOK, &questDetail)

	if questDetail.Data.Runtime.SuggestedActionType != "quest_interact" {
		t.Fatalf("expected hard delivery to start with quest_interact, got %q", questDetail.Data.Runtime.SuggestedActionType)
	}
	interaction := asStringValue(questDetail.Data.Runtime.SuggestedActionArgs["interaction"])
	if interaction == "" {
		t.Fatal("expected hard delivery runtime to expose suggested interaction")
	}
	targetRegionID := questDetail.Data.Quest.TargetRegionID
	if targetRegionID == "" {
		t.Fatal("expected hard delivery quest to expose target region")
	}

	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/actions", map[string]any{
		"action_type": questDetail.Data.Runtime.SuggestedActionType,
		"action_args": map[string]any{
			"quest_id":    hardQuestID,
			"interaction": interaction,
		},
	}, token, http.StatusOK, nil)

	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/actions", map[string]any{
		"action_type": "travel",
		"action_args": map[string]any{
			"region_id": targetRegionID,
		},
	}, token, http.StatusOK, nil)

	assertQuestStatus(t, server, token, hardQuestID, "completed")
}

func TestFieldEncounterProgressionFlow(t *testing.T) {
	server := NewServer(config.API{Port: "8080"})
	token, questsResponse := createCharacterWithBoard(t, server, "bot-field-runner", "FieldLoop", func(quests []boardQuestView) bool {
		hasKill := false
		hasCollect := false
		for _, quest := range quests {
			if quest.TargetRegionID != "whispering_forest" {
				continue
			}
			if quest.TemplateType == "kill_region_enemies" {
				hasKill = true
			}
			if quest.TemplateType == "collect_materials" {
				hasCollect = true
			}
		}
		return hasKill && hasCollect
	})

	killQuestID := ""
	collectQuestID := ""
	for _, quest := range questsResponse {
		if quest.TargetRegionID != "whispering_forest" {
			continue
		}
		if quest.TemplateType == "kill_region_enemies" {
			killQuestID = quest.QuestID
		}
		if quest.TemplateType == "collect_materials" {
			collectQuestID = quest.QuestID
		}
	}
	if killQuestID == "" || collectQuestID == "" {
		t.Fatal("expected whispering_forest kill and collect quests")
	}

	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/travel", map[string]any{
		"region_id": "whispering_forest",
	}, token, http.StatusOK, nil)

	var fieldState struct {
		Data struct {
			ValidActions []struct {
				ActionType string `json:"action_type"`
			} `json:"valid_actions"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/state", nil, token, http.StatusOK, &fieldState)

	hasHuntAction := false
	hasGatherAction := false
	hasCurioAction := false
	for _, action := range fieldState.Data.ValidActions {
		switch action.ActionType {
		case "resolve_field_encounter:hunt":
			hasHuntAction = true
		case "resolve_field_encounter:gather":
			hasGatherAction = true
		case "resolve_field_encounter:curio":
			hasCurioAction = true
		}
	}
	if !hasHuntAction || !hasGatherAction || !hasCurioAction {
		t.Fatalf("expected field valid_actions to include hunt/gather/curio, got %#v", fieldState.Data.ValidActions)
	}

	var firstEncounter struct {
		Data struct {
			ActionResult struct {
				ActionType string `json:"action_type"`
				Approach   string `json:"approach"`
				EventType  string `json:"event_type"`
			} `json:"action_result"`
			Result struct {
				RewardGold         int `json:"reward_gold"`
				EnemiesDefeated    int `json:"enemies_defeated"`
				MaterialsCollected int `json:"materials_collected"`
			} `json:"result"`
			State struct {
				Character struct {
					Gold int `json:"gold"`
				} `json:"character"`
				Materials []struct {
					MaterialKey string `json:"material_key"`
					Quantity    int    `json:"quantity"`
				} `json:"materials"`
			} `json:"state"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/field-encounter", map[string]any{
		"approach": "hunt",
	}, token, http.StatusOK, &firstEncounter)

	if firstEncounter.Data.ActionResult.ActionType != "resolve_field_encounter" {
		t.Fatalf("expected dedicated endpoint to return resolve_field_encounter, got %q", firstEncounter.Data.ActionResult.ActionType)
	}
	if firstEncounter.Data.ActionResult.Approach != "hunt" {
		t.Fatalf("expected hunt approach, got %q", firstEncounter.Data.ActionResult.Approach)
	}
	if firstEncounter.Data.ActionResult.EventType != "field.encounter_resolved" {
		t.Fatalf("expected field.encounter_resolved, got %q", firstEncounter.Data.ActionResult.EventType)
	}
	if firstEncounter.Data.Result.RewardGold <= 0 || firstEncounter.Data.Result.EnemiesDefeated <= 0 || firstEncounter.Data.Result.MaterialsCollected <= 0 {
		t.Fatal("expected field encounter to grant gold, enemy progress, and material progress")
	}
	if len(firstEncounter.Data.State.Materials) == 0 {
		t.Fatal("expected field encounter to add materials to character state")
	}

	var secondEncounter struct {
		Data struct {
			ActionResult struct {
				ActionType string `json:"action_type"`
			} `json:"action_result"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/actions", map[string]any{
		"action_type": "resolve_field_encounter:gather",
		"action_args": map[string]any{},
	}, token, http.StatusOK, &secondEncounter)

	if secondEncounter.Data.ActionResult.ActionType != "resolve_field_encounter" {
		t.Fatalf("expected action router to resolve field encounter, got %q", secondEncounter.Data.ActionResult.ActionType)
	}

	var questsAfter struct {
		Data struct {
			Quests []struct {
				QuestID string `json:"quest_id"`
				Status  string `json:"status"`
			} `json:"quests"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/quests", nil, token, http.StatusOK, &questsAfter)

	killCompleted, collectCompleted := false, false
	for attempt := 0; attempt < 3; attempt++ {
		killCompleted = false
		collectCompleted = false
		for _, quest := range questsAfter.Data.Quests {
			if quest.QuestID == killQuestID && quest.Status == "completed" {
				killCompleted = true
			}
			if quest.QuestID == collectQuestID && quest.Status == "completed" {
				collectCompleted = true
			}
		}
		if killCompleted && collectCompleted {
			break
		}
		if !killCompleted {
			doJSONRequest(t, server, http.MethodPost, "/api/v1/me/actions", map[string]any{
				"action_type": "resolve_field_encounter:hunt",
				"action_args": map[string]any{},
			}, token, http.StatusOK, &secondEncounter)
		}
		if !collectCompleted {
			doJSONRequest(t, server, http.MethodPost, "/api/v1/me/actions", map[string]any{
				"action_type": "resolve_field_encounter:gather",
				"action_args": map[string]any{},
			}, token, http.StatusOK, &secondEncounter)
		}
		doJSONRequest(t, server, http.MethodGet, "/api/v1/me/quests", nil, token, http.StatusOK, &questsAfter)
	}
	if !killCompleted || !collectCompleted {
		t.Fatalf("expected field loop to complete both quests, got kill=%v collect=%v", killCompleted, collectCompleted)
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

	foundFieldEvent := false
	for _, item := range eventsResponse.Data.Items {
		if strings.HasPrefix(item.ActorName, "FieldLoop") && item.EventType == "field.encounter_resolved" {
			foundFieldEvent = true
			break
		}
	}
	if !foundFieldEvent {
		t.Fatal("expected public events to include the resolved field encounter")
	}
}

func TestCurioEncounterDoesNotOverflowFixedDailyBoard(t *testing.T) {
	server := NewServer(config.API{Port: "8080"})
	token, initialBoard := createCharacterWithBoard(t, server, "bot-curio-runner", "CurioLoop", func(quests []boardQuestView) bool {
		for _, quest := range quests {
			if quest.TemplateType == "deliver_supplies" && quest.TargetRegionID == "greenfield_village" {
				return true
			}
		}
		return false
	})
	deliveryQuestID := ""
	for _, quest := range initialBoard {
		if quest.TemplateType == "deliver_supplies" && quest.TargetRegionID == "greenfield_village" {
			deliveryQuestID = quest.QuestID
			break
		}
	}
	if deliveryQuestID == "" {
		t.Fatal("expected greenfield delivery contract before curio test")
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/travel", map[string]any{
		"region_id": "greenfield_village",
	}, token, http.StatusOK, nil)
	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/quests/"+deliveryQuestID+"/submit", nil, token, http.StatusOK, nil)
	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/travel", map[string]any{
		"region_id": "whispering_forest",
	}, token, http.StatusOK, nil)

	var encounterResponse struct {
		Data struct {
			Result struct {
				IsCurio       bool   `json:"is_curio"`
				CurioLabel    string `json:"curio_label"`
				CurioOutcome  string `json:"curio_outcome"`
				FollowupQuest any    `json:"followup_quest"`
			} `json:"result"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/field-encounter", map[string]any{
		"approach": "curio",
	}, token, http.StatusOK, &encounterResponse)

	if !encounterResponse.Data.Result.IsCurio {
		t.Fatal("expected curio encounter to mark result as curio")
	}
	if encounterResponse.Data.Result.CurioLabel == "" || encounterResponse.Data.Result.CurioOutcome == "" {
		t.Fatal("expected curio encounter to include curio label and outcome")
	}
	if encounterResponse.Data.Result.FollowupQuest != nil {
		t.Fatal("expected curio encounter not to inject a fifth same-day quest")
	}

	var questsResponse struct {
		Data struct {
			Quests []struct {
				QuestID string `json:"quest_id"`
				Status  string `json:"status"`
				Title   string `json:"title"`
			} `json:"quests"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/quests", nil, token, http.StatusOK, &questsResponse)

	if len(questsResponse.Data.Quests) != characters.DailyQuestBoardSize {
		t.Fatalf("expected quest board to remain at %d entries, got %d", characters.DailyQuestBoardSize, len(questsResponse.Data.Quests))
	}

	var eventsResponse struct {
		Data struct {
			Items []struct {
				EventType string `json:"event_type"`
				ActorName string `json:"actor_name"`
				Summary   string `json:"summary"`
			} `json:"items"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/public/events?limit=20", nil, "", http.StatusOK, &eventsResponse)

	foundCurioEvent := false
	for _, item := range eventsResponse.Data.Items {
		if !strings.HasPrefix(item.ActorName, "CurioLoop") {
			continue
		}
		if item.EventType == "field.curio_resolved" {
			foundCurioEvent = true
		}
	}
	if !foundCurioEvent {
		t.Fatal("expected public events to include the resolved curio encounter")
	}
}

func TestQuestSubmitRouteEnforcesDailyCompletionCap(t *testing.T) {
	server := NewServer(config.API{Port: "8080"})
	token, boardResponse := createCharacterWithBoard(t, server, "bot-quest-cap", "QuestCapper", func(quests []boardQuestView) bool {
		for _, quest := range quests {
			if quest.TemplateType == "deliver_supplies" && quest.TargetRegionID == "greenfield_village" {
				return true
			}
		}
		return false
	})

	deliveryQuestID := ""
	for _, quest := range boardResponse {
		if quest.TemplateType == "deliver_supplies" && quest.TargetRegionID == "greenfield_village" {
			deliveryQuestID = quest.QuestID
			break
		}
	}
	if deliveryQuestID == "" {
		t.Fatal("expected delivery quest for cap test")
	}

	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/travel", map[string]any{
		"region_id": "greenfield_village",
	}, token, http.StatusOK, nil)

	account, err := server.authService.Authenticate(token)
	if err != nil {
		t.Fatal("expected token to authenticate")
	}
	character, exists := server.characterService.GetCharacterByAccount(account)
	if !exists {
		t.Fatal("expected character to exist")
	}
	for i := 0; i < characters.DailyQuestBoardSize; i++ {
		if _, _, _, _, err := server.characterService.ApplyQuestSubmission(character.CharacterID, characters.QuestSummary{
			QuestID:          "quest_cap_seed_" + strconv.Itoa(i),
			Title:            "Cap Seed",
			RewardGold:       0,
			RewardReputation: 0,
		}); err != nil {
			t.Fatalf("seed quest completion cap: %v", err)
		}
	}

	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/quests/"+deliveryQuestID+"/submit", nil, token, http.StatusBadRequest, nil)
}

func TestMeStateInitializesDailyQuestBoard(t *testing.T) {
	server := NewServer(config.API{Port: "8080"})
	token := createQuestE2ECharacter(t, server, "bot-state-board", "StateBoarder")

	var stateResponse struct {
		Data struct {
			Objectives []boardQuestView `json:"objectives"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/state", nil, token, http.StatusOK, &stateResponse)

	if len(stateResponse.Data.Objectives) != characters.DailyQuestBoardSize {
		t.Fatalf("expected /me/state to expose %d objectives, got %d", characters.DailyQuestBoardSize, len(stateResponse.Data.Objectives))
	}
}

func TestMeActionsInitializesDailyQuestBoard(t *testing.T) {
	server := NewServer(config.API{Port: "8080"})
	token := createQuestE2ECharacter(t, server, "bot-actions-board", "ActionBoarder")

	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/actions", nil, token, http.StatusOK, nil)

	account, err := server.authService.Authenticate(token)
	if err != nil {
		t.Fatalf("expected token to authenticate: %v", err)
	}
	character, exists := server.characterService.GetCharacterByAccount(account)
	if !exists {
		t.Fatal("expected character to exist")
	}
	if objectives := server.questService.ActiveObjectives(character.CharacterID); len(objectives) != characters.DailyQuestBoardSize {
		t.Fatalf("expected /me/actions to initialize %d objectives, got %d", characters.DailyQuestBoardSize, len(objectives))
	}
}

func TestSubmitQuestActionDoesNotConsumeQuestWhenCapReached(t *testing.T) {
	server := NewServer(config.API{Port: "8080"})
	token, boardResponse := createCharacterWithBoard(t, server, "bot-action-quest-cap", "ActionQuestCap", func(quests []boardQuestView) bool {
		for _, quest := range quests {
			if quest.TemplateType == "deliver_supplies" && quest.TargetRegionID == "greenfield_village" {
				return true
			}
		}
		return false
	})

	deliveryQuestID := ""
	for _, quest := range boardResponse {
		if quest.TemplateType == "deliver_supplies" && quest.TargetRegionID == "greenfield_village" {
			deliveryQuestID = quest.QuestID
			break
		}
	}
	if deliveryQuestID == "" {
		t.Fatal("expected delivery quest for action cap test")
	}

	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/travel", map[string]any{
		"region_id": "greenfield_village",
	}, token, http.StatusOK, nil)

	account, err := server.authService.Authenticate(token)
	if err != nil {
		t.Fatalf("expected token to authenticate: %v", err)
	}
	character, exists := server.characterService.GetCharacterByAccount(account)
	if !exists {
		t.Fatal("expected character to exist")
	}
	for i := 0; i < characters.DailyQuestBoardSize; i++ {
		if _, _, _, _, err := server.characterService.ApplyQuestSubmission(character.CharacterID, characters.QuestSummary{
			QuestID:          "quest_action_cap_seed_" + strconv.Itoa(i),
			Title:            "Cap Seed",
			RewardGold:       0,
			RewardReputation: 0,
		}); err != nil {
			t.Fatalf("seed quest completion cap: %v", err)
		}
	}

	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/actions", map[string]any{
		"action_type": "submit_quest",
		"action_args": map[string]any{"quest_id": deliveryQuestID},
	}, token, http.StatusBadRequest, nil)

	var boardAfter struct {
		Data struct {
			Quests []boardQuestView `json:"quests"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/quests", nil, token, http.StatusOK, &boardAfter)

	for _, quest := range boardAfter.Data.Quests {
		if quest.QuestID == deliveryQuestID {
			if quest.Status != "completed" {
				t.Fatalf("expected quest to remain completed after capped action submit, got %s", quest.Status)
			}
			return
		}
	}
	t.Fatalf("expected quest %s to remain on board after capped action submit", deliveryQuestID)
}

func TestPublicBotDetailUsesBusinessDayResetForTodayHistory(t *testing.T) {
	server := NewServer(config.API{Port: "8080"})
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatalf("load location: %v", err)
	}
	server.worldService.SetClock(func() time.Time {
		return time.Date(2026, 4, 7, 2, 0, 0, 0, loc)
	})

	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/register", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-business-day",
		"password": "verysecure",
	}), "", http.StatusOK, nil)

	var loginResponse struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/login", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-business-day",
		"password": "verysecure",
	}), "", http.StatusOK, &loginResponse)

	doJSONRequest(t, server, http.MethodPost, "/api/v1/characters", map[string]any{
		"name":         "BusinessDayBot",
		"class":        "warrior",
		"weapon_style": "sword_shield",
	}, loginResponse.Data.AccessToken, http.StatusOK, nil)

	account, err := server.authService.Authenticate(loginResponse.Data.AccessToken)
	if err != nil {
		t.Fatal("expected access token to authenticate")
	}
	character, exists := server.characterService.GetCharacterByAccount(account)
	if !exists {
		t.Fatal("expected character to exist")
	}

	eventTime := time.Date(2026, 4, 6, 23, 30, 0, 0, loc).Format(time.RFC3339)
	if err := server.characterService.AppendEvents(character.CharacterID,
		world.WorldEvent{
			EventID:          "evt_business_quest",
			EventType:        "quest.submitted",
			Visibility:       "public",
			ActorCharacterID: character.CharacterID,
			ActorName:        character.Name,
			RegionID:         "greenfield_village",
			Summary:          character.Name + " submitted a late-night contract.",
			Payload: map[string]any{
				"quest_id":          "quest_business",
				"quest_title":       "Late Contract",
				"reward_gold":       100,
				"reward_reputation": 20,
			},
			OccurredAt: eventTime,
		},
		world.WorldEvent{
			EventID:          "evt_business_dungeon",
			EventType:        "dungeon.loot_granted",
			Visibility:       "public",
			ActorCharacterID: character.CharacterID,
			ActorName:        character.Name,
			RegionID:         "ancient_catacomb",
			Summary:          character.Name + " claimed a dungeon chest before reset.",
			Payload: map[string]any{
				"run_id":      "run_business",
				"dungeon_id":  "ancient_catacomb_v1",
				"reward_gold": 180,
				"rating":      "A",
			},
			OccurredAt: eventTime,
		},
	); err != nil {
		t.Fatalf("AppendEvents failed: %v", err)
	}

	var publicBotDetail struct {
		Data struct {
			CompletedQuestsToday []any `json:"completed_quests_today"`
			DungeonRunsToday     []any `json:"dungeon_runs_today"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/public/bots/"+character.CharacterID, nil, "", http.StatusOK, &publicBotDetail)
	if len(publicBotDetail.Data.CompletedQuestsToday) == 0 {
		t.Fatal("expected pre-4am quest submission to still count for today's business-day history")
	}
	if len(publicBotDetail.Data.DungeonRunsToday) == 0 {
		t.Fatal("expected pre-4am dungeon claim to still count for today's business-day history")
	}
}

func TestPublicRoutesReflectRuntimeData(t *testing.T) {
	server := NewServer(config.API{Port: "8080"})
	token, boardResponse := createCharacterWithBoard(t, server, "bot-public", "PublicRunner", func(quests []boardQuestView) bool {
		for _, quest := range quests {
			if quest.TemplateType == "deliver_supplies" && quest.TargetRegionID == "greenfield_village" {
				return true
			}
		}
		return false
	})

	deliveryQuestID := ""
	for _, quest := range boardResponse {
		if quest.TemplateType == "deliver_supplies" && quest.TargetRegionID == "greenfield_village" {
			deliveryQuestID = quest.QuestID
			break
		}
	}
	if deliveryQuestID == "" {
		t.Fatal("expected a greenfield delivery quest for runtime public data test")
	}

	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/travel", map[string]any{
		"region_id": "greenfield_village",
	}, token, http.StatusOK, nil)
	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/quests/"+deliveryQuestID+"/submit", nil, token, http.StatusOK, nil)

	var enterResponse struct {
		Data struct {
			RunID string `json:"run_id"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/dungeons/ancient_catacomb_v1/enter", dungeonEnterPayload(""), token, http.StatusOK, &enterResponse)
	if enterResponse.Data.RunID == "" {
		t.Fatal("expected dungeon run id for public history test")
	}

	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/runs/"+enterResponse.Data.RunID+"/claim", nil, token, http.StatusOK, nil)

	var worldStateResponse struct {
		Data struct {
			ActiveBotCount       int `json:"active_bot_count"`
			QuestsCompletedToday int `json:"quests_completed_today"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/public/world-state", nil, "", http.StatusOK, &worldStateResponse)

	if worldStateResponse.Data.ActiveBotCount < 1 {
		t.Fatalf("expected at least one active bot, got %d", worldStateResponse.Data.ActiveBotCount)
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
	if !strings.HasPrefix(eventsResponse.Data.Items[0].ActorName, "PublicRunner") {
		t.Fatalf("expected PublicRunner* as latest public actor, got %q", eventsResponse.Data.Items[0].ActorName)
	}
	activePublicActor := eventsResponse.Data.Items[0].ActorName

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

	if len(leaderboardsResponse.Data.Reputation) == 0 || !strings.HasPrefix(leaderboardsResponse.Data.Reputation[0].Name, "PublicRunner") {
		t.Fatal("expected runtime reputation leaderboard to include PublicRunner*")
	}
	if len(leaderboardsResponse.Data.Gold) == 0 || !strings.HasPrefix(leaderboardsResponse.Data.Gold[0].Name, "PublicRunner") {
		t.Fatal("expected runtime gold leaderboard to include PublicRunner*")
	}

	var publicBotsResponse struct {
		Data struct {
			Items []struct {
				CharacterSummary struct {
					CharacterID string `json:"character_id"`
					Name        string `json:"name"`
				} `json:"character_summary"`
			} `json:"items"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/public/bots?q=Public&limit=10", nil, "", http.StatusOK, &publicBotsResponse)
	if len(publicBotsResponse.Data.Items) == 0 {
		t.Fatal("expected bot search to return PublicRunner")
	}

	botID := ""
	var publicBotDetail struct {
		Data struct {
			CompletedQuestsToday []any `json:"completed_quests_today"`
			DungeonRunsToday     []any `json:"dungeon_runs_today"`
			QuestHistory7D       []any `json:"quest_history_7d"`
			DungeonHistory7D     []any `json:"dungeon_history_7d"`
		} `json:"data"`
	}
	for _, item := range publicBotsResponse.Data.Items {
		if item.CharacterSummary.Name != activePublicActor && !strings.HasPrefix(item.CharacterSummary.Name, "PublicRunner") {
			continue
		}
		doJSONRequest(t, server, http.MethodGet, "/api/v1/public/bots/"+item.CharacterSummary.CharacterID, nil, "", http.StatusOK, &publicBotDetail)
		if len(publicBotDetail.Data.QuestHistory7D) > 0 && len(publicBotDetail.Data.DungeonHistory7D) > 0 {
			botID = item.CharacterSummary.CharacterID
			break
		}
	}
	if botID == "" {
		t.Fatal("expected one searched public bot to expose both quest and dungeon 7-day history")
	}

	var questHistoryResponse struct {
		Data struct {
			Items []struct {
				QuestID string `json:"quest_id"`
			} `json:"items"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/public/bots/"+botID+"/quests/history?days=7&limit=20", nil, "", http.StatusOK, &questHistoryResponse)
	if len(questHistoryResponse.Data.Items) == 0 {
		t.Fatal("expected quest history endpoint to return data")
	}

	var dungeonHistoryResponse struct {
		Data struct {
			Items []struct {
				RunID string `json:"run_id"`
			} `json:"items"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/public/bots/"+botID+"/dungeon-runs?days=7&limit=20", nil, "", http.StatusOK, &dungeonHistoryResponse)
	if len(dungeonHistoryResponse.Data.Items) == 0 {
		t.Fatal("expected dungeon history endpoint to return data")
	}

	var dungeonRunDetailResponse struct {
		Data struct {
			RunID string `json:"run_id"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/public/bots/"+botID+"/dungeon-runs/"+enterResponse.Data.RunID, nil, "", http.StatusOK, &dungeonRunDetailResponse)
	if dungeonRunDetailResponse.Data.RunID != enterResponse.Data.RunID {
		t.Fatalf("expected dungeon run detail run id %q, got %q", enterResponse.Data.RunID, dungeonRunDetailResponse.Data.RunID)
	}
}

func TestDungeonRunHistoryListAndDetailLevels(t *testing.T) {
	server := NewServer(config.API{Port: "8080"})

	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/register", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-run-history",
		"password": "verysecure",
	}), "", http.StatusOK, nil)

	var loginResponse struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/login", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-run-history",
		"password": "verysecure",
	}), "", http.StatusOK, &loginResponse)

	doJSONRequest(t, server, http.MethodPost, "/api/v1/characters", map[string]any{
		"name":         "RunHistory",
		"class":        "warrior",
		"weapon_style": "sword_shield",
	}, loginResponse.Data.AccessToken, http.StatusOK, nil)

	var enterResponse struct {
		Data struct {
			RunID string `json:"run_id"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/dungeons/ancient_catacomb_v1/enter", dungeonEnterPayload(""), loginResponse.Data.AccessToken, http.StatusOK, &enterResponse)

	var listResponse struct {
		Data struct {
			Items []struct {
				RunID              string   `json:"run_id"`
				DungeonID          string   `json:"dungeon_id"`
				RunStatus          string   `json:"run_status"`
				HighestRoomCleared int      `json:"highest_room_cleared"`
				PotionLoadout      []string `json:"potion_loadout"`
				BossReached        bool     `json:"boss_reached"`
				SummaryTag         string   `json:"summary_tag"`
			} `json:"items"`
			NextCursor string `json:"next_cursor"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/runs?dungeon_id=ancient_catacomb_v1&result=cleared&limit=1", nil, loginResponse.Data.AccessToken, http.StatusOK, &listResponse)

	if len(listResponse.Data.Items) != 1 {
		t.Fatalf("expected one run summary, got %#v", listResponse.Data.Items)
	}
	if listResponse.Data.Items[0].RunID != enterResponse.Data.RunID {
		t.Fatalf("expected listed run_id %q, got %q", enterResponse.Data.RunID, listResponse.Data.Items[0].RunID)
	}
	if listResponse.Data.Items[0].SummaryTag == "" {
		t.Fatal("expected run summary_tag to be present")
	}
	if !listResponse.Data.Items[0].BossReached {
		t.Fatal("expected cleared run to mark boss_reached")
	}
	if len(listResponse.Data.Items[0].PotionLoadout) == 0 {
		t.Fatal("expected compact run summary to include potion_loadout")
	}

	var compactDetail struct {
		Data map[string]any `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/runs/"+enterResponse.Data.RunID+"?detail_level=compact", nil, loginResponse.Data.AccessToken, http.StatusOK, &compactDetail)

	if compactDetail.Data["summary_tag"] == nil {
		t.Fatal("expected compact detail to include summary_tag")
	}
	if _, exists := compactDetail.Data["recent_battle_log"]; exists {
		t.Fatalf("expected compact detail to omit recent_battle_log, got %#v", compactDetail.Data["recent_battle_log"])
	}
	if _, exists := compactDetail.Data["battle_log"]; exists {
		t.Fatalf("expected compact detail to omit battle_log, got %#v", compactDetail.Data["battle_log"])
	}

	var standardDetail struct {
		Data map[string]any `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/runs/"+enterResponse.Data.RunID, nil, loginResponse.Data.AccessToken, http.StatusOK, &standardDetail)

	if _, exists := standardDetail.Data["key_findings"]; !exists {
		t.Fatal("expected standard detail to include key_findings")
	}
	if _, exists := standardDetail.Data["danger_rooms"]; !exists {
		t.Fatal("expected standard detail to include danger_rooms")
	}
	if _, exists := standardDetail.Data["resource_pressure"]; !exists {
		t.Fatal("expected standard detail to include resource_pressure")
	}
	if _, exists := standardDetail.Data["reward_summary"]; !exists {
		t.Fatal("expected standard detail to include reward_summary")
	}
	if _, exists := standardDetail.Data["recent_battle_log"]; !exists {
		t.Fatal("expected standard detail to include recent_battle_log")
	}
	if _, exists := standardDetail.Data["battle_log"]; exists {
		t.Fatal("expected standard detail to omit battle_log")
	}

	var verboseDetail struct {
		Data map[string]any `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/runs/"+enterResponse.Data.RunID+"?detail_level=verbose", nil, loginResponse.Data.AccessToken, http.StatusOK, &verboseDetail)

	battleLog, ok := verboseDetail.Data["battle_log"].([]any)
	if !ok || len(battleLog) == 0 {
		t.Fatalf("expected verbose detail to include battle_log, got %#v", verboseDetail.Data["battle_log"])
	}
}

func TestInventoryIncludesUpgradeHintsAndPotionOptions(t *testing.T) {
	server := NewServer(config.API{Port: "8080"})

	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/register", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-prep-inventory",
		"password": "verysecure",
	}), "", http.StatusOK, nil)

	var loginResponse struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/login", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-prep-inventory",
		"password": "verysecure",
	}), "", http.StatusOK, &loginResponse)

	doJSONRequest(t, server, http.MethodPost, "/api/v1/characters", map[string]any{
		"name":         "PrepBag",
		"class":        "warrior",
		"weapon_style": "sword_shield",
	}, loginResponse.Data.AccessToken, http.StatusOK, nil)

	var inventoryResponse struct {
		Data struct {
			EquipmentScore int `json:"equipment_score"`
			UpgradeHints   []struct {
				Source     string `json:"source"`
				Slot       string `json:"slot"`
				ScoreDelta int    `json:"score_delta"`
			} `json:"upgrade_hints"`
			PotionLoadoutOptions []struct {
				CatalogID     string `json:"catalog_id"`
				AvailableNow  bool   `json:"available_now"`
				Recommended   bool   `json:"recommended"`
				QuantityOwned int    `json:"quantity_owned"`
			} `json:"potion_loadout_options"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/inventory", nil, loginResponse.Data.AccessToken, http.StatusOK, &inventoryResponse)

	if inventoryResponse.Data.EquipmentScore <= 0 {
		t.Fatalf("expected positive equipment_score, got %d", inventoryResponse.Data.EquipmentScore)
	}
	if len(inventoryResponse.Data.UpgradeHints) == 0 {
		t.Fatal("expected upgrade_hints to be present")
	}
	if len(inventoryResponse.Data.PotionLoadoutOptions) == 0 {
		t.Fatal("expected potion_loadout_options to be present")
	}
	foundRecommendedPotion := false
	for _, option := range inventoryResponse.Data.PotionLoadoutOptions {
		if option.Recommended && option.AvailableNow && option.QuantityOwned > 0 {
			foundRecommendedPotion = true
			break
		}
	}
	if !foundRecommendedPotion {
		t.Fatal("expected at least one recommended potion option to be immediately available")
	}
}

func TestBlacksmithSalvageAndEnhanceFlow(t *testing.T) {
	server := NewServer(config.API{Port: "8080"})

	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/register", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-blacksmith",
		"password": "verysecure",
	}), "", http.StatusOK, nil)

	var loginResponse struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/login", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-blacksmith",
		"password": "verysecure",
	}), "", http.StatusOK, &loginResponse)

	doJSONRequest(t, server, http.MethodPost, "/api/v1/characters", map[string]any{
		"name":         "ForgeRunner",
		"class":        "warrior",
		"weapon_style": "sword_shield",
	}, loginResponse.Data.AccessToken, http.StatusOK, nil)

	var inventoryResponse struct {
		Data struct {
			Equipped []struct {
				ItemID string `json:"item_id"`
				Slot   string `json:"slot"`
			} `json:"equipped"`
			Inventory []struct {
				ItemID string `json:"item_id"`
			} `json:"inventory"`
			SlotEnhancements []struct {
				Slot             string `json:"slot"`
				EnhancementLevel int    `json:"enhancement_level"`
			} `json:"slot_enhancements"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/inventory", nil, loginResponse.Data.AccessToken, http.StatusOK, &inventoryResponse)

	if len(inventoryResponse.Data.Equipped) == 0 {
		t.Fatal("expected equipped starter items")
	}
	if len(inventoryResponse.Data.SlotEnhancements) == 0 {
		t.Fatal("expected slot_enhancements in inventory response")
	}
	if len(inventoryResponse.Data.Inventory) < 2 {
		t.Fatal("expected at least two starter inventory items to salvage")
	}

	weaponItemID := ""
	for _, item := range inventoryResponse.Data.Equipped {
		if item.Slot == "weapon" {
			weaponItemID = item.ItemID
			break
		}
	}
	if weaponItemID == "" {
		t.Fatal("expected equipped weapon item")
	}

	for i := 0; i < 2; i++ {
		doJSONRequest(t, server, http.MethodPost, "/api/v1/buildings/blacksmith_main_city/salvage", map[string]any{
			"item_id": inventoryResponse.Data.Inventory[i].ItemID,
		}, loginResponse.Data.AccessToken, http.StatusOK, nil)
	}

	var stateBeforeEnhance struct {
		Data struct {
			Materials []struct {
				MaterialKey string `json:"material_key"`
				Quantity    int    `json:"quantity"`
			} `json:"materials"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/state", nil, loginResponse.Data.AccessToken, http.StatusOK, &stateBeforeEnhance)

	shardsBefore := 0
	for _, item := range stateBeforeEnhance.Data.Materials {
		if item.MaterialKey == "enhancement_shard" {
			shardsBefore = item.Quantity
			break
		}
	}
	if shardsBefore < 2 {
		t.Fatalf("expected at least 2 enhancement_shard after salvage, got %d", shardsBefore)
	}

	var enhanceResponse struct {
		Data struct {
			Result struct {
				Item struct {
					ItemID           string `json:"item_id"`
					EnhancementLevel int    `json:"enhancement_level"`
				} `json:"item"`
				EnhancementQuote struct {
					GoldCost int `json:"gold_cost"`
				} `json:"enhancement_quote"`
			} `json:"result"`
			State struct {
				Character struct {
					Gold int `json:"gold"`
				} `json:"character"`
				Materials []struct {
					MaterialKey string `json:"material_key"`
					Quantity    int    `json:"quantity"`
				} `json:"materials"`
			} `json:"state"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/buildings/blacksmith_main_city/enhance", map[string]any{
		"slot": "weapon",
	}, loginResponse.Data.AccessToken, http.StatusOK, &enhanceResponse)

	if enhanceResponse.Data.Result.Item.ItemID != weaponItemID {
		t.Fatalf("expected enhanced weapon item %q, got %q", weaponItemID, enhanceResponse.Data.Result.Item.ItemID)
	}
	if enhanceResponse.Data.Result.Item.EnhancementLevel != 1 {
		t.Fatalf("expected enhancement level 1, got %d", enhanceResponse.Data.Result.Item.EnhancementLevel)
	}
	if enhanceResponse.Data.Result.EnhancementQuote.GoldCost <= 0 {
		t.Fatal("expected positive enhancement gold cost")
	}

	shardsAfter := 0
	for _, item := range enhanceResponse.Data.State.Materials {
		if item.MaterialKey == "enhancement_shard" {
			shardsAfter = item.Quantity
			break
		}
	}
	if shardsAfter >= shardsBefore {
		t.Fatalf("expected enhancement materials to decrease after enhance, before=%d after=%d", shardsBefore, shardsAfter)
	}

	var stateAfterEnhance struct {
		Data struct {
			SlotEnhancements []struct {
				Slot             string `json:"slot"`
				EnhancementLevel int    `json:"enhancement_level"`
			} `json:"slot_enhancements"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/state", nil, loginResponse.Data.AccessToken, http.StatusOK, &stateAfterEnhance)
	weaponLevel := -1
	for _, item := range stateAfterEnhance.Data.SlotEnhancements {
		if item.Slot == "weapon" {
			weaponLevel = item.EnhancementLevel
			break
		}
	}
	if weaponLevel != 1 {
		t.Fatalf("expected weapon slot enhancement level 1 in state, got %d", weaponLevel)
	}
}

func TestGenericActionBusSellAndEnhanceFlow(t *testing.T) {
	server := NewServer(config.API{Port: "8080"})

	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/register", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-generic-action-buildings",
		"password": "verysecure",
	}), "", http.StatusOK, nil)

	var loginResponse struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/login", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-generic-action-buildings",
		"password": "verysecure",
	}), "", http.StatusOK, &loginResponse)

	doJSONRequest(t, server, http.MethodPost, "/api/v1/characters", map[string]any{
		"name":         "GenericBuilder",
		"class":        "warrior",
		"weapon_style": "sword_shield",
	}, loginResponse.Data.AccessToken, http.StatusOK, nil)

	var stateBefore struct {
		Data struct {
			Character struct {
				Gold int `json:"gold"`
			} `json:"character"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/state", nil, loginResponse.Data.AccessToken, http.StatusOK, &stateBefore)

	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/actions", map[string]any{
		"action_type": "unequip_item",
		"action_args": map[string]any{
			"slot": "chest",
		},
	}, loginResponse.Data.AccessToken, http.StatusOK, nil)

	var inventoryResponse struct {
		Data struct {
			Equipped []struct {
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

	if len(inventoryResponse.Data.Inventory) < 3 {
		t.Fatalf("expected at least three inventory items after unequip, got %d", len(inventoryResponse.Data.Inventory))
	}

	weaponItemID := ""
	for _, item := range inventoryResponse.Data.Equipped {
		if item.Slot == "weapon" {
			weaponItemID = item.ItemID
			break
		}
	}
	if weaponItemID == "" {
		t.Fatal("expected equipped weapon item")
	}

	sellItemID := inventoryResponse.Data.Inventory[0].ItemID
	salvageItemID1 := inventoryResponse.Data.Inventory[1].ItemID
	salvageItemID2 := inventoryResponse.Data.Inventory[2].ItemID

	var sellResponse struct {
		Data struct {
			ActionResult struct {
				BuildingID string `json:"building_id"`
			} `json:"action_result"`
			Result struct {
				GainGold int `json:"gain_gold"`
				Item     struct {
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
	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/actions", map[string]any{
		"action_type": "sell_item",
		"action_args": map[string]any{
			"item_id": sellItemID,
		},
	}, loginResponse.Data.AccessToken, http.StatusOK, &sellResponse)

	if sellResponse.Data.ActionResult.BuildingID != "equipment_shop_main_city" {
		t.Fatalf("expected generic sell_item to resolve equipment_shop_main_city, got %q", sellResponse.Data.ActionResult.BuildingID)
	}
	if sellResponse.Data.Result.Item.ItemID != sellItemID {
		t.Fatalf("expected sold item %q, got %q", sellItemID, sellResponse.Data.Result.Item.ItemID)
	}
	if sellResponse.Data.Result.GainGold <= 0 {
		t.Fatal("expected sell_item to grant positive gold")
	}
	if sellResponse.Data.State.Character.Gold <= stateBefore.Data.Character.Gold {
		t.Fatalf("expected gold to increase after sell_item, before=%d after=%d", stateBefore.Data.Character.Gold, sellResponse.Data.State.Character.Gold)
	}

	doJSONRequest(t, server, http.MethodPost, "/api/v1/buildings/blacksmith_main_city/salvage", map[string]any{
		"item_id": salvageItemID1,
	}, loginResponse.Data.AccessToken, http.StatusOK, nil)
	doJSONRequest(t, server, http.MethodPost, "/api/v1/buildings/blacksmith_main_city/salvage", map[string]any{
		"item_id": salvageItemID2,
	}, loginResponse.Data.AccessToken, http.StatusOK, nil)

	var enhanceResponse struct {
		Data struct {
			ActionResult struct {
				BuildingID string `json:"building_id"`
			} `json:"action_result"`
			Result struct {
				Item struct {
					ItemID           string `json:"item_id"`
					EnhancementLevel int    `json:"enhancement_level"`
				} `json:"item"`
				EnhancementQuote struct {
					GoldCost     int              `json:"gold_cost"`
					MaterialCost []map[string]any `json:"material_cost"`
				} `json:"enhancement_quote"`
			} `json:"result"`
			State struct {
				SlotEnhancements []struct {
					Slot             string `json:"slot"`
					EnhancementLevel int    `json:"enhancement_level"`
				} `json:"slot_enhancements"`
			} `json:"state"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/actions", map[string]any{
		"action_type": "enhance_item",
		"action_args": map[string]any{
			"slot": "weapon",
		},
	}, loginResponse.Data.AccessToken, http.StatusOK, &enhanceResponse)

	if enhanceResponse.Data.ActionResult.BuildingID != "blacksmith_main_city" {
		t.Fatalf("expected generic enhance_item to resolve blacksmith_main_city, got %q", enhanceResponse.Data.ActionResult.BuildingID)
	}
	if enhanceResponse.Data.Result.Item.ItemID != weaponItemID {
		t.Fatalf("expected enhanced weapon %q, got %q", weaponItemID, enhanceResponse.Data.Result.Item.ItemID)
	}
	if enhanceResponse.Data.Result.Item.EnhancementLevel != 1 {
		t.Fatalf("expected enhancement level 1, got %d", enhanceResponse.Data.Result.Item.EnhancementLevel)
	}
	if enhanceResponse.Data.Result.EnhancementQuote.GoldCost <= 0 {
		t.Fatal("expected positive enhancement gold cost")
	}

	shardCost := 0
	for _, material := range enhanceResponse.Data.Result.EnhancementQuote.MaterialCost {
		if key, _ := material["material_key"].(string); key == "enhancement_shard" {
			switch quantity := material["quantity"].(type) {
			case float64:
				shardCost = int(quantity)
			case int:
				shardCost = quantity
			}
			break
		}
	}
	if shardCost <= 0 {
		t.Fatal("expected enhancement quote to spend enhancement_shard")
	}

	weaponLevel := -1
	for _, item := range enhanceResponse.Data.State.SlotEnhancements {
		if item.Slot == "weapon" {
			weaponLevel = item.EnhancementLevel
			break
		}
	}
	if weaponLevel != 1 {
		t.Fatalf("expected weapon slot enhancement level 1 after generic enhance_item, got %d", weaponLevel)
	}
}

func TestPlannerIncludesDungeonPreparationHints(t *testing.T) {
	server := NewServer(config.API{Port: "8080"})

	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/register", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-prep-planner",
		"password": "verysecure",
	}), "", http.StatusOK, nil)

	var loginResponse struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/login", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-prep-planner",
		"password": "verysecure",
	}), "", http.StatusOK, &loginResponse)

	doJSONRequest(t, server, http.MethodPost, "/api/v1/characters", map[string]any{
		"name":         "PrepPlanner",
		"class":        "warrior",
		"weapon_style": "sword_shield",
	}, loginResponse.Data.AccessToken, http.StatusOK, nil)

	var plannerResponse struct {
		Data struct {
			LocalDungeons []struct {
				CurrentPower              int    `json:"current_power"`
				RecommendedPower          int    `json:"recommended_power"`
				PowerGap                  int    `json:"power_gap"`
				DungeonID                 string `json:"dungeon_id"`
				CurrentEquipmentScore     int    `json:"current_equipment_score"`
				RecommendedEquipmentScore int    `json:"recommended_equipment_score"`
				ScoreGap                  int    `json:"score_gap"`
				Readiness                 string `json:"readiness"`
			} `json:"local_dungeons"`
			DungeonPreparation struct {
				CurrentPower          int `json:"current_power"`
				CurrentEquipmentScore int `json:"current_equipment_score"`
				UpgradeHintCount      int `json:"upgrade_hint_count"`
				PotionOptionCount     int `json:"potion_option_count"`
				Items                 []struct {
					CurrentPower               int              `json:"current_power"`
					RecommendedPower           int              `json:"recommended_power"`
					PowerGap                   int              `json:"power_gap"`
					DungeonID                  string           `json:"dungeon_id"`
					Readiness                  string           `json:"readiness"`
					ScoreGap                   int              `json:"score_gap"`
					InventoryUpgradeCount      int              `json:"inventory_upgrade_count"`
					AffordableShopUpgradeCount int              `json:"affordable_shop_upgrade_count"`
					SuggestedPreparationSteps  []string         `json:"suggested_preparation_steps"`
					PotionOptions              []map[string]any `json:"potion_options"`
				} `json:"items"`
			} `json:"dungeon_preparation"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/planner?region_id=ancient_catacomb", nil, loginResponse.Data.AccessToken, http.StatusOK, &plannerResponse)

	if len(plannerResponse.Data.LocalDungeons) == 0 {
		t.Fatal("expected local_dungeons in planner response")
	}
	if plannerResponse.Data.DungeonPreparation.CurrentEquipmentScore <= 0 {
		t.Fatal("expected dungeon_preparation current_equipment_score")
	}
	if plannerResponse.Data.DungeonPreparation.CurrentPower <= 0 {
		t.Fatal("expected dungeon_preparation current_power")
	}
	if len(plannerResponse.Data.DungeonPreparation.Items) == 0 {
		t.Fatal("expected dungeon_preparation items")
	}
	item := plannerResponse.Data.DungeonPreparation.Items[0]
	if item.CurrentPower <= 0 || item.RecommendedPower <= 0 {
		t.Fatal("expected dungeon_preparation item current/recommended power")
	}
	if item.DungeonID == "" {
		t.Fatal("expected dungeon_preparation item dungeon_id")
	}
	if item.Readiness == "" {
		t.Fatal("expected dungeon_preparation readiness")
	}
	if len(item.PotionOptions) == 0 {
		t.Fatal("expected dungeon_preparation potion_options")
	}
	if len(item.SuggestedPreparationSteps) == 0 && item.ScoreGap > 0 {
		t.Fatal("expected preparation steps when score gap exists")
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
	doJSONRequest(t, server, http.MethodPost, "/api/v1/dungeons/ancient_catacomb_v1/enter", dungeonEnterPayload(""), loginResponse.Data.AccessToken, http.StatusOK, &enterResponse)

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
			RunID           string           `json:"run_id"`
			BattleState     map[string]any   `json:"battle_state"`
			RecentBattleLog []map[string]any `json:"recent_battle_log"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/runs/"+enterResponse.Data.RunID, nil, loginResponse.Data.AccessToken, http.StatusOK, &getRunResponse)
	if getRunResponse.Data.RunID != enterResponse.Data.RunID {
		t.Fatalf("expected run_id %q, got %q", enterResponse.Data.RunID, getRunResponse.Data.RunID)
	}
	if getRunResponse.Data.BattleState["engine_mode"] != "auto_turn_based" {
		t.Fatalf("expected battle_state.engine_mode auto_turn_based, got %#v", getRunResponse.Data.BattleState["engine_mode"])
	}
	if len(getRunResponse.Data.RecentBattleLog) == 0 {
		t.Fatal("expected recent_battle_log to be populated")
	}
	if getRunResponse.Data.RecentBattleLog[0]["event_type"] == nil {
		t.Fatal("expected recent_battle_log entries to include event_type")
	}
	actionFound := false
	for _, item := range getRunResponse.Data.RecentBattleLog {
		if item["event_type"] == "action" {
			actionFound = true
			if item["actor"] == nil || item["target"] == nil || item["turn"] == nil {
				t.Fatal("expected action log to include actor, target, and turn")
			}
			if item["target_hp_before"] == nil || item["target_hp_after"] == nil {
				t.Fatal("expected action log to include target hp before/after")
			}
			if item["cooldown_before_round"] == nil || item["cooldown_after_round"] == nil {
				t.Fatal("expected action log to include cooldown snapshots")
			}
			break
		}
	}
	if !actionFound {
		t.Fatal("expected at least one action event in recent_battle_log")
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

func TestBuildPublicDungeonRunDetailFallsBackToEventHistoryWhenRuntimeRunMissing(t *testing.T) {
	server := NewServer(config.API{Port: "8080"})

	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/register", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-history-fallback",
		"password": "verysecure",
	}), "", http.StatusOK, nil)

	var loginResponse struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/login", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-history-fallback",
		"password": "verysecure",
	}), "", http.StatusOK, &loginResponse)

	doJSONRequest(t, server, http.MethodPost, "/api/v1/characters", map[string]any{
		"name":         "HistoryFallback",
		"class":        "mage",
		"weapon_style": "staff",
	}, loginResponse.Data.AccessToken, http.StatusOK, nil)

	var enterResponse struct {
		Data struct {
			RunID string `json:"run_id"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/dungeons/ancient_catacomb_v1/enter", dungeonEnterPayload(""), loginResponse.Data.AccessToken, http.StatusOK, &enterResponse)

	var state struct {
		Data struct {
			Character struct {
				CharacterID string `json:"character_id"`
			} `json:"character"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/state", nil, loginResponse.Data.AccessToken, http.StatusOK, &state)

	_, _, _, events, ok := server.characterService.GetRuntimeDetailByCharacterID(state.Data.Character.CharacterID)
	if !ok {
		t.Fatal("expected runtime detail for fallback test character")
	}

	payload, found := buildPublicDungeonRunDetail(
		state.Data.Character.CharacterID,
		enterResponse.Data.RunID,
		events,
		dungeons.NewService(),
	)
	if !found {
		t.Fatal("expected fallback payload to be built from event history")
	}

	runID, _ := payload["run_id"].(string)
	if runID != enterResponse.Data.RunID {
		t.Fatalf("expected fallback run id %q, got %q", enterResponse.Data.RunID, runID)
	}

	result, _ := payload["result"].(map[string]any)
	if result["run_status"] != "cleared" {
		t.Fatalf("expected fallback run status cleared, got %#v", result["run_status"])
	}
	if result["runtime_phase"] != "history_only" {
		t.Fatalf("expected history_only runtime phase, got %#v", result["runtime_phase"])
	}
	if result["reward_claimable"] != true {
		t.Fatal("expected uncleared historical loot state to remain claimable in fallback payload")
	}

	battleLog, _ := payload["battle_log"].([]map[string]any)
	if len(battleLog) != 0 {
		t.Fatalf("expected no runtime battle log in fallback payload, got %d items", len(battleLog))
	}

	difficulty, _ := payload["difficulty"].(string)
	if difficulty != "unknown" {
		t.Fatalf("expected fallback difficulty unknown, got %q", difficulty)
	}
}

func TestDungeonEnterDifficultyQueryAndFallback(t *testing.T) {
	server := NewServer(config.API{Port: "8080"})

	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/register", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-dungeon-diff-query",
		"password": "verysecure",
	}), "", http.StatusOK, nil)

	var loginResponse struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/login", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-dungeon-diff-query",
		"password": "verysecure",
	}), "", http.StatusOK, &loginResponse)

	doJSONRequest(t, server, http.MethodPost, "/api/v1/characters", map[string]any{
		"name":         "DiffQueryRunner",
		"class":        "warrior",
		"weapon_style": "sword_shield",
	}, loginResponse.Data.AccessToken, http.StatusOK, nil)

	var hardEnter struct {
		Data struct {
			RunID        string `json:"run_id"`
			Difficulty   string `json:"difficulty"`
			RunStatus    string `json:"run_status"`
			RuntimePhase string `json:"runtime_phase"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/dungeons/ancient_catacomb_v1/enter", dungeonEnterPayload("hard"), loginResponse.Data.AccessToken, http.StatusOK, &hardEnter)

	if hardEnter.Data.RunID == "" {
		t.Fatal("expected hard difficulty run_id")
	}
	if hardEnter.Data.Difficulty != "hard" {
		t.Fatalf("expected difficulty hard, got %q", hardEnter.Data.Difficulty)
	}

	var hardPublicDetail struct {
		Data struct {
			RunID      string `json:"run_id"`
			Difficulty string `json:"difficulty"`
		} `json:"data"`
	}

	var state struct {
		Data struct {
			Character struct {
				CharacterID string `json:"character_id"`
			} `json:"character"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/state", nil, loginResponse.Data.AccessToken, http.StatusOK, &state)

	doJSONRequest(t, server, http.MethodGet, "/api/v1/public/bots/"+state.Data.Character.CharacterID+"/dungeon-runs/"+hardEnter.Data.RunID, nil, "", http.StatusOK, &hardPublicDetail)
	if hardPublicDetail.Data.Difficulty != "hard" {
		t.Fatalf("expected public run detail difficulty hard, got %q", hardPublicDetail.Data.Difficulty)
	}

	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/runs/"+hardEnter.Data.RunID+"/claim", nil, loginResponse.Data.AccessToken, http.StatusOK, nil)

	var fallbackEnter struct {
		Data struct {
			Difficulty string `json:"difficulty"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/dungeons/ancient_catacomb_v1/enter", dungeonEnterPayload("invalid_value"), loginResponse.Data.AccessToken, http.StatusOK, &fallbackEnter)
	if fallbackEnter.Data.Difficulty != "easy" {
		t.Fatalf("expected invalid difficulty to fallback to easy, got %q", fallbackEnter.Data.Difficulty)
	}
}

func TestActionEnterDungeonWithDifficulty(t *testing.T) {
	server := NewServer(config.API{Port: "8080"})

	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/register", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-dungeon-diff-action",
		"password": "verysecure",
	}), "", http.StatusOK, nil)

	var loginResponse struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/login", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-dungeon-diff-action",
		"password": "verysecure",
	}), "", http.StatusOK, &loginResponse)

	doJSONRequest(t, server, http.MethodPost, "/api/v1/characters", map[string]any{
		"name":         "DiffActionRunner",
		"class":        "warrior",
		"weapon_style": "sword_shield",
	}, loginResponse.Data.AccessToken, http.StatusOK, nil)

	var actionResponse struct {
		Data struct {
			State struct {
				Run struct {
					RunID      string `json:"run_id"`
					Difficulty string `json:"difficulty"`
				} `json:"run"`
			} `json:"state"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/actions", map[string]any{
		"action_type": "enter_dungeon",
		"action_args": map[string]any{
			"dungeon_id":     "ancient_catacomb_v1",
			"difficulty":     "nightmare",
			"potion_loadout": []string{"potion_hp_t2", "potion_atk_t2"},
		},
	}, loginResponse.Data.AccessToken, http.StatusOK, &actionResponse)

	if actionResponse.Data.State.Run.RunID == "" {
		t.Fatal("expected run_id from action enter_dungeon")
	}
	if actionResponse.Data.State.Run.Difficulty != "nightmare" {
		t.Fatalf("expected action difficulty nightmare, got %q", actionResponse.Data.State.Run.Difficulty)
	}
}

func TestActionClaimDungeonRewards(t *testing.T) {
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
	doJSONRequest(t, server, http.MethodPost, "/api/v1/dungeons/ancient_catacomb_v1/enter", dungeonEnterPayload(""), loginResponse.Data.AccessToken, http.StatusOK, &enterResponse)

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
		"action_type": "claim_dungeon_rewards",
		"action_args": map[string]any{
			"run_id": enterResponse.Data.RunID,
		},
	}, loginResponse.Data.AccessToken, http.StatusOK, &actionResponse)

	if actionResponse.Data.ActionResult.ActionType != "claim_dungeon_rewards" {
		t.Fatalf("expected action_type claim_dungeon_rewards, got %q", actionResponse.Data.ActionResult.ActionType)
	}
	if actionResponse.Data.ActionResult.RunID != enterResponse.Data.RunID {
		t.Fatalf("expected run_id %q, got %q", enterResponse.Data.RunID, actionResponse.Data.ActionResult.RunID)
	}
	if actionResponse.Data.State.Run.RuntimePhase != "claim_settled" {
		t.Fatalf("expected claim_settled runtime phase, got %q", actionResponse.Data.State.Run.RuntimePhase)
	}
}

func TestActionRestoreHp(t *testing.T) {
	server := NewServer(config.API{Port: "8080"})

	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/register", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-restore-hp-alias",
		"password": "verysecure",
	}), "", http.StatusOK, nil)

	var loginResponse struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/login", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-restore-hp-alias",
		"password": "verysecure",
	}), "", http.StatusOK, &loginResponse)

	doJSONRequest(t, server, http.MethodPost, "/api/v1/characters", map[string]any{
		"name":         "RestoreAliasRunner",
		"class":        "priest",
		"weapon_style": "scepter",
	}, loginResponse.Data.AccessToken, http.StatusOK, nil)

	var actionResponse struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/actions", map[string]any{
		"action_type": "restore_hp",
		"action_args": map[string]any{},
	}, loginResponse.Data.AccessToken, http.StatusBadRequest, &actionResponse)

	if actionResponse.Error.Code != "ACTION_NOT_SUPPORTED" {
		t.Fatalf("expected ACTION_NOT_SUPPORTED, got %q", actionResponse.Error.Code)
	}
}

func TestActionEnterBuildingRequiresCurrentRegion(t *testing.T) {
	server := NewServer(config.API{Port: "8080"})

	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/register", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-enter-building-scope",
		"password": "verysecure",
	}), "", http.StatusOK, nil)

	var loginResponse struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/login", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-enter-building-scope",
		"password": "verysecure",
	}), "", http.StatusOK, &loginResponse)

	doJSONRequest(t, server, http.MethodPost, "/api/v1/characters", map[string]any{
		"name":         "ScopedBuilder",
		"class":        "warrior",
		"weapon_style": "sword_shield",
	}, loginResponse.Data.AccessToken, http.StatusOK, nil)

	var errResponse struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/actions", map[string]any{
		"action_type": "enter_building",
		"action_args": map[string]any{
			"building_id": "apothecary_village",
		},
	}, loginResponse.Data.AccessToken, http.StatusBadRequest, &errResponse)

	if errResponse.Error.Code != "ACTION_NOT_SUPPORTED" {
		t.Fatalf("expected ACTION_NOT_SUPPORTED, got %q", errResponse.Error.Code)
	}
}

func TestActionsExposeSuggestedTargetsAndLinkedDungeon(t *testing.T) {
	server := NewServer(config.API{Port: "8080"})

	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/register", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-actions-targets",
		"password": "verysecure",
	}), "", http.StatusOK, nil)

	var loginResponse struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/login", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-actions-targets",
		"password": "verysecure",
	}), "", http.StatusOK, &loginResponse)

	doJSONRequest(t, server, http.MethodPost, "/api/v1/characters", map[string]any{
		"name":         "ActionTargeter",
		"class":        "warrior",
		"weapon_style": "sword_shield",
	}, loginResponse.Data.AccessToken, http.StatusOK, nil)

	var actionsResponse struct {
		Data struct {
			Actions []actionSchemaView `json:"actions"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/actions", nil, loginResponse.Data.AccessToken, http.StatusOK, &actionsResponse)

	foundTravelSuggestion := false
	foundBuildingSuggestion := false
	for _, action := range actionsResponse.Data.Actions {
		switch action.ActionType {
		case "travel":
			if suggestedRegionID, _ := action.ArgsSchema["suggested_region_id"].(string); suggestedRegionID != "" {
				foundTravelSuggestion = true
			}
		case "enter_building":
			if suggestedBuildingID, _ := action.ArgsSchema["suggested_building_id"].(string); suggestedBuildingID != "" {
				foundBuildingSuggestion = true
			}
		}
	}
	if !foundTravelSuggestion {
		t.Fatalf("expected /me/actions travel entries to expose suggested_region_id, got %#v", actionsResponse.Data.Actions)
	}
	if !foundBuildingSuggestion {
		t.Fatalf("expected /me/actions building entries to expose suggested_building_id, got %#v", actionsResponse.Data.Actions)
	}

	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/travel", map[string]any{
		"region_id": "whispering_forest",
	}, loginResponse.Data.AccessToken, http.StatusOK, nil)

	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/actions", nil, loginResponse.Data.AccessToken, http.StatusOK, &actionsResponse)

	enterDungeonAction := findActionByType(t, actionsResponse.Data.Actions, "enter_dungeon")
	suggestedDungeonID, _ := enterDungeonAction.ArgsSchema["suggested_dungeon_id"].(string)
	if suggestedDungeonID != "ancient_catacomb_v1" {
		t.Fatalf("expected linked dungeon suggested_dungeon_id ancient_catacomb_v1, got %#v", enterDungeonAction.ArgsSchema["suggested_dungeon_id"])
	}
}

func TestActionsExposeClaimAndSubmitNextSteps(t *testing.T) {
	server := NewServer(config.API{Port: "8080"})

	accessToken, boardResponse := createCharacterWithBoard(t, server, "bot-actions-next-steps", "ActionLooper", func(quests []boardQuestView) bool {
		for _, quest := range quests {
			if quest.TemplateType == "clear_dungeon" || quest.TemplateType == "kill_dungeon_elite" {
				return true
			}
		}
		return false
	})

	var dungeonQuestID string
	var dungeonQuestDifficulty string
	for _, quest := range boardResponse {
		if quest.TemplateType == "clear_dungeon" || quest.TemplateType == "kill_dungeon_elite" {
			dungeonQuestID = quest.QuestID
			dungeonQuestDifficulty = quest.Difficulty
			break
		}
	}
	if dungeonQuestID == "" {
		t.Fatal("expected dungeon quest on generated board")
	}

	if dungeonQuestDifficulty == "nightmare" {
		doJSONRequest(t, server, http.MethodPost, "/api/v1/me/quests/"+dungeonQuestID+"/interact", map[string]any{
			"interaction": "inspect_clue",
		}, accessToken, http.StatusOK, nil)
		doJSONRequest(t, server, http.MethodPost, "/api/v1/me/quests/"+dungeonQuestID+"/choice", map[string]any{
			"choice_key": "follow_standard_brief",
		}, accessToken, http.StatusOK, nil)
	}

	var enterResponse struct {
		Data struct {
			RunID string `json:"run_id"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/dungeons/ancient_catacomb_v1/enter", dungeonEnterPayload(""), accessToken, http.StatusOK, &enterResponse)

	var actionsResponse struct {
		Data struct {
			Actions []actionSchemaView `json:"actions"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/actions", nil, accessToken, http.StatusOK, &actionsResponse)

	foundClaimAction := false
	foundSubmitAction := false
	for _, action := range actionsResponse.Data.Actions {
		switch action.ActionType {
		case "claim_dungeon_rewards":
			if suggestedRunID, _ := action.ArgsSchema["suggested_run_id"].(string); suggestedRunID == enterResponse.Data.RunID {
				foundClaimAction = true
			}
		case "submit_quest":
			if suggestedQuestID, _ := action.ArgsSchema["suggested_quest_id"].(string); suggestedQuestID == dungeonQuestID {
				foundSubmitAction = true
			}
		}
	}

	if !foundClaimAction {
		t.Fatalf("expected /me/actions to expose claim_dungeon_rewards for run %q, got %#v", enterResponse.Data.RunID, actionsResponse.Data.Actions)
	}
	if !foundSubmitAction {
		t.Fatalf("expected /me/actions to expose submit_quest for completed quest %q, got %#v", dungeonQuestID, actionsResponse.Data.Actions)
	}
}

func TestDungeonEnterStillAllowedAfterDailyClaimCapReached(t *testing.T) {
	server := NewServer(config.API{Port: "8080"})

	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/register", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-dungeon-cap-enter",
		"password": "verysecure",
	}), "", http.StatusOK, nil)

	var loginResponse struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/login", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-dungeon-cap-enter",
		"password": "verysecure",
	}), "", http.StatusOK, &loginResponse)

	doJSONRequest(t, server, http.MethodPost, "/api/v1/characters", map[string]any{
		"name":         "CapEnterRunner",
		"class":        "warrior",
		"weapon_style": "sword_shield",
	}, loginResponse.Data.AccessToken, http.StatusOK, nil)

	for i := 0; i < 2; i++ {
		var run struct {
			Data struct {
				RunID string `json:"run_id"`
			} `json:"data"`
		}
		doJSONRequest(t, server, http.MethodPost, "/api/v1/dungeons/ancient_catacomb_v1/enter", dungeonEnterPayload(""), loginResponse.Data.AccessToken, http.StatusOK, &run)
		if run.Data.RunID == "" {
			t.Fatalf("expected run_id for capped-prep run #%d", i+1)
		}
		doJSONRequest(t, server, http.MethodPost, "/api/v1/me/runs/"+run.Data.RunID+"/claim", nil, loginResponse.Data.AccessToken, http.StatusOK, nil)
	}

	var thirdRun struct {
		Data struct {
			RunID           string `json:"run_id"`
			RewardClaimable bool   `json:"reward_claimable"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/dungeons/ancient_catacomb_v1/enter", dungeonEnterPayload(""), loginResponse.Data.AccessToken, http.StatusOK, &thirdRun)
	if thirdRun.Data.RunID == "" {
		t.Fatal("expected enter to still return a run_id after claim cap is reached")
	}
	if !thirdRun.Data.RewardClaimable {
		t.Fatal("expected entered run to be claimable even after claim cap is reached")
	}

	var claimError struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/runs/"+thirdRun.Data.RunID+"/claim", nil, loginResponse.Data.AccessToken, http.StatusBadRequest, &claimError)
	if claimError.Error.Code != "DUNGEON_REWARD_CLAIM_LIMIT_REACHED" {
		t.Fatalf("expected DUNGEON_REWARD_CLAIM_LIMIT_REACHED, got %q", claimError.Error.Code)
	}
}

func TestDungeonClaimGrantsItemAndClearsStagedRewards(t *testing.T) {
	server := NewServer(config.API{Port: "8080"})

	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/register", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-dungeon-loot",
		"password": "verysecure",
	}), "", http.StatusOK, nil)

	var loginResponse struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/login", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-dungeon-loot",
		"password": "verysecure",
	}), "", http.StatusOK, &loginResponse)

	doJSONRequest(t, server, http.MethodPost, "/api/v1/characters", map[string]any{
		"name":         "LootRunner",
		"class":        "warrior",
		"weapon_style": "sword_shield",
	}, loginResponse.Data.AccessToken, http.StatusOK, nil)

	var beforeInventory struct {
		Data struct {
			Equipped  []any `json:"equipped"`
			Inventory []any `json:"inventory"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/inventory", nil, loginResponse.Data.AccessToken, http.StatusOK, &beforeInventory)
	beforeTotalItems := len(beforeInventory.Data.Equipped) + len(beforeInventory.Data.Inventory)

	var run struct {
		Data struct {
			RunID string `json:"run_id"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/dungeons/ancient_catacomb_v1/enter", dungeonEnterPayload(""), loginResponse.Data.AccessToken, http.StatusOK, &run)

	var runDetail struct {
		Data struct {
			CurrentRating        *string          `json:"current_rating"`
			PendingRatingRewards []map[string]any `json:"pending_rating_rewards"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/runs/"+run.Data.RunID, nil, loginResponse.Data.AccessToken, http.StatusOK, &runDetail)
	if runDetail.Data.CurrentRating == nil || *runDetail.Data.CurrentRating != "S" {
		t.Fatalf("expected auto-resolve run rating S, got %#v", runDetail.Data.CurrentRating)
	}
	if len(runDetail.Data.PendingRatingRewards) < 2 {
		t.Fatalf("expected S rating to stage multiple rating rewards, got %#v", runDetail.Data.PendingRatingRewards)
	}

	var claimResponse struct {
		Data struct {
			PendingRatingRewards []map[string]any `json:"pending_rating_rewards"`
			StagedMaterialDrops  []map[string]any `json:"staged_material_drops"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/runs/"+run.Data.RunID+"/claim", nil, loginResponse.Data.AccessToken, http.StatusOK, &claimResponse)
	if len(claimResponse.Data.PendingRatingRewards) != 0 {
		t.Fatalf("expected pending_rating_rewards to be empty after claim, got %#v", claimResponse.Data.PendingRatingRewards)
	}
	if len(claimResponse.Data.StagedMaterialDrops) != 0 {
		t.Fatalf("expected staged_material_drops to be empty after claim, got %#v", claimResponse.Data.StagedMaterialDrops)
	}

	var afterInventory struct {
		Data struct {
			Equipped  []any `json:"equipped"`
			Inventory []any `json:"inventory"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/inventory", nil, loginResponse.Data.AccessToken, http.StatusOK, &afterInventory)
	afterTotalItems := len(afterInventory.Data.Equipped) + len(afterInventory.Data.Inventory)
	if afterTotalItems <= beforeTotalItems {
		t.Fatalf("expected dungeon claim to grant an inventory item, before=%d after=%d", beforeTotalItems, afterTotalItems)
	}

	var stateWithMaterials struct {
		Data struct {
			Materials []struct {
				MaterialKey string `json:"material_key"`
				Quantity    int    `json:"quantity"`
			} `json:"materials"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/state", nil, loginResponse.Data.AccessToken, http.StatusOK, &stateWithMaterials)

	foundEssence := false
	for _, material := range stateWithMaterials.Data.Materials {
		if material.MaterialKey == "dungeon_essence" && material.Quantity > 0 {
			foundEssence = true
			break
		}
	}
	if !foundEssence {
		t.Fatalf("expected claimed material dungeon_essence in state materials, got %#v", stateWithMaterials.Data.Materials)
	}
}

func TestDungeonQuestProgressesOnEnter(t *testing.T) {
	server := NewServer(config.API{Port: "8080"})

	accessToken, boardResponse := createCharacterWithBoard(t, server, "bot-dungeon-quest", "DungeonQuestor", func(quests []boardQuestView) bool {
		for _, quest := range quests {
			if quest.TemplateType == "clear_dungeon" || quest.TemplateType == "kill_dungeon_elite" {
				return true
			}
		}
		return false
	})

	var dungeonQuestID string
	var dungeonQuestDifficulty string
	for _, quest := range boardResponse {
		if quest.TemplateType == "clear_dungeon" || quest.TemplateType == "kill_dungeon_elite" {
			dungeonQuestID = quest.QuestID
			dungeonQuestDifficulty = quest.Difficulty
			break
		}
	}
	if dungeonQuestID == "" {
		t.Fatal("expected generated quest board to include a dungeon quest")
	}

	if dungeonQuestDifficulty == "nightmare" {
		doJSONRequest(t, server, http.MethodPost, "/api/v1/me/quests/"+dungeonQuestID+"/interact", map[string]any{
			"interaction": "inspect_clue",
		}, accessToken, http.StatusOK, nil)
		doJSONRequest(t, server, http.MethodPost, "/api/v1/me/quests/"+dungeonQuestID+"/choice", map[string]any{
			"choice_key": "follow_standard_brief",
		}, accessToken, http.StatusOK, nil)
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/dungeons/ancient_catacomb_v1/enter", dungeonEnterPayload(""), accessToken, http.StatusOK, nil)

	var questsAfter struct {
		Data struct {
			Quests []struct {
				QuestID         string `json:"quest_id"`
				TemplateType    string `json:"template_type"`
				Status          string `json:"status"`
				ProgressCurrent int    `json:"progress_current"`
				ProgressTarget  int    `json:"progress_target"`
			} `json:"quests"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/quests", nil, accessToken, http.StatusOK, &questsAfter)

	foundCompleted := false
	for _, quest := range questsAfter.Data.Quests {
		if quest.QuestID != dungeonQuestID {
			continue
		}
		if quest.Status != "completed" {
			t.Fatalf("expected accepted dungeon quest to become completed, got %q", quest.Status)
		}
		if quest.ProgressCurrent != quest.ProgressTarget {
			t.Fatalf("expected dungeon quest progress to reach target, got %d/%d", quest.ProgressCurrent, quest.ProgressTarget)
		}
		foundCompleted = true
		break
	}
	if !foundCompleted {
		t.Fatalf("expected quest %q after dungeon enter", dungeonQuestID)
	}
}

func TestStateIncludesDungeonDailyHints(t *testing.T) {
	server := NewServer(config.API{Port: "8080"})

	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/register", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-dungeon-hints",
		"password": "verysecure",
	}), "", http.StatusOK, nil)

	var loginResponse struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/login", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-dungeon-hints",
		"password": "verysecure",
	}), "", http.StatusOK, &loginResponse)

	doJSONRequest(t, server, http.MethodPost, "/api/v1/characters", map[string]any{
		"name":         "HintRunner",
		"class":        "warrior",
		"weapon_style": "sword_shield",
	}, loginResponse.Data.AccessToken, http.StatusOK, nil)

	var initialState struct {
		Data struct {
			DungeonDaily struct {
				HasRemainingQuota bool     `json:"has_remaining_quota"`
				HasClaimableRun   bool     `json:"has_claimable_run"`
				PendingRunIDs     []string `json:"pending_claim_run_ids"`
			} `json:"dungeon_daily"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/state", nil, loginResponse.Data.AccessToken, http.StatusOK, &initialState)

	if !initialState.Data.DungeonDaily.HasRemainingQuota {
		t.Fatal("expected remaining dungeon quota for a new character")
	}
	if initialState.Data.DungeonDaily.HasClaimableRun {
		t.Fatal("expected no claimable runs before entering dungeon")
	}

	var enterResponse struct {
		Data struct {
			RunID string `json:"run_id"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/dungeons/ancient_catacomb_v1/enter", dungeonEnterPayload(""), loginResponse.Data.AccessToken, http.StatusOK, &enterResponse)

	var pendingState struct {
		Data struct {
			DungeonDaily struct {
				HasClaimableRun bool     `json:"has_claimable_run"`
				PendingRunIDs   []string `json:"pending_claim_run_ids"`
			} `json:"dungeon_daily"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/state", nil, loginResponse.Data.AccessToken, http.StatusOK, &pendingState)

	if !pendingState.Data.DungeonDaily.HasClaimableRun {
		t.Fatal("expected claimable run after dungeon auto-resolve")
	}
	if len(pendingState.Data.DungeonDaily.PendingRunIDs) == 0 || pendingState.Data.DungeonDaily.PendingRunIDs[0] != enterResponse.Data.RunID {
		t.Fatalf("expected pending_claim_run_ids to include %q, got %#v", enterResponse.Data.RunID, pendingState.Data.DungeonDaily.PendingRunIDs)
	}
}

func TestPlannerEndpointReturnsTodayAndRegionalOptions(t *testing.T) {
	server := NewServer(config.API{Port: "8080"})

	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/register", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-planner",
		"password": "verysecure",
	}), "", http.StatusOK, nil)

	var loginResponse struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/login", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-planner",
		"password": "verysecure",
	}), "", http.StatusOK, &loginResponse)

	doJSONRequest(t, server, http.MethodPost, "/api/v1/characters", map[string]any{
		"name":         "PlannerOne",
		"class":        "warrior",
		"weapon_style": "sword_shield",
	}, loginResponse.Data.AccessToken, http.StatusOK, nil)

	var plannerResponse struct {
		Data struct {
			CharacterRegionID string `json:"character_region_id"`
			QueryRegionID     string `json:"query_region_id"`
			Today             struct {
				QuestCompletion struct {
					Used      int `json:"used"`
					Cap       int `json:"cap"`
					Remaining int `json:"remaining"`
				} `json:"quest_completion"`
				DungeonClaim struct {
					Used      int `json:"used"`
					Cap       int `json:"cap"`
					Remaining int `json:"remaining"`
				} `json:"dungeon_claim"`
			} `json:"today"`
			LocalQuests []struct {
				QuestID      string `json:"quest_id"`
				TargetRegion string `json:"target_region_id"`
				TemplateType string `json:"template_type"`
				Status       string `json:"status"`
			} `json:"local_quests"`
			LocalDungeons []struct {
				DungeonID string `json:"dungeon_id"`
				RegionID  string `json:"region_id"`
				CanEnter  bool   `json:"can_enter"`
			} `json:"local_dungeons"`
			SuggestedActions []string `json:"suggested_actions"`
		} `json:"data"`
	}

	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/planner?region_id=ancient_catacomb", nil, loginResponse.Data.AccessToken, http.StatusOK, &plannerResponse)

	if plannerResponse.Data.CharacterRegionID != "main_city" {
		t.Fatalf("expected character_region_id main_city, got %q", plannerResponse.Data.CharacterRegionID)
	}
	if plannerResponse.Data.QueryRegionID != "ancient_catacomb" {
		t.Fatalf("expected query_region_id ancient_catacomb, got %q", plannerResponse.Data.QueryRegionID)
	}
	if plannerResponse.Data.Today.QuestCompletion.Cap <= 0 || plannerResponse.Data.Today.DungeonClaim.Cap <= 0 {
		t.Fatal("expected positive daily caps in planner today summary")
	}
	if plannerResponse.Data.Today.QuestCompletion.Used != 0 || plannerResponse.Data.Today.DungeonClaim.Used != 0 {
		t.Fatal("expected fresh character planner used counters to be 0")
	}

	if len(plannerResponse.Data.LocalDungeons) == 0 {
		t.Fatal("expected planner local_dungeons for ancient_catacomb")
	}
	if plannerResponse.Data.LocalDungeons[0].DungeonID != "ancient_catacomb_v1" {
		t.Fatalf("expected ancient_catacomb_v1, got %q", plannerResponse.Data.LocalDungeons[0].DungeonID)
	}
	if !plannerResponse.Data.LocalDungeons[0].CanEnter {
		t.Fatal("expected character to be able to enter ancient_catacomb_v1")
	}

	hasEnterDungeon := false
	for _, action := range plannerResponse.Data.SuggestedActions {
		if action == "enter_dungeon" {
			hasEnterDungeon = true
		}
	}
	if !hasEnterDungeon {
		t.Fatalf("expected suggested_actions include enter_dungeon, got %#v", plannerResponse.Data.SuggestedActions)
	}

	var fieldPlannerResponse struct {
		Data struct {
			SuggestedActions []string `json:"suggested_actions"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/planner?region_id=whispering_forest", nil, loginResponse.Data.AccessToken, http.StatusOK, &fieldPlannerResponse)

	hasFieldHunt := false
	hasFieldGather := false
	hasFieldCurio := false
	for _, action := range fieldPlannerResponse.Data.SuggestedActions {
		switch action {
		case "resolve_field_encounter:hunt":
			hasFieldHunt = true
		case "resolve_field_encounter:gather":
			hasFieldGather = true
		case "resolve_field_encounter:curio":
			hasFieldCurio = true
		}
	}
	if !hasFieldHunt || !hasFieldGather || !hasFieldCurio {
		t.Fatalf("expected whispering_forest planner suggested_actions to include field encounter modes, got %#v", fieldPlannerResponse.Data.SuggestedActions)
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
	account, err := server.authService.Authenticate(loginResponse.Data.AccessToken)
	if err != nil {
		t.Fatalf("failed to authenticate building test account: %v", err)
	}
	summary, ok := server.characterService.GetCharacterByAccount(account)
	if !ok {
		t.Fatal("expected building test character to exist")
	}
	if _, err := server.characterService.GrantGold(summary.CharacterID, 3000); err != nil {
		t.Fatalf("failed to grant gold for building purchase test: %v", err)
	}

	var shopInventory struct {
		Data struct {
			Items []struct {
				CatalogID string `json:"catalog_id"`
				PriceGold int    `json:"price_gold"`
			} `json:"items"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/buildings/equipment_shop_main_city/shop-inventory", nil, loginResponse.Data.AccessToken, http.StatusOK, &shopInventory)
	if len(shopInventory.Data.Items) == 0 {
		t.Fatal("expected non-empty equipment shop inventory")
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
	doJSONRequest(t, server, http.MethodPost, "/api/v1/buildings/equipment_shop_main_city/purchase", map[string]any{
		"catalog_id": purchaseCatalogID,
	}, loginResponse.Data.AccessToken, http.StatusOK, &purchaseResponse)

	if purchaseResponse.Data.Result.Item.ItemID == "" {
		t.Fatal("expected purchased item_id")
	}
	if purchaseResponse.Data.Result.PriceGold != purchasePrice {
		t.Fatalf("expected purchase price %d, got %d", purchasePrice, purchaseResponse.Data.Result.PriceGold)
	}
	if purchaseResponse.Data.State.Character.Gold != 3100-purchasePrice {
		t.Fatalf("expected gold %d after purchase, got %d", 3100-purchasePrice, purchaseResponse.Data.State.Character.Gold)
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
	doJSONRequest(t, server, http.MethodPost, "/api/v1/buildings/equipment_shop_main_city/sell", map[string]any{
		"item_id": purchaseResponse.Data.Result.Item.ItemID,
	}, loginResponse.Data.AccessToken, http.StatusOK, &sellResponse)
	if sellResponse.Data.Result.GainGold <= 0 {
		t.Fatalf("expected gain_gold > 0, got %d", sellResponse.Data.Result.GainGold)
	}
	if sellResponse.Data.State.Character.Gold <= purchaseResponse.Data.State.Character.Gold {
		t.Fatal("expected gold to increase after sell")
	}

	var apothecaryInventory struct {
		Data struct {
			Items []struct {
				CatalogID string `json:"catalog_id"`
				ItemType  string `json:"item_type"`
				Family    string `json:"family"`
				Tier      int    `json:"tier"`
				PriceGold int    `json:"price_gold"`
			} `json:"items"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/buildings/apothecary_main_city/shop-inventory", nil, loginResponse.Data.AccessToken, http.StatusOK, &apothecaryInventory)
	if len(apothecaryInventory.Data.Items) == 0 {
		t.Fatal("expected non-empty apothecary inventory")
	}
	if apothecaryInventory.Data.Items[0].ItemType != "consumable" {
		t.Fatalf("expected apothecary items to be consumables, got %q", apothecaryInventory.Data.Items[0].ItemType)
	}

	var apothecaryPurchase struct {
		Data struct {
			Result struct {
				PriceGold  int `json:"price_gold"`
				Consumable struct {
					CatalogID string `json:"catalog_id"`
					Quantity  int    `json:"quantity"`
				} `json:"consumable"`
			} `json:"result"`
			State struct {
				Character struct {
					Gold int `json:"gold"`
				} `json:"character"`
			} `json:"state"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/buildings/apothecary_main_city/purchase", map[string]any{
		"catalog_id": apothecaryInventory.Data.Items[0].CatalogID,
	}, loginResponse.Data.AccessToken, http.StatusOK, &apothecaryPurchase)
	if apothecaryPurchase.Data.Result.Consumable.CatalogID == "" {
		t.Fatal("expected purchased consumable catalog_id")
	}
	if apothecaryPurchase.Data.Result.Consumable.Quantity <= 0 {
		t.Fatalf("expected purchased consumable quantity > 0, got %d", apothecaryPurchase.Data.Result.Consumable.Quantity)
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
			Consumables []struct {
				CatalogID string `json:"catalog_id"`
				Quantity  int    `json:"quantity"`
			} `json:"consumables"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/inventory", nil, loginResponse.Data.AccessToken, http.StatusOK, &inventoryView)
	if len(inventoryView.Data.Equipped) == 0 {
		t.Fatal("expected at least one equipped item")
	}
	if len(inventoryView.Data.Consumables) == 0 {
		t.Fatal("expected purchased consumable to appear in inventory")
	}

	doJSONRequest(t, server, http.MethodPost, "/api/v1/buildings/equipment_shop_main_city/sell", map[string]any{
		"item_id": inventoryView.Data.Equipped[0].ItemID,
	}, loginResponse.Data.AccessToken, http.StatusBadRequest, &sellEquippedErr)
	if sellEquippedErr.Error.Code != "INVALID_ACTION_STATE" {
		t.Fatalf("expected INVALID_ACTION_STATE, got %q", sellEquippedErr.Error.Code)
	}
}

func TestArenaRatingBoardUsesPanelPowerScore(t *testing.T) {
	server := NewServer(config.API{Port: "8080"})
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatalf("load location: %v", err)
	}
	fixedNow := time.Date(2026, 4, 2, 8, 30, 0, 0, loc)
	server.worldService.SetClock(func() time.Time { return fixedNow })
	server.arenaService.SetClock(func() time.Time { return fixedNow })

	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/register", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-arena-power",
		"password": "verysecure",
	}), "", http.StatusOK, nil)

	var loginResponse struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/login", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-arena-power",
		"password": "verysecure",
	}), "", http.StatusOK, &loginResponse)

	doJSONRequest(t, server, http.MethodPost, "/api/v1/characters", map[string]any{
		"name":         "ArenaPower",
		"class":        "warrior",
		"weapon_style": "great_axe",
	}, loginResponse.Data.AccessToken, http.StatusOK, nil)

	var stateResponse struct {
		Data struct {
			Character struct {
				CharacterID string `json:"character_id"`
			} `json:"character"`
			CombatPower struct {
				PanelPowerScore int `json:"panel_power_score"`
			} `json:"combat_power"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/state", nil, loginResponse.Data.AccessToken, http.StatusOK, &stateResponse)
	if stateResponse.Data.CombatPower.PanelPowerScore <= 0 {
		t.Fatal("expected positive panel_power_score before arena signup")
	}
	if _, _, _, _, err := server.characterService.ApplyQuestSubmission(stateResponse.Data.Character.CharacterID, characters.QuestSummary{
		QuestID:          "quest_rank_seed",
		Title:            "Arena eligibility warmup",
		RewardReputation: 250,
		RewardGold:       0,
		ProgressTarget:   1,
	}); err != nil {
		t.Fatalf("seed arena eligibility reputation: %v", err)
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/state", nil, loginResponse.Data.AccessToken, http.StatusOK, &stateResponse)

	var boardResponse struct {
		Data struct {
			CharacterID string `json:"character_id"`
			Rating      int    `json:"rating"`
			Candidates  []struct {
				CharacterID     string `json:"character_id"`
				PanelPowerScore int    `json:"panel_power_score"`
				EquipmentScore  int    `json:"equipment_score"`
			} `json:"candidates"`
			Leaderboard []struct {
				CharacterID string `json:"character_id"`
				Rating      int    `json:"rating"`
			} `json:"leaderboard"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/arena/rating-board", nil, loginResponse.Data.AccessToken, http.StatusOK, &boardResponse)
	if boardResponse.Data.CharacterID == "" {
		t.Fatal("expected rating board character id")
	}
	if boardResponse.Data.Rating != 1000 {
		t.Fatalf("expected starting rating 1000, got %d", boardResponse.Data.Rating)
	}
	if len(boardResponse.Data.Leaderboard) == 0 {
		t.Fatal("expected rating leaderboard entries")
	}

	var currentResponse struct {
		Data struct {
			HighestPower    int `json:"highest_panel_power"`
			LowestPower     int `json:"lowest_panel_power"`
			MedianPower     int `json:"median_panel_power"`
			FeaturedEntries []struct {
				CharacterID     string `json:"character_id"`
				PanelPowerScore int    `json:"panel_power_score"`
			} `json:"featured_entries"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/arena/current", nil, "", http.StatusOK, &currentResponse)
	if len(currentResponse.Data.FeaturedEntries) == 0 {
		t.Fatal("expected arena current featured entries during rating week")
	}
	if currentResponse.Data.HighestPower < currentResponse.Data.LowestPower {
		t.Fatalf("expected current highest power >= lowest power, got %d < %d", currentResponse.Data.HighestPower, currentResponse.Data.LowestPower)
	}
	if currentResponse.Data.MedianPower <= 0 {
		t.Fatal("expected arena current median power summary")
	}
	if currentResponse.Data.FeaturedEntries[0].PanelPowerScore <= 0 {
		t.Fatal("expected arena current featured entry panel_power_score")
	}
}

func TestArenaHistoryEndpointsExposeResolvedBattles(t *testing.T) {
	server := NewServer(config.API{Port: "8080"})
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatalf("load location: %v", err)
	}
	currentNow := time.Date(2026, 4, 2, 8, 30, 0, 0, loc)
	server.worldService.SetClock(func() time.Time { return currentNow })
	server.arenaService.SetClock(func() time.Time { return currentNow })

	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/register", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-arena-history",
		"password": "verysecure",
	}), "", http.StatusOK, nil)

	var loginResponse struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/login", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-arena-history",
		"password": "verysecure",
	}), "", http.StatusOK, &loginResponse)

	doJSONRequest(t, server, http.MethodPost, "/api/v1/characters", map[string]any{
		"name":         "ArenaArchivist",
		"class":        "warrior",
		"weapon_style": "sword_shield",
	}, loginResponse.Data.AccessToken, http.StatusOK, nil)

	var stateResponse struct {
		Data struct {
			Character struct {
				CharacterID string `json:"character_id"`
			} `json:"character"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/state", nil, loginResponse.Data.AccessToken, http.StatusOK, &stateResponse)
	if _, _, _, _, err := server.characterService.ApplyQuestSubmission(stateResponse.Data.Character.CharacterID, characters.QuestSummary{
		QuestID:          "quest_rank_seed_history",
		Title:            "Arena eligibility warmup",
		RewardReputation: 250,
	}); err != nil {
		t.Fatalf("seed arena eligibility reputation: %v", err)
	}

	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/register", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-arena-history-rival",
		"password": "verysecure",
	}), "", http.StatusOK, nil)

	var rivalLogin struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/login", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-arena-history-rival",
		"password": "verysecure",
	}), "", http.StatusOK, &rivalLogin)

	doJSONRequest(t, server, http.MethodPost, "/api/v1/characters", map[string]any{
		"name":         "ArenaHistorianRival",
		"class":        "mage",
		"weapon_style": "spellbook",
	}, rivalLogin.Data.AccessToken, http.StatusOK, nil)

	var boardResponse struct {
		Data struct {
			Candidates []struct {
				CharacterID string `json:"character_id"`
			} `json:"candidates"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/arena/rating-board", nil, loginResponse.Data.AccessToken, http.StatusOK, &boardResponse)
	if len(boardResponse.Data.Candidates) == 0 {
		t.Fatal("expected arena rating candidates")
	}

	doJSONRequest(t, server, http.MethodPost, "/api/v1/arena/rating-challenges", map[string]any{
		"target_character_id": boardResponse.Data.Candidates[0].CharacterID,
	}, loginResponse.Data.AccessToken, http.StatusOK, nil)

	currentNow = time.Date(2026, 4, 2, 10, 0, 0, 0, loc)

	var historyResponse struct {
		Data struct {
			Items []struct {
				MatchID string `json:"match_id"`
				Result  string `json:"result"`
			} `json:"items"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/arena-history?limit=20", nil, loginResponse.Data.AccessToken, http.StatusOK, &historyResponse)
	if len(historyResponse.Data.Items) == 0 {
		t.Fatal("expected arena history items after arena resolution")
	}

	matchID := historyResponse.Data.Items[0].MatchID
	if matchID == "" {
		t.Fatal("expected arena history item to include match_id")
	}

	var detailResponse struct {
		Data struct {
			MatchID      string         `json:"match_id"`
			Result       string         `json:"result"`
			BattleReport map[string]any `json:"battle_report"`
			BattleLog    []any          `json:"battle_log"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/arena-history/"+matchID+"?detail_level=verbose", nil, loginResponse.Data.AccessToken, http.StatusOK, &detailResponse)
	if detailResponse.Data.MatchID != matchID {
		t.Fatalf("expected arena history detail match %q, got %q", matchID, detailResponse.Data.MatchID)
	}
	if detailResponse.Data.BattleReport == nil {
		t.Fatal("expected arena history detail to include battle_report")
	}
	if detailResponse.Data.Result != "bye" && len(detailResponse.Data.BattleLog) == 0 {
		t.Fatal("expected non-bye arena history detail to include battle log")
	}

	var publicDetailResponse struct {
		Data struct {
			MatchID      string         `json:"match_id"`
			BattleReport map[string]any `json:"battle_report"`
			BattleLog    []any          `json:"battle_log"`
			LeftEntry    map[string]any `json:"left_entry"`
			RightEntry   map[string]any `json:"right_entry"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/arena/matches/"+matchID+"?detail_level=verbose", nil, "", http.StatusOK, &publicDetailResponse)
	if publicDetailResponse.Data.MatchID != matchID {
		t.Fatalf("expected public arena match detail match %q, got %q", matchID, publicDetailResponse.Data.MatchID)
	}
	if publicDetailResponse.Data.BattleReport == nil {
		t.Fatal("expected public arena match detail to include battle_report")
	}
	if detailResponse.Data.Result != "bye" {
		if _, ok := publicDetailResponse.Data.BattleReport["left_final_hp"]; !ok {
			t.Fatal("expected public arena match detail to include left_final_hp")
		}
		if _, ok := publicDetailResponse.Data.BattleReport["right_final_hp"]; !ok {
			t.Fatal("expected public arena match detail to include right_final_hp")
		}
		if _, ok := publicDetailResponse.Data.BattleReport["end_reason"]; !ok {
			t.Fatal("expected public arena match detail to include end_reason")
		}
		if _, ok := publicDetailResponse.Data.BattleReport["adjudication"]; !ok {
			t.Fatal("expected public arena match detail to include adjudication")
		}
	}
	if detailResponse.Data.Result != "bye" && len(publicDetailResponse.Data.BattleLog) == 0 {
		t.Fatal("expected public non-bye arena match detail to include battle log")
	}
}

func TestRegionBuildingsExposeFunctionalAndNeutralCategories(t *testing.T) {
	server := NewServer(config.API{Port: "8080"})

	var regionDetail struct {
		Data struct {
			Buildings []struct {
				BuildingID string `json:"building_id"`
				Type       string `json:"type"`
				Category   string `json:"category"`
			} `json:"buildings"`
		} `json:"data"`
	}

	doJSONRequest(t, server, http.MethodGet, "/api/v1/regions/greenfield_village", nil, "", http.StatusOK, &regionDetail)

	categoryByID := make(map[string]string, len(regionDetail.Data.Buildings))
	for _, building := range regionDetail.Data.Buildings {
		categoryByID[building.BuildingID] = building.Category
	}

	if categoryByID["guild_outpost_village"] != "functional_building" {
		t.Fatalf("expected guild outpost to be functional_building, got %q", categoryByID["guild_outpost_village"])
	}
	if categoryByID["caravan_dispatch_village"] != "neutral_interaction_point" {
		t.Fatalf("expected caravan dispatch to be neutral_interaction_point, got %q", categoryByID["caravan_dispatch_village"])
	}
}

func TestProfessionChangeResponseIncludesBotFriendlySummary(t *testing.T) {
	server := NewServer(config.API{Port: "8080"})

	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/register", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-prof-summary",
		"password": "verysecure",
	}), "", http.StatusOK, nil)

	var loginResponse struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/login", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-prof-summary",
		"password": "verysecure",
	}), "", http.StatusOK, &loginResponse)

	doJSONRequest(t, server, http.MethodPost, "/api/v1/characters", map[string]any{
		"name": "ProfSummary",
	}, loginResponse.Data.AccessToken, http.StatusOK, nil)

	account, err := server.authService.Authenticate(loginResponse.Data.AccessToken)
	if err != nil {
		t.Fatalf("failed to authenticate test account: %v", err)
	}
	summary, ok := server.characterService.GetSummary(account.AccountID)
	if !ok {
		t.Fatal("expected character summary")
	}
	if _, err := server.characterService.GrantSeasonXP(summary.CharacterID, 5000); err != nil {
		t.Fatalf("failed to grant season xp: %v", err)
	}
	if _, err := server.characterService.GrantGold(summary.CharacterID, 900); err != nil {
		t.Fatalf("failed to grant gold: %v", err)
	}

	var professionResponse struct {
		Data struct {
			Character struct {
				Class       string `json:"class"`
				WeaponStyle string `json:"weapon_style"`
				Gold        int    `json:"gold"`
			} `json:"character"`
			ProfessionChangeResult struct {
				RequestedClass        string   `json:"requested_class"`
				FromClass             string   `json:"from_class"`
				ToClass               string   `json:"to_class"`
				GoldCost              int      `json:"gold_cost"`
				GoldBefore            int      `json:"gold_before"`
				GoldAfter             int      `json:"gold_after"`
				SkillLevelsPreserved  bool     `json:"skill_levels_preserved"`
				StarterWeaponGranted  bool     `json:"starter_weapon_granted"`
				StarterWeaponEquipped bool     `json:"starter_weapon_equipped"`
				Warnings              []string `json:"warnings"`
			} `json:"profession_change_result"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/profession", map[string]any{
		"class_id": "warrior",
	}, loginResponse.Data.AccessToken, http.StatusOK, &professionResponse)

	if professionResponse.Data.Character.Class != "warrior" {
		t.Fatalf("expected warrior class after profession change, got %q", professionResponse.Data.Character.Class)
	}
	if professionResponse.Data.Character.WeaponStyle != "sword_shield" {
		t.Fatalf("expected starter weapon style sword_shield, got %q", professionResponse.Data.Character.WeaponStyle)
	}
	if professionResponse.Data.ProfessionChangeResult.RequestedClass != "warrior" {
		t.Fatalf("expected requested_class warrior, got %q", professionResponse.Data.ProfessionChangeResult.RequestedClass)
	}
	if professionResponse.Data.ProfessionChangeResult.FromClass != "civilian" || professionResponse.Data.ProfessionChangeResult.ToClass != "warrior" {
		t.Fatalf("expected civilian -> warrior summary, got %#v", professionResponse.Data.ProfessionChangeResult)
	}
	if professionResponse.Data.ProfessionChangeResult.GoldCost != characters.ProfessionChangeGoldCost {
		t.Fatalf("expected profession gold cost %d, got %d", characters.ProfessionChangeGoldCost, professionResponse.Data.ProfessionChangeResult.GoldCost)
	}
	if professionResponse.Data.ProfessionChangeResult.GoldBefore != 1000 || professionResponse.Data.ProfessionChangeResult.GoldAfter != 200 {
		t.Fatalf("expected gold before/after 1000 -> 200, got %d -> %d", professionResponse.Data.ProfessionChangeResult.GoldBefore, professionResponse.Data.ProfessionChangeResult.GoldAfter)
	}
	if !professionResponse.Data.ProfessionChangeResult.SkillLevelsPreserved {
		t.Fatal("expected skill preservation flag")
	}
	if !professionResponse.Data.ProfessionChangeResult.StarterWeaponGranted || !professionResponse.Data.ProfessionChangeResult.StarterWeaponEquipped {
		t.Fatalf("expected starter weapon grant+equip summary, got %#v", professionResponse.Data.ProfessionChangeResult)
	}
	if len(professionResponse.Data.ProfessionChangeResult.Warnings) != 0 {
		t.Fatalf("expected no warnings for first promotion, got %#v", professionResponse.Data.ProfessionChangeResult.Warnings)
	}
}

func TestProfessionChangeErrorIncludesBotFriendlyDetails(t *testing.T) {
	server := NewServer(config.API{Port: "8080"})

	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/register", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-prof-error",
		"password": "verysecure",
	}), "", http.StatusOK, nil)

	var loginResponse struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/login", withAuthChallenge(t, server, map[string]any{
		"bot_name": "bot-prof-error",
		"password": "verysecure",
	}), "", http.StatusOK, &loginResponse)

	doJSONRequest(t, server, http.MethodPost, "/api/v1/characters", map[string]any{
		"name": "ProfError",
	}, loginResponse.Data.AccessToken, http.StatusOK, nil)

	account, err := server.authService.Authenticate(loginResponse.Data.AccessToken)
	if err != nil {
		t.Fatalf("failed to authenticate test account: %v", err)
	}
	summary, ok := server.characterService.GetSummary(account.AccountID)
	if !ok {
		t.Fatal("expected character summary")
	}
	if _, err := server.characterService.GrantSeasonXP(summary.CharacterID, 5000); err != nil {
		t.Fatalf("failed to grant season xp: %v", err)
	}

	var errorResponse struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
			Details struct {
				RequestedClass string   `json:"requested_class"`
				CurrentClass   string   `json:"current_class"`
				CurrentGold    int      `json:"current_gold"`
				RequiredGold   int      `json:"required_gold"`
				GoldShortfall  int      `json:"gold_shortfall"`
				SeasonLevel    int      `json:"season_level"`
				RequiredLevel  int      `json:"required_level"`
				ReasonHint     string   `json:"reason_hint"`
				Supported      []string `json:"supported_classes"`
			} `json:"details"`
		} `json:"error"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/profession", map[string]any{
		"class_id": "mage",
	}, loginResponse.Data.AccessToken, http.StatusBadRequest, &errorResponse)

	if errorResponse.Error.Code != "CHARACTER_PROFESSION_GOLD_INSUFFICIENT" {
		t.Fatalf("expected gold-insufficient profession error, got %q", errorResponse.Error.Code)
	}
	if errorResponse.Error.Details.RequestedClass != "mage" || errorResponse.Error.Details.CurrentClass != "civilian" {
		t.Fatalf("expected requested mage from civilian details, got %#v", errorResponse.Error.Details)
	}
	if errorResponse.Error.Details.CurrentGold != 100 || errorResponse.Error.Details.RequiredGold != characters.ProfessionChangeGoldCost {
		t.Fatalf("expected current/required gold 100/%d, got %d/%d", characters.ProfessionChangeGoldCost, errorResponse.Error.Details.CurrentGold, errorResponse.Error.Details.RequiredGold)
	}
	if errorResponse.Error.Details.GoldShortfall != characters.ProfessionChangeGoldCost-100 {
		t.Fatalf("expected shortfall %d, got %d", characters.ProfessionChangeGoldCost-100, errorResponse.Error.Details.GoldShortfall)
	}
	if errorResponse.Error.Details.SeasonLevel != 10 || errorResponse.Error.Details.RequiredLevel != 10 {
		t.Fatalf("expected level details 10/10, got %d/%d", errorResponse.Error.Details.SeasonLevel, errorResponse.Error.Details.RequiredLevel)
	}
	if errorResponse.Error.Details.ReasonHint != "gold_insufficient" {
		t.Fatalf("expected reason_hint gold_insufficient, got %q", errorResponse.Error.Details.ReasonHint)
	}
	if len(errorResponse.Error.Details.Supported) != 4 {
		t.Fatalf("expected supported class list in details, got %#v", errorResponse.Error.Details.Supported)
	}
}

func decodeJSON(t *testing.T, recorder *httptest.ResponseRecorder, target any) {
	t.Helper()

	if err := json.Unmarshal(recorder.Body.Bytes(), target); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
}

type questPlannerHintE2E struct {
	QuestID             string         `json:"quest_id"`
	CurrentStepKey      string         `json:"current_step_key"`
	CurrentStepLabel    string         `json:"current_step_label"`
	CurrentStepHint     string         `json:"current_step_hint"`
	SuggestedActionType string         `json:"suggested_action_type"`
	SuggestedActionArgs map[string]any `json:"suggested_action_args"`
	TargetRegionID      string         `json:"target_region_id"`
}

type plannerE2EView struct {
	Data struct {
		SuggestedActions []string              `json:"suggested_actions"`
		RuntimeHints     []questPlannerHintE2E `json:"quest_runtime_hints"`
	} `json:"data"`
}

type actionSchemaView struct {
	ActionType string         `json:"action_type"`
	ArgsSchema map[string]any `json:"args_schema"`
}

type boardQuestView struct {
	QuestID        string `json:"quest_id"`
	TemplateType   string `json:"template_type"`
	ContractType   string `json:"contract_type"`
	Difficulty     string `json:"difficulty"`
	FlowKind       string `json:"flow_kind"`
	TargetRegionID string `json:"target_region_id"`
	Status         string `json:"status"`
}

func createQuestE2ECharacter(t *testing.T, server *Server, botName, characterName string) string {
	t.Helper()

	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/register", withAuthChallenge(t, server, map[string]any{
		"bot_name": botName,
		"password": "verysecure",
	}), "", http.StatusOK, nil)

	var loginResponse struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/login", withAuthChallenge(t, server, map[string]any{
		"bot_name": botName,
		"password": "verysecure",
	}), "", http.StatusOK, &loginResponse)

	doJSONRequest(t, server, http.MethodPost, "/api/v1/characters", map[string]any{
		"name": characterName,
	}, loginResponse.Data.AccessToken, http.StatusOK, nil)

	account, err := server.authService.Authenticate(loginResponse.Data.AccessToken)
	if err != nil {
		t.Fatalf("failed to authenticate quest e2e helper token: %v", err)
	}
	summary, ok := server.characterService.GetCharacterByAccount(account)
	if !ok {
		t.Fatal("expected helper character to exist")
	}
	if _, err := server.characterService.GrantSeasonXP(summary.CharacterID, 5000); err != nil {
		t.Fatalf("failed to grant season xp for quest e2e helper: %v", err)
	}
	if _, err := server.characterService.GrantGold(summary.CharacterID, characters.ProfessionChangeGoldCost); err != nil {
		t.Fatalf("failed to grant profession change gold for quest e2e helper: %v", err)
	}

	doJSONRequest(t, server, http.MethodPost, "/api/v1/me/profession-route", map[string]any{
		"route_id": "aoe_burst",
	}, loginResponse.Data.AccessToken, http.StatusOK, nil)
	server.questService.ResetDailyQuestBoard(summary.CharacterID)

	return loginResponse.Data.AccessToken
}

func createCharacterWithBoard(t *testing.T, server *Server, botPrefix, characterPrefix string, matcher func([]boardQuestView) bool) (string, []boardQuestView) {
	t.Helper()

	for attempt := 0; attempt < 24; attempt++ {
		token := createQuestE2ECharacter(t, server, botPrefix+"-"+strconv.Itoa(attempt), characterPrefix+strconv.Itoa(attempt))

		var boardResponse struct {
			Data struct {
				Quests []boardQuestView `json:"quests"`
			} `json:"data"`
		}
		doJSONRequest(t, server, http.MethodGet, "/api/v1/me/quests", nil, token, http.StatusOK, &boardResponse)

		if matcher(boardResponse.Data.Quests) {
			return token, boardResponse.Data.Quests
		}
	}

	t.Fatalf("failed to find a quest board matching predicate for %s", botPrefix)
	return "", nil
}

func getPlannerE2EView(t *testing.T, server *Server, token, regionID string) plannerE2EView {
	t.Helper()

	var plannerResponse plannerE2EView
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/planner?region_id="+regionID, nil, token, http.StatusOK, &plannerResponse)
	return plannerResponse
}

func findQuestRuntimeHint(t *testing.T, hints []questPlannerHintE2E, questID string) questPlannerHintE2E {
	t.Helper()

	for _, hint := range hints {
		if hint.QuestID == questID {
			return hint
		}
	}
	t.Fatalf("expected runtime hint for quest %q, got %#v", questID, hints)
	return questPlannerHintE2E{}
}

func findActionByType(t *testing.T, actions []actionSchemaView, actionType string) actionSchemaView {
	t.Helper()

	for _, action := range actions {
		if action.ActionType == actionType {
			return action
		}
	}
	t.Fatalf("expected action %q in %#v", actionType, actions)
	return actionSchemaView{}
}

func decodeChoiceOptions(t *testing.T, raw any) []struct {
	ChoiceKey string
	Label     string
} {
	t.Helper()

	items, ok := raw.([]any)
	if !ok {
		return nil
	}
	choices := make([]struct {
		ChoiceKey string
		Label     string
	}, 0, len(items))
	for _, item := range items {
		payload, ok := item.(map[string]any)
		if !ok {
			continue
		}
		choices = append(choices, struct {
			ChoiceKey string
			Label     string
		}{
			ChoiceKey: asStringValue(payload["choice_key"]),
			Label:     asStringValue(payload["label"]),
		})
	}
	return choices
}

func assertQuestStatus(t *testing.T, server *Server, token, questID, expectedStatus string) {
	t.Helper()

	var questsResponse struct {
		Data struct {
			Quests []struct {
				QuestID string `json:"quest_id"`
				Status  string `json:"status"`
			} `json:"quests"`
		} `json:"data"`
	}
	doJSONRequest(t, server, http.MethodGet, "/api/v1/me/quests", nil, token, http.StatusOK, &questsResponse)

	for _, quest := range questsResponse.Data.Quests {
		if quest.QuestID == questID {
			if quest.Status != expectedStatus {
				t.Fatalf("expected quest %q status %q, got %q", questID, expectedStatus, quest.Status)
			}
			return
		}
	}
	t.Fatalf("expected quest %q in quest board", questID)
}

func asStringValue(value any) string {
	text, _ := value.(string)
	return text
}

func doJSONRequest(t *testing.T, server *Server, method, path string, body any, bearerToken string, expectedStatus int, target any) {
	t.Helper()

	legacyClassID := ""
	if method == http.MethodPost && path == "/api/v1/characters" && expectedStatus == http.StatusOK {
		if payloadMap, ok := body.(map[string]any); ok {
			className := strings.TrimSpace(asStringValue(payloadMap["class"]))
			if className != "" && className != "civilian" {
				legacyClassID = className
				body = map[string]any{
					"name": payloadMap["name"],
				}
			}
		}
	}

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

	if legacyClassID != "" {
		account, err := server.authService.Authenticate(bearerToken)
		if err != nil {
			t.Fatalf("failed to authenticate legacy create flow in test helper: %v", err)
		}
		summary, ok := server.characterService.GetCharacterByAccount(account)
		if !ok {
			t.Fatal("expected character after creation in legacy create flow")
		}
		if _, err := server.characterService.GrantSeasonXP(summary.CharacterID, 5000); err != nil {
			t.Fatalf("failed to grant season xp in legacy create flow: %v", err)
		}
		if _, err := server.characterService.GrantGold(summary.CharacterID, characters.ProfessionChangeGoldCost); err != nil {
			t.Fatalf("failed to grant profession change gold in legacy create flow: %v", err)
		}

		doJSONRequest(t, server, http.MethodPost, "/api/v1/me/profession-route", map[string]any{
			"class_id": legacyClassID,
		}, bearerToken, http.StatusOK, nil)
		return
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

func dungeonEnterPayload(difficulty string) map[string]any {
	payload := map[string]any{
		"potion_loadout": []string{"potion_hp_t2", "potion_atk_t2"},
	}
	if strings.TrimSpace(difficulty) != "" {
		payload["difficulty"] = difficulty
	}
	return payload
}
