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
	"clawgame/apps/api/internal/combat"
)

var (
	ErrDungeonNotFound              = errors.New("dungeon not found")
	ErrDungeonRunAlreadyActive      = errors.New("dungeon run already active")
	ErrDungeonRunNotFound           = errors.New("dungeon run not found")
	ErrDungeonRunForbidden          = errors.New("dungeon run forbidden")
	ErrDungeonRewardClaimNotAllowed = errors.New("dungeon reward not claimable")
	ErrDungeonRewardClaimCapReached = errors.New("dungeon reward claim cap reached")
	ErrDungeonPotionLoadoutInvalid  = errors.New("dungeon potion loadout invalid")
)

type DefinitionView struct {
	DungeonID           string              `json:"dungeon_id"`
	Name                string              `json:"name"`
	RegionID            string              `json:"region_id"`
	RoomCount           int                 `json:"room_count"`
	BossRoomIndex       int                 `json:"boss_room_index"`
	RecommendedLevelMin int                 `json:"recommended_level_min"`
	RecommendedLevelMax int                 `json:"recommended_level_max"`
	IsNovice            bool                `json:"is_novice"`
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
	Difficulty                string                   `json:"difficulty"`
	PotionLoadout             []string                 `json:"potion_loadout"`
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

type RunListFilters struct {
	DungeonID  string
	Difficulty string
	Result     string
	Limit      int
	Cursor     string
}

type RunSummaryView struct {
	RunID              string   `json:"run_id"`
	DungeonID          string   `json:"dungeon_id"`
	Difficulty         string   `json:"difficulty"`
	StartedAt          string   `json:"started_at"`
	ResolvedAt         string   `json:"resolved_at"`
	RunStatus          string   `json:"run_status"`
	HighestRoomCleared int      `json:"highest_room_cleared"`
	CurrentRating      *string  `json:"current_rating"`
	PotionLoadout      []string `json:"potion_loadout"`
	BossReached        bool     `json:"boss_reached"`
	SummaryTag         string   `json:"summary_tag"`
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

func BuildPlayerCombatant(character characters.Summary, stats characters.StatsSnapshot, skills characters.SkillsStateView) combat.Combatant {
	return combat.Combatant{
		EntityID:        character.CharacterID,
		Name:            character.Name,
		Team:            "a",
		IsPlayerSide:    true,
		Class:           character.Class,
		MaxHP:           stats.MaxHP,
		PhysAtk:         stats.PhysicalAttack,
		MagAtk:          stats.MagicAttack,
		PhysDef:         stats.PhysicalDefense,
		MagDef:          stats.MagicDefense,
		Speed:           stats.Speed,
		HealPow:         stats.HealingPower,
		CritRate:        stats.CritRate,
		CritDamage:      stats.CritDamage,
		BlockRate:       stats.BlockRate,
		Precision:       stats.Precision,
		EvasionRate:     stats.EvasionRate,
		PhysicalMastery: stats.PhysicalMastery,
		MagicMastery:    stats.MagicMastery,
		CurrentHP:       maxInt(1, stats.MaxHP),
		EquippedSkills:  buildEquippedSkills(skills),
		SkillCooldowns:  map[string]int{},
		BasicSkillID:    skills.BasicAttack.SkillID,
	}
}

func buildEquippedSkills(skills characters.SkillsStateView) []combat.SkillAction {
	levelByID := make(map[string]int, len(skills.Universal)+len(skills.ClassSkills))
	for _, skill := range skills.Universal {
		levelByID[skill.SkillID] = skill.Level
	}
	for _, skill := range skills.ClassSkills {
		levelByID[skill.SkillID] = skill.Level
	}

	items := make([]combat.SkillAction, 0, len(skills.ActiveLoadout))
	for _, skillID := range skills.ActiveLoadout {
		items = append(items, skillActionForID(skillID, levelByID[skillID]))
	}
	return items
}

func skillActionForID(skillID string, level int) combat.SkillAction {
	switch skillID {
	case "Quickstep":
		return combat.SkillAction{SkillID: skillID, ActionKind: "buff", Tier: "normal", CooldownRounds: 1, BuffFamily: "spd", BuffValue: 0.12, Level: level}
	case "Pocket Sand":
		return combat.SkillAction{SkillID: skillID, ActionKind: "damage", DamageType: "physical", Tier: "normal", CooldownRounds: 1, PowerScale: 0.92, DebuffFamily: "weaken", DebuffValue: 0.10, Level: level}
	case "Emergency Roll":
		return combat.SkillAction{SkillID: skillID, ActionKind: "buff", Tier: "advanced", CooldownRounds: 2, BuffFamily: "def", BuffValue: 0.16, ShieldScale: 0.08, Level: level}
	case "Signal Flare":
		return combat.SkillAction{SkillID: skillID, ActionKind: "debuff", Tier: "advanced", CooldownRounds: 2, TargetPattern: "target_3", DebuffFamily: "vulnerable", DebuffValue: 0.12, Level: level}
	case "Field Tonic":
		return combat.SkillAction{SkillID: skillID, ActionKind: "heal", Tier: "advanced", CooldownRounds: 2, HealScale: 0.85, RegenScale: 0.04, Level: level}
	case "Tripwire Kit":
		return combat.SkillAction{SkillID: skillID, ActionKind: "damage", DamageType: "physical", Tier: "ultimate", CooldownRounds: 3, PowerScale: 1.18, DebuffFamily: "slow", DebuffValue: 0.18, Level: level}
	case "Guard Stance":
		return combat.SkillAction{SkillID: skillID, ActionKind: "buff", Tier: "normal", CooldownRounds: 1, BuffFamily: "def", BuffValue: 0.16, ShieldScale: 0.10, Level: level}
	case "War Cry":
		return combat.SkillAction{SkillID: skillID, ActionKind: "buff", Tier: "normal", CooldownRounds: 1, BuffFamily: "atk", BuffValue: 0.14, Level: level}
	case "Intercept":
		return combat.SkillAction{SkillID: skillID, ActionKind: "damage", Tier: "advanced", CooldownRounds: 2, PowerScale: 1.02, DebuffFamily: "weaken", DebuffValue: 0.10, Level: level}
	case "Shield Bash":
		return combat.SkillAction{SkillID: skillID, ActionKind: "damage", DamageType: "physical", Tier: "normal", CooldownRounds: 1, PowerScale: 1.04, DebuffFamily: "slow", DebuffValue: 0.12, Level: level}
	case "Fortified Slash":
		return combat.SkillAction{SkillID: skillID, ActionKind: "damage", DamageType: "physical", Tier: "advanced", CooldownRounds: 2, PowerScale: 1.12, BuffFamily: "def", BuffValue: 0.10, Level: level}
	case "Bulwark Field":
		return combat.SkillAction{SkillID: skillID, ActionKind: "buff", Tier: "advanced", CooldownRounds: 2, BuffFamily: "def", BuffValue: 0.18, ShieldScale: 0.14, Level: level}
	case "Linebreaker":
		return combat.SkillAction{SkillID: skillID, ActionKind: "damage", DamageType: "physical", Tier: "ultimate", CooldownRounds: 3, TargetPattern: "target_3", PowerScale: 1.32, DebuffFamily: "shred", DebuffValue: 0.16, Level: level}
	case "Cleave":
		return combat.SkillAction{SkillID: skillID, ActionKind: "damage", DamageType: "physical", Tier: "normal", CooldownRounds: 1, TargetPattern: "target_3", PowerScale: 1.10, Level: level}
	case "Blood Roar":
		return combat.SkillAction{SkillID: skillID, ActionKind: "buff", Tier: "advanced", CooldownRounds: 2, BuffFamily: "atk", BuffValue: 0.18, Level: level}
	case "Execution Rush":
		return combat.SkillAction{SkillID: skillID, ActionKind: "damage", DamageType: "physical", Tier: "advanced", CooldownRounds: 2, PowerScale: 1.36, Level: level}
	case "Rending Arc":
		return combat.SkillAction{SkillID: skillID, ActionKind: "damage", DamageType: "physical", Tier: "ultimate", CooldownRounds: 3, TargetPattern: "target_3", PowerScale: 1.42, DebuffFamily: "weaken", DebuffValue: 0.12, Level: level}
	case "Runic Brand":
		return combat.SkillAction{SkillID: skillID, ActionKind: "damage", DamageType: "magic", Tier: "normal", CooldownRounds: 1, PowerScale: 1.08, DebuffFamily: "vulnerable", DebuffValue: 0.10, Level: level}
	case "Arcsteel Surge":
		return combat.SkillAction{SkillID: skillID, ActionKind: "damage", DamageType: "magic", Tier: "advanced", CooldownRounds: 2, PowerScale: 1.22, BuffFamily: "atk", BuffValue: 0.10, Level: level}
	case "Spellrend Wave":
		return combat.SkillAction{SkillID: skillID, ActionKind: "damage", DamageType: "magic", Tier: "advanced", CooldownRounds: 2, TargetPattern: "target_3", PowerScale: 1.18, DebuffFamily: "shred", DebuffValue: 0.14, Level: level}
	case "Astral Breaker":
		return combat.SkillAction{SkillID: skillID, ActionKind: "damage", DamageType: "magic", Tier: "ultimate", CooldownRounds: 3, PowerScale: 1.48, Level: level}
	case "Arc Veil":
		return combat.SkillAction{SkillID: skillID, ActionKind: "buff", Tier: "normal", CooldownRounds: 1, BuffFamily: "def", BuffValue: 0.14, ShieldScale: 0.08, Level: level}
	case "Focus Pulse":
		return combat.SkillAction{SkillID: skillID, ActionKind: "buff", Tier: "normal", CooldownRounds: 1, BuffFamily: "atk", BuffValue: 0.14, Level: level}
	case "Disrupt Ray":
		return combat.SkillAction{SkillID: skillID, ActionKind: "damage", DamageType: "magic", Tier: "advanced", CooldownRounds: 2, PowerScale: 1.06, DebuffFamily: "slow", DebuffValue: 0.12, Level: level}
	case "Hex Mark":
		return combat.SkillAction{SkillID: skillID, ActionKind: "debuff", Tier: "normal", CooldownRounds: 1, DebuffFamily: "vulnerable", DebuffValue: 0.14, Level: level}
	case "Seal Fracture":
		return combat.SkillAction{SkillID: skillID, ActionKind: "damage", DamageType: "magic", Tier: "advanced", CooldownRounds: 2, PowerScale: 1.16, DebuffFamily: "shred", DebuffValue: 0.14, Level: level}
	case "Detonate Sigil":
		return combat.SkillAction{SkillID: skillID, ActionKind: "damage", DamageType: "magic", Tier: "advanced", CooldownRounds: 2, PowerScale: 1.34, Level: level}
	case "Star Pierce":
		return combat.SkillAction{SkillID: skillID, ActionKind: "damage", DamageType: "magic", Tier: "ultimate", CooldownRounds: 3, PowerScale: 1.52, Level: level}
	case "Flame Burst":
		return combat.SkillAction{SkillID: skillID, ActionKind: "damage", DamageType: "magic", Tier: "normal", CooldownRounds: 1, TargetPattern: "target_3", PowerScale: 1.10, Level: level}
	case "Meteor Shard":
		return combat.SkillAction{SkillID: skillID, ActionKind: "damage", DamageType: "magic", Tier: "advanced", CooldownRounds: 2, PowerScale: 1.28, Level: level}
	case "Chain Script":
		return combat.SkillAction{SkillID: skillID, ActionKind: "damage", DamageType: "magic", Tier: "advanced", CooldownRounds: 2, TargetPattern: "target_3", PowerScale: 1.20, DebuffFamily: "vulnerable", DebuffValue: 0.10, Level: level}
	case "Ember Field":
		return combat.SkillAction{SkillID: skillID, ActionKind: "damage", DamageType: "magic", Tier: "ultimate", CooldownRounds: 3, TargetPattern: "all_enemies", PowerScale: 1.38, DebuffFamily: "weaken", DebuffValue: 0.10, Level: level}
	case "Frost Bind":
		return combat.SkillAction{SkillID: skillID, ActionKind: "damage", DamageType: "magic", Tier: "normal", CooldownRounds: 1, PowerScale: 1.02, DebuffFamily: "slow", DebuffValue: 0.16, Level: level}
	case "Gravity Knot":
		return combat.SkillAction{SkillID: skillID, ActionKind: "debuff", Tier: "advanced", CooldownRounds: 2, TargetPattern: "target_3", DebuffFamily: "slow", DebuffValue: 0.22, Level: level}
	case "Silencing Prism":
		return combat.SkillAction{SkillID: skillID, ActionKind: "debuff", Tier: "advanced", CooldownRounds: 2, TargetPattern: "target_2", SilenceRounds: 2, Level: level}
	case "Time Lock":
		return combat.SkillAction{SkillID: skillID, ActionKind: "debuff", Tier: "ultimate", CooldownRounds: 3, TargetPattern: "target_2", DebuffFamily: "slow", DebuffValue: 0.24, SilenceRounds: 1, Level: level}
	case "Restore":
		return combat.SkillAction{SkillID: skillID, ActionKind: "heal", Tier: "normal", CooldownRounds: 1, HealScale: 1.02, Level: level}
	case "Sanctuary Mark":
		return combat.SkillAction{SkillID: skillID, ActionKind: "buff", Tier: "normal", CooldownRounds: 1, BuffFamily: "def", BuffValue: 0.14, ShieldScale: 0.10, Level: level}
	case "Purge":
		return combat.SkillAction{SkillID: skillID, ActionKind: "heal", Tier: "advanced", CooldownRounds: 2, HealScale: 0.84, Level: level}
	case "Grace Field":
		return combat.SkillAction{SkillID: skillID, ActionKind: "buff", Tier: "normal", CooldownRounds: 1, RegenScale: 0.05, BuffFamily: "def", BuffValue: 0.08, Level: level}
	case "Purifying Wave":
		return combat.SkillAction{SkillID: skillID, ActionKind: "heal", Tier: "advanced", CooldownRounds: 2, HealScale: 1.08, Level: level}
	case "Prayer of Renewal":
		return combat.SkillAction{SkillID: skillID, ActionKind: "heal", Tier: "advanced", CooldownRounds: 2, HealScale: 1.20, RegenScale: 0.04, Level: level}
	case "Bless Armor":
		return combat.SkillAction{SkillID: skillID, ActionKind: "buff", Tier: "ultimate", CooldownRounds: 3, BuffFamily: "def", BuffValue: 0.20, ShieldScale: 0.16, Level: level}
	case "Judged Weakness":
		return combat.SkillAction{SkillID: skillID, ActionKind: "damage", DamageType: "magic", Tier: "normal", CooldownRounds: 1, PowerScale: 1.04, DebuffFamily: "weaken", DebuffValue: 0.12, Level: level}
	case "Seal of Silence":
		return combat.SkillAction{SkillID: skillID, ActionKind: "debuff", Tier: "advanced", CooldownRounds: 2, SilenceRounds: 2, Level: level}
	case "Wither Prayer":
		return combat.SkillAction{SkillID: skillID, ActionKind: "debuff", Tier: "advanced", CooldownRounds: 2, DebuffFamily: "vulnerable", DebuffValue: 0.14, Level: level}
	case "Judgment":
		return combat.SkillAction{SkillID: skillID, ActionKind: "damage", DamageType: "magic", Tier: "ultimate", CooldownRounds: 3, PowerScale: 1.46, Level: level}
	case "Sanctified Blow":
		return combat.SkillAction{SkillID: skillID, ActionKind: "damage", DamageType: "magic", Tier: "normal", CooldownRounds: 1, PowerScale: 1.00, HealScale: 0.28, Level: level}
	case "Lantern Servitor":
		return combat.SkillAction{SkillID: skillID, ActionKind: "summon", Tier: "advanced", CooldownRounds: 2, SummonTurns: 3, SummonMend: 0.22, Level: level}
	case "Censer Guardian":
		return combat.SkillAction{SkillID: skillID, ActionKind: "summon", Tier: "advanced", CooldownRounds: 2, SummonTurns: 3, SummonStrike: 0.68, ShieldScale: 0.08, Level: level}
	case "Choir Invocation":
		return combat.SkillAction{SkillID: skillID, ActionKind: "summon", Tier: "ultimate", CooldownRounds: 3, SummonTurns: 4, SummonStrike: 0.78, SummonMend: 0.18, Level: level}
	default:
		return combat.SkillAction{SkillID: skillID, ActionKind: "damage", DamageType: "", Tier: tierForSkill(skillID), CooldownRounds: cooldownForSkill(skillID), PowerScale: powerForSkill(skillID), Level: level}
	}
}

func tierForSkill(skillID string) string {
	switch skillID {
	case "Shield Bash", "Cleave", "Runic Brand", "Hex Mark", "Flame Burst", "Frost Bind", "Judged Weakness", "Sanctified Blow":
		return "normal"
	case "Fortified Slash", "Blood Roar", "Execution Rush", "Arcsteel Surge", "Spellrend Wave", "Seal Fracture", "Detonate Sigil", "Meteor Shard", "Chain Script", "Gravity Knot", "Silencing Prism", "Wither Prayer", "Seal of Silence", "Lantern Servitor", "Censer Guardian":
		return "advanced"
	case "Linebreaker", "Rending Arc", "Astral Breaker", "Star Pierce", "Ember Field", "Time Lock", "Judgment", "Choir Invocation":
		return "ultimate"
	default:
		return "advanced"
	}
}

func cooldownForSkill(skillID string) int {
	switch tierForSkill(skillID) {
	case "normal":
		return 1
	case "ultimate":
		return 3
	default:
		return 2
	}
}

func powerForSkill(skillID string) float64 {
	switch tierForSkill(skillID) {
	case "normal":
		return 1.08
	case "ultimate":
		return 1.48
	default:
		return 1.26
	}
}

func (s *Service) EnterDungeon(character characters.Summary, _ characters.DailyLimits, player combat.Combatant, dungeonID, difficulty string, potionLoadout []string, potionBag []combat.PotionItem) (RunView, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	def, ok := dungeonDefinitions[dungeonID]
	if !ok {
		return RunView{}, ErrDungeonNotFound
	}
	if len(potionLoadout) > 2 || len(potionBag) > 2 {
		return RunView{}, ErrDungeonPotionLoadoutInvalid
	}
	if len(potionLoadout) == 2 && (strings.TrimSpace(potionLoadout[0]) == "" || strings.TrimSpace(potionLoadout[1]) == "" || potionLoadout[0] == potionLoadout[1]) {
		return RunView{}, ErrDungeonPotionLoadoutInvalid
	}
	if len(potionLoadout) == 1 && strings.TrimSpace(potionLoadout[0]) == "" {
		return RunView{}, ErrDungeonPotionLoadoutInvalid
	}
	difficulty = normalizeDifficulty(difficulty)
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

	battleResult := simulateDungeonRun(player, def, difficulty, runID, potionBag)
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
		stagedMaterialDrops = stagedMaterialDropsFromClear(battleResult.HighestRoomCleared, difficulty)
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
		Difficulty:                difficulty,
		PotionLoadout:             append([]string(nil), potionLoadout...),
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
			"difficulty":      difficulty,
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
			"difficulty":          difficulty,
			"basic_attack_skill":  player.BasicSkillID,
			"equipped_skills":     append([]combat.SkillAction(nil), player.EquippedSkills...),
		},
		StagedMaterialDrops:  stagedMaterialDrops,
		PendingRatingRewards: pendingRatingRewards,
		AvailableActions:     availableActions,
		RecentBattleLog:      battleResult.Log,
	}

	rewardGold := 0
	if battleResult.Cleared {
		goldMin, goldMax := definitionRewardGoldRange(def)
		rewardGold = scaledRewardGold(goldMin, goldMax, battleResult.HighestRoomCleared, def.RoomCount, difficulty, runID)
	}

	s.runsByID[runID] = runRecord{
		CharacterID: character.CharacterID,
		Run:         run,
		RewardGold:  rewardGold,
	}
	delete(s.activeRunByCharacter, character.CharacterID)

	return run, nil
}

