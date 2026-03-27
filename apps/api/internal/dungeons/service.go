package dungeons

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"clawgame/apps/api/internal/characters"
)

var (
	ErrDungeonNotFound              = errors.New("dungeon not found")
	ErrDungeonRankNotEligible       = errors.New("dungeon rank not eligible")
	ErrDungeonRunAlreadyActive      = errors.New("dungeon run already active")
	ErrDungeonRunNotFound           = errors.New("dungeon run not found")
	ErrDungeonRunForbidden          = errors.New("dungeon run forbidden")
	ErrDungeonRewardClaimNotAllowed = errors.New("dungeon reward not claimable")
	ErrDungeonRewardClaimCapReached = errors.New("dungeon reward claim cap reached")
)

type DefinitionView struct {
	DungeonID           string              `json:"dungeon_id"`
	Name                string              `json:"name"`
	RegionID            string              `json:"region_id"`
	MinRank             string              `json:"min_rank"`
	RoomCount           int                 `json:"room_count"`
	BossRoomIndex       int                 `json:"boss_room_index"`
	RecommendedLevelMin int                 `json:"recommended_level_min"`
	RecommendedLevelMax int                 `json:"recommended_level_max"`
	RatingRules         []DungeonRatingRule `json:"rating_rules"`
	RewardSummary       map[string]any      `json:"reward_summary"`
}

type DungeonRatingRule struct {
	HighestRoomCleared int    `json:"highest_room_cleared"`
	Rating             string `json:"rating"`
}

type RunView struct {
	RunID                     string                   `json:"run_id"`
	DungeonID                 string                   `json:"dungeon_id"`
	RunStatus                 string                   `json:"run_status"`
	RuntimePhase              string                   `json:"runtime_phase"`
	RewardClaimable           bool                     `json:"reward_claimable"`
	RewardClaimedAt           *string                  `json:"reward_claimed_at"`
	ClaimConsumesDailyCounter bool                     `json:"claim_consumes_daily_counter"`
	CurrentRoomIndex          int                      `json:"current_room_index"`
	HighestRoomCleared        int                      `json:"highest_room_cleared"`
	ProjectedRating           *string                  `json:"projected_rating"`
	CurrentRating             *string                  `json:"current_rating"`
	RoomSummary               map[string]any           `json:"room_summary"`
	BattleState               map[string]any           `json:"battle_state"`
	StagedMaterialDrops       []map[string]any         `json:"staged_material_drops"`
	PendingRatingRewards      []map[string]any         `json:"pending_rating_rewards"`
	AvailableActions          []characters.ValidAction `json:"available_actions"`
	RecentBattleLog           []map[string]any         `json:"recent_battle_log"`
}

type runRecord struct {
	CharacterID string
	Run         RunView
	RewardGold  int
}

type Service struct {
	mu                   sync.RWMutex
	clock                func() time.Time
	runsByID             map[string]runRecord
	activeRunByCharacter map[string]string
}

var runCounter uint64

func NewService() *Service {
	return &Service{
		clock:                time.Now,
		runsByID:             make(map[string]runRecord),
		activeRunByCharacter: make(map[string]string),
	}
}

func (s *Service) GetDungeonDefinition(dungeonID string) (DefinitionView, bool) {
	def, ok := dungeonDefinitions[dungeonID]
	return def, ok
}

func (s *Service) EnterDungeon(character characters.Summary, limits characters.DailyLimits, dungeonID string) (RunView, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	def, ok := dungeonDefinitions[dungeonID]
	if !ok {
		return RunView{}, ErrDungeonNotFound
	}
	if !rankAllows(character.Rank, def.MinRank) {
		return RunView{}, ErrDungeonRankNotEligible
	}
	if limits.DungeonEntryUsed >= limits.DungeonEntryCap {
		return RunView{}, ErrDungeonRewardClaimCapReached
	}
	if activeRunID, ok := s.activeRunByCharacter[character.CharacterID]; ok {
		if existing, exists := s.runsByID[activeRunID]; exists {
			if existing.Run.RunStatus == "active" || existing.Run.RunStatus == "resolving" {
				return RunView{}, ErrDungeonRunAlreadyActive
			}
		}
		delete(s.activeRunByCharacter, character.CharacterID)
	}

	now := s.clock().Format(time.RFC3339)
	runID := nextID("run")
	rating := ratingFromRoomClear(def.RoomCount, def.RoomCount)

	run := RunView{
		RunID:                     runID,
		DungeonID:                 def.DungeonID,
		RunStatus:                 "cleared",
		RuntimePhase:              "result_ready",
		RewardClaimable:           true,
		RewardClaimedAt:           nil,
		ClaimConsumesDailyCounter: true,
		CurrentRoomIndex:          def.RoomCount,
		HighestRoomCleared:        def.RoomCount,
		ProjectedRating:           &rating,
		CurrentRating:             &rating,
		RoomSummary: map[string]any{
			"room_count":      def.RoomCount,
			"boss_room_index": def.BossRoomIndex,
		},
		BattleState: map[string]any{
			"engine_mode": "auto",
			"resolved_at": now,
		},
		StagedMaterialDrops: []map[string]any{
			{
				"material_key": "dungeon_essence",
				"quantity":     3,
			},
		},
		PendingRatingRewards: []map[string]any{
			{
				"reward_type": "rating_chest",
				"rating":      rating,
			},
		},
		AvailableActions: []characters.ValidAction{
			{
				ActionType: "claim_run_rewards",
				Label:      "Claim Dungeon Rewards",
				ArgsSchema: map[string]any{"run_id": "string"},
			},
		},
		RecentBattleLog: []map[string]any{
			{
				"step":    "auto_resolve_start",
				"message": "backend engine started simulation",
			},
			{
				"step":    "auto_resolve_end",
				"message": fmt.Sprintf("dungeon cleared with rating %s", rating),
			},
		},
	}

	s.runsByID[runID] = runRecord{
		CharacterID: character.CharacterID,
		Run:         run,
		RewardGold:  definitionRewardGold(def),
	}
	delete(s.activeRunByCharacter, character.CharacterID)

	return run, nil
}

