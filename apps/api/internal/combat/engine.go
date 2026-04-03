package combat

import (
	"fmt"
	"hash/fnv"
	"strings"
)

const (
	maxTurnsDefault    = 24
	RunPotionCapPerRun = 3
)

// Combatant represents one fighter in a battle.
type Combatant struct {
	EntityID     string
	Name         string
	Team         string // "a" or "b"
	IsPlayerSide bool   // true = heal+burst+basic+potions; false = burst+basic, no potions
	Class        string // "warrior" | "mage" | "priest"
	Role         string

	MaxHP   int
	PhysAtk int
	MagAtk  int
	PhysDef int
	MagDef  int
	Speed   int
	HealPow int

	CritRate        float64
	CritDamage      float64
	BlockRate       float64
	Precision       float64
	EvasionRate     float64
	PhysicalMastery float64
	MagicMastery    float64

	// Runtime state. CurrentHP must be set by caller; others start at zero.
	CurrentHP      int
	ShieldHP       int
	BurstCD        int
	HealCD         int
	EquippedSkills []SkillAction
	SkillCooldowns map[string]int
	ActiveBuffs    []StatusBuff
	PotionBag      []PotionItem
	SummonSkillID  string
	SummonTurns    int
	SummonStrike   float64
	SummonMend     float64
	BasicSkillID   string
}

type SkillAction struct {
	SkillID        string
	ActionKind     string
	DamageType     string
	Tier           string
	CooldownRounds int
	PowerScale     float64
	HealScale      float64
	TargetPattern  string
	BuffFamily     string
	BuffValue      float64
	DebuffFamily   string
	DebuffValue    float64
	ShieldScale    float64
	RegenScale     float64
	SilenceRounds  int
	SummonTurns    int
	SummonStrike   float64
	SummonMend     float64
	Level          int
	HalfHPOnly     bool
}

// StatusBuff is a temporary combat buff (applied by a potion).
type StatusBuff struct {
	Family    string // "atk" | "def" | "spd"
	PotionID  string
	Value     float64 // multiplicative bonus, e.g. 0.16 = +16 %
	Remaining int     // rounds until expiry, decremented at round end
}

// PotionItem is one potion type with a remaining quantity.
type PotionItem struct {
	PotionID string
	Family   string // "hp" | "atk" | "def" | "spd"
	Tier     int

	// HP potion
	HealPct float64 // fraction of MaxHP to restore, e.g. 0.25
	HealCap int     // upper cap on heal amount

	// Buff potion (atk / def / spd)
	BuffValue float64 // multiplicative bonus
	Duration  int     // rounds the buff lasts

	Quantity int
}

// BattleConfig defines one battle invocation.
type BattleConfig struct {
	BattleType string // "dungeon_wave" | "arena_duel"
	RunID      string
	RoomIndex  int  // used in log events
	IsBossRoom bool // triggers ATK-potion auto-use at turn 1

	MaxTurns int // 0 = default 24

	SideA Combatant
	SideB Combatant

	// Potions already consumed by each side in earlier dungeon rooms.
	// The engine adds its own consumption on top of these.
	RunPotionUsedA int
	RunPotionUsedB int
}

// BattleResult is the output of SimulateBattle.
type BattleResult struct {
	SideAWon         bool
	SideAFinalHP     int
	SideBFinalHP     int
	PotionsConsumedA int
	PotionsConsumedB int
	Log              []map[string]any
}

type SkirmishConfig struct {
	BattleType string
	RunID      string
	RoomIndex  int
	IsBossRoom bool
	MaxTurns   int

	Player  Combatant
	Enemies []Combatant

	RunPotionUsed int
}

type SkirmishResult struct {
	PlayerWon        bool
	PlayerFinalHP    int
	EnemiesRemaining int
	PotionsConsumed  int
	Log              []map[string]any
}

type RaidBossConfig struct {
	BattleType string
	RunID      string
	MaxTurns   int
	Players    []Combatant
	Boss       Combatant
}

type RaidParticipantResult struct {
	EntityID    string
	Name        string
	FinalHP     int
	DamageDealt int
	Survived    bool
}

type RaidBossResult struct {
	BossDefeated bool
	BossFinalHP  int
	TotalDamage  int
	Participants []RaidParticipantResult
	Log          []map[string]any
}

// BaselineCombatant returns a Combatant pre-filled with class-based stats.
// Caller must set EntityID, Name, Team, IsPlayerSide, CurrentHP, and PotionBag.
func BaselineCombatant(class string) Combatant {
	switch class {
	case "warrior":
		return Combatant{Class: class, MaxHP: 132, PhysAtk: 28, MagAtk: 6, PhysDef: 18, MagDef: 10, Speed: 10, HealPow: 4, CritRate: 0.20, CritDamage: 0.50, BlockRate: 0.05}
	case "mage":
		return Combatant{Class: class, MaxHP: 92, PhysAtk: 11, MagAtk: 34, PhysDef: 9, MagDef: 18, Speed: 16, HealPow: 8, CritRate: 0.20, CritDamage: 0.50, BlockRate: 0.05}
	case "priest":
		return Combatant{Class: class, MaxHP: 104, PhysAtk: 10, MagAtk: 26, PhysDef: 11, MagDef: 17, Speed: 14, HealPow: 20, CritRate: 0.20, CritDamage: 0.50, BlockRate: 0.05}
	default:
		return Combatant{Class: class, MaxHP: 100, PhysAtk: 16, MagAtk: 16, PhysDef: 12, MagDef: 12, Speed: 12, HealPow: 6, CritRate: 0.20, CritDamage: 0.50, BlockRate: 0.05}
	}
}

// SimulateBattle runs a complete auto-resolved battle and returns the result.
func SimulateBattle(cfg BattleConfig) BattleResult {
	a := cfg.SideA
	b := cfg.SideB

	if a.CurrentHP <= 0 {
		a.CurrentHP = a.MaxHP
	}
	if b.CurrentHP <= 0 {
		b.CurrentHP = b.MaxHP
	}

	maxTurns := cfg.MaxTurns
	if maxTurns <= 0 {
		maxTurns = maxTurnsDefault
	}

	potionConsumedA := 0
	potionConsumedB := 0

	log := make([]map[string]any, 0, maxTurns*8)
	log = append(log, map[string]any{
		"step":                  "room_start",
		"event_type":            "room_start",
		"room_index":            cfg.RoomIndex,
		"turn":                  0,
		"player_hp":             a.CurrentHP,
		"enemy_hp":              b.CurrentHP,
		"enemy_name":            b.Name,
		"player_speed":          effectiveSpeed(a),
		"enemy_speed":           effectiveSpeed(b),
		"cooldown_before_round": roundCooldownMap(a, b),
		"cooldown_after_round":  roundCooldownMap(a, b),
		"message":               "auto battle room started",
	})

	var finalTurn int
	for turn := 1; turn <= maxTurns && a.CurrentHP > 0 && b.CurrentHP > 0; turn++ {
		finalTurn = turn

		aFirst := effectiveSpeed(a) >= effectiveSpeed(b)
		initiative := []string{"b", "a"}
		if aFirst {
			initiative = []string{"a", "b"}
		}

		log = append(log, map[string]any{
			"step":                  "turn_start",
			"event_type":            "turn_start",
			"room_index":            cfg.RoomIndex,
			"turn":                  turn,
			"actor":                 "system",
			"initiative_order":      initiative,
			"player_hp":             a.CurrentHP,
			"enemy_hp":              b.CurrentHP,
			"cooldown_before_round": roundCooldownMap(a, b),
			"message":               "turn begins",
		})

		budgetA := RunPotionCapPerRun - cfg.RunPotionUsedA - potionConsumedA
		budgetB := RunPotionCapPerRun - cfg.RunPotionUsedB - potionConsumedB

		if aFirst {
			ca, newLog := actOnce(&a, &b, cfg, turn, budgetA, log)
			log = newLog
			potionConsumedA += ca
			if a.CurrentHP > 0 && b.CurrentHP > 0 {
				cb, newLog := actOnce(&b, &a, cfg, turn, budgetB, log)
				log = newLog
				potionConsumedB += cb
			}
		} else {
			cb, newLog := actOnce(&b, &a, cfg, turn, budgetB, log)
			log = newLog
			potionConsumedB += cb
			if a.CurrentHP > 0 && b.CurrentHP > 0 {
				ca, newLog := actOnce(&a, &b, cfg, turn, budgetA, log)
				log = newLog
				potionConsumedA += ca
			}
		}

		// If battle ended mid-turn, skip tick and turn_end.
		if a.CurrentHP <= 0 || b.CurrentHP <= 0 {
			break
		}

		// Tick status buff durations, decrement cooldowns.
		tickStatusBuffs(&a)
		tickStatusBuffs(&b)
		tickSummon(&a)
		tickSummon(&b)
		a.BurstCD = imax(0, a.BurstCD-1)
		a.HealCD = imax(0, a.HealCD-1)
		b.BurstCD = imax(0, b.BurstCD-1)
		b.HealCD = imax(0, b.HealCD-1)
		tickSkillCooldowns(&a)
		tickSkillCooldowns(&b)

		log = append(log, map[string]any{
			"step":                 "turn_end",
			"event_type":           "turn_end",
			"room_index":           cfg.RoomIndex,
			"turn":                 turn,
			"actor":                "system",
			"player_hp":            a.CurrentHP,
			"enemy_hp":             b.CurrentHP,
			"cooldown_after_round": roundCooldownMap(a, b),
			"message":              "turn ended",
		})
	}

	if b.CurrentHP <= 0 {
		log = append(log, map[string]any{
			"step":       "room_end",
			"event_type": "room_end",
			"room_index": cfg.RoomIndex,
			"turn":       finalTurn,
			"result":     "cleared",
			"player_hp":  imax(0, a.CurrentHP),
			"enemy_hp":   0,
			"message":    "room cleared",
		})
		return BattleResult{
			SideAWon:         true,
			SideAFinalHP:     imax(1, a.CurrentHP),
			SideBFinalHP:     0,
			PotionsConsumedA: potionConsumedA,
			PotionsConsumedB: potionConsumedB,
			Log:              log,
		}
	}

	result := "failed"
	if a.CurrentHP > 0 && b.CurrentHP > 0 {
		result = "timeout"
	}
	log = append(log, map[string]any{
		"step":       "room_end",
		"event_type": "room_end",
		"room_index": cfg.RoomIndex,
		"turn":       finalTurn,
		"result":     result,
		"player_hp":  imax(0, a.CurrentHP),
		"enemy_hp":   imax(0, b.CurrentHP),
		"message":    "room failed",
	})
	return BattleResult{
		SideAWon:         false,
		SideAFinalHP:     imax(0, a.CurrentHP),
		SideBFinalHP:     imax(0, b.CurrentHP),
		PotionsConsumedA: potionConsumedA,
		PotionsConsumedB: potionConsumedB,
		Log:              log,
	}
}