func normalizeDifficulty(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "hard":
		return "hard"
	case "nightmare":
		return "nightmare"
	default:
		return "easy"
	}
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

func (s *Service) ListRuns(characterID string, filters RunListFilters) []RunSummaryView {
	runs := s.ListRunsByCharacter(characterID)
	if filters.Limit <= 0 {
		filters.Limit = 20
	}
	if filters.Limit > 100 {
		filters.Limit = 100
	}

	startIndex := 0
	if cursor := strings.TrimSpace(filters.Cursor); cursor != "" {
		for index, run := range runs {
			if run.RunID == cursor {
				startIndex = index + 1
				break
			}
		}
	}

	items := make([]RunSummaryView, 0, minInt(filters.Limit, len(runs)))
	for _, run := range runs[startIndex:] {
		if strings.TrimSpace(filters.DungeonID) != "" && run.DungeonID != strings.TrimSpace(filters.DungeonID) {
			continue
		}
		if difficulty := normalizeDifficulty(filters.Difficulty); strings.TrimSpace(filters.Difficulty) != "" && run.Difficulty != difficulty {
			continue
		}
		if result := normalizeRunResult(filters.Result); result != "" && run.RunStatus != result {
			continue
		}

		items = append(items, buildRunSummary(run))
		if len(items) >= filters.Limit {
			break
		}
	}

	return items
}

func (s *Service) BuildRunPayload(run RunView, detailLevel string) map[string]any {
	detailLevel = normalizeRunDetailLevel(detailLevel)

	payload := map[string]any{
		"run_id":                       run.RunID,
		"dungeon_id":                   run.DungeonID,
		"difficulty":                   run.Difficulty,
		"potion_loadout":               append([]string(nil), run.PotionLoadout...),
		"started_at":                   run.StartedAt,
		"resolved_at":                  run.ResolvedAt,
		"run_status":                   run.RunStatus,
		"runtime_phase":                run.RuntimePhase,
		"reward_claimable":             run.RewardClaimable,
		"reward_claimed_at":            run.RewardClaimedAt,
		"claim_consumes_daily_counter": run.ClaimConsumesDailyCounter,
		"current_room_index":           run.CurrentRoomIndex,
		"highest_room_cleared":         run.HighestRoomCleared,
		"projected_rating":             run.ProjectedRating,
		"current_rating":               run.CurrentRating,
		"summary_tag":                  runSummaryTag(run),
		"boss_reached":                 bossReached(run),
	}

	switch detailLevel {
	case "compact":
		return payload
	case "verbose":
		payload["room_summary"] = copyMap(run.RoomSummary)
		payload["battle_state"] = copyMap(run.BattleState)
		payload["staged_material_drops"] = copyMapSlice(run.StagedMaterialDrops)
		payload["pending_rating_rewards"] = copyMapSlice(run.PendingRatingRewards)
		payload["available_actions"] = copyValidActions(run.AvailableActions)
		payload["recent_battle_log"] = copyMapSlice(run.RecentBattleLog)
		payload["battle_log"] = copyMapSlice(run.RecentBattleLog)
		payload["key_findings"] = buildKeyFindings(run)
		payload["danger_rooms"] = buildDangerRooms(run)
		payload["resource_pressure"] = buildResourcePressure(run)
		payload["reward_summary"] = buildRewardSummary(run)
		return payload
	default:
		payload["room_summary"] = copyMap(run.RoomSummary)
		payload["battle_state"] = copyMap(run.BattleState)
		payload["staged_material_drops"] = copyMapSlice(run.StagedMaterialDrops)
		payload["pending_rating_rewards"] = copyMapSlice(run.PendingRatingRewards)
		payload["available_actions"] = copyValidActions(run.AvailableActions)
		payload["recent_battle_log"] = copyRecentBattleLog(run.RecentBattleLog, 20)
		payload["key_findings"] = buildKeyFindings(run)
		payload["danger_rooms"] = buildDangerRooms(run)
		payload["resource_pressure"] = buildResourcePressure(run)
		payload["reward_summary"] = buildRewardSummary(run)
		return payload
	}
}

