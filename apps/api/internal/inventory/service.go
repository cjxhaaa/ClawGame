package inventory

import (
	"errors"
	"fmt"
	"hash/fnv"
	"math"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"clawgame/apps/api/internal/characters"
	"clawgame/apps/api/internal/combat"
)

var (
	ErrItemNotOwned       = errors.New("item not owned")
	ErrItemNotEquippable  = errors.New("item not equippable")
	ErrSlotNotOccupied    = errors.New("slot not occupied")
	ErrCatalogNotFound    = errors.New("catalog item not found")
	ErrItemEquipped       = errors.New("item is currently equipped")
	ErrConsumableMissing  = errors.New("consumable missing")
	ErrItemNotEnhanceable = errors.New("item not enhanceable")
	ErrEnhancementCap     = errors.New("enhancement cap reached")
	ErrItemNotReforgeable = errors.New("item not reforgeable")
	ErrReforgeNotPending  = errors.New("reforge result not pending")
)

const EnhancementMaterialKey = "enhancement_shard"
const ReforgeMaterialKey = "reforge_stone"

type EquipmentAffix struct {
	AffixKey string `json:"affix_key"`
	Value    int    `json:"value"`
}

type PendingReforgeResult struct {
	AttemptID        string           `json:"attempt_id"`
	MaterialKey      string           `json:"material_key"`
	MaterialQuantity int              `json:"material_quantity"`
	PreviousAffixes  []EquipmentAffix `json:"previous_affixes"`
	PreviewAffixes   []EquipmentAffix `json:"preview_affixes"`
	CreatedAt        string           `json:"created_at"`
}

type EquipmentItem struct {
	ItemID                string                `json:"item_id"`
	CatalogID             string                `json:"catalog_id"`
	Name                  string                `json:"name"`
	Slot                  string                `json:"slot"`
	Rarity                string                `json:"rarity"`
	RequiredClass         string                `json:"required_class,omitempty"`
	RequiredWeaponStyle   string                `json:"required_weapon_style,omitempty"`
	EnhancementLevel      int                   `json:"enhancement_level"`
	Durability            int                   `json:"durability"`
	BaseStats             map[string]int        `json:"base_stats,omitempty"`
	Stats                 map[string]int        `json:"stats"`
	PassiveAffix          map[string]any        `json:"passive_affix,omitempty"`
	ExtraAffixes          []EquipmentAffix      `json:"extra_affixes,omitempty"`
	PendingReforge        *PendingReforgeResult `json:"pending_reforge,omitempty"`
	State                 string                `json:"state"`
	EnhancementPreviewPct float64               `json:"enhancement_preview_pct,omitempty"`
}

type InventoryView struct {
	EquipmentScore       int                              `json:"equipment_score"`
	Equipped             []EquipmentItem                  `json:"equipped"`
	Inventory            []EquipmentItem                  `json:"inventory"`
	Consumables          []ConsumableStack                `json:"consumables"`
	SlotEnhancements     []characters.SlotEnhancementView `json:"slot_enhancements"`
	UpgradeHints         []UpgradeHint                    `json:"upgrade_hints"`
	PotionLoadoutOptions []PotionLoadoutOption            `json:"potion_loadout_options"`
}

type EnhancementQuote struct {
	ItemID            string           `json:"item_id"`
	TargetSlot        string           `json:"target_slot"`
	CurrentLevel      int              `json:"current_level"`
	NextLevel         int              `json:"next_level"`
	GoldCost          int              `json:"gold_cost"`
	MaterialCost      []map[string]any `json:"material_cost"`
	PreviewBonusPct   float64          `json:"preview_bonus_pct"`
	MaxEnhancement    int              `json:"max_enhancement"`
	EnhancementTarget string           `json:"enhancement_target"`
}

type UpgradeHint struct {
	Source            string `json:"source"`
	ItemID            string `json:"item_id,omitempty"`
	CatalogID         string `json:"catalog_id,omitempty"`
	Name              string `json:"name"`
	Slot              string `json:"slot"`
	ScoreDelta        int    `json:"score_delta"`
	PriceGold         int    `json:"price_gold,omitempty"`
	Affordable        bool   `json:"affordable"`
	DirectlyEquipable bool   `json:"directly_equippable"`
}

type PotionLoadoutOption struct {
	CatalogID     string `json:"catalog_id"`
	Name          string `json:"name"`
	Family        string `json:"family"`
	Tier          int    `json:"tier"`
	QuantityOwned int    `json:"quantity_owned"`
	PriceGold     int    `json:"price_gold"`
	AvailableNow  bool   `json:"available_now"`
	CanPurchase   bool   `json:"can_purchase"`
	Recommended   bool   `json:"recommended"`
	EffectSummary string `json:"effect_summary"`
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
	slotEnhancementsByChar map[string]map[string]int
}

