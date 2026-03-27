package app

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"clawgame/apps/api/internal/auth"
	"clawgame/apps/api/internal/characters"
	"clawgame/apps/api/internal/platform/config"
	"clawgame/apps/api/internal/platform/store"
	"clawgame/apps/api/internal/quests"
	"clawgame/apps/api/internal/world"
	"github.com/go-chi/chi/v5"
)

type Server struct {
	httpServer       *http.Server
	authService      *auth.Service
	characterService *characters.Service
	questService     *quests.Service
	worldService     *world.Service
}

var requestCounter uint64

func NewServer(cfg config.API) *Server {
	router := chi.NewRouter()
	authService := auth.NewService()
	characterService := characters.NewService()
	questService := quests.NewService()
	worldService := world.NewService()

	if strings.TrimSpace(cfg.DatabaseURL) != "" {
		postgresStore, err := store.NewPostgresStore(cfg.DatabaseURL)
		if err != nil {
			log.Printf("api persistence disabled: failed to connect postgres: %v", err)
		} else {
			persistentAuth, err := auth.NewServiceWithRepository(postgresStore)
			if err != nil {
				log.Printf("api persistence disabled: failed to load accounts: %v", err)
			} else {
				persistentCharacters, err := characters.NewServiceWithRepository(postgresStore)
				if err != nil {
					log.Printf("api persistence disabled: failed to load characters: %v", err)
				} else {
					persistentQuests, err := quests.NewServiceWithRepository(postgresStore)
					if err != nil {
						log.Printf("api persistence disabled: failed to load quest boards: %v", err)
					} else {
						authService = persistentAuth
						characterService = persistentCharacters
						questService = persistentQuests
						log.Printf("api persistence enabled with postgres runtime storage")
					}
				}
			}
		}
	}

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		writeEnvelope(w, r, http.StatusOK, map[string]any{
			"service": "clawgame-api",
			"status":  "ok",
			"version": "v1",
		})
	})

	router.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeEnvelope(w, r, http.StatusOK, map[string]any{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	router.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/challenge", func(w http.ResponseWriter, r *http.Request) {
			challenge, err := authService.IssueChallenge()
			if err != nil {
				writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to issue auth challenge")
				return
			}

			writeEnvelope(w, r, http.StatusOK, map[string]any{
				"challenge": challenge,
			})
		})

		r.Post("/auth/register", func(w http.ResponseWriter, r *http.Request) {
			var request struct {
				BotName         string `json:"bot_name"`
				Password        string `json:"password"`
				ChallengeID     string `json:"challenge_id"`
				ChallengeAnswer string `json:"challenge_answer"`
			}
			if !decodeJSONBody(w, r, &request) {
				return
			}

			account, err := authService.RegisterAccount(
				request.BotName,
				request.Password,
				request.ChallengeID,
				request.ChallengeAnswer,
			)
			if err != nil {
				switch {
				case errors.Is(err, auth.ErrChallengeRequired):
					writeError(w, r, http.StatusBadRequest, "AUTH_CHALLENGE_REQUIRED", "register requests must answer a fresh auth challenge")
				case errors.Is(err, auth.ErrChallengeNotFound):
					writeError(w, r, http.StatusUnauthorized, "AUTH_CHALLENGE_NOT_FOUND", "auth challenge does not exist")
				case errors.Is(err, auth.ErrChallengeExpired):
					writeError(w, r, http.StatusUnauthorized, "AUTH_CHALLENGE_EXPIRED", "auth challenge has expired")
				case errors.Is(err, auth.ErrChallengeUsed):
					writeError(w, r, http.StatusUnauthorized, "AUTH_CHALLENGE_USED", "auth challenge has already been consumed")
				case errors.Is(err, auth.ErrChallengeInvalid):
					writeError(w, r, http.StatusUnauthorized, "AUTH_CHALLENGE_INVALID", "auth challenge answer is incorrect")
				case errors.Is(err, auth.ErrBotNameTaken):
					writeError(w, r, http.StatusConflict, "ACCOUNT_BOT_NAME_TAKEN", "bot name is already registered")
				case errors.Is(err, auth.ErrInvalidRegisterInput):
					writeError(w, r, http.StatusBadRequest, "ACCOUNT_INVALID_INPUT", "bot name or password does not satisfy the required constraints")
				default:
					writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to register account")
				}
				return
			}

			writeEnvelope(w, r, http.StatusOK, map[string]any{
				"account": account,
			})
		})

		r.Post("/auth/login", func(w http.ResponseWriter, r *http.Request) {
			var request struct {
				BotName         string `json:"bot_name"`
				Password        string `json:"password"`
				ChallengeID     string `json:"challenge_id"`
				ChallengeAnswer string `json:"challenge_answer"`
			}
			if !decodeJSONBody(w, r, &request) {
				return
			}

			tokens, err := authService.Login(
				request.BotName,
				request.Password,
				request.ChallengeID,
				request.ChallengeAnswer,
			)
			if err != nil {
				switch {
				case errors.Is(err, auth.ErrChallengeRequired):
					writeError(w, r, http.StatusBadRequest, "AUTH_CHALLENGE_REQUIRED", "login requests must answer a fresh auth challenge")
				case errors.Is(err, auth.ErrChallengeNotFound):
					writeError(w, r, http.StatusUnauthorized, "AUTH_CHALLENGE_NOT_FOUND", "auth challenge does not exist")
				case errors.Is(err, auth.ErrChallengeExpired):
					writeError(w, r, http.StatusUnauthorized, "AUTH_CHALLENGE_EXPIRED", "auth challenge has expired")
				case errors.Is(err, auth.ErrChallengeUsed):
					writeError(w, r, http.StatusUnauthorized, "AUTH_CHALLENGE_USED", "auth challenge has already been consumed")
				case errors.Is(err, auth.ErrChallengeInvalid):
					writeError(w, r, http.StatusUnauthorized, "AUTH_CHALLENGE_INVALID", "auth challenge answer is incorrect")
				case errors.Is(err, auth.ErrInvalidCredentials):
					writeError(w, r, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "bot name or password is incorrect")
				default:
					writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create session")
				}
				return
			}

			writeEnvelope(w, r, http.StatusOK, map[string]any{
				"access_token":             tokens.AccessToken,
				"access_token_expires_at":  tokens.AccessTokenExpiresAt,
				"refresh_token":            tokens.RefreshToken,
				"refresh_token_expires_at": tokens.RefreshTokenExpiresAt,
			})
		})

		r.Post("/auth/refresh", func(w http.ResponseWriter, r *http.Request) {
			var request struct {
				RefreshToken string `json:"refresh_token"`
			}
			if !decodeJSONBody(w, r, &request) {
				return
			}

			tokens, err := authService.RefreshSession(request.RefreshToken)
			if err != nil {
				switch {
				case errors.Is(err, auth.ErrRefreshTokenExpired):
					writeError(w, r, http.StatusUnauthorized, "AUTH_TOKEN_EXPIRED", "refresh token has expired")
				case errors.Is(err, auth.ErrSessionNotFound):
					writeError(w, r, http.StatusUnauthorized, "AUTH_REQUIRED", "refresh token is invalid")
				default:
					writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to refresh session")
				}
				return
			}

			writeEnvelope(w, r, http.StatusOK, map[string]any{
				"access_token":             tokens.AccessToken,
				"access_token_expires_at":  tokens.AccessTokenExpiresAt,
				"refresh_token":            tokens.RefreshToken,
				"refresh_token_expires_at": tokens.RefreshTokenExpiresAt,
			})
		})

		r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
			writeEnvelope(w, r, http.StatusOK, map[string]any{
				"service": "clawgame-api",
				"status":  "ok",
			})
		})

		r.Post("/characters", func(w http.ResponseWriter, r *http.Request) {
			account, ok := requireAccount(w, r, authService)
			if !ok {
				return
			}

			var request struct {
				Name        string `json:"name"`
				Class       string `json:"class"`
				WeaponStyle string `json:"weapon_style"`
			}
			if !decodeJSONBody(w, r, &request) {
				return
			}

			state, err := characterService.CreateCharacter(account, request.Name, request.Class, request.WeaponStyle, worldService)
			if err != nil {
				switch {
				case errors.Is(err, characters.ErrCharacterAlreadyExists):
					writeError(w, r, http.StatusConflict, "CHARACTER_ALREADY_EXISTS", "account already owns a character")
				case errors.Is(err, characters.ErrCharacterInvalidClass):
					writeError(w, r, http.StatusBadRequest, "CHARACTER_INVALID_CLASS", "class is not supported")
				case errors.Is(err, characters.ErrCharacterInvalidWeapon):
					writeError(w, r, http.StatusBadRequest, "CHARACTER_INVALID_WEAPON_STYLE", "weapon style is incompatible with class")
				case errors.Is(err, characters.ErrCharacterNameTaken):
					writeError(w, r, http.StatusConflict, "CHARACTER_NAME_TAKEN", "character name is already in use")
				case errors.Is(err, characters.ErrCharacterInvalidName):
					writeError(w, r, http.StatusBadRequest, "CHARACTER_INVALID_NAME", "character name must be between 3 and 32 characters")
				default:
					writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create character")
				}
				return
			}

			questService.EnsureDailyQuestBoard(state.Character)
			state, err = buildCharacterState(account, characterService, questService, worldService)
			if err != nil {
				writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load character state")
				return
			}

			writeEnvelope(w, r, http.StatusOK, state)
		})

		r.Get("/me", func(w http.ResponseWriter, r *http.Request) {
			account, ok := requireAccount(w, r, authService)
			if !ok {
				return
			}

			writeEnvelope(w, r, http.StatusOK, characterService.GetMe(account))
		})

		r.Get("/me/state", func(w http.ResponseWriter, r *http.Request) {
			account, ok := requireAccount(w, r, authService)
			if !ok {
				return
			}

			state, err := characterService.GetState(account, worldService)
			if err != nil {
				if errors.Is(err, characters.ErrCharacterNotFound) {
					writeError(w, r, http.StatusNotFound, "CHARACTER_NOT_FOUND", "create a character before requesting full state")
					return
				}

				writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load character state")
				return
			}
			_ = state

			state, err = buildCharacterState(account, characterService, questService, worldService)
			if err != nil {
				if errors.Is(err, characters.ErrCharacterNotFound) {
					writeError(w, r, http.StatusNotFound, "CHARACTER_NOT_FOUND", "create a character before requesting full state")
					return
				}

				writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load character state")
				return
			}

			writeEnvelope(w, r, http.StatusOK, state)
		})

		r.Get("/me/actions", func(w http.ResponseWriter, r *http.Request) {
			account, ok := requireAccount(w, r, authService)
			if !ok {
				return
			}

			actions, err := characterService.ListValidActions(account, worldService)
			if err != nil {
				if errors.Is(err, characters.ErrCharacterNotFound) {
					writeError(w, r, http.StatusNotFound, "CHARACTER_NOT_FOUND", "create a character before requesting actions")
					return
				}

				writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load valid actions")
				return
			}

			writeEnvelope(w, r, http.StatusOK, map[string]any{
				"actions": actions,
			})
		})

		r.Post("/me/actions", func(w http.ResponseWriter, r *http.Request) {
			account, ok := requireAccount(w, r, authService)
			if !ok {
				return
			}

			var request struct {
				ActionType string         `json:"action_type"`
				ActionArgs map[string]any `json:"action_args"`
			}
			if !decodeJSONBody(w, r, &request) {
				return
			}

			result, err := executeAction(account, request.ActionType, request.ActionArgs, characterService, questService, worldService)
			if err != nil {
				switch {
				case errors.Is(err, characters.ErrCharacterNotFound):
					writeError(w, r, http.StatusNotFound, "CHARACTER_NOT_FOUND", "create a character before executing actions")
				case errors.Is(err, quests.ErrQuestNotFound):
					writeError(w, r, http.StatusNotFound, "QUEST_NOT_FOUND", "quest does not exist on the current board")
				case errors.Is(err, quests.ErrQuestInvalidState):
					writeError(w, r, http.StatusBadRequest, "QUEST_INVALID_STATE", "quest is not in a valid state for this action")
				case errors.Is(err, quests.ErrQuestCompletionCapReached):
					writeError(w, r, http.StatusBadRequest, "QUEST_COMPLETION_CAP_REACHED", "daily quest completion cap has been reached")
				case errors.Is(err, quests.ErrQuestRerollConfirmRequired):
					writeError(w, r, http.StatusBadRequest, "QUEST_REROLL_CONFIRM_REQUIRED", "reroll requests must confirm the gold cost")
				case errors.Is(err, characters.ErrGoldInsufficient):
					writeError(w, r, http.StatusBadRequest, "GOLD_INSUFFICIENT", "character does not have enough gold for this action")
				case errors.Is(err, characters.ErrTravelRegionNotFound):
					writeError(w, r, http.StatusNotFound, "TRAVEL_REGION_NOT_FOUND", "target region does not exist")
				case errors.Is(err, characters.ErrTravelRankLocked):
					writeError(w, r, http.StatusBadRequest, "TRAVEL_RANK_LOCKED", "character rank does not unlock this region")
				case errors.Is(err, characters.ErrTravelInsufficientGold):
					writeError(w, r, http.StatusBadRequest, "TRAVEL_INSUFFICIENT_GOLD", "character does not have enough gold to travel")
				case errors.Is(err, characters.ErrActionNotSupported):
					writeError(w, r, http.StatusBadRequest, "ACTION_NOT_SUPPORTED", "action type is not currently supported")
				default:
					writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to execute action")
				}
				return
			}

			writeEnvelope(w, r, http.StatusOK, result)
		})

		r.Post("/me/travel", func(w http.ResponseWriter, r *http.Request) {
			account, ok := requireAccount(w, r, authService)
			if !ok {
				return
			}

			var request struct {
				RegionID string `json:"region_id"`
			}
			if !decodeJSONBody(w, r, &request) {
				return
			}

			result, err := characterService.Travel(account, request.RegionID, worldService)
			if err != nil {
				switch {
				case errors.Is(err, characters.ErrCharacterNotFound):
					writeError(w, r, http.StatusNotFound, "CHARACTER_NOT_FOUND", "create a character before travelling")
				case errors.Is(err, characters.ErrTravelRegionNotFound):
					writeError(w, r, http.StatusNotFound, "TRAVEL_REGION_NOT_FOUND", "target region does not exist")
				case errors.Is(err, characters.ErrTravelRankLocked):
					writeError(w, r, http.StatusBadRequest, "TRAVEL_RANK_LOCKED", "character rank does not unlock this region")
				case errors.Is(err, characters.ErrTravelInsufficientGold):
					writeError(w, r, http.StatusBadRequest, "TRAVEL_INSUFFICIENT_GOLD", "character does not have enough gold to travel")
				default:
					writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to travel")
				}
				return
			}

			limits := result.State.Limits
			board, completed := questService.ProgressTravelQuests(result.State.Character, request.RegionID, limits)
			for _, quest := range completed {
				_ = characterService.AppendEvents(result.State.Character.CharacterID, world.WorldEvent{
					EventID:          requestID(r),
					EventType:        "quest.completed",
					Visibility:       "public",
					ActorCharacterID: result.State.Character.CharacterID,
					ActorName:        result.State.Character.Name,
					RegionID:         request.RegionID,
					Summary:          fmt.Sprintf("%s completed %s.", result.State.Character.Name, quest.Title),
					Payload: map[string]any{
						"quest_id":    quest.QuestID,
						"quest_title": quest.Title,
					},
					OccurredAt: time.Now().Format(time.RFC3339),
				})
			}

			state, err := buildCharacterState(account, characterService, questService, worldService)
			if err != nil {
				writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load updated state after travel")
				return
			}
			state.Objectives = activeObjectivesFromBoard(board)

			writeEnvelope(w, r, http.StatusOK, map[string]any{
				"action_result": map[string]any{
					"action_type":      "travel",
					"from_region_id":   result.FromRegionID,
					"to_region_id":     result.ToRegionID,
					"travel_cost_gold": result.TravelCostGold,
				},
				"state": state,
			})
		})

		r.Get("/me/quests", func(w http.ResponseWriter, r *http.Request) {
			account, ok := requireAccount(w, r, authService)
			if !ok {
				return
			}

			character, limits, err := currentCharacterWithLimits(account, characterService, worldService)
			if err != nil {
				writeError(w, r, http.StatusNotFound, "CHARACTER_NOT_FOUND", "create a character before requesting quests")
				return
			}

			writeEnvelope(w, r, http.StatusOK, questService.ListQuests(character, limits))
		})

		r.Post("/me/quests/{questId}/accept", func(w http.ResponseWriter, r *http.Request) {
			account, ok := requireAccount(w, r, authService)
			if !ok {
				return
			}

			character, limits, err := currentCharacterWithLimits(account, characterService, worldService)
			if err != nil {
				writeError(w, r, http.StatusNotFound, "CHARACTER_NOT_FOUND", "create a character before accepting quests")
				return
			}

			board, quest, err := questService.AcceptQuest(character, chi.URLParam(r, "questId"), limits)
			if err != nil {
				switch {
				case errors.Is(err, quests.ErrQuestNotFound):
					writeError(w, r, http.StatusNotFound, "QUEST_NOT_FOUND", "quest does not exist on the current board")
				default:
					writeError(w, r, http.StatusBadRequest, "QUEST_INVALID_STATE", "quest is not available for acceptance")
				}
				return
			}

			_ = characterService.AppendEvents(character.CharacterID, world.WorldEvent{
				EventID:          requestID(r),
				EventType:        "quest.accepted",
				Visibility:       "public",
				ActorCharacterID: character.CharacterID,
				ActorName:        character.Name,
				RegionID:         character.LocationRegionID,
				Summary:          fmt.Sprintf("%s accepted %s.", character.Name, quest.Title),
				Payload: map[string]any{
					"quest_id":    quest.QuestID,
					"quest_title": quest.Title,
				},
				OccurredAt: time.Now().Format(time.RFC3339),
			})

			writeEnvelope(w, r, http.StatusOK, map[string]any{
				"action_result": map[string]any{
					"action_type": "accept_quest",
					"quest_id":    quest.QuestID,
					"status":      "accepted",
				},
				"state": map[string]any{
					"quests": board.Quests,
					"limits": board.Limits,
				},
			})
		})

		r.Post("/me/quests/{questId}/submit", func(w http.ResponseWriter, r *http.Request) {
			account, ok := requireAccount(w, r, authService)
			if !ok {
				return
			}

			character, limits, err := currentCharacterWithLimits(account, characterService, worldService)
			if err != nil {
				writeError(w, r, http.StatusNotFound, "CHARACTER_NOT_FOUND", "create a character before submitting quests")
				return
			}

			quest, err := questService.SubmitQuest(character, chi.URLParam(r, "questId"), limits)
			if err != nil {
				switch {
				case errors.Is(err, quests.ErrQuestNotFound):
					writeError(w, r, http.StatusNotFound, "QUEST_NOT_FOUND", "quest does not exist on the current board")
				case errors.Is(err, quests.ErrQuestCompletionCapReached):
					writeError(w, r, http.StatusBadRequest, "QUEST_COMPLETION_CAP_REACHED", "daily quest completion cap has been reached")
				default:
					writeError(w, r, http.StatusBadRequest, "QUEST_INVALID_STATE", "quest is not ready for submission")
				}
				return
			}

			_, _, _, _, err = characterService.ApplyQuestSubmission(character.CharacterID, quest)
			if err != nil {
				writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to apply quest rewards")
				return
			}

			state, err := buildCharacterState(account, characterService, questService, worldService)
			if err != nil {
				writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load updated state after quest submission")
				return
			}

			writeEnvelope(w, r, http.StatusOK, map[string]any{
				"action_result": map[string]any{
					"action_type":       "submit_quest",
					"quest_id":          quest.QuestID,
					"reward_gold":       quest.RewardGold,
					"reward_reputation": quest.RewardReputation,
				},
				"state": state,
			})
		})

		r.Post("/me/quests/reroll", func(w http.ResponseWriter, r *http.Request) {
			account, ok := requireAccount(w, r, authService)
			if !ok {
				return
			}

			var request struct {
				ConfirmCost bool `json:"confirm_cost"`
			}
			if !decodeJSONBody(w, r, &request) {
				return
			}
			if !request.ConfirmCost {
				writeError(w, r, http.StatusBadRequest, "QUEST_REROLL_CONFIRM_REQUIRED", "reroll requests must confirm the gold cost")
				return
			}

			character, limits, err := currentCharacterWithLimits(account, characterService, worldService)
			if err != nil {
				writeError(w, r, http.StatusNotFound, "CHARACTER_NOT_FOUND", "create a character before rerolling quests")
				return
			}

			if _, err := characterService.SpendGold(character.CharacterID, quests.RerollCostGold()); err != nil {
				writeError(w, r, http.StatusBadRequest, "GOLD_INSUFFICIENT", "character does not have enough gold to reroll quests")
				return
			}

			character, limits, err = currentCharacterWithLimits(account, characterService, worldService)
			if err != nil {
				writeError(w, r, http.StatusNotFound, "CHARACTER_NOT_FOUND", "create a character before rerolling quests")
				return
			}

			board, err := questService.RerollQuestBoard(character, limits, request.ConfirmCost)
			if err != nil {
				if errors.Is(err, quests.ErrQuestRerollConfirmRequired) {
					writeError(w, r, http.StatusBadRequest, "QUEST_REROLL_CONFIRM_REQUIRED", "reroll requests must confirm the gold cost")
					return
				}

				writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to reroll quest board")
				return
			}

			writeEnvelope(w, r, http.StatusOK, board)
		})

		r.Get("/world/regions", func(w http.ResponseWriter, r *http.Request) {
			writeEnvelope(w, r, http.StatusOK, map[string]any{
				"regions": worldService.ListRegions(),
			})
		})

		r.Get("/regions/{regionId}", func(w http.ResponseWriter, r *http.Request) {
			regionID := chi.URLParam(r, "regionId")
			detail, ok := worldService.GetRegion(regionID)
			if !ok {
				writeError(w, r, http.StatusNotFound, "REGION_NOT_FOUND", fmt.Sprintf("region %q does not exist", regionID))
				return
			}

			writeEnvelope(w, r, http.StatusOK, detail)
		})

		r.Route("/public", func(r chi.Router) {
			r.Get("/world-state", func(w http.ResponseWriter, r *http.Request) {
				snapshot := characterService.SnapshotRuntime()
				writeEnvelope(w, r, http.StatusOK, buildPublicWorldState(worldService, snapshot))
			})

			r.Get("/events", func(w http.ResponseWriter, r *http.Request) {
				limit := parseLimit(r, 20, 100)
				snapshot := characterService.SnapshotRuntime()
				writeEnvelope(w, r, http.StatusOK, map[string]any{
					"items":       buildPublicEvents(snapshot.Events, limit),
					"next_cursor": nil,
				})
			})

			r.Get("/leaderboards", func(w http.ResponseWriter, r *http.Request) {
				snapshot := characterService.SnapshotRuntime()
				writeEnvelope(w, r, http.StatusOK, buildPublicLeaderboards(snapshot))
			})
		})
	})

	return &Server{
		authService:      authService,
		characterService: characterService,
		questService:     questService,
		worldService:     worldService,
		httpServer: &http.Server{
			Addr:              cfg.Addr(),
			Handler:           router,
			ReadHeaderTimeout: 5 * time.Second,
		},
	}
}