func SimulateSkirmish(cfg SkirmishConfig) SkirmishResult {
	player := cfg.Player
	if player.CurrentHP <= 0 {
		player.CurrentHP = player.MaxHP
	}
	enemies := cloneCombatants(cfg.Enemies)
	for i := range enemies {
		if enemies[i].CurrentHP <= 0 {
			enemies[i].CurrentHP = enemies[i].MaxHP
		}
	}

	maxTurns := cfg.MaxTurns
	if maxTurns <= 0 {
		maxTurns = maxTurnsDefault
	}

	potionsConsumed := 0
	log := []map[string]any{
		{
			"step":        "room_start",
			"event_type":  "room_start",
			"room_index":  cfg.RoomIndex,
			"turn":        0,
			"player_hp":   player.CurrentHP,
			"enemy_count": countAlive(enemies),
			"message":     "multi-unit battle room started",
		},
	}

	for turn := 1; turn <= maxTurns && player.CurrentHP > 0 && countAlive(enemies) > 0; turn++ {
		log = append(log, map[string]any{
			"step":        "turn_start",
			"event_type":  "turn_start",
			"room_index":  cfg.RoomIndex,
			"turn":        turn,
			"actor":       "system",
			"player_hp":   player.CurrentHP,
			"enemy_count": countAlive(enemies),
			"message":     "turn begins",
		})

		budget := RunPotionCapPerRun - cfg.RunPotionUsed - potionsConsumed
		if budget > 0 && countAlive(enemies) > 0 {
			primary := choosePrimaryEnemyTarget(enemies)
			if primary >= 0 {
				c, newLog := tryUsePotion(&player, &enemies[primary], BattleConfig{RunID: cfg.RunID, RoomIndex: cfg.RoomIndex, IsBossRoom: cfg.IsBossRoom}, turn, log)
				potionsConsumed += c
				log = newLog
			}
		}

		if player.CurrentHP > 0 && countAlive(enemies) > 0 {
			log = applyStartTurnEffectsMulti(&player, enemies, cfg, turn, log)
			if player.CurrentHP > 0 && countAlive(enemies) > 0 {
				log = executePlayerSkirmishTurn(&player, enemies, cfg, turn, log)
			}
		}

		for _, enemyIndex := range enemyInitiativeOrder(enemies) {
			if player.CurrentHP <= 0 {
				break
			}
			if !isAlive(enemies[enemyIndex]) {
				continue
			}
			log = applyStartTurnEffects(&enemies[enemyIndex], &player, BattleConfig{RunID: cfg.RunID, RoomIndex: cfg.RoomIndex}, turn, log)
			if player.CurrentHP <= 0 {
				break
			}
			log = executeAction(&enemies[enemyIndex], &player, BattleConfig{BattleType: cfg.BattleType, RunID: cfg.RunID, RoomIndex: cfg.RoomIndex}, turn, log)
		}

		if player.CurrentHP <= 0 || countAlive(enemies) == 0 {
			break
		}

		tickStatusBuffs(&player)
		tickSummon(&player)
		tickSkillCooldowns(&player)
		player.BurstCD = imax(0, player.BurstCD-1)
		player.HealCD = imax(0, player.HealCD-1)
		for i := range enemies {
			if !isAlive(enemies[i]) {
				continue
			}
			tickStatusBuffs(&enemies[i])
			tickSummon(&enemies[i])
			tickSkillCooldowns(&enemies[i])
			enemies[i].BurstCD = imax(0, enemies[i].BurstCD-1)
			enemies[i].HealCD = imax(0, enemies[i].HealCD-1)
		}
		log = append(log, map[string]any{
			"step":        "turn_end",
			"event_type":  "turn_end",
			"room_index":  cfg.RoomIndex,
			"turn":        turn,
			"actor":       "system",
			"player_hp":   player.CurrentHP,
			"enemy_count": countAlive(enemies),
			"message":     "turn ended",
		})
	}

	result := SkirmishResult{
		PlayerWon:        player.CurrentHP > 0 && countAlive(enemies) == 0,
		PlayerFinalHP:    imax(0, player.CurrentHP),
		EnemiesRemaining: countAlive(enemies),
		PotionsConsumed:  potionsConsumed,
		Log:              log,
	}
	log = append(log, map[string]any{
		"step":        "room_end",
		"event_type":  "room_end",
		"room_index":  cfg.RoomIndex,
		"result":      map[bool]string{true: "cleared", false: "failed"}[result.PlayerWon],
		"player_hp":   result.PlayerFinalHP,
		"enemy_count": result.EnemiesRemaining,
		"message":     "multi-unit battle finished",
	})
	result.Log = log
	return result
}

func SimulateRaidBoss(cfg RaidBossConfig) RaidBossResult {
	players := cloneCombatants(cfg.Players)
	boss := cfg.Boss
	if boss.CurrentHP <= 0 {
		boss.CurrentHP = boss.MaxHP
	}
	for i := range players {
		if players[i].CurrentHP <= 0 {
			players[i].CurrentHP = players[i].MaxHP
		}
		if players[i].SkillCooldowns == nil {
			players[i].SkillCooldowns = map[string]int{}
		}
	}
	if boss.SkillCooldowns == nil {
		boss.SkillCooldowns = map[string]int{}
	}

	maxTurns := cfg.MaxTurns
	if maxTurns <= 0 {
		maxTurns = maxTurnsDefault
	}

	potionsConsumed := make(map[string]int, len(players))
	damageByPlayer := make(map[string]int, len(players))
	log := []map[string]any{{
		"step":        "raid_start",
		"event_type":  "raid_start",
		"turn":        0,
		"boss_hp":     boss.CurrentHP,
		"alive_count": alivePlayers(players),
		"message":     "raid boss battle started",
	}}

	for turn := 1; turn <= maxTurns && boss.CurrentHP > 0 && alivePlayers(players) > 0; turn++ {
		log = append(log, map[string]any{
			"step":        "turn_start",
			"event_type":  "turn_start",
			"turn":        turn,
			"boss_hp":     boss.CurrentHP,
			"alive_count": alivePlayers(players),
			"message":     "raid turn begins",
		})

		playerOrder := raidPlayerOrder(players)
		for _, index := range playerOrder {
			if boss.CurrentHP <= 0 {
				break
			}
			if !isAlive(players[index]) {
				continue
			}
			log = applyStartTurnEffects(&players[index], &boss, BattleConfig{BattleType: cfg.BattleType, RunID: cfg.RunID, RoomIndex: 1, IsBossRoom: true}, turn, log)
			if players[index].CurrentHP <= 0 || boss.CurrentHP <= 0 {
				continue
			}
			budget := RunPotionCapPerRun - potionsConsumed[players[index].EntityID]
			if budget > 0 {
				var consumed int
				consumed, log = tryUsePotion(&players[index], &boss, BattleConfig{BattleType: cfg.BattleType, RunID: cfg.RunID, RoomIndex: 1, IsBossRoom: true}, turn, log)
				potionsConsumed[players[index].EntityID] += consumed
			}
			beforeBossHP := boss.CurrentHP
			log = executeAction(&players[index], &boss, BattleConfig{BattleType: cfg.BattleType, RunID: cfg.RunID, RoomIndex: 1, IsBossRoom: true}, turn, log)
			if beforeBossHP > boss.CurrentHP {
				damageByPlayer[players[index].EntityID] += beforeBossHP - boss.CurrentHP
			}
		}

		if boss.CurrentHP > 0 && alivePlayers(players) > 0 {
			primaryIndex := chooseRaidPrimaryTarget(players)
			if primaryIndex >= 0 {
				log = applyStartTurnEffects(&boss, &players[primaryIndex], BattleConfig{BattleType: cfg.BattleType, RunID: cfg.RunID, RoomIndex: 1, IsBossRoom: true}, turn, log)
				if boss.CurrentHP > 0 && alivePlayers(players) > 0 {
					log = executeBossRaidTurn(&boss, players, cfg, turn, log)
				}
			}
		}

		if boss.CurrentHP <= 0 || alivePlayers(players) == 0 {
			break
		}

		tickStatusBuffs(&boss)
		tickSummon(&boss)
		tickSkillCooldowns(&boss)
		boss.BurstCD = imax(0, boss.BurstCD-1)
		boss.HealCD = imax(0, boss.HealCD-1)
		for i := range players {
			if !isAlive(players[i]) {
				continue
			}
			tickStatusBuffs(&players[i])
			tickSummon(&players[i])
			tickSkillCooldowns(&players[i])
			players[i].BurstCD = imax(0, players[i].BurstCD-1)
			players[i].HealCD = imax(0, players[i].HealCD-1)
		}

		log = append(log, map[string]any{
			"step":        "turn_end",
			"event_type":  "turn_end",
			"turn":        turn,
			"boss_hp":     boss.CurrentHP,
			"alive_count": alivePlayers(players),
			"message":     "raid turn ended",
		})
	}

	totalDamage := boss.MaxHP - boss.CurrentHP
	if totalDamage < 0 {
		totalDamage = 0
	}
	results := make([]RaidParticipantResult, 0, len(players))
	for _, player := range players {
		results = append(results, RaidParticipantResult{
			EntityID:    player.EntityID,
			Name:        player.Name,
			FinalHP:     imax(0, player.CurrentHP),
			DamageDealt: damageByPlayer[player.EntityID],
			Survived:    player.CurrentHP > 0,
		})
	}

	log = append(log, map[string]any{
		"step":        "raid_end",
		"event_type":  "raid_end",
		"boss_hp":     boss.CurrentHP,
		"alive_count": alivePlayers(players),
		"result":      map[bool]string{true: "boss_defeated", false: "raid_finished"}[boss.CurrentHP <= 0],
		"message":     "raid boss battle finished",
	})

	return RaidBossResult{
		BossDefeated: boss.CurrentHP <= 0,
		BossFinalHP:  imax(0, boss.CurrentHP),
		TotalDamage:  totalDamage,
		Participants: results,
		Log:          log,
	}
}

