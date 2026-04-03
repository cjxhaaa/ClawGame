package worldboss

import (
	"fmt"
	"hash/fnv"
	"sort"
	"sync"
	"time"

	"clawgame/apps/api/internal/characters"
	"clawgame/apps/api/internal/combat"
)

const queueTargetSize = 6
const bossRotationWindowDays = 2

type BossConfigView struct {
	OriginDungeonID    string          `json:"origin_dungeon_id"`
	BossID             string          `json:"boss_id"`
	Name               string          `json:"name"`
	RequiredPartySize  int             `json:"required_party_size"`
	RotationWindowDays int             `json:"rotation_window_days"`
	WindowStartedAt    string          `json:"window_started_at"`
	WindowEndsAt       string          `json:"window_ends_at"`
	MaxHP              int             `json:"max_hp"`
	MaxTurns           int             `json:"max_turns"`
	RecommendedPower   map[string]int  `json:"recommended_power"`
	BossStats          map[string]any  `json:"boss_stats"`
	MechanicTags       []string        `json:"mechanic_tags"`
	Skills             []BossSkillView `json:"skills"`
	RewardTiers        []RewardTier    `json:"reward_tiers"`
	QueueOpen          bool            `json:"queue_open"`
	SeasonTag          string          `json:"season_tag"`
}

type BossSkillView struct {
	SkillID        string   `json:"skill_id"`
	Name           string   `json:"name"`
	Phase          string   `json:"phase"`
	TargetPattern  string   `json:"target_pattern"`
	CooldownRounds int      `json:"cooldown_rounds"`
	HalfHPOnly     bool     `json:"half_hp_only"`
	MechanicTags   []string `json:"mechanic_tags"`
	Description    string   `json:"description"`
}

type RewardTier struct {
	Tier           string           `json:"tier"`
	RequiredDamage int              `json:"required_damage"`
	RewardGold     int              `json:"reward_gold"`
	MaterialDrops  []map[string]any `json:"material_drops"`
}

type QueueStatusView struct {
	BossID             string   `json:"boss_id"`
	CharacterID        string   `json:"character_id"`
	Queued             bool     `json:"queued"`
	CurrentQueuedCount int      `json:"current_queued_count"`
	RequiredPartySize  int      `json:"required_party_size"`
	LastRaidID         string   `json:"last_raid_id,omitempty"`
	LastRewardTier     string   `json:"last_reward_tier,omitempty"`
	PendingRaidIDs     []string `json:"pending_raid_ids,omitempty"`
}

type RaidMemberView struct {
	CharacterID string `json:"character_id"`
	Name        string `json:"name"`
	Class       string `json:"class"`
	DamageDealt int    `json:"damage_dealt"`
	Survived    bool   `json:"survived"`
	PowerScore  int    `json:"power_score"`
}

type RaidView struct {
	RaidID        string           `json:"raid_id"`
	BossID        string           `json:"boss_id"`
	BossName      string           `json:"boss_name"`
	Status        string           `json:"status"`
	TotalDamage   int              `json:"total_damage"`
	BossMaxHP     int              `json:"boss_max_hp"`
	RewardTier    string           `json:"reward_tier"`
	RewardPackage RewardTier       `json:"reward_package"`
	Members       []RaidMemberView `json:"members"`
	ResolvedAt    string           `json:"resolved_at"`
}

type JoinResult struct {
	Status              QueueStatusView `json:"status"`
	MatchedCharacterIDs []string        `json:"matched_character_ids,omitempty"`
	ResolvedRaid        *RaidView       `json:"resolved_raid,omitempty"`
}

type ParticipantSnapshot struct {
	Character characters.Summary
	Power     int
	Player    combat.Combatant
}

type queueEntry struct {
	CharacterID string
	Name        string
	QueuedAt    time.Time
}

