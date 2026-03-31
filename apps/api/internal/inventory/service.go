package inventory

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"clawgame/apps/api/internal/characters"
	"clawgame/apps/api/internal/combat"
)

var (
	ErrItemNotOwned      = errors.New("item not owned")
	ErrItemNotEquippable = errors.New("item not equippable")
	ErrSlotNotOccupied   = errors.New("slot not occupied")
	ErrCatalogNotFound   = errors.New("catalog item not found")
	ErrItemEquipped      = errors.New("item is currently equipped")
	ErrConsumableMissing = errors.New("consumable missing")
)

type EquipmentItem struct {
	ItemID              string         `json:"item_id"`
	CatalogID           string         `json:"catalog_id"`
	Name                string         `json:"name"`
	Slot                string         `json:"slot"`
	Rarity              string         `json:"rarity"`
	RequiredClass       string         `json:"required_class,omitempty"`
	RequiredWeaponStyle string         `json:"required_weapon_style,omitempty"`
	EnhancementLevel    int            `json:"enhancement_level"`
	Durability          int            `json:"durability"`
	Stats               map[string]int `json:"stats"`
	PassiveAffix        map[string]any `json:"passive_affix,omitempty"`
	State               string         `json:"state"`
}

type InventoryView struct {
	EquipmentScore int               `json:"equipment_score"`
	Equipped       []EquipmentItem   `json:"equipped"`
	Inventory      []EquipmentItem   `json:"inventory"`
	Consumables    []ConsumableStack `json:"consumables"`
}

type ShopItem struct {
	CatalogID      string         `json:"catalog_id"`
	Name           string         `json:"name"`
	ItemType       string         `json:"item_type"`
	Slot           string         `json:"slot"`
	Rarity         string         `json:"rarity"`
	PriceGold      int            `json:"price_gold"`
	RequiredClass  string         `json:"required_class,omitempty"`
	RequiredWeapon string         `json:"required_weapon_style,omitempty"`
	Stats          map[string]int `json:"stats"`
	Family         string         `json:"family,omitempty"`
	Tier           int            `json:"tier,omitempty"`
	EffectSummary  string         `json:"effect_summary,omitempty"`
}

type ConsumableStack struct {
	CatalogID     string `json:"catalog_id"`
	Name          string `json:"name"`
	ItemType      string `json:"item_type"`
	Family        string `json:"family"`
	Tier          int    `json:"tier"`
	Quantity      int    `json:"quantity"`
	EffectSummary string `json:"effect_summary"`
}

type catalogItem struct {
	CatalogID           string
	Name                string
	Slot                string
	Rarity              string
	RequiredClass       string
	RequiredWeaponStyle string
	Stats               map[string]int
}

type shopCatalogItem struct {
	catalogItem
	PriceGold int
}

type consumableCatalogItem struct {
	CatalogID     string
	Name          string
	ItemType      string
	Family        string
	Tier          int
	PriceGold     int
	MinRank       string
	EffectSummary string
	BuildingTypes []string
}

type Service struct {
	mu                     sync.Mutex
	itemsByCharacter       map[string][]EquipmentItem
	consumablesByCharacter map[string]map[string]int
}

var itemCounter uint64

func NewService() *Service {
	return &Service{
		itemsByCharacter:       make(map[string][]EquipmentItem),
		consumablesByCharacter: make(map[string]map[string]int),
	}
}

func (s *Service) GetInventory(character characters.Summary) InventoryView {
	s.mu.Lock()
	defer s.mu.Unlock()

	items := s.ensureCharacterItemsLocked(character)
	consumables := s.ensureConsumablesForCharacterLocked(character)
	return buildView(items, consumables)
}

func (s *Service) EquipItem(character characters.Summary, itemID string) (InventoryView, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	items := s.ensureCharacterItemsLocked(character)
	index := indexByID(items, itemID)
	if index < 0 {
		return InventoryView{}, ErrItemNotOwned
	}

	item := items[index]
	if item.State != "inventory" {
		return InventoryView{}, ErrItemNotEquippable
	}
	if !itemCompatible(character, item) {
		return InventoryView{}, ErrItemNotEquippable
	}

	for i := range items {
		if items[i].State == "equipped" && items[i].Slot == item.Slot {
			items[i].State = "inventory"
		}
	}

	items[index].State = "equipped"
	s.itemsByCharacter[character.CharacterID] = items
	return buildView(items, s.ensureConsumablesForCharacterLocked(character)), nil
}