func (s *Service) PreviewRunRewards(characterID, runID string, limits characters.DailyLimits) (RunView, ClaimRewardPackage, error) {
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

func (s *Service) FinalizeRunRewards(characterID, runID string) (RunView, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.runsByID[runID]
	if !ok {
		return RunView{}, ErrDungeonRunNotFound
	}
	if record.CharacterID != characterID {
		return RunView{}, ErrDungeonRunForbidden
	}
	if record.Run.RunStatus != "cleared" || !record.Run.RewardClaimable {
		return RunView{}, ErrDungeonRewardClaimNotAllowed
	}

	now := s.clock().Format(time.RFC3339)
	record.Run.RewardClaimable = false
	record.Run.RewardClaimedAt = &now
	record.Run.RuntimePhase = "claim_settled"
	record.Run.AvailableActions = []characters.ValidAction{}
	record.Run.ClaimConsumesDailyCounter = true
	record.Run.StagedMaterialDrops = []map[string]any{}
	record.Run.PendingRatingRewards = []map[string]any{}

	s.runsByID[runID] = record
	return record.Run, nil
}

func nextID(prefix string) string {
	return fmt.Sprintf("%s_%d_%06d", prefix, time.Now().UnixNano(), atomic.AddUint64(&runCounter, 1))
}

func definitionRewardGoldRange(def DefinitionView) (int, int) {
	minValue := 120
	maxValue := 120
	if value, ok := def.RewardSummary["gold_min"].(int); ok {
		minValue = value
	}
	if value, ok := def.RewardSummary["gold_min"].(float64); ok {
		minValue = int(value)
	}
	if value, ok := def.RewardSummary["gold_max"].(int); ok {
		maxValue = value
	}
	if value, ok := def.RewardSummary["gold_max"].(float64); ok {
		maxValue = int(value)
	}
	if maxValue < minValue {
		maxValue = minValue
	}
	return minValue, maxValue
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

func simulateDungeonRun(player combat.Combatant, def DefinitionView, difficulty, runID string, potionBag []combat.PotionItem) simulatedRunResult {
	player.Team = "a"
	player.IsPlayerSide = true
	player.CurrentHP = maxInt(1, player.MaxHP)
	player.PotionBag = copyPotionBag(potionBag)

	startHP := player.CurrentHP
	runPotionUsed := 0
	log := make([]map[string]any, 0, def.RoomCount*12)

	for roomIndex := 1; roomIndex <= def.RoomCount; roomIndex++ {
		composition := buildRoomComposition(def.DungeonID, difficulty, roomIndex, def.BossRoomIndex)
		enemies := buildRoomEnemyCombatants(def, difficulty, roomIndex, composition)
		isBoss := composition.hasBoss(def.BossRoomIndex, roomIndex)
		log = append(log, map[string]any{
			"step":       "room_composition",
			"event_type": "room_composition",
			"room_index": roomIndex,
			"difficulty": difficulty,
			"monsters":   composition.logView(),
			"message":    "room enemy composition prepared",
		})

		result := combat.SimulateSkirmish(combat.SkirmishConfig{
			BattleType:    "dungeon_wave",
			RunID:         runID,
			RoomIndex:     roomIndex,
			IsBossRoom:    isBoss,
			Player:        player,
			Enemies:       enemies,
			RunPotionUsed: runPotionUsed,
		})

		log = append(log, result.Log...)
		if len(log) > 120 {
			log = log[len(log)-120:]
		}
		runPotionUsed += result.PotionsConsumed

		if !result.PlayerWon {
			return simulatedRunResult{
				Cleared:            false,
				PlayerSurvived:     result.PlayerFinalHP > 0,
				StartHP:            startHP,
				RemainingHP:        maxInt(0, result.PlayerFinalHP),
				CurrentRoomIndex:   roomIndex,
				HighestRoomCleared: roomIndex - 1,
				EndedInRoomIndex:   roomIndex,
				Log:                log,
			}
		}

		// HP carries over to the next room.
		player.CurrentHP = result.PlayerFinalHP

		// Recovery between rooms: none for standard dungeons, 5 % max_hp for novice dungeons.
		if roomIndex < def.RoomCount && def.IsNovice {
			recovery := maxInt(1, int(float64(player.MaxHP)*0.05))
			beforeHP := player.CurrentHP
			player.CurrentHP = minInt(player.MaxHP, player.CurrentHP+recovery)
			log = append(log, map[string]any{
				"step":                  "room_recovery",
				"event_type":            "room_recovery",
				"room_index":            roomIndex,
				"turn":                  0,
				"actor":                 "system",
				"target":                "player",
				"action":                "novice_rest",
				"value":                 player.CurrentHP - beforeHP,
				"value_type":            "heal",
				"target_hp_before":      beforeHP,
				"target_hp_after":       player.CurrentHP,
				"player_hp":             player.CurrentHP,
				"enemy_hp":              0,
				"cooldown_before_round": map[string]int{},
				"cooldown_after_round":  map[string]int{},
				"message":               "novice dungeon recovery between rooms",
			})
		}
	}

	return simulatedRunResult{
		Cleared:            true,
		PlayerSurvived:     true,
		StartHP:            startHP,
		RemainingHP:        player.CurrentHP,
		CurrentRoomIndex:   def.RoomCount,
		HighestRoomCleared: def.RoomCount,
		EndedInRoomIndex:   def.RoomCount,
		Log:                log,
	}
}

func copyPotionBag(items []combat.PotionItem) []combat.PotionItem {
	copied := make([]combat.PotionItem, len(items))
	copy(copied, items)
	return copied
}

func buildRunSummary(run RunView) RunSummaryView {
	return RunSummaryView{
		RunID:              run.RunID,
		DungeonID:          run.DungeonID,
		Difficulty:         run.Difficulty,
		StartedAt:          run.StartedAt,
		ResolvedAt:         run.ResolvedAt,
		RunStatus:          run.RunStatus,
		HighestRoomCleared: run.HighestRoomCleared,
		CurrentRating:      run.CurrentRating,
		PotionLoadout:      append([]string(nil), run.PotionLoadout...),
		BossReached:        bossReached(run),
		SummaryTag:         runSummaryTag(run),
	}
}

func normalizeRunResult(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "cleared":
		return "cleared"
	case "failed":
		return "failed"
	case "abandoned":
		return "abandoned"
	case "expired":
		return "expired"
	default:
		return ""
	}
}

func normalizeRunDetailLevel(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "compact":
		return "compact"
	case "verbose":
		return "verbose"
	default:
		return "standard"
	}
}

func bossReached(run RunView) bool {
	bossRoomIndex := 0
	switch value := run.RoomSummary["boss_room_index"].(type) {
	case int:
		bossRoomIndex = value
	case int32:
		bossRoomIndex = int(value)
	case int64:
		bossRoomIndex = int(value)
	case float64:
		bossRoomIndex = int(value)
	}
	if bossRoomIndex <= 0 {
		return false
	}
	return run.HighestRoomCleared >= bossRoomIndex || run.CurrentRoomIndex >= bossRoomIndex
}

func runSummaryTag(run RunView) string {
	startHP := intFromMap(run.BattleState, "start_hp")
	remainingHP := intFromMap(run.BattleState, "remaining_hp")
	if run.RunStatus == "cleared" {
		if startHP > 0 && remainingHP*100 < startHP*55 {
			return "boss_low_hp_clear"
		}
		return "cleared_stable"
	}
	if bossReached(run) {
		return "failed_at_boss"
	}
	if run.CurrentRoomIndex > 0 {
		return fmt.Sprintf("failed_room_%d", run.CurrentRoomIndex)
	}
	return "failed_before_boss"
}

func buildKeyFindings(run RunView) []string {
	findings := make([]string, 0, 4)
	findings = append(findings, runSummaryTag(run))
	if usage := PotionUsageFromLog(run.RecentBattleLog); len(usage) > 0 {
		findings = append(findings, fmt.Sprintf("potion_types_used_%d", len(usage)))
	}
	if hp := buildResourcePressure(run); hp["boss_entry_hp_percent"] != nil {
		findings = append(findings, fmt.Sprintf("boss_entry_hp_%d_percent", hp["boss_entry_hp_percent"]))
	}
	if len(buildDangerRooms(run)) > 0 {
		first := buildDangerRooms(run)[0]
		if roomIndex, ok := first["room_index"].(int); ok {
			findings = append(findings, fmt.Sprintf("room_%d_damage_spike", roomIndex))
		}
	}
	return findings
}

func buildDangerRooms(run RunView) []map[string]any {
	damageByRoom := map[int]int{}
	for _, entry := range run.RecentBattleLog {
		valueType, _ := entry["value_type"].(string)
		target, _ := entry["target"].(string)
		if valueType != "damage" || target != "player" {
			continue
		}
		roomIndex := intFromAny(entry["room_index"])
		if roomIndex <= 0 {
			continue
		}
		damageByRoom[roomIndex] += intFromAny(entry["value"])
	}
	type roomDamage struct {
		roomIndex int
		damage    int
	}
	rooms := make([]roomDamage, 0, len(damageByRoom))
	for roomIndex, damage := range damageByRoom {
		rooms = append(rooms, roomDamage{roomIndex: roomIndex, damage: damage})
	}
	sort.Slice(rooms, func(i, j int) bool {
		if rooms[i].damage == rooms[j].damage {
			return rooms[i].roomIndex < rooms[j].roomIndex
		}
		return rooms[i].damage > rooms[j].damage
	})
	if len(rooms) > 3 {
		rooms = rooms[:3]
	}
	items := make([]map[string]any, 0, len(rooms))
	for _, room := range rooms {
		items = append(items, map[string]any{
			"room_index":   room.roomIndex,
			"damage_taken": room.damage,
			"is_boss_room": room.roomIndex == intFromMap(run.RoomSummary, "boss_room_index"),
		})
	}
	return items
}

func buildResourcePressure(run RunView) map[string]any {
	startHP := intFromMap(run.BattleState, "start_hp")
	remainingHP := intFromMap(run.BattleState, "remaining_hp")
	hpEndPercent := 0
	if startHP > 0 {
		hpEndPercent = int(math.Round(float64(remainingHP) * 100 / float64(startHP)))
	}
	usage := PotionUsageFromLog(run.RecentBattleLog)
	totalPotionUsed := 0
	for _, count := range usage {
		totalPotionUsed += count
	}
	result := map[string]any{
		"hp_end_percent":     hpEndPercent,
		"potion_usage":       usage,
		"total_potions_used": totalPotionUsed,
	}
	bossRoomIndex := intFromMap(run.RoomSummary, "boss_room_index")
	if bossRoomIndex > 0 {
		if bossEntryHP := bossEntryHPPercent(run.RecentBattleLog, startHP, bossRoomIndex); bossEntryHP >= 0 {
			result["boss_entry_hp_percent"] = bossEntryHP
		}
	}
	return result
}