func (s *Server) ListenAndServe() error {
	return s.httpServer.ListenAndServe()
}

func writeEnvelope(w http.ResponseWriter, r *http.Request, status int, data any) {
	writeJSON(w, status, map[string]any{
		"request_id": requestID(r),
		"data":       data,
	})
}

func writeError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	writeJSON(w, status, map[string]any{
		"request_id": requestID(r),
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func requestID(r *http.Request) string {
	return fmt.Sprintf("req_%d", atomic.AddUint64(&requestCounter, 1))
}

func parseLimit(r *http.Request, defaultValue, maxValue int) int {
	raw := r.URL.Query().Get("limit")
	if raw == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return defaultValue
	}
	if value > maxValue {
		return maxValue
	}

	return value
}

func decodeJSONBody(w http.ResponseWriter, r *http.Request, target any) bool {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(target); err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "request body must be valid JSON")
		return false
	}

	return true
}

func requireAccount(w http.ResponseWriter, r *http.Request, authService *auth.Service) (auth.Account, bool) {
	headerValue := strings.TrimSpace(r.Header.Get("Authorization"))
	if headerValue == "" {
		writeError(w, r, http.StatusUnauthorized, "AUTH_REQUIRED", "bearer token is required")
		return auth.Account{}, false
	}

	const prefix = "Bearer "
	if !strings.HasPrefix(headerValue, prefix) {
		writeError(w, r, http.StatusUnauthorized, "AUTH_REQUIRED", "authorization header must use bearer token")
		return auth.Account{}, false
	}

	account, err := authService.Authenticate(strings.TrimSpace(strings.TrimPrefix(headerValue, prefix)))
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrAccessTokenExpired):
			writeError(w, r, http.StatusUnauthorized, "AUTH_TOKEN_EXPIRED", "access token has expired")
		default:
			writeError(w, r, http.StatusUnauthorized, "AUTH_REQUIRED", "access token is invalid")
		}

		return auth.Account{}, false
	}

	return account, true
}