var itemCounter uint64

func NewService() *Service {
	return &Service{
		itemsByCharacter:       make(map[string][]EquipmentItem),
		consumablesByCharacter: make(map[string]map[string]int),
		slotEnhancementsByChar: make(map[string]map[string]int),
	}
}

func (s *Service) GetInventory(character characters.Summary) InventoryView {
	s.mu.Lock()
	defer s.mu.Unlock()

	items := s.ensureCharacterItemsLocked(character)
	consumables := s.ensureConsumablesForCharacterLocked(character)
	return buildView(character, items, consumables, s.ensureSlotEnhancementsLocked(character.CharacterID))
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
	return buildView(character, items, s.ensureConsumablesForCharacterLocked(character), s.ensureSlotEnhancementsLocked(character.CharacterID)), nil
}

func (s *Service) UnequipItem(character characters.Summary, slot string) (InventoryView, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	items := s.ensureCharacterItemsLocked(character)
	for i := range items {
		if items[i].State == "equipped" && items[i].Slot == slot {
			items[i].State = "inventory"
			s.itemsByCharacter[character.CharacterID] = items
			return buildView(character, items, s.ensureConsumablesForCharacterLocked(character), s.ensureSlotEnhancementsLocked(character.CharacterID)), nil
		}
	}

	return InventoryView{}, ErrSlotNotOccupied
}

func (s *Service) ComputeEquipmentScore(character characters.Summary) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	items := s.ensureCharacterItemsLocked(character)
	score := 0
	slotEnhancements := s.ensureSlotEnhancementsLocked(character.CharacterID)
	for _, item := range items {
		if item.State != "equipped" {
			continue
		}
		score += rarityScore(item.Rarity)
		score += slotEnhancements[item.Slot] * 5
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

	view := buildView(character, items, consumables, s.ensureSlotEnhancementsLocked(character.CharacterID))
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

	purchased := buildEquipmentItemFromCatalog(catalogItem{
		CatalogID:           template.CatalogID,
		Name:                template.Name,
		Slot:                template.Slot,
		Rarity:              template.Rarity,
		RequiredClass:       template.RequiredClass,
		RequiredWeaponStyle: template.RequiredWeaponStyle,
		Stats:               copyStats(template.Stats),
	}, "inventory")

	s.mu.Lock()
	defer s.mu.Unlock()

	items := s.ensureCharacterItemsLocked(character)
	items = append(items, purchased)
	s.itemsByCharacter[character.CharacterID] = items

	return buildView(character, items, s.ensureConsumablesForCharacterLocked(character), s.ensureSlotEnhancementsLocked(character.CharacterID)), purchased, template.PriceGold, nil
}

func (s *Service) GrantItemFromCatalog(character characters.Summary, catalogID string) (InventoryView, EquipmentItem, error) {
	template, ok := catalogByID(strings.TrimSpace(catalogID))
	if !ok {
		return InventoryView{}, EquipmentItem{}, ErrCatalogNotFound
	}
	if !itemCompatible(character, EquipmentItem{RequiredClass: template.RequiredClass, RequiredWeaponStyle: template.RequiredWeaponStyle}) {
		return InventoryView{}, EquipmentItem{}, ErrItemNotEquippable
	}

	reward := buildEquipmentItemFromCatalog(template, "inventory")

	s.mu.Lock()
	defer s.mu.Unlock()

	items := s.ensureCharacterItemsLocked(character)
	items = append(items, reward)
	s.itemsByCharacter[character.CharacterID] = items

	return buildView(character, items, s.ensureConsumablesForCharacterLocked(character), s.ensureSlotEnhancementsLocked(character.CharacterID)), reward, nil
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

	return buildView(character, items, s.ensureConsumablesForCharacterLocked(character), s.ensureSlotEnhancementsLocked(character.CharacterID)), target, price, nil
}

func (s *Service) SalvageItem(character characters.Summary, itemID string) (InventoryView, EquipmentItem, []map[string]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	items := s.ensureCharacterItemsLocked(character)
	index := indexByID(items, strings.TrimSpace(itemID))
	if index < 0 {
		return InventoryView{}, EquipmentItem{}, nil, ErrItemNotOwned
	}
	target := items[index]
	if target.State == "equipped" {
		return InventoryView{}, EquipmentItem{}, nil, ErrItemEquipped
	}

	items = append(items[:index], items[index+1:]...)
	s.itemsByCharacter[character.CharacterID] = items
	drops := []map[string]any{{
		"material_key": EnhancementMaterialKey,
		"quantity":     salvageYieldFor(target),
	}}
	return buildView(character, items, s.ensureConsumablesForCharacterLocked(character), s.ensureSlotEnhancementsLocked(character.CharacterID)), target, drops, nil
}