func buildRewardSummary(run RunView) map[string]any {
	return map[string]any{
		"reward_claimable":       run.RewardClaimable,
		"staged_material_drops":  copyMapSlice(run.StagedMaterialDrops),
		"pending_rating_rewards": copyMapSlice(run.PendingRatingRewards),
	}
}

func bossEntryHPPercent(log []map[string]any, startHP, bossRoomIndex int) int {
	if startHP <= 0 || bossRoomIndex <= 0 {
		return -1
	}
	for _, entry := range log {
		if intFromAny(entry["room_index"]) != bossRoomIndex {
			continue
		}
		if eventType, _ := entry["event_type"].(string); eventType == "room_composition" {
			switch hp := entry["player_hp"].(type) {
			case int:
				return int(math.Round(float64(hp) * 100 / float64(startHP)))
			case float64:
				return int(math.Round(hp * 100 / float64(startHP)))
			}
		}
	}
	return -1
}

func copyRecentBattleLog(log []map[string]any, limit int) []map[string]any {
	if limit <= 0 || len(log) <= limit {
		return copyMapSlice(log)
	}
	return copyMapSlice(log[len(log)-limit:])
}

func copyMap(input map[string]any) map[string]any {
	if input == nil {
		return map[string]any{}
	}
	output := make(map[string]any, len(input))
	for key, value := range input {
		output[key] = value
	}
	return output
}

func copyValidActions(actions []characters.ValidAction) []characters.ValidAction {
	copied := make([]characters.ValidAction, len(actions))
	copy(copied, actions)
	return copied
}

func intFromMap(payload map[string]any, key string) int {
	if payload == nil {
		return 0
	}
	return intFromAny(payload[key])
}

func intFromAny(value any) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int32:
		return int(typed)
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	default:
		return 0
	}
}

func PotionUsageFromLog(log []map[string]any) map[string]int {
	usage := map[string]int{}
	for _, entry := range log {
		if eventType, _ := entry["event_type"].(string); eventType != "potion.consumed" {
			continue
		}
		potionID, _ := entry["potion_id"].(string)
		if strings.TrimSpace(potionID) == "" {
			continue
		}
		usage[potionID]++
	}
	return usage
}

type difficultyMultiplier struct {
	hp      float64
	damage  float64
	defense float64
	speed   float64
}

type roomMonsterSlot struct {
	MonsterID string
	Tier      string
	Role      string
	Count     int
}

type roomComposition struct {
	slots []roomMonsterSlot
}

func (c roomComposition) hasBoss(bossRoomIndex, roomIndex int) bool {
	if roomIndex != bossRoomIndex {
		return false
	}
	for _, slot := range c.slots {
		if slot.Count > 0 && strings.EqualFold(slot.Tier, "boss") {
			return true
		}
	}
	return false
}

func (c roomComposition) logView() []map[string]any {
	items := make([]map[string]any, 0, len(c.slots))
	for _, slot := range c.slots {
		if slot.Count <= 0 {
			continue
		}
		items = append(items, map[string]any{
			"monster_id": slot.MonsterID,
			"tier":       slot.Tier,
			"role":       slot.Role,
			"count":      slot.Count,
		})
	}
	return items
}

type monsterBlueprint struct {
	Name    string
	Tier    string
	Role    string
	MaxHP   int
	PhysAtk int
	MagAtk  int
	PhysDef int
	MagDef  int
	Speed   int
	HealPow int
}

func buildRoomEnemyCombatants(def DefinitionView, difficulty string, roomIndex int, composition roomComposition) []combat.Combatant {
	mult := difficultyMultipliers[normalizeDifficulty(difficulty)]
	roomScale := 0.70 + 0.05*float64(roomIndex-1)
	enemies := make([]combat.Combatant, 0, 4)

	for _, slot := range composition.slots {
		if slot.Count <= 0 {
			continue
		}
		blueprint, ok := monsterBlueprints[slot.MonsterID]
		if !ok {
			continue
		}
		for copyIndex := 0; copyIndex < slot.Count; copyIndex++ {
			maxHP := maxInt(1, int(float64(blueprint.MaxHP)*0.26*roomScale*mult.hp))
			physAtk := maxInt(1, int(float64(blueprint.PhysAtk)*0.22*mult.damage))
			magAtk := maxInt(0, int(float64(blueprint.MagAtk)*0.22*mult.damage))
			physDef := maxInt(1, int(float64(blueprint.PhysDef)*0.38*mult.defense))
			magDef := maxInt(1, int(float64(blueprint.MagDef)*0.38*mult.defense))
			speed := maxInt(1, int(float64(blueprint.Speed)*0.95*mult.speed))
			heal := maxInt(0, int(float64(blueprint.HealPow)*0.40*mult.defense))

			enemies = append(enemies, combat.Combatant{
				EntityID:     fmt.Sprintf("%s_%s_room_%d_%s_%d", def.DungeonID, normalizeDifficulty(difficulty), roomIndex, slot.MonsterID, copyIndex+1),
				Name:         blueprint.Name,
				Team:         "b",
				IsPlayerSide: false,
				Role:         firstNonEmpty(slot.Role, blueprint.Role),
				MaxHP:        maxHP,
				PhysAtk:      physAtk,
				MagAtk:       magAtk,
				PhysDef:      physDef,
				MagDef:       magDef,
				Speed:        speed,
				HealPow:      heal,
				CritRate:     0,
				CritDamage:   0,
				BlockRate:    0,
				Precision:    0,
				EvasionRate:  0,
				CurrentHP:    maxHP,
			})
		}
	}

	if len(enemies) > 0 {
		return enemies
	}

	fallback := 40 + roomIndex*12
	return []combat.Combatant{{
		EntityID:     fmt.Sprintf("%s_room_%d", def.DungeonID, roomIndex),
		Name:         "room_enemy",
		Team:         "b",
		IsPlayerSide: false,
		Role:         "bruiser",
		MaxHP:        fallback,
		PhysAtk:      7 + roomIndex*2,
		PhysDef:      3 + roomIndex/2,
		Speed:        7 + roomIndex/2,
		CritRate:     0,
		CritDamage:   0,
		BlockRate:    0,
		Precision:    0,
		EvasionRate:  0,
		CurrentHP:    fallback,
	}}
}

func buildRoomComposition(dungeonID, difficulty string, roomIndex, bossRoomIndex int) roomComposition {
	difficulty = normalizeDifficulty(difficulty)
	if byDifficulty, ok := dungeonRoomCompositions[dungeonID]; ok {
		if byRoom, ok := byDifficulty[difficulty]; ok {
			if composition, ok := byRoom[roomIndex]; ok {
				return composition
			}
		}
	}

	// Fallback keeps the historical behavior when no table is found.
	tier := "normal"
	name := "fallback_enemy"
	if roomIndex == bossRoomIndex {
		tier = "boss"
		name = "fallback_boss"
	}
	return roomComposition{slots: []roomMonsterSlot{{MonsterID: name, Tier: tier, Role: "bruiser", Count: 1}}}
}

var difficultyMultipliers = map[string]difficultyMultiplier{
	"easy":      {hp: 1.00, damage: 1.00, defense: 1.00, speed: 1.00},
	"hard":      {hp: 1.12, damage: 1.08, defense: 1.05, speed: 1.04},
	"nightmare": {hp: 1.38, damage: 1.28, defense: 1.18, speed: 1.10},
}

var monsterBlueprints = map[string]monsterBlueprint{
	// Ancient Catacomb
	"catacomb_boneguard":    {Name: "Catacomb Boneguard", Tier: "normal", Role: "bruiser", MaxHP: 132, PhysAtk: 18, MagAtk: 0, PhysDef: 10, MagDef: 6, Speed: 8, HealPow: 0},
	"ashen_skull_caster":    {Name: "Ashen Skull Caster", Tier: "normal", Role: "caster", MaxHP: 102, PhysAtk: 4, MagAtk: 20, PhysDef: 5, MagDef: 9, Speed: 10, HealPow: 0},
	"grave_rat_swarm":       {Name: "Grave Rat Swarm", Tier: "normal", Role: "assassin", MaxHP: 94, PhysAtk: 16, MagAtk: 0, PhysDef: 4, MagDef: 4, Speed: 14, HealPow: 0},
	"warden_of_seals":       {Name: "Warden of Seals", Tier: "elite", Role: "tank", MaxHP: 286, PhysAtk: 22, MagAtk: 8, PhysDef: 18, MagDef: 12, Speed: 7, HealPow: 0},
	"tomb_hexer":            {Name: "Tomb Hexer", Tier: "elite", Role: "controller", MaxHP: 234, PhysAtk: 6, MagAtk: 26, PhysDef: 8, MagDef: 14, Speed: 12, HealPow: 0},
	"morthis_chapel_keeper": {Name: "Morthis, Chapel Keeper", Tier: "boss", Role: "boss", MaxHP: 620, PhysAtk: 16, MagAtk: 38, PhysDef: 14, MagDef: 16, Speed: 11, HealPow: 18},

	// Thorned Hollow
	"thornclaw_stalker":     {Name: "Thornclaw Stalker", Tier: "normal", Role: "assassin", MaxHP: 124, PhysAtk: 24, MagAtk: 0, PhysDef: 8, MagDef: 6, Speed: 17, HealPow: 0},
	"briar_hexcaller":       {Name: "Briar Hexcaller", Tier: "normal", Role: "caster", MaxHP: 110, PhysAtk: 5, MagAtk: 24, PhysDef: 6, MagDef: 10, Speed: 13, HealPow: 0},
	"hollow_trapper":        {Name: "Hollow Trapper", Tier: "normal", Role: "controller", MaxHP: 132, PhysAtk: 18, MagAtk: 6, PhysDef: 10, MagDef: 8, Speed: 14, HealPow: 0},
	"needlemane_huntmaster": {Name: "Needlemane Huntmaster", Tier: "elite", Role: "bruiser", MaxHP: 292, PhysAtk: 32, MagAtk: 0, PhysDef: 16, MagDef: 10, Speed: 14, HealPow: 0},
	"thornveil_predator":    {Name: "Thornveil Predator", Tier: "elite", Role: "assassin", MaxHP: 248, PhysAtk: 34, MagAtk: 0, PhysDef: 10, MagDef: 8, Speed: 19, HealPow: 0},
	"briarqueen_ravastra":   {Name: "Briarqueen Ravastra", Tier: "boss", Role: "boss", MaxHP: 670, PhysAtk: 36, MagAtk: 18, PhysDef: 14, MagDef: 12, Speed: 18, HealPow: 0},

	// Sunscar Warvault
	"warvault_legionary":  {Name: "Warvault Legionary", Tier: "normal", Role: "bruiser", MaxHP: 148, PhysAtk: 26, MagAtk: 0, PhysDef: 13, MagDef: 8, Speed: 11, HealPow: 0},
	"sunforge_channeler":  {Name: "Sunforge Channeler", Tier: "normal", Role: "caster", MaxHP: 118, PhysAtk: 4, MagAtk: 26, PhysDef: 7, MagDef: 11, Speed: 12, HealPow: 0},
	"shieldline_sentinel": {Name: "Shieldline Sentinel", Tier: "normal", Role: "tank", MaxHP: 164, PhysAtk: 20, MagAtk: 0, PhysDef: 16, MagDef: 10, Speed: 9, HealPow: 0},
	"vault_breaker":       {Name: "Vault Breaker", Tier: "elite", Role: "bruiser", MaxHP: 320, PhysAtk: 38, MagAtk: 0, PhysDef: 18, MagDef: 10, Speed: 12, HealPow: 0},
	"standard_bearer":     {Name: "Standard Bearer", Tier: "elite", Role: "controller", MaxHP: 274, PhysAtk: 20, MagAtk: 12, PhysDef: 15, MagDef: 12, Speed: 13, HealPow: 0},
	"sunscar_warmarshal":  {Name: "Sunscar Warmarshal", Tier: "boss", Role: "boss", MaxHP: 740, PhysAtk: 40, MagAtk: 18, PhysDef: 18, MagDef: 12, Speed: 15, HealPow: 0},

	// Obsidian Spire
	"blackglass_sentinel": {Name: "Blackglass Sentinel", Tier: "normal", Role: "tank", MaxHP: 150, PhysAtk: 22, MagAtk: 0, PhysDef: 15, MagDef: 12, Speed: 10, HealPow: 0},
	"void_adept":          {Name: "Void Adept", Tier: "normal", Role: "caster", MaxHP: 116, PhysAtk: 4, MagAtk: 28, PhysDef: 7, MagDef: 13, Speed: 14, HealPow: 0},
	"eclipse_scribe":      {Name: "Eclipse Scribe", Tier: "normal", Role: "controller", MaxHP: 108, PhysAtk: 4, MagAtk: 24, PhysDef: 6, MagDef: 12, Speed: 15, HealPow: 0},
	"mirrorbound_rector":  {Name: "Mirrorbound Rector", Tier: "elite", Role: "caster", MaxHP: 258, PhysAtk: 6, MagAtk: 34, PhysDef: 10, MagDef: 16, Speed: 15, HealPow: 0},
	"obsidian_lancer":     {Name: "Obsidian Lancer", Tier: "elite", Role: "bruiser", MaxHP: 282, PhysAtk: 30, MagAtk: 10, PhysDef: 16, MagDef: 13, Speed: 13, HealPow: 0},
	"obsidian_archon":     {Name: "Obsidian Archon", Tier: "boss", Role: "boss", MaxHP: 710, PhysAtk: 18, MagAtk: 42, PhysDef: 15, MagDef: 18, Speed: 16, HealPow: 16},

	// Sandworm Den (archived reference)
	"dune_skitterer":                {Name: "Dune Skitterer", Tier: "normal", Role: "assassin", MaxHP: 108, PhysAtk: 22, MagAtk: 0, PhysDef: 6, MagDef: 6, Speed: 16, HealPow: 0},
	"sand_burrower":                 {Name: "Sand Burrower", Tier: "normal", Role: "bruiser", MaxHP: 140, PhysAtk: 24, MagAtk: 0, PhysDef: 12, MagDef: 8, Speed: 9, HealPow: 0},
	"scorched_spitter":              {Name: "Scorched Spitter", Tier: "normal", Role: "caster", MaxHP: 112, PhysAtk: 4, MagAtk: 22, PhysDef: 6, MagDef: 10, Speed: 12, HealPow: 0},
	"carapace_crusher":              {Name: "Carapace Crusher", Tier: "elite", Role: "bruiser", MaxHP: 310, PhysAtk: 36, MagAtk: 8, PhysDef: 20, MagDef: 12, Speed: 8, HealPow: 0},
	"venom_herald":                  {Name: "Venom Herald", Tier: "elite", Role: "controller", MaxHP: 262, PhysAtk: 8, MagAtk: 28, PhysDef: 10, MagDef: 16, Speed: 13, HealPow: 0},
	"kharzug_dunescourge_matriarch": {Name: "Kharzug, Dunescourge Matriarch", Tier: "boss", Role: "boss", MaxHP: 720, PhysAtk: 32, MagAtk: 18, PhysDef: 20, MagDef: 14, Speed: 10, HealPow: 0},
	"sandworm_larva":                {Name: "Sandworm Larva", Tier: "normal", Role: "summoner", MaxHP: 88, PhysAtk: 14, MagAtk: 0, PhysDef: 5, MagDef: 4, Speed: 11, HealPow: 0},
}