type Service struct {
	mu          sync.Mutex
	clock       func() time.Time
	config      BossConfigView
	queue       []queueEntry
	raidsByID   map[string]RaidView
	raidsByChar map[string][]string
}

func NewService() *Service {
	config := currentBossConfigAt(time.Now())
	return &Service{
		clock:       time.Now,
		config:      config,
		queue:       make([]queueEntry, 0, queueTargetSize),
		raidsByID:   make(map[string]RaidView),
		raidsByChar: make(map[string][]string),
	}
}

func (s *Service) CurrentBoss() BossConfigView {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.refreshConfigLocked()
	return s.config
}

func (s *Service) QueueStatus(characterID string) QueueStatusView {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.refreshConfigLocked()
	return s.queueStatusLocked(characterID)
}

func (s *Service) GetRaid(characterID, raidID string) (RaidView, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	raid, ok := s.raidsByID[raidID]
	if !ok {
		return RaidView{}, false
	}
	if characterID == "" {
		return raid, true
	}
	for _, member := range raid.Members {
		if member.CharacterID == characterID {
			return raid, true
		}
	}
	return RaidView{}, false
}

func (s *Service) JoinQueue(characterID, name string) JoinResult {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.refreshConfigLocked()

	if !s.isQueuedLocked(characterID) {
		s.queue = append(s.queue, queueEntry{
			CharacterID: characterID,
			Name:        name,
			QueuedAt:    s.clock(),
		})
	}

	status := s.queueStatusLocked(characterID)
	if len(s.queue) < s.config.RequiredPartySize {
		return JoinResult{Status: status}
	}

	entries := append([]queueEntry(nil), s.queue[:s.config.RequiredPartySize]...)
	s.queue = append([]queueEntry(nil), s.queue[s.config.RequiredPartySize:]...)
	matchedIDs := make([]string, 0, len(entries))
	for _, entry := range entries {
		matchedIDs = append(matchedIDs, entry.CharacterID)
	}

	return JoinResult{
		Status:              s.queueStatusLocked(characterID),
		MatchedCharacterIDs: matchedIDs,
	}
}

func (s *Service) ResolveMatchedRaid(snapshots []ParticipantSnapshot) RaidView {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.refreshConfigLocked()
	raid := s.resolveRaidLocked(snapshots)
	s.raidsByID[raid.RaidID] = raid
	for _, member := range raid.Members {
		s.raidsByChar[member.CharacterID] = append([]string{raid.RaidID}, s.raidsByChar[member.CharacterID]...)
	}
	return raid
}

func (s *Service) refreshConfigLocked() {
	next := currentBossConfigAt(s.clock())
	if next.BossID != s.config.BossID {
		s.queue = make([]queueEntry, 0, queueTargetSize)
	}
	s.config = next
}

func (s *Service) isQueuedLocked(characterID string) bool {
	for _, entry := range s.queue {
		if entry.CharacterID == characterID {
			return true
		}
	}
	return false
}

func (s *Service) queueStatusLocked(characterID string) QueueStatusView {
	status := QueueStatusView{
		BossID:             s.config.BossID,
		CharacterID:        characterID,
		Queued:             s.isQueuedLocked(characterID),
		CurrentQueuedCount: len(s.queue),
		RequiredPartySize:  s.config.RequiredPartySize,
	}
	if raids := s.raidsByChar[characterID]; len(raids) > 0 {
		status.LastRaidID = raids[0]
		if raid, ok := s.raidsByID[raids[0]]; ok {
			status.LastRewardTier = raid.RewardTier
		}
	}
	return status
}

