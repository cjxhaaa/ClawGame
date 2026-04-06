package combat

// PotionCatalog holds the full definition of every potion by potion_id.
var PotionCatalog = map[string]PotionItem{
	// HP Potions
	"potion_hp_t1": {PotionID: "potion_hp_t1", Family: "hp", Tier: 1, HealPct: 0.25, HealCap: 220},
	"potion_hp_t2": {PotionID: "potion_hp_t2", Family: "hp", Tier: 2, HealPct: 0.35, HealCap: 520},
	"potion_hp_t3": {PotionID: "potion_hp_t3", Family: "hp", Tier: 3, HealPct: 0.45, HealCap: 980},

	// ATK Potions
	"potion_atk_t1": {PotionID: "potion_atk_t1", Family: "atk", Tier: 1, BuffValue: 0.10, Duration: 3},
	"potion_atk_t2": {PotionID: "potion_atk_t2", Family: "atk", Tier: 2, BuffValue: 0.16, Duration: 3},
	"potion_atk_t3": {PotionID: "potion_atk_t3", Family: "atk", Tier: 3, BuffValue: 0.24, Duration: 4},

	// DEF Potions
	"potion_def_t1": {PotionID: "potion_def_t1", Family: "def", Tier: 1, BuffValue: 0.10, Duration: 3},
	"potion_def_t2": {PotionID: "potion_def_t2", Family: "def", Tier: 2, BuffValue: 0.16, Duration: 3},
	"potion_def_t3": {PotionID: "potion_def_t3", Family: "def", Tier: 3, BuffValue: 0.24, Duration: 4},

	// SPD Potions
	"potion_spd_t1": {PotionID: "potion_spd_t1", Family: "spd", Tier: 1, BuffValue: 0.08, Duration: 3},
	"potion_spd_t2": {PotionID: "potion_spd_t2", Family: "spd", Tier: 2, BuffValue: 0.12, Duration: 3},
	"potion_spd_t3": {PotionID: "potion_spd_t3", Family: "spd", Tier: 3, BuffValue: 0.18, Duration: 4},
}

// DefaultPotionBag returns the baseline starter bag for a bot character until
// real shop purchases and loadout choices take over.
func DefaultPotionBag() []PotionItem {
	return []PotionItem{
		withQty(PotionCatalog["potion_hp_t2"], 2),
		withQty(PotionCatalog["potion_atk_t2"], 1),
	}
}

func withQty(p PotionItem, qty int) PotionItem {
	p.Quantity = qty
	return p
}