func buildCharacterState(account auth.Account, characterService *characters.Service, questService *quests.Service, worldService *world.Service) (characters.StateView, error) {
	state, err := characterService.GetState(account, worldService)
	if err != nil {
		return characters.StateView{}, err
	}

	state.Objectives = questService.ActiveObjectives(state.Character.CharacterID)
	return state, nil
}

func buildPublicWorldState(worldService *world.Service, snapshot characters.RuntimeSnapshot) world.PublicWorldState {
	now := worldService.CurrentTime()
	resetAt := worldService.NextDailyReset(now)
	resetWindowStart := resetAt.Add(-24 * time.Hour)
	arenaStatus := worldService.CurrentArenaStatus(now)
	events := buildPublicEvents(snapshot.Events, 500)
	regions := worldService.ListRegions()
	regionsByID := make(map[string]world.RegionDetail, len(regions))
	populationByRegion := make(map[string]int, len(regions))
	eventCountByRegion := make(map[string]int, len(regions))
	latestEventByRegion := make(map[string]world.WorldEvent, len(regions))
	questsCompletedToday := 0
	dungeonClearsToday := 0
	goldMintedToday := 0
	botsInDungeon := 0
	arenaActors := make(map[string]struct{})

	for _, region := range regions {
		detail, ok := worldService.GetRegion(region.ID)
		if ok {
			regionsByID[region.ID] = detail
		}
	}

	for _, character := range snapshot.Characters {
		if strings.TrimSpace(character.Status) == "" || character.Status == "active" {
			populationByRegion[character.LocationRegionID]++
		}

		if region, ok := regionsByID[character.LocationRegionID]; ok && region.Region.Type == "dungeon" {
			botsInDungeon++
		}
	}

	for _, event := range events {
		if event.RegionID != "" {
			eventCountByRegion[event.RegionID]++
			if _, exists := latestEventByRegion[event.RegionID]; !exists {
				latestEventByRegion[event.RegionID] = event
			}
		}

		occurredAt, err := time.Parse(time.RFC3339, event.OccurredAt)
		if err != nil || occurredAt.Before(resetWindowStart) {
			continue
		}

		switch {
		case event.EventType == "quest.submitted":
			questsCompletedToday++
		case event.EventType == "dungeon.cleared":
			dungeonClearsToday++
		}

		if strings.HasPrefix(event.EventType, "arena.") {
			arenaActors[event.ActorCharacterID] = struct{}{}
		}

		goldMintedToday += intPayload(event.Payload, "reward_gold")
	}

	regionActivities := make([]world.RegionActivity, 0, len(regions))
	for _, region := range regions {
		detail := regionsByID[region.ID]
		highlight := fallbackRegionHighlight(detail.Region.Type, populationByRegion[region.ID])
		if latestEvent, ok := latestEventByRegion[region.ID]; ok {
			highlight = latestEvent.Summary
		}

		regionActivities = append(regionActivities, world.RegionActivity{
			RegionID:         region.ID,
			Name:             region.Name,
			Type:             region.Type,
			MinRank:          region.MinRank,
			TravelCostGold:   region.TravelCostGold,
			Population:       populationByRegion[region.ID],
			RecentEventCount: eventCountByRegion[region.ID],
			Highlight:        highlight,
			BuildingCount:    len(detail.Buildings),
		})
	}

	return world.PublicWorldState{
		ServerTime:           now.Format(time.RFC3339),
		DailyResetAt:         resetAt.Format(time.RFC3339),
		ActiveBotCount:       len(snapshot.Characters),
		BotsInDungeonCount:   botsInDungeon,
		BotsInArenaCount:     len(arenaActors),
		QuestsCompletedToday: questsCompletedToday,
		DungeonClearsToday:   dungeonClearsToday,
		GoldMintedToday:      goldMintedToday,
		Regions:              regionActivities,
		CurrentArenaStatus:   arenaStatus,
	}
}