func (s *Service) GetActiveRun(characterID string) (*RunView, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	runID, ok := s.activeRunByCharacter[characterID]
	if !ok {
		return nil, nil
	}
	record, ok := s.runsByID[runID]
	if !ok {
		return nil, nil
	}
	if record.CharacterID != characterID {
		return nil, ErrDungeonRunForbidden
	}

	copyRun := record.Run
	return &copyRun, nil
}

func (s *Service) GetRun(characterID, runID string) (RunView, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	record, ok := s.runsByID[runID]
	if !ok {
		return RunView{}, ErrDungeonRunNotFound
	}
	if record.CharacterID != characterID {
		return RunView{}, ErrDungeonRunForbidden
	}

	return record.Run, nil
}

func (s *Service) ClaimRunRewards(characterID, runID string, limits characters.DailyLimits) (RunView, int, string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.runsByID[runID]
	if !ok {
		return RunView{}, 0, "", ErrDungeonRunNotFound
	}
	if record.CharacterID != characterID {
		return RunView{}, 0, "", ErrDungeonRunForbidden
	}
	if limits.DungeonEntryUsed >= limits.DungeonEntryCap {
		return RunView{}, 0, "", ErrDungeonRewardClaimCapReached
	}
	if record.Run.RunStatus != "cleared" || !record.Run.RewardClaimable {
		return RunView{}, 0, "", ErrDungeonRewardClaimNotAllowed
	}

	now := s.clock().Format(time.RFC3339)
	record.Run.RewardClaimable = false
	record.Run.RewardClaimedAt = &now
	record.Run.RuntimePhase = "claim_settled"
	record.Run.AvailableActions = []characters.ValidAction{}
	record.Run.ClaimConsumesDailyCounter = true

	s.runsByID[runID] = record

	rating := ""
	if record.Run.CurrentRating != nil {
		rating = *record.Run.CurrentRating
	}

	return record.Run, record.RewardGold, rating, nil
}

func nextID(prefix string) string {
	return fmt.Sprintf("%s_%d_%06d", prefix, time.Now().UnixNano(), atomic.AddUint64(&runCounter, 1))
}

func definitionRewardGold(def DefinitionView) int {
	if value, ok := def.RewardSummary["gold_min"].(int); ok {
		return value
	}
	if value, ok := def.RewardSummary["gold_min"].(float64); ok {
		return int(value)
	}
	return 120
}

func ratingFromRoomClear(highestRoom, totalRooms int) string {
	if totalRooms <= 0 {
		return "E"
	}
	ratio := float64(highestRoom) / float64(totalRooms)
	switch {
	case ratio >= 1.0:
		return "A"
	case ratio >= 0.8:
		return "B"
	case ratio >= 0.6:
		return "C"
	case ratio >= 0.4:
		return "D"
	default:
		return "E"
	}
}

func rankAllows(currentRank, requiredRank string) bool {
	return rankOrder[currentRank] >= rankOrder[requiredRank]
}

var rankOrder = map[string]int{
	"low":  1,
	"mid":  2,
	"high": 3,
}

var dungeonDefinitions = map[string]DefinitionView{
	"ancient_catacomb_v1": {
		DungeonID:           "ancient_catacomb_v1",
		Name:                "Ancient Catacomb",
		RegionID:            "ancient_catacomb",
		MinRank:             "low",
		RoomCount:           6,
		BossRoomIndex:       6,
		RecommendedLevelMin: 1,
		RecommendedLevelMax: 30,
		RatingRules: []DungeonRatingRule{
			{HighestRoomCleared: 2, Rating: "E"},
			{HighestRoomCleared: 3, Rating: "D"},
			{HighestRoomCleared: 4, Rating: "C"},
			{HighestRoomCleared: 5, Rating: "B"},
			{HighestRoomCleared: 6, Rating: "A"},
		},
		RewardSummary: map[string]any{
			"gold_min":   180,
			"gold_max":   260,
			"drop_table": "catacomb_v1",
		},
	},
	"sandworm_den_v1": {
		DungeonID:           "sandworm_den_v1",
		Name:                "Sandworm Den",
		RegionID:            "sandworm_den",
		MinRank:             "high",
		RoomCount:           6,
		BossRoomIndex:       6,
		RecommendedLevelMin: 70,
		RecommendedLevelMax: 100,
		RatingRules: []DungeonRatingRule{
			{HighestRoomCleared: 2, Rating: "E"},
			{HighestRoomCleared: 3, Rating: "D"},
			{HighestRoomCleared: 4, Rating: "C"},
			{HighestRoomCleared: 5, Rating: "B"},
			{HighestRoomCleared: 6, Rating: "A"},
		},
		RewardSummary: map[string]any{
			"gold_min":   320,
			"gold_max":   460,
			"drop_table": "sandworm_v1",
		},
	},
}