func (s *Service) GetEnhancementQuote(character characters.Summary, itemID, slot string) (EquipmentItem, EnhancementQuote, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	items := s.ensureCharacterItemsLocked(character)
	targetSlot, item, err := s.resolveEnhancementTargetLocked(items, itemID, slot)
	if err != nil {
		return EquipmentItem{}, EnhancementQuote{}, err
	}
	slotEnhancements := s.ensureSlotEnhancementsLocked(character.CharacterID)
	currentLevel := slotEnhancements[targetSlot]
	if currentLevel >= maxEnhancementLevelFor(item) {
		return EquipmentItem{}, EnhancementQuote{}, ErrEnhancementCap
	}
	if !isEnhanceable(item) {
		return EquipmentItem{}, EnhancementQuote{}, ErrItemNotEnhanceable
	}
	annotated := applySlotEnhancement(item, currentLevel)
	nextLevel := currentLevel + 1
	return annotated, EnhancementQuote{
		ItemID:            annotated.ItemID,
		CurrentLevel:      currentLevel,
		NextLevel:         nextLevel,
		GoldCost:          enhancementGoldCost(item, currentLevel),
		MaterialCost:      enhancementMaterialCost(item, currentLevel),
		PreviewBonusPct:   enhancementPreviewPct(nextLevel),
		MaxEnhancement:    maxEnhancementLevelFor(item),
		EnhancementTarget: "base_stats_only",
		TargetSlot:        targetSlot,
	}, nil
}

func (s *Service) EnhanceItem(character characters.Summary, itemID, slot string) (InventoryView, EquipmentItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	items := s.ensureCharacterItemsLocked(character)
	targetSlot, item, err := s.resolveEnhancementTargetLocked(items, itemID, slot)
	if err != nil {
		return InventoryView{}, EquipmentItem{}, err
	}
	if !isEnhanceable(item) {
		return InventoryView{}, EquipmentItem{}, ErrItemNotEnhanceable
	}
	slotEnhancements := s.ensureSlotEnhancementsLocked(character.CharacterID)
	if slotEnhancements[targetSlot] >= maxEnhancementLevelFor(item) {
		return InventoryView{}, EquipmentItem{}, ErrEnhancementCap
	}
	slotEnhancements[targetSlot]++
	s.slotEnhancementsByChar[character.CharacterID] = slotEnhancements

	enhanced := applySlotEnhancement(item, slotEnhancements[targetSlot])
	return buildView(character, items, s.ensureConsumablesForCharacterLocked(character), slotEnhancements), enhanced, nil
}

func (s *Service) ReforgeItem(character characters.Summary, itemID string) (InventoryView, EquipmentItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	items := s.ensureCharacterItemsLocked(character)
	index := indexByID(items, strings.TrimSpace(itemID))
	if index < 0 {
		return InventoryView{}, EquipmentItem{}, ErrItemNotOwned
	}
	item := items[index]
	if !isReforgeable(item) {
		return InventoryView{}, EquipmentItem{}, ErrItemNotReforgeable
	}

	attemptID := nextItemID()
	preview := rollExtraAffixes(item.ItemID, item.Slot, item.Rarity, attemptID)
	items[index].PendingReforge = &PendingReforgeResult{
		AttemptID:        attemptID,
		MaterialKey:      ReforgeMaterialKey,
		MaterialQuantity: reforgeStoneCost(item),
		PreviousAffixes:  cloneAffixes(item.ExtraAffixes),
		PreviewAffixes:   preview,
		CreatedAt:        time.Now().Format(time.RFC3339),
	}
	s.itemsByCharacter[character.CharacterID] = items
	view := buildView(character, items, s.ensureConsumablesForCharacterLocked(character), s.ensureSlotEnhancementsLocked(character.CharacterID))
	reforged := findItemByID(view.Equipped, view.Inventory, item.ItemID)
	return view, reforged, nil
}

func (s *Service) GetReforgeCost(character characters.Summary, itemID string) (EquipmentItem, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	items := s.ensureCharacterItemsLocked(character)
	index := indexByID(items, strings.TrimSpace(itemID))
	if index < 0 {
		return EquipmentItem{}, 0, ErrItemNotOwned
	}
	if !isReforgeable(items[index]) {
		return EquipmentItem{}, 0, ErrItemNotReforgeable
	}
	return items[index], reforgeStoneCost(items[index]), nil
}

func (s *Service) ValidateReforgeTarget(character characters.Summary, itemID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	items := s.ensureCharacterItemsLocked(character)
	index := indexByID(items, strings.TrimSpace(itemID))
	if index < 0 {
		return ErrItemNotOwned
	}
	if !isReforgeable(items[index]) {
		return ErrItemNotReforgeable
	}
	return nil
}