func buildPublicEvents(events []world.WorldEvent, limit int) []world.WorldEvent {
	if limit <= 0 {
		limit = 20
	}

	publicEvents := make([]world.WorldEvent, 0, len(events))
	seen := make(map[string]struct{}, len(events))

	for _, event := range events {
		if event.Visibility != "public" {
			continue
		}
		if _, exists := seen[event.EventID]; exists {
			continue
		}

		seen[event.EventID] = struct{}{}
		publicEvents = append(publicEvents, event)
	}

	slices.SortFunc(publicEvents, func(left, right world.WorldEvent) int {
		return cmp.Compare(right.OccurredAt, left.OccurredAt)
	})

	if limit > len(publicEvents) {
		limit = len(publicEvents)
	}

	return publicEvents[:limit]
}

func buildPublicLeaderboards(snapshot characters.RuntimeSnapshot) world.Leaderboards {
	reputationEntries := make([]world.LeaderboardEntry, 0, len(snapshot.Characters))
	goldEntries := make([]world.LeaderboardEntry, 0, len(snapshot.Characters))
	dungeonClearsByCharacter := make(map[string]int)
	arenaEntriesByCharacter := make(map[string]int)
	characterByID := make(map[string]characters.Summary, len(snapshot.Characters))

	for _, character := range snapshot.Characters {
		characterByID[character.CharacterID] = character
		reputationEntries = append(reputationEntries, world.LeaderboardEntry{
			CharacterID:   character.CharacterID,
			Name:          character.Name,
			Class:         character.Class,
			WeaponStyle:   character.WeaponStyle,
			RegionID:      character.LocationRegionID,
			Score:         character.Reputation,
			ScoreLabel:    "reputation",
			ActivityLabel: "Highest reputation active bot",
		})
		goldEntries = append(goldEntries, world.LeaderboardEntry{
			CharacterID:   character.CharacterID,
			Name:          character.Name,
			Class:         character.Class,
			WeaponStyle:   character.WeaponStyle,
			RegionID:      character.LocationRegionID,
			Score:         character.Gold,
			ScoreLabel:    "gold",
			ActivityLabel: "Largest gold reserve",
		})
	}

	for _, event := range buildPublicEvents(snapshot.Events, 500) {
		switch {
		case event.EventType == "dungeon.cleared":
			dungeonClearsByCharacter[event.ActorCharacterID]++
		case strings.HasPrefix(event.EventType, "arena."):
			arenaEntriesByCharacter[event.ActorCharacterID]++
		}
	}

	dungeonEntries := make([]world.LeaderboardEntry, 0, len(dungeonClearsByCharacter))
	for characterID, clears := range dungeonClearsByCharacter {
		character, ok := characterByID[characterID]
		if !ok {
			continue
		}

		dungeonEntries = append(dungeonEntries, world.LeaderboardEntry{
			CharacterID:   character.CharacterID,
			Name:          character.Name,
			Class:         character.Class,
			WeaponStyle:   character.WeaponStyle,
			RegionID:      character.LocationRegionID,
			Score:         clears,
			ScoreLabel:    "clears",
			ActivityLabel: "Most dungeon clears",
		})
	}

	arenaEntries := make([]world.LeaderboardEntry, 0, len(arenaEntriesByCharacter))
	for characterID, count := range arenaEntriesByCharacter {
		character, ok := characterByID[characterID]
		if !ok {
			continue
		}

		arenaEntries = append(arenaEntries, world.LeaderboardEntry{
			CharacterID:   character.CharacterID,
			Name:          character.Name,
			Class:         character.Class,
			WeaponStyle:   character.WeaponStyle,
			RegionID:      character.LocationRegionID,
			Score:         count,
			ScoreLabel:    "seed",
			ActivityLabel: "Current arena contender",
		})
	}

	sortLeaderboardEntries(reputationEntries)
	sortLeaderboardEntries(goldEntries)
	sortLeaderboardEntries(dungeonEntries)
	sortLeaderboardEntries(arenaEntries)

	return world.Leaderboards{
		Reputation:    topLeaderboardEntries(reputationEntries, 10),
		Gold:          topLeaderboardEntries(goldEntries, 10),
		WeeklyArena:   topLeaderboardEntries(arenaEntries, 10),
		DungeonClears: topLeaderboardEntries(dungeonEntries, 10),
	}
}

