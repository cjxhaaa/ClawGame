package inventory

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"clawgame/apps/api/internal/characters"
)

func TestDungeonRewardCatalogUsesOnlyLiveSetFamilies(t *testing.T) {
	allowedPrefixes := []string{
		"gravewake_bastion_",
		"briarbound_sight_",
		"sunscar_assault_",
		"nightglass_arcanum_",
	}

	for catalogID, item := range dungeonRewardCatalog {
		if item.Slot == "wrist" {
			t.Fatalf("expected dungeon reward catalog %s not to use retired wrist slot", catalogID)
		}

		matched := false
		for _, prefix := range allowedPrefixes {
			if strings.HasPrefix(catalogID, prefix) {
				matched = true
				break
			}
		}
		if !matched {
			t.Fatalf("unexpected legacy dungeon reward catalog id %s", catalogID)
		}
	}
}

func TestDungeonRewardCatalogIncludesAllWeaponStylesPerSet(t *testing.T) {
	expectedStyles := []string{"sword_shield", "great_axe", "staff", "spellbook", "scepter", "holy_tome"}
	for _, prefix := range []string{
		"gravewake_bastion",
		"briarbound_sight",
		"sunscar_assault",
		"nightglass_arcanum",
	} {
		for _, style := range expectedStyles {
			catalogID := prefix + "_weapon_" + style + "_red"
			item, ok := dungeonRewardCatalog[catalogID]
			if !ok {
				t.Fatalf("expected live dungeon reward catalog %s", catalogID)
			}
			if item.Slot != "weapon" {
				t.Fatalf("expected %s to be a weapon, got %s", catalogID, item.Slot)
			}
			if item.RequiredWeaponStyle != style {
				t.Fatalf("expected %s to require style %s, got %s", catalogID, style, item.RequiredWeaponStyle)
			}
		}
	}
}

func TestDeriveStatsAppliesEquippedSetBonuses(t *testing.T) {
	service := NewService()
	character := characters.Summary{
		CharacterID: "char_set_bonus",
		Name:        "Set Bonus",
		Class:       "warrior",
		WeaponStyle: "sword_shield",
	}
	service.itemsByCharacter[character.CharacterID] = []EquipmentItem{
		buildEquipmentItemFromCatalog(dungeonRewardCatalog["gravewake_bastion_weapon_sword_shield_red"], "equipped"),
		buildEquipmentItemFromCatalog(dungeonRewardCatalog["gravewake_bastion_chest_red"], "equipped"),
	}

	base := characters.StatsSnapshot{
		MaxHP:           1000,
		PhysicalDefense: 100,
		MagicDefense:    100,
	}
	derived := service.DeriveStats(character, base)

	if derived.MaxHP <= base.MaxHP {
		t.Fatalf("expected set bonus to increase max HP, base=%d derived=%d", base.MaxHP, derived.MaxHP)
	}
	if derived.PhysicalDefense <= base.PhysicalDefense {
		t.Fatalf("expected set bonus to increase physical defense, base=%d derived=%d", base.PhysicalDefense, derived.PhysicalDefense)
	}

	view := service.GetInventory(character)
	if len(view.EquippedSetBonuses) != 1 {
		t.Fatalf("expected one equipped set bonus view, got %d", len(view.EquippedSetBonuses))
	}
	if view.EquippedSetBonuses[0].SetID != "gravewake_bastion" {
		t.Fatalf("expected gravewake_bastion set view, got %s", view.EquippedSetBonuses[0].SetID)
	}
	if len(view.EquippedSetBonuses[0].ActiveTiers) != 1 || view.EquippedSetBonuses[0].ActiveTiers[0] != 2 {
		t.Fatalf("expected 2-piece tier to be active, got %#v", view.EquippedSetBonuses[0].ActiveTiers)
	}
}

