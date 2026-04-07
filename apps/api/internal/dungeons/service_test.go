package dungeons

import (
	"strings"
	"testing"

	"clawgame/apps/api/internal/characters"
	"clawgame/apps/api/internal/combat"
)

func TestEnterDungeonUsesProvidedPlayerStats(t *testing.T) {
	service := NewService()
	character := characters.Summary{
		CharacterID: "char_test",
		Name:        "DungeonPrep",
		Class:       "warrior",
	}
	player := combat.Combatant{
		EntityID:        character.CharacterID,
		Name:            character.Name,
		Team:            "a",
		IsPlayerSide:    true,
		Class:           character.Class,
		MaxHP:           777,
		CurrentHP:       777,
		PhysAtk:         180,
		MagAtk:          10,
		PhysDef:         80,
		MagDef:          50,
		Speed:           30,
		HealPow:         20,
		CritRate:        0.20,
		CritDamage:      0.50,
		BlockRate:       0.05,
		PhysicalMastery: 0.25,
	}

	run, err := service.EnterDungeon(character, characters.DailyLimits{}, player, "ancient_catacomb_v1", "normal", nil, nil)
	if err != nil {
		t.Fatalf("expected enter dungeon to succeed, got %v", err)
	}

	startHP, _ := run.BattleState["start_hp"].(int)
	if startHP != 777 {
		t.Fatalf("expected dungeon battle_state.start_hp to use provided player stats 777, got %#v", run.BattleState["start_hp"])
	}
	if run.RunStatus == "" {
		t.Fatal("expected dungeon run status to be populated")
	}
}

func TestListDungeonDefinitionsIncludesFourSeasonDungeons(t *testing.T) {
	service := NewService()
	definitions := service.ListDungeonDefinitions()
	if len(definitions) != 4 {
		t.Fatalf("expected 4 active dungeon definitions, got %d", len(definitions))
	}

	expected := map[string]struct{}{
		"ancient_catacomb_v1": {},
		"thorned_hollow_v1":   {},
		"sunscar_warvault_v1": {},
		"obsidian_spire_v1":   {},
	}
	for _, definition := range definitions {
		if definition.RoomCount != 6 || definition.BossRoomIndex != 6 {
			t.Fatalf("expected %s to use 6 rooms with boss in room 6, got rooms=%d boss=%d", definition.DungeonID, definition.RoomCount, definition.BossRoomIndex)
		}
		delete(expected, definition.DungeonID)
	}
	if len(expected) != 0 {
		t.Fatalf("missing active dungeon definitions: %#v", expected)
	}
}

func TestRunSummaryTagMarksBossFailureCorrectly(t *testing.T) {
	run := RunView{
		RunStatus:          "failed",
		HighestRoomCleared: 5,
		CurrentRoomIndex:   6,
		RoomSummary:        map[string]any{"boss_room_index": 6},
		BattleState:        map[string]any{"start_hp": 500, "remaining_hp": 0},
	}

	if got := runSummaryTag(run); got != "failed_at_boss" {
		t.Fatalf("expected failed_at_boss, got %q", got)
	}
}

func TestPreviewAndFinalizeRunRewards(t *testing.T) {
	service := NewService()
	character := characters.Summary{
		CharacterID: "char_claim",
		Name:        "Claimer",
		Class:       "warrior",
	}
	player := combat.Combatant{
		EntityID:        character.CharacterID,
		Name:            character.Name,
		Team:            "a",
		IsPlayerSide:    true,
		Class:           character.Class,
		MaxHP:           900,
		CurrentHP:       900,
		PhysAtk:         260,
		MagAtk:          20,
		PhysDef:         100,
		MagDef:          80,
		Speed:           35,
		HealPow:         20,
		CritRate:        0.20,
		CritDamage:      0.60,
		PhysicalMastery: 0.30,
	}

	run, err := service.EnterDungeon(character, characters.DailyLimits{DungeonEntryCap: 6}, player, "ancient_catacomb_v1", "easy", nil, nil)
	if err != nil {
		t.Fatalf("expected enter dungeon to succeed, got %v", err)
	}
	if !run.RewardClaimable {
		t.Fatal("expected run rewards to be claimable before finalize")
	}

	previewRun, claimPackage, err := service.PreviewRunRewards(character.CharacterID, run.RunID, characters.DailyLimits{DungeonEntryCap: 6})
	if err != nil {
		t.Fatalf("expected preview rewards to succeed, got %v", err)
	}
	if !previewRun.RewardClaimable {
		t.Fatal("expected preview not to settle reward claim")
	}
	if claimPackage.RewardGold <= 0 {
		t.Fatal("expected preview package to include gold")
	}

	finalRun, err := service.FinalizeRunRewards(character.CharacterID, run.RunID)
	if err != nil {
		t.Fatalf("expected finalize rewards to succeed, got %v", err)
	}
	if finalRun.RewardClaimable {
		t.Fatal("expected finalized run to be non-claimable")
	}
}