func (s *Service) resolveRaidLocked(snapshots []ParticipantSnapshot) RaidView {
	raidID := fmt.Sprintf("raid_%d", s.clock().UnixNano())
	boss := bossCombatant(s.config)
	players := make([]combat.Combatant, 0, len(snapshots))
	powerByCharacter := make(map[string]int, len(snapshots))
	for _, snap := range snapshots {
		player := snap.Player
		if player.EntityID == "" {
			player = combat.BaselineCombatant(snap.Character.Class)
			player.EntityID = snap.Character.CharacterID
			player.Name = snap.Character.Name
			player.Team = "a"
			player.IsPlayerSide = true
			player.CurrentHP = player.MaxHP
		}
		player.Team = "a"
		player.IsPlayerSide = true
		players = append(players, player)
		powerByCharacter[snap.Character.CharacterID] = snap.Power
	}
	result := combat.SimulateRaidBoss(combat.RaidBossConfig{
		BattleType: "world_boss",
		RunID:      raidID,
		MaxTurns:   s.config.MaxTurns,
		Players:    players,
		Boss:       boss,
	})
	members := make([]RaidMemberView, 0, len(snapshots))
	totalDamage := result.TotalDamage
	classByCharacter := make(map[string]string, len(snapshots))
	for _, snap := range snapshots {
		classByCharacter[snap.Character.CharacterID] = snap.Character.Class
	}
	for _, participant := range result.Participants {
		members = append(members, RaidMemberView{
			CharacterID: participant.EntityID,
			Name:        participant.Name,
			Class:       classByCharacter[participant.EntityID],
			DamageDealt: participant.DamageDealt,
			Survived:    participant.Survived,
			PowerScore:  powerByCharacter[participant.EntityID],
		})
	}
	sort.Slice(members, func(i, j int) bool {
		if members[i].DamageDealt != members[j].DamageDealt {
			return members[i].DamageDealt > members[j].DamageDealt
		}
		return members[i].CharacterID < members[j].CharacterID
	})
	reward := rewardTierForDamage(s.config, totalDamage)
	return RaidView{
		RaidID:        raidID,
		BossID:        s.config.BossID,
		BossName:      s.config.Name,
		Status:        "resolved",
		TotalDamage:   totalDamage,
		BossMaxHP:     s.config.MaxHP,
		RewardTier:    reward.Tier,
		RewardPackage: reward,
		Members:       members,
		ResolvedAt:    s.clock().Format(time.RFC3339),
	}
}

func currentBossConfigAt(now time.Time) BossConfigView {
	tiers := []RewardTier{
		{Tier: "D", RequiredDamage: 250, RewardGold: 90, MaterialDrops: []map[string]any{{"material_key": "reforge_stone", "quantity": 1}}},
		{Tier: "C", RequiredDamage: 550, RewardGold: 150, MaterialDrops: []map[string]any{{"material_key": "reforge_stone", "quantity": 2}}},
		{Tier: "B", RequiredDamage: 925, RewardGold: 230, MaterialDrops: []map[string]any{{"material_key": "reforge_stone", "quantity": 3}}},
		{Tier: "A", RequiredDamage: 1400, RewardGold: 340, MaterialDrops: []map[string]any{{"material_key": "reforge_stone", "quantity": 5}}},
		{Tier: "S", RequiredDamage: 2100, RewardGold: 500, MaterialDrops: []map[string]any{{"material_key": "reforge_stone", "quantity": 8}}},
	}
	windowStart := rotationWindowStart(now)
	windowEnd := windowStart.AddDate(0, 0, bossRotationWindowDays)
	roster := worldBossRoster()
	selected := roster[deterministicIndex("worldboss:"+windowStart.Format("2006-01-02"), len(roster))]
	selected.RotationWindowDays = bossRotationWindowDays
	selected.WindowStartedAt = windowStart.Format(time.RFC3339)
	selected.WindowEndsAt = windowEnd.Format(time.RFC3339)
	selected.RewardTiers = tiers
	selected.QueueOpen = true
	selected.SeasonTag = "season_1"
	return selected
}

func rewardTierForDamage(config BossConfigView, damage int) RewardTier {
	reached := config.RewardTiers[0]
	for _, tier := range config.RewardTiers {
		if damage >= tier.RequiredDamage {
			reached = tier
		}
	}
	return reached
}

