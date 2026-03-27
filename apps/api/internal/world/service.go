package world

import (
	"fmt"
	"slices"
	"time"
)

const businessTimezone = "Asia/Shanghai"

type Region struct {
	ID             string `json:"region_id"`
	Name           string `json:"name"`
	Type           string `json:"type"`
	MinRank        string `json:"min_rank"`
	TravelCostGold int    `json:"travel_cost_gold"`
}

type Building struct {
	ID      string   `json:"building_id"`
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	Actions []string `json:"actions"`
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

type RegionDetail struct {
	Region           Region            `json:"region"`
	Description      string            `json:"description"`
	Buildings        []Building        `json:"buildings"`
	TravelOptions    []TravelOption    `json:"travel_options"`
	EncounterSummary *EncounterSummary `json:"encounter_summary,omitempty"`
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

func mustLocation(name string) *time.Location {
	location, err := time.LoadLocation(name)
	if err != nil {
		panic(err)
	}

	return location
}

func nextDailyReset(now time.Time) time.Time {
	resetToday := time.Date(now.Year(), now.Month(), now.Day(), 4, 0, 0, 0, now.Location())
	if now.Before(resetToday) {
		return resetToday
	}

	return resetToday.Add(24 * time.Hour)
}

func currentArenaStatus(now time.Time) ArenaStatus {
	weekday := now.Weekday()
	hour := now.Hour()
	minute := now.Minute()

	if weekday == time.Saturday && (hour < 19 || (hour == 19 && minute < 50)) {
		return ArenaStatus{
			Code:          "signup_open",
			Label:         "Signups Open",
			Details:       "Eligible high-rank bots can still enter this week's bracket.",
			NextMilestone: "Signup closes Saturday 19:50",
		}
	}

	if weekday == time.Saturday && hour == 19 && minute >= 50 {
		return ArenaStatus{
			Code:          "signup_locked",
			Label:         "Bracket Locking",
			Details:       "Seeding is being finalized before the first arena round.",
			NextMilestone: "Tournament starts Saturday 20:00",
		}
	}

	if weekday == time.Saturday && hour >= 20 {
		return ArenaStatus{
			Code:          "in_progress",
			Label:         "In Progress",
			Details:       "Arena rounds resolve every five minutes until a champion is crowned.",
			NextMilestone: "Next round in 5 minutes",
		}
	}

	if weekday == time.Sunday {
		return ArenaStatus{
			Code:          "results_live",
			Label:         "Results Live",
			Details:       "This week's bracket is complete and standings are visible on the public board.",
			NextMilestone: "Next signup window opens Saturday",
		}
	}

	return ArenaStatus{
		Code:          "preparing",
		Label:         "Preparing",
		Details:       "Guilds are training through the week while the next arena window approaches.",
		NextMilestone: "Signup opens Saturday",
	}
}

func arenaStatusArenaPopulation(code string) int {
	switch code {
	case "signup_open", "signup_locked":
		return 12
	case "in_progress":
		return 16
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
		Buildings: []Building{
			{ID: "guild_main_city", Name: "Adventurers Guild", Type: "guild", Actions: []string{"list_quests", "accept_quest", "submit_quest", "reroll_quests"}},
			{ID: "weapon_shop_main_city", Name: "Weapon Shop", Type: "weapon_shop", Actions: []string{"browse_stock", "purchase", "sell_loot"}},
			{ID: "armor_shop_main_city", Name: "Armor Shop", Type: "armor_shop", Actions: []string{"browse_stock", "purchase", "sell_loot"}},
			{ID: "temple_main_city", Name: "Temple", Type: "temple", Actions: []string{"restore_hp_mp", "remove_status"}},
			{ID: "blacksmith_main_city", Name: "Blacksmith", Type: "blacksmith", Actions: []string{"enhance_item", "repair_item"}},
			{ID: "arena_hall_main_city", Name: "Arena Hall", Type: "arena_hall", Actions: []string{"view_bracket", "signup"}},
			{ID: "warehouse_main_city", Name: "Warehouse", Type: "warehouse", Actions: []string{"view_storage"}},
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
		Buildings: []Building{
			{ID: "quest_outpost_village", Name: "Quest Outpost", Type: "quest_outpost", Actions: []string{"pick_up_supplies", "turn_in_contracts"}},
			{ID: "general_store_village", Name: "General Store", Type: "general_store", Actions: []string{"buy_consumables", "sell_loot"}},
			{ID: "field_healer_village", Name: "Field Healer", Type: "healer", Actions: []string{"restore_hp_mp", "remove_status"}},
		},
		TravelOptions: []TravelOption{
			{RegionID: "main_city", Name: "Main City", TravelCostGold: 0, RequiresRank: "low"},
			{RegionID: "whispering_forest", Name: "Whispering Forest", TravelCostGold: 10, RequiresRank: "low"},
		},
	},
	{
		Region:      Region{ID: "whispering_forest", Name: "Whispering Forest", Type: "field", MinRank: "low", TravelCostGold: 10},
		Description: "The first major hunting ground, filled with predictable contracts and light dungeon pressure.",
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
		Summary:          "NovaScript locked in an arena signup for the coming weekend bracket.",
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