var dungeonRoomCompositions = map[string]map[string]map[int]roomComposition{
	"ancient_catacomb_v1": {
		"easy": {
			1: {slots: []roomMonsterSlot{{MonsterID: "catacomb_boneguard", Tier: "normal", Role: "bruiser", Count: 1}, {MonsterID: "ashen_skull_caster", Tier: "normal", Role: "caster", Count: 1}}},
			2: {slots: []roomMonsterSlot{{MonsterID: "catacomb_boneguard", Tier: "normal", Role: "bruiser", Count: 1}, {MonsterID: "grave_rat_swarm", Tier: "normal", Role: "assassin", Count: 1}, {MonsterID: "ashen_skull_caster", Tier: "normal", Role: "caster", Count: 1}}},
			3: {slots: []roomMonsterSlot{{MonsterID: "catacomb_boneguard", Tier: "normal", Role: "bruiser", Count: 1}, {MonsterID: "warden_of_seals", Tier: "elite", Role: "tank", Count: 1}}},
			4: {slots: []roomMonsterSlot{{MonsterID: "grave_rat_swarm", Tier: "normal", Role: "assassin", Count: 1}, {MonsterID: "ashen_skull_caster", Tier: "normal", Role: "caster", Count: 1}, {MonsterID: "tomb_hexer", Tier: "elite", Role: "controller", Count: 1}}},
			5: {slots: []roomMonsterSlot{{MonsterID: "warden_of_seals", Tier: "elite", Role: "tank", Count: 1}, {MonsterID: "tomb_hexer", Tier: "elite", Role: "controller", Count: 1}, {MonsterID: "catacomb_boneguard", Tier: "normal", Role: "bruiser", Count: 1}}},
			6: {slots: []roomMonsterSlot{{MonsterID: "morthis_chapel_keeper", Tier: "boss", Role: "boss", Count: 1}, {MonsterID: "catacomb_boneguard", Tier: "normal", Role: "bruiser", Count: 1}}},
		},
		"hard": {
			1: {slots: []roomMonsterSlot{{MonsterID: "catacomb_boneguard", Tier: "normal", Role: "bruiser", Count: 2}, {MonsterID: "ashen_skull_caster", Tier: "normal", Role: "caster", Count: 1}}},
			2: {slots: []roomMonsterSlot{{MonsterID: "catacomb_boneguard", Tier: "normal", Role: "bruiser", Count: 1}, {MonsterID: "grave_rat_swarm", Tier: "normal", Role: "assassin", Count: 2}, {MonsterID: "ashen_skull_caster", Tier: "normal", Role: "caster", Count: 1}}},
			3: {slots: []roomMonsterSlot{{MonsterID: "warden_of_seals", Tier: "elite", Role: "tank", Count: 1}, {MonsterID: "catacomb_boneguard", Tier: "normal", Role: "bruiser", Count: 1}, {MonsterID: "ashen_skull_caster", Tier: "normal", Role: "caster", Count: 1}}},
			4: {slots: []roomMonsterSlot{{MonsterID: "tomb_hexer", Tier: "elite", Role: "controller", Count: 1}, {MonsterID: "grave_rat_swarm", Tier: "normal", Role: "assassin", Count: 1}, {MonsterID: "ashen_skull_caster", Tier: "normal", Role: "caster", Count: 1}, {MonsterID: "catacomb_boneguard", Tier: "normal", Role: "bruiser", Count: 1}}},
			5: {slots: []roomMonsterSlot{{MonsterID: "warden_of_seals", Tier: "elite", Role: "tank", Count: 1}, {MonsterID: "tomb_hexer", Tier: "elite", Role: "controller", Count: 1}, {MonsterID: "catacomb_boneguard", Tier: "normal", Role: "bruiser", Count: 1}, {MonsterID: "grave_rat_swarm", Tier: "normal", Role: "assassin", Count: 1}}},
			6: {slots: []roomMonsterSlot{{MonsterID: "morthis_chapel_keeper", Tier: "boss", Role: "boss", Count: 1}, {MonsterID: "warden_of_seals", Tier: "elite", Role: "tank", Count: 1}, {MonsterID: "ashen_skull_caster", Tier: "normal", Role: "caster", Count: 1}}},
		},
		"nightmare": {
			1: {slots: []roomMonsterSlot{{MonsterID: "catacomb_boneguard", Tier: "normal", Role: "bruiser", Count: 2}, {MonsterID: "ashen_skull_caster", Tier: "normal", Role: "caster", Count: 1}, {MonsterID: "grave_rat_swarm", Tier: "normal", Role: "assassin", Count: 1}}},
			2: {slots: []roomMonsterSlot{{MonsterID: "warden_of_seals", Tier: "elite", Role: "tank", Count: 1}, {MonsterID: "grave_rat_swarm", Tier: "normal", Role: "assassin", Count: 1}, {MonsterID: "ashen_skull_caster", Tier: "normal", Role: "caster", Count: 1}, {MonsterID: "catacomb_boneguard", Tier: "normal", Role: "bruiser", Count: 1}}},
			3: {slots: []roomMonsterSlot{{MonsterID: "warden_of_seals", Tier: "elite", Role: "tank", Count: 1}, {MonsterID: "tomb_hexer", Tier: "elite", Role: "controller", Count: 1}, {MonsterID: "catacomb_boneguard", Tier: "normal", Role: "bruiser", Count: 1}, {MonsterID: "grave_rat_swarm", Tier: "normal", Role: "assassin", Count: 1}}},
			4: {slots: []roomMonsterSlot{{MonsterID: "tomb_hexer", Tier: "elite", Role: "controller", Count: 1}, {MonsterID: "grave_rat_swarm", Tier: "normal", Role: "assassin", Count: 2}, {MonsterID: "ashen_skull_caster", Tier: "normal", Role: "caster", Count: 1}, {MonsterID: "catacomb_boneguard", Tier: "normal", Role: "bruiser", Count: 1}}},
			5: {slots: []roomMonsterSlot{{MonsterID: "warden_of_seals", Tier: "elite", Role: "tank", Count: 1}, {MonsterID: "tomb_hexer", Tier: "elite", Role: "controller", Count: 1}, {MonsterID: "ashen_skull_caster", Tier: "normal", Role: "caster", Count: 1}, {MonsterID: "catacomb_boneguard", Tier: "normal", Role: "bruiser", Count: 1}, {MonsterID: "grave_rat_swarm", Tier: "normal", Role: "assassin", Count: 1}}},
			6: {slots: []roomMonsterSlot{{MonsterID: "morthis_chapel_keeper", Tier: "boss", Role: "boss", Count: 1}, {MonsterID: "warden_of_seals", Tier: "elite", Role: "tank", Count: 1}, {MonsterID: "tomb_hexer", Tier: "elite", Role: "controller", Count: 1}, {MonsterID: "grave_rat_swarm", Tier: "normal", Role: "assassin", Count: 1}}},
		},
	},
	"thorned_hollow_v1": {
		"easy": {
			1: {slots: []roomMonsterSlot{{MonsterID: "thornclaw_stalker", Tier: "normal", Role: "assassin", Count: 1}, {MonsterID: "briar_hexcaller", Tier: "normal", Role: "caster", Count: 1}}},
			2: {slots: []roomMonsterSlot{{MonsterID: "thornclaw_stalker", Tier: "normal", Role: "assassin", Count: 1}, {MonsterID: "hollow_trapper", Tier: "normal", Role: "controller", Count: 1}, {MonsterID: "briar_hexcaller", Tier: "normal", Role: "caster", Count: 1}}},
			3: {slots: []roomMonsterSlot{{MonsterID: "needlemane_huntmaster", Tier: "elite", Role: "bruiser", Count: 1}, {MonsterID: "thornclaw_stalker", Tier: "normal", Role: "assassin", Count: 1}}},
			4: {slots: []roomMonsterSlot{{MonsterID: "thornveil_predator", Tier: "elite", Role: "assassin", Count: 1}, {MonsterID: "hollow_trapper", Tier: "normal", Role: "controller", Count: 1}, {MonsterID: "briar_hexcaller", Tier: "normal", Role: "caster", Count: 1}}},
			5: {slots: []roomMonsterSlot{{MonsterID: "needlemane_huntmaster", Tier: "elite", Role: "bruiser", Count: 1}, {MonsterID: "thornveil_predator", Tier: "elite", Role: "assassin", Count: 1}, {MonsterID: "thornclaw_stalker", Tier: "normal", Role: "assassin", Count: 1}}},
			6: {slots: []roomMonsterSlot{{MonsterID: "briarqueen_ravastra", Tier: "boss", Role: "boss", Count: 1}, {MonsterID: "thornclaw_stalker", Tier: "normal", Role: "assassin", Count: 1}}},
		},
		"hard": {
			1: {slots: []roomMonsterSlot{{MonsterID: "thornclaw_stalker", Tier: "normal", Role: "assassin", Count: 2}, {MonsterID: "briar_hexcaller", Tier: "normal", Role: "caster", Count: 1}}},
			2: {slots: []roomMonsterSlot{{MonsterID: "thornclaw_stalker", Tier: "normal", Role: "assassin", Count: 1}, {MonsterID: "hollow_trapper", Tier: "normal", Role: "controller", Count: 1}, {MonsterID: "briar_hexcaller", Tier: "normal", Role: "caster", Count: 1}, {MonsterID: "thornclaw_stalker", Tier: "normal", Role: "assassin", Count: 1}}},
			3: {slots: []roomMonsterSlot{{MonsterID: "needlemane_huntmaster", Tier: "elite", Role: "bruiser", Count: 1}, {MonsterID: "thornclaw_stalker", Tier: "normal", Role: "assassin", Count: 1}, {MonsterID: "briar_hexcaller", Tier: "normal", Role: "caster", Count: 1}}},
			4: {slots: []roomMonsterSlot{{MonsterID: "thornveil_predator", Tier: "elite", Role: "assassin", Count: 1}, {MonsterID: "hollow_trapper", Tier: "normal", Role: "controller", Count: 1}, {MonsterID: "needlemane_huntmaster", Tier: "elite", Role: "bruiser", Count: 1}}},
			5: {slots: []roomMonsterSlot{{MonsterID: "needlemane_huntmaster", Tier: "elite", Role: "bruiser", Count: 1}, {MonsterID: "thornveil_predator", Tier: "elite", Role: "assassin", Count: 1}, {MonsterID: "hollow_trapper", Tier: "normal", Role: "controller", Count: 1}, {MonsterID: "briar_hexcaller", Tier: "normal", Role: "caster", Count: 1}}},
			6: {slots: []roomMonsterSlot{{MonsterID: "briarqueen_ravastra", Tier: "boss", Role: "boss", Count: 1}, {MonsterID: "thornveil_predator", Tier: "elite", Role: "assassin", Count: 1}, {MonsterID: "briar_hexcaller", Tier: "normal", Role: "caster", Count: 1}}},
		},
		"nightmare": {
			1: {slots: []roomMonsterSlot{{MonsterID: "thornclaw_stalker", Tier: "normal", Role: "assassin", Count: 2}, {MonsterID: "briar_hexcaller", Tier: "normal", Role: "caster", Count: 1}, {MonsterID: "hollow_trapper", Tier: "normal", Role: "controller", Count: 1}}},
			2: {slots: []roomMonsterSlot{{MonsterID: "needlemane_huntmaster", Tier: "elite", Role: "bruiser", Count: 1}, {MonsterID: "thornclaw_stalker", Tier: "normal", Role: "assassin", Count: 2}, {MonsterID: "briar_hexcaller", Tier: "normal", Role: "caster", Count: 1}}},
			3: {slots: []roomMonsterSlot{{MonsterID: "thornveil_predator", Tier: "elite", Role: "assassin", Count: 1}, {MonsterID: "hollow_trapper", Tier: "normal", Role: "controller", Count: 1}, {MonsterID: "briar_hexcaller", Tier: "normal", Role: "caster", Count: 1}, {MonsterID: "thornclaw_stalker", Tier: "normal", Role: "assassin", Count: 1}}},
			4: {slots: []roomMonsterSlot{{MonsterID: "needlemane_huntmaster", Tier: "elite", Role: "bruiser", Count: 1}, {MonsterID: "thornveil_predator", Tier: "elite", Role: "assassin", Count: 1}, {MonsterID: "hollow_trapper", Tier: "normal", Role: "controller", Count: 1}, {MonsterID: "thornclaw_stalker", Tier: "normal", Role: "assassin", Count: 1}}},
			5: {slots: []roomMonsterSlot{{MonsterID: "needlemane_huntmaster", Tier: "elite", Role: "bruiser", Count: 1}, {MonsterID: "thornveil_predator", Tier: "elite", Role: "assassin", Count: 1}, {MonsterID: "briar_hexcaller", Tier: "normal", Role: "caster", Count: 1}, {MonsterID: "hollow_trapper", Tier: "normal", Role: "controller", Count: 1}, {MonsterID: "thornclaw_stalker", Tier: "normal", Role: "assassin", Count: 1}}},
			6: {slots: []roomMonsterSlot{{MonsterID: "briarqueen_ravastra", Tier: "boss", Role: "boss", Count: 1}, {MonsterID: "thornveil_predator", Tier: "elite", Role: "assassin", Count: 1}, {MonsterID: "needlemane_huntmaster", Tier: "elite", Role: "bruiser", Count: 1}}},
		},
	},
	"sunscar_warvault_v1": {
		"easy": {
			1: {slots: []roomMonsterSlot{{MonsterID: "warvault_legionary", Tier: "normal", Role: "bruiser", Count: 1}, {MonsterID: "sunforge_channeler", Tier: "normal", Role: "caster", Count: 1}}},
			2: {slots: []roomMonsterSlot{{MonsterID: "warvault_legionary", Tier: "normal", Role: "bruiser", Count: 1}, {MonsterID: "shieldline_sentinel", Tier: "normal", Role: "tank", Count: 1}, {MonsterID: "sunforge_channeler", Tier: "normal", Role: "caster", Count: 1}}},
			3: {slots: []roomMonsterSlot{{MonsterID: "vault_breaker", Tier: "elite", Role: "bruiser", Count: 1}, {MonsterID: "warvault_legionary", Tier: "normal", Role: "bruiser", Count: 1}}},
			4: {slots: []roomMonsterSlot{{MonsterID: "standard_bearer", Tier: "elite", Role: "controller", Count: 1}, {MonsterID: "shieldline_sentinel", Tier: "normal", Role: "tank", Count: 1}, {MonsterID: "sunforge_channeler", Tier: "normal", Role: "caster", Count: 1}}},
			5: {slots: []roomMonsterSlot{{MonsterID: "vault_breaker", Tier: "elite", Role: "bruiser", Count: 1}, {MonsterID: "standard_bearer", Tier: "elite", Role: "controller", Count: 1}, {MonsterID: "warvault_legionary", Tier: "normal", Role: "bruiser", Count: 1}}},
			6: {slots: []roomMonsterSlot{{MonsterID: "sunscar_warmarshal", Tier: "boss", Role: "boss", Count: 1}, {MonsterID: "shieldline_sentinel", Tier: "normal", Role: "tank", Count: 1}}},
		},
		"hard": {
			1: {slots: []roomMonsterSlot{{MonsterID: "warvault_legionary", Tier: "normal", Role: "bruiser", Count: 2}, {MonsterID: "sunforge_channeler", Tier: "normal", Role: "caster", Count: 1}}},
			2: {slots: []roomMonsterSlot{{MonsterID: "warvault_legionary", Tier: "normal", Role: "bruiser", Count: 1}, {MonsterID: "shieldline_sentinel", Tier: "normal", Role: "tank", Count: 2}, {MonsterID: "sunforge_channeler", Tier: "normal", Role: "caster", Count: 1}}},
			3: {slots: []roomMonsterSlot{{MonsterID: "vault_breaker", Tier: "elite", Role: "bruiser", Count: 1}, {MonsterID: "warvault_legionary", Tier: "normal", Role: "bruiser", Count: 1}, {MonsterID: "sunforge_channeler", Tier: "normal", Role: "caster", Count: 1}}},
			4: {slots: []roomMonsterSlot{{MonsterID: "standard_bearer", Tier: "elite", Role: "controller", Count: 1}, {MonsterID: "vault_breaker", Tier: "elite", Role: "bruiser", Count: 1}, {MonsterID: "shieldline_sentinel", Tier: "normal", Role: "tank", Count: 1}}},
			5: {slots: []roomMonsterSlot{{MonsterID: "vault_breaker", Tier: "elite", Role: "bruiser", Count: 1}, {MonsterID: "standard_bearer", Tier: "elite", Role: "controller", Count: 1}, {MonsterID: "warvault_legionary", Tier: "normal", Role: "bruiser", Count: 1}, {MonsterID: "sunforge_channeler", Tier: "normal", Role: "caster", Count: 1}}},
			6: {slots: []roomMonsterSlot{{MonsterID: "sunscar_warmarshal", Tier: "boss", Role: "boss", Count: 1}, {MonsterID: "vault_breaker", Tier: "elite", Role: "bruiser", Count: 1}, {MonsterID: "standard_bearer", Tier: "elite", Role: "controller", Count: 1}}},
		},
		"nightmare": {
			1: {slots: []roomMonsterSlot{{MonsterID: "warvault_legionary", Tier: "normal", Role: "bruiser", Count: 2}, {MonsterID: "shieldline_sentinel", Tier: "normal", Role: "tank", Count: 1}, {MonsterID: "sunforge_channeler", Tier: "normal", Role: "caster", Count: 1}}},
			2: {slots: []roomMonsterSlot{{MonsterID: "vault_breaker", Tier: "elite", Role: "bruiser", Count: 1}, {MonsterID: "shieldline_sentinel", Tier: "normal", Role: "tank", Count: 2}, {MonsterID: "sunforge_channeler", Tier: "normal", Role: "caster", Count: 1}}},
			3: {slots: []roomMonsterSlot{{MonsterID: "vault_breaker", Tier: "elite", Role: "bruiser", Count: 1}, {MonsterID: "standard_bearer", Tier: "elite", Role: "controller", Count: 1}, {MonsterID: "warvault_legionary", Tier: "normal", Role: "bruiser", Count: 1}, {MonsterID: "sunforge_channeler", Tier: "normal", Role: "caster", Count: 1}}},
			4: {slots: []roomMonsterSlot{{MonsterID: "standard_bearer", Tier: "elite", Role: "controller", Count: 1}, {MonsterID: "vault_breaker", Tier: "elite", Role: "bruiser", Count: 1}, {MonsterID: "shieldline_sentinel", Tier: "normal", Role: "tank", Count: 1}, {MonsterID: "warvault_legionary", Tier: "normal", Role: "bruiser", Count: 1}}},
			5: {slots: []roomMonsterSlot{{MonsterID: "vault_breaker", Tier: "elite", Role: "bruiser", Count: 1}, {MonsterID: "standard_bearer", Tier: "elite", Role: "controller", Count: 1}, {MonsterID: "sunforge_channeler", Tier: "normal", Role: "caster", Count: 1}, {MonsterID: "shieldline_sentinel", Tier: "normal", Role: "tank", Count: 1}, {MonsterID: "warvault_legionary", Tier: "normal", Role: "bruiser", Count: 1}}},
			6: {slots: []roomMonsterSlot{{MonsterID: "sunscar_warmarshal", Tier: "boss", Role: "boss", Count: 1}, {MonsterID: "vault_breaker", Tier: "elite", Role: "bruiser", Count: 1}, {MonsterID: "standard_bearer", Tier: "elite", Role: "controller", Count: 1}}},
		},
	},
	"obsidian_spire_v1": {
		"easy": {
			1: {slots: []roomMonsterSlot{{MonsterID: "blackglass_sentinel", Tier: "normal", Role: "tank", Count: 1}, {MonsterID: "void_adept", Tier: "normal", Role: "caster", Count: 1}}},
			2: {slots: []roomMonsterSlot{{MonsterID: "blackglass_sentinel", Tier: "normal", Role: "tank", Count: 1}, {MonsterID: "eclipse_scribe", Tier: "normal", Role: "controller", Count: 1}, {MonsterID: "void_adept", Tier: "normal", Role: "caster", Count: 1}}},
			3: {slots: []roomMonsterSlot{{MonsterID: "obsidian_lancer", Tier: "elite", Role: "bruiser", Count: 1}, {MonsterID: "void_adept", Tier: "normal", Role: "caster", Count: 1}}},
			4: {slots: []roomMonsterSlot{{MonsterID: "mirrorbound_rector", Tier: "elite", Role: "caster", Count: 1}, {MonsterID: "eclipse_scribe", Tier: "normal", Role: "controller", Count: 1}, {MonsterID: "blackglass_sentinel", Tier: "normal", Role: "tank", Count: 1}}},
			5: {slots: []roomMonsterSlot{{MonsterID: "obsidian_lancer", Tier: "elite", Role: "bruiser", Count: 1}, {MonsterID: "mirrorbound_rector", Tier: "elite", Role: "caster", Count: 1}, {MonsterID: "void_adept", Tier: "normal", Role: "caster", Count: 1}}},
			6: {slots: []roomMonsterSlot{{MonsterID: "obsidian_archon", Tier: "boss", Role: "boss", Count: 1}, {MonsterID: "blackglass_sentinel", Tier: "normal", Role: "tank", Count: 1}}},
		},
		"hard": {
			1: {slots: []roomMonsterSlot{{MonsterID: "blackglass_sentinel", Tier: "normal", Role: "tank", Count: 2}, {MonsterID: "void_adept", Tier: "normal", Role: "caster", Count: 1}}},
			2: {slots: []roomMonsterSlot{{MonsterID: "blackglass_sentinel", Tier: "normal", Role: "tank", Count: 1}, {MonsterID: "eclipse_scribe", Tier: "normal", Role: "controller", Count: 1}, {MonsterID: "void_adept", Tier: "normal", Role: "caster", Count: 2}}},
			3: {slots: []roomMonsterSlot{{MonsterID: "obsidian_lancer", Tier: "elite", Role: "bruiser", Count: 1}, {MonsterID: "void_adept", Tier: "normal", Role: "caster", Count: 1}, {MonsterID: "blackglass_sentinel", Tier: "normal", Role: "tank", Count: 1}}},
			4: {slots: []roomMonsterSlot{{MonsterID: "mirrorbound_rector", Tier: "elite", Role: "caster", Count: 1}, {MonsterID: "eclipse_scribe", Tier: "normal", Role: "controller", Count: 1}, {MonsterID: "obsidian_lancer", Tier: "elite", Role: "bruiser", Count: 1}}},
			5: {slots: []roomMonsterSlot{{MonsterID: "obsidian_lancer", Tier: "elite", Role: "bruiser", Count: 1}, {MonsterID: "mirrorbound_rector", Tier: "elite", Role: "caster", Count: 1}, {MonsterID: "void_adept", Tier: "normal", Role: "caster", Count: 1}, {MonsterID: "eclipse_scribe", Tier: "normal", Role: "controller", Count: 1}}},
			6: {slots: []roomMonsterSlot{{MonsterID: "obsidian_archon", Tier: "boss", Role: "boss", Count: 1}, {MonsterID: "mirrorbound_rector", Tier: "elite", Role: "caster", Count: 1}, {MonsterID: "blackglass_sentinel", Tier: "normal", Role: "tank", Count: 1}}},
		},
		"nightmare": {
			1: {slots: []roomMonsterSlot{{MonsterID: "blackglass_sentinel", Tier: "normal", Role: "tank", Count: 2}, {MonsterID: "void_adept", Tier: "normal", Role: "caster", Count: 1}, {MonsterID: "eclipse_scribe", Tier: "normal", Role: "controller", Count: 1}}},
			2: {slots: []roomMonsterSlot{{MonsterID: "obsidian_lancer", Tier: "elite", Role: "bruiser", Count: 1}, {MonsterID: "void_adept", Tier: "normal", Role: "caster", Count: 2}, {MonsterID: "blackglass_sentinel", Tier: "normal", Role: "tank", Count: 1}}},
			3: {slots: []roomMonsterSlot{{MonsterID: "mirrorbound_rector", Tier: "elite", Role: "caster", Count: 1}, {MonsterID: "eclipse_scribe", Tier: "normal", Role: "controller", Count: 1}, {MonsterID: "void_adept", Tier: "normal", Role: "caster", Count: 1}, {MonsterID: "blackglass_sentinel", Tier: "normal", Role: "tank", Count: 1}}},
			4: {slots: []roomMonsterSlot{{MonsterID: "obsidian_lancer", Tier: "elite", Role: "bruiser", Count: 1}, {MonsterID: "mirrorbound_rector", Tier: "elite", Role: "caster", Count: 1}, {MonsterID: "eclipse_scribe", Tier: "normal", Role: "controller", Count: 1}, {MonsterID: "void_adept", Tier: "normal", Role: "caster", Count: 1}}},
			5: {slots: []roomMonsterSlot{{MonsterID: "obsidian_lancer", Tier: "elite", Role: "bruiser", Count: 1}, {MonsterID: "mirrorbound_rector", Tier: "elite", Role: "caster", Count: 1}, {MonsterID: "void_adept", Tier: "normal", Role: "caster", Count: 1}, {MonsterID: "eclipse_scribe", Tier: "normal", Role: "controller", Count: 1}, {MonsterID: "blackglass_sentinel", Tier: "normal", Role: "tank", Count: 1}}},
			6: {slots: []roomMonsterSlot{{MonsterID: "obsidian_archon", Tier: "boss", Role: "boss", Count: 1}, {MonsterID: "mirrorbound_rector", Tier: "elite", Role: "caster", Count: 1}, {MonsterID: "obsidian_lancer", Tier: "elite", Role: "bruiser", Count: 1}}},
		},
	},
	"sandworm_den_v1": {
		"easy": {
			1: {slots: []roomMonsterSlot{{MonsterID: "sand_burrower", Tier: "normal", Role: "bruiser", Count: 1}, {MonsterID: "scorched_spitter", Tier: "normal", Role: "caster", Count: 1}}},
			2: {slots: []roomMonsterSlot{{MonsterID: "dune_skitterer", Tier: "normal", Role: "assassin", Count: 1}, {MonsterID: "sand_burrower", Tier: "normal", Role: "bruiser", Count: 1}, {MonsterID: "scorched_spitter", Tier: "normal", Role: "caster", Count: 1}}},
			3: {slots: []roomMonsterSlot{{MonsterID: "carapace_crusher", Tier: "elite", Role: "bruiser", Count: 1}, {MonsterID: "sand_burrower", Tier: "normal", Role: "bruiser", Count: 1}}},
			4: {slots: []roomMonsterSlot{{MonsterID: "venom_herald", Tier: "elite", Role: "controller", Count: 1}, {MonsterID: "scorched_spitter", Tier: "normal", Role: "caster", Count: 1}, {MonsterID: "dune_skitterer", Tier: "normal", Role: "assassin", Count: 1}}},
			5: {slots: []roomMonsterSlot{{MonsterID: "carapace_crusher", Tier: "elite", Role: "bruiser", Count: 1}, {MonsterID: "venom_herald", Tier: "elite", Role: "controller", Count: 1}, {MonsterID: "sand_burrower", Tier: "normal", Role: "bruiser", Count: 1}}},
			6: {slots: []roomMonsterSlot{{MonsterID: "kharzug_dunescourge_matriarch", Tier: "boss", Role: "boss", Count: 1}, {MonsterID: "sand_burrower", Tier: "normal", Role: "bruiser", Count: 1}}},
		},
		"hard": {
			1: {slots: []roomMonsterSlot{{MonsterID: "sand_burrower", Tier: "normal", Role: "bruiser", Count: 2}, {MonsterID: "scorched_spitter", Tier: "normal", Role: "caster", Count: 1}}},
			2: {slots: []roomMonsterSlot{{MonsterID: "dune_skitterer", Tier: "normal", Role: "assassin", Count: 1}, {MonsterID: "sand_burrower", Tier: "normal", Role: "bruiser", Count: 2}, {MonsterID: "scorched_spitter", Tier: "normal", Role: "caster", Count: 1}}},
			3: {slots: []roomMonsterSlot{{MonsterID: "carapace_crusher", Tier: "elite", Role: "bruiser", Count: 1}, {MonsterID: "sand_burrower", Tier: "normal", Role: "bruiser", Count: 1}, {MonsterID: "scorched_spitter", Tier: "normal", Role: "caster", Count: 1}}},
			4: {slots: []roomMonsterSlot{{MonsterID: "venom_herald", Tier: "elite", Role: "controller", Count: 1}, {MonsterID: "carapace_crusher", Tier: "elite", Role: "bruiser", Count: 1}, {MonsterID: "dune_skitterer", Tier: "normal", Role: "assassin", Count: 1}, {MonsterID: "scorched_spitter", Tier: "normal", Role: "caster", Count: 1}}},
			5: {slots: []roomMonsterSlot{{MonsterID: "venom_herald", Tier: "elite", Role: "controller", Count: 1}, {MonsterID: "carapace_crusher", Tier: "elite", Role: "bruiser", Count: 1}, {MonsterID: "sand_burrower", Tier: "normal", Role: "bruiser", Count: 1}, {MonsterID: "dune_skitterer", Tier: "normal", Role: "assassin", Count: 1}}},
			6: {slots: []roomMonsterSlot{{MonsterID: "kharzug_dunescourge_matriarch", Tier: "boss", Role: "boss", Count: 1}, {MonsterID: "carapace_crusher", Tier: "elite", Role: "bruiser", Count: 1}, {MonsterID: "venom_herald", Tier: "elite", Role: "controller", Count: 1}}},
		},
		"nightmare": {
			1: {slots: []roomMonsterSlot{{MonsterID: "sand_burrower", Tier: "normal", Role: "bruiser", Count: 2}, {MonsterID: "scorched_spitter", Tier: "normal", Role: "caster", Count: 1}, {MonsterID: "dune_skitterer", Tier: "normal", Role: "assassin", Count: 1}}},
			2: {slots: []roomMonsterSlot{{MonsterID: "carapace_crusher", Tier: "elite", Role: "bruiser", Count: 1}, {MonsterID: "sand_burrower", Tier: "normal", Role: "bruiser", Count: 1}, {MonsterID: "scorched_spitter", Tier: "normal", Role: "caster", Count: 1}, {MonsterID: "dune_skitterer", Tier: "normal", Role: "assassin", Count: 1}}},
			3: {slots: []roomMonsterSlot{{MonsterID: "carapace_crusher", Tier: "elite", Role: "bruiser", Count: 1}, {MonsterID: "venom_herald", Tier: "elite", Role: "controller", Count: 1}, {MonsterID: "sand_burrower", Tier: "normal", Role: "bruiser", Count: 1}, {MonsterID: "scorched_spitter", Tier: "normal", Role: "caster", Count: 1}}},
			4: {slots: []roomMonsterSlot{{MonsterID: "venom_herald", Tier: "elite", Role: "controller", Count: 1}, {MonsterID: "carapace_crusher", Tier: "elite", Role: "bruiser", Count: 1}, {MonsterID: "dune_skitterer", Tier: "normal", Role: "assassin", Count: 2}, {MonsterID: "scorched_spitter", Tier: "normal", Role: "caster", Count: 1}}},
			5: {slots: []roomMonsterSlot{{MonsterID: "venom_herald", Tier: "elite", Role: "controller", Count: 1}, {MonsterID: "carapace_crusher", Tier: "elite", Role: "bruiser", Count: 1}, {MonsterID: "sand_burrower", Tier: "normal", Role: "bruiser", Count: 1}, {MonsterID: "scorched_spitter", Tier: "normal", Role: "caster", Count: 1}, {MonsterID: "dune_skitterer", Tier: "normal", Role: "assassin", Count: 1}}},
			6: {slots: []roomMonsterSlot{{MonsterID: "kharzug_dunescourge_matriarch", Tier: "boss", Role: "boss", Count: 1}, {MonsterID: "carapace_crusher", Tier: "elite", Role: "bruiser", Count: 1}, {MonsterID: "venom_herald", Tier: "elite", Role: "controller", Count: 1}, {MonsterID: "sandworm_larva", Tier: "normal", Role: "summoner", Count: 2}}},
		},
	},
}

