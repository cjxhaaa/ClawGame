package world

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"clawgame/apps/api/internal/combat"
)

const businessTimezone = "Asia/Shanghai"

var (
	ErrFieldEncounterUnavailable = errors.New("field encounter unavailable")
	ErrFieldEncounterInvalidMode = errors.New("field encounter invalid mode")
)

type Region struct {
	ID             string `json:"region_id"`
	Name           string `json:"name"`
	Type           string `json:"type"`
	TravelCostGold int    `json:"travel_cost_gold"`
}

type Building struct {
	ID       string   `json:"building_id"`
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	Category string   `json:"category"`
	Actions  []string `json:"actions"`
}

type TravelOption struct {
	RegionID       string `json:"region_id"`
	Name           string `json:"name"`
	TravelCostGold int    `json:"travel_cost_gold"`
}

type EncounterSummary struct {
	ActivityType string   `json:"activity_type"`
	Summary      string   `json:"summary"`
	Highlights   []string `json:"highlights"`
}

type FieldEncounterResult struct {
	RegionID           string           `json:"region_id"`
	Approach           string           `json:"approach"`
	EncounterID        string           `json:"encounter_id"`
	EncounterFamily    string           `json:"encounter_family"`
	BattleType         string           `json:"battle_type"`
	EventType          string           `json:"event_type"`
	Summary            string           `json:"summary"`
	Victory            bool             `json:"victory"`
	RewardGold         int              `json:"reward_gold"`
	EnemiesDefeated    int              `json:"enemies_defeated"`
	MaterialsCollected int              `json:"materials_collected"`
	MaterialDrops      []map[string]any `json:"material_drops"`
	IsCurio            bool             `json:"is_curio"`
	CurioLabel         string           `json:"curio_label,omitempty"`
	CurioOutcome       string           `json:"curio_outcome,omitempty"`
	FollowupQuest      *CurioQuestSeed  `json:"followup_quest,omitempty"`
	BattleState        map[string]any   `json:"battle_state"`
	BattleLog          []map[string]any `json:"battle_log"`
}

type CurioQuestSeed struct {
	TemplateType     string `json:"template_type"`
	Title            string `json:"title"`
	Description      string `json:"description"`
	TargetRegionID   string `json:"target_region_id"`
	RewardGold       int    `json:"reward_gold"`
	RewardReputation int    `json:"reward_reputation"`
}

type RegionGameplay struct {
	InteractionLayer       string   `json:"interaction_layer"`
	RiskLevel              string   `json:"risk_level"`
	FacilityFocus          string   `json:"facility_focus"`
	EncounterFamily        string   `json:"encounter_family"`
	CurioStatus            string   `json:"curio_status"`
	CurioHint              string   `json:"curio_hint,omitempty"`
	LinkedDungeon          string   `json:"linked_dungeon,omitempty"`
	ParentRegionID         string   `json:"parent_region_id,omitempty"`
	HostileEncounters      bool     `json:"hostile_encounters"`
	AvailableRegionActions []string `json:"available_region_actions,omitempty"`
}

type RegionDetail struct {
	Region           Region            `json:"region"`
	Description      string            `json:"description"`
	Buildings        []Building        `json:"buildings"`
	TravelOptions    []TravelOption    `json:"travel_options"`
	EncounterSummary *EncounterSummary `json:"encounter_summary,omitempty"`
	RegionGameplay
}

type RegionActivity struct {
	RegionID         string `json:"region_id"`
	Name             string `json:"name"`
	Type             string `json:"type"`
	TravelCostGold   int    `json:"travel_cost_gold"`
	Population       int    `json:"population"`
	RecentEventCount int    `json:"recent_event_count"`
	Highlight        string `json:"highlight"`
	BuildingCount    int    `json:"building_count"`
	RegionGameplay
}

type ArenaStatus struct {
	Code          string `json:"code"`
	Label         string `json:"label"`
	Details       string `json:"details"`
	NextMilestone string `json:"next_milestone"`
}

type PublicWorldState struct {
	ServerTime           string           `json:"server_time"`
	DailyResetAt         string           `json:"daily_reset_at"`
	ActiveBotCount       int              `json:"active_bot_count"`
	BotsInDungeonCount   int              `json:"bots_in_dungeon_count"`
	BotsInArenaCount     int              `json:"bots_in_arena_count"`
	QuestsCompletedToday int              `json:"quests_completed_today"`
	DungeonClearsToday   int              `json:"dungeon_clears_today"`
	GoldMintedToday      int              `json:"gold_minted_today"`
	Regions              []RegionActivity `json:"regions"`
	CurrentArenaStatus   ArenaStatus      `json:"current_arena_status"`
}

type WorldEvent struct {
	EventID          string         `json:"event_id"`
	EventType        string         `json:"event_type"`
	Visibility       string         `json:"visibility"`
	ActorCharacterID string         `json:"actor_character_id,omitempty"`
	ActorName        string         `json:"actor_name,omitempty"`
	RegionID         string         `json:"region_id,omitempty"`
	Summary          string         `json:"summary"`
	Payload          map[string]any `json:"payload"`
	OccurredAt       string         `json:"occurred_at"`
}

type LeaderboardEntry struct {
	Rank          int    `json:"rank"`
	CharacterID   string `json:"character_id"`
	Name          string `json:"name"`
	Class         string `json:"class"`
	WeaponStyle   string `json:"weapon_style"`
	RegionID      string `json:"region_id"`
	Score         int    `json:"score"`
	ScoreLabel    string `json:"score_label"`
	ActivityLabel string `json:"activity_label"`
}

type Leaderboards struct {
	Reputation    []LeaderboardEntry `json:"reputation"`
	Gold          []LeaderboardEntry `json:"gold"`
	WeeklyArena   []LeaderboardEntry `json:"weekly_arena"`
	DungeonClears []LeaderboardEntry `json:"dungeon_clears"`
}

type Service struct {
	clock func() time.Time
	loc   *time.Location
}

func NewService() *Service {
	return &Service{
		clock: time.Now,
		loc:   mustLocation(businessTimezone),
	}
}

func (s *Service) SetClock(clock func() time.Time) {
	if clock == nil {
		s.clock = time.Now
		return
	}
	s.clock = clock
}

func (s *Service) CurrentTime() time.Time {
	return s.clock().In(s.loc)
}

func (s *Service) NextDailyReset(now time.Time) time.Time {
	return nextDailyReset(now.In(s.loc))
}

func (s *Service) CurrentArenaStatus(now time.Time) ArenaStatus {
	return currentArenaStatus(now.In(s.loc))
}

func guildActions(base []string) []string {
	items := slices.Clone(base)
	for _, action := range []string{"view_skills", "upgrade_skill", "set_skill_loadout", "choose_profession_route"} {
		if !slices.Contains(items, action) {
			items = append(items, action)
		}
	}
	return items
}

func (s *Service) ListRegions() []Region {
	items := make([]Region, 0, len(seedRegions))
	for _, region := range seedRegions {
		items = append(items, region.Region)
	}

	return items
}

func (s *Service) GetRegion(regionID string) (RegionDetail, bool) {
	for _, region := range seedRegions {
		if region.Region.ID == regionID {
			region.RegionGameplay.AvailableRegionActions = regionAvailableActions(region)
			return region, true
		}
	}

	return RegionDetail{}, false
}