func cloneCombatants(items []Combatant) []Combatant {
	cloned := make([]Combatant, len(items))
	for i, item := range items {
		cloned[i] = item
		if len(item.EquippedSkills) > 0 {
			cloned[i].EquippedSkills = append([]SkillAction(nil), item.EquippedSkills...)
		}
		if len(item.ActiveBuffs) > 0 {
			cloned[i].ActiveBuffs = append([]StatusBuff(nil), item.ActiveBuffs...)
		}
		if len(item.PotionBag) > 0 {
			cloned[i].PotionBag = append([]PotionItem(nil), item.PotionBag...)
		}
		if item.SkillCooldowns != nil {
			cloned[i].SkillCooldowns = make(map[string]int, len(item.SkillCooldowns))
			for key, value := range item.SkillCooldowns {
				cloned[i].SkillCooldowns[key] = value
			}
		}
	}
	return cloned
}

func countAlive(items []Combatant) int {
	total := 0
	for _, item := range items {
		if isAlive(item) {
			total++
		}
	}
	return total
}

func isAlive(item Combatant) bool {
	return item.CurrentHP > 0
}

func choosePrimaryEnemyTarget(enemies []Combatant) int {
	indexes := prioritizedEnemyIndexes(enemies)
	if len(indexes) == 0 {
		return -1
	}
	return indexes[0]
}

func enemyRolePriority(role string) int {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "controller":
		return 1
	case "caster":
		return 2
	case "assassin":
		return 3
	case "summoner":
		return 4
	case "bruiser":
		return 5
	case "tank":
		return 6
	case "boss":
		return 7
	default:
		return 8
	}
}

func enemyInitiativeOrder(enemies []Combatant) []int {
	order := make([]int, 0, len(enemies))
	for i, enemy := range enemies {
		if isAlive(enemy) {
			order = append(order, i)
		}
	}
	for i := 0; i < len(order); i++ {
		for j := i + 1; j < len(order); j++ {
			left := enemies[order[i]]
			right := enemies[order[j]]
			if effectiveSpeed(right) > effectiveSpeed(left) {
				order[i], order[j] = order[j], order[i]
			}
		}
	}
	return order
}

func prioritizedEnemyIndexes(enemies []Combatant) []int {
	indexes := make([]int, 0, len(enemies))
	for i, enemy := range enemies {
		if isAlive(enemy) {
			indexes = append(indexes, i)
		}
	}
	for i := 0; i < len(indexes); i++ {
		for j := i + 1; j < len(indexes); j++ {
			left := enemies[indexes[i]]
			right := enemies[indexes[j]]
			leftPriority := enemyRolePriority(left.Role)
			rightPriority := enemyRolePriority(right.Role)
			if rightPriority < leftPriority || (rightPriority == leftPriority && right.CurrentHP < left.CurrentHP) {
				indexes[i], indexes[j] = indexes[j], indexes[i]
			}
		}
	}
	return indexes
}

func applyStartTurnEffectsMulti(actor *Combatant, enemies []Combatant, cfg SkirmishConfig, turn int, log []map[string]any) []map[string]any {
	for _, buf := range actor.ActiveBuffs {
		if buf.Family != "regen" {
			continue
		}
		before := actor.CurrentHP
		heal := imax(4, int(float64(actor.MaxHP)*buf.Value))
		actor.CurrentHP = imin(actor.MaxHP, actor.CurrentHP+heal)
		log = append(log, map[string]any{
			"step":             "action",
			"event_type":       "status_tick",
			"room_index":       cfg.RoomIndex,
			"turn":             turn,
			"actor":            actor.Team,
			"target":           actor.Team,
			"action":           "regen_tick",
			"skill_id":         buf.PotionID,
			"value":            actor.CurrentHP - before,
			"value_type":       "heal",
			"target_hp_before": before,
			"target_hp_after":  actor.CurrentHP,
			"player_hp":        actor.CurrentHP,
			"enemy_count":      countAlive(enemies),
			"message":          fmt.Sprintf("%s recovers from %s", actor.Name, buf.PotionID),
		})
	}

	if actor.SummonTurns <= 0 {
		return log
	}

	if actor.SummonMend > 0 {
		before := actor.CurrentHP
		heal := imax(4, int(float64(actor.HealPow+imax(actor.MagAtk, actor.PhysAtk))*actor.SummonMend))
		actor.CurrentHP = imin(actor.MaxHP, actor.CurrentHP+heal)
		log = append(log, map[string]any{
			"step":             "action",
			"event_type":       "summon_tick",
			"room_index":       cfg.RoomIndex,
			"turn":             turn,
			"actor":            actor.Team,
			"target":           actor.Team,
			"action":           "summon_mend",
			"skill_id":         actor.SummonSkillID,
			"value":            actor.CurrentHP - before,
			"value_type":       "heal",
			"target_hp_before": before,
			"target_hp_after":  actor.CurrentHP,
			"player_hp":        actor.CurrentHP,
			"enemy_count":      countAlive(enemies),
			"message":          fmt.Sprintf("%s is sustained by %s", actor.Name, actor.SummonSkillID),
		})
	}

	if actor.SummonStrike <= 0 {
		return log
	}
	targetIndex := choosePrimaryEnemyTarget(enemies)
	if targetIndex < 0 {
		return log
	}
	target := &enemies[targetIndex]
	damageType := primaryDamageType(*actor)
	profile := resolveTypedAttack(*actor, *target, BattleConfig{BattleType: cfg.BattleType, RunID: cfg.RunID, RoomIndex: cfg.RoomIndex}, turn, strings.ToLower(strings.ReplaceAll(actor.SummonSkillID, " ", "_"))+"_summon", damageType, actor.SummonStrike)
	beforeHP := target.CurrentHP
	beforeShield := target.ShieldHP
	target.ShieldHP, target.CurrentHP = absorbDamage(target.ShieldHP, target.CurrentHP, profile.Damage)
	log = append(log, map[string]any{
		"step":                 "action",
		"event_type":           "summon_tick",
		"room_index":           cfg.RoomIndex,
		"turn":                 turn,
		"actor":                actor.Team,
		"target":               target.Team,
		"target_entity_id":     target.EntityID,
		"target_name":          target.Name,
		"action":               "summon_strike",
		"skill_id":             actor.SummonSkillID,
		"damage_type":          profile.DamageType,
		"value":                profile.Damage,
		"value_type":           "damage",
		"is_critical":          profile.IsCritical,
		"is_blocked":           profile.IsBlocked,
		"is_evaded":            profile.IsEvaded,
		"target_hp_before":     beforeHP,
		"target_hp_after":      target.CurrentHP,
		"target_shield_before": beforeShield,
		"target_shield_after":  target.ShieldHP,
		"player_hp":            actor.CurrentHP,
		"enemy_count":          countAlive(enemies),
		"message":              fmt.Sprintf("%s's summon acts via %s", actor.Name, actor.SummonSkillID),
	})
	return log
}

func executePlayerSkirmishTurn(actor *Combatant, enemies []Combatant, cfg SkirmishConfig, turn int, log []map[string]any) []map[string]any {
	primaryIndex := choosePrimaryEnemyTarget(enemies)
	if primaryIndex < 0 {
		return log
	}
	primary := enemies[primaryIndex]
	cdBefore := actorCooldownSnap(*actor)

	if skill, ok := chooseSkill(*actor, primary); ok {
		if actor.SkillCooldowns == nil {
			actor.SkillCooldowns = map[string]int{}
		}
		switch skill.ActionKind {
		case "heal":
			heal := computeSkillHeal(*actor, skill)
			before := actor.CurrentHP
			actor.CurrentHP = imin(actor.MaxHP, actor.CurrentHP+heal)
			actor.SkillCooldowns[skill.SkillID] = skill.CooldownRounds
			return append(log, map[string]any{
				"step":                  "action",
				"event_type":            "action",
				"room_index":            cfg.RoomIndex,
				"turn":                  turn,
				"actor":                 actor.Team,
				"target":                actor.Team,
				"action":                "skill_heal",
				"skill_id":              skill.SkillID,
				"value":                 actor.CurrentHP - before,
				"value_type":            "heal",
				"target_hp_before":      before,
				"target_hp_after":       actor.CurrentHP,
				"player_hp":             actor.CurrentHP,
				"enemy_count":           countAlive(enemies),
				"cooldown_before_round": cdBefore,
				"cooldown_after_round":  actorCooldownSnap(*actor),
				"message":               fmt.Sprintf("%s uses %s", actor.Name, skill.SkillID),
			})
		case "buff", "debuff", "summon":
			return executeUtilitySkillMulti(actor, enemies, cfg, turn, log, cdBefore, skill)
		default:
			return executeDamageSkillMulti(actor, enemies, cfg, turn, log, cdBefore, skill)
		}
	}

	target := &enemies[primaryIndex]
	profile := resolveAttack(*actor, *target, BattleConfig{BattleType: cfg.BattleType, RunID: cfg.RunID, RoomIndex: cfg.RoomIndex}, turn, "basic_attack", 1.0)
	before := target.CurrentHP
	beforeShield := target.ShieldHP
	target.ShieldHP, target.CurrentHP = absorbDamage(target.ShieldHP, target.CurrentHP, profile.Damage)
	return append(log, map[string]any{
		"step":                   "action",
		"event_type":             "action",
		"room_index":             cfg.RoomIndex,
		"turn":                   turn,
		"actor":                  actor.Team,
		"target":                 target.Team,
		"target_entity_id":       target.EntityID,
		"target_name":            target.Name,
		"action":                 "basic_attack",
		"skill_id":               basicAttackID(*actor),
		"damage_type":            profile.DamageType,
		"value":                  profile.Damage,
		"value_type":             "damage",
		"is_critical":            profile.IsCritical,
		"is_blocked":             profile.IsBlocked,
		"is_evaded":              profile.IsEvaded,
		"effective_block_rate":   profile.EffectiveBlockRate,
		"effective_crit_rate":    profile.EffectiveCritRate,
		"effective_evasion_rate": profile.EffectiveEvasionRate,
		"target_hp_before":       before,
		"target_hp_after":        target.CurrentHP,
		"target_shield_before":   beforeShield,
		"target_shield_after":    target.ShieldHP,
		"player_hp":              actor.CurrentHP,
		"enemy_count":            countAlive(enemies),
		"cooldown_before_round":  cdBefore,
		"cooldown_after_round":   actorCooldownSnap(*actor),
		"message":                fmt.Sprintf("%s attacks %s", actor.Name, target.Name),
	})
}