func (s *Service) SaveReforge(character characters.Summary, itemID string) (InventoryView, EquipmentItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	items := s.ensureCharacterItemsLocked(character)
	index := indexByID(items, strings.TrimSpace(itemID))
	if index < 0 {
		return InventoryView{}, EquipmentItem{}, ErrItemNotOwned
	}
	if items[index].PendingReforge == nil {
		return InventoryView{}, EquipmentItem{}, ErrReforgeNotPending
	}

	items[index].ExtraAffixes = cloneAffixes(items[index].PendingReforge.PreviewAffixes)
	items[index].PendingReforge = nil
	items[index] = recomputeItemStats(items[index])
	s.itemsByCharacter[character.CharacterID] = items
	view := buildView(character, items, s.ensureConsumablesForCharacterLocked(character), s.ensureSlotEnhancementsLocked(character.CharacterID))
	reforged := findItemByID(view.Equipped, view.Inventory, itemID)
	return view, reforged, nil
}

func (s *Service) DiscardReforge(character characters.Summary, itemID string) (InventoryView, EquipmentItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	items := s.ensureCharacterItemsLocked(character)
	index := indexByID(items, strings.TrimSpace(itemID))
	if index < 0 {
		return InventoryView{}, EquipmentItem{}, ErrItemNotOwned
	}
	if items[index].PendingReforge == nil {
		return InventoryView{}, EquipmentItem{}, ErrReforgeNotPending
	}

	items[index].PendingReforge = nil
	s.itemsByCharacter[character.CharacterID] = items
	view := buildView(character, items, s.ensureConsumablesForCharacterLocked(character), s.ensureSlotEnhancementsLocked(character.CharacterID))
	reforged := findItemByID(view.Equipped, view.Inventory, itemID)
	return view, reforged, nil
}

func (s *Service) DeriveStats(character characters.Summary, base characters.StatsSnapshot) characters.StatsSnapshot {
	s.mu.Lock()
	defer s.mu.Unlock()

	items := s.ensureCharacterItemsLocked(character)
	return deriveStats(base, items, s.ensureSlotEnhancementsLocked(character.CharacterID))
}