// stagedMaterialDropsFromClear returns material drops for a cleared dungeon run.
// Per doc 12 §7: higher difficulty increases the weight of rare materials;
// nightmare improves access to red and prismatic material resources.
func stagedMaterialDropsFromClear(highestRoomCleared int, difficulty string) []map[string]any {
	if highestRoomCleared <= 0 {
		return []map[string]any{}
	}
	baseQty := maxInt(1, highestRoomCleared/2+1)
	drops := []map[string]any{{
		"material_key": "dungeon_essence",
		"quantity":     baseQty,
	}}
	switch normalizeDifficulty(difficulty) {
	case "hard":
		// hard adds rare-tier fragments at reduced quantity
		drops = append(drops, map[string]any{
			"material_key": "dungeon_fragment",
			"quantity":     maxInt(1, highestRoomCleared/3),
		})
	case "nightmare":
		// nightmare adds both rare fragments and high-tier cores
		drops = append(drops, map[string]any{
			"material_key": "dungeon_fragment",
			"quantity":     maxInt(1, highestRoomCleared/2),
		})
		if highestRoomCleared >= 4 {
			drops = append(drops, map[string]any{
				"material_key": "dungeon_core",
				"quantity":     maxInt(1, (highestRoomCleared-3)/2),
			})
		}
	}
	return drops
}