func TestCivilianCanEquipAnyWeaponStyle(t *testing.T) {
	service := NewService()
	character := characters.Summary{
		CharacterID: "char_civilian_any_weapon",
		Name:        "Flexible Civilian",
		Class:       "civilian",
		WeaponStyle: "",
	}

	_, item, err := service.GrantItemFromCatalog(character, "gravewake_bastion_weapon_staff_red")
	if err != nil {
		t.Fatalf("expected weapon grant to succeed, got %v", err)
	}

	view, err := service.EquipItem(character, item.ItemID)
	if err != nil {
		t.Fatalf("expected civilian to equip any weapon style, got %v", err)
	}
	if len(view.Equipped) == 0 {
		t.Fatal("expected equipped items after civilian equip")
	}

	foundWeapon := false
	for _, equipped := range view.Equipped {
		if equipped.ItemID == item.ItemID && equipped.Slot == "weapon" {
			foundWeapon = true
			break
		}
	}
	if !foundWeapon {
		t.Fatal("expected granted weapon to be equipped for civilian")
	}
}

func TestPromotedCharacterStillRespectsWeaponStyleRestriction(t *testing.T) {
	service := NewService()
	character := characters.Summary{
		CharacterID: "char_warrior_weapon_lock",
		Name:        "Focused Warrior",
		Class:       "warrior",
		WeaponStyle: "sword_shield",
	}

	_, item, err := service.GrantItemFromCatalog(character, "gravewake_bastion_weapon_staff_red")
	if err != nil {
		t.Fatalf("expected weapon grant to succeed, got %v", err)
	}

	if _, err := service.EquipItem(character, item.ItemID); err != ErrItemNotEquippable {
		t.Fatalf("expected non-civilian weapon-style mismatch to fail with ErrItemNotEquippable, got %v", err)
	}
}

func TestEquipmentShopInventoryIsPersonalizedAndDungeonBound(t *testing.T) {
	service := NewService()
	fixedNow := time.Date(2026, 4, 8, 10, 0, 0, 0, time.Local)
	service.clock = func() time.Time { return fixedNow }

	civilian := characters.Summary{CharacterID: "char_civilian_shop_a", Name: "Shop Civilian A", Class: "civilian"}
	other := characters.Summary{CharacterID: "char_civilian_shop_b", Name: "Shop Civilian B", Class: "civilian"}

	itemsA := service.ListShopInventory("equipment_shop", civilian)
	itemsARepeat := service.ListShopInventory("equipment_shop", civilian)
	itemsB := service.ListShopInventory("equipment_shop", other)

	if len(itemsA) != 6 {
		t.Fatalf("expected 6 daily equipment shop items, got %d", len(itemsA))
	}
	if !reflect.DeepEqual(itemsA, itemsARepeat) {
		t.Fatalf("expected same-day shop inventory to stay stable, got %#v vs %#v", itemsA, itemsARepeat)
	}
	if reflect.DeepEqual(itemsA, itemsB) {
		t.Fatalf("expected different characters to see different shop inventories, got %#v", itemsA)
	}

	seen := map[string]bool{}
	for _, item := range itemsA {
		if _, ok := dungeonRewardCatalog[item.CatalogID]; !ok {
			t.Fatalf("expected shop item %s to come from dungeon reward catalog", item.CatalogID)
		}
		if seen[item.CatalogID] {
			t.Fatalf("expected no duplicate catalog ids on one daily shop board, got %s", item.CatalogID)
		}
		seen[item.CatalogID] = true
		if item.PriceGold < 200 {
			t.Fatalf("expected equipment shop prices to stay expensive, got %d for %s", item.PriceGold, item.CatalogID)
		}
	}
}

func TestEquipmentShopPurchaseConsumesDailyOffer(t *testing.T) {
	service := NewService()
	service.clock = func() time.Time { return time.Date(2026, 4, 8, 10, 0, 0, 0, time.Local) }
	character := characters.Summary{CharacterID: "char_daily_shop_purchase", Name: "Daily Buyer", Class: "civilian"}

	items := service.ListShopInventory("equipment_shop", character)
	if len(items) == 0 {
		t.Fatal("expected non-empty daily equipment shop inventory")
	}

	before := len(items)
	view, purchased, price, err := service.PurchaseItem(character, items[0].CatalogID)
	if err != nil {
		t.Fatalf("expected purchase to succeed, got %v", err)
	}
	if price != items[0].PriceGold {
		t.Fatalf("expected purchase price %d, got %d", items[0].PriceGold, price)
	}
	if purchased.CatalogID != items[0].CatalogID {
		t.Fatalf("expected purchased catalog %s, got %s", items[0].CatalogID, purchased.CatalogID)
	}
	remaining := service.ListShopInventory("equipment_shop", character)
	if len(remaining) != before-1 {
		t.Fatalf("expected purchased offer to disappear, got %d items after buying from %d", len(remaining), before)
	}
	for _, item := range remaining {
		if item.CatalogID == purchased.CatalogID {
			t.Fatalf("expected purchased catalog %s to be removed from daily shop", purchased.CatalogID)
		}
	}
	found := false
	for _, item := range view.Inventory {
		if item.CatalogID == purchased.CatalogID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected purchased item %s to be added to inventory", purchased.CatalogID)
	}
}