func TestScaledRewardGoldUsesConfiguredRange(t *testing.T) {
	low := scaledRewardGold(100, 200, 6, 6, "easy", "seed_a")
	high := scaledRewardGold(100, 200, 6, 6, "easy", "seed_b")
	if low < 100 || low > 200 {
		t.Fatalf("expected gold within configured range, got %d", low)
	}
	if high < 100 || high > 200 {
		t.Fatalf("expected gold within configured range, got %d", high)
	}
	if low == high {
		t.Fatalf("expected different seeds to produce different values, both were %d", low)
	}
}

func TestDungeonPoolsUseDistinctSetCatalogs(t *testing.T) {
	expectedPrefixes := map[string]string{
		"ancient_catacomb_v1": "gravewake_bastion_",
		"thorned_hollow_v1":   "briarbound_sight_",
		"sunscar_warvault_v1": "sunscar_assault_",
		"obsidian_spire_v1":   "nightglass_arcanum_",
	}

	for dungeonID, qualityPools := range dungeonQualityCatalogPools {
		prefix, ok := expectedPrefixes[dungeonID]
		if !ok {
			continue
		}
		for quality, pool := range qualityPools {
			if len(pool) == 0 {
				t.Fatalf("expected %s %s pool to be populated", dungeonID, quality)
			}
			for _, catalogID := range pool {
				if len(catalogID) < len(prefix) || catalogID[:len(prefix)] != prefix {
					t.Fatalf("expected %s %s pool item %s to use prefix %s", dungeonID, quality, catalogID, prefix)
				}
			}
		}
	}
}

func TestDungeonPoolsIncludeAllWeaponStyles(t *testing.T) {
	expectedStyles := []string{"sword_shield", "great_axe", "staff", "spellbook", "scepter", "holy_tome"}
	active := map[string]bool{
		"ancient_catacomb_v1": true,
		"thorned_hollow_v1":   true,
		"sunscar_warvault_v1": true,
		"obsidian_spire_v1":   true,
	}
	for dungeonID, qualityPools := range dungeonQualityCatalogPools {
		if !active[dungeonID] {
			continue
		}
		for quality, pool := range qualityPools {
			found := map[string]bool{}
			for _, catalogID := range pool {
				for _, style := range expectedStyles {
					if strings.Contains(catalogID, "_weapon_"+style+"_"+quality) {
						found[style] = true
					}
				}
			}
			for _, style := range expectedStyles {
				if !found[style] {
					t.Fatalf("expected %s %s pool to include weapon style %s", dungeonID, quality, style)
				}
			}
		}
	}
}

func TestSkillActionMetadataMatchesDocumentedCooldowns(t *testing.T) {
	tests := []struct {
		skillID  string
		tier     string
		cooldown int
	}{
		{skillID: "Guard Stance", tier: "advanced", cooldown: 2},
		{skillID: "War Cry", tier: "ultimate", cooldown: 3},
		{skillID: "Intercept", tier: "normal", cooldown: 1},
		{skillID: "Bulwark Field", tier: "ultimate", cooldown: 3},
		{skillID: "Linebreaker", tier: "advanced", cooldown: 2},
		{skillID: "Execution Rush", tier: "ultimate", cooldown: 3},
		{skillID: "Arc Veil", tier: "advanced", cooldown: 2},
		{skillID: "Detonate Sigil", tier: "ultimate", cooldown: 3},
		{skillID: "Flame Burst", tier: "advanced", cooldown: 2},
		{skillID: "Sanctuary Mark", tier: "advanced", cooldown: 2},
		{skillID: "Purge", tier: "normal", cooldown: 1},
		{skillID: "Grace Field", tier: "advanced", cooldown: 2},
		{skillID: "Purifying Wave", tier: "ultimate", cooldown: 3},
		{skillID: "Prayer of Renewal", tier: "ultimate", cooldown: 3},
		{skillID: "Judged Weakness", tier: "advanced", cooldown: 2},
		{skillID: "Wither Prayer", tier: "ultimate", cooldown: 3},
	}

	for _, tt := range tests {
		action := skillActionForID(tt.skillID, 1)
		if action.Tier != tt.tier || action.CooldownRounds != tt.cooldown {
			t.Fatalf("expected %s tier/cooldown %s/%d, got %s/%d", tt.skillID, tt.tier, tt.cooldown, action.Tier, action.CooldownRounds)
		}
	}
}