func (s *Service) UnequipItem(character characters.Summary, slot string) (InventoryView, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	items := s.ensureCharacterItemsLocked(character)
	for i := range items {
		if items[i].State == "equipped" && items[i].Slot == slot {
			items[i].State = "inventory"
			s.itemsByCharacter[character.CharacterID] = items
			return buildView(items, s.ensureConsumablesForCharacterLocked(character)), nil
		}
	}

	return InventoryView{}, ErrSlotNotOccupied
}

func (s *Service) ComputeEquipmentScore(character characters.Summary) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	items := s.ensureCharacterItemsLocked(character)
	score := 0
	for _, item := range items {
		if item.State != "equipped" {
			continue
		}
		score += rarityScore(item.Rarity)
		score += item.EnhancementLevel * 5
		for _, value := range item.Stats {
			score += value
		}
	}
	return score
}

func (s *Service) ListShopInventory(buildingType string, character characters.Summary) []ShopItem {
	items := make([]ShopItem, 0)
	for _, entry := range shopCatalog {
		if len(entry.BuildingTypes) > 0 && !containsString(entry.BuildingTypes, buildingType) {
			continue
		}

		if strings.TrimSpace(entry.Item.RequiredClass) != "" && entry.Item.RequiredClass != character.Class {
			continue
		}
		if strings.TrimSpace(entry.Item.RequiredWeaponStyle) != "" && entry.Item.RequiredWeaponStyle != character.WeaponStyle {
			continue
		}

		items = append(items, ShopItem{
			CatalogID:      entry.Item.CatalogID,
			Name:           entry.Item.Name,
			ItemType:       "equipment",
			Slot:           entry.Item.Slot,
			Rarity:         entry.Item.Rarity,
			PriceGold:      entry.Item.PriceGold,
			RequiredClass:  entry.Item.RequiredClass,
			RequiredWeapon: entry.Item.RequiredWeaponStyle,
			Stats:          copyStats(entry.Item.Stats),
		})
	}

	for _, entry := range consumableShopCatalog {
		if len(entry.BuildingTypes) > 0 && !containsString(entry.BuildingTypes, buildingType) {
			continue
		}
		if !rankAtLeast(character.Rank, entry.MinRank) {
			continue
		}

		items = append(items, ShopItem{
			CatalogID:     entry.CatalogID,
			Name:          entry.Name,
			ItemType:      entry.ItemType,
			PriceGold:     entry.PriceGold,
			Family:        entry.Family,
			Tier:          entry.Tier,
			EffectSummary: entry.EffectSummary,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].PriceGold != items[j].PriceGold {
			return items[i].PriceGold < items[j].PriceGold
		}
		return items[i].CatalogID < items[j].CatalogID
	})

	return items
}