func bossCombatant(config BossConfigView) combat.Combatant {
	physAtk, _ := config.BossStats["physical_attack"].(int)
	magAtk, _ := config.BossStats["magic_attack"].(int)
	physDef, _ := config.BossStats["physical_defense"].(int)
	magDef, _ := config.BossStats["magic_defense"].(int)
	speed, _ := config.BossStats["speed"].(int)
	blockRate, _ := config.BossStats["block_rate"].(float64)
	critRate, _ := config.BossStats["crit_rate"].(float64)
	critDamage, _ := config.BossStats["crit_damage"].(float64)
	return combat.Combatant{
		EntityID:       config.BossID,
		Name:           config.Name,
		Team:           "b",
		IsPlayerSide:   false,
		Class:          "warrior",
		Role:           "boss",
		MaxHP:          config.MaxHP,
		CurrentHP:      config.MaxHP,
		PhysAtk:        physAtk,
		MagAtk:         magAtk,
		PhysDef:        physDef,
		MagDef:         magDef,
		Speed:          speed,
		BlockRate:      blockRate,
		CritRate:       critRate,
		CritDamage:     critDamage,
		EquippedSkills: worldBossSkillActions(config),
		SkillCooldowns: map[string]int{},
	}
}

func recommendedWorldBossPower() map[string]int {
	return map[string]int{
		"per_player_floor": 6400,
		"per_player_mid":   7600,
		"per_player_high":  9000,
		"party_floor":      38400,
		"party_mid":        45600,
		"party_high":       54000,
	}
}