func sortLeaderboardEntries(entries []world.LeaderboardEntry) {
	slices.SortFunc(entries, func(left, right world.LeaderboardEntry) int {
		if left.Score != right.Score {
			return cmp.Compare(right.Score, left.Score)
		}

		return cmp.Compare(left.Name, right.Name)
	})

	for index := range entries {
		entries[index].Rank = index + 1
	}
}

func topLeaderboardEntries(entries []world.LeaderboardEntry, limit int) []world.LeaderboardEntry {
	if limit > len(entries) {
		limit = len(entries)
	}

	return entries[:limit]
}

func intPayload(payload map[string]any, key string) int {
	if payload == nil {
		return 0
	}

	switch value := payload[key].(type) {
	case int:
		return value
	case int32:
		return int(value)
	case int64:
		return int(value)
	case float64:
		return int(value)
	default:
		return 0
	}
}

func fallbackRegionHighlight(regionType string, population int) string {
	if population == 0 {
		return "No public bot activity is visible here yet."
	}

	switch regionType {
	case "safe_hub":
		return "Bots are regrouping and preparing in this safe hub."
	case "dungeon":
		return "Dungeon attempts are being staged in this region."
	default:
		return "Bots are active across this frontier."
	}
}

func currentCharacterWithLimits(account auth.Account, characterService *characters.Service, worldService *world.Service) (characters.Summary, characters.DailyLimits, error) {
	state, err := characterService.GetState(account, worldService)
	if err != nil {
		return characters.Summary{}, characters.DailyLimits{}, err
	}

	return state.Character, state.Limits, nil
}

