package app

import (
	"math"
	"sort"
	"strings"

	"clawgame/apps/api/internal/characters"
	"clawgame/apps/api/internal/dungeons"
	"clawgame/apps/api/internal/inventory"
)

const powerFormulaVersion = "power_score_v1_0"

type powerItemScore struct {
	ItemID           string `json:"item_id"`
	CatalogID        string `json:"catalog_id"`
	Name             string `json:"name"`
	Slot             string `json:"slot"`
	Rarity           string `json:"rarity"`
	EnhancementLevel int    `json:"enhancement_level"`
	IsEquipped       bool   `json:"is_equipped"`
	PowerScore       int    `json:"power_score"`
	DeltaVsEquipped  int    `json:"delta_vs_equipped"`
}

func buildCombatPower(summary characters.Summary, stats characters.StatsSnapshot, inv inventory.InventoryView, dungeonService *dungeons.Service) (characters.CombatPowerSummary, []powerItemScore) {
	effectiveLevel := estimateLevel(summary)
	progressionCoeff := progressionCoeff()
	baseGrowth := computeBaseGrowthScore(summary, stats, effectiveLevel, progressionCoeff)
	itemScores := scoreInventoryItems(summary, inv)
	equipmentScore := 0
	equippedCount := 0
	for _, item := range itemScores {
		if item.IsEquipped {
			equipmentScore += item.PowerScore
			equippedCount++
		}
	}
	buildModifier := computeBuildModifier(stats, summary, equippedCount)
	panel := baseGrowth + equipmentScore + buildModifier
	if panel < 1 {
		panel = 1
	}
	view := characters.CombatPowerSummary{
		FormulaVersion:     powerFormulaVersion,
		EffectiveLevel:     effectiveLevel,
		ProgressionCoeff:   progressionCoeff,
		BaseGrowthScore:    baseGrowth,
		EquipmentScore:     equipmentScore,
		BuildModifierScore: buildModifier,
		PanelPowerScore:    panel,
		PowerTier:          powerTier(panel),
		ArenaPreview:       buildArenaPreview(panel),
		DungeonPreviews:    buildDungeonPreviews(panel, dungeonService),
	}
	return view, itemScores
}

func estimateLevel(summary characters.Summary) int {
	if summary.SeasonLevel > 0 {
		if summary.SeasonLevel > 100 {
			return 100
		}
		return summary.SeasonLevel
	}
	level := 1 + summary.Reputation/80
	if level < 40 {
		level = 40
	}
	if level > 69 {
		level = 69
	}
	if level < 1 {
		level = 1
	}
	if level > 100 {
		level = 100
	}
	return level
}

func progressionCoeff() float64 {
	return 1.35
}

func computeBaseGrowthScore(summary characters.Summary, stats characters.StatsSnapshot, level int, coeff float64) int {
	primary := stats.PhysicalAttack
	secondary := stats.MagicAttack
	if strings.EqualFold(summary.Class, "mage") || strings.EqualFold(summary.Class, "priest") {
		primary = stats.MagicAttack
		secondary = stats.PhysicalAttack
	}
	levelScore := 20.0 * float64(level)
	smoothScore := 8.0 * math.Sqrt(float64(level))
	statScore :=
		1.2*float64(stats.MaxHP) +
			6.0*float64(primary) +
			4.5*float64(secondary) +
			5.0*float64(stats.PhysicalDefense) +
			5.0*float64(stats.MagicDefense) +
			4.0*float64(stats.Speed) +
			4.0*float64(stats.HealingPower)
	return int(coeff * (levelScore + smoothScore + statScore))
}

func scoreInventoryItems(summary characters.Summary, inv inventory.InventoryView) []powerItemScore {
	all := make([]inventory.EquipmentItem, 0, len(inv.Equipped)+len(inv.Inventory))
	all = append(all, inv.Equipped...)
	all = append(all, inv.Inventory...)

	equippedBySlot := map[string]int{}
	raw := make([]powerItemScore, 0, len(all))
	for _, it := range all {
		score := scoreOneItem(summary, it)
		entry := powerItemScore{
			ItemID:           it.ItemID,
			CatalogID:        it.CatalogID,
			Name:             it.Name,
			Slot:             it.Slot,
			Rarity:           it.Rarity,
			EnhancementLevel: it.EnhancementLevel,
			IsEquipped:       strings.EqualFold(it.State, "equipped"),
			PowerScore:       score,
		}
		raw = append(raw, entry)
		if entry.IsEquipped {
			equippedBySlot[it.Slot] = score
		}
	}

	for i := range raw {
		equipped := equippedBySlot[raw[i].Slot]
		raw[i].DeltaVsEquipped = raw[i].PowerScore - equipped
	}

	sort.Slice(raw, func(i, j int) bool {
		if raw[i].IsEquipped != raw[j].IsEquipped {
			return raw[i].IsEquipped
		}
		if raw[i].Slot == raw[j].Slot {
			return raw[i].PowerScore > raw[j].PowerScore
		}
		return raw[i].Slot < raw[j].Slot
	})
	return raw
}

func scoreOneItem(summary characters.Summary, item inventory.EquipmentItem) int {
	score := rarityBase(item.Rarity)
	for key, value := range item.Stats {
		weight := statWeight(key)
		eff := subStatEfficiency(summary.Class, key)
		score += int(weight * float64(value) * eff)
	}
	lvl := float64(item.EnhancementLevel)
	score += int(12.0*lvl + 1.8*lvl*lvl)
	if score < 0 {
		return 0
	}
	return score
}