func (s *Service) ListSlotEnhancements(character characters.Summary) []characters.SlotEnhancementView {
	s.mu.Lock()
	defer s.mu.Unlock()

	return buildSlotEnhancementViews(s.ensureSlotEnhancementsLocked(character.CharacterID))
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

func (s *Service) ensureSlotEnhancementsLocked(characterID string) map[string]int {
	if levels, ok := s.slotEnhancementsByChar[characterID]; ok {
		copied := make(map[string]int, len(levels))
		for key, value := range levels {
			copied[key] = value
		}
		return copied
	}
	s.slotEnhancementsByChar[characterID] = map[string]int{}
	return map[string]int{}
}

func (s *Service) resolveEnhancementTargetLocked(items []EquipmentItem, itemID, slot string) (string, EquipmentItem, error) {
	itemID = strings.TrimSpace(itemID)
	slot = strings.TrimSpace(slot)

	if itemID != "" {
		index := indexByID(items, itemID)
		if index < 0 {
			return "", EquipmentItem{}, ErrItemNotOwned
		}
		return items[index].Slot, items[index], nil
	}

	if slot == "" {
		return "", EquipmentItem{}, ErrItemNotOwned
	}

	for _, item := range items {
		if item.Slot == slot {
			return slot, item, nil
		}
	}

	return slot, EquipmentItem{Slot: slot, Rarity: "common", Stats: map[string]int{}}, nil
}

func applySlotEnhancement(item EquipmentItem, level int) EquipmentItem {
	item.EnhancementLevel = level
	item.EnhancementPreviewPct = enhancementPreviewPct(level)
	return item
}

func buildEquipmentItemFromCatalog(catalog catalogItem, state string) EquipmentItem {
	item := EquipmentItem{
		ItemID:              nextItemID(),
		CatalogID:           catalog.CatalogID,
		Name:                catalog.Name,
		Slot:                catalog.Slot,
		Rarity:              catalog.Rarity,
		RequiredClass:       catalog.RequiredClass,
		RequiredWeaponStyle: catalog.RequiredWeaponStyle,
		EnhancementLevel:    0,
		Durability:          100,
		BaseStats:           copyStats(catalog.Stats),
		PassiveAffix:        nil,
		ExtraAffixes:        nil,
		State:               state,
	}
	item.ExtraAffixes = rollExtraAffixes(item.ItemID, item.Slot, item.Rarity, item.CatalogID)
	return recomputeItemStats(item)
}

func recomputeItemStats(item EquipmentItem) EquipmentItem {
	if item.BaseStats == nil {
		item.BaseStats = copyStats(item.Stats)
	}
	item.Stats = copyStats(item.BaseStats)
	for _, affix := range item.ExtraAffixes {
		item.Stats[affix.AffixKey] += affix.Value
	}
	return item
}

func cloneAffixes(items []EquipmentAffix) []EquipmentAffix {
	cloned := make([]EquipmentAffix, len(items))
	copy(cloned, items)
	return cloned
}

func findItemByID(equipped []EquipmentItem, bag []EquipmentItem, itemID string) EquipmentItem {
	for _, item := range equipped {
		if item.ItemID == itemID {
			return item
		}
	}
	for _, item := range bag {
		if item.ItemID == itemID {
			return item
		}
	}
	return EquipmentItem{}
}

func isReforgeable(item EquipmentItem) bool {
	return len(item.ExtraAffixes) > 0
}

func rollExtraAffixes(itemID, slot, rarity, seed string) []EquipmentAffix {
	count := extraAffixCountForRarity(rarity)
	if count <= 0 {
		return nil
	}
	pool := affixPoolForSlot(slot)
	used := map[string]struct{}{}
	affixes := make([]EquipmentAffix, 0, count)
	for i := 0; i < count && len(used) < len(pool); i++ {
		key := pickAffixKey(pool, used, itemID, seed, i)
		used[key] = struct{}{}
		affixes = append(affixes, EquipmentAffix{
			AffixKey: key,
			Value:    rollAffixValue(key, rarity, itemID, seed, i),
		})
	}
	return affixes
}

func extraAffixCountForRarity(rarity string) int {
	switch strings.ToLower(strings.TrimSpace(rarity)) {
	case "blue", "rare":
		return 1
	case "purple", "epic":
		return 2
	case "gold":
		return 3
	case "red", "prismatic":
		return 4
	default:
		return 0
	}
}

func affixPoolForSlot(slot string) []string {
	switch strings.ToLower(strings.TrimSpace(slot)) {
	case "weapon":
		return []string{"physical_attack", "magic_attack", "healing_power", "precision", "crit_rate", "crit_damage", "speed"}
	case "head":
		return []string{"max_hp", "physical_defense", "magic_defense", "speed", "healing_power"}
	case "chest":
		return []string{"max_hp", "physical_defense", "magic_defense", "healing_power"}
	case "boots":
		return []string{"speed", "max_hp", "physical_defense", "magic_defense", "precision"}
	case "ring":
		return []string{"physical_attack", "magic_attack", "healing_power", "precision", "crit_rate", "crit_damage", "speed"}
	case "necklace":
		return []string{"max_hp", "healing_power", "magic_attack", "magic_defense", "speed", "precision"}
	default:
		return []string{"max_hp", "physical_attack", "magic_attack", "physical_defense", "magic_defense", "speed", "healing_power"}
	}
}

func pickAffixKey(pool []string, used map[string]struct{}, itemID, seed string, offset int) string {
	available := make([]string, 0, len(pool))
	for _, key := range pool {
		if _, exists := used[key]; !exists {
			available = append(available, key)
		}
	}
	if len(available) == 0 {
		return pool[0]
	}
	index := deterministicIndex(itemID, seed, offset, len(available))
	return available[index]
}

func rollAffixValue(key, rarity, itemID, seed string, offset int) int {
	minValue, maxValue := affixRange(key)
	if minValue >= maxValue {
		return maxValue
	}
	if strings.EqualFold(strings.TrimSpace(rarity), "prismatic") {
		return maxValue
	}
	minRoll, maxRoll := qualityRollBand(rarity)
	position := minRoll + deterministicFloat(itemID, seed, offset)*(maxRoll-minRoll)
	value := float64(minValue) + position*float64(maxValue-minValue)
	return int(math.Round(value))
}

func affixRange(key string) (int, int) {
	switch key {
	case "max_hp":
		return 10, 38
	case "physical_attack", "magic_attack", "healing_power":
		return 2, 8
	case "physical_defense", "magic_defense":
		return 1, 6
	case "speed":
		return 1, 3
	case "precision":
		return 2, 8
	case "crit_rate":
		return 1, 5
	case "crit_damage":
		return 3, 10
	default:
		return 1, 4
	}
}

func qualityRollBand(rarity string) (float64, float64) {
	switch strings.ToLower(strings.TrimSpace(rarity)) {
	case "blue", "rare":
		return 0.70, 0.82
	case "purple", "epic":
		return 0.78, 0.88
	case "gold":
		return 0.84, 0.93
	case "red":
		return 0.90, 0.97
	case "prismatic":
		return 1.0, 1.0
	default:
		return 0.70, 0.82
	}
}

func reforgeStoneCost(item EquipmentItem) int {
	switch strings.ToLower(strings.TrimSpace(item.Rarity)) {
	case "blue":
		return 1
	case "purple":
		return 2
	case "gold":
		return 3
	case "red":
		return 5
	case "prismatic":
		return 8
	default:
		return 3
	}
}

func deterministicIndex(itemID, seed string, offset, length int) int {
	if length <= 1 {
		return 0
	}
	return int(deterministicFloat(itemID, seed, offset)*float64(length)) % length
}

func deterministicFloat(itemID, seed string, offset int) float64 {
	hasher := fnv.New64a()
	_, _ = hasher.Write([]byte(fmt.Sprintf("%s|%s|%d", itemID, seed, offset)))
	return float64(hasher.Sum64()%10000) / 10000
}

func starterItemsFor(character characters.Summary) []EquipmentItem {
	if strings.EqualFold(character.Class, "civilian") {
		catalogIDs := []string{"starter_chest_cloth", "starter_boots"}
		items := make([]EquipmentItem, 0, len(catalogIDs))
		for _, id := range catalogIDs {
			catalog, ok := starterCatalog[id]
			if !ok {
				continue
			}
			items = append(items, buildEquipmentItemFromCatalog(catalog, "equipped"))
		}
		for _, id := range catalogIDs {
			catalog, ok := starterCatalog[id]
			if !ok {
				continue
			}
			items = append(items, buildEquipmentItemFromCatalog(catalog, "inventory"))
		}
		return items
	}

	weaponCatalogID := ProfessionStarterCatalogID(character.ProfessionRoute)

	chestCatalogID := "starter_chest_cloth"
	if character.Class == "warrior" {
		chestCatalogID = "starter_chest_armor"
	}

	catalogIDs := []string{weaponCatalogID, chestCatalogID, "starter_boots"}
	items := make([]EquipmentItem, 0, len(catalogIDs))
	for _, id := range catalogIDs {
		catalog, ok := starterCatalog[id]
		if !ok {
			continue
		}
		items = append(items, buildEquipmentItemFromCatalog(catalog, "inventory"))
	}

	for i := range items {
		items[i].State = "equipped"
	}
	if len(items) >= 3 {
		for _, id := range []string{chestCatalogID, "starter_boots"} {
			catalog, ok := starterCatalog[id]
			if !ok {
				continue
			}
			items = append(items, buildEquipmentItemFromCatalog(catalog, "inventory"))
		}
	}
	return items
}

func ProfessionStarterCatalogID(routeID string) string {
	switch strings.ToLower(strings.TrimSpace(routeID)) {
	case "tank", "magic_burst":
		return "warrior_sword_starter"
	case "physical_burst":
		return "warrior_axe_starter"
	case "aoe_burst":
		return "mage_staff_starter"
	case "single_burst", "control":
		return "mage_spellbook_starter"
	case "curse":
		return "priest_scepter_starter"
	case "healing_support", "summon":
		return "priest_tome_starter"
	default:
		return ""
	}
}

func buildView(character characters.Summary, items []EquipmentItem, consumables map[string]int, slotEnhancements map[string]int) InventoryView {
	equipped := make([]EquipmentItem, 0)
	bag := make([]EquipmentItem, 0)
	for _, item := range items {
		annotated := applySlotEnhancement(item, slotEnhancements[item.Slot])
		if item.State == "equipped" {
			equipped = append(equipped, annotated)
		} else {
			bag = append(bag, annotated)
		}
	}

	sort.Slice(equipped, func(i, j int) bool { return equipped[i].Slot < equipped[j].Slot })
	sort.Slice(bag, func(i, j int) bool { return bag[i].ItemID < bag[j].ItemID })

	score := 0
	for _, item := range equipped {
		score += rarityScore(item.Rarity)
		score += item.EnhancementLevel * 5
		for _, value := range item.Stats {
			score += value
		}
	}

	return InventoryView{
		EquipmentScore:       score,
		Equipped:             equipped,
		Inventory:            bag,
		Consumables:          buildConsumableStacks(consumables),
		SlotEnhancements:     buildSlotEnhancementViews(slotEnhancements),
		UpgradeHints:         buildUpgradeHints(character, equipped, bag),
		PotionLoadoutOptions: buildPotionLoadoutOptions(character, consumables),
	}
}

func buildSlotEnhancementViews(slotEnhancements map[string]int) []characters.SlotEnhancementView {
	slots := []string{"weapon", "head", "chest", "hands", "legs", "boots", "accessory"}
	items := make([]characters.SlotEnhancementView, 0, len(slots))
	for _, slot := range slots {
		level := slotEnhancements[slot]
		items = append(items, characters.SlotEnhancementView{
			Slot:                  slot,
			EnhancementLevel:      level,
			EnhancementPreviewPct: enhancementPreviewPct(level),
			MaxEnhancementLevel:   20,
		})
	}
	return items
}

func buildUpgradeHints(character characters.Summary, equipped []EquipmentItem, bag []EquipmentItem) []UpgradeHint {
	equippedBySlot := make(map[string]EquipmentItem, len(equipped))
	for _, item := range equipped {
		equippedBySlot[item.Slot] = item
	}

	hints := make([]UpgradeHint, 0, 6)
	bestInventoryBySlot := make(map[string]EquipmentItem)
	for _, item := range bag {
		current, ok := bestInventoryBySlot[item.Slot]
		if !ok || equipmentItemScore(item) > equipmentItemScore(current) {
			bestInventoryBySlot[item.Slot] = item
		}
	}
	for slot, candidate := range bestInventoryBySlot {
		delta := equipmentItemScore(candidate) - equipmentItemScore(equippedBySlot[slot])
		if delta <= 0 {
			continue
		}
		hints = append(hints, UpgradeHint{
			Source:            "inventory",
			ItemID:            candidate.ItemID,
			CatalogID:         candidate.CatalogID,
			Name:              candidate.Name,
			Slot:              slot,
			ScoreDelta:        delta,
			Affordable:        true,
			DirectlyEquipable: itemCompatible(character, candidate),
		})
	}

	bestShopBySlot := make(map[string]ShopItem)
	for _, item := range buildEquipmentShopItems(character) {
		current, ok := bestShopBySlot[item.Slot]
		if !ok || shopItemScore(item) > shopItemScore(current) {
			bestShopBySlot[item.Slot] = item
		}
	}
	for slot, candidate := range bestShopBySlot {
		delta := shopItemScore(candidate) - equipmentItemScore(equippedBySlot[slot])
		if delta <= 0 {
			continue
		}
		hints = append(hints, UpgradeHint{
			Source:            "shop",
			CatalogID:         candidate.CatalogID,
			Name:              candidate.Name,
			Slot:              slot,
			ScoreDelta:        delta,
			PriceGold:         candidate.PriceGold,
			Affordable:        character.Gold >= candidate.PriceGold,
			DirectlyEquipable: true,
		})
	}

	sort.Slice(hints, func(i, j int) bool {
		if hints[i].ScoreDelta != hints[j].ScoreDelta {
			return hints[i].ScoreDelta > hints[j].ScoreDelta
		}
		if hints[i].Affordable != hints[j].Affordable {
			return hints[i].Affordable
		}
		if hints[i].Source != hints[j].Source {
			return hints[i].Source < hints[j].Source
		}
		return hints[i].Slot < hints[j].Slot
	})
	if len(hints) > 8 {
		hints = hints[:8]
	}
	return hints
}

func buildPotionLoadoutOptions(character characters.Summary, consumables map[string]int) []PotionLoadoutOption {
	options := make([]PotionLoadoutOption, 0, len(consumableShopCatalog))
	for _, entry := range consumableShopCatalog {
		if !rankAtLeast(character.Rank, entry.MinRank) {
			continue
		}
		quantity := consumables[entry.CatalogID]
		options = append(options, PotionLoadoutOption{
			CatalogID:     entry.CatalogID,
			Name:          entry.Name,
			Family:        entry.Family,
			Tier:          entry.Tier,
			QuantityOwned: quantity,
			PriceGold:     entry.PriceGold,
			AvailableNow:  quantity > 0,
			CanPurchase:   character.Gold >= entry.PriceGold,
			Recommended:   potionRecommendedForClass(character.Class, entry.Family),
			EffectSummary: entry.EffectSummary,
		})
	}
	sort.Slice(options, func(i, j int) bool {
		if options[i].Recommended != options[j].Recommended {
			return options[i].Recommended
		}
		if options[i].AvailableNow != options[j].AvailableNow {
			return options[i].AvailableNow
		}
		if options[i].Tier != options[j].Tier {
			return options[i].Tier < options[j].Tier
		}
		return options[i].CatalogID < options[j].CatalogID
	})
	return options
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
	case "prismatic":
		return 50
	case "red":
		return 40
	case "gold":
		return 35
	case "epic":
		return 30
	case "rare":
		return 20
	default:
		return 10
	}
}