func executeUtilitySkillMulti(actor *Combatant, enemies []Combatant, cfg SkirmishConfig, turn int, log []map[string]any, cdBefore map[string]int, skill SkillAction) []map[string]any {
	actor.SkillCooldowns[skill.SkillID] = skill.CooldownRounds
	duration := 2
	if skill.Tier == "ultimate" {
		duration = 3
	}
	if skill.BuffFamily != "" && !hasActiveBuff(*actor, skill.BuffFamily) {
		actor.ActiveBuffs = append(actor.ActiveBuffs, StatusBuff{
			Family:    skill.BuffFamily,
			PotionID:  skill.SkillID,
			Value:     skill.BuffValue,
			Remaining: duration,
		})
	}
	if skill.RegenScale > 0 {
		actor.ActiveBuffs = append(actor.ActiveBuffs, StatusBuff{
			Family:    "regen",
			PotionID:  skill.SkillID,
			Value:     skill.RegenScale,
			Remaining: duration,
		})
	}
	if skill.ShieldScale > 0 {
		actor.ShieldHP += imax(6, int(float64(actor.MaxHP)*skill.ShieldScale*skillLevelScalar(skill.Level)))
	}
	targetIndexes := multiTargetIndexes(enemies, choosePrimaryEnemyTarget(enemies), skill.TargetPattern)
	affected := 0
	for _, index := range targetIndexes {
		target := &enemies[index]
		if skill.DebuffFamily != "" && !hasActiveBuff(*target, skill.DebuffFamily) {
			target.ActiveBuffs = append(target.ActiveBuffs, StatusBuff{
				Family:    skill.DebuffFamily,
				PotionID:  skill.SkillID,
				Value:     skill.DebuffValue,
				Remaining: duration,
			})
			affected++
		}
		if skill.SilenceRounds > 0 && !hasActiveBuff(*target, "silence") {
			target.ActiveBuffs = append(target.ActiveBuffs, StatusBuff{
				Family:    "silence",
				PotionID:  skill.SkillID,
				Value:     float64(skill.SilenceRounds),
				Remaining: skill.SilenceRounds,
			})
			affected++
		}
	}
	if skill.ActionKind == "summon" {
		actor.SummonSkillID = skill.SkillID
		actor.SummonTurns = skill.SummonTurns
		actor.SummonStrike = skill.SummonStrike * skillLevelScalar(skill.Level)
		actor.SummonMend = skill.SummonMend * skillLevelScalar(skill.Level)
	}
	return append(log, map[string]any{
		"step":                  "action",
		"event_type":            "action",
		"room_index":            cfg.RoomIndex,
		"turn":                  turn,
		"actor":                 actor.Team,
		"target":                "multi",
		"action":                "skill_utility",
		"skill_id":              skill.SkillID,
		"value":                 skill.BuffValue,
		"value_type":            "utility",
		"buff_family":           skill.BuffFamily,
		"debuff_family":         skill.DebuffFamily,
		"affected_targets":      affected,
		"target_shield_after":   actor.ShieldHP,
		"player_hp":             actor.CurrentHP,
		"enemy_count":           countAlive(enemies),
		"cooldown_before_round": cdBefore,
		"cooldown_after_round":  actorCooldownSnap(*actor),
		"message":               fmt.Sprintf("%s uses %s", actor.Name, skill.SkillID),
	})
}

func executeDamageSkillMulti(actor *Combatant, enemies []Combatant, cfg SkirmishConfig, turn int, log []map[string]any, cdBefore map[string]int, skill SkillAction) []map[string]any {
	actionID := strings.ToLower(strings.ReplaceAll(skill.SkillID, " ", "_"))
	damageType := skill.DamageType
	if damageType == "" {
		damageType = primaryDamageType(*actor)
	}
	targetIndexes := multiTargetIndexes(enemies, choosePrimaryEnemyTarget(enemies), skill.TargetPattern)
	actor.SkillCooldowns[skill.SkillID] = skill.CooldownRounds
	for _, index := range targetIndexes {
		target := &enemies[index]
		if !isAlive(*target) {
			continue
		}
		profile := resolveTypedAttack(*actor, *target, BattleConfig{BattleType: cfg.BattleType, RunID: cfg.RunID, RoomIndex: cfg.RoomIndex}, turn, actionID, damageType, skillMultiplier(skill))
		beforeHP := target.CurrentHP
		beforeShield := target.ShieldHP
		target.ShieldHP, target.CurrentHP = absorbDamage(target.ShieldHP, target.CurrentHP, profile.Damage)
		if skill.DebuffFamily != "" && !hasActiveBuff(*target, skill.DebuffFamily) {
			target.ActiveBuffs = append(target.ActiveBuffs, StatusBuff{
				Family:    skill.DebuffFamily,
				PotionID:  skill.SkillID,
				Value:     skill.DebuffValue,
				Remaining: 2,
			})
		}
		if skill.SilenceRounds > 0 && !hasActiveBuff(*target, "silence") {
			target.ActiveBuffs = append(target.ActiveBuffs, StatusBuff{
				Family:    "silence",
				PotionID:  skill.SkillID,
				Value:     float64(skill.SilenceRounds),
				Remaining: skill.SilenceRounds,
			})
		}
		log = append(log, map[string]any{
			"step":                   "action",
			"event_type":             "action",
			"room_index":             cfg.RoomIndex,
			"turn":                   turn,
			"actor":                  actor.Team,
			"target":                 target.Team,
			"target_entity_id":       target.EntityID,
			"target_name":            target.Name,
			"action":                 "skill_attack",
			"skill_id":               skill.SkillID,
			"damage_type":            profile.DamageType,
			"value":                  profile.Damage,
			"value_type":             "damage",
			"is_critical":            profile.IsCritical,
			"is_blocked":             profile.IsBlocked,
			"is_evaded":              profile.IsEvaded,
			"effective_block_rate":   profile.EffectiveBlockRate,
			"effective_crit_rate":    profile.EffectiveCritRate,
			"effective_evasion_rate": profile.EffectiveEvasionRate,
			"target_hp_before":       beforeHP,
			"target_hp_after":        target.CurrentHP,
			"target_shield_before":   beforeShield,
			"target_shield_after":    target.ShieldHP,
			"player_hp":              actor.CurrentHP,
			"enemy_count":            countAlive(enemies),
			"cooldown_before_round":  cdBefore,
			"cooldown_after_round":   actorCooldownSnap(*actor),
			"message":                fmt.Sprintf("%s uses %s on %s", actor.Name, skill.SkillID, target.Name),
		})
	}
	if skill.HealScale > 0 {
		actor.CurrentHP = imin(actor.MaxHP, actor.CurrentHP+imax(5, int(float64(actor.HealPow+imax(actor.MagAtk, actor.PhysAtk))*skill.HealScale)))
	}
	if skill.BuffFamily != "" && !hasActiveBuff(*actor, skill.BuffFamily) {
		actor.ActiveBuffs = append(actor.ActiveBuffs, StatusBuff{
			Family:    skill.BuffFamily,
			PotionID:  skill.SkillID,
			Value:     skill.BuffValue,
			Remaining: 2,
		})
	}
	return log
}

func multiTargetIndexes(enemies []Combatant, primaryIndex int, pattern string) []int {
	if primaryIndex < 0 {
		return nil
	}
	switch strings.ToLower(strings.TrimSpace(pattern)) {
	case "target_2":
		return limitedTargetIndexes(enemies, 2)
	case "target_3":
		return limitedTargetIndexes(enemies, 3)
	case "all_enemies", "aoe":
		indexes := make([]int, 0, len(enemies))
		for i, enemy := range enemies {
			if isAlive(enemy) {
				indexes = append(indexes, i)
			}
		}
		return indexes
	default:
		return []int{primaryIndex}
	}
}

func limitedTargetIndexes(enemies []Combatant, limit int) []int {
	indexes := prioritizedEnemyIndexes(enemies)
	if len(indexes) > limit {
		indexes = indexes[:limit]
	}
	return indexes
}

func alivePlayers(players []Combatant) int {
	total := 0
	for _, player := range players {
		if isAlive(player) {
			total++
		}
	}
	return total
}

func raidPlayerOrder(players []Combatant) []int {
	order := make([]int, 0, len(players))
	for i, player := range players {
		if isAlive(player) {
			order = append(order, i)
		}
	}
	for i := 0; i < len(order); i++ {
		for j := i + 1; j < len(order); j++ {
			left := players[order[i]]
			right := players[order[j]]
			if effectiveSpeed(right) > effectiveSpeed(left) {
				order[i], order[j] = order[j], order[i]
			}
		}
	}
	return order
}