// difficultyGoldMultiplier returns a gold bonus multiplier for harder difficulties.
// hard gives +20%, nightmare gives +50%, aligned with the difficulty risk increase.
func difficultyGoldMultiplier(difficulty string) float64 {
	switch normalizeDifficulty(difficulty) {
	case "hard":
		return 1.20
	case "nightmare":
		return 1.50
	default:
		return 1.00
	}
}

func scaledRewardGold(baseGoldMin, baseGoldMax, highestRoomCleared, totalRooms int, difficulty string, seed string) int {
	if totalRooms <= 0 {
		return baseGoldMin
	}
	baseGold := deterministicGoldInRange(seed, baseGoldMin, baseGoldMax)
	cleared := maxInt(0, minInt(highestRoomCleared, totalRooms))
	if cleared == 0 {
		return maxInt(1, baseGold/5)
	}
	ratio := float64(cleared) / float64(totalRooms)
	base := float64(baseGold) * (0.45 + ratio*0.55)
	return maxInt(1, int(base*difficultyGoldMultiplier(difficulty)))
}

func deterministicGoldInRange(seed string, minGold, maxGold int) int {
	if maxGold <= minGold {
		return minGold
	}
	return minGold + deterministicIndex(seed, maxGold-minGold+1)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
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
		"blue":      expandDungeonRewardPool([]string{"gravewake_bastion_boots_blue", "gravewake_bastion_helm_blue"}, "gravewake_bastion", "blue"),
		"purple":    expandDungeonRewardPool([]string{"gravewake_bastion_ring_purple", "gravewake_bastion_necklace_purple"}, "gravewake_bastion", "purple"),
		"gold":      expandDungeonRewardPool([]string{"gravewake_bastion_chest_gold", "gravewake_bastion_greaves_gold"}, "gravewake_bastion", "gold"),
		"red":       expandDungeonRewardPool([]string{"gravewake_bastion_chest_red", "gravewake_bastion_ring_red"}, "gravewake_bastion", "red"),
		"prismatic": expandDungeonRewardPool([]string{"gravewake_bastion_necklace_prismatic"}, "gravewake_bastion", "prismatic"),
	},
	"thorned_hollow_v1": {
		"blue":      expandDungeonRewardPool([]string{"briarbound_sight_boots_blue", "briarbound_sight_hood_blue"}, "briarbound_sight", "blue"),
		"purple":    expandDungeonRewardPool([]string{"briarbound_sight_ring_purple", "briarbound_sight_necklace_purple"}, "briarbound_sight", "purple"),
		"gold":      expandDungeonRewardPool([]string{"briarbound_sight_chest_gold", "briarbound_sight_boots_gold"}, "briarbound_sight", "gold"),
		"red":       expandDungeonRewardPool([]string{"briarbound_sight_chest_red", "briarbound_sight_ring_red"}, "briarbound_sight", "red"),
		"prismatic": expandDungeonRewardPool([]string{"briarbound_sight_necklace_prismatic"}, "briarbound_sight", "prismatic"),
	},
	"sunscar_warvault_v1": {
		"blue":      expandDungeonRewardPool([]string{"sunscar_assault_boots_blue", "sunscar_assault_helm_blue"}, "sunscar_assault", "blue"),
		"purple":    expandDungeonRewardPool([]string{"sunscar_assault_ring_purple", "sunscar_assault_necklace_purple"}, "sunscar_assault", "purple"),
		"gold":      expandDungeonRewardPool([]string{"sunscar_assault_chest_gold", "sunscar_assault_boots_gold"}, "sunscar_assault", "gold"),
		"red":       expandDungeonRewardPool([]string{"sunscar_assault_chest_red", "sunscar_assault_ring_red"}, "sunscar_assault", "red"),
		"prismatic": expandDungeonRewardPool([]string{"sunscar_assault_necklace_prismatic"}, "sunscar_assault", "prismatic"),
	},
	"obsidian_spire_v1": {
		"blue":      expandDungeonRewardPool([]string{"nightglass_arcanum_boots_blue", "nightglass_arcanum_hood_blue"}, "nightglass_arcanum", "blue"),
		"purple":    expandDungeonRewardPool([]string{"nightglass_arcanum_ring_purple", "nightglass_arcanum_necklace_purple"}, "nightglass_arcanum", "purple"),
		"gold":      expandDungeonRewardPool([]string{"nightglass_arcanum_chest_gold", "nightglass_arcanum_boots_gold"}, "nightglass_arcanum", "gold"),
		"red":       expandDungeonRewardPool([]string{"nightglass_arcanum_chest_red", "nightglass_arcanum_ring_red"}, "nightglass_arcanum", "red"),
		"prismatic": expandDungeonRewardPool([]string{"nightglass_arcanum_necklace_prismatic"}, "nightglass_arcanum", "prismatic"),
	},
	"sandworm_den_v1": {
		"blue":      {"dunescourge_burrowstep_blue", "dunescourge_coilguards_blue"},
		"purple":    {"dunescourge_fang_ring_purple", "dunescourge_crownshell_purple"},
		"gold":      {"dunescourge_heartspine_chain_gold", "dunescourge_carapace_mail_gold"},
		"red":       {"dunescourge_carapace_mail_red", "dunescourge_fang_ring_red"},
		"prismatic": {"dunescourge_heartspine_chain_prismatic"},
	},
}