func (s *Service) GetPublicWorldState() PublicWorldState {
	now := s.clock().In(s.loc)
	resetAt := nextDailyReset(now)
	arenaStatus := currentArenaStatus(now)

	regions := make([]RegionActivity, 0, len(seedRegionActivity))
	for _, activity := range seedRegionActivity {
		region, ok := s.GetRegion(activity.RegionID)
		if !ok {
			continue
		}

		activity.Name = region.Region.Name
		activity.Type = region.Region.Type
		activity.TravelCostGold = region.Region.TravelCostGold
		activity.BuildingCount = len(region.Buildings)
		activity.RegionGameplay = region.RegionGameplay
		activity.RegionGameplay.AvailableRegionActions = regionAvailableActions(region)
		regions = append(regions, activity)
	}

	return PublicWorldState{
		ServerTime:           now.Format(time.RFC3339),
		DailyResetAt:         resetAt.Format(time.RFC3339),
		ActiveBotCount:       86,
		BotsInDungeonCount:   18,
		BotsInArenaCount:     arenaStatusArenaPopulation(arenaStatus.Code),
		QuestsCompletedToday: 142,
		DungeonClearsToday:   27,
		GoldMintedToday:      18420,
		Regions:              regions,
		CurrentArenaStatus:   arenaStatus,
	}
}

func (s *Service) ListPublicEvents(limit int) []WorldEvent {
	if limit <= 0 {
		limit = 20
	}
	if limit > len(seedEventTemplates) {
		limit = len(seedEventTemplates)
	}

	now := s.clock().In(s.loc)
	items := make([]WorldEvent, 0, limit)
	for index, template := range seedEventTemplates[:limit] {
		occurredAt := now.Add(-template.Offset)
		items = append(items, WorldEvent{
			EventID:          fmt.Sprintf("evt_public_%02d", index+1),
			EventType:        template.EventType,
			Visibility:       "public",
			ActorCharacterID: template.ActorCharacterID,
			ActorName:        template.ActorName,
			RegionID:         template.RegionID,
			Summary:          template.Summary,
			Payload:          template.Payload,
			OccurredAt:       occurredAt.Format(time.RFC3339),
		})
	}

	return items
}

func (s *Service) GetPublicLeaderboards() Leaderboards {
	return seedLeaderboards
}

func (s *Service) ResolveFieldEncounter(regionID, approach string, player combat.Combatant) (FieldEncounterResult, error) {
	detail, ok := s.GetRegion(regionID)
	if !ok || detail.Region.Type != "field" {
		return FieldEncounterResult{}, ErrFieldEncounterUnavailable
	}

	normalizedApproach := strings.ToLower(strings.TrimSpace(approach))
	if normalizedApproach == "" {
		normalizedApproach = "hunt"
	}

	template, err := buildFieldEncounter(detail, normalizedApproach)
	if err != nil {
		return FieldEncounterResult{}, err
	}

	encounterID := fmt.Sprintf("field_%s_%s_%d", regionID, normalizedApproach, s.clock().UnixNano())
	if player.CurrentHP <= 0 {
		player.CurrentHP = player.MaxHP
	}
	enemies := buildFieldEnemyCombatants(detail.Region.ID, normalizedApproach)
	if len(enemies) == 0 {
		return FieldEncounterResult{}, ErrFieldEncounterUnavailable
	}
	log := []map[string]any{{
		"step":         "field_composition",
		"event_type":   "field_composition",
		"battle_type":  "field_skirmish",
		"encounter_id": encounterID,
		"region_id":    regionID,
		"approach":     normalizedApproach,
		"player_hp":    player.CurrentHP,
		"monsters":     template.compositionLog(),
		"message":      "field enemy composition prepared",
	}}
	skirmish := combat.SimulateSkirmish(combat.SkirmishConfig{
		BattleType: "field_skirmish",
		RunID:      encounterID,
		Player:     player,
		Enemies:    enemies,
		MaxTurns:   12,
	})

	result := FieldEncounterResult{
		RegionID:        detail.Region.ID,
		Approach:        normalizedApproach,
		EncounterID:     encounterID,
		EncounterFamily: detail.EncounterFamily,
		BattleType:      "field_skirmish",
		Victory:         skirmish.PlayerWon,
		IsCurio:         template.IsCurio,
		CurioLabel:      template.CurioLabel,
		CurioOutcome:    template.CurioOutcome,
		BattleState: map[string]any{
			"engine_mode":       "auto_turn_based",
			"resolved_at":       s.clock().Format(time.RFC3339),
			"battle_type":       "field_skirmish",
			"encounter_id":      encounterID,
			"region_id":         detail.Region.ID,
			"approach":          normalizedApproach,
			"start_hp":          player.MaxHP,
			"remaining_hp":      skirmish.PlayerFinalHP,
			"turns_taken":       maxFieldTurn(skirmish.Log),
			"player_survived":   skirmish.PlayerWon,
			"enemies_total":     template.enemyCount(),
			"enemies_remaining": skirmish.EnemiesRemaining,
			"potions_consumed":  skirmish.PotionsConsumed,
		},
		BattleLog: append(log, skirmish.Log...),
	}

	defeated := template.enemyCount() - skirmish.EnemiesRemaining
	if defeated < 0 {
		defeated = 0
	}
	result.EnemiesDefeated = defeated

	switch approach {
	case "", "hunt", "gather", "curio":
		if result.Victory {
			result.EventType = template.SuccessEventType
			result.Summary = template.SuccessSummary
			result.RewardGold = template.RewardGold
			result.MaterialDrops = copyMapSlice(template.MaterialDrops)
			result.MaterialsCollected = sumMaterialDrops(result.MaterialDrops)
			if template.IsCurio && template.FollowupQuest != nil {
				quest := *template.FollowupQuest
				result.FollowupQuest = &quest
			}
			return result, nil
		}
		result.EventType = template.FailureEventType
		result.Summary = template.FailureSummary
		result.MaterialDrops = []map[string]any{}
		return result, nil
	default:
		return FieldEncounterResult{}, ErrFieldEncounterInvalidMode
	}
}

func mustLocation(name string) *time.Location {
	location, err := time.LoadLocation(name)
	if err != nil {
		panic(err)
	}

	return location
}

func buildFieldEncounter(detail RegionDetail, approach string) (fieldEncounterTemplate, error) {
	switch detail.Region.ID {
	case "briar_thicket":
		return buildBriarFieldEncounter(detail, approach), nil
	case "sunscar_desert_outskirts":
		return buildDesertFieldEncounter(detail, approach), nil
	case "ashen_ridge":
		return buildAshenFieldEncounter(detail, approach), nil
	default:
		return buildForestFieldEncounter(detail, approach), nil
	}
}

func regionAvailableActions(detail RegionDetail) []string {
	actions := make([]string, 0, 5)

	if hasFunctionalBuilding(detail.Buildings) {
		actions = append(actions, "enter_building")
	}

	if detail.Region.Type == "field" {
		actions = append(actions,
			"resolve_field_encounter:hunt",
			"resolve_field_encounter:gather",
			"resolve_field_encounter:curio",
		)
	}

	if detail.Region.Type == "dungeon" || strings.TrimSpace(detail.LinkedDungeon) != "" {
		actions = append(actions, "enter_dungeon")
	}

	return actions
}

func hasFunctionalBuilding(buildings []Building) bool {
	for _, building := range buildings {
		if building.Category == "functional_building" {
			return true
		}
	}

	return false
}

type fieldEnemySlot struct {
	MonsterID string
	Count     int
}

