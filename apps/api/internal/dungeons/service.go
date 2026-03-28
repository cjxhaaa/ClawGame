package dungeons

import (
	"errors"
	"fmt"
	"hash/fnv"
	"math"
	"sort"
	"strings"
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
	StartedAt                 string                   `json:"started_at"`
	ResolvedAt                string                   `json:"resolved_at"`
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

type ClaimRewardPackage struct {
	RewardGold    int              `json:"reward_gold"`
	Rating        string           `json:"rating"`
	MaterialDrops []map[string]any `json:"material_drops"`
	RatingRewards []map[string]any `json:"rating_rewards"`
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

func (s *Service) ListDungeonDefinitions() []DefinitionView {
	definitions := make([]DefinitionView, 0, len(dungeonDefinitions))
	for _, definition := range dungeonDefinitions {
		definitions = append(definitions, definition)
	}

	sort.Slice(definitions, func(i, j int) bool {
		if definitions[i].RegionID == definitions[j].RegionID {
			return definitions[i].DungeonID < definitions[j].DungeonID
		}

		return definitions[i].RegionID < definitions[j].RegionID
	})

	return definitions
}

func RankAllows(currentRank, requiredRank string) bool {
	return rankAllows(currentRank, requiredRank)
}

func (s *Service) EnterDungeon(character characters.Summary, _ characters.DailyLimits, dungeonID string) (RunView, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	def, ok := dungeonDefinitions[dungeonID]
	if !ok {
		return RunView{}, ErrDungeonNotFound
	}
	if !rankAllows(character.Rank, def.MinRank) {
		return RunView{}, ErrDungeonRankNotEligible
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

	battleResult := simulateDungeonRun(character, def, runID)
	rating := ratingFromRoomClear(battleResult.HighestRoomCleared, def.RoomCount)
	pendingRatingRewards := []map[string]any{}
	if battleResult.Cleared {
		pendingRatingRewards = buildPendingRatingRewards(runID, def.DungeonID, rating, character)
	}

	runStatus := "failed"
	if battleResult.Cleared {
		runStatus = "cleared"
	}
	rewardClaimable := battleResult.Cleared
	stagedMaterialDrops := []map[string]any{}
	if rewardClaimable {
		stagedMaterialDrops = stagedMaterialDropsFromClear(battleResult.HighestRoomCleared)
	}
	availableActions := []characters.ValidAction{}
	if rewardClaimable {
		availableActions = []characters.ValidAction{
			{
				ActionType: "claim_dungeon_rewards",
				Label:      "Claim Dungeon Rewards",
				ArgsSchema: map[string]any{"run_id": "string"},
			},
		}
	}

	run := RunView{
		RunID:                     runID,
		DungeonID:                 def.DungeonID,
		StartedAt:                 now,
		ResolvedAt:                now,
		RunStatus:                 runStatus,
		RuntimePhase:              "result_ready",
		RewardClaimable:           rewardClaimable,
		RewardClaimedAt:           nil,
		ClaimConsumesDailyCounter: rewardClaimable,
		CurrentRoomIndex:          battleResult.CurrentRoomIndex,
		HighestRoomCleared:        battleResult.HighestRoomCleared,
		ProjectedRating:           &rating,
		CurrentRating:             &rating,
		RoomSummary: map[string]any{
			"room_count":      def.RoomCount,
			"boss_room_index": def.BossRoomIndex,
			"rooms_cleared":   battleResult.HighestRoomCleared,
		},
		BattleState: map[string]any{
			"engine_mode":         "auto_turn_based",
			"resolved_at":         now,
			"start_hp":            battleResult.StartHP,
			"remaining_hp":        battleResult.RemainingHP,
			"rooms_attempted":     battleResult.CurrentRoomIndex,
			"player_survived":     battleResult.PlayerSurvived,
			"final_result":        runStatus,
			"ended_in_room_index": battleResult.EndedInRoomIndex,
		},
		StagedMaterialDrops:  stagedMaterialDrops,
		PendingRatingRewards: pendingRatingRewards,
		AvailableActions:     availableActions,
		RecentBattleLog:      battleResult.Log,
	}

	rewardGold := 0
	if battleResult.Cleared {
		rewardGold = scaledRewardGold(definitionRewardGold(def), battleResult.HighestRoomCleared, def.RoomCount)
	}

	s.runsByID[runID] = runRecord{
		CharacterID: character.CharacterID,
		Run:         run,
		RewardGold:  rewardGold,
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

func (s *Service) ListRunsByCharacter(characterID string) []RunView {
	s.mu.RLock()
	defer s.mu.RUnlock()

	runs := make([]RunView, 0, len(s.runsByID))
	for _, record := range s.runsByID {
		if characterID != "" && record.CharacterID != characterID {
			continue
		}
		runs = append(runs, record.Run)
	}

	sortRunsByResolvedAtDesc(runs)
	return runs
}

func (s *Service) ClaimRunRewards(characterID, runID string, limits characters.DailyLimits) (RunView, ClaimRewardPackage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.runsByID[runID]
	if !ok {
		return RunView{}, ClaimRewardPackage{}, ErrDungeonRunNotFound
	}
	if record.CharacterID != characterID {
		return RunView{}, ClaimRewardPackage{}, ErrDungeonRunForbidden
	}
	if limits.DungeonEntryUsed >= limits.DungeonEntryCap {
		return RunView{}, ClaimRewardPackage{}, ErrDungeonRewardClaimCapReached
	}
	if record.Run.RunStatus != "cleared" || !record.Run.RewardClaimable {
		return RunView{}, ClaimRewardPackage{}, ErrDungeonRewardClaimNotAllowed
	}

	materialDrops := copyMapSlice(record.Run.StagedMaterialDrops)
	ratingRewards := copyMapSlice(record.Run.PendingRatingRewards)

	now := s.clock().Format(time.RFC3339)
	record.Run.RewardClaimable = false
	record.Run.RewardClaimedAt = &now
	record.Run.RuntimePhase = "claim_settled"
	record.Run.AvailableActions = []characters.ValidAction{}
	record.Run.ClaimConsumesDailyCounter = true
	record.Run.StagedMaterialDrops = []map[string]any{}
	record.Run.PendingRatingRewards = []map[string]any{}

	s.runsByID[runID] = record

	rating := ""
	if record.Run.CurrentRating != nil {
		rating = *record.Run.CurrentRating
	}

	return record.Run, ClaimRewardPackage{
		RewardGold:    record.RewardGold,
		Rating:        rating,
		MaterialDrops: materialDrops,
		RatingRewards: ratingRewards,
	}, nil
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

type simulatedRunResult struct {
	Cleared            bool
	PlayerSurvived     bool
	StartHP            int
	RemainingHP        int
	CurrentRoomIndex   int
	HighestRoomCleared int
	EndedInRoomIndex   int
	Log                []map[string]any
}

func simulateDungeonRun(character characters.Summary, def DefinitionView, runID string) simulatedRunResult {
	stats := baselineStatsForClass(character.Class)
	startHP := maxInt(1, stats.maxHP)
	currentHP := startHP
	log := make([]map[string]any, 0, def.RoomCount*8)

	for roomIndex := 1; roomIndex <= def.RoomCount; roomIndex++ {
		enemy := enemyForRoom(def, roomIndex)
		outcome := simulateRoomCombat(runID, roomIndex, currentHP, stats, enemy)
		log = append(log, outcome.Log...)
		if len(log) > 120 {
			log = log[len(log)-120:]
		}

		if !outcome.PlayerWon {
			return simulatedRunResult{
				Cleared:            false,
				PlayerSurvived:     outcome.PlayerHP > 0,
				StartHP:            startHP,
				RemainingHP:        maxInt(0, outcome.PlayerHP),
				CurrentRoomIndex:   roomIndex,
				HighestRoomCleared: roomIndex - 1,
				EndedInRoomIndex:   roomIndex,
				Log:                log,
			}
		}

		currentHP = outcome.PlayerHP
		if roomIndex < def.RoomCount {
			recovery := maxInt(6, int(float64(stats.maxHP)*0.18)+stats.healingPower)
			beforeHP := currentHP
			currentHP = minInt(stats.maxHP, currentHP+recovery)
			log = append(log, map[string]any{
				"step":                  "room_recovery",
				"event_type":            "room_recovery",
				"room_index":            roomIndex,
				"turn":                  0,
				"actor":                 "system",
				"target":                "player",
				"action":                "short_rest",
				"value":                 recovery,
				"value_type":            "heal",
				"target_hp_before":      beforeHP,
				"target_hp_after":       currentHP,
				"player_hp":             currentHP,
				"enemy_hp":              0,
				"cooldown_before_round": map[string]int{},
				"cooldown_after_round":  map[string]int{},
				"message":               "player recovers between rooms",
			})
		}
	}

	return simulatedRunResult{
		Cleared:            true,
		PlayerSurvived:     true,
		StartHP:            startHP,
		RemainingHP:        currentHP,
		CurrentRoomIndex:   def.RoomCount,
		HighestRoomCleared: def.RoomCount,
		EndedInRoomIndex:   def.RoomCount,
		Log:                log,
	}
}

type combatStats struct {
	maxHP           int
	physicalAttack  int
	magicAttack     int
	physicalDefense int
	magicDefense    int
	speed           int
	healingPower    int
}

type roomEnemy struct {
	name            string
	hp              int
	attack          int
	physicalDefense int
	magicDefense    int
	speed           int
}

type roomCombatResult struct {
	PlayerWon bool
	PlayerHP  int
	Log       []map[string]any
}

func baselineStatsForClass(class string) combatStats {
	class = strings.TrimSpace(strings.ToLower(class))
	switch class {
	case "warrior":
		return combatStats{maxHP: 132, physicalAttack: 28, magicAttack: 6, physicalDefense: 18, magicDefense: 10, speed: 10, healingPower: 4}
	case "mage":
		return combatStats{maxHP: 92, physicalAttack: 11, magicAttack: 34, physicalDefense: 9, magicDefense: 18, speed: 16, healingPower: 8}
	case "priest":
		return combatStats{maxHP: 104, physicalAttack: 10, magicAttack: 26, physicalDefense: 11, magicDefense: 17, speed: 14, healingPower: 20}
	default:
		return combatStats{maxHP: 100, physicalAttack: 16, magicAttack: 16, physicalDefense: 12, magicDefense: 12, speed: 12, healingPower: 6}
	}
}

func enemyForRoom(def DefinitionView, roomIndex int) roomEnemy {
	isBoss := roomIndex == def.BossRoomIndex
	baseHP := 40 + roomIndex*12
	baseAtk := 7 + roomIndex*2
	basePDef := 3 + roomIndex/2
	baseMDef := 3 + roomIndex/2
	baseSpeed := 7 + roomIndex/2

	if strings.Contains(def.DungeonID, "sandworm") {
		baseHP += 45
		baseAtk += 6
		basePDef += 4
		baseMDef += 4
		baseSpeed += 2
	}

	if isBoss {
		baseHP = int(float64(baseHP) * 1.5)
		baseAtk = int(float64(baseAtk) * 1.25)
		basePDef += 3
		baseMDef += 3
		baseSpeed += 3
	}

	name := fmt.Sprintf("room_%d_enemy", roomIndex)
	if isBoss {
		name = "boss"
	}

	return roomEnemy{
		name:            name,
		hp:              baseHP,
		attack:          baseAtk,
		physicalDefense: basePDef,
		magicDefense:    baseMDef,
		speed:           baseSpeed,
	}
}

func simulateRoomCombat(runID string, roomIndex int, startingPlayerHP int, player combatStats, enemy roomEnemy) roomCombatResult {
	playerHP := maxInt(1, startingPlayerHP)
	enemyHP := enemy.hp
	turn := 1
	maxTurns := 24

	basicCooldown := 0
	burstCooldown := 0
	healCooldown := 0
	enemyBurstCooldown := 0

	log := []map[string]any{
		{
			"step":                  "room_start",
			"event_type":            "room_start",
			"room_index":            roomIndex,
			"turn":                  0,
			"player_hp":             playerHP,
			"enemy_hp":              enemyHP,
			"enemy_name":            enemy.name,
			"player_speed":          player.speed,
			"enemy_speed":           enemy.speed,
			"cooldown_before_round": map[string]int{"player_burst": burstCooldown, "player_heal": healCooldown, "enemy_burst": enemyBurstCooldown},
			"cooldown_after_round":  map[string]int{"player_burst": burstCooldown, "player_heal": healCooldown, "enemy_burst": enemyBurstCooldown},
			"message":               "auto battle room started",
		},
	}

	for turn <= maxTurns && playerHP > 0 && enemyHP > 0 {
		playerActsFirst := player.speed >= enemy.speed
		initiative := []string{"enemy", "player"}
		if playerActsFirst {
			initiative = []string{"player", "enemy"}
		}
		log = append(log, map[string]any{
			"step":                  "turn_start",
			"event_type":            "turn_start",
			"room_index":            roomIndex,
			"turn":                  turn,
			"actor":                 "system",
			"initiative_order":      initiative,
			"player_hp":             playerHP,
			"enemy_hp":              enemyHP,
			"cooldown_before_round": map[string]int{"player_burst": burstCooldown, "player_heal": healCooldown, "enemy_burst": enemyBurstCooldown},
			"message":               "turn begins",
		})
		if playerActsFirst {
			playerHP, enemyHP, basicCooldown, burstCooldown, healCooldown, log = playerTurn(turn, roomIndex, player, enemy, playerHP, enemyHP, basicCooldown, burstCooldown, healCooldown, log)
			if enemyHP <= 0 || playerHP <= 0 {
				break
			}
			playerHP, enemyHP, enemyBurstCooldown, log = enemyTurn(turn, roomIndex, player, enemy, playerHP, enemyHP, enemyBurstCooldown, log)
		} else {
			playerHP, enemyHP, enemyBurstCooldown, log = enemyTurn(turn, roomIndex, player, enemy, playerHP, enemyHP, enemyBurstCooldown, log)
			if enemyHP <= 0 || playerHP <= 0 {
				break
			}
			playerHP, enemyHP, basicCooldown, burstCooldown, healCooldown, log = playerTurn(turn, roomIndex, player, enemy, playerHP, enemyHP, basicCooldown, burstCooldown, healCooldown, log)
		}

		basicCooldown = maxInt(0, basicCooldown-1)
		burstCooldown = maxInt(0, burstCooldown-1)
		healCooldown = maxInt(0, healCooldown-1)
		enemyBurstCooldown = maxInt(0, enemyBurstCooldown-1)
		log = append(log, map[string]any{
			"step":                 "turn_end",
			"event_type":           "turn_end",
			"room_index":           roomIndex,
			"turn":                 turn,
			"actor":                "system",
			"player_hp":            playerHP,
			"enemy_hp":             enemyHP,
			"cooldown_after_round": map[string]int{"player_burst": burstCooldown, "player_heal": healCooldown, "enemy_burst": enemyBurstCooldown},
			"message":              "turn ended",
		})
		turn++
	}

	if enemyHP <= 0 {
		log = append(log, map[string]any{
			"step":       "room_end",
			"event_type": "room_end",
			"room_index": roomIndex,
			"turn":       turn,
			"result":     "cleared",
			"player_hp":  maxInt(0, playerHP),
			"enemy_hp":   0,
			"message":    "room cleared",
		})
		return roomCombatResult{PlayerWon: true, PlayerHP: maxInt(1, playerHP), Log: log}
	}

	result := "failed"
	if turn > maxTurns {
		result = "timeout"
	}
	log = append(log, map[string]any{
		"step":       "room_end",
		"event_type": "room_end",
		"room_index": roomIndex,
		"turn":       turn,
		"result":     result,
		"player_hp":  maxInt(0, playerHP),
		"enemy_hp":   maxInt(0, enemyHP),
		"message":    "room failed",
	})
	return roomCombatResult{PlayerWon: false, PlayerHP: maxInt(0, playerHP), Log: log}
}

func playerTurn(turn, roomIndex int, player combatStats, enemy roomEnemy, playerHP, enemyHP, basicCooldown, burstCooldown, healCooldown int, log []map[string]any) (int, int, int, int, int, []map[string]any) {
	cooldownBefore := map[string]int{"player_burst": burstCooldown, "player_heal": healCooldown}
	healThreshold := int(float64(player.maxHP) * 0.55)
	if player.healingPower > 0 && playerHP <= healThreshold && healCooldown == 0 {
		healValue := maxInt(8, int(float64(player.healingPower)*1.6)+int(float64(player.maxHP)*0.06))
		before := playerHP
		playerHP = minInt(player.maxHP, playerHP+healValue)
		healCooldown = 3
		log = append(log, map[string]any{
			"step":                  "action",
			"event_type":            "action",
			"room_index":            roomIndex,
			"turn":                  turn,
			"actor":                 "player",
			"target":                "player",
			"action":                "recovery_wave",
			"skill_id":              "player_recovery_wave",
			"value":                 playerHP - before,
			"value_type":            "heal",
			"target_hp_before":      before,
			"target_hp_after":       playerHP,
			"player_hp":             playerHP,
			"enemy_hp":              enemyHP,
			"cooldown_before_round": cooldownBefore,
			"cooldown_after_round":  map[string]int{"player_burst": burstCooldown, "player_heal": healCooldown},
			"message":               "player casts healing skill",
		})
		return playerHP, enemyHP, basicCooldown, burstCooldown, healCooldown, log
	}

	if burstCooldown == 0 {
		damage := computeDamage(maxInt(player.physicalAttack, player.magicAttack), enemy.physicalDefense, 1.75)
		before := enemyHP
		enemyHP = maxInt(0, enemyHP-damage)
		burstCooldown = 2
		log = append(log, map[string]any{
			"step":                  "action",
			"event_type":            "action",
			"room_index":            roomIndex,
			"turn":                  turn,
			"actor":                 "player",
			"target":                "enemy",
			"action":                "burst_skill",
			"skill_id":              "player_burst_skill",
			"damage_type":           "physical",
			"value":                 damage,
			"value_type":            "damage",
			"target_hp_before":      before,
			"target_hp_after":       enemyHP,
			"player_hp":             playerHP,
			"enemy_hp":              enemyHP,
			"cooldown_before_round": cooldownBefore,
			"cooldown_after_round":  map[string]int{"player_burst": burstCooldown, "player_heal": healCooldown},
			"message":               "player uses burst skill",
		})
		return playerHP, enemyHP, basicCooldown, burstCooldown, healCooldown, log
	}

	damage := computeDamage(maxInt(player.physicalAttack, player.magicAttack), enemy.physicalDefense, 1.0)
	before := enemyHP
	enemyHP = maxInt(0, enemyHP-damage)
	log = append(log, map[string]any{
		"step":                  "action",
		"event_type":            "action",
		"room_index":            roomIndex,
		"turn":                  turn,
		"actor":                 "player",
		"target":                "enemy",
		"action":                "basic_attack",
		"skill_id":              "player_basic_attack",
		"damage_type":           "physical",
		"value":                 damage,
		"value_type":            "damage",
		"target_hp_before":      before,
		"target_hp_after":       enemyHP,
		"player_hp":             playerHP,
		"enemy_hp":              enemyHP,
		"cooldown_before_round": cooldownBefore,
		"cooldown_after_round":  map[string]int{"player_burst": burstCooldown, "player_heal": healCooldown},
		"message":               "player attacks",
	})
	return playerHP, enemyHP, basicCooldown, burstCooldown, healCooldown, log
}

func enemyTurn(turn, roomIndex int, player combatStats, enemy roomEnemy, playerHP, enemyHP, enemyBurstCooldown int, log []map[string]any) (int, int, int, []map[string]any) {
	cooldownBefore := map[string]int{"enemy_burst": enemyBurstCooldown}
	action := "enemy_attack"
	multiplier := 1.0
	skillID := "enemy_basic_attack"
	if enemyBurstCooldown == 0 {
		action = "enemy_skill"
		skillID = "enemy_burst_skill"
		multiplier = 1.35
		enemyBurstCooldown = 2
	}

	damage := computeDamage(enemy.attack, player.physicalDefense, multiplier)
	before := playerHP
	playerHP = maxInt(0, playerHP-damage)
	log = append(log, map[string]any{
		"step":                  "action",
		"event_type":            "action",
		"room_index":            roomIndex,
		"turn":                  turn,
		"actor":                 "enemy",
		"target":                "player",
		"action":                action,
		"skill_id":              skillID,
		"damage_type":           "physical",
		"value":                 damage,
		"value_type":            "damage",
		"target_hp_before":      before,
		"target_hp_after":       playerHP,
		"player_hp":             playerHP,
		"enemy_hp":              enemyHP,
		"cooldown_before_round": cooldownBefore,
		"cooldown_after_round":  map[string]int{"enemy_burst": enemyBurstCooldown},
		"message":               "enemy attacks",
	})

	return playerHP, enemyHP, enemyBurstCooldown, log
}

func computeDamage(attack, defense int, multiplier float64) int {
	raw := int(float64(attack)*multiplier) - int(float64(defense)*0.35)
	return maxInt(1, raw)
}

func stagedMaterialDropsFromClear(highestRoomCleared int) []map[string]any {
	if highestRoomCleared <= 0 {
		return []map[string]any{}
	}
	quantity := maxInt(1, highestRoomCleared/2+1)
	return []map[string]any{{
		"material_key": "dungeon_essence",
		"quantity":     quantity,
	}}
}

func scaledRewardGold(baseGold, highestRoomCleared, totalRooms int) int {
	if totalRooms <= 0 {
		return baseGold
	}
	cleared := maxInt(0, minInt(highestRoomCleared, totalRooms))
	if cleared == 0 {
		return maxInt(1, baseGold/5)
	}
	ratio := float64(cleared) / float64(totalRooms)
	return maxInt(1, int(float64(baseGold)*(0.45+ratio*0.55)))
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func ratingFromRoomClear(highestRoom, totalRooms int) string {
	if totalRooms <= 0 {
		return "E"
	}
	if highestRoom >= totalRooms {
		return "S"
	}
	if highestRoom <= 1 {
		return "E"
	}

	stepsFromFull := totalRooms - highestRoom
	switch stepsFromFull {
	case 1:
		return "A"
	case 2:
		return "B"
	case 3:
		return "C"
	case 4:
		return "D"
	default:
		return "E"
	}
}

func buildPendingRatingRewards(runID, dungeonID, rating string, character characters.Summary) []map[string]any {
	rollCount := rollCountForRating(runID, rating)
	rewards := make([]map[string]any, 0, rollCount)
	add := func(catalogID, quality string, rollIndex int) {
		catalogID = strings.TrimSpace(catalogID)
		if catalogID == "" {
			return
		}
		rewards = append(rewards, map[string]any{
			"reward_type": "rating_chest",
			"rating":      rating,
			"quality":     quality,
			"roll_index":  rollIndex,
			"catalog_id":  catalogID,
		})
	}

	for i := 0; i < rollCount; i++ {
		forceHigh := strings.EqualFold(strings.TrimSpace(rating), "S") && i == 0
		quality := rollQualityForRating(runID, rating, i, forceHigh)
		catalogID := catalogForDungeonQuality(dungeonID, quality, character, i)
		add(catalogID, quality, i+1)
	}

	if len(rewards) == 0 {
		fallbackQuality := "blue"
		add(catalogForDungeonQuality(dungeonID, fallbackQuality, character, 0), fallbackQuality, 1)
	}

	return rewards
}

func rollCountForRating(runID, rating string) int {
	rating = strings.ToUpper(strings.TrimSpace(rating))
	seedBase := fmt.Sprintf("%s|%s|roll_count", runID, rating)
	roll := deterministicUnitFloat(seedBase)

	switch rating {
	case "S":
		return 2
	case "A":
		if roll < 0.35 {
			return 2
		}
		return 1
	case "B":
		return 1
	case "C":
		if roll < 0.75 {
			return 1
		}
		return 0
	case "D":
		if roll < 0.45 {
			return 1
		}
		return 0
	default:
		if roll < 0.20 {
			return 1
		}
		return 0
	}
}

func rollQualityForRating(runID, rating string, rollIndex int, forceHighQuality bool) string {
	rating = strings.ToUpper(strings.TrimSpace(rating))
	seedBase := fmt.Sprintf("%s|%s|quality|%d", runID, rating, rollIndex)
	roll := deterministicUnitFloat(seedBase)

	distribution := map[string]float64{}
	if forceHighQuality {
		distribution = map[string]float64{"gold": 42, "red": 32, "prismatic": 8}
	} else {
		switch rating {
		case "S":
			distribution = map[string]float64{"blue": 0, "purple": 18, "gold": 42, "red": 32, "prismatic": 8}
		case "A":
			distribution = map[string]float64{"blue": 8, "purple": 34, "gold": 36, "red": 18, "prismatic": 4}
		case "B":
			distribution = map[string]float64{"blue": 24, "purple": 42, "gold": 24, "red": 9, "prismatic": 1}
		case "C":
			distribution = map[string]float64{"blue": 46, "purple": 36, "gold": 14, "red": 4, "prismatic": 0}
		case "D":
			distribution = map[string]float64{"blue": 65, "purple": 27, "gold": 7, "red": 1, "prismatic": 0}
		default:
			distribution = map[string]float64{"blue": 82, "purple": 16, "gold": 2, "red": 0, "prismatic": 0}
		}
	}

	return pickQualityByDistribution(roll, distribution)
}

func pickQualityByDistribution(roll float64, distribution map[string]float64) string {
	orderedQualities := []string{"blue", "purple", "gold", "red", "prismatic"}
	total := 0.0
	for _, quality := range orderedQualities {
		total += distribution[quality]
	}
	if total <= 0 {
		return "blue"
	}

	threshold := roll * total
	cursor := 0.0
	for _, quality := range orderedQualities {
		cursor += distribution[quality]
		if threshold <= cursor && distribution[quality] > 0 {
			return quality
		}
	}

	for i := len(orderedQualities) - 1; i >= 0; i-- {
		quality := orderedQualities[i]
		if distribution[quality] > 0 {
			return quality
		}
	}

	return "blue"
}

func catalogForDungeonQuality(dungeonID, quality string, character characters.Summary, rollIndex int) string {
	quality = strings.ToLower(strings.TrimSpace(quality))

	if pool, ok := dungeonQualityCatalogPools[dungeonID]; ok {
		candidates := pool[quality]
		if len(candidates) > 0 {
			indexSeed := fmt.Sprintf("%s|%s|%s|%d", dungeonID, quality, character.WeaponStyle, rollIndex)
			index := deterministicIndex(indexSeed, len(candidates))
			return candidates[index]
		}
	}

	fallback := []string{"starter_boots", "starter_chest_armor", "starter_chest_cloth"}
	return fallback[deterministicIndex(fmt.Sprintf("fallback|%s|%s|%d", dungeonID, quality, rollIndex), len(fallback))]
}

func deterministicUnitFloat(seed string) float64 {
	hasher := fnv.New32a()
	_, _ = hasher.Write([]byte(seed))
	value := hasher.Sum32()
	return float64(value%1000000) / 1000000.0
}

func deterministicIndex(seed string, size int) int {
	if size <= 1 {
		return 0
	}
	hasher := fnv.New32a()
	_, _ = hasher.Write([]byte(seed))
	return int(math.Abs(float64(int32(hasher.Sum32())))) % size
}

var dungeonQualityCatalogPools = map[string]map[string][]string{
	"ancient_catacomb_v1": {
		"blue":      {"gravewake_marchers_blue", "gravewake_shackles_blue"},
		"purple":    {"gravewake_seal_purple", "gravewake_hood_purple"},
		"gold":      {"gravewake_reliquary_gold", "gravewake_vestment_gold"},
		"red":       {"gravewake_vestment_red", "gravewake_seal_red"},
		"prismatic": {"gravewake_reliquary_prismatic"},
	},
	"sandworm_den_v1": {
		"blue":      {"dunescourge_burrowstep_blue", "dunescourge_coilguards_blue"},
		"purple":    {"dunescourge_fang_ring_purple", "dunescourge_crownshell_purple"},
		"gold":      {"dunescourge_heartspine_chain_gold", "dunescourge_carapace_mail_gold"},
		"red":       {"dunescourge_carapace_mail_red", "dunescourge_fang_ring_red"},
		"prismatic": {"dunescourge_heartspine_chain_prismatic"},
	},
}

func copyMapSlice(input []map[string]any) []map[string]any {
	if len(input) == 0 {
		return []map[string]any{}
	}

	result := make([]map[string]any, 0, len(input))
	for _, item := range input {
		copied := make(map[string]any, len(item))
		for key, value := range item {
			copied[key] = value
		}
		result = append(result, copied)
	}

	return result
}

func rankAllows(currentRank, requiredRank string) bool {
	return rankOrder[currentRank] >= rankOrder[requiredRank]
}

func sortRunsByResolvedAtDesc(runs []RunView) {
	sort.Slice(runs, func(i, j int) bool {
		left := parseRunTime(runs[i].ResolvedAt)
		right := parseRunTime(runs[j].ResolvedAt)
		return left.After(right)
	})
}

func parseRunTime(value string) time.Time {
	parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(value))
	if err != nil {
		return time.Time{}
	}

	return parsed
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
			{HighestRoomCleared: 1, Rating: "E"},
			{HighestRoomCleared: 2, Rating: "D"},
			{HighestRoomCleared: 3, Rating: "C"},
			{HighestRoomCleared: 4, Rating: "B"},
			{HighestRoomCleared: 5, Rating: "A"},
			{HighestRoomCleared: 6, Rating: "S"},
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
			{HighestRoomCleared: 1, Rating: "E"},
			{HighestRoomCleared: 2, Rating: "D"},
			{HighestRoomCleared: 3, Rating: "C"},
			{HighestRoomCleared: 4, Rating: "B"},
			{HighestRoomCleared: 5, Rating: "A"},
			{HighestRoomCleared: 6, Rating: "S"},
		},
		RewardSummary: map[string]any{
			"gold_min":   320,
			"gold_max":   460,
			"drop_table": "sandworm_v1",
		},
	},
}