const (
	sharedDungeonGoldMin = 265
	sharedDungeonGoldMax = 370
)

func expandDungeonRewardPool(base []string, setPrefix, rarity string) []string {
	pool := make([]string, 0, len(base)+6)
	pool = append(pool, base...)
	for _, style := range []string{"sword_shield", "great_axe", "staff", "spellbook", "scepter", "holy_tome"} {
		pool = append(pool, fmt.Sprintf("%s_weapon_%s_%s", setPrefix, style, rarity))
	}
	return pool
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

var dungeonDefinitions = map[string]DefinitionView{
	"ancient_catacomb_v1": {
		DungeonID:           "ancient_catacomb_v1",
		Name:                "Ancient Catacomb",
		RegionID:            "ancient_catacomb",
		RoomCount:           6,
		BossRoomIndex:       6,
		RecommendedLevelMin: 1,
		RecommendedLevelMax: 30,
		IsNovice:            true,
		RatingRules: []DungeonRatingRule{
			{HighestRoomCleared: 1, Rating: "E"},
			{HighestRoomCleared: 2, Rating: "D"},
			{HighestRoomCleared: 3, Rating: "C"},
			{HighestRoomCleared: 4, Rating: "B"},
			{HighestRoomCleared: 5, Rating: "A"},
			{HighestRoomCleared: 6, Rating: "S"},
		},
		RewardSummary: map[string]any{
			"gold_min":   sharedDungeonGoldMin,
			"gold_max":   sharedDungeonGoldMax,
			"drop_table": "catacomb_v1",
		},
	},
	"thorned_hollow_v1": {
		DungeonID:           "thorned_hollow_v1",
		Name:                "Thorned Hollow",
		RegionID:            "thorned_hollow",
		RoomCount:           6,
		BossRoomIndex:       6,
		RecommendedLevelMin: 26,
		RecommendedLevelMax: 55,
		RatingRules: []DungeonRatingRule{
			{HighestRoomCleared: 1, Rating: "E"},
			{HighestRoomCleared: 2, Rating: "D"},
			{HighestRoomCleared: 3, Rating: "C"},
			{HighestRoomCleared: 4, Rating: "B"},
			{HighestRoomCleared: 5, Rating: "A"},
			{HighestRoomCleared: 6, Rating: "S"},
		},
		RewardSummary: map[string]any{
			"gold_min":   sharedDungeonGoldMin,
			"gold_max":   sharedDungeonGoldMax,
			"drop_table": "thorned_hollow_v1",
		},
	},
	"sunscar_warvault_v1": {
		DungeonID:           "sunscar_warvault_v1",
		Name:                "Sunscar Warvault",
		RegionID:            "sunscar_warvault",
		RoomCount:           6,
		BossRoomIndex:       6,
		RecommendedLevelMin: 56,
		RecommendedLevelMax: 82,
		RatingRules: []DungeonRatingRule{
			{HighestRoomCleared: 1, Rating: "E"},
			{HighestRoomCleared: 2, Rating: "D"},
			{HighestRoomCleared: 3, Rating: "C"},
			{HighestRoomCleared: 4, Rating: "B"},
			{HighestRoomCleared: 5, Rating: "A"},
			{HighestRoomCleared: 6, Rating: "S"},
		},
		RewardSummary: map[string]any{
			"gold_min":   sharedDungeonGoldMin,
			"gold_max":   sharedDungeonGoldMax,
			"drop_table": "sunscar_warvault_v1",
		},
	},
	"obsidian_spire_v1": {
		DungeonID:           "obsidian_spire_v1",
		Name:                "Obsidian Spire",
		RegionID:            "obsidian_spire",
		RoomCount:           6,
		BossRoomIndex:       6,
		RecommendedLevelMin: 68,
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
			"gold_min":   sharedDungeonGoldMin,
			"gold_max":   sharedDungeonGoldMax,
			"drop_table": "obsidian_spire_v1",
		},
	},
}