type fieldEncounterTemplate struct {
	RewardGold       int
	MaterialDrops    []map[string]any
	SuccessSummary   string
	FailureSummary   string
	SuccessEventType string
	FailureEventType string
	IsCurio          bool
	CurioLabel       string
	CurioOutcome     string
	FollowupQuest    *CurioQuestSeed
	Enemies          []fieldEnemySlot
}

func (t fieldEncounterTemplate) enemyCount() int {
	total := 0
	for _, slot := range t.Enemies {
		total += slot.Count
	}
	return total
}

func (t fieldEncounterTemplate) compositionLog() []map[string]any {
	items := make([]map[string]any, 0, len(t.Enemies))
	for _, slot := range t.Enemies {
		items = append(items, map[string]any{
			"monster_id": slot.MonsterID,
			"count":      slot.Count,
		})
	}
	return items
}

type fieldMonsterBlueprint struct {
	Name    string
	Role    string
	MaxHP   int
	PhysAtk int
	MagAtk  int
	PhysDef int
	MagDef  int
	Speed   int
	HealPow int
}

func buildForestFieldEncounter(detail RegionDetail, approach string) fieldEncounterTemplate {
	switch approach {
	case "gather":
		drops := []map[string]any{
			{"material_key": "whisperleaf", "quantity": 2},
			{"material_key": "damp_moss", "quantity": 1},
			{"material_key": "thorn_vine", "quantity": 1},
		}
		return fieldEncounterTemplate{
			RewardGold:       22,
			MaterialDrops:    drops,
			SuccessSummary:   "secured a reagent route through Whispering Forest",
			FailureSummary:   "was forced off the reagent route before the pack scattered",
			SuccessEventType: "field.gathering_resolved",
			FailureEventType: "field.skirmish_failed",
			Enemies: []fieldEnemySlot{
				{MonsterID: "forest_stalker", Count: 1},
				{MonsterID: "vine_seer", Count: 1},
			},
		}
	case "curio":
		drops := []map[string]any{
			{"material_key": "whisperleaf", "quantity": 2},
			{"material_key": "shrine_amber", "quantity": 1},
			{"material_key": "wolf_pelt", "quantity": 1},
		}
		return fieldEncounterTemplate{
			RewardGold:       48,
			MaterialDrops:    drops,
			SuccessSummary:   "resolved a shrine echo and recovered a scout from its guarded cache",
			FailureSummary:   "was driven away from the shrine cache before the rescue could finish",
			SuccessEventType: "field.curio_resolved",
			FailureEventType: "field.curio_failed",
			IsCurio:          true,
			CurioLabel:       "Shrine Echo Cache",
			CurioOutcome:     "rescue_followup",
			FollowupQuest: &CurioQuestSeed{
				TemplateType:     "curio_followup_delivery",
				Title:            "Escort the Shrine Witness",
				Description:      "Deliver the rescued forest scout and shrine notes to Greenfield Outpost for debrief.",
				TargetRegionID:   "greenfield_village",
				RewardGold:       56,
				RewardReputation: 9,
			},
			Enemies: []fieldEnemySlot{
				{MonsterID: "forest_stalker", Count: 1},
				{MonsterID: "vine_seer", Count: 1},
			},
		}
	default:
		drops := []map[string]any{
			{"material_key": "wolf_pelt", "quantity": 1},
			{"material_key": "whisperleaf", "quantity": 1},
			{"material_key": "beast_bone_shard", "quantity": 1},
		}
		return fieldEncounterTemplate{
			RewardGold:       34,
			MaterialDrops:    drops,
			SuccessSummary:   "cleared a forest hunt route",
			FailureSummary:   "was repelled by the forest pack before the route was secured",
			SuccessEventType: "field.encounter_resolved",
			FailureEventType: "field.skirmish_failed",
			Enemies: []fieldEnemySlot{
				{MonsterID: "forest_stalker", Count: 2},
				{MonsterID: "vine_seer", Count: 1},
			},
		}
	}
}

func buildBriarFieldEncounter(detail RegionDetail, approach string) fieldEncounterTemplate {
	switch approach {
	case "gather":
		drops := []map[string]any{
			{"material_key": "thorn_vine", "quantity": 2},
			{"material_key": "moonbloom_pollen", "quantity": 1},
			{"material_key": "amber_bark", "quantity": 1},
		}
		return fieldEncounterTemplate{
			RewardGold:       30,
			MaterialDrops:    drops,
			SuccessSummary:   "cut a clean supply lane through the briar growth",
			FailureSummary:   "lost control of the briar lane under pressure from prowlers",
			SuccessEventType: "field.gathering_resolved",
			FailureEventType: "field.skirmish_failed",
			Enemies: []fieldEnemySlot{
				{MonsterID: "thornprowler", Count: 1},
				{MonsterID: "briar_hexcaller", Count: 1},
			},
		}
	case "curio":
		drops := []map[string]any{
			{"material_key": "amber_bark", "quantity": 2},
			{"material_key": "moonbloom_pollen", "quantity": 1},
			{"material_key": "predator_fang", "quantity": 1},
		}
		return fieldEncounterTemplate{
			RewardGold:       56,
			MaterialDrops:    drops,
			SuccessSummary:   "stabilized an overgrown altar and recovered its hunt marker",
			FailureSummary:   "lost the altar marker when the hollow pack counterattacked",
			SuccessEventType: "field.curio_resolved",
			FailureEventType: "field.curio_failed",
			IsCurio:          true,
			CurioLabel:       "Overgrown Hunt Altar",
			CurioOutcome:     "predator_trace",
			FollowupQuest: &CurioQuestSeed{
				TemplateType:     "curio_followup_delivery",
				Title:            "Trace the Briarqueen's Route",
				Description:      "Bring the recovered hunt marker back to Greenfield Outpost so the guild can mark the predator lane.",
				TargetRegionID:   "greenfield_village",
				RewardGold:       72,
				RewardReputation: 10,
			},
			Enemies: []fieldEnemySlot{
				{MonsterID: "thornprowler", Count: 2},
				{MonsterID: "briar_hexcaller", Count: 1},
			},
		}
	default:
		drops := []map[string]any{
			{"material_key": "predator_fang", "quantity": 1},
			{"material_key": "amber_bark", "quantity": 1},
			{"material_key": "thorn_vine", "quantity": 1},
		}
		return fieldEncounterTemplate{
			RewardGold:       42,
			MaterialDrops:    drops,
			SuccessSummary:   "broke a briar ambush route",
			FailureSummary:   "was outpaced by the thorn pack on the ambush route",
			SuccessEventType: "field.encounter_resolved",
			FailureEventType: "field.skirmish_failed",
			Enemies: []fieldEnemySlot{
				{MonsterID: "thornprowler", Count: 2},
				{MonsterID: "briar_trapper", Count: 1},
			},
		}
	}
}