func TestBuildSlotEnhancementViewsUsesCanonicalSixEquipmentSlots(t *testing.T) {
	views := buildSlotEnhancementViews(map[string]int{
		"weapon":   3,
		"necklace": 2,
		"ring":     1,
	})

	expected := []string{"head", "chest", "necklace", "ring", "boots", "weapon"}
	if len(views) != len(expected) {
		t.Fatalf("expected %d slot enhancement views, got %d", len(expected), len(views))
	}

	for i, slot := range expected {
		if views[i].Slot != slot {
			t.Fatalf("expected slot %d to be %s, got %s", i, slot, views[i].Slot)
		}
	}

	if views[2].EnhancementLevel != 2 {
		t.Fatalf("expected necklace enhancement level 2, got %d", views[2].EnhancementLevel)
	}
	if views[3].EnhancementLevel != 1 {
		t.Fatalf("expected ring enhancement level 1, got %d", views[3].EnhancementLevel)
	}
	if views[5].EnhancementLevel != 3 {
		t.Fatalf("expected weapon enhancement level 3, got %d", views[5].EnhancementLevel)
	}
}

func TestSellPriceUsesCanonicalEquipmentShopEstimate(t *testing.T) {
	item := buildEquipmentItemFromCatalog(dungeonRewardCatalog["gravewake_bastion_chest_red"], "inventory")

	price := sellPriceFor(item)

	expected := equipmentShopEstimatedPrice(dungeonRewardCatalog["gravewake_bastion_chest_red"], 1.0) / 2
	if price != expected {
		t.Fatalf("expected sell price %d from canonical equipment estimate, got %d", expected, price)
	}
}

func TestBluePurpleRarityDriveEnhancementAndSalvage(t *testing.T) {
	blue := EquipmentItem{Rarity: "blue"}
	purple := EquipmentItem{Rarity: "purple"}

	if got := rarityEnhancementMultiplier(blue.Rarity); got != 1.3 {
		t.Fatalf("expected blue enhancement multiplier 1.3, got %v", got)
	}
	if got := rarityEnhancementMultiplier(purple.Rarity); got != 1.7 {
		t.Fatalf("expected purple enhancement multiplier 1.7, got %v", got)
	}
	if got := salvageYieldFor(blue); got != 3 {
		t.Fatalf("expected blue salvage yield 3, got %d", got)
	}
	if got := salvageYieldFor(purple); got != 7 {
		t.Fatalf("expected purple salvage yield 7, got %d", got)
	}
}

func TestUnequipWeaponIfIncompatibleMovesWeaponBackToInventory(t *testing.T) {
	service := NewService()
	character := characters.Summary{
		CharacterID: "char_swap_weapon",
		Name:        "Swap Weapon",
		Class:       "mage",
		WeaponStyle: "spellbook",
	}
	service.itemsByCharacter[character.CharacterID] = []EquipmentItem{
		buildEquipmentItemFromCatalog(starterCatalog["warrior_sword_starter"], "equipped"),
	}

	changed, err := service.UnequipWeaponIfIncompatible(character)
	if err != nil {
		t.Fatalf("expected incompatible weapon check to succeed, got %v", err)
	}
	if !changed {
		t.Fatal("expected incompatible weapon to be unequipped")
	}

	view := service.GetInventory(character)
	if len(view.Equipped) != 0 {
		t.Fatalf("expected no equipped weapon after auto-unequip, got %#v", view.Equipped)
	}
	if len(view.Inventory) != 1 || view.Inventory[0].State != "inventory" {
		t.Fatalf("expected weapon moved to inventory, got %#v", view.Inventory)
	}
}