func (s *Service) PurchaseShopItem(character characters.Summary, catalogID string) (InventoryView, *EquipmentItem, *ConsumableStack, int, error) {
	if _, ok := shopCatalogByID[strings.TrimSpace(catalogID)]; ok {
		view, purchased, price, err := s.PurchaseItem(character, catalogID)
		if err != nil {
			return InventoryView{}, nil, nil, 0, err
		}
		return view, &purchased, nil, price, nil
	}

	entry, ok := consumableShopCatalogByID[strings.TrimSpace(catalogID)]
	if !ok || !rankAtLeast(character.Rank, entry.MinRank) {
		return InventoryView{}, nil, nil, 0, ErrCatalogNotFound
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	items := s.ensureCharacterItemsLocked(character)
	consumables := s.ensureConsumablesForCharacterLocked(character)
	consumables[entry.CatalogID]++
	s.consumablesByCharacter[character.CharacterID] = consumables

	view := buildView(items, consumables)
	stack := buildConsumableStack(entry, consumables[entry.CatalogID])
	return view, nil, &stack, entry.PriceGold, nil
}

func (s *Service) PurchaseItem(character characters.Summary, catalogID string) (InventoryView, EquipmentItem, int, error) {
	entry, ok := shopCatalogByID[strings.TrimSpace(catalogID)]
	if !ok {
		return InventoryView{}, EquipmentItem{}, 0, ErrCatalogNotFound
	}

	template := entry
	if !itemCompatible(character, EquipmentItem{RequiredClass: template.RequiredClass, RequiredWeaponStyle: template.RequiredWeaponStyle}) {
		return InventoryView{}, EquipmentItem{}, 0, ErrItemNotEquippable
	}

	purchased := EquipmentItem{
		ItemID:              nextItemID(),
		CatalogID:           template.CatalogID,
		Name:                template.Name,
		Slot:                template.Slot,
		Rarity:              template.Rarity,
		RequiredClass:       template.RequiredClass,
		RequiredWeaponStyle: template.RequiredWeaponStyle,
		EnhancementLevel:    0,
		Durability:          100,
		Stats:               copyStats(template.Stats),
		State:               "inventory",
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	items := s.ensureCharacterItemsLocked(character)
	items = append(items, purchased)
	s.itemsByCharacter[character.CharacterID] = items

	return buildView(items, s.ensureConsumablesForCharacterLocked(character)), purchased, template.PriceGold, nil
}

func (s *Service) GrantItemFromCatalog(character characters.Summary, catalogID string) (InventoryView, EquipmentItem, error) {
	template, ok := catalogByID(strings.TrimSpace(catalogID))
	if !ok {
		return InventoryView{}, EquipmentItem{}, ErrCatalogNotFound
	}
	if !itemCompatible(character, EquipmentItem{RequiredClass: template.RequiredClass, RequiredWeaponStyle: template.RequiredWeaponStyle}) {
		return InventoryView{}, EquipmentItem{}, ErrItemNotEquippable
	}

	reward := EquipmentItem{
		ItemID:              nextItemID(),
		CatalogID:           template.CatalogID,
		Name:                template.Name,
		Slot:                template.Slot,
		Rarity:              template.Rarity,
		RequiredClass:       template.RequiredClass,
		RequiredWeaponStyle: template.RequiredWeaponStyle,
		EnhancementLevel:    0,
		Durability:          100,
		Stats:               copyStats(template.Stats),
		State:               "inventory",
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	items := s.ensureCharacterItemsLocked(character)
	items = append(items, reward)
	s.itemsByCharacter[character.CharacterID] = items

	return buildView(items, s.ensureConsumablesForCharacterLocked(character)), reward, nil
}

func (s *Service) SellItem(character characters.Summary, itemID string) (InventoryView, EquipmentItem, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	items := s.ensureCharacterItemsLocked(character)
	index := indexByID(items, strings.TrimSpace(itemID))
	if index < 0 {
		return InventoryView{}, EquipmentItem{}, 0, ErrItemNotOwned
	}

	target := items[index]
	if target.State == "equipped" {
		return InventoryView{}, EquipmentItem{}, 0, ErrItemEquipped
	}

	price := sellPriceFor(target)
	items = append(items[:index], items[index+1:]...)
	s.itemsByCharacter[character.CharacterID] = items

	return buildView(items, s.ensureConsumablesForCharacterLocked(character)), target, price, nil
}

func (s *Service) ensureCharacterItemsLocked(character characters.Summary) []EquipmentItem {
	if items, ok := s.itemsByCharacter[character.CharacterID]; ok {
		copied := make([]EquipmentItem, len(items))
		copy(copied, items)
		return copied
	}

	items := starterItemsFor(character)
	s.itemsByCharacter[character.CharacterID] = items
	copied := make([]EquipmentItem, len(items))
	copy(copied, items)
	return copied
}

func (s *Service) ensureConsumablesLocked(characterID string) map[string]int {
	if items, ok := s.consumablesByCharacter[characterID]; ok {
		copied := make(map[string]int, len(items))
		for key, value := range items {
			copied[key] = value
		}
		return copied
	}

	s.consumablesByCharacter[characterID] = map[string]int{}
	return map[string]int{}
}

func (s *Service) ensureConsumablesForCharacterLocked(character characters.Summary) map[string]int {
	if items, ok := s.consumablesByCharacter[character.CharacterID]; ok {
		copied := make(map[string]int, len(items))
		for key, value := range items {
			copied[key] = value
		}
		return copied
	}

	starter := starterConsumablesFor(character.Rank)
	s.consumablesByCharacter[character.CharacterID] = starter

	copied := make(map[string]int, len(starter))
	for key, value := range starter {
		copied[key] = value
	}
	return copied
}

func starterItemsFor(character characters.Summary) []EquipmentItem {
	weaponCatalogID := map[string]string{
		"sword_shield": "warrior_sword_starter",
		"great_axe":    "warrior_axe_starter",
		"staff":        "mage_staff_starter",
		"spellbook":    "mage_spellbook_starter",
		"scepter":      "priest_scepter_starter",
		"holy_tome":    "priest_tome_starter",
	}[character.WeaponStyle]

	chestCatalogID := "starter_chest_armor"
	if character.Class != "warrior" {
		chestCatalogID = "starter_chest_cloth"
	}

	catalogIDs := []string{weaponCatalogID, chestCatalogID, "starter_boots"}
	items := make([]EquipmentItem, 0, len(catalogIDs))
	for _, id := range catalogIDs {
		catalog, ok := starterCatalog[id]
		if !ok {
			continue
		}
		items = append(items, EquipmentItem{
			ItemID:              nextItemID(),
			CatalogID:           catalog.CatalogID,
			Name:                catalog.Name,
			Slot:                catalog.Slot,
			Rarity:              catalog.Rarity,
			RequiredClass:       catalog.RequiredClass,
			RequiredWeaponStyle: catalog.RequiredWeaponStyle,
			EnhancementLevel:    0,
			Durability:          100,
			Stats:               copyStats(catalog.Stats),
			State:               "inventory",
		})
	}

	if len(items) > 0 {
		items[0].State = "equipped"
	}
	return items
}

func buildView(items []EquipmentItem, consumables map[string]int) InventoryView {
	equipped := make([]EquipmentItem, 0)
	bag := make([]EquipmentItem, 0)
	for _, item := range items {
		if item.State == "equipped" {
			equipped = append(equipped, item)
		} else {
			bag = append(bag, item)
		}
	}

	sort.Slice(equipped, func(i, j int) bool { return equipped[i].Slot < equipped[j].Slot })
	sort.Slice(bag, func(i, j int) bool { return bag[i].ItemID < bag[j].ItemID })

	score := 0
	for _, item := range equipped {
		score += rarityScore(item.Rarity)
		for _, value := range item.Stats {
			score += value
		}
	}

	return InventoryView{
		EquipmentScore: score,
		Equipped:       equipped,
		Inventory:      bag,
		Consumables:    buildConsumableStacks(consumables),
	}
}

func buildConsumableStacks(consumables map[string]int) []ConsumableStack {
	stacks := make([]ConsumableStack, 0, len(consumables))
	for catalogID, quantity := range consumables {
		if quantity <= 0 {
			continue
		}
		entry, ok := consumableShopCatalogByID[catalogID]
		if !ok {
			continue
		}
		stacks = append(stacks, buildConsumableStack(entry, quantity))
	}

	sort.Slice(stacks, func(i, j int) bool {
		if stacks[i].Tier != stacks[j].Tier {
			return stacks[i].Tier < stacks[j].Tier
		}
		return stacks[i].CatalogID < stacks[j].CatalogID
	})

	return stacks
}

func buildConsumableStack(entry consumableCatalogItem, quantity int) ConsumableStack {
	return ConsumableStack{
		CatalogID:     entry.CatalogID,
		Name:          entry.Name,
		ItemType:      entry.ItemType,
		Family:        entry.Family,
		Tier:          entry.Tier,
		Quantity:      quantity,
		EffectSummary: entry.EffectSummary,
	}
}

func starterConsumablesFor(rank string) map[string]int {
	potions := combat.DefaultPotionBag(rank)
	items := make(map[string]int, len(potions))
	for _, potion := range potions {
		quantity := potion.Quantity
		if quantity < 6 {
			quantity = 6
		}
		items[potion.PotionID] = quantity
	}
	return items
}

func itemCompatible(character characters.Summary, item EquipmentItem) bool {
	if strings.TrimSpace(item.RequiredClass) != "" && item.RequiredClass != character.Class {
		return false
	}
	if strings.TrimSpace(item.RequiredWeaponStyle) != "" && item.RequiredWeaponStyle != character.WeaponStyle {
		return false
	}
	return true
}

func indexByID(items []EquipmentItem, itemID string) int {
	for i := range items {
		if items[i].ItemID == itemID {
			return i
		}
	}
	return -1
}

func rarityScore(rarity string) int {
	switch strings.ToLower(strings.TrimSpace(rarity)) {
	case "epic":
		return 30
	case "rare":
		return 20
	default:
		return 10
	}
}

func copyStats(input map[string]int) map[string]int {
	output := make(map[string]int, len(input))
	for key, value := range input {
		output[key] = value
	}
	return output
}

func nextItemID() string {
	return fmt.Sprintf("item_%d_%06d", time.Now().UnixNano(), atomic.AddUint64(&itemCounter, 1))
}

func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func rankAtLeast(currentRank, requiredRank string) bool {
	order := map[string]int{
		"low":  1,
		"mid":  2,
		"high": 3,
	}

	if strings.TrimSpace(requiredRank) == "" {
		return true
	}

	return order[strings.TrimSpace(currentRank)] >= order[strings.TrimSpace(requiredRank)]
}

func (s *Service) BuildPotionLoadout(character characters.Summary, potionIDs []string) ([]combat.PotionItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(potionIDs) == 0 {
		return []combat.PotionItem{}, nil
	}
	if len(potionIDs) > 2 {
		return nil, ErrConsumableMissing
	}

	consumables := s.ensureConsumablesForCharacterLocked(character)
	loadout := make([]combat.PotionItem, 0, len(potionIDs))
	seen := map[string]struct{}{}
	for _, potionID := range potionIDs {
		potionID = strings.TrimSpace(potionID)
		if potionID == "" {
			return nil, ErrConsumableMissing
		}
		if _, ok := seen[potionID]; ok {
			return nil, ErrConsumableMissing
		}
		seen[potionID] = struct{}{}
		entry, ok := consumableShopCatalogByID[potionID]
		if !ok || !rankAtLeast(character.Rank, entry.MinRank) {
			return nil, ErrConsumableMissing
		}
		quantity := consumables[potionID]
		if quantity <= 0 {
			return nil, ErrConsumableMissing
		}
		potion, ok := combat.PotionCatalog[potionID]
		if !ok {
			return nil, ErrConsumableMissing
		}
		potion.Quantity = quantity
		loadout = append(loadout, potion)
	}

	return loadout, nil
}

func (s *Service) ConsumeConsumables(characterID string, usage map[string]int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	consumables := s.ensureConsumablesLocked(characterID)
	for catalogID, quantity := range usage {
		if quantity <= 0 {
			continue
		}
		remaining := consumables[catalogID] - quantity
		if remaining <= 0 {
			delete(consumables, catalogID)
			continue
		}
		consumables[catalogID] = remaining
	}
	s.consumablesByCharacter[characterID] = consumables
}

func sellPriceFor(item EquipmentItem) int {
	if entry, ok := shopCatalogByID[item.CatalogID]; ok {
		if entry.PriceGold <= 1 {
			return 1
		}
		return entry.PriceGold / 2
	}

	base := 8 + rarityScore(item.Rarity)
	for _, value := range item.Stats {
		base += value
	}
	if base <= 1 {
		return 1
	}
	return base / 2
}

func catalogByID(catalogID string) (catalogItem, bool) {
	if entry, ok := shopCatalogByID[catalogID]; ok {
		return catalogItem{
			CatalogID:           entry.CatalogID,
			Name:                entry.Name,
			Slot:                entry.Slot,
			Rarity:              entry.Rarity,
			RequiredClass:       entry.RequiredClass,
			RequiredWeaponStyle: entry.RequiredWeaponStyle,
			Stats:               copyStats(entry.Stats),
		}, true
	}

	if entry, ok := starterCatalog[catalogID]; ok {
		return catalogItem{
			CatalogID:           entry.CatalogID,
			Name:                entry.Name,
			Slot:                entry.Slot,
			Rarity:              entry.Rarity,
			RequiredClass:       entry.RequiredClass,
			RequiredWeaponStyle: entry.RequiredWeaponStyle,
			Stats:               copyStats(entry.Stats),
		}, true
	}

	if entry, ok := dungeonRewardCatalog[catalogID]; ok {
		return catalogItem{
			CatalogID:           entry.CatalogID,
			Name:                entry.Name,
			Slot:                entry.Slot,
			Rarity:              entry.Rarity,
			RequiredClass:       entry.RequiredClass,
			RequiredWeaponStyle: entry.RequiredWeaponStyle,
			Stats:               copyStats(entry.Stats),
		}, true
	}

	return catalogItem{}, false
}

var starterCatalog = map[string]catalogItem{
	"warrior_sword_starter": {
		CatalogID:           "warrior_sword_starter",
		Name:                "Recruit Sword",
		Slot:                "weapon",
		Rarity:              "common",
		RequiredClass:       "warrior",
		RequiredWeaponStyle: "sword_shield",
		Stats:               map[string]int{"physical_attack": 6},
	},
	"warrior_axe_starter": {
		CatalogID:           "warrior_axe_starter",
		Name:                "Recruit Axe",
		Slot:                "weapon",
		Rarity:              "common",
		RequiredClass:       "warrior",
		RequiredWeaponStyle: "great_axe",
		Stats:               map[string]int{"physical_attack": 7},
	},
	"mage_staff_starter": {
		CatalogID:           "mage_staff_starter",
		Name:                "Ashwood Staff",
		Slot:                "weapon",
		Rarity:              "common",
		RequiredClass:       "mage",
		RequiredWeaponStyle: "staff",
		Stats:               map[string]int{"magic_attack": 8},
	},
	"mage_spellbook_starter": {
		CatalogID:           "mage_spellbook_starter",
		Name:                "Trainee Spellbook",
		Slot:                "weapon",
		Rarity:              "common",
		RequiredClass:       "mage",
		RequiredWeaponStyle: "spellbook",
		Stats:               map[string]int{"magic_attack": 8},
	},
	"priest_scepter_starter": {
		CatalogID:           "priest_scepter_starter",
		Name:                "Pilgrim Scepter",
		Slot:                "weapon",
		Rarity:              "common",
		RequiredClass:       "priest",
		RequiredWeaponStyle: "scepter",
		Stats:               map[string]int{"healing_power": 6, "magic_attack": 4},
	},
	"priest_tome_starter": {
		CatalogID:           "priest_tome_starter",
		Name:                "Pilgrim Tome",
		Slot:                "weapon",
		Rarity:              "common",
		RequiredClass:       "priest",
		RequiredWeaponStyle: "holy_tome",
		Stats:               map[string]int{"healing_power": 5, "magic_attack": 5},
	},
	"starter_chest_cloth": {
		CatalogID: "starter_chest_cloth",
		Name:      "Novice Robe",
		Slot:      "chest",
		Rarity:    "common",
		Stats:     map[string]int{"magic_defense": 3},
	},
	"starter_chest_armor": {
		CatalogID: "starter_chest_armor",
		Name:      "Novice Armor",
		Slot:      "chest",
		Rarity:    "common",
		Stats:     map[string]int{"physical_defense": 4, "max_hp": 12},
	},
	"starter_boots": {
		CatalogID: "starter_boots",
		Name:      "Trail Boots",
		Slot:      "boots",
		Rarity:    "common",
		Stats:     map[string]int{"speed": 2},
	},
}

var dungeonRewardCatalog = map[string]catalogItem{
	"gravewake_marchers_blue": {
		CatalogID: "gravewake_marchers_blue",
		Name:      "Gravewake Marchers",
		Slot:      "boots",
		Rarity:    "blue",
		Stats:     map[string]int{"speed": 4, "max_hp": 45, "physical_defense": 4},
	},
	"gravewake_shackles_blue": {
		CatalogID: "gravewake_shackles_blue",
		Name:      "Gravewake Shackles",
		Slot:      "wrist",
		Rarity:    "blue",
		Stats:     map[string]int{"physical_attack": 8, "speed": 2},
	},
	"gravewake_seal_purple": {
		CatalogID: "gravewake_seal_purple",
		Name:      "Gravewake Seal",
		Slot:      "ring",
		Rarity:    "purple",
		Stats:     map[string]int{"magic_attack": 10, "physical_attack": 10},
	},
	"gravewake_hood_purple": {
		CatalogID: "gravewake_hood_purple",
		Name:      "Gravewake Hood",
		Slot:      "head",
		Rarity:    "purple",
		Stats:     map[string]int{"max_hp": 60, "physical_defense": 6, "magic_defense": 6},
	},
	"gravewake_reliquary_gold": {
		CatalogID: "gravewake_reliquary_gold",
		Name:      "Gravewake Reliquary",
		Slot:      "necklace",
		Rarity:    "gold",
		Stats:     map[string]int{"max_hp": 45, "magic_attack": 6, "healing_power": 8},
	},
	"gravewake_vestment_gold": {
		CatalogID: "gravewake_vestment_gold",
		Name:      "Gravewake Vestment",
		Slot:      "chest",
		Rarity:    "gold",
		Stats:     map[string]int{"max_hp": 90, "physical_defense": 10, "magic_defense": 10},
	},
	"gravewake_vestment_red": {
		CatalogID: "gravewake_vestment_red",
		Name:      "Gravewake Vestment",
		Slot:      "chest",
		Rarity:    "red",
		Stats:     map[string]int{"max_hp": 115, "physical_defense": 13, "magic_defense": 13},
	},
	"gravewake_seal_red": {
		CatalogID: "gravewake_seal_red",
		Name:      "Gravewake Seal",
		Slot:      "ring",
		Rarity:    "red",
		Stats:     map[string]int{"magic_attack": 14, "physical_attack": 14},
	},
	"gravewake_reliquary_prismatic": {
		CatalogID: "gravewake_reliquary_prismatic",
		Name:      "Gravewake Reliquary",
		Slot:      "necklace",
		Rarity:    "prismatic",
		Stats:     map[string]int{"max_hp": 70, "magic_attack": 12, "healing_power": 14},
	},
	"dunescourge_burrowstep_blue": {
		CatalogID: "dunescourge_burrowstep_blue",
		Name:      "Dunescourge Burrowstep Boots",
		Slot:      "boots",
		Rarity:    "blue",
		Stats:     map[string]int{"speed": 12, "max_hp": 230, "physical_defense": 18},
	},
	"dunescourge_coilguards_blue": {
		CatalogID: "dunescourge_coilguards_blue",
		Name:      "Dunescourge Coilguards",
		Slot:      "wrist",
		Rarity:    "blue",
		Stats:     map[string]int{"physical_attack": 44, "speed": 11},
	},
	"dunescourge_fang_ring_purple": {
		CatalogID: "dunescourge_fang_ring_purple",
		Name:      "Dunescourge Fang Ring",
		Slot:      "ring",
		Rarity:    "purple",
		Stats:     map[string]int{"physical_attack": 56, "healing_power": 42},
	},
	"dunescourge_crownshell_purple": {
		CatalogID: "dunescourge_crownshell_purple",
		Name:      "Dunescourge Crownshell",
		Slot:      "head",
		Rarity:    "purple",
		Stats:     map[string]int{"max_hp": 360, "physical_defense": 30, "magic_defense": 30},
	},
	"dunescourge_heartspine_chain_gold": {
		CatalogID: "dunescourge_heartspine_chain_gold",
		Name:      "Dunescourge Heartspine Chain",
		Slot:      "necklace",
		Rarity:    "gold",
		Stats:     map[string]int{"max_hp": 265, "magic_attack": 36, "healing_power": 36},
	},
	"dunescourge_carapace_mail_gold": {
		CatalogID: "dunescourge_carapace_mail_gold",
		Name:      "Dunescourge Carapace Mail",
		Slot:      "chest",
		Rarity:    "gold",
		Stats:     map[string]int{"max_hp": 470, "physical_defense": 45, "magic_defense": 40},
	},
	"dunescourge_carapace_mail_red": {
		CatalogID: "dunescourge_carapace_mail_red",
		Name:      "Dunescourge Carapace Mail",
		Slot:      "chest",
		Rarity:    "red",
		Stats:     map[string]int{"max_hp": 520, "physical_defense": 50, "magic_defense": 45},
	},
	"dunescourge_fang_ring_red": {
		CatalogID: "dunescourge_fang_ring_red",
		Name:      "Dunescourge Fang Ring",
		Slot:      "ring",
		Rarity:    "red",
		Stats:     map[string]int{"physical_attack": 64, "healing_power": 50},
	},
	"dunescourge_heartspine_chain_prismatic": {
		CatalogID: "dunescourge_heartspine_chain_prismatic",
		Name:      "Dunescourge Heartspine Chain",
		Slot:      "necklace",
		Rarity:    "prismatic",
		Stats:     map[string]int{"max_hp": 310, "magic_attack": 42, "healing_power": 42},
	},
}

type shopEntry struct {
	Item          shopCatalogItem
	BuildingTypes []string
}

var shopCatalog = []shopEntry{
	{
		Item: shopCatalogItem{
			catalogItem: catalogItem{
				CatalogID:           "warrior_sword_bronze",
				Name:                "Bronze Longsword",
				Slot:                "weapon",
				Rarity:              "common",
				RequiredClass:       "warrior",
				RequiredWeaponStyle: "sword_shield",
				Stats:               map[string]int{"physical_attack": 10},
			},
			PriceGold: 65,
		},
		BuildingTypes: []string{"equipment_shop"},
	},
	{
		Item: shopCatalogItem{
			catalogItem: catalogItem{
				CatalogID:           "mage_staff_oak",
				Name:                "Oak Arcane Staff",
				Slot:                "weapon",
				Rarity:              "common",
				RequiredClass:       "mage",
				RequiredWeaponStyle: "staff",
				Stats:               map[string]int{"magic_attack": 12},
			},
			PriceGold: 65,
		},
		BuildingTypes: []string{"equipment_shop"},
	},
	{
		Item: shopCatalogItem{
			catalogItem: catalogItem{
				CatalogID:           "priest_scepter_ash",
				Name:                "Ashwood Scepter",
				Slot:                "weapon",
				Rarity:              "common",
				RequiredClass:       "priest",
				RequiredWeaponStyle: "scepter",
				Stats:               map[string]int{"healing_power": 10, "magic_attack": 6},
			},
			PriceGold: 62,
		},
		BuildingTypes: []string{"equipment_shop"},
	},
	{
		Item: shopCatalogItem{
			catalogItem: catalogItem{
				CatalogID: "armor_chain_vest",
				Name:      "Chain Vest",
				Slot:      "chest",
				Rarity:    "common",
				Stats:     map[string]int{"physical_defense": 8, "max_hp": 14},
			},
			PriceGold: 58,
		},
		BuildingTypes: []string{"equipment_shop"},
	},
	{
		Item: shopCatalogItem{
			catalogItem: catalogItem{
				CatalogID: "armor_apprentice_robe",
				Name:      "Apprentice Robe",
				Slot:      "chest",
				Rarity:    "common",
				Stats:     map[string]int{"magic_defense": 8},
			},
			PriceGold: 58,
		},
		BuildingTypes: []string{"equipment_shop"},
	},
	{
		Item: shopCatalogItem{
			catalogItem: catalogItem{
				CatalogID: "boots_pathfinder",
				Name:      "Pathfinder Boots",
				Slot:      "boots",
				Rarity:    "common",
				Stats:     map[string]int{"speed": 5},
			},
			PriceGold: 40,
		},
		BuildingTypes: []string{"equipment_shop"},
	},
}

var shopCatalogByID = func() map[string]shopCatalogItem {
	items := make(map[string]shopCatalogItem, len(shopCatalog))
	for _, entry := range shopCatalog {
		items[entry.Item.CatalogID] = entry.Item
	}
	return items
}()

var consumableShopCatalog = []consumableCatalogItem{
	{CatalogID: "potion_hp_t1", Name: "Minor HP Potion", ItemType: "consumable", Family: "hp", Tier: 1, PriceGold: 12, MinRank: "low", EffectSummary: "Restore 25% max HP, capped at 220.", BuildingTypes: []string{"apothecary"}},
	{CatalogID: "potion_atk_t1", Name: "Minor Attack Potion", ItemType: "consumable", Family: "atk", Tier: 1, PriceGold: 14, MinRank: "low", EffectSummary: "Increase primary attack by 10% for 3 rounds.", BuildingTypes: []string{"apothecary"}},
	{CatalogID: "potion_def_t1", Name: "Minor Defense Potion", ItemType: "consumable", Family: "def", Tier: 1, PriceGold: 14, MinRank: "low", EffectSummary: "Increase defenses by 10% for 3 rounds.", BuildingTypes: []string{"apothecary"}},
	{CatalogID: "potion_spd_t1", Name: "Minor Speed Potion", ItemType: "consumable", Family: "spd", Tier: 1, PriceGold: 14, MinRank: "low", EffectSummary: "Increase speed by 8% for 3 rounds.", BuildingTypes: []string{"apothecary"}},
	{CatalogID: "potion_hp_t2", Name: "Standard HP Potion", ItemType: "consumable", Family: "hp", Tier: 2, PriceGold: 22, MinRank: "mid", EffectSummary: "Restore 35% max HP, capped at 520.", BuildingTypes: []string{"apothecary"}},
	{CatalogID: "potion_atk_t2", Name: "Standard Attack Potion", ItemType: "consumable", Family: "atk", Tier: 2, PriceGold: 24, MinRank: "mid", EffectSummary: "Increase primary attack by 16% for 3 rounds.", BuildingTypes: []string{"apothecary"}},
	{CatalogID: "potion_def_t2", Name: "Standard Defense Potion", ItemType: "consumable", Family: "def", Tier: 2, PriceGold: 24, MinRank: "mid", EffectSummary: "Increase defenses by 16% for 3 rounds.", BuildingTypes: []string{"apothecary"}},
	{CatalogID: "potion_spd_t2", Name: "Standard Speed Potion", ItemType: "consumable", Family: "spd", Tier: 2, PriceGold: 24, MinRank: "mid", EffectSummary: "Increase speed by 12% for 3 rounds.", BuildingTypes: []string{"apothecary"}},
	{CatalogID: "potion_hp_t3", Name: "Superior HP Potion", ItemType: "consumable", Family: "hp", Tier: 3, PriceGold: 36, MinRank: "high", EffectSummary: "Restore 45% max HP, capped at 980.", BuildingTypes: []string{"apothecary"}},
	{CatalogID: "potion_atk_t3", Name: "Superior Attack Potion", ItemType: "consumable", Family: "atk", Tier: 3, PriceGold: 38, MinRank: "high", EffectSummary: "Increase primary attack by 24% for 4 rounds.", BuildingTypes: []string{"apothecary"}},
	{CatalogID: "potion_def_t3", Name: "Superior Defense Potion", ItemType: "consumable", Family: "def", Tier: 3, PriceGold: 38, MinRank: "high", EffectSummary: "Increase defenses by 24% for 4 rounds.", BuildingTypes: []string{"apothecary"}},
	{CatalogID: "potion_spd_t3", Name: "Superior Speed Potion", ItemType: "consumable", Family: "spd", Tier: 3, PriceGold: 38, MinRank: "high", EffectSummary: "Increase speed by 18% for 4 rounds.", BuildingTypes: []string{"apothecary"}},
}

var consumableShopCatalogByID = func() map[string]consumableCatalogItem {
	items := make(map[string]consumableCatalogItem, len(consumableShopCatalog))
	for _, entry := range consumableShopCatalog {
		if potion, ok := combat.PotionCatalog[entry.CatalogID]; ok {
			entry.Family = potion.Family
			entry.Tier = potion.Tier
		}
		items[entry.CatalogID] = entry
	}
	return items
}()