func buildDesertFieldEncounter(detail RegionDetail, approach string) fieldEncounterTemplate {
	switch approach {
	case "gather":
		drops := []map[string]any{
			{"material_key": "dry_resin", "quantity": 2},
			{"material_key": "dust_crystal", "quantity": 1},
			{"material_key": "sunscorched_ore", "quantity": 1},
		}
		return fieldEncounterTemplate{
			RewardGold:       30,
			MaterialDrops:    drops,
			SuccessSummary:   "secured a desert salvage route under light pressure",
			FailureSummary:   "was pushed off the salvage route by frontier raiders",
			SuccessEventType: "field.gathering_resolved",
			FailureEventType: "field.skirmish_failed",
			Enemies: []fieldEnemySlot{
				{MonsterID: "sunscar_raider", Count: 1},
				{MonsterID: "dust_channeler", Count: 1},
			},
		}
	case "curio":
		drops := []map[string]any{
			{"material_key": "sunscorched_ore", "quantity": 2},
			{"material_key": "venom_sac", "quantity": 1},
			{"material_key": "dust_crystal", "quantity": 1},
		}
		return fieldEncounterTemplate{
			RewardGold:       64,
			MaterialDrops:    drops,
			SuccessSummary:   "investigated a ruined beacon and extracted its distress core after the ambush",
			FailureSummary:   "failed to hold the beacon site long enough to recover its distress core",
			SuccessEventType: "field.curio_resolved",
			FailureEventType: "field.curio_failed",
			IsCurio:          true,
			CurioLabel:       "Ruined Beacon Distress",
			CurioOutcome:     "contract_redirect",
			FollowupQuest: &CurioQuestSeed{
				TemplateType:     "curio_followup_delivery",
				Title:            "Return the Beacon Core",
				Description:      "Bring the recovered beacon core back to Ironbanner City so the guild can redirect the frontier contract line.",
				TargetRegionID:   "main_city",
				RewardGold:       84,
				RewardReputation: 12,
			},
			Enemies: []fieldEnemySlot{
				{MonsterID: "sunscar_raider", Count: 1},
				{MonsterID: "dust_channeler", Count: 1},
				{MonsterID: "warvault_scout", Count: 1},
			},
		}
	default:
		drops := []map[string]any{
			{"material_key": "sand_carapace", "quantity": 1},
			{"material_key": "sunscorched_ore", "quantity": 1},
			{"material_key": "venom_sac", "quantity": 1},
		}
		return fieldEncounterTemplate{
			RewardGold:       52,
			MaterialDrops:    drops,
			SuccessSummary:   "broke an elite desert patrol and looted the route",
			FailureSummary:   "was forced to disengage from the elite frontier patrol",
			SuccessEventType: "field.encounter_resolved",
			FailureEventType: "field.skirmish_failed",
			Enemies: []fieldEnemySlot{
				{MonsterID: "sunscar_raider", Count: 2},
				{MonsterID: "dust_channeler", Count: 1},
			},
		}
	}
}

func buildAshenFieldEncounter(detail RegionDetail, approach string) fieldEncounterTemplate {
	switch approach {
	case "gather":
		drops := []map[string]any{
			{"material_key": "obsidian_shard", "quantity": 2},
			{"material_key": "void_residue", "quantity": 1},
			{"material_key": "blackglass_dust", "quantity": 1},
		}
		return fieldEncounterTemplate{
			RewardGold:       44,
			MaterialDrops:    drops,
			SuccessSummary:   "secured a blackglass salvage pass along the ridge",
			FailureSummary:   "lost the salvage pass when the ridge sentries converged",
			SuccessEventType: "field.gathering_resolved",
			FailureEventType: "field.skirmish_failed",
			Enemies: []fieldEnemySlot{
				{MonsterID: "ridge_sentinel", Count: 1},
				{MonsterID: "void_adept", Count: 1},
			},
		}
	case "curio":
		drops := []map[string]any{
			{"material_key": "obsidian_shard", "quantity": 2},
			{"material_key": "void_residue", "quantity": 1},
			{"material_key": "fractured_focus", "quantity": 1},
		}
		return fieldEncounterTemplate{
			RewardGold:       72,
			MaterialDrops:    drops,
			SuccessSummary:   "decoded a fractured spire relay and captured its focusing lens",
			FailureSummary:   "lost the spire relay lens when the ridge casters overwhelmed the team",
			SuccessEventType: "field.curio_resolved",
			FailureEventType: "field.curio_failed",
			IsCurio:          true,
			CurioLabel:       "Fractured Spire Relay",
			CurioOutcome:     "arcane_redirect",
			FollowupQuest: &CurioQuestSeed{
				TemplateType:     "curio_followup_delivery",
				Title:            "Deliver the Relay Lens",
				Description:      "Return the recovered relay lens to Ironbanner City for arcane analysis and route clearance.",
				TargetRegionID:   "main_city",
				RewardGold:       96,
				RewardReputation: 13,
			},
			Enemies: []fieldEnemySlot{
				{MonsterID: "ridge_sentinel", Count: 1},
				{MonsterID: "void_adept", Count: 2},
			},
		}
	default:
		drops := []map[string]any{
			{"material_key": "obsidian_shard", "quantity": 1},
			{"material_key": "blackglass_dust", "quantity": 1},
			{"material_key": "fractured_focus", "quantity": 1},
		}
		return fieldEncounterTemplate{
			RewardGold:       58,
			MaterialDrops:    drops,
			SuccessSummary:   "cleared an obsidian ridge patrol route",
			FailureSummary:   "was broken by the ridge patrol before the route could be held",
			SuccessEventType: "field.encounter_resolved",
			FailureEventType: "field.skirmish_failed",
			Enemies: []fieldEnemySlot{
				{MonsterID: "ridge_sentinel", Count: 2},
				{MonsterID: "void_adept", Count: 1},
			},
		}
	}
}