func chooseRaidPrimaryTarget(players []Combatant) int {
	best := -1
	for i, player := range players {
		if !isAlive(player) {
			continue
		}
		if best < 0 || player.CurrentHP < players[best].CurrentHP || (player.CurrentHP == players[best].CurrentHP && player.MaxHP < players[best].MaxHP) {
			best = i
		}
	}
	return best
}

func raidTargetIndexes(players []Combatant, primaryIndex int, pattern string) []int {
	if primaryIndex < 0 {
		return nil
	}
	switch strings.ToLower(strings.TrimSpace(pattern)) {
	case "target_3":
		return limitedRaidTargetIndexes(players, 3)
	case "all_players", "all_enemies", "aoe":
		indexes := make([]int, 0, len(players))
		for i, player := range players {
			if isAlive(player) {
				indexes = append(indexes, i)
			}
		}
		return indexes
	default:
		return []int{primaryIndex}
	}
}

func limitedRaidTargetIndexes(players []Combatant, limit int) []int {
	indexes := make([]int, 0, len(players))
	for i, player := range players {
		if isAlive(player) {
			indexes = append(indexes, i)
		}
	}
	for i := 0; i < len(indexes); i++ {
		for j := i + 1; j < len(indexes); j++ {
			left := players[indexes[i]]
			right := players[indexes[j]]
			if right.CurrentHP < left.CurrentHP || (right.CurrentHP == left.CurrentHP && effectiveSpeed(right) > effectiveSpeed(left)) {
				indexes[i], indexes[j] = indexes[j], indexes[i]
			}
		}
	}
	if len(indexes) > limit {
		indexes = indexes[:limit]
	}
	return indexes
}

func executeBossRaidTurn(boss *Combatant, players []Combatant, cfg RaidBossConfig, turn int, log []map[string]any) []map[string]any {
	primaryIndex := chooseRaidPrimaryTarget(players)
	if primaryIndex < 0 {
		return log
	}
	primary := players[primaryIndex]
	cdBefore := actorCooldownSnap(*boss)
	skill, ok := chooseSkill(*boss, primary)
	if !ok {
		profile := resolveAttack(*boss, primary, BattleConfig{BattleType: cfg.BattleType, RunID: cfg.RunID, RoomIndex: 1, IsBossRoom: true}, turn, "enemy_attack", 1.0)
		beforeHP := players[primaryIndex].CurrentHP
		beforeShield := players[primaryIndex].ShieldHP
		players[primaryIndex].ShieldHP, players[primaryIndex].CurrentHP = absorbDamage(players[primaryIndex].ShieldHP, players[primaryIndex].CurrentHP, profile.Damage)
		return append(log, map[string]any{
			"step":                  "action",
			"event_type":            "action",
			"turn":                  turn,
			"actor":                 boss.Team,
			"actor_entity_id":       boss.EntityID,
			"target":                players[primaryIndex].Team,
			"target_entity_id":      players[primaryIndex].EntityID,
			"target_name":           players[primaryIndex].Name,
			"action":                "enemy_attack",
			"skill_id":              enemyAttackID(*boss, "enemy_basic_attack"),
			"damage_type":           profile.DamageType,
			"value":                 profile.Damage,
			"value_type":            "damage",
			"target_hp_before":      beforeHP,
			"target_hp_after":       players[primaryIndex].CurrentHP,
			"target_shield_before":  beforeShield,
			"target_shield_after":   players[primaryIndex].ShieldHP,
			"boss_hp":               boss.CurrentHP,
			"alive_count":           alivePlayers(players),
			"cooldown_before_round": cdBefore,
			"cooldown_after_round":  actorCooldownSnap(*boss),
			"message":               fmt.Sprintf("%s attacks %s", boss.Name, players[primaryIndex].Name),
		})
	}

	if boss.SkillCooldowns == nil {
		boss.SkillCooldowns = map[string]int{}
	}
	boss.SkillCooldowns[skill.SkillID] = skill.CooldownRounds

	if skill.ActionKind == "buff" || skill.ActionKind == "debuff" {
		duration := 2
		if skill.Tier == "ultimate" {
			duration = 3
		}
		if skill.BuffFamily != "" && !hasActiveBuff(*boss, skill.BuffFamily) {
			boss.ActiveBuffs = append(boss.ActiveBuffs, StatusBuff{
				Family:    skill.BuffFamily,
				PotionID:  skill.SkillID,
				Value:     skill.BuffValue,
				Remaining: duration,
			})
		}
		if skill.ShieldScale > 0 {
			boss.ShieldHP += imax(6, int(float64(boss.MaxHP)*skill.ShieldScale*skillLevelScalar(skill.Level)))
		}
		affected := 0
		for _, index := range raidTargetIndexes(players, primaryIndex, skill.TargetPattern) {
			if skill.DebuffFamily != "" && !hasActiveBuff(players[index], skill.DebuffFamily) {
				players[index].ActiveBuffs = append(players[index].ActiveBuffs, StatusBuff{
					Family:    skill.DebuffFamily,
					PotionID:  skill.SkillID,
					Value:     skill.DebuffValue,
					Remaining: duration,
				})
				affected++
			}
			if skill.SilenceRounds > 0 && !hasActiveBuff(players[index], "silence") {
				players[index].ActiveBuffs = append(players[index].ActiveBuffs, StatusBuff{
					Family:    "silence",
					PotionID:  skill.SkillID,
					Value:     float64(skill.SilenceRounds),
					Remaining: skill.SilenceRounds,
				})
				affected++
			}
		}
		return append(log, map[string]any{
			"step":                  "action",
			"event_type":            "action",
			"turn":                  turn,
			"actor":                 boss.Team,
			"actor_entity_id":       boss.EntityID,
			"target":                "multi",
			"action":                "boss_utility",
			"skill_id":              skill.SkillID,
			"buff_family":           skill.BuffFamily,
			"debuff_family":         skill.DebuffFamily,
			"affected_targets":      affected,
			"target_shield_after":   boss.ShieldHP,
			"boss_hp":               boss.CurrentHP,
			"alive_count":           alivePlayers(players),
			"cooldown_before_round": cdBefore,
			"cooldown_after_round":  actorCooldownSnap(*boss),
			"message":               fmt.Sprintf("%s uses %s", boss.Name, skill.SkillID),
		})
	}

	targetIndexes := raidTargetIndexes(players, primaryIndex, skill.TargetPattern)
	actionID := strings.ToLower(strings.ReplaceAll(skill.SkillID, " ", "_"))
	damageType := skill.DamageType
	if damageType == "" {
		damageType = primaryDamageType(*boss)
	}
	for _, index := range targetIndexes {
		if !isAlive(players[index]) {
			continue
		}
		profile := resolveTypedAttack(*boss, players[index], BattleConfig{BattleType: cfg.BattleType, RunID: cfg.RunID, RoomIndex: 1, IsBossRoom: true}, turn, actionID, damageType, skillMultiplier(skill))
		beforeHP := players[index].CurrentHP
		beforeShield := players[index].ShieldHP
		players[index].ShieldHP, players[index].CurrentHP = absorbDamage(players[index].ShieldHP, players[index].CurrentHP, profile.Damage)
		if skill.DebuffFamily != "" && !hasActiveBuff(players[index], skill.DebuffFamily) {
			players[index].ActiveBuffs = append(players[index].ActiveBuffs, StatusBuff{
				Family:    skill.DebuffFamily,
				PotionID:  skill.SkillID,
				Value:     skill.DebuffValue,
				Remaining: 2,
			})
		}
		if skill.SilenceRounds > 0 && !hasActiveBuff(players[index], "silence") {
			players[index].ActiveBuffs = append(players[index].ActiveBuffs, StatusBuff{
				Family:    "silence",
				PotionID:  skill.SkillID,
				Value:     float64(skill.SilenceRounds),
				Remaining: skill.SilenceRounds,
			})
		}
		log = append(log, map[string]any{
			"step":                  "action",
			"event_type":            "action",
			"turn":                  turn,
			"actor":                 boss.Team,
			"actor_entity_id":       boss.EntityID,
			"target":                players[index].Team,
			"target_entity_id":      players[index].EntityID,
			"target_name":           players[index].Name,
			"action":                "boss_skill",
			"skill_id":              skill.SkillID,
			"damage_type":           profile.DamageType,
			"value":                 profile.Damage,
			"value_type":            "damage",
			"target_hp_before":      beforeHP,
			"target_hp_after":       players[index].CurrentHP,
			"target_shield_before":  beforeShield,
			"target_shield_after":   players[index].ShieldHP,
			"boss_hp":               boss.CurrentHP,
			"alive_count":           alivePlayers(players),
			"cooldown_before_round": cdBefore,
			"cooldown_after_round":  actorCooldownSnap(*boss),
			"message":               fmt.Sprintf("%s uses %s on %s", boss.Name, skill.SkillID, players[index].Name),
		})
	}
	return log
}

// actOnce handles one combatant's full turn: potion check + action selection + execution.
// Returns potions consumed this call and updated log.
func actOnce(actor, opponent *Combatant, cfg BattleConfig, turn, budget int, log []map[string]any) (int, []map[string]any) {
	log = applyStartTurnEffects(actor, opponent, cfg, turn, log)
	if actor.CurrentHP <= 0 || opponent.CurrentHP <= 0 {
		return 0, log
	}
	consumed := 0
	if actor.IsPlayerSide && budget > 0 {
		c, newLog := tryUsePotion(actor, opponent, cfg, turn, log)
		consumed += c
		log = newLog
	}
	log = executeAction(actor, opponent, cfg, turn, log)
	return consumed, log
}

