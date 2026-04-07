package e2e

import (
	"fmt"
	"testing"
)

func TestWorldBossSixAccountRaid(t *testing.T) {
	harness := NewHarness(t)
	clients := provisionWorldBossParty(t, harness, "worldboss")

	currentBoss, err := clients[0].WorldBossCurrent()
	if err != nil {
		t.Fatalf("load current world boss: %v", err)
	}
	if currentBoss.RequiredPartySize != 6 {
		t.Fatalf("expected world boss required party size 6, got %d", currentBoss.RequiredPartySize)
	}

	initialStates := make([]State, len(clients))
	for index, client := range clients {
		state, err := client.State()
		if err != nil {
			t.Fatalf("load initial state for client %d: %v", index, err)
		}
		initialStates[index] = state
		logAccountSnapshot(t, client, fmt.Sprintf("combat-boss client-%d before boss", index))
	}

	var resolvedRaid *WorldBossRaid
	for index, client := range clients {
		joinResult, err := client.JoinWorldBossQueue()
		if err != nil {
			t.Fatalf("join world boss queue for client %d: %v", index, err)
		}

		if index < len(clients)-1 {
			if joinResult.ResolvedRaid != nil {
				t.Fatalf("expected no raid resolution before sixth join, got raid %s on join %d", joinResult.ResolvedRaid.RaidID, index+1)
			}
			if joinResult.Status.CurrentQueuedCount != index+1 {
				t.Fatalf("expected queued count %d after join %d, got %d", index+1, index+1, joinResult.Status.CurrentQueuedCount)
			}
			continue
		}

		if joinResult.ResolvedRaid == nil {
			t.Fatal("expected sixth queue join to resolve a world boss raid")
		}
		resolvedRaid = joinResult.ResolvedRaid
	}

	if resolvedRaid == nil {
		t.Fatal("expected resolved raid")
	}
	logWorldBossRaidSummary(t, resolvedRaid, "combat-boss raid")
	if resolvedRaid.RaidID == "" {
		t.Fatal("expected resolved raid id")
	}
	if resolvedRaid.RewardTier == "" {
		t.Fatal("expected resolved raid reward tier")
	}
	if resolvedRaid.TotalDamage <= 0 {
		t.Fatalf("expected positive total damage, got %d", resolvedRaid.TotalDamage)
	}
	if len(resolvedRaid.Members) != 6 {
		t.Fatalf("expected 6 raid members, got %d", len(resolvedRaid.Members))
	}
	if resolvedRaid.RewardPackage.RewardGold <= 0 {
		t.Fatalf("expected positive gold reward package, got %d", resolvedRaid.RewardPackage.RewardGold)
	}

	memberByCharacter := make(map[string]WorldBossRaidMember, len(resolvedRaid.Members))
	for _, member := range resolvedRaid.Members {
		memberByCharacter[member.CharacterID] = member
		if member.DamageDealt <= 0 {
			t.Fatalf("expected member %s to deal positive damage, got %d", member.CharacterID, member.DamageDealt)
		}
	}

	for index, client := range clients {
		stateAfter, err := client.State()
		if err != nil {
			t.Fatalf("load post-raid state for client %d: %v", index, err)
		}
		if stateAfter.Character.Gold <= initialStates[index].Character.Gold {
			t.Fatalf("expected gold to increase for client %d, before=%d after=%d", index, initialStates[index].Character.Gold, stateAfter.Character.Gold)
		}
		if materialQuantity(stateAfter.Materials, "reforge_stone") <= materialQuantity(initialStates[index].Materials, "reforge_stone") {
			t.Fatalf("expected reforge stones to increase for client %d", index)
		}

		status, err := client.WorldBossQueueStatus()
		if err != nil {
			t.Fatalf("load queue status for client %d: %v", index, err)
		}
		if status.Queued {
			t.Fatalf("expected client %d to leave queue after raid resolution", index)
		}
		if status.LastRaidID != resolvedRaid.RaidID {
			t.Fatalf("expected client %d last raid id %s, got %s", index, resolvedRaid.RaidID, status.LastRaidID)
		}
		if status.LastRewardTier != resolvedRaid.RewardTier {
			t.Fatalf("expected client %d reward tier %s, got %s", index, resolvedRaid.RewardTier, status.LastRewardTier)
		}

		raidDetail, err := client.WorldBossRaid(resolvedRaid.RaidID)
		if err != nil {
			t.Fatalf("load raid detail for client %d: %v", index, err)
		}
		if len(raidDetail.Members) != 6 {
			t.Fatalf("expected raid detail to expose 6 members for client %d, got %d", index, len(raidDetail.Members))
		}
		if _, ok := memberByCharacter[stateAfter.Character.CharacterID]; !ok {
			t.Fatalf("expected character %s to appear in resolved raid member list", stateAfter.Character.CharacterID)
		}
		logAccountSnapshot(t, client, fmt.Sprintf("combat-boss client-%d after boss", index))
	}
}