func buildFieldEnemyCombatants(regionID, approach string) []combat.Combatant {
	templates := map[string]fieldEncounterTemplate{
		"whispering_forest:hunt":          buildForestFieldEncounter(RegionDetail{}, "hunt"),
		"whispering_forest:gather":        buildForestFieldEncounter(RegionDetail{}, "gather"),
		"whispering_forest:curio":         buildForestFieldEncounter(RegionDetail{}, "curio"),
		"briar_thicket:hunt":              buildBriarFieldEncounter(RegionDetail{}, "hunt"),
		"briar_thicket:gather":            buildBriarFieldEncounter(RegionDetail{}, "gather"),
		"briar_thicket:curio":             buildBriarFieldEncounter(RegionDetail{}, "curio"),
		"sunscar_desert_outskirts:hunt":   buildDesertFieldEncounter(RegionDetail{}, "hunt"),
		"sunscar_desert_outskirts:gather": buildDesertFieldEncounter(RegionDetail{}, "gather"),
		"sunscar_desert_outskirts:curio":  buildDesertFieldEncounter(RegionDetail{}, "curio"),
		"ashen_ridge:hunt":                buildAshenFieldEncounter(RegionDetail{}, "hunt"),
		"ashen_ridge:gather":              buildAshenFieldEncounter(RegionDetail{}, "gather"),
		"ashen_ridge:curio":               buildAshenFieldEncounter(RegionDetail{}, "curio"),
	}
	template, ok := templates[regionID+":"+approach]
	if !ok {
		return nil
	}

	blueprints := map[string]fieldMonsterBlueprint{
		"forest_stalker":  {Name: "Forest Stalker", Role: "assassin", MaxHP: 92, PhysAtk: 16, PhysDef: 6, MagDef: 5, Speed: 13},
		"vine_seer":       {Name: "Vine Seer", Role: "caster", MaxHP: 84, PhysAtk: 3, MagAtk: 16, PhysDef: 4, MagDef: 8, Speed: 10},
		"shrine_guardian": {Name: "Shrine Guardian", Role: "tank", MaxHP: 120, PhysAtk: 14, MagAtk: 6, PhysDef: 10, MagDef: 10, Speed: 8},
		"thornprowler":    {Name: "Thornprowler", Role: "assassin", MaxHP: 136, PhysAtk: 24, PhysDef: 8, MagDef: 7, Speed: 17},
		"briar_hexcaller": {Name: "Briar Hexcaller", Role: "caster", MaxHP: 116, PhysAtk: 5, MagAtk: 24, PhysDef: 6, MagDef: 11, Speed: 12},
		"briar_trapper":   {Name: "Briar Trapper", Role: "controller", MaxHP: 148, PhysAtk: 21, MagAtk: 6, PhysDef: 11, MagDef: 9, Speed: 13},
		"sunscar_raider":  {Name: "Sunscar Raider", Role: "bruiser", MaxHP: 152, PhysAtk: 26, PhysDef: 12, MagDef: 9, Speed: 11},
		"dust_channeler":  {Name: "Dust Channeler", Role: "caster", MaxHP: 120, PhysAtk: 5, MagAtk: 25, PhysDef: 7, MagDef: 12, Speed: 13},
		"warvault_scout":  {Name: "Warvault Scout", Role: "assassin", MaxHP: 132, PhysAtk: 23, PhysDef: 8, MagDef: 8, Speed: 16},
		"ridge_sentinel":  {Name: "Ridge Sentinel", Role: "tank", MaxHP: 168, PhysAtk: 24, PhysDef: 14, MagDef: 12, Speed: 10},
		"void_adept":      {Name: "Void Adept", Role: "caster", MaxHP: 122, PhysAtk: 4, MagAtk: 28, PhysDef: 7, MagDef: 13, Speed: 14},
	}

	enemies := make([]combat.Combatant, 0, template.enemyCount())
	for _, slot := range template.Enemies {
		blueprint, ok := blueprints[slot.MonsterID]
		if !ok {
			continue
		}
		for idx := 0; idx < slot.Count; idx++ {
			maxHP := blueprint.MaxHP
			enemies = append(enemies, combat.Combatant{
				EntityID:     fmt.Sprintf("%s_%s_%d", regionID, slot.MonsterID, idx+1),
				Name:         blueprint.Name,
				Team:         "b",
				IsPlayerSide: false,
				Role:         blueprint.Role,
				MaxHP:        maxHP,
				CurrentHP:    maxHP,
				PhysAtk:      blueprint.PhysAtk,
				MagAtk:       blueprint.MagAtk,
				PhysDef:      blueprint.PhysDef,
				MagDef:       blueprint.MagDef,
				Speed:        blueprint.Speed,
				HealPow:      blueprint.HealPow,
			})
		}
	}
	return enemies
}

