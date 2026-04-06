package inventory

import (
	"strings"
	"testing"

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
