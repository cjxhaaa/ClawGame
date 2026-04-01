package combat

import (
	"fmt"
	"hash/fnv"
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
	CurrentHP   int
	BurstCD     int
	HealCD      int
	ActiveBuffs []StatusBuff
	PotionBag   []PotionItem
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
		a.BurstCD = imax(0, a.BurstCD-1)
		a.HealCD = imax(0, a.HealCD-1)
		b.BurstCD = imax(0, b.BurstCD-1)
		b.HealCD = imax(0, b.HealCD-1)

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

// actOnce handles one combatant's full turn: potion check + action selection + execution.
// Returns potions consumed this call and updated log.
func actOnce(actor, opponent *Combatant, cfg BattleConfig, turn, budget int, log []map[string]any) (int, []map[string]any) {
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
		// Heal: HP ≤ 55 % and skill off cooldown.
		healThreshold := int(float64(actor.MaxHP) * 0.55)
		if actor.HealPow > 0 && actor.CurrentHP <= healThreshold && actor.HealCD == 0 {
			heal := imax(8, int(float64(actor.HealPow)*1.6)+int(float64(actor.MaxHP)*0.06))
			before := actor.CurrentHP
			actor.CurrentHP = imin(actor.MaxHP, actor.CurrentHP+heal)
			actor.HealCD = 3
			pHP, _ = logHP(actor, opponent)
			return append(log, map[string]any{
				"step":                  "action",
				"event_type":            "action",
				"room_index":            cfg.RoomIndex,
				"turn":                  turn,
				"actor":                 actor.Team,
				"target":                actor.Team,
				"action":                "recovery_wave",
				"skill_id":              "player_recovery_wave",
				"value":                 actor.CurrentHP - before,
				"value_type":            "heal",
				"target_hp_before":      before,
				"target_hp_after":       actor.CurrentHP,
				"player_hp":             pHP,
				"enemy_hp":              pEn,
				"cooldown_before_round": cdBefore,
				"cooldown_after_round":  actorCooldownSnap(*actor),
				"message":               fmt.Sprintf("%s casts healing skill", actor.Name),
			})
		}

		// Burst: off cooldown.
		if actor.BurstCD == 0 {
			profile := resolveAttack(*actor, *opponent, cfg, turn, "burst_skill", 1.75)
			before := opponent.CurrentHP
			opponent.CurrentHP = imax(0, opponent.CurrentHP-profile.Damage)
			actor.BurstCD = 2
			_, pEn = logHP(actor, opponent)
			return append(log, map[string]any{
				"step":                   "action",
				"event_type":             "action",
				"room_index":             cfg.RoomIndex,
				"turn":                   turn,
				"actor":                  actor.Team,
				"target":                 opponent.Team,
				"action":                 "burst_skill",
				"skill_id":               "player_burst_skill",
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
				"player_hp":              pHP,
				"enemy_hp":               pEn,
				"cooldown_before_round":  cdBefore,
				"cooldown_after_round":   actorCooldownSnap(*actor),
				"message":                fmt.Sprintf("%s uses burst skill", actor.Name),
			})
		}

		// Basic attack.
		profile := resolveAttack(*actor, *opponent, cfg, turn, "basic_attack", 1.0)
		before := opponent.CurrentHP
		opponent.CurrentHP = imax(0, opponent.CurrentHP-profile.Damage)
		_, pEn = logHP(actor, opponent)
		return append(log, map[string]any{
			"step":                   "action",
			"event_type":             "action",
			"room_index":             cfg.RoomIndex,
			"turn":                   turn,
			"actor":                  actor.Team,
			"target":                 opponent.Team,
			"action":                 "basic_attack",
			"skill_id":               "player_basic_attack",
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
			"player_hp":              pHP,
			"enemy_hp":               pEn,
			"cooldown_before_round":  cdBefore,
			"cooldown_after_round":   actorCooldownSnap(*actor),
			"message":                fmt.Sprintf("%s attacks", actor.Name),
		})
	}

	// Enemy side: burst → basic.
	action := "enemy_attack"
	skillID := "enemy_basic_attack"
	multiplier := 1.0
	if actor.BurstCD == 0 {
		action = "enemy_skill"
		skillID = "enemy_burst_skill"
		multiplier = 1.35
		actor.BurstCD = 2
	}
	profile := resolveAttack(*actor, *opponent, cfg, turn, action, multiplier)
	before := opponent.CurrentHP
	opponent.CurrentHP = imax(0, opponent.CurrentHP-profile.Damage)
	_, pEn = logHP(actor, opponent)
	return append(log, map[string]any{
		"step":                   "action",
		"event_type":             "action",
		"room_index":             cfg.RoomIndex,
		"turn":                   turn,
		"actor":                  actor.Team,
		"target":                 opponent.Team,
		"action":                 action,
		"skill_id":               skillID,
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
	return map[string]int{
		"player_burst": a.BurstCD,
		"player_heal":  a.HealCD,
		"enemy_burst":  b.BurstCD,
	}
}

// actorCooldownSnap returns the actor-scoped cooldown map used in per-action log events.
func actorCooldownSnap(c Combatant) map[string]int {
	if c.Team == "a" {
		return map[string]int{"player_burst": c.BurstCD, "player_heal": c.HealCD}
	}
	return map[string]int{"enemy_burst": c.BurstCD}
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