func maxFieldTurn(log []map[string]any) int {
	maxTurn := 0
	for _, entry := range log {
		switch turn := entry["turn"].(type) {
		case int:
			if turn > maxTurn {
				maxTurn = turn
			}
		case float64:
			if int(turn) > maxTurn {
				maxTurn = int(turn)
			}
		}
	}
	return maxTurn
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

func sumMaterialDrops(drops []map[string]any) int {
	total := 0
	for _, drop := range drops {
		switch quantity := drop["quantity"].(type) {
		case int:
			total += quantity
		case int32:
			total += int(quantity)
		case int64:
			total += int(quantity)
		case float64:
			total += int(quantity)
		}
	}

	return total
}

func nextDailyReset(now time.Time) time.Time {
	resetToday := time.Date(now.Year(), now.Month(), now.Day(), 4, 0, 0, 0, now.Location())
	if now.Before(resetToday) {
		return resetToday
	}

	return resetToday.Add(24 * time.Hour)
}

func currentArenaStatus(now time.Time) ArenaStatus {
	switch now.Weekday() {
	case time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday:
		return ArenaStatus{
			Code:          "rating_open",
			Label:         "Rating Season Live",
			Details:       "Bots can make three free rating challenges each day, buy extra attempts, and climb toward Saturday's top-64 knockout.",
			NextMilestone: "Weekly knockout seeds lock on Saturday 09:00",
		}
	case time.Saturday:
		hour := now.Hour()
		minute := now.Minute()
		stageMinute := minute - 5
		if hour < 9 {
			return ArenaStatus{
				Code:          "knockout_pending",
				Label:         "Knockout Pending",
				Details:       "This week's top 64 rating seeds are locked in and today's elimination bracket is about to begin.",
				NextMilestone: "Round of 64 starts at 09:05",
			}
		}
		if hour == 9 && minute < 35 {
			roundLabel := "Round of 64"
			nextMilestone := "Main bracket rounds continue resolving every five minutes"
			switch {
			case stageMinute < 5:
				roundLabel = "Round of 64"
				nextMilestone = "Round of 32 starts at 09:10"
			case stageMinute < 10:
				roundLabel = "Round of 32"
				nextMilestone = "Round of 16 starts at 09:15"
			case stageMinute < 15:
				roundLabel = "Round of 16"
				nextMilestone = "Quarterfinals start at 09:20"
			case stageMinute < 20:
				roundLabel = "Quarterfinals"
				nextMilestone = "Semifinals start at 09:25"
			case stageMinute < 25:
				roundLabel = "Semifinals"
				nextMilestone = "Final starts at 09:30"
			default:
				roundLabel = "Final"
				nextMilestone = "Champion is declared at 09:35"
			}
			return ArenaStatus{
				Code:          "knockout_in_progress",
				Label:         "Knockout In Progress",
				Details:       fmt.Sprintf("%s matches are auto-resolving now from this week's 64 seeded qualifiers.", roundLabel),
				NextMilestone: nextMilestone,
			}
		}
		return ArenaStatus{
			Code:          "knockout_results_live",
			Label:         "Knockout Results Live",
			Details:       "This week's knockout bracket is complete and title assignment will be finalized during Sunday's rest window.",
			NextMilestone: "Titles and seasonal rest begin on Sunday",
		}
	default:
		return ArenaStatus{
			Code:          "rest_day",
			Label:         "Arena Rest Day",
			Details:       "Arena combat is paused today while weekly titles remain active and a new rating week prepares to open on Monday.",
			NextMilestone: "The next rating week opens on Monday 00:00",
		}
	}
}

func arenaStatusArenaPopulation(code string) int {
	switch code {
	case "rating_open":
		return 18
	case "knockout_pending":
		return 64
	case "knockout_in_progress":
		return 64
	case "knockout_results_live":
		return 1
	case "rest_day":
		return 0
	default:
		return 0
	}
}

type eventTemplate struct {
	EventType        string
	ActorCharacterID string
	ActorName        string
	RegionID         string
	Summary          string
	Payload          map[string]any
	Offset           time.Duration
}

var seedRegions = []RegionDetail{
	{
		Region:      Region{ID: "main_city", Name: "Main City", Type: "safe_hub", TravelCostGold: 0},
		Description: "The capital hub for guild work, gearing, and arena registration.",
		RegionGameplay: RegionGameplay{
			InteractionLayer:  "safe_hub",
			RiskLevel:         "low",
			FacilityFocus:     "guild_services",
			EncounterFamily:   "non_combat_hub",
			CurioStatus:       "dormant",
			CurioHint:         "Guild dispatches, arena notices, and logistics requests may surface here.",
			HostileEncounters: false,
		},
		Buildings: []Building{
			{ID: "guild_main_city", Name: "Adventurers Guild", Type: "guild", Category: "functional_building", Actions: guildActions([]string{"list_quests", "submit_quest"})},
			{ID: "equipment_shop_main_city", Name: "Equipment Shop", Type: "equipment_shop", Category: "functional_building", Actions: []string{"browse_stock", "purchase", "sell_loot"}},
			{ID: "apothecary_main_city", Name: "Apothecary", Type: "apothecary", Category: "functional_building", Actions: []string{"purchase", "restore_hp"}},
			{ID: "blacksmith_main_city", Name: "Blacksmith", Type: "blacksmith", Category: "functional_building", Actions: []string{"enhance_item", "salvage_item"}},
			{ID: "arena_main_city", Name: "Arena", Type: "arena", Category: "functional_building", Actions: []string{"view_bracket", "signup"}},
			{ID: "warehouse_main_city", Name: "Warehouse", Type: "warehouse", Category: "functional_building", Actions: []string{"view_storage"}},
		},
		TravelOptions: []TravelOption{
			{RegionID: "greenfield_village", Name: "Greenfield Village", TravelCostGold: 0},
			{RegionID: "whispering_forest", Name: "Whispering Forest", TravelCostGold: 10},
			{RegionID: "briar_thicket", Name: "Briar Thicket", TravelCostGold: 20},
			{RegionID: "ancient_catacomb", Name: "Ancient Catacomb", TravelCostGold: 15},
			{RegionID: "thorned_hollow", Name: "Thorned Hollow", TravelCostGold: 25},
			{RegionID: "sunscar_desert_outskirts", Name: "Sunscar Desert Outskirts", TravelCostGold: 30},
			{RegionID: "sunscar_warvault", Name: "Sunscar Warvault", TravelCostGold: 40},
			{RegionID: "ashen_ridge", Name: "Ashen Ridge", TravelCostGold: 42},
			{RegionID: "obsidian_spire", Name: "Obsidian Spire", TravelCostGold: 52},
		},
	},
	{
		Region:      Region{ID: "greenfield_village", Name: "Greenfield Village", Type: "safe_hub", TravelCostGold: 0},
		Description: "A logistics stop for early contracts and recovery before pushing back into the wild.",
		RegionGameplay: RegionGameplay{
			InteractionLayer:  "safe_hub",
			RiskLevel:         "low",
			FacilityFocus:     "supply_services",
			EncounterFamily:   "route_support",
			CurioStatus:       "dormant",
			CurioHint:         "Escort requests and clinic-side supply incidents can begin here.",
			HostileEncounters: false,
		},
		Buildings: []Building{
			{ID: "guild_outpost_village", Name: "Adventurers Guild Outpost", Type: "guild", Category: "functional_building", Actions: guildActions([]string{"list_quests", "submit_quest"})},
			{ID: "equipment_shop_village", Name: "Equipment Shop", Type: "equipment_shop", Category: "functional_building", Actions: []string{"browse_stock", "purchase", "sell_loot"}},
			{ID: "apothecary_village", Name: "Apothecary", Type: "apothecary", Category: "functional_building", Actions: []string{"purchase", "restore_hp"}},
			{ID: "caravan_dispatch_village", Name: "Caravan Dispatch Point", Type: "caravan_dispatch", Category: "neutral_interaction_point", Actions: []string{"pick_up_supplies", "turn_in_contracts"}},
		},
		TravelOptions: []TravelOption{
			{RegionID: "main_city", Name: "Main City", TravelCostGold: 0},
			{RegionID: "whispering_forest", Name: "Whispering Forest", TravelCostGold: 10},
			{RegionID: "briar_thicket", Name: "Briar Thicket", TravelCostGold: 20},
		},
	},
	{
		Region:      Region{ID: "whispering_forest", Name: "Whispering Forest", Type: "field", TravelCostGold: 10},
		Description: "The first major hunting ground, built around reusable field skirmishes and the first six-room dungeon gate.",
		RegionGameplay: RegionGameplay{
			InteractionLayer:  "field",
			RiskLevel:         "low",
			FacilityFocus:     "hunt_camp",
			EncounterFamily:   "forest_hunt",
			CurioStatus:       "dormant",
			CurioHint:         "Shrine echoes, rare herb blooms, and scout rescues may appear in local routes.",
			LinkedDungeon:     "ancient_catacomb",
			HostileEncounters: true,
		},
		TravelOptions: []TravelOption{
			{RegionID: "main_city", Name: "Main City", TravelCostGold: 10},
			{RegionID: "greenfield_village", Name: "Greenfield Village", TravelCostGold: 10},
			{RegionID: "ancient_catacomb", Name: "Ancient Catacomb", TravelCostGold: 15},
		},
		EncounterSummary: &EncounterSummary{
			ActivityType: "field_combat",
			Summary:      "Early-route bots resolve real field skirmishes here for daily quests, materials, and route control.",
			Highlights:   []string{"Forest wolf packs", "Poison vine casters", "Supply delivery routes"},
		},
	},
	{
		Region:      Region{ID: "ancient_catacomb", Name: "Ancient Catacomb", Type: "dungeon", TravelCostGold: 15},
		Description: "A starter six-room dungeon with escalating undead pressure and a necromancer boss in room six.",
		RegionGameplay: RegionGameplay{
			InteractionLayer:  "dungeon",
			RiskLevel:         "mid",
			FacilityFocus:     "dungeon_gate",
			EncounterFamily:   "undead_corridors",
			CurioStatus:       "dormant",
			CurioHint:         "Sealed crypts and cursed caches can branch from the main run.",
			ParentRegionID:    "whispering_forest",
			HostileEncounters: true,
		},
		TravelOptions: []TravelOption{
			{RegionID: "main_city", Name: "Main City", TravelCostGold: 15},
			{RegionID: "whispering_forest", Name: "Whispering Forest", TravelCostGold: 15},
		},
		EncounterSummary: &EncounterSummary{
			ActivityType: "dungeon_run",
			Summary:      "Bots enter for limited daily clears, deterministic logs, and fixed six-room pacing.",
			Highlights:   []string{"6 rooms per run", "Necromancer boss", "Gold and starter gear upgrades"},
		},
	},
	{
		Region:      Region{ID: "briar_thicket", Name: "Briar Thicket", Type: "field", TravelCostGold: 20},
		Description: "A frontier woodland route where predator packs pressure daily hunt and curio loops.",
		RegionGameplay: RegionGameplay{
			InteractionLayer:  "field",
			RiskLevel:         "mid",
			FacilityFocus:     "predator_contracts",
			EncounterFamily:   "briar_skirmish",
			CurioStatus:       "dormant",
			CurioHint:         "Overgrown altars, hunt markers, and trapped caravans can turn into skirmish objectives.",
			LinkedDungeon:     "thorned_hollow",
			HostileEncounters: true,
		},
		TravelOptions: []TravelOption{
			{RegionID: "main_city", Name: "Main City", TravelCostGold: 20},
			{RegionID: "greenfield_village", Name: "Greenfield Village", TravelCostGold: 20},
			{RegionID: "thorned_hollow", Name: "Thorned Hollow", TravelCostGold: 25},
		},
		EncounterSummary: &EncounterSummary{
			ActivityType: "field_combat",
			Summary:      "Frontier bots run true field skirmishes here to support daily contracts and unlock the hollow route.",
			Highlights:   []string{"Predator ambush lanes", "Briar hexcasters", "Escort disruption sites"},
		},
	},
	{
		Region:      Region{ID: "thorned_hollow", Name: "Thorned Hollow", Type: "dungeon", TravelCostGold: 25},
		Description: "A six-room predator den focused on speed checks, focus fire, and a lethal room-six apex hunt.",
		RegionGameplay: RegionGameplay{
			InteractionLayer:  "dungeon",
			RiskLevel:         "mid",
			FacilityFocus:     "dungeon_gate",
			EncounterFamily:   "thorn_predator",
			CurioStatus:       "dormant",
			CurioHint:         "Snare nests and hunt shrines can branch off the main route.",
			ParentRegionID:    "briar_thicket",
			HostileEncounters: true,
		},
		TravelOptions: []TravelOption{
			{RegionID: "briar_thicket", Name: "Briar Thicket", TravelCostGold: 25},
			{RegionID: "main_city", Name: "Main City", TravelCostGold: 25},
		},
		EncounterSummary: &EncounterSummary{
			ActivityType: "dungeon_run",
			Summary:      "This dungeon keeps the season's six-room cadence while shifting the fight toward speed and crit pressure.",
			Highlights:   []string{"6 rooms per run", "Predator boss", "Precision-heavy clears"},
		},
	},
	{
		Region:      Region{ID: "sunscar_desert_outskirts", Name: "Sunscar Desert Outskirts", Type: "field", TravelCostGold: 30},
		Description: "A frontier route where real desert skirmishes feed daily objectives and the warvault lane.",
		RegionGameplay: RegionGameplay{
			InteractionLayer:  "field",
			RiskLevel:         "mid",
			FacilityFocus:     "frontier_contracts",
			EncounterFamily:   "desert_ambush",
			CurioStatus:       "dormant",
			CurioHint:         "Buried caches, ruin signals, and escort incidents can disrupt frontier routes.",
			LinkedDungeon:     "sunscar_warvault",
			HostileEncounters: true,
		},
		TravelOptions: []TravelOption{
			{RegionID: "main_city", Name: "Main City", TravelCostGold: 30},
			{RegionID: "sunscar_warvault", Name: "Sunscar Warvault", TravelCostGold: 40},
			{RegionID: "ashen_ridge", Name: "Ashen Ridge", TravelCostGold: 42},
		},
		EncounterSummary: &EncounterSummary{
			ActivityType: "field_combat",
			Summary:      "Frontier patrols fight real ambushes here to support daily contracts and open the warvault lane.",
			Highlights:   []string{"Sand skirmisher packs", "Dust mage ambushes", "Elite courier interceptions"},
		},
	},
	{
		Region:      Region{ID: "sunscar_warvault", Name: "Sunscar Warvault", Type: "dungeon", TravelCostGold: 40},
		Description: "A six-room fortress run built around shield breaks, burst windows, and a room-six war marshal.",
		RegionGameplay: RegionGameplay{
			InteractionLayer:  "dungeon",
			RiskLevel:         "high",
			FacilityFocus:     "elite_dungeon_camp",
			EncounterFamily:   "warvault_legion",
			CurioStatus:       "dormant",
			CurioHint:         "Supply caches, command seals, and artillery relays can appear near the run path.",
			ParentRegionID:    "sunscar_desert_outskirts",
			HostileEncounters: true,
		},
		TravelOptions: []TravelOption{
			{RegionID: "sunscar_desert_outskirts", Name: "Sunscar Desert Outskirts", TravelCostGold: 40},
			{RegionID: "main_city", Name: "Main City", TravelCostGold: 40},
		},
		EncounterSummary: &EncounterSummary{
			ActivityType: "dungeon_run",
			Summary:      "This fortress keeps the shared six-room rule while emphasizing physical breakpoints and burst.",
			Highlights:   []string{"6 rooms per run", "War marshal boss", "High-tier dungeon rewards"},
		},
	},
	{
		Region:      Region{ID: "ashen_ridge", Name: "Ashen Ridge", Type: "field", TravelCostGold: 42},
		Description: "A blackglass ridge where late-route bots clear harsh skirmishes before entering the spire line.",
		RegionGameplay: RegionGameplay{
			InteractionLayer:  "field",
			RiskLevel:         "high",
			FacilityFocus:     "arcane_salvage",
			EncounterFamily:   "obsidian_skirmish",
			CurioStatus:       "dormant",
			CurioHint:         "Fractured relays, shard storms, and void mirrors can convert into field combat objectives.",
			LinkedDungeon:     "obsidian_spire",
			HostileEncounters: true,
		},
		TravelOptions: []TravelOption{
			{RegionID: "main_city", Name: "Main City", TravelCostGold: 42},
			{RegionID: "sunscar_desert_outskirts", Name: "Sunscar Desert Outskirts", TravelCostGold: 42},
			{RegionID: "obsidian_spire", Name: "Obsidian Spire", TravelCostGold: 52},
		},
		EncounterSummary: &EncounterSummary{
			ActivityType: "field_combat",
			Summary:      "Late-route bots fight true skirmishes here to maintain daily routes and stabilize the spire approach.",
			Highlights:   []string{"Blackglass sentries", "Void casters", "Relay salvage skirmishes"},
		},
	},
	{
		Region:      Region{ID: "obsidian_spire", Name: "Obsidian Spire", Type: "dungeon", TravelCostGold: 52},
		Description: "A six-room arcane tower where magic throughput, silence pressure, and the room-six archon decide the clear.",
		RegionGameplay: RegionGameplay{
			InteractionLayer:  "dungeon",
			RiskLevel:         "high",
			FacilityFocus:     "arcane_dungeon_gate",
			EncounterFamily:   "void_spire",
			CurioStatus:       "dormant",
			CurioHint:         "Mirror halls, rupture seals, and blackglass caches can branch near the ascent.",
			ParentRegionID:    "ashen_ridge",
			HostileEncounters: true,
		},
		TravelOptions: []TravelOption{
			{RegionID: "ashen_ridge", Name: "Ashen Ridge", TravelCostGold: 52},
			{RegionID: "main_city", Name: "Main City", TravelCostGold: 52},
		},
		EncounterSummary: &EncounterSummary{
			ActivityType: "dungeon_run",
			Summary:      "This arcane tower keeps the season's six-room structure while leaning into magic pressure and anti-caster checks.",
			Highlights:   []string{"6 rooms per run", "Archon boss", "Spell-chain pressure"},
		},
	},
}

var seedRegionActivity = []RegionActivity{
	{RegionID: "main_city", Population: 24, RecentEventCount: 18, Highlight: "Arena entrants are checking brackets and upgrading gear."},
	{RegionID: "greenfield_village", Population: 12, RecentEventCount: 9, Highlight: "Supply runners are rotating through healer and outpost loops."},
	{RegionID: "whispering_forest", Population: 21, RecentEventCount: 26, Highlight: "Quest traffic is heavy as early-route bots farm reputation."},
	{RegionID: "ancient_catacomb", Population: 11, RecentEventCount: 14, Highlight: "Starter dungeon clears are pacing today's gold income."},
	{RegionID: "briar_thicket", Population: 15, RecentEventCount: 17, Highlight: "Frontier parties are trading clean skirmish clears for contract progress."},
	{RegionID: "thorned_hollow", Population: 9, RecentEventCount: 10, Highlight: "Predator clears are becoming the new frontier benchmark."},
	{RegionID: "sunscar_desert_outskirts", Population: 10, RecentEventCount: 11, Highlight: "Frontier parties are pushing ambush contracts with real battle logs."},
	{RegionID: "sunscar_warvault", Population: 8, RecentEventCount: 9, Highlight: "Warvault attempts are rising as players test burst windows."},
	{RegionID: "ashen_ridge", Population: 7, RecentEventCount: 8, Highlight: "Blackglass skirmishes are feeding late-route daily objectives."},
	{RegionID: "obsidian_spire", Population: 6, RecentEventCount: 8, Highlight: "Arcane tower attempts remain dangerous but increasingly watchable."},
}

var seedEventTemplates = []eventTemplate{
	{
		EventType:        "dungeon.cleared",
		ActorCharacterID: "char_ferrin",
		ActorName:        "Ferrin-7",
		RegionID:         "ancient_catacomb",
		Summary:          "Ferrin-7 cleared Ancient Catacomb and extracted with upgraded gear.",
		Payload:          map[string]any{"dungeon_id": "ancient_catacomb_v1", "reward_gold": 238},
		Offset:           6 * time.Minute,
	},
	{
		EventType:        "quest.submitted",
		ActorCharacterID: "char_lyra",
		ActorName:        "LyraLoop",
		RegionID:         "main_city",
		Summary:          "LyraLoop submitted a guild contract and pushed deeper into the frontier reputation track.",
		Payload:          map[string]any{"reward_reputation": 26, "reward_gold": 120},
		Offset:           13 * time.Minute,
	},
	{
		EventType:        "travel.completed",
		ActorCharacterID: "char_toma",
		ActorName:        "TomaSeed",
		RegionID:         "sunscar_desert_outskirts",
		Summary:          "TomaSeed fast-travelled to Sunscar Desert Outskirts for elite patrol contracts.",
		Payload:          map[string]any{"from_region_id": "main_city", "travel_cost_gold": 30},
		Offset:           21 * time.Minute,
	},
	{
		EventType:        "field.encounter_resolved",
		ActorCharacterID: "char_mira",
		ActorName:        "MiraBot",
		RegionID:         "briar_thicket",
		Summary:          "MiraBot closed a briar ambush lane and converted the skirmish into contract progress.",
		Payload:          map[string]any{"battle_type": "field_skirmish", "victory": true},
		Offset:           28 * time.Minute,
	},
	{
		EventType:        "arena.entry_accepted",
		ActorCharacterID: "char_nova",
		ActorName:        "NovaScript",
		RegionID:         "main_city",
		Summary:          "NovaScript locked in a signup for tomorrow's 09:00 arena qualifiers.",
		Payload:          map[string]any{"status": "signed_up"},
		Offset:           34 * time.Minute,
	},
	{
		EventType:        "dungeon.cleared",
		ActorCharacterID: "char_kiro",
		ActorName:        "KiroNode",
		RegionID:         "obsidian_spire",
		Summary:          "KiroNode cleared Obsidian Spire and stabilized the room-six archon route.",
		Payload:          map[string]any{"dungeon_id": "obsidian_spire_v1", "reward_gold": 412},
		Offset:           42 * time.Minute,
	},
	{
		EventType:        "dungeon.entered",
		ActorCharacterID: "char_kiro",
		ActorName:        "KiroNode",
		RegionID:         "sunscar_warvault",
		Summary:          "KiroNode entered Sunscar Warvault and started a six-room fortress run.",
		Payload:          map[string]any{"dungeon_id": "sunscar_warvault_v1"},
		Offset:           57 * time.Minute,
	},
}

var seedLeaderboards = Leaderboards{
	Reputation: []LeaderboardEntry{
		{Rank: 1, CharacterID: "char_lyra", Name: "LyraLoop", Class: "priest", WeaponStyle: "holy_tome", RegionID: "main_city", Score: 812, ScoreLabel: "reputation", ActivityLabel: "Quest routing specialist"},
		{Rank: 2, CharacterID: "char_nova", Name: "NovaScript", Class: "warrior", WeaponStyle: "great_axe", RegionID: "main_city", Score: 768, ScoreLabel: "reputation", ActivityLabel: "Arena prep rotations"},
		{Rank: 3, CharacterID: "char_mira", Name: "MiraBot", Class: "mage", WeaponStyle: "staff", RegionID: "briar_thicket", Score: 731, ScoreLabel: "reputation", ActivityLabel: "Forest contract grinder"},
	},
	Gold: []LeaderboardEntry{
		{Rank: 1, CharacterID: "char_ferrin", Name: "Ferrin-7", Class: "warrior", WeaponStyle: "sword_shield", RegionID: "ancient_catacomb", Score: 12640, ScoreLabel: "gold", ActivityLabel: "Starter dungeon farming"},
		{Rank: 2, CharacterID: "char_kiro", Name: "KiroNode", Class: "mage", WeaponStyle: "spellbook", RegionID: "sunscar_warvault", Score: 11890, ScoreLabel: "gold", ActivityLabel: "Late-route dungeon loop"},
		{Rank: 3, CharacterID: "char_toma", Name: "TomaSeed", Class: "priest", WeaponStyle: "scepter", RegionID: "sunscar_desert_outskirts", Score: 11220, ScoreLabel: "gold", ActivityLabel: "Frontier courier disruptor"},
	},
	WeeklyArena: []LeaderboardEntry{
		{Rank: 1, CharacterID: "char_nova", Name: "NovaScript", Class: "warrior", WeaponStyle: "great_axe", RegionID: "main_city", Score: 1, ScoreLabel: "seed", ActivityLabel: "Projected top seed"},
		{Rank: 2, CharacterID: "char_lyra", Name: "LyraLoop", Class: "priest", WeaponStyle: "holy_tome", RegionID: "main_city", Score: 2, ScoreLabel: "seed", ActivityLabel: "Bracket control pick"},
		{Rank: 3, CharacterID: "char_kiro", Name: "KiroNode", Class: "mage", WeaponStyle: "spellbook", RegionID: "obsidian_spire", Score: 3, ScoreLabel: "seed", ActivityLabel: "Burst finisher"},
	},
	DungeonClears: []LeaderboardEntry{
		{Rank: 1, CharacterID: "char_ferrin", Name: "Ferrin-7", Class: "warrior", WeaponStyle: "sword_shield", RegionID: "ancient_catacomb", Score: 19, ScoreLabel: "clears", ActivityLabel: "Ancient Catacomb specialist"},
		{Rank: 2, CharacterID: "char_kiro", Name: "KiroNode", Class: "mage", WeaponStyle: "spellbook", RegionID: "obsidian_spire", Score: 14, ScoreLabel: "clears", ActivityLabel: "Obsidian Spire frontrunner"},
		{Rank: 3, CharacterID: "char_mira", Name: "MiraBot", Class: "mage", WeaponStyle: "staff", RegionID: "thorned_hollow", Score: 12, ScoreLabel: "clears", ActivityLabel: "Fast resolver"},
	},
}

func init() {
	for index := range seedRegions {
		if len(seedRegions[index].TravelOptions) == 0 {
			seedRegions[index].TravelOptions = defaultTravelOptions(seedRegions[index].Region.ID)
		}
	}
}

func defaultTravelOptions(currentRegionID string) []TravelOption {
	options := make([]TravelOption, 0, len(seedRegions)-1)
	for _, region := range seedRegions {
		if region.Region.ID == currentRegionID {
			continue
		}

		options = append(options, TravelOption{
			RegionID:       region.Region.ID,
			Name:           region.Region.Name,
			TravelCostGold: region.Region.TravelCostGold,
		})
	}

	slices.SortFunc(options, func(left, right TravelOption) int {
		if left.TravelCostGold == right.TravelCostGold {
			switch {
			case left.RegionID < right.RegionID:
				return -1
			case left.RegionID > right.RegionID:
				return 1
			default:
				return 0
			}
		}

		return left.TravelCostGold - right.TravelCostGold
	})

	return options
}
