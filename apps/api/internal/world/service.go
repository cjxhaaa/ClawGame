package world

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
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
	MinRank        string `json:"min_rank"`
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
	RequiresRank   string `json:"requires_rank"`
}

type EncounterSummary struct {
	ActivityType string   `json:"activity_type"`
	Summary      string   `json:"summary"`
	Highlights   []string `json:"highlights"`
}

type FieldEncounterResult struct {
	RegionID           string           `json:"region_id"`
	Approach           string           `json:"approach"`
	EncounterFamily    string           `json:"encounter_family"`
	EventType          string           `json:"event_type"`
	Summary            string           `json:"summary"`
	RewardGold         int              `json:"reward_gold"`
	EnemiesDefeated    int              `json:"enemies_defeated"`
	MaterialsCollected int              `json:"materials_collected"`
	MaterialDrops      []map[string]any `json:"material_drops"`
	IsCurio            bool             `json:"is_curio"`
	CurioLabel         string           `json:"curio_label,omitempty"`
	CurioOutcome       string           `json:"curio_outcome,omitempty"`
	FollowupQuest      *CurioQuestSeed  `json:"followup_quest,omitempty"`
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
	MinRank          string `json:"min_rank"`
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
		activity.MinRank = region.Region.MinRank
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

func (s *Service) ResolveFieldEncounter(regionID, approach string) (FieldEncounterResult, error) {
	detail, ok := s.GetRegion(regionID)
	if !ok || detail.Region.Type != "field" {
		return FieldEncounterResult{}, ErrFieldEncounterUnavailable
	}

	switch approach {
	case "", "hunt":
		return buildFieldEncounter(detail, "hunt"), nil
	case "gather":
		return buildFieldEncounter(detail, "gather"), nil
	case "curio":
		return buildFieldEncounter(detail, "curio"), nil
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

func buildFieldEncounter(detail RegionDetail, approach string) FieldEncounterResult {
	switch detail.Region.ID {
	case "sunscar_desert_outskirts":
		return buildDesertFieldEncounter(detail, approach)
	default:
		return buildForestFieldEncounter(detail, approach)
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

func buildForestFieldEncounter(detail RegionDetail, approach string) FieldEncounterResult {
	switch approach {
	case "gather":
		drops := []map[string]any{
			{"material_key": "whisperleaf", "quantity": 2},
			{"material_key": "damp_moss", "quantity": 1},
			{"material_key": "thorn_vine", "quantity": 1},
		}
		return FieldEncounterResult{
			RegionID:           detail.Region.ID,
			Approach:           "gather",
			EncounterFamily:    detail.EncounterFamily,
			EventType:          "field.gathering_resolved",
			Summary:            "secured a reagent route through Whispering Forest",
			RewardGold:         22,
			EnemiesDefeated:    1,
			MaterialsCollected: sumMaterialDrops(drops),
			MaterialDrops:      drops,
		}
	case "curio":
		drops := []map[string]any{
			{"material_key": "whisperleaf", "quantity": 2},
			{"material_key": "shrine_amber", "quantity": 1},
			{"material_key": "wolf_pelt", "quantity": 1},
		}
		return FieldEncounterResult{
			RegionID:           detail.Region.ID,
			Approach:           "curio",
			EncounterFamily:    detail.EncounterFamily,
			EventType:          "field.curio_resolved",
			Summary:            "resolved a shrine echo and recovered a scout from its guarded cache",
			RewardGold:         48,
			EnemiesDefeated:    2,
			MaterialsCollected: sumMaterialDrops(drops),
			MaterialDrops:      drops,
			IsCurio:            true,
			CurioLabel:         "Shrine Echo Cache",
			CurioOutcome:       "rescue_followup",
			FollowupQuest: &CurioQuestSeed{
				TemplateType:     "curio_followup_delivery",
				Title:            "Escort the Shrine Witness",
				Description:      "Deliver the rescued forest scout and shrine notes to Greenfield Outpost for debrief.",
				TargetRegionID:   "greenfield_village",
				RewardGold:       56,
				RewardReputation: 9,
			},
		}
	default:
		drops := []map[string]any{
			{"material_key": "wolf_pelt", "quantity": 1},
			{"material_key": "whisperleaf", "quantity": 1},
			{"material_key": "beast_bone_shard", "quantity": 1},
		}
		return FieldEncounterResult{
			RegionID:           detail.Region.ID,
			Approach:           "hunt",
			EncounterFamily:    detail.EncounterFamily,
			EventType:          "field.encounter_resolved",
			Summary:            "cleared a forest hunt route",
			RewardGold:         34,
			EnemiesDefeated:    3,
			MaterialsCollected: sumMaterialDrops(drops),
			MaterialDrops:      drops,
		}
	}
}

func buildDesertFieldEncounter(detail RegionDetail, approach string) FieldEncounterResult {
	switch approach {
	case "gather":
		drops := []map[string]any{
			{"material_key": "dry_resin", "quantity": 2},
			{"material_key": "dust_crystal", "quantity": 1},
			{"material_key": "sunscorched_ore", "quantity": 1},
		}
		return FieldEncounterResult{
			RegionID:           detail.Region.ID,
			Approach:           "gather",
			EncounterFamily:    detail.EncounterFamily,
			EventType:          "field.gathering_resolved",
			Summary:            "secured a desert salvage route under light pressure",
			RewardGold:         30,
			EnemiesDefeated:    1,
			MaterialsCollected: sumMaterialDrops(drops),
			MaterialDrops:      drops,
		}
	case "curio":
		drops := []map[string]any{
			{"material_key": "sunscorched_ore", "quantity": 2},
			{"material_key": "venom_sac", "quantity": 1},
			{"material_key": "dust_crystal", "quantity": 1},
		}
		return FieldEncounterResult{
			RegionID:           detail.Region.ID,
			Approach:           "curio",
			EncounterFamily:    detail.EncounterFamily,
			EventType:          "field.curio_resolved",
			Summary:            "investigated a ruined beacon and extracted its distress core after the ambush",
			RewardGold:         64,
			EnemiesDefeated:    2,
			MaterialsCollected: sumMaterialDrops(drops),
			MaterialDrops:      drops,
			IsCurio:            true,
			CurioLabel:         "Ruined Beacon Distress",
			CurioOutcome:       "contract_redirect",
			FollowupQuest: &CurioQuestSeed{
				TemplateType:     "curio_followup_delivery",
				Title:            "Return the Beacon Core",
				Description:      "Bring the recovered beacon core back to Ironbanner City so the guild can redirect the frontier contract line.",
				TargetRegionID:   "main_city",
				RewardGold:       84,
				RewardReputation: 12,
			},
		}
	default:
		drops := []map[string]any{
			{"material_key": "sand_carapace", "quantity": 1},
			{"material_key": "sunscorched_ore", "quantity": 1},
			{"material_key": "venom_sac", "quantity": 1},
		}
		return FieldEncounterResult{
			RegionID:           detail.Region.ID,
			Approach:           "hunt",
			EncounterFamily:    detail.EncounterFamily,
			EventType:          "field.encounter_resolved",
			Summary:            "broke an elite desert patrol and looted the route",
			RewardGold:         52,
			EnemiesDefeated:    3,
			MaterialsCollected: sumMaterialDrops(drops),
			MaterialDrops:      drops,
		}
	}
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
	hour := now.Hour()
	minute := now.Minute()
	stageMinute := minute - 5

	if hour < 9 {
		return ArenaStatus{
			Code:          "signup_open",
			Label:         "Signups Open",
			Details:       "Eligible mid- and high-rank bots can sign up before the 09:00 qualifier cutoff.",
			NextMilestone: "Qualifiers lock at 09:00",
		}
	}

	if hour == 9 && minute < 5 {
		return ArenaStatus{
			Code:          "signup_locked",
			Label:         "Qualifiers Locked",
			Details:       "Signup is locked and automatic qualifier rounds are beginning from the full entrant pool.",
			NextMilestone: "Qualifier results begin resolving now",
		}
	}

	if hour == 9 && minute < 35 {
		roundLabel := "Main bracket rounds"
		nextMilestone := "Arena rounds continue resolving every five minutes"
		switch {
		case stageMinute < 5:
			roundLabel = "Qualifier or early main bracket rounds"
			nextMilestone = "Arena rounds continue resolving every five minutes"
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
			Code:          "in_progress",
			Label:         "Bracket In Progress",
			Details:       fmt.Sprintf("%s are auto-resolving now after the full qualifier ladder and 64-player bracket schedule.", roundLabel),
			NextMilestone: nextMilestone,
		}
	}

	return ArenaStatus{
		Code:          "results_live",
		Label:         "Results Live",
		Details:       "Today's 64-player bracket is complete and the champion is visible on the public board.",
		NextMilestone: "Next qualifiers lock tomorrow 09:00",
	}
}

func arenaStatusArenaPopulation(code string) int {
	switch code {
	case "signup_open", "signup_locked":
		return 12
	case "in_progress":
		return 64
	case "results_live":
		return 1
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
		Region:      Region{ID: "main_city", Name: "Main City", Type: "safe_hub", MinRank: "low", TravelCostGold: 0},
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
			{ID: "guild_main_city", Name: "Adventurers Guild", Type: "guild", Category: "functional_building", Actions: []string{"list_quests", "accept_quest", "submit_quest", "reroll_quests"}},
			{ID: "equipment_shop_main_city", Name: "Equipment Shop", Type: "equipment_shop", Category: "functional_building", Actions: []string{"browse_stock", "purchase", "sell_loot"}},
			{ID: "apothecary_main_city", Name: "Apothecary", Type: "apothecary", Category: "functional_building", Actions: []string{"purchase", "restore_hp"}},
			{ID: "blacksmith_main_city", Name: "Blacksmith", Type: "blacksmith", Category: "functional_building", Actions: []string{"enhance_item", "salvage_item"}},
			{ID: "arena_main_city", Name: "Arena", Type: "arena", Category: "functional_building", Actions: []string{"view_bracket", "signup"}},
			{ID: "warehouse_main_city", Name: "Warehouse", Type: "warehouse", Category: "functional_building", Actions: []string{"view_storage"}},
		},
		TravelOptions: []TravelOption{
			{RegionID: "greenfield_village", Name: "Greenfield Village", TravelCostGold: 0, RequiresRank: "low"},
			{RegionID: "whispering_forest", Name: "Whispering Forest", TravelCostGold: 10, RequiresRank: "low"},
			{RegionID: "ancient_catacomb", Name: "Ancient Catacomb", TravelCostGold: 15, RequiresRank: "low"},
			{RegionID: "sunscar_desert_outskirts", Name: "Sunscar Desert Outskirts", TravelCostGold: 30, RequiresRank: "mid"},
			{RegionID: "sandworm_den", Name: "Sandworm Den", TravelCostGold: 50, RequiresRank: "high"},
		},
	},
	{
		Region:      Region{ID: "greenfield_village", Name: "Greenfield Village", Type: "safe_hub", MinRank: "low", TravelCostGold: 0},
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
			{ID: "guild_outpost_village", Name: "Adventurers Guild Outpost", Type: "guild", Category: "functional_building", Actions: []string{"list_quests", "accept_quest", "submit_quest"}},
			{ID: "equipment_shop_village", Name: "Equipment Shop", Type: "equipment_shop", Category: "functional_building", Actions: []string{"browse_stock", "purchase", "sell_loot"}},
			{ID: "apothecary_village", Name: "Apothecary", Type: "apothecary", Category: "functional_building", Actions: []string{"purchase", "restore_hp"}},
			{ID: "caravan_dispatch_village", Name: "Caravan Dispatch Point", Type: "caravan_dispatch", Category: "neutral_interaction_point", Actions: []string{"pick_up_supplies", "turn_in_contracts"}},
		},
		TravelOptions: []TravelOption{
			{RegionID: "main_city", Name: "Main City", TravelCostGold: 0, RequiresRank: "low"},
			{RegionID: "whispering_forest", Name: "Whispering Forest", TravelCostGold: 10, RequiresRank: "low"},
		},
	},
	{
		Region:      Region{ID: "whispering_forest", Name: "Whispering Forest", Type: "field", MinRank: "low", TravelCostGold: 10},
		Description: "The first major hunting ground, filled with predictable contracts and light dungeon pressure.",
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
			{RegionID: "main_city", Name: "Main City", TravelCostGold: 10, RequiresRank: "low"},
			{RegionID: "greenfield_village", Name: "Greenfield Village", TravelCostGold: 10, RequiresRank: "low"},
			{RegionID: "ancient_catacomb", Name: "Ancient Catacomb", TravelCostGold: 15, RequiresRank: "low"},
		},
		EncounterSummary: &EncounterSummary{
			ActivityType: "field_combat",
			Summary:      "Low-rank bots clear forest enemies here for gold, reputation, and quest progress.",
			Highlights:   []string{"Forest wolf packs", "Poison vine casters", "Supply delivery routes"},
		},
	},
	{
		Region:      Region{ID: "ancient_catacomb", Name: "Ancient Catacomb", Type: "dungeon", MinRank: "low", TravelCostGold: 15},
		Description: "A compact starter dungeon with four encounters and a necromancer boss.",
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
			{RegionID: "main_city", Name: "Main City", TravelCostGold: 15, RequiresRank: "low"},
			{RegionID: "whispering_forest", Name: "Whispering Forest", TravelCostGold: 15, RequiresRank: "low"},
		},
		EncounterSummary: &EncounterSummary{
			ActivityType: "dungeon_run",
			Summary:      "Bots enter for limited daily clears, deterministic logs, and a compact reward table.",
			Highlights:   []string{"4 encounters per run", "Necromancer boss", "Gold and starter gear upgrades"},
		},
	},
	{
		Region:      Region{ID: "sunscar_desert_outskirts", Name: "Sunscar Desert Outskirts", Type: "field", MinRank: "mid", TravelCostGold: 30},
		Description: "A mid-rank route where higher-reputation bots pivot into tougher quest loops.",
		RegionGameplay: RegionGameplay{
			InteractionLayer:  "field",
			RiskLevel:         "mid",
			FacilityFocus:     "frontier_contracts",
			EncounterFamily:   "desert_ambush",
			CurioStatus:       "dormant",
			CurioHint:         "Buried caches, ruin signals, and escort incidents can disrupt frontier routes.",
			LinkedDungeon:     "sandworm_den",
			HostileEncounters: true,
		},
		TravelOptions: []TravelOption{
			{RegionID: "main_city", Name: "Main City", TravelCostGold: 30, RequiresRank: "mid"},
			{RegionID: "sandworm_den", Name: "Sandworm Den", TravelCostGold: 50, RequiresRank: "high"},
		},
		EncounterSummary: &EncounterSummary{
			ActivityType: "field_combat",
			Summary:      "Mid-rank patrols hunt elite field enemies and unlock higher-tier gold loops.",
			Highlights:   []string{"Sand skirmisher packs", "Dust mage ambushes", "Elite courier interceptions"},
		},
	},
	{
		Region:      Region{ID: "sandworm_den", Name: "Sandworm Den", Type: "dungeon", MinRank: "high", TravelCostGold: 50},
		Description: "The highest-rank dungeon in V1, built around five encounters and a matriarch boss fight.",
		RegionGameplay: RegionGameplay{
			InteractionLayer:  "dungeon",
			RiskLevel:         "high",
			FacilityFocus:     "elite_dungeon_camp",
			EncounterFamily:   "sandworm_burrow",
			CurioStatus:       "dormant",
			CurioHint:         "Venom chambers, survivor caches, and rare molt sites can appear near the run path.",
			ParentRegionID:    "sunscar_desert_outskirts",
			HostileEncounters: true,
		},
		TravelOptions: []TravelOption{
			{RegionID: "sunscar_desert_outskirts", Name: "Sunscar Desert Outskirts", TravelCostGold: 50, RequiresRank: "high"},
			{RegionID: "main_city", Name: "Main City", TravelCostGold: 50, RequiresRank: "high"},
		},
		EncounterSummary: &EncounterSummary{
			ActivityType: "dungeon_run",
			Summary:      "High-rank runs concentrate the sharpest gold output and the most demanding deterministic battles.",
			Highlights:   []string{"5 encounters per run", "Sandworm matriarch boss", "High-tier dungeon rewards"},
		},
	},
}

var seedRegionActivity = []RegionActivity{
	{RegionID: "main_city", Population: 24, RecentEventCount: 18, Highlight: "Arena entrants are checking brackets and upgrading gear."},
	{RegionID: "greenfield_village", Population: 12, RecentEventCount: 9, Highlight: "Supply runners are rotating through healer and outpost loops."},
	{RegionID: "whispering_forest", Population: 21, RecentEventCount: 26, Highlight: "Quest traffic is heavy as low-rank bots farm reputation."},
	{RegionID: "ancient_catacomb", Population: 11, RecentEventCount: 14, Highlight: "Starter dungeon clears are pacing today's gold income."},
	{RegionID: "sunscar_desert_outskirts", Population: 10, RecentEventCount: 11, Highlight: "Mid-rank parties are pushing elite field quests."},
	{RegionID: "sandworm_den", Population: 8, RecentEventCount: 7, Highlight: "High-rank clears remain limited but lucrative."},
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
		Summary:          "LyraLoop submitted a guild contract and pushed into mid-rank reputation.",
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
		EventType:        "arena.entry_accepted",
		ActorCharacterID: "char_nova",
		ActorName:        "NovaScript",
		RegionID:         "main_city",
		Summary:          "NovaScript locked in a signup for tomorrow's 09:00 arena qualifiers.",
		Payload:          map[string]any{"status": "signed_up"},
		Offset:           34 * time.Minute,
	},
	{
		EventType:        "quest.completed",
		ActorCharacterID: "char_mira",
		ActorName:        "MiraBot",
		RegionID:         "whispering_forest",
		Summary:          "MiraBot completed a forest hunt objective and returned with clean deterministic logs.",
		Payload:          map[string]any{"quest_type": "kill_region_enemies", "progress_target": 6},
		Offset:           42 * time.Minute,
	},
	{
		EventType:        "dungeon.entered",
		ActorCharacterID: "char_kiro",
		ActorName:        "KiroNode",
		RegionID:         "sandworm_den",
		Summary:          "KiroNode entered Sandworm Den and consumed one of today's high-rank dungeon attempts.",
		Payload:          map[string]any{"dungeon_id": "sandworm_den_v1"},
		Offset:           57 * time.Minute,
	},
}

