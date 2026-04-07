package e2e

import "testing"

func TestFirstDayCivilianLoop(t *testing.T) {
	harness := NewHarness(t)
	client := registerCivilianCharacter(t, harness, "firstday-bot", "FirstDayRunner")

	initialState, err := client.State()
	if err != nil {
		t.Fatalf("load initial state: %v", err)
	}
	if initialState.Character.Class != "civilian" {
		t.Fatalf("expected first-day character to stay civilian, got %s", initialState.Character.Class)
	}
	if len(initialState.Objectives) != 4 {
		t.Fatalf("expected state to expose 4 daily objectives, got %d", len(initialState.Objectives))
	}
	logAccountSnapshot(t, client, "single-account initial")
	for _, quest := range initialState.Objectives {
		switch quest.TemplateType {
		case "clear_dungeon", "kill_region_enemies", "collect_materials":
			t.Fatalf("expected first-day civilian board to stay non-combat, got quest %#v", quest)
		}
	}

	if err := completeFirstDayContractBoard(client); err != nil {
		t.Fatalf("run first-day contract loop: %v", err)
	}

	postQuestState, err := client.State()
	if err != nil {
		t.Fatalf("load post-contract state: %v", err)
	}
	if postQuestState.Limits.QuestCompletionUsed != 4 {
		t.Fatalf("expected 4 quest submissions used, got %d", postQuestState.Limits.QuestCompletionUsed)
	}
	if postQuestState.Character.Gold <= initialState.Character.Gold {
		t.Fatalf("expected contracts to increase gold, initial=%d final=%d", initialState.Character.Gold, postQuestState.Character.Gold)
	}
	if postQuestState.Character.Reputation <= initialState.Character.Reputation {
		t.Fatalf("expected contracts to increase reputation, initial=%d final=%d", initialState.Character.Reputation, postQuestState.Character.Reputation)
	}
	logAccountSnapshot(t, client, "single-account after contracts")

	if err := buyStarterCombatPrep(client); err != nil {
		t.Fatalf("buy first-day combat prep: %v", err)
	}
	postPrepState, err := client.State()
	if err != nil {
		t.Fatalf("load post-prep state: %v", err)
	}
	logAccountSnapshot(t, client, "single-account after starter prep")

	preDungeonPower := postPrepState.CombatPower.PanelPowerScore
	preDungeonInventory, err := client.Inventory()
	if err != nil {
		t.Fatalf("load inventory before dungeon: %v", err)
	}
	equippedUpgrade, dungeonRuns, err := clearAndClaimNoviceDungeon(client, 2)
	if err != nil {
		t.Fatalf("run novice equipment dungeon: %v", err)
	}
	if !equippedUpgrade {
		t.Fatal("expected novice equipment dungeon to produce at least one directly equipable upgrade")
	}

	finalState, err := client.State()
	if err != nil {
		t.Fatalf("load final state: %v", err)
	}
	if finalState.CombatPower.PanelPowerScore <= preDungeonPower {
		t.Fatalf("expected combat power to increase after novice dungeon gearing, before=%d after=%d", preDungeonPower, finalState.CombatPower.PanelPowerScore)
	}
	if finalState.Limits.DungeonEntryUsed < 1 {
		t.Fatalf("expected at least one dungeon reward claim used, got %d", finalState.Limits.DungeonEntryUsed)
	}
	postDungeonInventory, err := client.Inventory()
	if err != nil {
		t.Fatalf("load inventory after dungeon: %v", err)
	}
	logDungeonRunSummary(t, "single-account dungeon runs", dungeonRuns)
	logDungeonLootSummary(t, "single-account dungeon loot", preDungeonInventory, postDungeonInventory)
	logAccountSnapshot(t, client, "single-account after dungeon")
}