func provisionWorldBossParty(t *testing.T, harness *Harness, prefix string) []*Client {
	t.Helper()

	clients := make([]*Client, 0, 6)
	for index := 0; index < 6; index++ {
		botName := fmt.Sprintf("%s-bot-%d", prefix, index)
		characterName := fmt.Sprintf("%s-char-%d", prefix, index)
		client := registerCombatReadyCharacter(t, harness, botName, characterName, "mage", "staff")
		clients = append(clients, client)
	}

	return clients
}

func materialQuantity(materials []MaterialBalance, key string) int {
	for _, material := range materials {
		if material.MaterialKey == key {
			return material.Quantity
		}
	}
	return 0
}

func TestFirstDaySixAccountBossProgression(t *testing.T) {
	harness := NewHarness(t)
	clients := make([]*Client, 0, 6)
	preBossStates := make([]State, 0, 6)

	for index := 0; index < 6; index++ {
		botName := fmt.Sprintf("firstdayboss-bot-%d", index)
		characterName := fmt.Sprintf("FirstDayBoss-%d", index)
		client := registerCivilianCharacter(t, harness, botName, characterName)
		preSoloInventory, err := client.Inventory()
		if err != nil {
			t.Fatalf("load inventory before first-day loop for client %d: %v", index, err)
		}
		if err := completeFirstDayContractBoard(client); err != nil {
			t.Fatalf("complete first-day contracts for client %d: %v", index, err)
		}
		logAccountSnapshot(t, client, fmt.Sprintf("firstday-boss client-%d after contracts", index))
		if err := buyStarterCombatPrep(client); err != nil {
			t.Fatalf("buy first-day prep for client %d: %v", index, err)
		}
		logAccountSnapshot(t, client, fmt.Sprintf("firstday-boss client-%d after starter prep", index))
		postPrepState, err := client.State()
		if err != nil {
			t.Fatalf("load post-prep state for client %d: %v", index, err)
		}
		preDungeonPower := postPrepState.CombatPower.PanelPowerScore
		equippedUpgrade, dungeonRuns, err := clearAndClaimNoviceDungeon(client, 2)
		if err != nil {
			t.Fatalf("run first-day novice dungeon for client %d: %v", index, err)
		}
		if !equippedUpgrade {
			t.Fatalf("expected first-day client %d to receive a directly equipable upgrade", index)
		}
		state, err := client.State()
		if err != nil {
			t.Fatalf("load final first-day state for client %d: %v", index, err)
		}
		if state.CombatPower.PanelPowerScore <= preDungeonPower {
			t.Fatalf("expected first-day client %d power to increase after dungeon, before=%d after=%d", index, preDungeonPower, state.CombatPower.PanelPowerScore)
		}
		logDungeonRunSummary(t, fmt.Sprintf("firstday-boss client-%d dungeon runs", index), dungeonRuns)
		postSoloInventory, err := client.Inventory()
		if err != nil {
			t.Fatalf("load inventory after first-day loop for client %d: %v", index, err)
		}
		logDungeonLootSummary(t, fmt.Sprintf("firstday-boss client-%d solo loot", index), preSoloInventory, postSoloInventory)
		logAccountSnapshot(t, client, fmt.Sprintf("firstday-boss client-%d after solo loop", index))
		clients = append(clients, client)
		preBossStates = append(preBossStates, state)
	}

	var resolvedRaid *WorldBossRaid
	for index, client := range clients {
		joinResult, err := client.JoinWorldBossQueue()
		if err != nil {
			t.Fatalf("join world boss queue for first-day client %d: %v", index, err)
		}
		if index < len(clients)-1 {
			if joinResult.ResolvedRaid != nil {
				t.Fatalf("expected no raid resolution before sixth first-day join, got %s", joinResult.ResolvedRaid.RaidID)
			}
			continue
		}
		if joinResult.ResolvedRaid == nil {
			t.Fatal("expected sixth first-day join to resolve a world boss raid")
		}
		resolvedRaid = joinResult.ResolvedRaid
	}

	if resolvedRaid == nil {
		t.Fatal("expected resolved raid for first-day boss progression")
	}
	logWorldBossRaidSummary(t, resolvedRaid, "firstday-boss raid")
	if resolvedRaid.RewardPackage.RewardGold <= 0 {
		t.Fatalf("expected world boss reward package to include gold, got %d", resolvedRaid.RewardPackage.RewardGold)
	}

	for index, client := range clients {
		stateAfter, err := client.State()
		if err != nil {
			t.Fatalf("load post-boss state for first-day client %d: %v", index, err)
		}
		if stateAfter.Character.Gold <= preBossStates[index].Character.Gold {
			t.Fatalf("expected boss rewards to increase gold for first-day client %d, before=%d after=%d", index, preBossStates[index].Character.Gold, stateAfter.Character.Gold)
		}
		if materialQuantity(stateAfter.Materials, "reforge_stone") <= materialQuantity(preBossStates[index].Materials, "reforge_stone") {
			t.Fatalf("expected boss rewards to grant reforge stones for first-day client %d", index)
		}
		if stateAfter.Character.Gold < 800 {
			t.Fatalf("expected first-day client %d to afford profession change after boss rewards, got %d gold", index, stateAfter.Character.Gold)
		}
		logAccountSnapshot(t, client, fmt.Sprintf("firstday-boss client-%d after boss", index))
	}
}