func worldBossRoster() []BossConfigView {
	return []BossConfigView{
		{
			OriginDungeonID:   "ancient_catacomb_v1",
			BossID:            "gravewake_overlord",
			Name:              "Gravewake Overlord",
			RequiredPartySize: queueTargetSize,
			MaxHP:             2200,
			MaxTurns:          15,
			RecommendedPower:  recommendedWorldBossPower(),
			BossStats:         map[string]any{"physical_attack": 44, "magic_attack": 34, "physical_defense": 28, "magic_defense": 28, "speed": 16, "block_rate": 0.12, "crit_rate": 0.06, "crit_damage": 0.25},
			MechanicTags:      []string{"guard_break_check", "sustain_pressure", "half_hp_ultimate"},
			Skills:            []BossSkillView{{SkillID: "sepulcher_slam", Name: "Sepulcher Slam", Phase: "all", TargetPattern: "single", CooldownRounds: 2, MechanicTags: []string{"heavy_physical", "defense_break"}, Description: "Crushes one target with heavy physical damage and a short defense break."}, {SkillID: "cryptward_bulwark", Name: "Cryptward Bulwark", Phase: "all", TargetPattern: "self", CooldownRounds: 4, MechanicTags: []string{"shield", "block_up"}, Description: "Raises a stone barrier that increases block and absorbs burst damage."}, {SkillID: "grave_tide", Name: "Grave Tide", Phase: "all", TargetPattern: "all_players", CooldownRounds: 5, MechanicTags: []string{"chip_aoe", "healing_reduction"}, Description: "A grave-cold wave hits every participant for moderate damage and suppresses healing briefly."}, {SkillID: "catacomb_annihilation", Name: "Catacomb Annihilation", Phase: "half_hp", TargetPattern: "all_players", CooldownRounds: 8, HalfHPOnly: true, MechanicTags: []string{"ultimate", "full_raid_hit"}, Description: "Below half HP, the Overlord collapses the catacomb and deals massive raid-wide damage to all participants."}},
		},
		{
			OriginDungeonID:   "thorned_hollow_v1",
			BossID:            "briarqueen_predator",
			Name:              "Briarqueen Predator",
			RequiredPartySize: queueTargetSize,
			MaxHP:             2050,
			MaxTurns:          15,
			RecommendedPower:  recommendedWorldBossPower(),
			BossStats:         map[string]any{"physical_attack": 46, "magic_attack": 30, "physical_defense": 22, "magic_defense": 22, "speed": 21, "block_rate": 0.06, "crit_rate": 0.14, "crit_damage": 0.35},
			MechanicTags:      []string{"accuracy_check", "crit_pressure", "half_hp_ultimate"},
			Skills:            []BossSkillView{{SkillID: "thornrush_pounce", Name: "Thornrush Pounce", Phase: "all", TargetPattern: "single", CooldownRounds: 2, MechanicTags: []string{"precision_strike", "backline_pressure"}, Description: "Leaps onto one target with a high-precision critical strike."}, {SkillID: "needleburst_fan", Name: "Needleburst Fan", Phase: "all", TargetPattern: "all_players", CooldownRounds: 4, MechanicTags: []string{"light_aoe", "evasion_down"}, Description: "Scatters thorn needles across the battlefield and lowers evasion briefly."}, {SkillID: "hunter_focus", Name: "Hunter Focus", Phase: "all", TargetPattern: "self", CooldownRounds: 4, MechanicTags: []string{"crit_up", "precision_up"}, Description: "Sharpens the boss's hunt state and raises crit pressure for several turns."}, {SkillID: "crown_of_thorns", Name: "Crown of Thorns", Phase: "half_hp", TargetPattern: "all_players", CooldownRounds: 8, HalfHPOnly: true, MechanicTags: []string{"ultimate", "full_raid_hit"}, Description: "Below half HP, the predator erupts into a field-wide thorn storm that damages every participant."}},
		},
		{
			OriginDungeonID:   "sunscar_warvault_v1",
			BossID:            "sunscar_warmarshal",
			Name:              "Sunscar Warmarshal",
			RequiredPartySize: queueTargetSize,
			MaxHP:             2100,
			MaxTurns:          15,
			RecommendedPower:  recommendedWorldBossPower(),
			BossStats:         map[string]any{"physical_attack": 52, "magic_attack": 26, "physical_defense": 24, "magic_defense": 20, "speed": 18, "block_rate": 0.08, "crit_rate": 0.10, "crit_damage": 0.30},
			MechanicTags:      []string{"burst_window", "armor_break", "half_hp_ultimate"},
			Skills:            []BossSkillView{{SkillID: "warvault_cleave", Name: "Warvault Cleave", Phase: "all", TargetPattern: "target_3", CooldownRounds: 3, MechanicTags: []string{"multi_hit", "physical_burst"}, Description: "Sweeps through up to three participants with a heavy physical cleave."}, {SkillID: "sunder_standard", Name: "Sunder Standard", Phase: "all", TargetPattern: "all_players", CooldownRounds: 5, MechanicTags: []string{"armor_break", "raid_pressure"}, Description: "Plants a blazing war-banner that weakens the entire party's defenses."}, {SkillID: "march_of_iron", Name: "March of Iron", Phase: "all", TargetPattern: "self", CooldownRounds: 4, MechanicTags: []string{"attack_up", "speed_up"}, Description: "Accelerates into a battle-rhythm that raises attack and tempo."}, {SkillID: "sunfall_bombardment", Name: "Sunfall Bombardment", Phase: "half_hp", TargetPattern: "all_players", CooldownRounds: 8, HalfHPOnly: true, MechanicTags: []string{"ultimate", "full_raid_hit"}, Description: "Below half HP, artillery sigils detonate across the whole arena and hit every participant."}},
		},
		{
			OriginDungeonID:   "obsidian_spire_v1",
			BossID:            "obsidian_archon",
			Name:              "Obsidian Archon",
			RequiredPartySize: queueTargetSize,
			MaxHP:             2080,
			MaxTurns:          15,
			RecommendedPower:  recommendedWorldBossPower(),
			BossStats:         map[string]any{"physical_attack": 30, "magic_attack": 54, "physical_defense": 21, "magic_defense": 26, "speed": 19, "block_rate": 0.05, "crit_rate": 0.10, "crit_damage": 0.35},
			MechanicTags:      []string{"magic_pressure", "silence_check", "half_hp_ultimate"},
			Skills:            []BossSkillView{{SkillID: "void_lance", Name: "Void Lance", Phase: "all", TargetPattern: "single", CooldownRounds: 2, MechanicTags: []string{"magic_burst", "anti_mage"}, Description: "Launches a condensed void spear at one participant for severe magic damage."}, {SkillID: "eclipse_field", Name: "Eclipse Field", Phase: "all", TargetPattern: "all_players", CooldownRounds: 5, MechanicTags: []string{"raid_dot", "silence_pressure"}, Description: "Darkens the arena and washes the whole team with magic damage plus brief casting disruption."}, {SkillID: "blackglass_mirror", Name: "Blackglass Mirror", Phase: "all", TargetPattern: "self", CooldownRounds: 4, MechanicTags: []string{"magic_shield", "anti_burst"}, Description: "Wraps the Archon in mirrored obsidian that softens incoming magic bursts."}, {SkillID: "spires_end_requiem", Name: "Spire's End Requiem", Phase: "half_hp", TargetPattern: "all_players", CooldownRounds: 8, HalfHPOnly: true, MechanicTags: []string{"ultimate", "full_raid_hit"}, Description: "Below half HP, the Archon releases a delayed arena-wide blast that hits every participant."}},
		},
	}
}

