package e2e

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"testing"
)

const professionChangeGoldTarget = 800

func logAccountSnapshot(t *testing.T, client *Client, label string) {
	t.Helper()

	state, err := client.State()
	if err != nil {
		t.Logf("[%s] snapshot skipped: state error: %v", label, err)
		return
	}
	inventory, err := client.Inventory()
	if err != nil {
		t.Logf("[%s] snapshot skipped: inventory error: %v", label, err)
		return
	}

	goldGap := professionChangeGoldTarget - state.Character.Gold
	if goldGap < 0 {
		goldGap = 0
	}

	t.Logf("[%s] name=%s class=%s weapon=%s lvl=%d gold=%d rep=%d power=%d tier=%s equip_score=%d quests=%d/%d dungeon_claims=%d/%d bonus_claims=%d gold_to_prof=%d mats=%s equipped=%s",
		label,
		state.Character.Name,
		state.Character.Class,
		state.Character.WeaponStyle,
		state.Character.SeasonLevel,
		state.Character.Gold,
		state.Character.Reputation,
		state.CombatPower.PanelPowerScore,
		state.CombatPower.PowerTier,
		inventory.EquipmentScore,
		state.Limits.QuestCompletionUsed,
		state.Limits.QuestCompletionCap,
		state.Limits.DungeonEntryUsed,
		state.Limits.DungeonEntryCap,
		state.Limits.BonusDungeonEntryPurchased,
		goldGap,
		formatMaterialBalances(state.Materials),
		formatEquippedItems(inventory.Equipped),
	)
}

func logWorldBossRaidSummary(t *testing.T, raid *WorldBossRaid, label string) {
	t.Helper()
	if raid == nil {
		t.Logf("[%s] no raid summary available", label)
		return
	}
	members := make([]string, 0, len(raid.Members))
	for _, member := range raid.Members {
		share := 0.0
		if raid.TotalDamage > 0 {
			share = math.Round((float64(member.DamageDealt)/float64(raid.TotalDamage))*1000) / 10
		}
		members = append(members, fmt.Sprintf("%s:%d(%.1f%%)", member.Name, member.DamageDealt, share))
	}
	t.Logf("[%s] raid=%s tier=%s total_damage=%d boss_hp=%d reward_gold=%d members=%s",
		label,
		raid.RaidID,
		raid.RewardTier,
		raid.TotalDamage,
		raid.BossMaxHP,
		raid.RewardPackage.RewardGold,
		strings.Join(members, ", "),
	)
}

func logDungeonRunSummary(t *testing.T, label string, runs []Run) {
	t.Helper()
	if len(runs) == 0 {
		t.Logf("[%s] no dungeon runs recorded", label)
		return
	}
	parts := make([]string, 0, len(runs))
	for index, run := range runs {
		rating := "unrated"
		if run.CurrentRating != nil && strings.TrimSpace(*run.CurrentRating) != "" {
			rating = strings.TrimSpace(*run.CurrentRating)
		}
		parts = append(parts, fmt.Sprintf("run%d:%s rooms=%d rating=%s claimable=%v", index+1, run.RunStatus, run.HighestRoomCleared, rating, run.RewardClaimable))
	}
	t.Logf("[%s] %s", label, strings.Join(parts, "; "))
}

func logDungeonLootSummary(t *testing.T, label string, before, after InventoryView) {
	t.Helper()
	beforeIDs := make(map[string]InventoryItem, len(before.Inventory))
	for _, item := range before.Inventory {
		beforeIDs[item.ItemID] = item
	}
	newItems := make([]string, 0)
	for _, item := range after.Inventory {
		if _, ok := beforeIDs[item.ItemID]; ok {
			continue
		}
		newItems = append(newItems, formatInventoryItem(item))
	}
	if len(newItems) == 0 {
		t.Logf("[%s] no new inventory items detected after dungeon claim", label)
		return
	}
	sort.Strings(newItems)
	t.Logf("[%s] new_items=%s", label, strings.Join(newItems, "; "))
}

func formatMaterialBalances(materials []MaterialBalance) string {
	if len(materials) == 0 {
		return "none"
	}
	items := make([]string, 0, len(materials))
	for _, material := range materials {
		if material.Quantity <= 0 {
			continue
		}
		items = append(items, fmt.Sprintf("%s=%d", material.MaterialKey, material.Quantity))
	}
	if len(items) == 0 {
		return "none"
	}
	sort.Strings(items)
	return strings.Join(items, ",")
}

func formatEquippedItems(items []InventoryItem) string {
	if len(items) == 0 {
		return "none"
	}
	parts := make([]string, 0, len(items))
	for _, item := range items {
		parts = append(parts, formatInventoryItem(item))
	}
	sort.Strings(parts)
	return strings.Join(parts, "; ")
}

func formatInventoryItem(item InventoryItem) string {
	setPart := ""
	if strings.TrimSpace(item.SetID) != "" {
		setPart = fmt.Sprintf("[%s]", item.SetID)
	}
	return fmt.Sprintf("%s:%s:%s%s", item.Slot, item.Rarity, item.Name, setPart)
}