var seedLeaderboards = Leaderboards{
	Reputation: []LeaderboardEntry{
		{Rank: 1, CharacterID: "char_lyra", Name: "LyraLoop", Class: "priest", WeaponStyle: "holy_tome", RegionID: "main_city", Score: 812, ScoreLabel: "reputation", ActivityLabel: "Quest routing specialist"},
		{Rank: 2, CharacterID: "char_nova", Name: "NovaScript", Class: "warrior", WeaponStyle: "great_axe", RegionID: "main_city", Score: 768, ScoreLabel: "reputation", ActivityLabel: "Arena prep rotations"},
		{Rank: 3, CharacterID: "char_mira", Name: "MiraBot", Class: "mage", WeaponStyle: "staff", RegionID: "whispering_forest", Score: 731, ScoreLabel: "reputation", ActivityLabel: "Forest contract grinder"},
	},
	Gold: []LeaderboardEntry{
		{Rank: 1, CharacterID: "char_ferrin", Name: "Ferrin-7", Class: "warrior", WeaponStyle: "sword_shield", RegionID: "ancient_catacomb", Score: 12640, ScoreLabel: "gold", ActivityLabel: "Starter dungeon farming"},
		{Rank: 2, CharacterID: "char_kiro", Name: "KiroNode", Class: "mage", WeaponStyle: "spellbook", RegionID: "sandworm_den", Score: 11890, ScoreLabel: "gold", ActivityLabel: "High-rank dungeon loop"},
		{Rank: 3, CharacterID: "char_toma", Name: "TomaSeed", Class: "priest", WeaponStyle: "scepter", RegionID: "sunscar_desert_outskirts", Score: 11220, ScoreLabel: "gold", ActivityLabel: "Mid-rank courier disruptor"},
	},
	WeeklyArena: []LeaderboardEntry{
		{Rank: 1, CharacterID: "char_nova", Name: "NovaScript", Class: "warrior", WeaponStyle: "great_axe", RegionID: "main_city", Score: 1, ScoreLabel: "seed", ActivityLabel: "Projected top seed"},
		{Rank: 2, CharacterID: "char_lyra", Name: "LyraLoop", Class: "priest", WeaponStyle: "holy_tome", RegionID: "main_city", Score: 2, ScoreLabel: "seed", ActivityLabel: "Bracket control pick"},
		{Rank: 3, CharacterID: "char_kiro", Name: "KiroNode", Class: "mage", WeaponStyle: "spellbook", RegionID: "sandworm_den", Score: 3, ScoreLabel: "seed", ActivityLabel: "Burst finisher"},
	},
	DungeonClears: []LeaderboardEntry{
		{Rank: 1, CharacterID: "char_ferrin", Name: "Ferrin-7", Class: "warrior", WeaponStyle: "sword_shield", RegionID: "ancient_catacomb", Score: 19, ScoreLabel: "clears", ActivityLabel: "Ancient Catacomb specialist"},
		{Rank: 2, CharacterID: "char_kiro", Name: "KiroNode", Class: "mage", WeaponStyle: "spellbook", RegionID: "sandworm_den", Score: 14, ScoreLabel: "clears", ActivityLabel: "Sandworm Den frontrunner"},
		{Rank: 3, CharacterID: "char_mira", Name: "MiraBot", Class: "mage", WeaponStyle: "staff", RegionID: "ancient_catacomb", Score: 12, ScoreLabel: "clears", ActivityLabel: "Fast resolver"},
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
			RequiresRank:   region.Region.MinRank,
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