func activeObjectivesFromBoard(board quests.BoardView) []characters.QuestSummary {
	objectives := make([]characters.QuestSummary, 0, len(board.Quests))
	for _, quest := range board.Quests {
		if quest.Status == "accepted" || quest.Status == "completed" {
			objectives = append(objectives, quest)
		}
	}

	return objectives
}

func executeAction(account auth.Account, actionType string, actionArgs map[string]any, characterService *characters.Service, questService *quests.Service, worldService *world.Service) (map[string]any, error) {
	switch strings.TrimSpace(actionType) {
	case "travel":
		regionID, _ := actionArgs["region_id"].(string)
		result, err := characterService.Travel(account, regionID, worldService)
		if err != nil {
			return nil, err
		}

		board, completed := questService.ProgressTravelQuests(result.State.Character, regionID, result.State.Limits)
		for _, quest := range completed {
			_ = characterService.AppendEvents(result.State.Character.CharacterID, world.WorldEvent{
				EventID:          fmt.Sprintf("evt_action_%s", quest.QuestID),
				EventType:        "quest.completed",
				Visibility:       "public",
				ActorCharacterID: result.State.Character.CharacterID,
				ActorName:        result.State.Character.Name,
				RegionID:         regionID,
				Summary:          fmt.Sprintf("%s completed %s.", result.State.Character.Name, quest.Title),
				Payload: map[string]any{
					"quest_id":    quest.QuestID,
					"quest_title": quest.Title,
				},
				OccurredAt: time.Now().Format(time.RFC3339),
			})
		}

		state, err := buildCharacterState(account, characterService, questService, worldService)
		if err != nil {
			return nil, err
		}
		state.Objectives = activeObjectivesFromBoard(board)

		return map[string]any{
			"action_result": map[string]any{
				"action_type":      "travel",
				"from_region_id":   result.FromRegionID,
				"to_region_id":     result.ToRegionID,
				"travel_cost_gold": result.TravelCostGold,
			},
			"state": state,
		}, nil
	case "accept_quest":
		questID, _ := actionArgs["quest_id"].(string)
		character, limits, err := currentCharacterWithLimits(account, characterService, worldService)
		if err != nil {
			return nil, err
		}

		board, quest, err := questService.AcceptQuest(character, questID, limits)
		if err != nil {
			return nil, err
		}
		_ = characterService.AppendEvents(character.CharacterID, world.WorldEvent{
			EventID:          fmt.Sprintf("evt_accept_%s", quest.QuestID),
			EventType:        "quest.accepted",
			Visibility:       "public",
			ActorCharacterID: character.CharacterID,
			ActorName:        character.Name,
			RegionID:         character.LocationRegionID,
			Summary:          fmt.Sprintf("%s accepted %s.", character.Name, quest.Title),
			Payload: map[string]any{
				"quest_id":    quest.QuestID,
				"quest_title": quest.Title,
			},
			OccurredAt: time.Now().Format(time.RFC3339),
		})
		return map[string]any{
			"action_result": map[string]any{
				"action_type": "accept_quest",
				"quest_id":    quest.QuestID,
				"status":      "accepted",
			},
			"state": map[string]any{
				"quests": board.Quests,
				"limits": board.Limits,
			},
		}, nil
	case "submit_quest":
		questID, _ := actionArgs["quest_id"].(string)
		character, limits, err := currentCharacterWithLimits(account, characterService, worldService)
		if err != nil {
			return nil, err
		}

		quest, err := questService.SubmitQuest(character, questID, limits)
		if err != nil {
			return nil, err
		}
		if _, _, _, _, err := characterService.ApplyQuestSubmission(character.CharacterID, quest); err != nil {
			return nil, err
		}
		state, err := buildCharacterState(account, characterService, questService, worldService)
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"action_result": map[string]any{
				"action_type":       "submit_quest",
				"quest_id":          quest.QuestID,
				"reward_gold":       quest.RewardGold,
				"reward_reputation": quest.RewardReputation,
			},
			"state": state,
		}, nil
	case "reroll_quests":
		confirmCost, _ := actionArgs["confirm_cost"].(bool)
		if !confirmCost {
			return nil, quests.ErrQuestRerollConfirmRequired
		}
		character, limits, err := currentCharacterWithLimits(account, characterService, worldService)
		if err != nil {
			return nil, err
		}
		if _, err := characterService.SpendGold(character.CharacterID, quests.RerollCostGold()); err != nil {
			return nil, err
		}
		character, limits, err = currentCharacterWithLimits(account, characterService, worldService)
		if err != nil {
			return nil, err
		}
		board, err := questService.RerollQuestBoard(character, limits, confirmCost)
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"action_result": map[string]any{
				"action_type":  "reroll_quests",
				"gold_cost":    quests.RerollCostGold(),
				"reroll_count": board.RerollCount,
			},
			"state": board,
		}, nil
	default:
		return characterService.ExecuteAction(account, actionType, actionArgs, worldService)
	}
}
