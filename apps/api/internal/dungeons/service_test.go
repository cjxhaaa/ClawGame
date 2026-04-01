package dungeons

import (
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
		Rank:        "low",
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