func rotationWindowStart(now time.Time) time.Time {
	local := now.In(now.Location())
	dayStart := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, local.Location())
	anchor := time.Date(2026, time.April, 1, 0, 0, 0, 0, local.Location())
	diffDays := int(dayStart.Sub(anchor).Hours() / 24)
	if diffDays < 0 {
		diffDays = 0
	}
	windowIndex := diffDays / bossRotationWindowDays
	return anchor.AddDate(0, 0, windowIndex*bossRotationWindowDays)
}

func deterministicIndex(seed string, size int) int {
	if size <= 0 {
		return 0
	}
	h := fnv.New32a()
	_, _ = h.Write([]byte(seed))
	return int(h.Sum32() % uint32(size))
}

func worldBossSkillActions(config BossConfigView) []combat.SkillAction {
	actions := make([]combat.SkillAction, 0, len(config.Skills))
	for _, skill := range config.Skills {
		switch skill.SkillID {
		case "sepulcher_slam":
			actions = append(actions, combat.SkillAction{SkillID: skill.SkillID, ActionKind: "damage", DamageType: "physical", CooldownRounds: skill.CooldownRounds, PowerScale: 1.45, DebuffFamily: "shred", DebuffValue: 0.12, Level: 1, HalfHPOnly: skill.HalfHPOnly})
		case "cryptward_bulwark":
			actions = append(actions, combat.SkillAction{SkillID: skill.SkillID, ActionKind: "buff", CooldownRounds: skill.CooldownRounds, BuffFamily: "def", BuffValue: 0.22, ShieldScale: 0.12, Level: 1, HalfHPOnly: skill.HalfHPOnly})
		case "grave_tide":
			actions = append(actions, combat.SkillAction{SkillID: skill.SkillID, ActionKind: "damage", DamageType: "magic", CooldownRounds: skill.CooldownRounds, PowerScale: 1.05, DebuffFamily: "weaken", DebuffValue: 0.10, Level: 1, HalfHPOnly: skill.HalfHPOnly})
		case "catacomb_annihilation":
			actions = append(actions, combat.SkillAction{SkillID: skill.SkillID, ActionKind: "damage", DamageType: "magic", Tier: "ultimate", CooldownRounds: skill.CooldownRounds, PowerScale: 1.90, Level: 1, HalfHPOnly: true})
		case "thornrush_pounce":
			actions = append(actions, combat.SkillAction{SkillID: skill.SkillID, ActionKind: "damage", DamageType: "physical", CooldownRounds: skill.CooldownRounds, PowerScale: 1.38, Level: 1, HalfHPOnly: skill.HalfHPOnly})
		case "needleburst_fan":
			actions = append(actions, combat.SkillAction{SkillID: skill.SkillID, ActionKind: "damage", DamageType: "physical", CooldownRounds: skill.CooldownRounds, PowerScale: 1.02, DebuffFamily: "slow", DebuffValue: 0.10, Level: 1, HalfHPOnly: skill.HalfHPOnly})
		case "hunter_focus":
			actions = append(actions, combat.SkillAction{SkillID: skill.SkillID, ActionKind: "buff", CooldownRounds: skill.CooldownRounds, BuffFamily: "atk", BuffValue: 0.20, Level: 1, HalfHPOnly: skill.HalfHPOnly})
		case "crown_of_thorns":
			actions = append(actions, combat.SkillAction{SkillID: skill.SkillID, ActionKind: "damage", DamageType: "physical", Tier: "ultimate", CooldownRounds: skill.CooldownRounds, PowerScale: 1.82, Level: 1, HalfHPOnly: true})
		case "warvault_cleave":
			actions = append(actions, combat.SkillAction{SkillID: skill.SkillID, ActionKind: "damage", DamageType: "physical", CooldownRounds: skill.CooldownRounds, PowerScale: 1.48, Level: 1, HalfHPOnly: skill.HalfHPOnly})
		case "sunder_standard":
			actions = append(actions, combat.SkillAction{SkillID: skill.SkillID, ActionKind: "debuff", CooldownRounds: skill.CooldownRounds, DebuffFamily: "shred", DebuffValue: 0.16, Level: 1, HalfHPOnly: skill.HalfHPOnly})
		case "march_of_iron":
			actions = append(actions, combat.SkillAction{SkillID: skill.SkillID, ActionKind: "buff", CooldownRounds: skill.CooldownRounds, BuffFamily: "atk", BuffValue: 0.18, Level: 1, HalfHPOnly: skill.HalfHPOnly})
		case "sunfall_bombardment":
			actions = append(actions, combat.SkillAction{SkillID: skill.SkillID, ActionKind: "damage", DamageType: "physical", Tier: "ultimate", CooldownRounds: skill.CooldownRounds, PowerScale: 1.95, Level: 1, HalfHPOnly: true})
		case "void_lance":
			actions = append(actions, combat.SkillAction{SkillID: skill.SkillID, ActionKind: "damage", DamageType: "magic", CooldownRounds: skill.CooldownRounds, PowerScale: 1.46, Level: 1, HalfHPOnly: skill.HalfHPOnly})
		case "eclipse_field":
			actions = append(actions, combat.SkillAction{SkillID: skill.SkillID, ActionKind: "damage", DamageType: "magic", CooldownRounds: skill.CooldownRounds, PowerScale: 1.08, SilenceRounds: 1, Level: 1, HalfHPOnly: skill.HalfHPOnly})
		case "blackglass_mirror":
			actions = append(actions, combat.SkillAction{SkillID: skill.SkillID, ActionKind: "buff", CooldownRounds: skill.CooldownRounds, BuffFamily: "def", BuffValue: 0.18, ShieldScale: 0.10, Level: 1, HalfHPOnly: skill.HalfHPOnly})
		case "spires_end_requiem":
			actions = append(actions, combat.SkillAction{SkillID: skill.SkillID, ActionKind: "damage", DamageType: "magic", Tier: "ultimate", CooldownRounds: skill.CooldownRounds, PowerScale: 1.92, SilenceRounds: 1, Level: 1, HalfHPOnly: true})
		}
	}
	return actions
}
