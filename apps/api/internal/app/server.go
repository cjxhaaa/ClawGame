package app

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"clawgame/apps/api/internal/arena"
	"clawgame/apps/api/internal/auth"
	"clawgame/apps/api/internal/characters"
	"clawgame/apps/api/internal/combat"
	"clawgame/apps/api/internal/dungeons"
	"clawgame/apps/api/internal/inventory"
	"clawgame/apps/api/internal/platform/config"
	"clawgame/apps/api/internal/platform/store"
	"clawgame/apps/api/internal/quests"
	"clawgame/apps/api/internal/world"
	"clawgame/apps/api/internal/worldboss"

	"github.com/go-chi/chi/v5"
)

type Server struct {
	httpServer       *http.Server
	authService      *auth.Service
	arenaService     *arena.Service
	characterService *characters.Service
	dungeonService   *dungeons.Service
	inventoryService *inventory.Service
	questService     *quests.Service
	worldService     *world.Service
	worldBossService *worldboss.Service
}

var requestCounter uint64

func NewServer(cfg config.API) *Server {
	router := chi.NewRouter()
	authService := auth.NewService()
	arenaService := arena.NewService()
	characterService := characters.NewService()
	dungeonService := dungeons.NewService()
	inventoryService := inventory.NewService()
	questService := quests.NewService()
	worldService := world.NewService()
	worldBossService := worldboss.NewService()

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
				Name string `json:"name"`
			}
			if !decodeJSONBody(w, r, &request) {
				return
			}

			state, err := characterService.CreateCharacter(account, request.Name, "", "", worldService)
			if err != nil {
				switch {
				case errors.Is(err, characters.ErrCharacterAlreadyExists):
					writeError(w, r, http.StatusConflict, "CHARACTER_ALREADY_EXISTS", "account already owns a character")
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
			state, err = buildCharacterState(account, characterService, inventoryService, questService, dungeonService, arenaService, worldService)
			if err != nil {
				writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load character state")
				return
			}

			writeEnvelope(w, r, http.StatusOK, state)
		})

		r.Get("/me/skills", func(w http.ResponseWriter, r *http.Request) {
			account, ok := requireAccount(w, r, authService)
			if !ok {
				return
			}

			skillsState, err := characterService.SkillsState(account)
			if err != nil {
				if errors.Is(err, characters.ErrCharacterNotFound) {
					writeError(w, r, http.StatusNotFound, "CHARACTER_NOT_FOUND", "create a character before requesting skills")
					return
				}
				writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load skills")
				return
			}

			writeEnvelope(w, r, http.StatusOK, skillsState)
		})

		r.Post("/me/skills/{skillId}/upgrade", func(w http.ResponseWriter, r *http.Request) {
			account, ok := requireAccount(w, r, authService)
			if !ok {
				return
			}

			skillsState, character, err := characterService.UpgradeSkill(account, chi.URLParam(r, "skillId"))
			if err != nil {
				switch {
				case errors.Is(err, characters.ErrCharacterNotFound):
					writeError(w, r, http.StatusNotFound, "CHARACTER_NOT_FOUND", "create a character before upgrading skills")
				case errors.Is(err, characters.ErrSkillNotFound):
					writeError(w, r, http.StatusNotFound, "SKILL_NOT_FOUND", "skill does not exist")
				case errors.Is(err, characters.ErrSkillLocked):
					writeError(w, r, http.StatusBadRequest, "SKILL_LOCKED", "skill is not available to the current character")
				case errors.Is(err, characters.ErrSkillMaxLevel):
					writeError(w, r, http.StatusBadRequest, "SKILL_MAX_LEVEL", "skill is already at max level")
				case errors.Is(err, characters.ErrGoldInsufficient):
					writeError(w, r, http.StatusBadRequest, "GOLD_INSUFFICIENT", "character does not have enough gold to upgrade this skill")
				default:
					writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to upgrade skill")
				}
				return
			}

			writeEnvelope(w, r, http.StatusOK, map[string]any{
				"character": character,
				"skills":    skillsState,
			})
		})

		r.Post("/me/skills/loadout", func(w http.ResponseWriter, r *http.Request) {
			account, ok := requireAccount(w, r, authService)
			if !ok {
				return
			}

			var request struct {
				SkillIDs []string `json:"skill_ids"`
			}
			if !decodeJSONBody(w, r, &request) {
				return
			}

			skillsState, err := characterService.SetSkillLoadout(account, request.SkillIDs)
			if err != nil {
				switch {
				case errors.Is(err, characters.ErrCharacterNotFound):
					writeError(w, r, http.StatusNotFound, "CHARACTER_NOT_FOUND", "create a character before configuring skills")
				case errors.Is(err, characters.ErrSkillLocked):
					writeError(w, r, http.StatusBadRequest, "SKILL_LOCKED", "loadout contains locked or unavailable skills")
				case errors.Is(err, characters.ErrSkillInvalidLoadout):
					writeError(w, r, http.StatusBadRequest, "SKILL_LOADOUT_INVALID", "loadout must contain up to four unique unlocked skills")
				default:
					writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update skill loadout")
				}
				return
			}

			writeEnvelope(w, r, http.StatusOK, skillsState)
		})

		r.Post("/me/profession-route", func(w http.ResponseWriter, r *http.Request) {
			account, ok := requireAccount(w, r, authService)
			if !ok {
				return
			}

			var request struct {
				RouteID string `json:"route_id"`
			}
			if !decodeJSONBody(w, r, &request) {
				return
			}

			state, err := characterService.ChooseProfessionRoute(account, request.RouteID, worldService)
			if err != nil {
				switch {
				case errors.Is(err, characters.ErrCharacterNotFound):
					writeError(w, r, http.StatusNotFound, "CHARACTER_NOT_FOUND", "create a character first")
				case errors.Is(err, characters.ErrCharacterInvalidRoute):
					writeError(w, r, http.StatusBadRequest, "CHARACTER_INVALID_ROUTE", "profession route is not supported")
				case errors.Is(err, characters.ErrCharacterRouteLocked):
					writeError(w, r, http.StatusBadRequest, "CHARACTER_ROUTE_LOCKED", "profession route is unavailable for the current character state")
				default:
					writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to choose profession route")
				}
				return
			}

			if catalogID := inventory.ProfessionStarterCatalogID(state.Character.ProfessionRoute); catalogID != "" {
				if _, item, grantErr := inventoryService.GrantItemFromCatalog(state.Character, catalogID); grantErr != nil && !errors.Is(grantErr, inventory.ErrCatalogNotFound) {
					writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to grant profession starter weapon")
					return
				} else if grantErr == nil && item.ItemID != "" {
					if _, equipErr := inventoryService.EquipItem(state.Character, item.ItemID); equipErr != nil {
						writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to equip profession starter weapon")
						return
					}
				}
			}

			state, err = buildCharacterState(account, characterService, inventoryService, questService, dungeonService, arenaService, worldService)
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

			state, err = buildCharacterState(account, characterService, inventoryService, questService, dungeonService, arenaService, worldService)
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

		r.Get("/me/planner", func(w http.ResponseWriter, r *http.Request) {
			account, ok := requireAccount(w, r, authService)
			if !ok {
				return
			}

			character, limits, err := currentCharacterWithLimits(account, characterService, worldService)
			if err != nil {
				writeError(w, r, http.StatusNotFound, "CHARACTER_NOT_FOUND", "create a character before requesting planner context")
				return
			}

			queryRegionID := strings.TrimSpace(r.URL.Query().Get("region_id"))
			if queryRegionID == "" {
				queryRegionID = character.LocationRegionID
			}

			if _, exists := worldService.GetRegion(queryRegionID); !exists {
				writeError(w, r, http.StatusNotFound, "REGION_NOT_FOUND", fmt.Sprintf("region %q does not exist", queryRegionID))
				return
			}

			state, err := buildCharacterState(account, characterService, inventoryService, questService, dungeonService, arenaService, worldService)
			if err != nil {
				writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load planner state")
				return
			}

			board := questService.ListQuests(character, limits)
			localQuests := make([]characters.QuestSummary, 0, len(board.Quests))
			suggestedActions := make([]string, 0, 6)
			questRuntimeHints := make([]map[string]any, 0, 4)
			for _, quest := range board.Quests {
				if quest.Status == "accepted" {
					if _, runtime, err := questService.GetQuestRuntime(character.CharacterID, quest.QuestID); err == nil {
						if runtime.SuggestedActionType != "" {
							suggestedActions = appendUniqueString(suggestedActions, runtime.SuggestedActionType)
						}
						if runtime.CurrentStepKey != "" {
							questRuntimeHints = append(questRuntimeHints, map[string]any{
								"quest_id":              quest.QuestID,
								"template_type":         quest.TemplateType,
								"difficulty":            quest.Difficulty,
								"flow_kind":             quest.FlowKind,
								"current_step_key":      runtime.CurrentStepKey,
								"current_step_label":    runtime.CurrentStepLabel,
								"current_step_hint":     runtime.CurrentStepHint,
								"suggested_action_type": runtime.SuggestedActionType,
								"suggested_action_args": runtime.SuggestedActionArgs,
								"target_region_id":      quest.TargetRegionID,
								"selected_choice":       runtime.State["selected_choice_key"],
								"selected_choice_label": runtime.State["selected_choice_label"],
								"available_choices":     runtime.AvailableChoices,
								"completed_step_keys":   runtime.CompletedStepKeys,
							})
						}
					}
				}
				if quest.TargetRegionID != queryRegionID {
					continue
				}
				if quest.Status == "submitted" || quest.Status == "expired" {
					continue
				}
				localQuests = append(localQuests, quest)
			}

			localDungeons := make([]map[string]any, 0, 2)
			inventoryView := inventoryService.GetInventory(character)
			prepItems := make([]map[string]any, 0, 2)
			for _, definition := range dungeonService.ListDungeonDefinitions() {
				if definition.RegionID != queryRegionID {
					continue
				}

				hasRemainingQuota := limits.DungeonEntryUsed < limits.DungeonEntryCap
				canEnter := hasRemainingQuota
				prep := buildDungeonPreparationEntry(character, definition, inventoryView, state.CombatPower)

				localDungeons = append(localDungeons, map[string]any{
					"dungeon_id":                  definition.DungeonID,
					"name":                        definition.Name,
					"region_id":                   definition.RegionID,
					"has_remaining_quota":         hasRemainingQuota,
					"can_enter":                   canEnter,
					"requires_potion_loadout":     false,
					"potion_slot_count":           2,
					"recommended_level_min":       definition.RecommendedLevelMin,
					"recommended_level_max":       definition.RecommendedLevelMax,
					"current_power":               prep["current_power"],
					"recommended_power":           prep["recommended_power"],
					"power_gap":                   prep["power_gap"],
					"current_equipment_score":     prep["current_equipment_score"],
					"recommended_equipment_score": prep["recommended_equipment_score"],
					"score_gap":                   prep["score_gap"],
					"readiness":                   prep["readiness"],
				})
				prepItems = append(prepItems, prep)
			}

			if state.DungeonDaily.HasClaimableRun {
				suggestedActions = append(suggestedActions, "claim_dungeon_rewards")
				if !state.DungeonDaily.HasRemainingQuota && state.Character.Reputation >= state.Limits.ReputationPerBonusClaim {
					suggestedActions = appendUniqueString(suggestedActions, "exchange_dungeon_reward_claims")
				}
			}
			for _, quest := range localQuests {
				if quest.Status == "completed" {
					suggestedActions = appendUniqueString(suggestedActions, "submit_quest")
				}
			}
			for _, dungeon := range localDungeons {
				if canEnter, _ := dungeon["can_enter"].(bool); canEnter {
					suggestedActions = appendUniqueString(suggestedActions, "enter_dungeon")
					break
				}
			}
			if detail, ok := worldService.GetRegion(queryRegionID); ok {
				for _, action := range detail.AvailableRegionActions {
					switch action {
					case "resolve_field_encounter:hunt", "resolve_field_encounter:gather", "resolve_field_encounter:curio", "enter_dungeon":
						suggestedActions = appendUniqueString(suggestedActions, action)
					}
				}
			}

			writeEnvelope(w, r, http.StatusOK, map[string]any{
				"today": map[string]any{
					"daily_reset_at": limits.DailyResetAt,
					"quest_completion": map[string]any{
						"used":      limits.QuestCompletionUsed,
						"cap":       limits.QuestCompletionCap,
						"remaining": max(0, limits.QuestCompletionCap-limits.QuestCompletionUsed),
					},
					"dungeon_claim": map[string]any{
						"used":      limits.DungeonEntryUsed,
						"cap":       limits.DungeonEntryCap,
						"remaining": max(0, limits.DungeonEntryCap-limits.DungeonEntryUsed),
					},
				},
				"character_region_id": character.LocationRegionID,
				"query_region_id":     queryRegionID,
				"local_quests":        localQuests,
				"quest_runtime_hints": questRuntimeHints,
				"local_dungeons":      localDungeons,
				"dungeon_preparation": map[string]any{
					"current_power":           state.CombatPower.PanelPowerScore,
					"current_equipment_score": inventoryView.EquipmentScore,
					"upgrade_hint_count":      len(inventoryView.UpgradeHints),
					"potion_option_count":     len(inventoryView.PotionLoadoutOptions),
					"items":                   prepItems,
				},
				"dungeon_daily":     state.DungeonDaily,
				"suggested_actions": suggestedActions,
			})
		})

		r.Get("/me/actions", func(w http.ResponseWriter, r *http.Request) {
			account, ok := requireAccount(w, r, authService)
			if !ok {
				return
			}

			character, limits, err := currentCharacterWithLimits(account, characterService, worldService)
			if err != nil {
				if errors.Is(err, characters.ErrCharacterNotFound) {
					writeError(w, r, http.StatusNotFound, "CHARACTER_NOT_FOUND", "create a character before requesting actions")
					return
				}
				writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to initialize quest board for actions")
				return
			}
			_ = questService.ListQuests(character, limits)

			actions, err := characterService.ListValidActions(account, worldService)
			if err != nil {
				if errors.Is(err, characters.ErrCharacterNotFound) {
					writeError(w, r, http.StatusNotFound, "CHARACTER_NOT_FOUND", "create a character before requesting actions")
					return
				}

				writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load valid actions")
				return
			}
			actions = appendQuestRuntimeValidActions(actions, account, characterService, questService)

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

			result, err := executeAction(account, request.ActionType, request.ActionArgs, characterService, questService, dungeonService, inventoryService, arenaService, worldService)
			if err != nil {
				switch {
				case errors.Is(err, characters.ErrCharacterNotFound):
					writeError(w, r, http.StatusNotFound, "CHARACTER_NOT_FOUND", "create a character before executing actions")
				case errors.Is(err, quests.ErrQuestNotFound):
					writeError(w, r, http.StatusNotFound, "QUEST_NOT_FOUND", "quest does not exist on the current board")
				case errors.Is(err, quests.ErrQuestInvalidState):
					writeError(w, r, http.StatusBadRequest, "QUEST_INVALID_STATE", "quest is not in a valid state for this action")
				case errors.Is(err, quests.ErrQuestChoiceNotAvailable):
					writeError(w, r, http.StatusBadRequest, "QUEST_CHOICE_NOT_AVAILABLE", "quest choice is not available")
				case errors.Is(err, quests.ErrQuestInteractionInvalid):
					writeError(w, r, http.StatusBadRequest, "QUEST_INTERACTION_INVALID", "quest interaction is invalid")
				case errors.Is(err, characters.ErrQuestCompletionCap):
					writeError(w, r, http.StatusBadRequest, "QUEST_COMPLETION_LIMIT_REACHED", "daily quest completion cap has been reached")
				case errors.Is(err, characters.ErrReputationInsufficient):
					writeError(w, r, http.StatusBadRequest, "REPUTATION_INSUFFICIENT", "character does not have enough reputation for this exchange")
				case errors.Is(err, characters.ErrGoldInsufficient):
					writeError(w, r, http.StatusBadRequest, "GOLD_INSUFFICIENT", "character does not have enough gold for this action")
				case errors.Is(err, characters.ErrTravelRegionNotFound):
					writeError(w, r, http.StatusNotFound, "TRAVEL_REGION_NOT_FOUND", "target region does not exist")
				case errors.Is(err, characters.ErrTravelInsufficientGold):
					writeError(w, r, http.StatusBadRequest, "TRAVEL_INSUFFICIENT_GOLD", "character does not have enough gold to travel")
				case errors.Is(err, world.ErrFieldEncounterUnavailable):
					writeError(w, r, http.StatusBadRequest, "FIELD_ENCOUNTER_UNAVAILABLE", "field encounters are not available in the current region")
				case errors.Is(err, world.ErrFieldEncounterInvalidMode):
					writeError(w, r, http.StatusBadRequest, "FIELD_ENCOUNTER_INVALID_MODE", "field encounter approach is not supported")
				case errors.Is(err, characters.ErrActionNotSupported):
					writeError(w, r, http.StatusBadRequest, "ACTION_NOT_SUPPORTED", "action type is not currently supported")
				case errors.Is(err, dungeons.ErrDungeonNotFound):
					writeError(w, r, http.StatusNotFound, "DUNGEON_NOT_FOUND", "dungeon definition does not exist")
				case errors.Is(err, dungeons.ErrDungeonRunAlreadyActive):
					writeError(w, r, http.StatusConflict, "DUNGEON_RUN_ALREADY_ACTIVE", "character already has an active dungeon run")
				case errors.Is(err, dungeons.ErrDungeonRunNotFound):
					writeError(w, r, http.StatusNotFound, "DUNGEON_RUN_NOT_FOUND", "dungeon run does not exist")
				case errors.Is(err, dungeons.ErrDungeonRunForbidden):
					writeError(w, r, http.StatusForbidden, "DUNGEON_RUN_FORBIDDEN", "dungeon run does not belong to caller")
				case errors.Is(err, dungeons.ErrDungeonRewardClaimNotAllowed):
					writeError(w, r, http.StatusBadRequest, "DUNGEON_REWARD_NOT_CLAIMABLE", "run rewards are not claimable")
				case errors.Is(err, dungeons.ErrDungeonRewardClaimCapReached), errors.Is(err, characters.ErrDungeonRewardClaimCap):
					writeError(w, r, http.StatusBadRequest, "DUNGEON_REWARD_CLAIM_LIMIT_REACHED", "daily dungeon reward claim cap has been reached")
				case errors.Is(err, dungeons.ErrDungeonPotionLoadoutInvalid):
					writeError(w, r, http.StatusBadRequest, "DUNGEON_POTION_LOADOUT_INVALID", "select up to two owned potion types before entering a dungeon")
				case errors.Is(err, inventory.ErrItemNotOwned):
					writeError(w, r, http.StatusNotFound, "ITEM_NOT_OWNED", "item does not belong to character")
				case errors.Is(err, inventory.ErrItemNotEquippable):
					writeError(w, r, http.StatusBadRequest, "ITEM_NOT_EQUIPPABLE", "item cannot be equipped by current character")
				case errors.Is(err, inventory.ErrSlotNotOccupied):
					writeError(w, r, http.StatusBadRequest, "ITEM_SLOT_EMPTY", "slot is not currently occupied")
				case errors.Is(err, arena.ErrSignupClosed):
					writeError(w, r, http.StatusBadRequest, "ARENA_SIGNUP_CLOSED", "arena signup window is closed")
				case errors.Is(err, arena.ErrAlreadySignedUp):
					writeError(w, r, http.StatusConflict, "ARENA_ALREADY_SIGNED_UP", "character already signed up for current arena")
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

			state, err := buildCharacterState(account, characterService, inventoryService, questService, dungeonService, arenaService, worldService)
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

		r.Get("/me/quests/{questId}", func(w http.ResponseWriter, r *http.Request) {
			account, ok := requireAccount(w, r, authService)
			if !ok {
				return
			}

			character, limits, err := currentCharacterWithLimits(account, characterService, worldService)
			if err != nil {
				writeError(w, r, http.StatusNotFound, "CHARACTER_NOT_FOUND", "create a character before requesting quest detail")
				return
			}
			_ = questService.ListQuests(character, limits)

			quest, runtime, err := questService.GetQuestRuntime(character.CharacterID, chi.URLParam(r, "questId"))
			if err != nil {
				if errors.Is(err, quests.ErrQuestNotFound) {
					writeError(w, r, http.StatusNotFound, "QUEST_NOT_FOUND", "quest does not exist on the current board")
					return
				}
				writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load quest runtime")
				return
			}

			writeEnvelope(w, r, http.StatusOK, map[string]any{
				"quest":    quest,
				"runtime":  runtime,
				"can_view": true,
			})
		})

		r.Post("/me/quests/{questId}/choice", func(w http.ResponseWriter, r *http.Request) {
			account, ok := requireAccount(w, r, authService)
			if !ok {
				return
			}

			var request struct {
				ChoiceKey string `json:"choice_key"`
			}
			if !decodeJSONBody(w, r, &request) {
				return
			}

			character, limits, err := currentCharacterWithLimits(account, characterService, worldService)
			if err != nil {
				writeError(w, r, http.StatusNotFound, "CHARACTER_NOT_FOUND", "create a character before choosing quest branches")
				return
			}
			_ = questService.ListQuests(character, limits)

			quest, runtime, err := questService.ApplyQuestChoice(character, chi.URLParam(r, "questId"), request.ChoiceKey)
			if err != nil {
				switch {
				case errors.Is(err, quests.ErrQuestNotFound):
					writeError(w, r, http.StatusNotFound, "QUEST_NOT_FOUND", "quest does not exist on the current board")
				case errors.Is(err, quests.ErrQuestChoiceNotAvailable):
					writeError(w, r, http.StatusBadRequest, "QUEST_CHOICE_NOT_AVAILABLE", "choice is not available for this quest")
				default:
					writeError(w, r, http.StatusBadRequest, "QUEST_INVALID_STATE", "quest choice could not be applied")
				}
				return
			}

			writeEnvelope(w, r, http.StatusOK, map[string]any{
				"quest":   quest,
				"runtime": runtime,
			})
		})

		r.Post("/me/quests/{questId}/interact", func(w http.ResponseWriter, r *http.Request) {
			account, ok := requireAccount(w, r, authService)
			if !ok {
				return
			}

			var request struct {
				Interaction string `json:"interaction"`
			}
			if !decodeJSONBody(w, r, &request) {
				return
			}

			character, limits, err := currentCharacterWithLimits(account, characterService, worldService)
			if err != nil {
				writeError(w, r, http.StatusNotFound, "CHARACTER_NOT_FOUND", "create a character before advancing quest interactions")
				return
			}
			_ = questService.ListQuests(character, limits)

			quest, runtime, err := questService.AdvanceQuestInteraction(character, chi.URLParam(r, "questId"), request.Interaction)
			if err != nil {
				switch {
				case errors.Is(err, quests.ErrQuestNotFound):
					writeError(w, r, http.StatusNotFound, "QUEST_NOT_FOUND", "quest does not exist on the current board")
				case errors.Is(err, quests.ErrQuestInteractionInvalid):
					writeError(w, r, http.StatusBadRequest, "QUEST_INTERACTION_INVALID", "interaction is not valid for this quest")
				default:
					writeError(w, r, http.StatusBadRequest, "QUEST_INVALID_STATE", "quest interaction could not be applied")
				}
				return
			}

			writeEnvelope(w, r, http.StatusOK, map[string]any{
				"quest":   quest,
				"runtime": runtime,
			})
		})

		r.Post("/me/field-encounter", func(w http.ResponseWriter, r *http.Request) {
			account, ok := requireAccount(w, r, authService)
			if !ok {
				return
			}

			var request struct {
				Approach string `json:"approach"`
			}
			if !decodeJSONBody(w, r, &request) {
				return
			}

			result, err := resolveFieldEncounter(account, request.Approach, characterService, inventoryService, questService, dungeonService, arenaService, worldService)
			if err != nil {
				switch {
				case errors.Is(err, characters.ErrCharacterNotFound):
					writeError(w, r, http.StatusNotFound, "CHARACTER_NOT_FOUND", "create a character before entering field encounters")
				case errors.Is(err, world.ErrFieldEncounterUnavailable):
					writeError(w, r, http.StatusBadRequest, "FIELD_ENCOUNTER_UNAVAILABLE", "field encounters are not available in the current region")
				case errors.Is(err, world.ErrFieldEncounterInvalidMode):
					writeError(w, r, http.StatusBadRequest, "FIELD_ENCOUNTER_INVALID_MODE", "field encounter approach is not supported")
				default:
					writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to resolve field encounter")
				}
				return
			}

			writeEnvelope(w, r, http.StatusOK, result)
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
			if limits.QuestCompletionUsed >= limits.QuestCompletionCap {
				writeError(w, r, http.StatusBadRequest, "QUEST_COMPLETION_LIMIT_REACHED", "daily quest completion cap has been reached")
				return
			}

			quest, err := questService.PrepareQuestSubmission(character, chi.URLParam(r, "questId"), limits)
			if err != nil {
				switch {
				case errors.Is(err, quests.ErrQuestNotFound):
					writeError(w, r, http.StatusNotFound, "QUEST_NOT_FOUND", "quest does not exist on the current board")
				default:
					writeError(w, r, http.StatusBadRequest, "QUEST_INVALID_STATE", "quest is not ready for submission")
				}
				return
			}

			_, _, _, _, err = characterService.ApplyQuestSubmission(character.CharacterID, quest)
			if err != nil {
				if errors.Is(err, characters.ErrQuestCompletionCap) {
					writeError(w, r, http.StatusBadRequest, "QUEST_COMPLETION_LIMIT_REACHED", "daily quest completion cap has been reached")
					return
				}
				writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to apply quest rewards")
				return
			}
			quest, err = questService.FinalizeQuestSubmission(character.CharacterID, quest.QuestID)
			if err != nil {
				writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to finalize quest submission")
				return
			}

			state, err := buildCharacterState(account, characterService, inventoryService, questService, dungeonService, arenaService, worldService)
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

		r.Post("/me/dungeons/reward-claims/exchange", func(w http.ResponseWriter, r *http.Request) {
			account, ok := requireAccount(w, r, authService)
			if !ok {
				return
			}

			var request struct {
				Quantity int `json:"quantity"`
			}
			if !decodeJSONBody(w, r, &request) {
				return
			}

			character, exists := characterService.GetCharacterByAccount(account)
			if !exists {
				writeError(w, r, http.StatusNotFound, "CHARACTER_NOT_FOUND", "create a character before exchanging dungeon reward claims")
				return
			}

			summary, limits, err := characterService.PurchaseDungeonRewardClaims(character.CharacterID, request.Quantity)
			if err != nil {
				if errors.Is(err, characters.ErrReputationInsufficient) {
					writeError(w, r, http.StatusBadRequest, "REPUTATION_INSUFFICIENT", "character does not have enough reputation for this exchange")
					return
				}
				writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to exchange dungeon reward claims")
				return
			}

			writeEnvelope(w, r, http.StatusOK, map[string]any{
				"character": summary,
				"limits":    limits,
			})
		})

		r.Get("/me/inventory", func(w http.ResponseWriter, r *http.Request) {
			account, ok := requireAccount(w, r, authService)
			if !ok {
				return
			}

			character, exists := characterService.GetCharacterByAccount(account)
			if !exists {
				writeError(w, r, http.StatusNotFound, "CHARACTER_NOT_FOUND", "create a character before requesting inventory")
				return
			}

			writeEnvelope(w, r, http.StatusOK, inventoryService.GetInventory(character))
		})

		r.Get("/world-boss/current", func(w http.ResponseWriter, r *http.Request) {
			if _, ok := requireAccount(w, r, authService); !ok {
				return
			}
			writeEnvelope(w, r, http.StatusOK, worldBossService.CurrentBoss())
		})

		r.Get("/world-boss/queue-status", func(w http.ResponseWriter, r *http.Request) {
			account, ok := requireAccount(w, r, authService)
			if !ok {
				return
			}
			character, exists := characterService.GetCharacterByAccount(account)
			if !exists {
				writeError(w, r, http.StatusNotFound, "CHARACTER_NOT_FOUND", "create a character before requesting world boss queue status")
				return
			}
			writeEnvelope(w, r, http.StatusOK, worldBossService.QueueStatus(character.CharacterID))
		})

		r.Post("/world-boss/queue", func(w http.ResponseWriter, r *http.Request) {
			account, ok := requireAccount(w, r, authService)
			if !ok {
				return
			}
			character, exists := characterService.GetCharacterByAccount(account)
			if !exists {
				writeError(w, r, http.StatusNotFound, "CHARACTER_NOT_FOUND", "create a character before joining world boss queue")
				return
			}

			join := worldBossService.JoinQueue(character.CharacterID, character.Name)
			if len(join.MatchedCharacterIDs) > 0 {
				participants := make([]worldboss.ParticipantSnapshot, 0, len(join.MatchedCharacterIDs))
				for _, characterID := range join.MatchedCharacterIDs {
					summary, baseStats, skills, found := characterService.GetCharacterByID(characterID)
					if !found {
						continue
					}
					derivedStats := inventoryService.DeriveStats(summary, baseStats)
					player := dungeons.BuildPlayerCombatant(summary, derivedStats, skills)
					inv := inventoryService.GetInventory(summary)
					power, _ := buildCombatPower(summary, derivedStats, inv, dungeonService)
					participants = append(participants, worldboss.ParticipantSnapshot{
						Character: summary,
						Power:     power.PanelPowerScore,
						Player:    player,
					})
				}
				if len(participants) == worldBossService.CurrentBoss().RequiredPartySize {
					raid := worldBossService.ResolveMatchedRaid(participants)
					join.ResolvedRaid = &raid
					for _, member := range raid.Members {
						if _, err := characterService.GrantGold(member.CharacterID, raid.RewardPackage.RewardGold); err == nil {
							_, _ = characterService.GrantMaterials(member.CharacterID, raid.RewardPackage.MaterialDrops)
						}
						_ = characterService.AppendEvents(member.CharacterID, world.WorldEvent{
							EventID:          requestID(r),
							EventType:        "world_boss.reward_granted",
							Visibility:       "public",
							ActorCharacterID: member.CharacterID,
							ActorName:        member.Name,
							RegionID:         "main_city",
							Summary:          fmt.Sprintf("%s earned %s-tier world boss rewards.", member.Name, raid.RewardTier),
							Payload: map[string]any{
								"raid_id":      raid.RaidID,
								"reward_tier":  raid.RewardTier,
								"damage_dealt": member.DamageDealt,
							},
							OccurredAt: time.Now().Format(time.RFC3339),
						})
					}
					join.Status = worldBossService.QueueStatus(character.CharacterID)
				}
			}

			writeEnvelope(w, r, http.StatusOK, join)
		})

		r.Get("/world-boss/raids/{raidId}", func(w http.ResponseWriter, r *http.Request) {
			account, ok := requireAccount(w, r, authService)
			if !ok {
				return
			}
			character, exists := characterService.GetCharacterByAccount(account)
			if !exists {
				writeError(w, r, http.StatusNotFound, "CHARACTER_NOT_FOUND", "create a character before requesting world boss raid detail")
				return
			}
			raid, found := worldBossService.GetRaid(character.CharacterID, chi.URLParam(r, "raidId"))
			if !found {
				writeError(w, r, http.StatusNotFound, "WORLD_BOSS_RAID_NOT_FOUND", "world boss raid does not exist")
				return
			}
			writeEnvelope(w, r, http.StatusOK, raid)
		})

		r.Post("/items/{itemId}/reforge", func(w http.ResponseWriter, r *http.Request) {
			account, ok := requireAccount(w, r, authService)
			if !ok {
				return
			}
			character, exists := characterService.GetCharacterByAccount(account)
			if !exists {
				writeError(w, r, http.StatusNotFound, "CHARACTER_NOT_FOUND", "create a character before reforging items")
				return
			}
			itemBefore, reforgeStoneCost, err := inventoryService.GetReforgeCost(character, chi.URLParam(r, "itemId"))
			if err != nil {
				switch {
				case errors.Is(err, inventory.ErrItemNotOwned):
					writeError(w, r, http.StatusNotFound, "ITEM_NOT_OWNED", "item does not belong to character")
				case errors.Is(err, inventory.ErrItemNotReforgeable):
					writeError(w, r, http.StatusBadRequest, "ITEM_NOT_REFORGEABLE", "item does not support extra-affix reforge")
				default:
					writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to validate reforge target")
				}
				return
			}
			if _, err := characterService.SpendMaterials(character.CharacterID, []map[string]any{{"material_key": inventory.ReforgeMaterialKey, "quantity": reforgeStoneCost}}); err != nil {
				if errors.Is(err, characters.ErrMaterialsInsufficient) {
					writeError(w, r, http.StatusBadRequest, "MATERIALS_INSUFFICIENT", "character does not have enough reforge material")
					return
				}
				writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to spend reforge material")
				return
			}
			view, item, err := inventoryService.ReforgeItem(character, chi.URLParam(r, "itemId"))
			if err != nil {
				switch {
				case errors.Is(err, inventory.ErrItemNotOwned):
					writeError(w, r, http.StatusNotFound, "ITEM_NOT_OWNED", "item does not belong to character")
				case errors.Is(err, inventory.ErrItemNotReforgeable):
					writeError(w, r, http.StatusBadRequest, "ITEM_NOT_REFORGEABLE", "item does not support extra-affix reforge")
				default:
					writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to roll reforge result")
				}
				return
			}
			writeEnvelope(w, r, http.StatusOK, map[string]any{
				"item_before": itemBefore,
				"item":        item,
				"inventory":   view,
				"reforge_cost": map[string]any{
					"material_key": inventory.ReforgeMaterialKey,
					"quantity":     reforgeStoneCost,
				},
			})
		})

		r.Post("/items/{itemId}/reforge/save", func(w http.ResponseWriter, r *http.Request) {
			account, ok := requireAccount(w, r, authService)
			if !ok {
				return
			}
			character, exists := characterService.GetCharacterByAccount(account)
			if !exists {
				writeError(w, r, http.StatusNotFound, "CHARACTER_NOT_FOUND", "create a character before saving reforge results")
				return
			}
			view, item, err := inventoryService.SaveReforge(character, chi.URLParam(r, "itemId"))
			if err != nil {
				switch {
				case errors.Is(err, inventory.ErrItemNotOwned):
					writeError(w, r, http.StatusNotFound, "ITEM_NOT_OWNED", "item does not belong to character")
				case errors.Is(err, inventory.ErrReforgeNotPending):
					writeError(w, r, http.StatusBadRequest, "REFORGE_NOT_PENDING", "item has no pending reforge result")
				default:
					writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to save reforge result")
				}
				return
			}
			writeEnvelope(w, r, http.StatusOK, map[string]any{
				"item":      item,
				"inventory": view,
			})
		})

		r.Post("/items/{itemId}/reforge/discard", func(w http.ResponseWriter, r *http.Request) {
			account, ok := requireAccount(w, r, authService)
			if !ok {
				return
			}
			character, exists := characterService.GetCharacterByAccount(account)
			if !exists {
				writeError(w, r, http.StatusNotFound, "CHARACTER_NOT_FOUND", "create a character before discarding reforge results")
				return
			}
			view, item, err := inventoryService.DiscardReforge(character, chi.URLParam(r, "itemId"))
			if err != nil {
				switch {
				case errors.Is(err, inventory.ErrItemNotOwned):
					writeError(w, r, http.StatusNotFound, "ITEM_NOT_OWNED", "item does not belong to character")
				case errors.Is(err, inventory.ErrReforgeNotPending):
					writeError(w, r, http.StatusBadRequest, "REFORGE_NOT_PENDING", "item has no pending reforge result")
				default:
					writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to discard reforge result")
				}
				return
			}
			writeEnvelope(w, r, http.StatusOK, map[string]any{
				"item":      item,
				"inventory": view,
			})
		})

		r.Post("/me/equipment/equip", func(w http.ResponseWriter, r *http.Request) {
			account, ok := requireAccount(w, r, authService)
			if !ok {
				return
			}

			character, exists := characterService.GetCharacterByAccount(account)
			if !exists {
				writeError(w, r, http.StatusNotFound, "CHARACTER_NOT_FOUND", "create a character before equipping items")
				return
			}

			var request struct {
				ItemID string `json:"item_id"`
			}
			if !decodeJSONBody(w, r, &request) {
				return
			}

			view, err := inventoryService.EquipItem(character, request.ItemID)
			if err != nil {
				switch {
				case errors.Is(err, inventory.ErrItemNotOwned):
					writeError(w, r, http.StatusNotFound, "ITEM_NOT_OWNED", "item does not belong to character")
				case errors.Is(err, inventory.ErrItemNotEquippable):
					writeError(w, r, http.StatusBadRequest, "ITEM_NOT_EQUIPPABLE", "item cannot be equipped by current character")
				default:
					writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to equip item")
				}
				return
			}

			_ = characterService.AppendEvents(character.CharacterID, world.WorldEvent{
				EventID:          requestID(r),
				EventType:        "inventory.item_equipped",
				Visibility:       "public",
				ActorCharacterID: character.CharacterID,
				ActorName:        character.Name,
				RegionID:         character.LocationRegionID,
				Summary:          fmt.Sprintf("%s equipped an item.", character.Name),
				Payload:          map[string]any{"item_id": request.ItemID},
				OccurredAt:       time.Now().Format(time.RFC3339),
			})

			writeEnvelope(w, r, http.StatusOK, view)
		})

		r.Post("/me/equipment/unequip", func(w http.ResponseWriter, r *http.Request) {
			account, ok := requireAccount(w, r, authService)
			if !ok {
				return
			}

			character, exists := characterService.GetCharacterByAccount(account)
			if !exists {
				writeError(w, r, http.StatusNotFound, "CHARACTER_NOT_FOUND", "create a character before unequipping items")
				return
			}

			var request struct {
				Slot string `json:"slot"`
			}
			if !decodeJSONBody(w, r, &request) {
				return
			}

			view, err := inventoryService.UnequipItem(character, request.Slot)
			if err != nil {
				if errors.Is(err, inventory.ErrSlotNotOccupied) {
					writeError(w, r, http.StatusBadRequest, "ITEM_SLOT_EMPTY", "slot is not currently occupied")
					return
				}

				writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to unequip item")
				return
			}

			_ = characterService.AppendEvents(character.CharacterID, world.WorldEvent{
				EventID:          requestID(r),
				EventType:        "inventory.item_unequipped",
				Visibility:       "public",
				ActorCharacterID: character.CharacterID,
				ActorName:        character.Name,
				RegionID:         character.LocationRegionID,
				Summary:          fmt.Sprintf("%s unequipped an item from %s.", character.Name, request.Slot),
				Payload:          map[string]any{"slot": request.Slot},
				OccurredAt:       time.Now().Format(time.RFC3339),
			})

			writeEnvelope(w, r, http.StatusOK, view)
		})

		r.Get("/buildings/{buildingId}", func(w http.ResponseWriter, r *http.Request) {
			buildingID := chi.URLParam(r, "buildingId")
			building, region, found := findBuilding(worldService, buildingID)
			if !found {
				writeError(w, r, http.StatusNotFound, "BUILDING_NOT_FOUND", "building does not exist")
				return
			}

			writeEnvelope(w, r, http.StatusOK, map[string]any{
				"building":          building,
				"region":            region.Region,
				"supported_actions": building.Actions,
			})
		})

		r.Get("/buildings/{buildingId}/skills", func(w http.ResponseWriter, r *http.Request) {
			account, ok := requireAccount(w, r, authService)
			if !ok {
				return
			}

			building, _, found := findBuilding(worldService, chi.URLParam(r, "buildingId"))
			if !found {
				writeError(w, r, http.StatusNotFound, "BUILDING_NOT_FOUND", "building does not exist")
				return
			}
			if building.Type != "guild" {
				writeError(w, r, http.StatusBadRequest, "INVALID_ACTION_STATE", "skills are only available at the Adventurers Guild")
				return
			}

			skillsState, err := characterService.SkillsState(account)
			if err != nil {
				if errors.Is(err, characters.ErrCharacterNotFound) {
					writeError(w, r, http.StatusNotFound, "CHARACTER_NOT_FOUND", "create a character before requesting skills")
					return
				}
				writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load guild skills")
				return
			}

			writeEnvelope(w, r, http.StatusOK, map[string]any{
				"building_id": building.ID,
				"skills":      skillsState,
			})
		})

		r.Get("/buildings/{buildingId}/shop-inventory", func(w http.ResponseWriter, r *http.Request) {
			account, ok := requireAccount(w, r, authService)
			if !ok {
				return
			}

			character, exists := characterService.GetCharacterByAccount(account)
			if !exists {
				writeError(w, r, http.StatusNotFound, "CHARACTER_NOT_FOUND", "create a character before requesting shop inventory")
				return
			}

			buildingID := chi.URLParam(r, "buildingId")
			building, _, found := findBuilding(worldService, buildingID)
			if !found {
				writeError(w, r, http.StatusNotFound, "BUILDING_NOT_FOUND", "building does not exist")
				return
			}

			writeEnvelope(w, r, http.StatusOK, map[string]any{
				"building_id": building.ID,
				"items":       inventoryService.ListShopInventory(building.Type, character),
			})
		})

		r.Post("/buildings/{buildingId}/heal", func(w http.ResponseWriter, r *http.Request) {
			handleBuildingAction(w, r, authService, worldService, characterService, inventoryService, questService, dungeonService, arenaService, "heal")
		})
		r.Post("/buildings/{buildingId}/cleanse", func(w http.ResponseWriter, r *http.Request) {
			handleBuildingAction(w, r, authService, worldService, characterService, inventoryService, questService, dungeonService, arenaService, "cleanse")
		})
		r.Post("/buildings/{buildingId}/enhance", func(w http.ResponseWriter, r *http.Request) {
			handleBuildingAction(w, r, authService, worldService, characterService, inventoryService, questService, dungeonService, arenaService, "enhance")
		})
		r.Post("/buildings/{buildingId}/salvage", func(w http.ResponseWriter, r *http.Request) {
			handleBuildingAction(w, r, authService, worldService, characterService, inventoryService, questService, dungeonService, arenaService, "salvage")
		})
		r.Post("/buildings/{buildingId}/repair", func(w http.ResponseWriter, r *http.Request) {
			handleBuildingAction(w, r, authService, worldService, characterService, inventoryService, questService, dungeonService, arenaService, "repair")
		})
		r.Post("/buildings/{buildingId}/purchase", func(w http.ResponseWriter, r *http.Request) {
			handleBuildingAction(w, r, authService, worldService, characterService, inventoryService, questService, dungeonService, arenaService, "purchase")
		})
		r.Post("/buildings/{buildingId}/sell", func(w http.ResponseWriter, r *http.Request) {
			handleBuildingAction(w, r, authService, worldService, characterService, inventoryService, questService, dungeonService, arenaService, "sell")
		})
		r.Post("/buildings/{buildingId}/skills/{skillId}/upgrade", func(w http.ResponseWriter, r *http.Request) {
			account, ok := requireAccount(w, r, authService)
			if !ok {
				return
			}

			building, _, found := findBuilding(worldService, chi.URLParam(r, "buildingId"))
			if !found {
				writeError(w, r, http.StatusNotFound, "BUILDING_NOT_FOUND", "building does not exist")
				return
			}
			if building.Type != "guild" {
				writeError(w, r, http.StatusBadRequest, "INVALID_ACTION_STATE", "skill upgrades are only available at the Adventurers Guild")
				return
			}

			skillsState, character, err := characterService.UpgradeSkill(account, chi.URLParam(r, "skillId"))
			if err != nil {
				switch {
				case errors.Is(err, characters.ErrCharacterNotFound):
					writeError(w, r, http.StatusNotFound, "CHARACTER_NOT_FOUND", "create a character before upgrading skills")
				case errors.Is(err, characters.ErrSkillNotFound):
					writeError(w, r, http.StatusNotFound, "SKILL_NOT_FOUND", "skill does not exist")
				case errors.Is(err, characters.ErrSkillLocked):
					writeError(w, r, http.StatusBadRequest, "SKILL_LOCKED", "skill is not available to the current character")
				case errors.Is(err, characters.ErrSkillMaxLevel):
					writeError(w, r, http.StatusBadRequest, "SKILL_MAX_LEVEL", "skill is already at max level")
				case errors.Is(err, characters.ErrGoldInsufficient):
					writeError(w, r, http.StatusBadRequest, "GOLD_INSUFFICIENT", "character does not have enough gold to upgrade this skill")
				default:
					writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to upgrade guild skill")
				}
				return
			}

			writeEnvelope(w, r, http.StatusOK, map[string]any{
				"building_id": building.ID,
				"character":   character,
				"skills":      skillsState,
			})
		})
		r.Post("/buildings/{buildingId}/skill-loadout", func(w http.ResponseWriter, r *http.Request) {
			account, ok := requireAccount(w, r, authService)
			if !ok {
				return
			}

			building, _, found := findBuilding(worldService, chi.URLParam(r, "buildingId"))
			if !found {
				writeError(w, r, http.StatusNotFound, "BUILDING_NOT_FOUND", "building does not exist")
				return
			}
			if building.Type != "guild" {
				writeError(w, r, http.StatusBadRequest, "INVALID_ACTION_STATE", "skill loadouts are only available at the Adventurers Guild")
				return
			}

			var request struct {
				SkillIDs []string `json:"skill_ids"`
			}
			if !decodeJSONBody(w, r, &request) {
				return
			}

			skillsState, err := characterService.SetSkillLoadout(account, request.SkillIDs)
			if err != nil {
				switch {
				case errors.Is(err, characters.ErrCharacterNotFound):
					writeError(w, r, http.StatusNotFound, "CHARACTER_NOT_FOUND", "create a character before configuring skills")
				case errors.Is(err, characters.ErrSkillLocked):
					writeError(w, r, http.StatusBadRequest, "SKILL_LOCKED", "loadout contains locked or unavailable skills")
				case errors.Is(err, characters.ErrSkillInvalidLoadout):
					writeError(w, r, http.StatusBadRequest, "SKILL_LOADOUT_INVALID", "loadout must contain up to four unique unlocked skills")
				default:
					writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update guild skill loadout")
				}
				return
			}

			writeEnvelope(w, r, http.StatusOK, map[string]any{
				"building_id": building.ID,
				"skills":      skillsState,
			})
		})

		r.Post("/arena/signup", func(w http.ResponseWriter, r *http.Request) {
			account, ok := requireAccount(w, r, authService)
			if !ok {
				return
			}
			character, exists := characterService.GetCharacterByAccount(account)
			if !exists {
				writeError(w, r, http.StatusNotFound, "CHARACTER_NOT_FOUND", "create a character before arena signup")
				return
			}

			state, err := buildCharacterState(account, characterService, inventoryService, questService, dungeonService, arenaService, worldService)
			if err != nil {
				writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to build character state before arena signup")
				return
			}

			entry, err := arenaService.Signup(character, state.CombatPower.PanelPowerScore, inventoryService.ComputeEquipmentScore(character), worldService.CurrentArenaStatus(worldService.CurrentTime()))
			if err != nil {
				switch {
				case errors.Is(err, arena.ErrSignupClosed):
					writeError(w, r, http.StatusBadRequest, "ARENA_SIGNUP_CLOSED", "arena signup window is closed")
				case errors.Is(err, arena.ErrAlreadySignedUp):
					writeError(w, r, http.StatusConflict, "ARENA_ALREADY_SIGNED_UP", "character already signed up for current arena")
				default:
					writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to signup arena")
				}
				return
			}

			_ = characterService.AppendEvents(character.CharacterID, world.WorldEvent{
				EventID:          requestID(r),
				EventType:        "arena.entry_accepted",
				Visibility:       "public",
				ActorCharacterID: character.CharacterID,
				ActorName:        character.Name,
				RegionID:         character.LocationRegionID,
				Summary:          fmt.Sprintf("%s signed up for arena.", character.Name),
				Payload: map[string]any{
					"character_id":      entry.CharacterID,
					"panel_power_score": entry.PanelPowerScore,
					"equipment_score":   entry.EquipmentScore,
				},
				OccurredAt: time.Now().Format(time.RFC3339),
			})

			writeEnvelope(w, r, http.StatusOK, map[string]any{
				"signed_up": true,
				"entry":     entry,
			})
		})

		r.Get("/arena/rating-board", func(w http.ResponseWriter, r *http.Request) {
			account, ok := requireAccount(w, r, authService)
			if !ok {
				return
			}
			character, exists := characterService.GetCharacterByAccount(account)
			if !exists {
				writeError(w, r, http.StatusNotFound, "CHARACTER_NOT_FOUND", "create a character before viewing arena rating board")
				return
			}
			entries := buildArenaRuntimeEntries(characterService, inventoryService, arenaService, dungeonService)
			view, err := arenaService.GetRatingBoard(character.CharacterID, entries)
			if err != nil {
				writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load arena rating board")
				return
			}
			writeEnvelope(w, r, http.StatusOK, view)
		})

		r.Post("/arena/rating-challenges/purchase", func(w http.ResponseWriter, r *http.Request) {
			account, ok := requireAccount(w, r, authService)
			if !ok {
				return
			}
			character, exists := characterService.GetCharacterByAccount(account)
			if !exists {
				writeError(w, r, http.StatusNotFound, "CHARACTER_NOT_FOUND", "create a character before purchasing arena challenges")
				return
			}
			price, _, err := arenaService.PurchaseRatingChallenge(character.CharacterID)
			if err != nil {
				switch {
				case errors.Is(err, arena.ErrChallengeWindow):
					writeError(w, r, http.StatusBadRequest, "ARENA_RATING_CLOSED", "arena rating challenges are only available Monday through Friday")
				case errors.Is(err, arena.ErrPurchaseCapReached):
					writeError(w, r, http.StatusBadRequest, "ARENA_PURCHASE_CAP_REACHED", "daily arena challenge purchase cap has been reached")
				default:
					writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to quote arena challenge purchase")
				}
				return
			}
			summary, err := characterService.SpendGold(character.CharacterID, price)
			if err != nil {
				if errors.Is(err, characters.ErrGoldInsufficient) {
					writeError(w, r, http.StatusBadRequest, "GOLD_INSUFFICIENT", "character does not have enough gold to buy another arena challenge")
					return
				}
				writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to spend gold for arena challenge purchase")
				return
			}
			if err := arenaService.ConfirmPurchasedRatingChallenge(character.CharacterID); err != nil {
				writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to grant purchased arena challenge")
				return
			}
			entries := buildArenaRuntimeEntries(characterService, inventoryService, arenaService, dungeonService)
			view, err := arenaService.GetRatingBoard(character.CharacterID, entries)
			if err != nil {
				writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to reload arena rating board")
				return
			}
			writeEnvelope(w, r, http.StatusOK, map[string]any{
				"price_gold": price,
				"character":  summary,
				"board":      view,
			})
		})

		r.Post("/arena/rating-challenges", func(w http.ResponseWriter, r *http.Request) {
			account, ok := requireAccount(w, r, authService)
			if !ok {
				return
			}
			character, exists := characterService.GetCharacterByAccount(account)
			if !exists {
				writeError(w, r, http.StatusNotFound, "CHARACTER_NOT_FOUND", "create a character before starting arena challenges")
				return
			}
			var request struct {
				TargetCharacterID string `json:"target_character_id"`
			}
			if !decodeJSONBody(w, r, &request) {
				return
			}
			entries := buildArenaRuntimeEntries(characterService, inventoryService, arenaService, dungeonService)
			result, err := arenaService.ResolveRatingChallenge(character.CharacterID, strings.TrimSpace(request.TargetCharacterID), entries)
			if err != nil {
				switch {
				case errors.Is(err, arena.ErrChallengeWindow):
					writeError(w, r, http.StatusBadRequest, "ARENA_RATING_CLOSED", "arena rating challenges are only available Monday through Friday")
				case errors.Is(err, arena.ErrNoChallengeAttempts):
					writeError(w, r, http.StatusBadRequest, "ARENA_NO_ATTEMPTS", "no arena challenge attempts remain today")
				case errors.Is(err, arena.ErrInvalidChallengeTarget):
					writeError(w, r, http.StatusBadRequest, "ARENA_INVALID_TARGET", "target is not in the current challenge candidate pool")
				default:
					writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to resolve arena rating challenge")
				}
				return
			}
			writeEnvelope(w, r, http.StatusOK, result)
		})

		r.Get("/arena/current", func(w http.ResponseWriter, r *http.Request) {
			entries := buildArenaRuntimeEntries(characterService, inventoryService, arenaService, dungeonService)
			writeEnvelope(w, r, http.StatusOK, arenaService.GetCurrent(worldService.CurrentArenaStatus(worldService.CurrentTime()), entries))
		})

		r.Get("/arena/entries", func(w http.ResponseWriter, r *http.Request) {
			limit := 20
			if rawLimit := strings.TrimSpace(r.URL.Query().Get("limit")); rawLimit != "" {
				if parsed, err := strconv.Atoi(rawLimit); err == nil && parsed > 0 {
					limit = parsed
				}
			}
			items := arenaService.ListEntries(arena.EntryListFilters{
				Cursor: strings.TrimSpace(r.URL.Query().Get("cursor")),
				Limit:  limit,
			})
			nextCursor := ""
			if len(items) == limit {
				nextCursor = items[len(items)-1].CharacterID
			}
			writeEnvelope(w, r, http.StatusOK, map[string]any{
				"items":       items,
				"next_cursor": nextCursor,
			})
		})

		r.Get("/arena/matches/{matchId}", func(w http.ResponseWriter, r *http.Request) {
			detail, err := arenaService.GetPublicMatchDetail(chi.URLParam(r, "matchId"), strings.TrimSpace(r.URL.Query().Get("detail_level")))
			if err != nil {
				switch {
				case errors.Is(err, arena.ErrArenaMatchNotFound):
					writeError(w, r, http.StatusNotFound, "ARENA_MATCH_NOT_FOUND", "arena match does not exist")
				default:
					writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load arena match detail")
				}
				return
			}
			writeEnvelope(w, r, http.StatusOK, detail)
		})

		r.Get("/arena/leaderboard", func(w http.ResponseWriter, r *http.Request) {
			writeEnvelope(w, r, http.StatusOK, map[string]any{"entries": arenaService.GetLeaderboard()})
		})

		r.Get("/me/arena-history", func(w http.ResponseWriter, r *http.Request) {
			account, ok := requireAccount(w, r, authService)
			if !ok {
				return
			}

			character, exists := characterService.GetCharacterByAccount(account)
			if !exists {
				writeError(w, r, http.StatusNotFound, "CHARACTER_NOT_FOUND", "create a character before viewing arena history")
				return
			}

			limit := 20
			if rawLimit := strings.TrimSpace(r.URL.Query().Get("limit")); rawLimit != "" {
				if parsed, err := strconv.Atoi(rawLimit); err == nil && parsed > 0 {
					limit = parsed
				}
			}

			items := arenaService.ListHistory(character.CharacterID, arena.HistoryFilters{
				Result:       strings.TrimSpace(r.URL.Query().Get("result")),
				TournamentID: strings.TrimSpace(r.URL.Query().Get("tournament_id")),
				Stage:        strings.TrimSpace(r.URL.Query().Get("stage")),
				Cursor:       strings.TrimSpace(r.URL.Query().Get("cursor")),
				Limit:        limit,
			})
			nextCursor := ""
			if len(items) == limit {
				nextCursor = items[len(items)-1].MatchID
			}

			writeEnvelope(w, r, http.StatusOK, map[string]any{
				"items":       items,
				"next_cursor": nextCursor,
			})
		})

		r.Get("/me/arena-history/{matchId}", func(w http.ResponseWriter, r *http.Request) {
			account, ok := requireAccount(w, r, authService)
			if !ok {
				return
			}

			character, exists := characterService.GetCharacterByAccount(account)
			if !exists {
				writeError(w, r, http.StatusNotFound, "CHARACTER_NOT_FOUND", "create a character before viewing arena history")
				return
			}

			detail, err := arenaService.GetHistoryDetail(character.CharacterID, chi.URLParam(r, "matchId"), strings.TrimSpace(r.URL.Query().Get("detail_level")))
			if err != nil {
				switch {
				case errors.Is(err, arena.ErrArenaMatchNotFound):
					writeError(w, r, http.StatusNotFound, "ARENA_MATCH_NOT_FOUND", "arena match does not exist")
				default:
					writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load arena match detail")
				}
				return
			}

			writeEnvelope(w, r, http.StatusOK, detail)
		})

		r.Get("/me/arena-title", func(w http.ResponseWriter, r *http.Request) {
			account, ok := requireAccount(w, r, authService)
			if !ok {
				return
			}
			character, exists := characterService.GetCharacterByAccount(account)
			if !exists {
				writeError(w, r, http.StatusNotFound, "CHARACTER_NOT_FOUND", "create a character before viewing arena title")
				return
			}
			entries := buildArenaRuntimeEntries(characterService, inventoryService, arenaService, dungeonService)
			title, found := arenaService.GetArenaTitle(character.CharacterID, entries)
			writeEnvelope(w, r, http.StatusOK, map[string]any{
				"title": foundIfAnyTitle(found, title),
			})
		})

		r.Get("/arena/duel", func(w http.ResponseWriter, r *http.Request) {
			charAID := strings.TrimSpace(r.URL.Query().Get("char_a_id"))
			charBID := strings.TrimSpace(r.URL.Query().Get("char_b_id"))
			if charAID == "" || charBID == "" {
				writeError(w, r, http.StatusBadRequest, "DUEL_MISSING_PARAMS", "char_a_id and char_b_id are required")
				return
			}
			summaryA, _, _, _, okA := characterService.GetRuntimeDetailByCharacterID(charAID)
			summaryB, _, _, _, okB := characterService.GetRuntimeDetailByCharacterID(charBID)
			if !okA || !okB {
				writeError(w, r, http.StatusNotFound, "CHARACTER_NOT_FOUND", "one or both characters not found")
				return
			}
			result := arenaService.SimulateDuel(summaryA, summaryB)
			writeEnvelope(w, r, http.StatusOK, result)
		})

		r.Get("/dungeons", func(w http.ResponseWriter, r *http.Request) {
			writeEnvelope(w, r, http.StatusOK, map[string]any{
				"items": dungeonService.ListDungeonDefinitions(),
			})
		})

		r.Get("/dungeons/{dungeonId}", func(w http.ResponseWriter, r *http.Request) {
			dungeonID := chi.URLParam(r, "dungeonId")
			definition, ok := dungeonService.GetDungeonDefinition(dungeonID)
			if !ok {
				writeError(w, r, http.StatusNotFound, "DUNGEON_NOT_FOUND", "dungeon definition does not exist")
				return
			}

			writeEnvelope(w, r, http.StatusOK, definition)
		})

		r.Post("/dungeons/{dungeonId}/enter", func(w http.ResponseWriter, r *http.Request) {
			account, ok := requireAccount(w, r, authService)
			if !ok {
				return
			}

			character, limits, err := currentCharacterWithLimits(account, characterService, worldService)
			if err != nil {
				writeError(w, r, http.StatusNotFound, "CHARACTER_NOT_FOUND", "create a character before entering dungeons")
				return
			}

			var request struct {
				Difficulty    string   `json:"difficulty"`
				PotionLoadout []string `json:"potion_loadout"`
			}
			if r.ContentLength > 0 {
				if !decodeJSONBody(w, r, &request) {
					return
				}
			}

			difficulty := strings.TrimSpace(request.Difficulty)
			if difficulty == "" {
				difficulty = strings.TrimSpace(r.URL.Query().Get("difficulty"))
			}
			potionBag, err := inventoryService.BuildPotionLoadout(character, request.PotionLoadout)
			if err != nil {
				writeError(w, r, http.StatusBadRequest, "DUNGEON_POTION_LOADOUT_INVALID", "select up to two owned potion types before entering a dungeon")
				return
			}

			state, err := buildCharacterState(account, characterService, inventoryService, questService, dungeonService, arenaService, worldService)
			if err != nil {
				writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to build character state before dungeon entry")
				return
			}
			player := dungeons.BuildPlayerCombatant(state.Character, state.Stats, state.Skills)

			run, err := dungeonService.EnterDungeon(character, limits, player, chi.URLParam(r, "dungeonId"), difficulty, request.PotionLoadout, potionBag)
			if err != nil {
				switch {
				case errors.Is(err, dungeons.ErrDungeonNotFound):
					writeError(w, r, http.StatusNotFound, "DUNGEON_NOT_FOUND", "dungeon definition does not exist")
				case errors.Is(err, dungeons.ErrDungeonRunAlreadyActive):
					writeError(w, r, http.StatusConflict, "DUNGEON_RUN_ALREADY_ACTIVE", "character already has an active dungeon run")
				case errors.Is(err, dungeons.ErrDungeonRewardClaimCapReached):
					writeError(w, r, http.StatusBadRequest, "DUNGEON_REWARD_CLAIM_LIMIT_REACHED", "daily dungeon reward claim cap has been reached")
				case errors.Is(err, dungeons.ErrDungeonPotionLoadoutInvalid):
					writeError(w, r, http.StatusBadRequest, "DUNGEON_POTION_LOADOUT_INVALID", "select up to two owned potion types before entering a dungeon")
				default:
					writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to enter dungeon")
				}
				return
			}
			inventoryService.ConsumeConsumables(character.CharacterID, dungeons.PotionUsageFromLog(run.RecentBattleLog))

			definition, _ := dungeonService.GetDungeonDefinition(run.DungeonID)
			completedQuests := []characters.QuestSummary{}
			if run.RunStatus == "cleared" {
				_, completedQuests = questService.ProgressDungeonQuests(character, definition.RegionID, limits)
			}
			events := []world.WorldEvent{{
				EventID:          requestID(r),
				EventType:        "dungeon.entered",
				Visibility:       "public",
				ActorCharacterID: character.CharacterID,
				ActorName:        character.Name,
				RegionID:         definition.RegionID,
				Summary:          fmt.Sprintf("%s entered %s.", character.Name, definition.Name),
				Payload: map[string]any{
					"run_id":     run.RunID,
					"dungeon_id": run.DungeonID,
				},
				OccurredAt: time.Now().Format(time.RFC3339),
			}}
			if run.RunStatus == "cleared" {
				events = append(events, world.WorldEvent{
					EventID:          fmt.Sprintf("evt_dungeon_cleared_%s", run.RunID),
					EventType:        "dungeon.cleared",
					Visibility:       "public",
					ActorCharacterID: character.CharacterID,
					ActorName:        character.Name,
					RegionID:         definition.RegionID,
					Summary:          fmt.Sprintf("%s cleared %s.", character.Name, definition.Name),
					Payload: map[string]any{
						"run_id":         run.RunID,
						"dungeon_id":     run.DungeonID,
						"current_rating": run.CurrentRating,
					},
					OccurredAt: time.Now().Format(time.RFC3339),
				})
			} else {
				events = append(events, world.WorldEvent{
					EventID:          fmt.Sprintf("evt_dungeon_failed_%s", run.RunID),
					EventType:        "dungeon.failed",
					Visibility:       "public",
					ActorCharacterID: character.CharacterID,
					ActorName:        character.Name,
					RegionID:         definition.RegionID,
					Summary:          fmt.Sprintf("%s failed in %s.", character.Name, definition.Name),
					Payload: map[string]any{
						"run_id":               run.RunID,
						"dungeon_id":           run.DungeonID,
						"highest_room_cleared": run.HighestRoomCleared,
					},
					OccurredAt: time.Now().Format(time.RFC3339),
				})
			}
			_ = characterService.AppendEvents(character.CharacterID, events...)
			for _, quest := range completedQuests {
				_ = characterService.AppendEvents(character.CharacterID, world.WorldEvent{
					EventID:          fmt.Sprintf("evt_quest_completed_dungeon_%s", quest.QuestID),
					EventType:        "quest.completed",
					Visibility:       "public",
					ActorCharacterID: character.CharacterID,
					ActorName:        character.Name,
					RegionID:         definition.RegionID,
					Summary:          fmt.Sprintf("%s completed %s.", character.Name, quest.Title),
					Payload: map[string]any{
						"quest_id":    quest.QuestID,
						"quest_title": quest.Title,
						"dungeon_id":  run.DungeonID,
						"run_id":      run.RunID,
					},
					OccurredAt: time.Now().Format(time.RFC3339),
				})
			}
			writeEnvelope(w, r, http.StatusOK, run)
		})

		r.Get("/me/runs/active", func(w http.ResponseWriter, r *http.Request) {
			account, ok := requireAccount(w, r, authService)
			if !ok {
				return
			}

			character, exists := characterService.GetCharacterByAccount(account)
			if !exists {
				writeError(w, r, http.StatusNotFound, "CHARACTER_NOT_FOUND", "create a character before viewing dungeon runs")
				return
			}

			run, err := dungeonService.GetActiveRun(character.CharacterID)
			if err != nil {
				writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load active dungeon run")
				return
			}

			if run == nil {
				writeEnvelope(w, r, http.StatusOK, nil)
				return
			}

			writeEnvelope(w, r, http.StatusOK, dungeonService.BuildRunPayload(*run, r.URL.Query().Get("detail_level")))
		})

		r.Get("/me/runs", func(w http.ResponseWriter, r *http.Request) {
			account, ok := requireAccount(w, r, authService)
			if !ok {
				return
			}

			character, exists := characterService.GetCharacterByAccount(account)
			if !exists {
				writeError(w, r, http.StatusNotFound, "CHARACTER_NOT_FOUND", "create a character before viewing dungeon runs")
				return
			}

			limit := 20
			if rawLimit := strings.TrimSpace(r.URL.Query().Get("limit")); rawLimit != "" {
				if parsed, err := strconv.Atoi(rawLimit); err == nil && parsed > 0 {
					limit = parsed
				}
			}
			items := dungeonService.ListRuns(character.CharacterID, dungeons.RunListFilters{
				DungeonID:  r.URL.Query().Get("dungeon_id"),
				Difficulty: r.URL.Query().Get("difficulty"),
				Result:     r.URL.Query().Get("result"),
				Cursor:     r.URL.Query().Get("cursor"),
				Limit:      limit,
			})

			nextCursor := ""
			if len(items) == limit {
				nextCursor = items[len(items)-1].RunID
			}

			writeEnvelope(w, r, http.StatusOK, map[string]any{
				"items":       items,
				"next_cursor": nextCursor,
			})
		})

		r.Get("/me/runs/{runId}", func(w http.ResponseWriter, r *http.Request) {
			account, ok := requireAccount(w, r, authService)
			if !ok {
				return
			}

			character, exists := characterService.GetCharacterByAccount(account)
			if !exists {
				writeError(w, r, http.StatusNotFound, "CHARACTER_NOT_FOUND", "create a character before viewing dungeon runs")
				return
			}

			run, err := dungeonService.GetRun(character.CharacterID, chi.URLParam(r, "runId"))
			if err != nil {
				switch {
				case errors.Is(err, dungeons.ErrDungeonRunNotFound):
					writeError(w, r, http.StatusNotFound, "DUNGEON_RUN_NOT_FOUND", "dungeon run does not exist")
				case errors.Is(err, dungeons.ErrDungeonRunForbidden):
					writeError(w, r, http.StatusForbidden, "DUNGEON_RUN_FORBIDDEN", "dungeon run does not belong to caller")
				default:
					writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load dungeon run")
				}
				return
			}

			writeEnvelope(w, r, http.StatusOK, dungeonService.BuildRunPayload(run, r.URL.Query().Get("detail_level")))
		})

		r.Post("/me/runs/{runId}/claim", func(w http.ResponseWriter, r *http.Request) {
			account, ok := requireAccount(w, r, authService)
			if !ok {
				return
			}

			character, limits, err := currentCharacterWithLimits(account, characterService, worldService)
			if err != nil {
				writeError(w, r, http.StatusNotFound, "CHARACTER_NOT_FOUND", "create a character before claiming rewards")
				return
			}

			run, claimPackage, err := dungeonService.PreviewRunRewards(character.CharacterID, chi.URLParam(r, "runId"), limits)
			if err != nil {
				switch {
				case errors.Is(err, dungeons.ErrDungeonRunNotFound):
					writeError(w, r, http.StatusNotFound, "DUNGEON_RUN_NOT_FOUND", "dungeon run does not exist")
				case errors.Is(err, dungeons.ErrDungeonRunForbidden):
					writeError(w, r, http.StatusForbidden, "DUNGEON_RUN_FORBIDDEN", "dungeon run does not belong to caller")
				case errors.Is(err, dungeons.ErrDungeonRewardClaimNotAllowed):
					writeError(w, r, http.StatusBadRequest, "DUNGEON_REWARD_NOT_CLAIMABLE", "run rewards are not claimable")
				case errors.Is(err, dungeons.ErrDungeonRewardClaimCapReached):
					writeError(w, r, http.StatusBadRequest, "DUNGEON_REWARD_CLAIM_LIMIT_REACHED", "daily dungeon reward claim cap has been reached")
				default:
					writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to claim dungeon rewards")
				}
				return
			}

			rewardItemCatalogIDs, err := grantDungeonInventoryRewards(inventoryService, character, claimPackage)
			if err != nil {
				writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to grant dungeon item rewards")
				return
			}

			if _, _, _, err := characterService.ApplyDungeonRewardClaim(character.CharacterID, run.RunID, run.DungeonID, claimPackage.RewardGold, claimPackage.Rating, rewardItemCatalogIDs, claimPackage.MaterialDrops); err != nil {
				if errors.Is(err, characters.ErrDungeonRewardClaimCap) {
					writeError(w, r, http.StatusBadRequest, "DUNGEON_REWARD_CLAIM_LIMIT_REACHED", "daily dungeon reward claim cap has been reached")
					return
				}

				writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to apply dungeon rewards")
				return
			}
			run, err = dungeonService.FinalizeRunRewards(character.CharacterID, run.RunID)
			if err != nil {
				writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to finalize dungeon rewards")
				return
			}

			writeEnvelope(w, r, http.StatusOK, run)
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
				cursor := strings.TrimSpace(r.URL.Query().Get("cursor"))
				snapshot := characterService.SnapshotRuntime()
				items := buildPublicEvents(snapshot.Events, len(snapshot.Events))
				start := 0
				if cursor != "" {
					if value, err := strconv.Atoi(cursor); err == nil && value >= 0 && value < len(items) {
						start = value
					}
				}
				end := start + limit
				if end > len(items) {
					end = len(items)
				}

				nextCursor := any(nil)
				if end < len(items) {
					nextCursor = strconv.Itoa(end)
				}

				writeEnvelope(w, r, http.StatusOK, map[string]any{
					"items":       items[start:end],
					"next_cursor": nextCursor,
				})
			})

			r.Get("/events/stream", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/event-stream")
				w.Header().Set("Cache-Control", "no-cache")
				w.Header().Set("Connection", "keep-alive")

				events := buildPublicEvents(characterService.SnapshotRuntime().Events, 1)
				if len(events) == 0 {
					fmt.Fprintf(w, "event: world.counter.updated\n")
					fmt.Fprintf(w, "data: {\"status\":\"idle\"}\n\n")
				} else {
					payload, _ := json.Marshal(events[0])
					fmt.Fprintf(w, "event: world.event.created\n")
					fmt.Fprintf(w, "data: %s\n\n", payload)
				}

				if flusher, ok := w.(http.Flusher); ok {
					flusher.Flush()
				}
			})

			r.Get("/events/{eventId}", func(w http.ResponseWriter, r *http.Request) {
				eventID := chi.URLParam(r, "eventId")
				snapshot := characterService.SnapshotRuntime()
				items := buildPublicEvents(snapshot.Events, len(snapshot.Events))
				for _, event := range items {
					if event.EventID == eventID {
						writeEnvelope(w, r, http.StatusOK, event)
						return
					}
				}

				writeError(w, r, http.StatusNotFound, "PUBLIC_EVENT_NOT_FOUND", "public event does not exist")
			})

			r.Get("/leaderboards", func(w http.ResponseWriter, r *http.Request) {
				snapshot := characterService.SnapshotRuntime()
				writeEnvelope(w, r, http.StatusOK, buildPublicLeaderboards(snapshot))
			})

			r.Get("/bots", func(w http.ResponseWriter, r *http.Request) {
				limit := parseLimit(r, 20, 100)
				cursor := strings.TrimSpace(r.URL.Query().Get("cursor"))
				queryName := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
				queryCharacterID := strings.TrimSpace(r.URL.Query().Get("character_id"))
				snapshot := characterService.SnapshotRuntime()
				items := buildPublicBots(snapshot.Characters, characterService, inventoryService, dungeonService)
				if queryName != "" || queryCharacterID != "" {
					filtered := make([]map[string]any, 0, len(items))
					for _, item := range items {
						summary, _ := item["character_summary"].(characters.Summary)
						if queryCharacterID != "" && summary.CharacterID != queryCharacterID {
							continue
						}
						if queryName != "" && !strings.Contains(strings.ToLower(summary.Name), queryName) {
							continue
						}
						filtered = append(filtered, item)
					}
					items = filtered
				}
				start := 0
				if cursor != "" {
					if value, err := strconv.Atoi(cursor); err == nil && value >= 0 && value < len(items) {
						start = value
					}
				}
				end := start + limit
				if end > len(items) {
					end = len(items)
				}

				nextCursor := any(nil)
				if end < len(items) {
					nextCursor = strconv.Itoa(end)
				}

				writeEnvelope(w, r, http.StatusOK, map[string]any{
					"items":       items[start:end],
					"next_cursor": nextCursor,
				})
			})

			r.Get("/bots/{botId}", func(w http.ResponseWriter, r *http.Request) {
				botID := chi.URLParam(r, "botId")
				summary, stats, limits, events, ok := characterService.GetRuntimeDetailByCharacterID(botID)
				if !ok {
					writeError(w, r, http.StatusNotFound, "BOT_NOT_FOUND", "public bot does not exist")
					return
				}

				now := worldService.CurrentTime()
				questHistory := buildQuestHistory(events, 7, now)
				dungeonHistory := buildDungeonHistory(events, 7, now, dungeonService)
				completedToday := filterQuestHistoryByDay(questHistory, now)
				dungeonToday := filterDungeonHistoryByDay(dungeonHistory, now)
				recentRuns := dungeonService.ListRunsByCharacter(botID)
				if len(recentRuns) > 5 {
					recentRuns = recentRuns[:5]
				}

				inventoryView := inventoryService.GetInventory(summary)
				combatPower, itemScores := buildCombatPower(summary, stats, inventoryView, dungeonService)

				writeEnvelope(w, r, http.StatusOK, map[string]any{
					"character_summary":      summary,
					"stats_snapshot":         stats,
					"equipment":              inventoryView,
					"equipment_item_scores":  itemScores,
					"combat_power":           combatPower,
					"daily_limits":           limits,
					"active_quests":          []any{},
					"recent_runs":            recentRuns,
					"arena_history":          arenaService.ListHistory(botID, arena.HistoryFilters{Limit: 10}),
					"recent_events":          events,
					"completed_quests_today": completedToday,
					"dungeon_runs_today":     dungeonToday,
					"quest_history_7d":       questHistory,
					"dungeon_history_7d":     dungeonHistory,
				})
			})

			r.Get("/bots/{botId}/quests/history", func(w http.ResponseWriter, r *http.Request) {
				botID := chi.URLParam(r, "botId")
				_, _, _, events, ok := characterService.GetRuntimeDetailByCharacterID(botID)
				if !ok {
					writeError(w, r, http.StatusNotFound, "BOT_NOT_FOUND", "public bot does not exist")
					return
				}

				days := parseDays(r, 7, 7)
				limit := parseLimit(r, 20, 100)
				history := buildQuestHistory(events, days, time.Now())
				if limit < len(history) {
					history = history[:limit]
				}

				writeEnvelope(w, r, http.StatusOK, map[string]any{
					"items":       history,
					"next_cursor": nil,
				})
			})

			r.Get("/bots/{botId}/dungeon-runs", func(w http.ResponseWriter, r *http.Request) {
				botID := chi.URLParam(r, "botId")
				_, _, _, events, ok := characterService.GetRuntimeDetailByCharacterID(botID)
				if !ok {
					writeError(w, r, http.StatusNotFound, "BOT_NOT_FOUND", "public bot does not exist")
					return
				}

				days := parseDays(r, 7, 7)
				limit := parseLimit(r, 20, 100)
				history := buildDungeonHistory(events, days, time.Now(), dungeonService)
				if limit < len(history) {
					history = history[:limit]
				}

				writeEnvelope(w, r, http.StatusOK, map[string]any{
					"items":       history,
					"next_cursor": nil,
				})
			})

			r.Get("/bots/{botId}/dungeon-runs/{runId}", func(w http.ResponseWriter, r *http.Request) {
				botID := chi.URLParam(r, "botId")
				runID := chi.URLParam(r, "runId")
				_, _, _, events, ok := characterService.GetRuntimeDetailByCharacterID(botID)
				if !ok {
					writeError(w, r, http.StatusNotFound, "BOT_NOT_FOUND", "public bot does not exist")
					return
				}

				payload, ok := buildPublicDungeonRunDetail(botID, runID, events, dungeonService)
				if !ok {
					writeError(w, r, http.StatusNotFound, "DUNGEON_RUN_NOT_FOUND", "dungeon run does not exist")
					return
				}

				writeEnvelope(w, r, http.StatusOK, payload)
			})
		})
	})

	return &Server{
		authService:      authService,
		arenaService:     arenaService,
		characterService: characterService,
		dungeonService:   dungeonService,
		inventoryService: inventoryService,
		questService:     questService,
		worldService:     worldService,
		worldBossService: worldBossService,
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

func parseDays(r *http.Request, defaultValue, maxValue int) int {
	raw := strings.TrimSpace(r.URL.Query().Get("days"))
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

func buildCharacterState(account auth.Account, characterService *characters.Service, inventoryService *inventory.Service, questService *quests.Service, dungeonService *dungeons.Service, arenaService *arena.Service, worldService *world.Service) (characters.StateView, error) {
	state, err := characterService.GetState(account, worldService)
	if err != nil {
		return characters.StateView{}, err
	}

	board := questService.ListQuests(state.Character, state.Limits)

	state.Stats = inventoryService.DeriveStats(state.Character, state.Stats)
	if arenaService != nil {
		titleEntries := buildArenaBaseEntries(characterService, inventoryService, dungeonService)
		if title, found := arenaService.GetArenaTitle(state.Character.CharacterID, titleEntries); found {
			state.Stats = applyArenaTitleBonusSnapshot(state.Stats, title.BonusSnapshot)
		}
	}
	inventoryView := inventoryService.GetInventory(state.Character)
	state.CombatPower, _ = buildCombatPower(state.Character, state.Stats, inventoryView, dungeonService)
	state.SlotEnhancements = inventoryService.ListSlotEnhancements(state.Character)
	state.Objectives = activeObjectivesFromBoard(board)
	state.ValidActions = append(state.ValidActions, questService.ListRuntimeActions(state.Character.CharacterID)...)
	state.DungeonDaily = buildDungeonDailyHint(state, dungeonService)
	return state, nil
}

func buildDungeonDailyHint(state characters.StateView, dungeonService *dungeons.Service) characters.DungeonDailyHint {
	remaining := state.Limits.DungeonEntryCap - state.Limits.DungeonEntryUsed
	if remaining < 0 {
		remaining = 0
	}

	runs := dungeonService.ListRunsByCharacter(state.Character.CharacterID)
	pending := make([]string, 0, len(runs))
	for _, run := range runs {
		if run.RunStatus == "cleared" && run.RewardClaimable {
			pending = append(pending, run.RunID)
		}
	}

	hint := characters.DungeonDailyHint{
		HasRemainingQuota:  state.Limits.DungeonEntryUsed < state.Limits.DungeonEntryCap,
		RemainingClaims:    remaining,
		HasClaimableRun:    len(pending) > 0,
		PendingClaimRunIDs: pending,
	}

	return hint
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
			TravelCostGold:   region.TravelCostGold,
			Population:       populationByRegion[region.ID],
			RecentEventCount: eventCountByRegion[region.ID],
			Highlight:        highlight,
			BuildingCount:    len(detail.Buildings),
			RegionGameplay: world.RegionGameplay{
				InteractionLayer:       detail.InteractionLayer,
				RiskLevel:              detail.RiskLevel,
				FacilityFocus:          detail.FacilityFocus,
				EncounterFamily:        detail.EncounterFamily,
				CurioStatus:            runtimeCurioStatus(detail, populationByRegion[region.ID], eventCountByRegion[region.ID]),
				CurioHint:              detail.CurioHint,
				LinkedDungeon:          detail.LinkedDungeon,
				ParentRegionID:         detail.ParentRegionID,
				HostileEncounters:      detail.HostileEncounters,
				AvailableRegionActions: detail.AvailableRegionActions,
			},
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

func appendUniqueString(values []string, candidate string) []string {
	for _, item := range values {
		if item == candidate {
			return values
		}
	}

	return append(values, candidate)
}

func buildDungeonPreparationEntry(character characters.Summary, definition dungeons.DefinitionView, inventoryView inventory.InventoryView, combatPower characters.CombatPowerSummary) map[string]any {
	recommendedScore := heuristicRecommendedEquipmentScore(definition)
	scoreGap := recommendedScore - inventoryView.EquipmentScore
	if scoreGap < 0 {
		scoreGap = 0
	}
	recommendedPower := heuristicRecommendedPanelPower(definition)
	powerGap := recommendedPower - combatPower.PanelPowerScore
	if powerGap < 0 {
		powerGap = 0
	}
	readiness := "ready"
	switch {
	case combatPower.PanelPowerScore+250 < recommendedPower:
		readiness = "underprepared"
	case combatPower.PanelPowerScore < recommendedPower:
		readiness = "caution"
	}

	inventoryUpgrades := make([]map[string]any, 0, len(inventoryView.UpgradeHints))
	shopUpgrades := make([]map[string]any, 0, len(inventoryView.UpgradeHints))
	for _, hint := range inventoryView.UpgradeHints {
		entry := map[string]any{
			"source":              hint.Source,
			"item_id":             hint.ItemID,
			"catalog_id":          hint.CatalogID,
			"name":                hint.Name,
			"slot":                hint.Slot,
			"score_delta":         hint.ScoreDelta,
			"price_gold":          hint.PriceGold,
			"affordable":          hint.Affordable,
			"directly_equippable": hint.DirectlyEquipable,
		}
		if hint.Source == "inventory" {
			inventoryUpgrades = append(inventoryUpgrades, entry)
		} else if hint.Source == "shop" {
			shopUpgrades = append(shopUpgrades, entry)
		}
	}

	potionOptions := make([]map[string]any, 0, len(inventoryView.PotionLoadoutOptions))
	for _, option := range inventoryView.PotionLoadoutOptions {
		potionOptions = append(potionOptions, map[string]any{
			"catalog_id":     option.CatalogID,
			"name":           option.Name,
			"family":         option.Family,
			"tier":           option.Tier,
			"quantity_owned": option.QuantityOwned,
			"available_now":  option.AvailableNow,
			"can_purchase":   option.CanPurchase,
			"recommended":    option.Recommended,
		})
	}

	steps := make([]string, 0, 3)
	if readiness != "ready" {
		if len(inventoryUpgrades) > 0 {
			steps = append(steps, "equip_stronger_inventory_items")
		}
		if affordableUpgradeCount(shopUpgrades) > 0 {
			steps = append(steps, "buy_equipment_upgrade")
		}
	}
	if recommendedPotionAvailable(potionOptions) {
		steps = append(steps, "review_potion_loadout")
	}

	return map[string]any{
		"current_power":                 combatPower.PanelPowerScore,
		"recommended_power":             recommendedPower,
		"power_gap":                     powerGap,
		"dungeon_id":                    definition.DungeonID,
		"name":                          definition.Name,
		"current_equipment_score":       inventoryView.EquipmentScore,
		"recommended_equipment_score":   recommendedScore,
		"score_gap":                     scoreGap,
		"readiness":                     readiness,
		"inventory_upgrade_count":       len(inventoryUpgrades),
		"affordable_shop_upgrade_count": affordableUpgradeCount(shopUpgrades),
		"inventory_upgrades":            inventoryUpgrades,
		"shop_upgrades":                 shopUpgrades,
		"potion_options":                potionOptions,
		"suggested_preparation_steps":   steps,
		"available_gold":                character.Gold,
	}
}

func buildArenaBaseEntries(characterService *characters.Service, inventoryService *inventory.Service, dungeonService *dungeons.Service) []arena.Entry {
	snapshot := characterService.SnapshotRuntime()
	entries := make([]arena.Entry, 0, len(snapshot.Characters))
	for _, summary := range snapshot.Characters {
		_, baseStats, _, _, ok := characterService.GetRuntimeDetailByCharacterID(summary.CharacterID)
		if !ok {
			continue
		}
		inv := inventoryService.GetInventory(summary)
		stats := inventoryService.DeriveStats(summary, baseStats)
		combatPower, _ := buildCombatPower(summary, stats, inv, dungeonService)
		entries = append(entries, arena.Entry{
			CharacterID:     summary.CharacterID,
			CharacterName:   summary.Name,
			Class:           summary.Class,
			WeaponStyle:     summary.WeaponStyle,
			PanelPowerScore: combatPower.PanelPowerScore,
			EquipmentScore:  inv.EquipmentScore,
			IsNPC:           strings.HasPrefix(summary.CharacterID, "npc_"),
		})
	}
	return entries
}

func buildArenaRuntimeEntries(characterService *characters.Service, inventoryService *inventory.Service, arenaService *arena.Service, dungeonService *dungeons.Service) []arena.Entry {
	baseEntries := buildArenaBaseEntries(characterService, inventoryService, dungeonService)
	if arenaService == nil {
		return baseEntries
	}

	snapshot := characterService.SnapshotRuntime()
	entries := make([]arena.Entry, 0, len(snapshot.Characters))
	for _, summary := range snapshot.Characters {
		_, baseStats, _, _, ok := characterService.GetRuntimeDetailByCharacterID(summary.CharacterID)
		if !ok {
			continue
		}
		inv := inventoryService.GetInventory(summary)
		stats := inventoryService.DeriveStats(summary, baseStats)
		if title, found := arenaService.GetArenaTitle(summary.CharacterID, baseEntries); found {
			stats = applyArenaTitleBonusSnapshot(stats, title.BonusSnapshot)
		}
		combatPower, _ := buildCombatPower(summary, stats, inv, dungeonService)
		entries = append(entries, arena.Entry{
			CharacterID:     summary.CharacterID,
			CharacterName:   summary.Name,
			Class:           summary.Class,
			WeaponStyle:     summary.WeaponStyle,
			PanelPowerScore: combatPower.PanelPowerScore,
			EquipmentScore:  inv.EquipmentScore,
			IsNPC:           strings.HasPrefix(summary.CharacterID, "npc_"),
		})
	}
	return entries
}

func applyArenaTitleBonusSnapshot(stats characters.StatsSnapshot, bonus map[string]any) characters.StatsSnapshot {
	applyInt := func(current int, key string) int {
		if raw, ok := bonus[key].(float64); ok && raw > 0 {
			return int(math.Round(float64(current) * (1 + raw)))
		}
		return current
	}
	applyFloat := func(current float64, key string) float64 {
		if raw, ok := bonus[key].(float64); ok && raw > 0 {
			return current * (1 + raw)
		}
		return current
	}

	stats.MaxHP = applyInt(stats.MaxHP, "max_hp")
	stats.PhysicalAttack = applyInt(stats.PhysicalAttack, "physical_attack")
	stats.MagicAttack = applyInt(stats.MagicAttack, "magic_attack")
	stats.PhysicalDefense = applyInt(stats.PhysicalDefense, "physical_defense")
	stats.MagicDefense = applyInt(stats.MagicDefense, "magic_defense")
	stats.Speed = applyInt(stats.Speed, "speed")
	stats.HealingPower = applyInt(stats.HealingPower, "healing_power")
	stats.CritRate = applyFloat(stats.CritRate, "crit_rate")
	stats.CritDamage = applyFloat(stats.CritDamage, "crit_damage")
	stats.BlockRate = applyFloat(stats.BlockRate, "block_rate")
	stats.Precision = applyFloat(stats.Precision, "precision")
	stats.EvasionRate = applyFloat(stats.EvasionRate, "evasion_rate")
	stats.PhysicalMastery = applyFloat(stats.PhysicalMastery, "physical_mastery")
	stats.MagicMastery = applyFloat(stats.MagicMastery, "magic_mastery")
	return stats
}

func foundIfAnyTitle(found bool, title arena.ArenaTitleView) any {
	if !found {
		return nil
	}
	return title
}

func heuristicRecommendedPanelPower(definition dungeons.DefinitionView) int {
	minPower := recommendedPowerForLevel(definition.RecommendedLevelMin)
	maxPower := recommendedPowerForLevel(definition.RecommendedLevelMax)
	if maxPower < minPower {
		maxPower = minPower
	}
	return (minPower + maxPower) / 2
}

func heuristicRecommendedEquipmentScore(definition dungeons.DefinitionView) int {
	base := definition.RecommendedLevelMin*22 + definition.RecommendedLevelMax*12
	if base < 40 {
		base = 40
	}
	return base
}

func affordableUpgradeCount(items []map[string]any) int {
	total := 0
	for _, item := range items {
		if affordable, _ := item["affordable"].(bool); affordable {
			total++
		}
	}
	return total
}

func recommendedPotionAvailable(items []map[string]any) bool {
	for _, item := range items {
		recommended, _ := item["recommended"].(bool)
		if !recommended {
			continue
		}
		if availableNow, _ := item["available_now"].(bool); availableNow {
			return true
		}
		if canPurchase, _ := item["can_purchase"].(bool); canPurchase {
			return true
		}
	}
	return false
}

func buildQuestHistory(events []world.WorldEvent, days int, now time.Time) []map[string]any {
	windowStart := now.Add(-time.Duration(days) * 24 * time.Hour)
	items := make([]map[string]any, 0, len(events))
	for _, event := range events {
		if event.EventType != "quest.submitted" {
			continue
		}

		occurredAt, err := time.Parse(time.RFC3339, event.OccurredAt)
		if err != nil || occurredAt.Before(windowStart) {
			continue
		}

		items = append(items, map[string]any{
			"quest_id":       stringPayload(event.Payload, "quest_id"),
			"quest_name":     stringPayload(event.Payload, "quest_title"),
			"status":         "submitted",
			"accepted_at":    nil,
			"submitted_at":   event.OccurredAt,
			"reward_summary": map[string]any{"gold": intPayload(event.Payload, "reward_gold"), "reputation": intPayload(event.Payload, "reward_reputation")},
		})
	}

	return items
}

func buildDungeonHistory(events []world.WorldEvent, days int, now time.Time, dungeonService *dungeons.Service) []map[string]any {
	windowStart := now.Add(-time.Duration(days) * 24 * time.Hour)
	runsByID := make(map[string]dungeons.RunView)
	for _, event := range events {
		runID := stringPayload(event.Payload, "run_id")
		if runID == "" {
			continue
		}
	}

	items := make([]map[string]any, 0)
	for _, event := range events {
		if event.EventType != "dungeon.cleared" && event.EventType != "dungeon.loot_granted" {
			continue
		}

		occurredAt, err := time.Parse(time.RFC3339, event.OccurredAt)
		if err != nil || occurredAt.Before(windowStart) {
			continue
		}

		runID := stringPayload(event.Payload, "run_id")
		dungeonID := stringPayload(event.Payload, "dungeon_id")
		runView, ok := runsByID[runID]
		if !ok {
			runView = findRunInServiceByID(dungeonService, runID)
			runsByID[runID] = runView
		}
		definition, _ := dungeonService.GetDungeonDefinition(dungeonID)

		items = append(items, map[string]any{
			"run_id":         runID,
			"dungeon_id":     dungeonID,
			"dungeon_name":   definition.Name,
			"started_at":     firstNonEmpty(runView.StartedAt, event.OccurredAt),
			"resolved_at":    firstNonEmpty(runView.ResolvedAt, event.OccurredAt),
			"result":         stringPayloadFromRun(runView.RunStatus, "cleared"),
			"reward_summary": map[string]any{"gold": intPayload(event.Payload, "reward_gold"), "rating": event.Payload["current_rating"]},
		})
	}

	slices.SortFunc(items, func(left, right map[string]any) int {
		leftAt, _ := left["resolved_at"].(string)
		rightAt, _ := right["resolved_at"].(string)
		return cmp.Compare(rightAt, leftAt)
	})

	return dedupeDungeonHistory(items)
}

func filterQuestHistoryByDay(items []map[string]any, now time.Time) []map[string]any {
	start := businessDayStart(now)
	filtered := make([]map[string]any, 0, len(items))
	for _, item := range items {
		submittedAt, _ := item["submitted_at"].(string)
		parsed, err := time.Parse(time.RFC3339, submittedAt)
		if err != nil || parsed.Before(start) {
			continue
		}
		filtered = append(filtered, item)
	}

	return filtered
}

func filterDungeonHistoryByDay(items []map[string]any, now time.Time) []map[string]any {
	start := businessDayStart(now)
	filtered := make([]map[string]any, 0, len(items))
	for _, item := range items {
		resolvedAt, _ := item["resolved_at"].(string)
		parsed, err := time.Parse(time.RFC3339, resolvedAt)
		if err != nil || parsed.Before(start) {
			continue
		}
		filtered = append(filtered, item)
	}

	return filtered
}

func businessDayStart(now time.Time) time.Time {
	loc := now.Location()
	start := time.Date(now.Year(), now.Month(), now.Day(), 4, 0, 0, 0, loc)
	if now.Before(start) {
		return start.Add(-24 * time.Hour)
	}
	return start
}

func stringPayload(payload map[string]any, key string) string {
	if payload == nil {
		return ""
	}
	value, _ := payload[key].(string)
	return strings.TrimSpace(value)
}

func findRunInServiceByID(dungeonService *dungeons.Service, runID string) dungeons.RunView {
	allRuns := dungeonService.ListRunsByCharacter("")
	for _, run := range allRuns {
		if run.RunID == runID {
			return run
		}
	}
	return dungeons.RunView{}
}

func buildPublicDungeonRunDetail(characterID, runID string, events []world.WorldEvent, dungeonService *dungeons.Service) (map[string]any, bool) {
	if run, err := dungeonService.GetRun(characterID, runID); err == nil {
		return publicDungeonRunPayloadFromRun(run, dungeonService), true
	}

	var enteredEvent *world.WorldEvent
	var clearedEvent *world.WorldEvent
	var lootEvent *world.WorldEvent

	for index := range events {
		event := events[index]
		if stringPayload(event.Payload, "run_id") != runID {
			continue
		}

		switch event.EventType {
		case "dungeon.entered":
			if enteredEvent == nil {
				enteredEvent = &event
			}
		case "dungeon.cleared":
			if clearedEvent == nil {
				clearedEvent = &event
			}
		case "dungeon.loot_granted":
			if lootEvent == nil {
				lootEvent = &event
			}
		}
	}

	if enteredEvent == nil && clearedEvent == nil && lootEvent == nil {
		return nil, false
	}

	dungeonID := firstNonEmpty(
		stringPayload(eventPayload(enteredEvent), "dungeon_id"),
		stringPayload(eventPayload(clearedEvent), "dungeon_id"),
		stringPayload(eventPayload(lootEvent), "dungeon_id"),
	)
	if dungeonID == "" {
		return nil, false
	}

	definition, _ := dungeonService.GetDungeonDefinition(dungeonID)
	dungeonName := firstNonEmpty(definition.Name, dungeonID)
	currentRating := firstNonEmpty(
		stringPayload(eventPayload(lootEvent), "rating"),
		stringPayload(eventPayload(clearedEvent), "current_rating"),
	)
	rewardClaimed := lootEvent != nil

	return map[string]any{
		"run_id":       runID,
		"dungeon_id":   dungeonID,
		"dungeon_name": dungeonName,
		"difficulty":   "unknown",
		"started_at": firstNonEmpty(
			occurredAtOrEmpty(enteredEvent),
			occurredAtOrEmpty(clearedEvent),
			occurredAtOrEmpty(lootEvent),
		),
		"resolved_at": firstNonEmpty(
			occurredAtOrEmpty(clearedEvent),
			occurredAtOrEmpty(lootEvent),
			occurredAtOrEmpty(enteredEvent),
		),
		"room_summary": map[string]any{
			"source": "event_history",
		},
		"battle_state": map[string]any{
			"engine_mode":  "history_only",
			"final_result": "cleared",
			"source":       "event_history",
		},
		"battle_log": []map[string]any{},
		"milestones": []map[string]any{
			{
				"type":    "history_fallback",
				"message": "runtime battle log unavailable; built from public event history",
			},
			{
				"type":    "rating",
				"rating":  currentRating,
				"message": "rating reconstructed from public event history",
			},
		},
		"result": map[string]any{
			"run_status":       "cleared",
			"runtime_phase":    "history_only",
			"reward_claimable": !rewardClaimed,
			"current_rating":   currentRating,
			"projected_rating": currentRating,
		},
		"reward_summary": map[string]any{
			"pending_rating_rewards": []map[string]any{},
			"staged_material_drops":  mapSlicePayload(eventPayload(lootEvent), "material_drops"),
			"reward_gold":            intPayload(eventPayload(lootEvent), "reward_gold"),
		},
	}, true
}

func publicDungeonRunPayloadFromRun(run dungeons.RunView, dungeonService *dungeons.Service) map[string]any {
	definition, _ := dungeonService.GetDungeonDefinition(run.DungeonID)

	return map[string]any{
		"run_id":         run.RunID,
		"dungeon_id":     run.DungeonID,
		"dungeon_name":   definition.Name,
		"difficulty":     run.Difficulty,
		"potion_loadout": run.PotionLoadout,
		"started_at":     run.StartedAt,
		"resolved_at":    run.ResolvedAt,
		"room_summary":   run.RoomSummary,
		"battle_state":   run.BattleState,
		"battle_log":     run.RecentBattleLog,
		"milestones": []map[string]any{
			{
				"type":    "rating",
				"rating":  run.CurrentRating,
				"message": "rating calculated from auto resolve result",
			},
		},
		"result": map[string]any{
			"run_status":       run.RunStatus,
			"runtime_phase":    run.RuntimePhase,
			"reward_claimable": run.RewardClaimable,
			"current_rating":   run.CurrentRating,
			"projected_rating": run.ProjectedRating,
		},
		"reward_summary": map[string]any{
			"pending_rating_rewards": run.PendingRatingRewards,
			"staged_material_drops":  run.StagedMaterialDrops,
		},
	}
}

func eventPayload(event *world.WorldEvent) map[string]any {
	if event == nil {
		return nil
	}
	return event.Payload
}

func occurredAtOrEmpty(event *world.WorldEvent) string {
	if event == nil {
		return ""
	}
	return event.OccurredAt
}

func mapSlicePayload(payload map[string]any, key string) []map[string]any {
	if payload == nil {
		return []map[string]any{}
	}

	switch value := payload[key].(type) {
	case []map[string]any:
		return value
	case []any:
		items := make([]map[string]any, 0, len(value))
		for _, item := range value {
			entry, ok := item.(map[string]any)
			if ok {
				items = append(items, entry)
			}
		}
		return items
	default:
		return []map[string]any{}
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func stringPayloadFromRun(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func dedupeDungeonHistory(items []map[string]any) []map[string]any {
	seen := make(map[string]struct{}, len(items))
	unique := make([]map[string]any, 0, len(items))
	for _, item := range items {
		runID, _ := item["run_id"].(string)
		if runID == "" {
			continue
		}
		if _, exists := seen[runID]; exists {
			continue
		}
		seen[runID] = struct{}{}
		unique = append(unique, item)
	}
	return unique
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

func runtimeCurioStatus(detail world.RegionDetail, population, recentEventCount int) string {
	switch detail.InteractionLayer {
	case "field", "dungeon":
		switch {
		case population > 0:
			return "active"
		case recentEventCount > 0:
			return "exhausted"
		default:
			return detail.CurioStatus
		}
	case "safe_hub":
		if recentEventCount >= 3 {
			return "active"
		}
	}

	return detail.CurioStatus
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

func appendQuestRuntimeValidActions(actions []characters.ValidAction, account auth.Account, characterService *characters.Service, questService *quests.Service) []characters.ValidAction {
	character, ok := characterService.GetCharacterByAccount(account)
	if !ok {
		return actions
	}
	return append(actions, questService.ListRuntimeActions(character.CharacterID)...)
}

func findBuilding(worldService *world.Service, buildingID string) (world.Building, world.RegionDetail, bool) {
	for _, region := range worldService.ListRegions() {
		detail, ok := worldService.GetRegion(region.ID)
		if !ok {
			continue
		}

		for _, building := range detail.Buildings {
			if building.ID == buildingID {
				return building, detail, true
			}
		}
	}

	return world.Building{}, world.RegionDetail{}, false
}

func handleBuildingAction(w http.ResponseWriter, r *http.Request, authService *auth.Service, worldService *world.Service, characterService *characters.Service, inventoryService *inventory.Service, questService *quests.Service, dungeonService *dungeons.Service, arenaService *arena.Service, actionName string) {
	account, ok := requireAccount(w, r, authService)
	if !ok {
		return
	}

	character, exists := characterService.GetCharacterByAccount(account)
	if !exists {
		writeError(w, r, http.StatusNotFound, "CHARACTER_NOT_FOUND", "create a character before building actions")
		return
	}

	buildingID := chi.URLParam(r, "buildingId")
	building, _, found := findBuilding(worldService, buildingID)
	if !found {
		writeError(w, r, http.StatusNotFound, "BUILDING_NOT_FOUND", "building does not exist")
		return
	}

	if !buildingSupportsAction(building, actionName) {
		writeError(w, r, http.StatusBadRequest, "INVALID_ACTION_STATE", "building does not support this action")
		return
	}

	response := map[string]any{}

	switch actionName {
	case "purchase":
		var payload struct {
			CatalogID string `json:"catalog_id"`
		}
		if !decodeJSONBody(w, r, &payload) {
			return
		}

		catalogID := strings.TrimSpace(payload.CatalogID)
		shopItems := inventoryService.ListShopInventory(building.Type, character)
		selectedIndex := -1
		for index := range shopItems {
			if shopItems[index].CatalogID == catalogID {
				selectedIndex = index
				break
			}
		}
		if selectedIndex < 0 {
			writeError(w, r, http.StatusBadRequest, "INVALID_ACTION_STATE", "shop item does not exist for this building")
			return
		}

		price := shopItems[selectedIndex].PriceGold
		if _, err := characterService.SpendGold(character.CharacterID, price); err != nil {
			if errors.Is(err, characters.ErrGoldInsufficient) {
				writeError(w, r, http.StatusBadRequest, "GOLD_INSUFFICIENT", "character does not have enough gold to purchase this item")
				return
			}
			writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to settle purchase")
			return
		}

		view, purchased, consumable, _, err := inventoryService.PurchaseShopItem(character, catalogID)
		if err != nil {
			_, _ = characterService.GrantGold(character.CharacterID, price)
			switch {
			case errors.Is(err, inventory.ErrCatalogNotFound):
				writeError(w, r, http.StatusBadRequest, "INVALID_ACTION_STATE", "shop item does not exist")
			case errors.Is(err, inventory.ErrItemNotEquippable):
				writeError(w, r, http.StatusBadRequest, "ITEM_NOT_EQUIPPABLE", "item cannot be equipped by current character")
			default:
				writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to purchase item")
			}
			return
		}

		response["inventory"] = view
		if purchased != nil {
			response["item"] = purchased
		}
		if consumable != nil {
			response["consumable"] = consumable
		}
		response["price_gold"] = price
	case "sell":
		var payload struct {
			ItemID string `json:"item_id"`
		}
		if !decodeJSONBody(w, r, &payload) {
			return
		}

		view, sold, gain, err := inventoryService.SellItem(character, payload.ItemID)
		if err != nil {
			switch {
			case errors.Is(err, inventory.ErrItemNotOwned):
				writeError(w, r, http.StatusNotFound, "ITEM_NOT_OWNED", "item does not belong to character")
			case errors.Is(err, inventory.ErrItemEquipped):
				writeError(w, r, http.StatusBadRequest, "INVALID_ACTION_STATE", "unequip item before selling")
			default:
				writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to sell item")
			}
			return
		}

		if _, err := characterService.GrantGold(character.CharacterID, gain); err != nil {
			writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to settle sell reward")
			return
		}

		response["inventory"] = view
		response["item"] = sold
		response["gain_gold"] = gain

		_ = characterService.AppendEvents(character.CharacterID, world.WorldEvent{
			EventID:          requestID(r),
			EventType:        "inventory.item_sold",
			Visibility:       "public",
			ActorCharacterID: character.CharacterID,
			ActorName:        character.Name,
			RegionID:         character.LocationRegionID,
			Summary:          fmt.Sprintf("%s sold %s.", character.Name, sold.Name),
			Payload: map[string]any{
				"item_id":    sold.ItemID,
				"catalog_id": sold.CatalogID,
				"gain_gold":  gain,
			},
			OccurredAt: time.Now().Format(time.RFC3339),
		})
	case "salvage":
		var payload struct {
			ItemID string `json:"item_id"`
		}
		if !decodeJSONBody(w, r, &payload) {
			return
		}

		view, salvaged, drops, err := inventoryService.SalvageItem(character, payload.ItemID)
		if err != nil {
			switch {
			case errors.Is(err, inventory.ErrItemNotOwned):
				writeError(w, r, http.StatusNotFound, "ITEM_NOT_OWNED", "item does not belong to character")
			case errors.Is(err, inventory.ErrItemEquipped):
				writeError(w, r, http.StatusBadRequest, "INVALID_ACTION_STATE", "unequip item before salvaging")
			default:
				writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to salvage item")
			}
			return
		}
		materials, err := characterService.GrantMaterials(character.CharacterID, drops)
		if err != nil {
			writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to grant salvage materials")
			return
		}

		response["inventory"] = view
		response["item"] = salvaged
		response["materials_granted"] = drops
		response["materials"] = materials

		_ = characterService.AppendEvents(character.CharacterID, world.WorldEvent{
			EventID:          requestID(r),
			EventType:        "inventory.item_salvaged",
			Visibility:       "public",
			ActorCharacterID: character.CharacterID,
			ActorName:        character.Name,
			RegionID:         character.LocationRegionID,
			Summary:          fmt.Sprintf("%s salvaged %s.", character.Name, salvaged.Name),
			Payload: map[string]any{
				"item_id":           salvaged.ItemID,
				"catalog_id":        salvaged.CatalogID,
				"materials_granted": drops,
			},
			OccurredAt: time.Now().Format(time.RFC3339),
		})
	case "enhance":
		var payload struct {
			ItemID string `json:"item_id"`
			Slot   string `json:"slot"`
		}
		if !decodeJSONBody(w, r, &payload) {
			return
		}

		item, quote, err := inventoryService.GetEnhancementQuote(character, payload.ItemID, payload.Slot)
		if err != nil {
			switch {
			case errors.Is(err, inventory.ErrItemNotOwned):
				writeError(w, r, http.StatusNotFound, "ITEM_NOT_OWNED", "item does not belong to character")
			case errors.Is(err, inventory.ErrItemNotEnhanceable):
				writeError(w, r, http.StatusBadRequest, "ITEM_NOT_ENHANCEABLE", "item cannot be enhanced")
			case errors.Is(err, inventory.ErrEnhancementCap):
				writeError(w, r, http.StatusBadRequest, "ENHANCEMENT_CAP_REACHED", "item has reached the enhancement cap")
			default:
				writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to quote enhancement")
			}
			return
		}

		if _, err := characterService.SpendGold(character.CharacterID, quote.GoldCost); err != nil {
			if errors.Is(err, characters.ErrGoldInsufficient) {
				writeError(w, r, http.StatusBadRequest, "GOLD_INSUFFICIENT", "character does not have enough gold to enhance this item")
				return
			}
			writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to settle enhancement gold cost")
			return
		}
		materials, err := characterService.SpendMaterials(character.CharacterID, quote.MaterialCost)
		if err != nil {
			_, _ = characterService.GrantGold(character.CharacterID, quote.GoldCost)
			if errors.Is(err, characters.ErrMaterialsInsufficient) {
				writeError(w, r, http.StatusBadRequest, "MATERIALS_INSUFFICIENT", "character does not have enough materials to enhance this item")
				return
			}
			writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to settle enhancement material cost")
			return
		}

		view, enhanced, err := inventoryService.EnhanceItem(character, payload.ItemID, payload.Slot)
		if err != nil {
			_, _ = characterService.GrantGold(character.CharacterID, quote.GoldCost)
			_, _ = characterService.GrantMaterials(character.CharacterID, quote.MaterialCost)
			switch {
			case errors.Is(err, inventory.ErrItemNotOwned):
				writeError(w, r, http.StatusNotFound, "ITEM_NOT_OWNED", "item does not belong to character")
			case errors.Is(err, inventory.ErrItemNotEnhanceable):
				writeError(w, r, http.StatusBadRequest, "ITEM_NOT_ENHANCEABLE", "item cannot be enhanced")
			case errors.Is(err, inventory.ErrEnhancementCap):
				writeError(w, r, http.StatusBadRequest, "ENHANCEMENT_CAP_REACHED", "item has reached the enhancement cap")
			default:
				writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to enhance item")
			}
			return
		}

		response["inventory"] = view
		response["item_before"] = item
		response["item"] = enhanced
		response["enhancement_quote"] = quote
		response["materials"] = materials

		_ = characterService.AppendEvents(character.CharacterID, world.WorldEvent{
			EventID:          requestID(r),
			EventType:        "inventory.item_enhanced",
			Visibility:       "public",
			ActorCharacterID: character.CharacterID,
			ActorName:        character.Name,
			RegionID:         character.LocationRegionID,
			Summary:          fmt.Sprintf("%s enhanced %s to +%d.", character.Name, enhanced.Name, enhanced.EnhancementLevel),
			Payload: map[string]any{
				"item_id":         enhanced.ItemID,
				"catalog_id":      enhanced.CatalogID,
				"from_level":      quote.CurrentLevel,
				"to_level":        enhanced.EnhancementLevel,
				"gold_cost":       quote.GoldCost,
				"material_cost":   quote.MaterialCost,
				"preview_bonus":   quote.PreviewBonusPct,
				"target_stat_set": quote.EnhancementTarget,
				"target_slot":     quote.TargetSlot,
			},
			OccurredAt: time.Now().Format(time.RFC3339),
		})
	}

	state, err := buildCharacterState(account, characterService, inventoryService, questService, dungeonService, arenaService, worldService)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load updated character state")
		return
	}

	writeEnvelope(w, r, http.StatusOK, map[string]any{
		"action_result": map[string]any{
			"action_type":   actionName,
			"building_id":   buildingID,
			"building_name": building.Name,
			"status":        "success",
		},
		"state":  state,
		"result": response,
	})
}

func buildingSupportsAction(building world.Building, actionName string) bool {
	mapping := map[string][]string{
		"heal":     {"restore_hp"},
		"cleanse":  {"remove_status"},
		"enhance":  {"enhance_item"},
		"salvage":  {"salvage_item"},
		"repair":   {"repair_item"},
		"purchase": {"purchase", "buy_consumables"},
		"sell":     {"sell_loot"},
	}

	required, ok := mapping[actionName]
	if !ok {
		return false
	}

	for _, candidate := range required {
		for _, supported := range building.Actions {
			if candidate == supported {
				return true
			}
		}
	}

	return false
}

func buildPublicBots(charactersList []characters.Summary, characterService *characters.Service, inventoryService *inventory.Service, dungeonService *dungeons.Service) []map[string]any {
	items := make([]map[string]any, 0, len(charactersList))
	for _, character := range charactersList {
		combatPowerSummary := map[string]any{}
		if summary, stats, _, _, ok := characterService.GetRuntimeDetailByCharacterID(character.CharacterID); ok {
			inventoryView := inventoryService.GetInventory(summary)
			combatPower, _ := buildCombatPower(summary, stats, inventoryView, dungeonService)
			combatPowerSummary = map[string]any{
				"panel_power_score": combatPower.PanelPowerScore,
				"power_tier":        combatPower.PowerTier,
			}
		}

		items = append(items, map[string]any{
			"character_summary":        character,
			"equipment_score":          inventoryService.ComputeEquipmentScore(character),
			"combat_power":             combatPowerSummary,
			"current_activity_type":    activityTypeFromRegion(character.LocationRegionID),
			"current_activity_summary": fmt.Sprintf("Active in %s", character.LocationRegionID),
			"last_seen_at":             time.Now().Format(time.RFC3339),
		})
	}

	slices.SortFunc(items, func(left, right map[string]any) int {
		leftName, _ := left["character_summary"].(characters.Summary)
		rightName, _ := right["character_summary"].(characters.Summary)
		return cmp.Compare(leftName.Name, rightName.Name)
	})

	return items
}

func activityTypeFromRegion(regionID string) string {
	if strings.Contains(regionID, "dungeon") || strings.Contains(regionID, "catacomb") || strings.Contains(regionID, "den") {
		return "dungeon"
	}
	if strings.Contains(regionID, "city") || strings.Contains(regionID, "village") {
		return "hub"
	}
	return "field"
}

func resolveFieldEncounter(account auth.Account, approach string, characterService *characters.Service, inventoryService *inventory.Service, questService *quests.Service, dungeonService *dungeons.Service, arenaService *arena.Service, worldService *world.Service) (map[string]any, error) {
	preState, err := buildCharacterState(account, characterService, inventoryService, questService, dungeonService, arenaService, worldService)
	if err != nil {
		return nil, err
	}
	character := preState.Character
	limits := preState.Limits

	player := dungeons.BuildPlayerCombatant(preState.Character, preState.Stats, preState.Skills)
	player.PotionBag = combat.DefaultPotionBag()

	result, err := worldService.ResolveFieldEncounter(character.LocationRegionID, approach, player)
	if err != nil {
		return nil, err
	}

	inventoryService.ConsumeConsumables(character.CharacterID, dungeons.PotionUsageFromLog(result.BattleLog))

	event := world.WorldEvent{
		EventID:          fmt.Sprintf("evt_field_%d", time.Now().UnixNano()),
		EventType:        result.EventType,
		Visibility:       "public",
		ActorCharacterID: character.CharacterID,
		ActorName:        character.Name,
		RegionID:         character.LocationRegionID,
		Summary:          fmt.Sprintf("%s %s.", character.Name, result.Summary),
		Payload: map[string]any{
			"battle_type":         result.BattleType,
			"encounter_id":        result.EncounterID,
			"approach":            result.Approach,
			"encounter_family":    result.EncounterFamily,
			"victory":             result.Victory,
			"reward_gold":         result.RewardGold,
			"enemies_defeated":    result.EnemiesDefeated,
			"materials_collected": result.MaterialsCollected,
			"material_drops":      result.MaterialDrops,
			"is_curio":            result.IsCurio,
			"curio_label":         result.CurioLabel,
			"curio_outcome":       result.CurioOutcome,
			"followup_quest":      result.FollowupQuest,
			"battle_state":        result.BattleState,
		},
		OccurredAt: time.Now().Format(time.RFC3339),
	}

	if _, _, err := characterService.ApplyFieldEncounter(character.CharacterID, result.RewardGold, result.MaterialDrops, result.Victory, event); err != nil {
		return nil, err
	}

	board := quests.BoardView{}
	completedQuests := []characters.QuestSummary{}
	if result.Victory {
		board, completedQuests = questService.ProgressFieldQuests(character, character.LocationRegionID, result.EnemiesDefeated, result.MaterialsCollected, limits)
	} else {
		board = questService.ListQuests(character, limits)
	}
	var triggeredQuest *characters.QuestSummary
	if result.Victory && result.FollowupQuest != nil {
		var err error
		board, triggeredQuest, err = questService.EnsureCurioFollowupQuest(character, *result.FollowupQuest, limits)
		if err != nil {
			return nil, err
		}
		if triggeredQuest != nil {
			_ = characterService.AppendEvents(character.CharacterID, world.WorldEvent{
				EventID:          fmt.Sprintf("evt_quest_curio_%s", triggeredQuest.QuestID),
				EventType:        "quest.accepted",
				Visibility:       "public",
				ActorCharacterID: character.CharacterID,
				ActorName:        character.Name,
				RegionID:         character.LocationRegionID,
				Summary:          fmt.Sprintf("%s triggered %s.", character.Name, triggeredQuest.Title),
				Payload: map[string]any{
					"quest_id":          triggeredQuest.QuestID,
					"quest_title":       triggeredQuest.Title,
					"quest_target":      triggeredQuest.TargetRegionID,
					"source":            "curio",
					"curio_label":       result.CurioLabel,
					"curio_outcome":     result.CurioOutcome,
					"curio_region_id":   result.RegionID,
					"followup_required": true,
				},
				OccurredAt: time.Now().Format(time.RFC3339),
			})
		}
	}
	for _, quest := range completedQuests {
		_ = characterService.AppendEvents(character.CharacterID, world.WorldEvent{
			EventID:          fmt.Sprintf("evt_quest_completed_field_%s", quest.QuestID),
			EventType:        "quest.completed",
			Visibility:       "public",
			ActorCharacterID: character.CharacterID,
			ActorName:        character.Name,
			RegionID:         character.LocationRegionID,
			Summary:          fmt.Sprintf("%s completed %s.", character.Name, quest.Title),
			Payload: map[string]any{
				"quest_id":    quest.QuestID,
				"quest_title": quest.Title,
				"approach":    result.Approach,
			},
			OccurredAt: time.Now().Format(time.RFC3339),
		})
	}

	state, err := buildCharacterState(account, characterService, inventoryService, questService, dungeonService, arenaService, worldService)
	if err != nil {
		return nil, err
	}
	state.Objectives = activeObjectivesFromBoard(board)

	return map[string]any{
		"action_result": map[string]any{
			"action_type":  "resolve_field_encounter",
			"region_id":    result.RegionID,
			"approach":     result.Approach,
			"event_type":   result.EventType,
			"encounter_id": result.EncounterID,
		},
		"state": state,
		"result": map[string]any{
			"battle_type":         result.BattleType,
			"encounter_id":        result.EncounterID,
			"encounter_family":    result.EncounterFamily,
			"victory":             result.Victory,
			"reward_gold":         result.RewardGold,
			"enemies_defeated":    result.EnemiesDefeated,
			"materials_collected": result.MaterialsCollected,
			"material_drops":      result.MaterialDrops,
			"is_curio":            result.IsCurio,
			"curio_label":         result.CurioLabel,
			"curio_outcome":       result.CurioOutcome,
			"followup_quest":      triggeredQuest,
			"battle_state":        result.BattleState,
			"battle_log":          result.BattleLog,
		},
	}, nil
}

func executeAction(account auth.Account, actionType string, actionArgs map[string]any, characterService *characters.Service, questService *quests.Service, dungeonService *dungeons.Service, inventoryService *inventory.Service, arenaService *arena.Service, worldService *world.Service) (map[string]any, error) {
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

		state, err := buildCharacterState(account, characterService, inventoryService, questService, dungeonService, arenaService, worldService)
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
	case "resolve_field_encounter":
		approach, _ := actionArgs["approach"].(string)
		return resolveFieldEncounter(account, approach, characterService, inventoryService, questService, dungeonService, arenaService, worldService)
	case "resolve_field_encounter:hunt":
		return resolveFieldEncounter(account, "hunt", characterService, inventoryService, questService, dungeonService, arenaService, worldService)
	case "resolve_field_encounter:gather":
		return resolveFieldEncounter(account, "gather", characterService, inventoryService, questService, dungeonService, arenaService, worldService)
	case "resolve_field_encounter:curio":
		return resolveFieldEncounter(account, "curio", characterService, inventoryService, questService, dungeonService, arenaService, worldService)
	case "submit_quest":
		questID, _ := actionArgs["quest_id"].(string)
		character, limits, err := currentCharacterWithLimits(account, characterService, worldService)
		if err != nil {
			return nil, err
		}

		if limits.QuestCompletionUsed >= limits.QuestCompletionCap {
			return nil, characters.ErrQuestCompletionCap
		}

		quest, err := questService.PrepareQuestSubmission(character, questID, limits)
		if err != nil {
			return nil, err
		}
		if _, _, _, _, err := characterService.ApplyQuestSubmission(character.CharacterID, quest); err != nil {
			return nil, err
		}
		quest, err = questService.FinalizeQuestSubmission(character.CharacterID, quest.QuestID)
		if err != nil {
			return nil, err
		}
		state, err := buildCharacterState(account, characterService, inventoryService, questService, dungeonService, arenaService, worldService)
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
	case "quest_choice":
		questID, _ := actionArgs["quest_id"].(string)
		choiceKey, _ := actionArgs["choice_key"].(string)
		character, limits, err := currentCharacterWithLimits(account, characterService, worldService)
		if err != nil {
			return nil, err
		}
		_ = questService.ListQuests(character, limits)
		quest, runtime, err := questService.ApplyQuestChoice(character, questID, choiceKey)
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"action_result": map[string]any{
				"action_type": "quest_choice",
				"quest_id":    quest.QuestID,
				"choice_key":  choiceKey,
			},
			"state": map[string]any{
				"quest":   quest,
				"runtime": runtime,
				"limits":  limits,
			},
		}, nil
	case "quest_interact":
		questID, _ := actionArgs["quest_id"].(string)
		interaction, _ := actionArgs["interaction"].(string)
		character, limits, err := currentCharacterWithLimits(account, characterService, worldService)
		if err != nil {
			return nil, err
		}
		_ = questService.ListQuests(character, limits)
		quest, runtime, err := questService.AdvanceQuestInteraction(character, questID, interaction)
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"action_result": map[string]any{
				"action_type": "quest_interact",
				"quest_id":    quest.QuestID,
				"interaction": interaction,
			},
			"state": map[string]any{
				"quest":   quest,
				"runtime": runtime,
				"limits":  limits,
			},
		}, nil
	case "exchange_dungeon_reward_claims":
		quantity := intPayload(actionArgs, "quantity")
		character, exists := characterService.GetCharacterByAccount(account)
		if !exists {
			return nil, characters.ErrCharacterNotFound
		}
		summary, limits, err := characterService.PurchaseDungeonRewardClaims(character.CharacterID, quantity)
		if err != nil {
			return nil, err
		}

		return map[string]any{
			"action_result": map[string]any{
				"action_type":                  "exchange_dungeon_reward_claims",
				"quantity":                     max(1, quantity),
				"reputation_cost_per_claim":    characters.BonusDungeonClaimCostRep,
				"bonus_claims_purchased_today": limits.BonusDungeonEntryPurchased,
			},
			"state": map[string]any{
				"character": summary,
				"limits":    limits,
			},
		}, nil
	case "enter_dungeon":
		dungeonID, _ := actionArgs["dungeon_id"].(string)
		difficulty, _ := actionArgs["difficulty"].(string)
		potionLoadout := make([]string, 0, 2)
		switch raw := actionArgs["potion_loadout"].(type) {
		case []any:
			for _, item := range raw {
				if value, ok := item.(string); ok {
					potionLoadout = append(potionLoadout, value)
				}
			}
		case []string:
			potionLoadout = append(potionLoadout, raw...)
		}
		character, limits, err := currentCharacterWithLimits(account, characterService, worldService)
		if err != nil {
			return nil, err
		}

		potionBag, err := inventoryService.BuildPotionLoadout(character, potionLoadout)
		if err != nil {
			return nil, dungeons.ErrDungeonPotionLoadoutInvalid
		}
		state, err := buildCharacterState(account, characterService, inventoryService, questService, dungeonService, arenaService, worldService)
		if err != nil {
			return nil, err
		}
		player := dungeons.BuildPlayerCombatant(state.Character, state.Stats, state.Skills)
		run, err := dungeonService.EnterDungeon(character, limits, player, dungeonID, difficulty, potionLoadout, potionBag)
		if err != nil {
			return nil, err
		}
		inventoryService.ConsumeConsumables(character.CharacterID, dungeons.PotionUsageFromLog(run.RecentBattleLog))

		definition, _ := dungeonService.GetDungeonDefinition(run.DungeonID)
		completedQuests := []characters.QuestSummary{}
		if run.RunStatus == "cleared" {
			_, completedQuests = questService.ProgressDungeonQuests(character, definition.RegionID, limits)
		}
		events := []world.WorldEvent{{
			EventID:          fmt.Sprintf("evt_dungeon_enter_%s", run.RunID),
			EventType:        "dungeon.entered",
			Visibility:       "public",
			ActorCharacterID: character.CharacterID,
			ActorName:        character.Name,
			RegionID:         definition.RegionID,
			Summary:          fmt.Sprintf("%s entered %s.", character.Name, definition.Name),
			Payload: map[string]any{
				"run_id":     run.RunID,
				"dungeon_id": run.DungeonID,
			},
			OccurredAt: time.Now().Format(time.RFC3339),
		}}
		if run.RunStatus == "cleared" {
			events = append(events, world.WorldEvent{
				EventID:          fmt.Sprintf("evt_dungeon_clear_%s", run.RunID),
				EventType:        "dungeon.cleared",
				Visibility:       "public",
				ActorCharacterID: character.CharacterID,
				ActorName:        character.Name,
				RegionID:         definition.RegionID,
				Summary:          fmt.Sprintf("%s cleared %s.", character.Name, definition.Name),
				Payload: map[string]any{
					"run_id":         run.RunID,
					"dungeon_id":     run.DungeonID,
					"current_rating": run.CurrentRating,
				},
				OccurredAt: time.Now().Format(time.RFC3339),
			})
		} else {
			events = append(events, world.WorldEvent{
				EventID:          fmt.Sprintf("evt_dungeon_failed_%s", run.RunID),
				EventType:        "dungeon.failed",
				Visibility:       "public",
				ActorCharacterID: character.CharacterID,
				ActorName:        character.Name,
				RegionID:         definition.RegionID,
				Summary:          fmt.Sprintf("%s failed in %s.", character.Name, definition.Name),
				Payload: map[string]any{
					"run_id":               run.RunID,
					"dungeon_id":           run.DungeonID,
					"highest_room_cleared": run.HighestRoomCleared,
				},
				OccurredAt: time.Now().Format(time.RFC3339),
			})
		}
		_ = characterService.AppendEvents(character.CharacterID, events...)
		for _, quest := range completedQuests {
			_ = characterService.AppendEvents(character.CharacterID, world.WorldEvent{
				EventID:          fmt.Sprintf("evt_quest_completed_dungeon_%s", quest.QuestID),
				EventType:        "quest.completed",
				Visibility:       "public",
				ActorCharacterID: character.CharacterID,
				ActorName:        character.Name,
				RegionID:         definition.RegionID,
				Summary:          fmt.Sprintf("%s completed %s.", character.Name, quest.Title),
				Payload: map[string]any{
					"quest_id":    quest.QuestID,
					"quest_title": quest.Title,
					"dungeon_id":  run.DungeonID,
					"run_id":      run.RunID,
				},
				OccurredAt: time.Now().Format(time.RFC3339),
			})
		}

		updatedState, err := buildCharacterState(account, characterService, inventoryService, questService, dungeonService, arenaService, worldService)
		if err != nil {
			return nil, err
		}

		return map[string]any{
			"action_result": map[string]any{
				"action_type": "enter_dungeon",
				"run_id":      run.RunID,
				"run_status":  run.RunStatus,
			},
			"state": map[string]any{
				"run":   run,
				"state": updatedState,
			},
		}, nil
	case "claim_dungeon_rewards":
		runID, _ := actionArgs["run_id"].(string)
		character, limits, err := currentCharacterWithLimits(account, characterService, worldService)
		if err != nil {
			return nil, err
		}

		run, claimPackage, err := dungeonService.PreviewRunRewards(character.CharacterID, runID, limits)
		if err != nil {
			return nil, err
		}

		rewardItemCatalogIDs, err := grantDungeonInventoryRewards(inventoryService, character, claimPackage)
		if err != nil {
			return nil, err
		}

		if _, _, _, err := characterService.ApplyDungeonRewardClaim(character.CharacterID, run.RunID, run.DungeonID, claimPackage.RewardGold, claimPackage.Rating, rewardItemCatalogIDs, claimPackage.MaterialDrops); err != nil {
			return nil, err
		}
		run, err = dungeonService.FinalizeRunRewards(character.CharacterID, run.RunID)
		if err != nil {
			return nil, err
		}

		state, err := buildCharacterState(account, characterService, inventoryService, questService, dungeonService, arenaService, worldService)
		if err != nil {
			return nil, err
		}

		return map[string]any{
			"action_result": map[string]any{
				"action_type":  "claim_dungeon_rewards",
				"run_id":       run.RunID,
				"reward_gold":  claimPackage.RewardGold,
				"reward_items": rewardItemCatalogIDs,
			},
			"state": map[string]any{
				"run":   run,
				"state": state,
			},
		}, nil
	case "equip_item":
		itemID, _ := actionArgs["item_id"].(string)
		character, exists := characterService.GetCharacterByAccount(account)
		if !exists {
			return nil, characters.ErrCharacterNotFound
		}

		view, err := inventoryService.EquipItem(character, itemID)
		if err != nil {
			return nil, err
		}

		return map[string]any{
			"action_result": map[string]any{"action_type": "equip_item", "item_id": itemID},
			"state":         view,
		}, nil
	case "enter_building":
		buildingID, _ := actionArgs["building_id"].(string)
		building, detail, found := findBuilding(worldService, buildingID)
		if !found {
			return nil, characters.ErrActionNotSupported
		}
		return map[string]any{
			"action_result": map[string]any{
				"action_type":   "enter_building",
				"building_id":   building.ID,
				"building_name": building.Name,
			},
			"state": map[string]any{
				"region":            detail.Region,
				"supported_actions": building.Actions,
			},
		}, nil
	case "restore_hp":
		return map[string]any{
			"action_result": map[string]any{
				"action_type": "restore_hp",
				"status":      "success",
			},
			"state": map[string]any{},
		}, nil
	case "remove_status", "enhance_item", "sell_item":
		return map[string]any{
			"action_result": map[string]any{
				"action_type": actionType,
				"status":      "success",
			},
			"state": map[string]any{},
		}, nil
	case "unequip_item":
		slot, _ := actionArgs["slot"].(string)
		character, exists := characterService.GetCharacterByAccount(account)
		if !exists {
			return nil, characters.ErrCharacterNotFound
		}

		view, err := inventoryService.UnequipItem(character, slot)
		if err != nil {
			return nil, err
		}

		return map[string]any{
			"action_result": map[string]any{"action_type": "unequip_item", "slot": slot},
			"state":         view,
		}, nil
	case "arena_signup":
		character, exists := characterService.GetCharacterByAccount(account)
		if !exists {
			return nil, characters.ErrCharacterNotFound
		}

		state, err := buildCharacterState(account, characterService, inventoryService, questService, dungeonService, arenaService, worldService)
		if err != nil {
			return nil, err
		}

		entry, err := arenaService.Signup(character, state.CombatPower.PanelPowerScore, inventoryService.ComputeEquipmentScore(character), worldService.CurrentArenaStatus(worldService.CurrentTime()))
		if err != nil {
			return nil, err
		}

		return map[string]any{
			"action_result": map[string]any{"action_type": "arena_signup", "signed_up": true},
			"state":         map[string]any{"entry": entry},
		}, nil
	default:
		return characterService.ExecuteAction(account, actionType, actionArgs, worldService)
	}
}

func grantDungeonInventoryRewards(inventoryService *inventory.Service, character characters.Summary, claimPackage dungeons.ClaimRewardPackage) ([]string, error) {
	if len(claimPackage.RatingRewards) == 0 {
		return []string{}, nil
	}

	grantedCatalogIDs := make([]string, 0, len(claimPackage.RatingRewards))
	for _, reward := range claimPackage.RatingRewards {
		catalogID, _ := reward["catalog_id"].(string)
		catalogID = strings.TrimSpace(catalogID)
		if catalogID == "" {
			continue
		}

		if _, _, err := inventoryService.GrantItemFromCatalog(character, catalogID); err != nil {
			return nil, err
		}
		grantedCatalogIDs = append(grantedCatalogIDs, catalogID)
	}

	return grantedCatalogIDs, nil
}