// tryUsePotion applies one potion from the bag if auto-use policy triggers.
func tryUsePotion(actor, opponent *Combatant, cfg BattleConfig, turn int, log []map[string]any) (int, []map[string]any) {
	pHP, pEn := logHP(actor, opponent)

	// Priority 1: HP potion when HP ≤ 35 % of max.
	hpThreshold := int(float64(actor.MaxHP) * 0.35)
	if actor.CurrentHP <= hpThreshold {
		for i := range actor.PotionBag {
			p := &actor.PotionBag[i]
			if p.Family == "hp" && p.Quantity > 0 {
				heal := int(float64(actor.MaxHP) * p.HealPct)
				if p.HealCap > 0 && heal > p.HealCap {
					heal = p.HealCap
				}
				before := actor.CurrentHP
				actor.CurrentHP = imin(actor.MaxHP, actor.CurrentHP+heal)
				p.Quantity--
				pHP, _ = logHP(actor, opponent)
				log = append(log, map[string]any{
					"step":             "action",
					"event_type":       "potion.consumed",
					"room_index":       cfg.RoomIndex,
					"turn":             turn,
					"actor":            actor.Team,
					"target":           actor.Team,
					"action":           "use_potion",
					"potion_id":        p.PotionID,
					"potion_family":    "hp",
					"potion_tier":      p.Tier,
					"value":            actor.CurrentHP - before,
					"value_type":       "heal",
					"target_hp_before": before,
					"target_hp_after":  actor.CurrentHP,
					"player_hp":        pHP,
					"enemy_hp":         pEn,
					"message":          fmt.Sprintf("%s uses %s", actor.Name, p.PotionID),
				})
				return 1, log
			}
		}
	}

	// Priority 2: ATK potion on boss room, turn 1, if no ATK buff is active.
	if cfg.IsBossRoom && turn == 1 && !hasActiveBuff(*actor, "atk") {
		for i := range actor.PotionBag {
			p := &actor.PotionBag[i]
			if p.Family == "atk" && p.Quantity > 0 {
				actor.ActiveBuffs = append(actor.ActiveBuffs, StatusBuff{
					Family:    "atk",
					PotionID:  p.PotionID,
					Value:     p.BuffValue,
					Remaining: p.Duration,
				})
				p.Quantity--
				log = append(log, map[string]any{
					"step":          "action",
					"event_type":    "potion.consumed",
					"room_index":    cfg.RoomIndex,
					"turn":          turn,
					"actor":         actor.Team,
					"target":        actor.Team,
					"action":        "use_potion",
					"potion_id":     p.PotionID,
					"potion_family": "atk",
					"potion_tier":   p.Tier,
					"value":         p.BuffValue,
					"value_type":    "buff",
					"buff_family":   "atk",
					"buff_duration": p.Duration,
					"player_hp":     pHP,
					"enemy_hp":      pEn,
					"message":       fmt.Sprintf("%s uses %s (ATK buff)", actor.Name, p.PotionID),
				})
				return 1, log
			}
		}
	}

	return 0, log
}

// executeAction selects and executes the actor's combat action for this turn.
func executeAction(actor, opponent *Combatant, cfg BattleConfig, turn int, log []map[string]any) []map[string]any {
	pHP, pEn := logHP(actor, opponent)
	cdBefore := actorCooldownSnap(*actor)

	isPlayerSide := actor.IsPlayerSide || cfg.BattleType == "arena_duel"

	if isPlayerSide {
		if actionLog, ok := executeEquippedSkill(actor, opponent, cfg, turn, log, cdBefore); ok {
			return actionLog
		}

		// Basic attack.
		profile := resolveAttack(*actor, *opponent, cfg, turn, "basic_attack", 1.0)
		before := opponent.CurrentHP
		beforeShield := opponent.ShieldHP
		opponent.ShieldHP, opponent.CurrentHP = absorbDamage(opponent.ShieldHP, opponent.CurrentHP, profile.Damage)
		_, pEn = logHP(actor, opponent)
		return append(log, map[string]any{
			"step":                   "action",
			"event_type":             "action",
			"room_index":             cfg.RoomIndex,
			"turn":                   turn,
			"actor":                  actor.Team,
			"target":                 opponent.Team,
			"action":                 "basic_attack",
			"skill_id":               basicAttackID(*actor),
			"damage_type":            profile.DamageType,
			"value":                  profile.Damage,
			"value_type":             "damage",
			"is_critical":            profile.IsCritical,
			"is_blocked":             profile.IsBlocked,
			"is_evaded":              profile.IsEvaded,
			"effective_block_rate":   profile.EffectiveBlockRate,
			"effective_crit_rate":    profile.EffectiveCritRate,
			"effective_evasion_rate": profile.EffectiveEvasionRate,
			"target_hp_before":       before,
			"target_hp_after":        opponent.CurrentHP,
			"target_shield_before":   beforeShield,
			"target_shield_after":    opponent.ShieldHP,
			"player_hp":              pHP,
			"enemy_hp":               pEn,
			"cooldown_before_round":  cdBefore,
			"cooldown_after_round":   actorCooldownSnap(*actor),
			"message":                fmt.Sprintf("%s attacks", actor.Name),
		})
	}

	// Enemy side: burst → basic.
	if len(actor.EquippedSkills) > 0 {
		if actionLog, ok := executeEquippedSkill(actor, opponent, cfg, turn, log, cdBefore); ok {
			return actionLog
		}
	}

	action := "enemy_attack"
	skillID := "enemy_basic_attack"
	multiplier := 1.0
	if actor.BurstCD == 0 && !hasActiveBuff(*actor, "silence") {
		action = "enemy_skill"
		skillID = "enemy_burst_skill"
		multiplier = 1.35
		actor.BurstCD = 2
	}
	profile := resolveAttack(*actor, *opponent, cfg, turn, action, multiplier)
	before := opponent.CurrentHP
	beforeShield := opponent.ShieldHP
	opponent.ShieldHP, opponent.CurrentHP = absorbDamage(opponent.ShieldHP, opponent.CurrentHP, profile.Damage)
	_, pEn = logHP(actor, opponent)
	return append(log, map[string]any{
		"step":                   "action",
		"event_type":             "action",
		"room_index":             cfg.RoomIndex,
		"turn":                   turn,
		"actor":                  actor.Team,
		"target":                 opponent.Team,
		"action":                 action,
		"skill_id":               enemyAttackID(*actor, skillID),
		"damage_type":            profile.DamageType,
		"value":                  profile.Damage,
		"value_type":             "damage",
		"is_critical":            profile.IsCritical,
		"is_blocked":             profile.IsBlocked,
		"is_evaded":              profile.IsEvaded,
		"effective_block_rate":   profile.EffectiveBlockRate,
		"effective_crit_rate":    profile.EffectiveCritRate,
		"effective_evasion_rate": profile.EffectiveEvasionRate,
		"target_hp_before":       before,
		"target_hp_after":        opponent.CurrentHP,
		"target_shield_before":   beforeShield,
		"target_shield_after":    opponent.ShieldHP,
		"player_hp":              pHP,
		"enemy_hp":               pEn,
		"cooldown_before_round":  cdBefore,
		"cooldown_after_round":   actorCooldownSnap(*actor),
		"message":                fmt.Sprintf("%s attacks", actor.Name),
	})
}

// tickStatusBuffs decrements remaining rounds on all active buffs, removing expired ones.
func tickStatusBuffs(c *Combatant) {
	active := c.ActiveBuffs[:0]
	for _, buf := range c.ActiveBuffs {
		buf.Remaining--
		if buf.Remaining > 0 {
			active = append(active, buf)
		}
	}
	c.ActiveBuffs = active
}

func tickSkillCooldowns(c *Combatant) {
	if len(c.SkillCooldowns) == 0 {
		return
	}
	for skillID, remaining := range c.SkillCooldowns {
		remaining--
		if remaining <= 0 {
			delete(c.SkillCooldowns, skillID)
			continue
		}
		c.SkillCooldowns[skillID] = remaining
	}
}

func tickSummon(c *Combatant) {
	if c.SummonTurns <= 0 {
		c.SummonTurns = 0
		c.SummonSkillID = ""
		c.SummonStrike = 0
		c.SummonMend = 0
		return
	}
	c.SummonTurns--
	if c.SummonTurns <= 0 {
		c.SummonTurns = 0
		c.SummonSkillID = ""
		c.SummonStrike = 0
		c.SummonMend = 0
	}
}

