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
)

var (
	ErrItemNotOwned      = errors.New("item not owned")
	ErrItemNotEquippable = errors.New("item not equippable")
	ErrSlotNotOccupied   = errors.New("slot not occupied")
	ErrCatalogNotFound   = errors.New("catalog item not found")
	ErrItemEquipped      = errors.New("item is currently equipped")
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
	EquipmentScore int             `json:"equipment_score"`
	Equipped       []EquipmentItem `json:"equipped"`
	Inventory      []EquipmentItem `json:"inventory"`
}

type ShopItem struct {
	CatalogID      string         `json:"catalog_id"`
	Name           string         `json:"name"`
	Slot           string         `json:"slot"`
	Rarity         string         `json:"rarity"`
	PriceGold      int            `json:"price_gold"`
	RequiredClass  string         `json:"required_class,omitempty"`
	RequiredWeapon string         `json:"required_weapon_style,omitempty"`
	Stats          map[string]int `json:"stats"`
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

type Service struct {
	mu               sync.Mutex
	itemsByCharacter map[string][]EquipmentItem
}

var itemCounter uint64

func NewService() *Service {
	return &Service{itemsByCharacter: make(map[string][]EquipmentItem)}
}

func (s *Service) GetInventory(character characters.Summary) InventoryView {
	s.mu.Lock()
	defer s.mu.Unlock()

	items := s.ensureCharacterItemsLocked(character)
	return buildView(items)
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
	return buildView(items), nil
}

func (s *Service) UnequipItem(character characters.Summary, slot string) (InventoryView, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	items := s.ensureCharacterItemsLocked(character)
	for i := range items {
		if items[i].State == "equipped" && items[i].Slot == slot {
			items[i].State = "inventory"
			s.itemsByCharacter[character.CharacterID] = items
			return buildView(items), nil
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
			Slot:           entry.Item.Slot,
			Rarity:         entry.Item.Rarity,
			PriceGold:      entry.Item.PriceGold,
			RequiredClass:  entry.Item.RequiredClass,
			RequiredWeapon: entry.Item.RequiredWeaponStyle,
			Stats:          copyStats(entry.Item.Stats),
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

	return buildView(items), purchased, template.PriceGold, nil
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

	return buildView(items), target, price, nil
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

func buildView(items []EquipmentItem) InventoryView {
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

	return InventoryView{EquipmentScore: score, Equipped: equipped, Inventory: bag}
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
		Stats:     map[string]int{"magic_defense": 3, "max_mp": 10},
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
		BuildingTypes: []string{"weapon_shop", "general_store"},
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
		BuildingTypes: []string{"weapon_shop", "general_store"},
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
		BuildingTypes: []string{"weapon_shop", "general_store"},
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
		BuildingTypes: []string{"armor_shop", "general_store"},
	},
	{
		Item: shopCatalogItem{
			catalogItem: catalogItem{
				CatalogID: "armor_apprentice_robe",
				Name:      "Apprentice Robe",
				Slot:      "chest",
				Rarity:    "common",
				Stats:     map[string]int{"magic_defense": 8, "max_mp": 16},
			},
			PriceGold: 58,
		},
		BuildingTypes: []string{"armor_shop", "general_store"},
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
		BuildingTypes: []string{"armor_shop", "general_store"},
	},
}

var shopCatalogByID = func() map[string]shopCatalogItem {
	items := make(map[string]shopCatalogItem, len(shopCatalog))
	for _, entry := range shopCatalog {
		items[entry.Item.CatalogID] = entry.Item
	}
	return items
}()