func equipmentItemScore(item EquipmentItem) int {
	score := rarityScore(item.Rarity) + item.EnhancementLevel*5
	for _, value := range item.Stats {
		score += value
	}
	return score
}

func shopItemScore(item ShopItem) int {
	score := rarityScore(item.Rarity)
	for _, value := range item.Stats {
		score += value
	}
	return score
}

func buildEquipmentShopItems(character characters.Summary) []ShopItem {
	items := make([]ShopItem, 0, len(shopCatalog))
	for _, entry := range shopCatalog {
		if len(entry.BuildingTypes) > 0 && !containsString(entry.BuildingTypes, "equipment_shop") {
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
	return items
}

func deriveStats(base characters.StatsSnapshot, items []EquipmentItem, slotEnhancements map[string]int) characters.StatsSnapshot {
	derived := base
	for _, item := range items {
		if item.State != "equipped" {
			continue
		}
		applyEquipmentStatMap(&derived, scaledBaseStats(item, slotEnhancements[item.Slot]))
		applyPassiveAffix(&derived, item.PassiveAffix)
	}
	return derived
}

func scaledBaseStats(item EquipmentItem, enhancementLevel int) map[string]float64 {
	multiplier := 1 + enhancementPreviewPct(enhancementLevel)
	values := make(map[string]float64, len(item.Stats))
	for key, value := range item.Stats {
		values[key] = float64(value) * multiplier
	}
	return values
}

func applyPassiveAffix(stats *characters.StatsSnapshot, affix map[string]any) {
	if affix == nil {
		return
	}
	values := make(map[string]float64, len(affix))
	for key, raw := range affix {
		switch typed := raw.(type) {
		case int:
			values[key] = float64(typed)
		case int32:
			values[key] = float64(typed)
		case int64:
			values[key] = float64(typed)
		case float64:
			values[key] = typed
		}
	}
	applyEquipmentStatMap(stats, values)
}

func applyEquipmentStatMap(stats *characters.StatsSnapshot, values map[string]float64) {
	for key, value := range values {
		switch key {
		case "max_hp":
			stats.MaxHP += int(math.Round(value))
		case "physical_attack":
			stats.PhysicalAttack += int(math.Round(value))
		case "magic_attack":
			stats.MagicAttack += int(math.Round(value))
		case "physical_defense":
			stats.PhysicalDefense += int(math.Round(value))
		case "magic_defense":
			stats.MagicDefense += int(math.Round(value))
		case "speed":
			stats.Speed += int(math.Round(value))
		case "healing_power":
			stats.HealingPower += int(math.Round(value))
		case "crit_rate":
			stats.CritRate += value / 100
		case "crit_damage":
			stats.CritDamage += value / 100
		case "block_rate":
			stats.BlockRate += value / 100
		case "precision":
			stats.Precision += value / 100
		case "evasion_rate":
			stats.EvasionRate += value / 100
		case "physical_mastery":
			stats.PhysicalMastery += value / 100
		case "magic_mastery":
			stats.MagicMastery += value / 100
		}
	}
}

func enhancementPreviewPct(level int) float64 {
	if level <= 0 {
		return 0
	}
	return float64(level) * 0.01
}

func isEnhanceable(item EquipmentItem) bool {
	return maxEnhancementLevelFor(item) > 0
}

func maxEnhancementLevelFor(_ EquipmentItem) int {
	return 20
}

func enhancementGoldCost(item EquipmentItem, level int) int {
	multiplier := rarityEnhancementMultiplier(item.Rarity)
	return int(math.Round((35 + float64(level*15)) * multiplier))
}

func enhancementMaterialCost(item EquipmentItem, level int) []map[string]any {
	multiplier := rarityEnhancementMultiplier(item.Rarity)
	qty := int(math.Ceil((2 + float64(level)) * multiplier))
	return []map[string]any{{
		"material_key": EnhancementMaterialKey,
		"quantity":     qty,
	}}
}

func salvageYieldFor(item EquipmentItem) int {
	base := map[string]int{
		"common":    1,
		"rare":      3,
		"epic":      7,
		"gold":      14,
		"red":       24,
		"prismatic": 40,
	}[strings.ToLower(strings.TrimSpace(item.Rarity))]
	if base == 0 {
		base = 1
	}
	return base
}

func rarityEnhancementMultiplier(rarity string) float64 {
	switch strings.ToLower(strings.TrimSpace(rarity)) {
	case "rare":
		return 1.3
	case "epic":
		return 1.7
	case "gold":
		return 2.2
	case "red":
		return 2.8
	case "prismatic":
		return 3.5
	default:
		return 1.0
	}
}

func potionRecommendedForClass(class, family string) bool {
	family = strings.ToLower(strings.TrimSpace(family))
	if family == "hp" {
		return true
	}
	switch strings.ToLower(strings.TrimSpace(class)) {
	case "warrior":
		return family == "def"
	case "mage":
		return family == "atk" || family == "spd"
	case "priest":
		return family == "def" || family == "atk"
	default:
		return false
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