func applyStartTurnEffects(actor, opponent *Combatant, cfg BattleConfig, turn int, log []map[string]any) []map[string]any {
	for _, buf := range actor.ActiveBuffs {
		if buf.Family != "regen" {
			continue
		}
		before := actor.CurrentHP
		heal := imax(4, int(float64(actor.MaxHP)*buf.Value))
		actor.CurrentHP = imin(actor.MaxHP, actor.CurrentHP+heal)
		pHP, pEn := logHP(actor, opponent)
		log = append(log, map[string]any{
			"step":             "action",
			"event_type":       "status_tick",
			"room_index":       cfg.RoomIndex,
			"turn":             turn,
			"actor":            actor.Team,
			"target":           actor.Team,
			"action":           "regen_tick",
			"skill_id":         buf.PotionID,
			"value":            actor.CurrentHP - before,
			"value_type":       "heal",
			"target_hp_before": before,
			"target_hp_after":  actor.CurrentHP,
			"player_hp":        pHP,
			"enemy_hp":         pEn,
			"message":          fmt.Sprintf("%s recovers from %s", actor.Name, buf.PotionID),
		})
	}

	if actor.SummonTurns > 0 {
		if actor.SummonMend > 0 {
			before := actor.CurrentHP
			heal := imax(4, int(float64(actor.HealPow+imax(actor.MagAtk, actor.PhysAtk))*actor.SummonMend))
			actor.CurrentHP = imin(actor.MaxHP, actor.CurrentHP+heal)
			pHP, pEn := logHP(actor, opponent)
			log = append(log, map[string]any{
				"step":             "action",
				"event_type":       "summon_tick",
				"room_index":       cfg.RoomIndex,
				"turn":             turn,
				"actor":            actor.Team,
				"target":           actor.Team,
				"action":           "summon_mend",
				"skill_id":         actor.SummonSkillID,
				"value":            actor.CurrentHP - before,
				"value_type":       "heal",
				"target_hp_before": before,
				"target_hp_after":  actor.CurrentHP,
				"player_hp":        pHP,
				"enemy_hp":         pEn,
				"message":          fmt.Sprintf("%s is sustained by %s", actor.Name, actor.SummonSkillID),
			})
		}
		if actor.SummonStrike > 0 && opponent.CurrentHP > 0 {
			damageType := primaryDamageType(*actor)
			profile := resolveTypedAttack(*actor, *opponent, cfg, turn, strings.ToLower(strings.ReplaceAll(actor.SummonSkillID, " ", "_"))+"_summon", damageType, actor.SummonStrike)
			beforeHP := opponent.CurrentHP
			beforeShield := opponent.ShieldHP
			opponent.ShieldHP, opponent.CurrentHP = absorbDamage(opponent.ShieldHP, opponent.CurrentHP, profile.Damage)
			pHP, pEn := logHP(actor, opponent)
			log = append(log, map[string]any{
				"step":                 "action",
				"event_type":           "summon_tick",
				"room_index":           cfg.RoomIndex,
				"turn":                 turn,
				"actor":                actor.Team,
				"target":               opponent.Team,
				"action":               "summon_strike",
				"skill_id":             actor.SummonSkillID,
				"damage_type":          profile.DamageType,
				"value":                profile.Damage,
				"value_type":           "damage",
				"is_critical":          profile.IsCritical,
				"is_blocked":           profile.IsBlocked,
				"is_evaded":            profile.IsEvaded,
				"target_hp_before":     beforeHP,
				"target_hp_after":      opponent.CurrentHP,
				"target_shield_before": beforeShield,
				"target_shield_after":  opponent.ShieldHP,
				"player_hp":            pHP,
				"enemy_hp":             pEn,
				"message":              fmt.Sprintf("%s's summon acts via %s", actor.Name, actor.SummonSkillID),
			})
		}
	}

	return log
}

func executeEquippedSkill(actor, opponent *Combatant, cfg BattleConfig, turn int, log []map[string]any, cdBefore map[string]int) ([]map[string]any, bool) {
	skill, ok := chooseSkill(*actor, *opponent)
	if !ok {
		return log, false
	}
	if actor.SkillCooldowns == nil {
		actor.SkillCooldowns = map[string]int{}
	}

	switch skill.ActionKind {
	case "heal":
		heal := computeSkillHeal(*actor, skill)
		before := actor.CurrentHP
		actor.CurrentHP = imin(actor.MaxHP, actor.CurrentHP+heal)
		actor.SkillCooldowns[skill.SkillID] = skill.CooldownRounds
		pHP, pEn := logHP(actor, opponent)
		return append(log, map[string]any{
			"step":                  "action",
			"event_type":            "action",
			"room_index":            cfg.RoomIndex,
			"turn":                  turn,
			"actor":                 actor.Team,
			"target":                actor.Team,
			"action":                "skill_heal",
			"skill_id":              skill.SkillID,
			"value":                 actor.CurrentHP - before,
			"value_type":            "heal",
			"target_hp_before":      before,
			"target_hp_after":       actor.CurrentHP,
			"player_hp":             pHP,
			"enemy_hp":              pEn,
			"cooldown_before_round": cdBefore,
			"cooldown_after_round":  actorCooldownSnap(*actor),
			"message":               fmt.Sprintf("%s uses %s", actor.Name, skill.SkillID),
		}), true
	case "buff", "debuff", "summon":
		return executeUtilitySkill(actor, opponent, cfg, turn, log, cdBefore, skill), true
	default:
		return executeDamageSkill(actor, opponent, cfg, turn, log, cdBefore, skill), true
	}
}

func executeUtilitySkill(actor, opponent *Combatant, cfg BattleConfig, turn int, log []map[string]any, cdBefore map[string]int, skill SkillAction) []map[string]any {
	actor.SkillCooldowns[skill.SkillID] = skill.CooldownRounds
	duration := 2
	if skill.Tier == "ultimate" {
		duration = 3
	}
	if skill.BuffFamily != "" && !hasActiveBuff(*actor, skill.BuffFamily) {
		actor.ActiveBuffs = append(actor.ActiveBuffs, StatusBuff{
			Family:    skill.BuffFamily,
			PotionID:  skill.SkillID,
			Value:     skill.BuffValue,
			Remaining: duration,
		})
	}
	if skill.RegenScale > 0 {
		actor.ActiveBuffs = append(actor.ActiveBuffs, StatusBuff{
			Family:    "regen",
			PotionID:  skill.SkillID,
			Value:     skill.RegenScale,
			Remaining: duration,
		})
	}
	if skill.ShieldScale > 0 {
		actor.ShieldHP += imax(6, int(float64(actor.MaxHP)*skill.ShieldScale*skillLevelScalar(skill.Level)))
	}
	if skill.DebuffFamily != "" && !hasActiveBuff(*opponent, skill.DebuffFamily) {
		opponent.ActiveBuffs = append(opponent.ActiveBuffs, StatusBuff{
			Family:    skill.DebuffFamily,
			PotionID:  skill.SkillID,
			Value:     skill.DebuffValue,
			Remaining: duration,
		})
	}
	if skill.SilenceRounds > 0 {
		opponent.ActiveBuffs = append(opponent.ActiveBuffs, StatusBuff{
			Family:    "silence",
			PotionID:  skill.SkillID,
			Value:     float64(skill.SilenceRounds),
			Remaining: skill.SilenceRounds,
		})
	}
	if skill.ActionKind == "summon" {
		actor.SummonSkillID = skill.SkillID
		actor.SummonTurns = skill.SummonTurns
		actor.SummonStrike = skill.SummonStrike * skillLevelScalar(skill.Level)
		actor.SummonMend = skill.SummonMend * skillLevelScalar(skill.Level)
	}
	pHP, pEn := logHP(actor, opponent)
	return append(log, map[string]any{
		"step":                  "action",
		"event_type":            "action",
		"room_index":            cfg.RoomIndex,
		"turn":                  turn,
		"actor":                 actor.Team,
		"target":                actor.Team,
		"action":                "skill_utility",
		"skill_id":              skill.SkillID,
		"value":                 skill.BuffValue,
		"value_type":            "utility",
		"buff_family":           skill.BuffFamily,
		"debuff_family":         skill.DebuffFamily,
		"target_shield_after":   actor.ShieldHP,
		"player_hp":             pHP,
		"enemy_hp":              pEn,
		"cooldown_before_round": cdBefore,
		"cooldown_after_round":  actorCooldownSnap(*actor),
		"message":               fmt.Sprintf("%s uses %s", actor.Name, skill.SkillID),
	})
}

func executeDamageSkill(actor, opponent *Combatant, cfg BattleConfig, turn int, log []map[string]any, cdBefore map[string]int, skill SkillAction) []map[string]any {
	actionID := strings.ToLower(strings.ReplaceAll(skill.SkillID, " ", "_"))
	damageType := skill.DamageType
	if damageType == "" {
		damageType = primaryDamageType(*actor)
	}
	profile := resolveTypedAttack(*actor, *opponent, cfg, turn, actionID, damageType, skillMultiplier(skill))
	beforeHP := opponent.CurrentHP
	beforeShield := opponent.ShieldHP
	opponent.ShieldHP, opponent.CurrentHP = absorbDamage(opponent.ShieldHP, opponent.CurrentHP, profile.Damage)
	actor.SkillCooldowns[skill.SkillID] = skill.CooldownRounds
	if skill.DebuffFamily != "" && !hasActiveBuff(*opponent, skill.DebuffFamily) {
		opponent.ActiveBuffs = append(opponent.ActiveBuffs, StatusBuff{
			Family:    skill.DebuffFamily,
			PotionID:  skill.SkillID,
			Value:     skill.DebuffValue,
			Remaining: 2,
		})
	}
	if skill.SilenceRounds > 0 {
		opponent.ActiveBuffs = append(opponent.ActiveBuffs, StatusBuff{
			Family:    "silence",
			PotionID:  skill.SkillID,
			Value:     float64(skill.SilenceRounds),
			Remaining: skill.SilenceRounds,
		})
	}
	if skill.HealScale > 0 {
		actor.CurrentHP = imin(actor.MaxHP, actor.CurrentHP+imax(5, int(float64(actor.HealPow+imax(actor.MagAtk, actor.PhysAtk))*skill.HealScale)))
	}
	pHP, pEn := logHP(actor, opponent)
	return append(log, map[string]any{
		"step":                   "action",
		"event_type":             "action",
		"room_index":             cfg.RoomIndex,
		"turn":                   turn,
		"actor":                  actor.Team,
		"target":                 opponent.Team,
		"action":                 "skill_attack",
		"skill_id":               skill.SkillID,
		"damage_type":            profile.DamageType,
		"value":                  profile.Damage,
		"value_type":             "damage",
		"is_critical":            profile.IsCritical,
		"is_blocked":             profile.IsBlocked,
		"is_evaded":              profile.IsEvaded,
		"effective_block_rate":   profile.EffectiveBlockRate,
		"effective_crit_rate":    profile.EffectiveCritRate,
		"effective_evasion_rate": profile.EffectiveEvasionRate,
		"target_hp_before":       beforeHP,
		"target_hp_after":        opponent.CurrentHP,
		"target_shield_before":   beforeShield,
		"target_shield_after":    opponent.ShieldHP,
		"player_hp":              pHP,
		"enemy_hp":               pEn,
		"cooldown_before_round":  cdBefore,
		"cooldown_after_round":   actorCooldownSnap(*actor),
		"message":                fmt.Sprintf("%s uses %s", actor.Name, skill.SkillID),
	})
}