func rarityBase(rarity string) int {
	switch strings.ToLower(strings.TrimSpace(rarity)) {
	case "mythic", "prismatic":
		return 320
	case "legendary", "red":
		return 220
	case "epic", "purple":
		return 145
	case "rare", "blue":
		return 90
	case "uncommon":
		return 55
	default:
		return 30
	}
}

func statWeight(key string) float64 {
	switch strings.ToLower(strings.TrimSpace(key)) {
	case "max_hp":
		return 1.2
	case "physical_attack", "magic_attack":
		return 6.0
	case "physical_defense", "magic_defense":
		return 5.0
	case "speed", "healing_power":
		return 4.0
	default:
		return 2.0
	}
}

// subStatEfficiency returns the efficiency multiplier for an item stat on a given class.
// Per doc 17 sub stat efficiency weights: effective 1.0x, semi-effective 0.55x, low-effective 0.25x.
// semi-effective applies to cross-class primary attacks.
// low-effective applies to stats that are essentially irrelevant for the class.
func subStatEfficiency(className, statKey string) float64 {
	className = strings.ToLower(strings.TrimSpace(className))
	statKey = strings.ToLower(strings.TrimSpace(statKey))
	// warrior: magic_attack is semi-effective, healing_power is low-effective
	if className == "warrior" {
		if statKey == "magic_attack" {
			return 0.55
		}
		if statKey == "healing_power" {
			return 0.25
		}
	}
	// mage/priest: physical_attack is semi-effective
	// priest's physical_attack is even less useful than mage's — make it low-effective
	if className == "mage" && statKey == "physical_attack" {
		return 0.55
	}
	if className == "priest" && statKey == "physical_attack" {
		return 0.25
	}
	return 1.0
}

func computeBuildModifier(stats characters.StatsSnapshot, summary characters.Summary, equippedCount int) int {
	modifier := 0
	// Skill loadout quality: fully equipped = up to +120 per doc 17.
	modifier += equippedCount * 17 // 7 slots × 17 ≈ 119, just under the +120 ceiling
	if strings.EqualFold(summary.Class, "priest") && stats.HealingPower >= 20 {
		modifier += 35
	}
	// Potion readiness uses the standard baseline bag in the current runtime.
	modifier += 80
	// Critical weakness penalties use a single open-era baseline threshold.
	speedThreshold := 15
	if stats.Speed < speedThreshold {
		modifier -= (speedThreshold - stats.Speed) * 12
	}
	defSum := stats.PhysicalDefense + stats.MagicDefense
	defThreshold := 34
	if defSum < defThreshold {
		modifier -= (defThreshold - defSum) * 8
	}
	if modifier > 200 {
		return 200
	}
	if modifier < -200 {
		return -200
	}
	return modifier
}

func powerTier(panel int) string {
	switch {
	case panel >= 11000:
		return "elite"
	case panel >= 7500:
		return "high"
	case panel >= 4200:
		return "mid"
	default:
		return "low"
	}
}

func buildArenaPreview(panel int) characters.ArenaPowerPreview {
	reference := arenaReferencePower()
	delta := panel - reference
	band := "close"
	tier := "balanced"
	switch {
	case delta <= -2200:
		band = "<30%"
		tier = "hard_disadvantage"
	case delta <= -900:
		band = "30%-45%"
		tier = "disadvantage"
	case delta < 900:
		band = "45%-55%"
		tier = "close"
	case delta < 2200:
		band = "55%-70%"
		tier = "advantage"
	default:
		band = ">70%"
		tier = "strong_advantage"
	}
	return characters.ArenaPowerPreview{
		ReferencePower:        reference,
		PowerDelta:            delta,
		EstimatedWinRateBand:  band,
		EstimatedStrengthTier: tier,
	}
}

func arenaReferencePower() int {
	return 6200
}

func buildDungeonPreviews(panel int, dungeonService *dungeons.Service) []characters.DungeonPowerPreview {
	definitions := dungeonService.ListDungeonDefinitions()
	items := make([]characters.DungeonPowerPreview, 0, len(definitions))
	for _, def := range definitions {
		minPower := recommendedPowerForLevel(def.RecommendedLevelMin)
		maxPower := recommendedPowerForLevel(def.RecommendedLevelMax)
		if maxPower < minPower {
			maxPower = minPower
		}
		conf, band := dungeonConfidence(panel, minPower, maxPower)
		items = append(items, characters.DungeonPowerPreview{
			DungeonID:           def.DungeonID,
			DungeonName:         def.Name,
			RecommendedPowerMin: minPower,
			RecommendedPowerMax: maxPower,
			CurrentPower:        panel,
			EstimatedConfidence: conf,
			EstimatedClearBand:  band,
		})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].RecommendedPowerMin < items[j].RecommendedPowerMin
	})
	return items
}

func recommendedPowerForLevel(level int) int {
	if level < 1 {
		level = 1
	}
	primary := 12 + level*3
	secondary := 8 + level*2
	pdef := 10 + level*2
	mdef := 10 + level*2
	speed := 10 + level
	heal := 6 + level
	hp := 100 + level*8
	levelScore := 20.0*float64(level) + 8.0*math.Sqrt(float64(level))
	statScore :=
		1.2*float64(hp) +
			6.0*float64(primary) +
			4.5*float64(secondary) +
			5.0*float64(pdef) +
			5.0*float64(mdef) +
			4.0*float64(speed) +
			4.0*float64(heal)
	base := int(levelScore + statScore)
	equipment := level * 35
	build := 80
	return base + equipment + build
}

func dungeonConfidence(panel, minPower, maxPower int) (string, string) {
	switch {
	case panel < minPower:
		return "low", "<35%"
	case panel < (minPower+maxPower)/2:
		return "medium", "35%-65%"
	case panel < maxPower:
		return "high", "65%-85%"
	default:
		return "very_high", ">85%"
	}
}