func chooseSkill(actor, opponent Combatant) (SkillAction, bool) {
	if hasActiveBuff(actor, "silence") {
		return SkillAction{}, false
	}
	available := make([]SkillAction, 0, len(actor.EquippedSkills))
	for _, skill := range actor.EquippedSkills {
		if actor.SkillCooldowns != nil && actor.SkillCooldowns[skill.SkillID] > 0 {
			continue
		}
		if skill.HalfHPOnly && actor.CurrentHP > actor.MaxHP/2 {
			continue
		}
		available = append(available, skill)
	}
	if len(available) == 0 {
		return SkillAction{}, false
	}

	healThreshold := int(float64(actor.MaxHP) * 0.55)
	if actor.CurrentHP <= healThreshold {
		for _, skill := range available {
			if skill.ActionKind == "heal" {
				return skill, true
			}
		}
	}

	defenseThreshold := int(float64(actor.MaxHP) * 0.45)
	if actor.CurrentHP <= defenseThreshold {
		for _, skill := range available {
			if skill.ActionKind == "buff" && (skill.BuffFamily == "def" || skill.BuffFamily == "spd") && !hasActiveBuff(actor, skill.BuffFamily) {
				return skill, true
			}
		}
	}

	if actor.SummonTurns <= 0 {
		for _, skill := range available {
			if skill.ActionKind == "summon" {
				return skill, true
			}
		}
	}

	for _, skill := range available {
		if skill.ActionKind == "debuff" {
			if skill.DebuffFamily != "" && !hasActiveBuff(opponent, skill.DebuffFamily) {
				return skill, true
			}
			if skill.SilenceRounds > 0 && !hasActiveBuff(opponent, "silence") {
				return skill, true
			}
		}
	}

	for _, skill := range available {
		if skill.ActionKind == "damage" {
			return skill, true
		}
	}

	for _, skill := range available {
		if skill.ActionKind == "buff" && skill.BuffFamily != "" && !hasActiveBuff(actor, skill.BuffFamily) {
			return skill, true
		}
	}

	return SkillAction{}, false
}

func computeSkillHeal(actor Combatant, skill SkillAction) int {
	primaryAttack := actor.PhysAtk
	if actor.MagAtk > primaryAttack {
		primaryAttack = actor.MagAtk
	}
	base := float64(actor.HealPow)*1.4 + float64(primaryAttack)*0.7
	return imax(10, int(base*skill.HealScale*skillLevelScalar(skill.Level)))
}

func skillLevelScalar(level int) float64 {
	if level <= 0 {
		return 1
	}
	return 1 + 0.02*float64(level)
}

func skillMultiplier(skill SkillAction) float64 {
	return skill.PowerScale * skillLevelScalar(skill.Level)
}

func hasActiveBuff(c Combatant, family string) bool {
	for _, buf := range c.ActiveBuffs {
		if buf.Family == family {
			return true
		}
	}
	return false
}

// effectiveSpeed returns speed with SPD buffs applied.
func effectiveSpeed(c Combatant) int {
	spd := c.Speed
	for _, buf := range c.ActiveBuffs {
		if buf.Family == "spd" {
			spd = int(float64(spd) * (1 + buf.Value))
		}
		if buf.Family == "slow" {
			spd = int(float64(spd) * (1 - buf.Value))
		}
	}
	return spd
}

type attackProfile struct {
	Damage               int
	DamageType           string
	IsCritical           bool
	IsBlocked            bool
	IsEvaded             bool
	EffectiveBlockRate   float64
	EffectiveCritRate    float64
	EffectiveEvasionRate float64
}

func resolveAttack(actor, opponent Combatant, cfg BattleConfig, turn int, action string, multiplier float64) attackProfile {
	damageType := primaryDamageType(actor)
	return resolveTypedAttack(actor, opponent, cfg, turn, action, damageType, multiplier)
}

func resolveTypedAttack(actor, opponent Combatant, cfg BattleConfig, turn int, action, damageType string, multiplier float64) attackProfile {
	attack := effectiveAttack(actor, damageType)
	defense := effectiveDefense(opponent, damageType)

	evasionRate := clampRate(opponent.EvasionRate)
	if deterministicRoll(cfg, turn, actor, opponent, action, "evade") < evasionRate {
		return attackProfile{
			Damage:               0,
			DamageType:           damageType,
			IsEvaded:             true,
			EffectiveCritRate:    clampRate(actor.CritRate),
			EffectiveBlockRate:   clampRate(opponent.BlockRate - actor.Precision),
			EffectiveEvasionRate: evasionRate,
		}
	}

	damage := computeDamage(attack, defense, multiplier)
	damage = applyMasteryDamage(actor, damageType, damage)

	critRate := clampRate(actor.CritRate)
	isCritical := deterministicRoll(cfg, turn, actor, opponent, action, "crit") < critRate
	if isCritical {
		damage = imax(1, int(float64(damage)*(1+actor.CritDamage)))
	}

	effectiveBlockRate := clampRate(opponent.BlockRate - actor.Precision)
	isBlocked := deterministicRoll(cfg, turn, actor, opponent, action, "block") < effectiveBlockRate
	if isBlocked {
		damage = imax(1, int(float64(damage)*0.5))
	}
	if hasActiveBuff(opponent, "vulnerable") {
		damage = imax(1, int(float64(damage)*1.12))
	}

	return attackProfile{
		Damage:               damage,
		DamageType:           damageType,
		IsCritical:           isCritical,
		IsBlocked:            isBlocked,
		IsEvaded:             false,
		EffectiveBlockRate:   effectiveBlockRate,
		EffectiveCritRate:    critRate,
		EffectiveEvasionRate: evasionRate,
	}
}

func primaryDamageType(c Combatant) string {
	if c.MagAtk > c.PhysAtk {
		return "magic"
	}
	return "physical"
}

func effectiveAttack(c Combatant, damageType string) int {
	base := c.PhysAtk
	if damageType == "magic" {
		base = c.MagAtk
	}
	for _, buf := range c.ActiveBuffs {
		if buf.Family == "atk" {
			base = int(float64(base) * (1 + buf.Value))
		}
		if buf.Family == "weaken" {
			base = int(float64(base) * (1 - buf.Value))
		}
	}
	return base
}

func effectiveDefense(c Combatant, damageType string) int {
	def := c.PhysDef
	if damageType == "magic" {
		def = c.MagDef
	}
	for _, buf := range c.ActiveBuffs {
		if buf.Family == "def" {
			def = int(float64(def) * (1 + buf.Value))
		}
		if buf.Family == "shred" {
			def = int(float64(def) * (1 - buf.Value))
		}
	}
	return def
}

func applyMasteryDamage(c Combatant, damageType string, damage int) int {
	mastery := c.PhysicalMastery
	if damageType == "magic" {
		mastery = c.MagicMastery
	}
	if mastery <= 0 {
		return damage
	}
	return imax(1, int(float64(damage)*(1+mastery)))
}

func computeDamage(attack, defense int, multiplier float64) int {
	raw := int(float64(attack)*multiplier) - int(float64(defense)*0.35)
	return imax(1, raw)
}

func absorbDamage(shield, hp, damage int) (int, int) {
	if damage <= 0 {
		return shield, hp
	}
	if shield > 0 {
		if shield >= damage {
			return shield - damage, hp
		}
		damage -= shield
		shield = 0
	}
	return shield, imax(0, hp-damage)
}

func deterministicRoll(cfg BattleConfig, turn int, actor, opponent Combatant, action, suffix string) float64 {
	hasher := fnv.New64a()
	_, _ = hasher.Write([]byte(fmt.Sprintf("%s|%d|%d|%s|%s|%s|%s", cfg.RunID, cfg.RoomIndex, turn, actor.EntityID, opponent.EntityID, action, suffix)))
	return float64(hasher.Sum64()%10000) / 10000
}

func clampRate(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 0.95 {
		return 0.95
	}
	return value
}

// logHP returns (player_hp, enemy_hp) for log events, where "player" always means SideA (team "a").
func logHP(actor, opponent *Combatant) (int, int) {
	if actor.Team == "a" {
		return actor.CurrentHP, opponent.CurrentHP
	}
	return opponent.CurrentHP, actor.CurrentHP
}

// roundCooldownMap returns the combined cooldown snapshot used in turn_start / turn_end log events.
func roundCooldownMap(a, b Combatant) map[string]int {
	if len(a.SkillCooldowns) > 0 || len(b.SkillCooldowns) > 0 {
		items := map[string]int{}
		for skillID, remaining := range a.SkillCooldowns {
			items["player:"+skillID] = remaining
		}
		for skillID, remaining := range b.SkillCooldowns {
			items["enemy:"+skillID] = remaining
		}
		return items
	}
	return map[string]int{
		"player_burst": a.BurstCD,
		"player_heal":  a.HealCD,
		"enemy_burst":  b.BurstCD,
	}
}

// actorCooldownSnap returns the actor-scoped cooldown map used in per-action log events.
func actorCooldownSnap(c Combatant) map[string]int {
	if len(c.SkillCooldowns) > 0 {
		items := make(map[string]int, len(c.SkillCooldowns))
		for skillID, remaining := range c.SkillCooldowns {
			items[skillID] = remaining
		}
		return items
	}
	if c.Team == "a" {
		return map[string]int{"player_burst": c.BurstCD, "player_heal": c.HealCD}
	}
	return map[string]int{"enemy_burst": c.BurstCD}
}

func basicAttackID(c Combatant) string {
	if strings.TrimSpace(c.BasicSkillID) != "" {
		return c.BasicSkillID
	}
	switch c.Class {
	case "mage":
		return "Arc Bolt"
	case "priest":
		return "Smite"
	default:
		return "Strike"
	}
}

func enemyAttackID(c Combatant, fallback string) string {
	if strings.TrimSpace(fallback) != "" && fallback != "enemy_basic_attack" && fallback != "enemy_burst_skill" {
		return fallback
	}
	if strings.TrimSpace(c.BasicSkillID) != "" {
		return c.BasicSkillID
	}
	return fallback
}

func imax(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func imin(a, b int) int {
	if a < b {
		return a
	}
	return b
}
