package e2e

import (
	"fmt"
	"sort"
	"strings"
	"testing"
)

func registerCivilianCharacter(t *testing.T, harness *Harness, botName, characterName string) *Client {
	t.Helper()

	client := harness.client.Clone()
	if err := client.Register(botName, harness.password); err != nil {
		t.Fatalf("register %s: %v", botName, err)
	}
	if err := client.Login(botName, harness.password); err != nil {
		t.Fatalf("login %s: %v", botName, err)
	}
	if err := client.CreateCharacter(characterName, "civilian", ""); err != nil {
		t.Fatalf("create character %s: %v", characterName, err)
	}
	return client
}

func registerCombatReadyCharacter(t *testing.T, harness *Harness, botName, characterName, className, weaponStyle string) *Client {
	t.Helper()

	client := registerCivilianCharacter(t, harness, botName, characterName)
	if err := harness.api.GrantSeasonXP(client.token, 15000); err != nil {
		t.Fatalf("grant season xp for %s: %v", characterName, err)
	}
	if err := harness.api.GrantGold(client.token, 1200); err != nil {
		t.Fatalf("grant gold for %s: %v", characterName, err)
	}
	if err := client.ChooseProfession(className, weaponStyle); err != nil {
		t.Fatalf("choose profession for %s: %v", characterName, err)
	}
	return client
}

func completeFirstDayContractBoard(client *Client) error {
	if err := runQuestStage(client, isNonCombatQuest, 48); err != nil {
		return fmt.Errorf("complete non-combat contract stage: %w", err)
	}

	board, err := client.Quests()
	if err != nil {
		return fmt.Errorf("reload quest board after first-day stages: %w", err)
	}
	for _, quest := range board {
		if quest.Status != "submitted" {
			return fmt.Errorf("quest %s remained %s after first-day stages", quest.QuestID, quest.Status)
		}
	}
	return nil
}

func runQuestStage(client *Client, shouldProgress func(Quest) bool, maxSteps int) error {
	for step := 0; step < maxSteps; step++ {
		board, err := client.Quests()
		if err != nil {
			return fmt.Errorf("reload quest board: %w", err)
		}

		pending := false
		for _, quest := range board {
			if quest.Status == "completed" {
				pending = true
				break
			}
			if quest.Status != "submitted" && shouldProgress(quest) {
				pending = true
				break
			}
		}
		if !pending {
			return nil
		}

		progressed, err := advanceQuestStageStep(client, board, shouldProgress)
		if err != nil {
			return err
		}
		if !progressed {
			return fmt.Errorf("quest stage stalled")
		}
	}

	return fmt.Errorf("quest stage exceeded step budget")
}

func advanceQuestStageStep(client *Client, board []Quest, shouldProgress func(Quest) bool) (bool, error) {
	for _, quest := range board {
		if quest.Status != "completed" {
			continue
		}
		if err := client.SubmitQuest(quest.QuestID); err != nil {
			return false, fmt.Errorf("submit quest %s: %w", quest.QuestID, err)
		}
		return true, nil
	}

	for _, quest := range board {
		if quest.Status != "accepted" || !shouldProgress(quest) {
			continue
		}
		return advanceAcceptedQuest(client, quest)
	}

	return false, nil
}

func advanceAcceptedQuest(client *Client, quest Quest) (bool, error) {
	detail, err := client.QuestDetail(quest.QuestID)
	if err != nil {
		return false, fmt.Errorf("load quest detail %s: %w", quest.QuestID, err)
	}

	switch strings.TrimSpace(detail.Runtime.SuggestedActionType) {
	case "quest_interact":
		interaction, _ := detail.Runtime.SuggestedActionArgs["interaction"].(string)
		if strings.TrimSpace(interaction) == "" {
			return false, fmt.Errorf("quest %s missing suggested interaction", quest.QuestID)
		}
		if err := client.QuestInteract(quest.QuestID, interaction); err != nil {
			return false, fmt.Errorf("interact quest %s: %w", quest.QuestID, err)
		}
		return true, nil
	case "quest_choice":
		if len(detail.Runtime.AvailableChoices) == 0 {
			return false, fmt.Errorf("quest %s missing available choices", quest.QuestID)
		}
		if err := client.QuestChoice(quest.QuestID, detail.Runtime.AvailableChoices[0].ChoiceKey); err != nil {
			return false, fmt.Errorf("choose quest %s: %w", quest.QuestID, err)
		}
		return true, nil
	}

	state, err := client.State()
	if err != nil {
		return false, fmt.Errorf("load state for quest %s: %w", quest.QuestID, err)
	}

	switch quest.TemplateType {
	case "kill_region_enemies":
		if state.Character.LocationRegionID != quest.TargetRegionID {
			if err := client.Travel(quest.TargetRegionID); err != nil {
				return false, fmt.Errorf("travel to kill region %s: %w", quest.TargetRegionID, err)
			}
			return true, nil
		}
		result, err := client.FieldEncounter("hunt")
		if err != nil {
			return false, fmt.Errorf("hunt encounter for quest %s: %w", quest.QuestID, err)
		}
		if !result.ActionResult.Victory {
			return false, fmt.Errorf("hunt encounter for quest %s resolved as defeat", quest.QuestID)
		}
		return true, nil
	case "collect_materials":
		if state.Character.LocationRegionID != quest.TargetRegionID {
			if err := client.Travel(quest.TargetRegionID); err != nil {
				return false, fmt.Errorf("travel to gather region %s: %w", quest.TargetRegionID, err)
			}
			return true, nil
		}
		result, err := client.FieldEncounter("gather")
		if err != nil {
			return false, fmt.Errorf("gather encounter for quest %s: %w", quest.QuestID, err)
		}
		if !result.ActionResult.Victory {
			return false, fmt.Errorf("gather encounter for quest %s resolved as defeat", quest.QuestID)
		}
		return true, nil
	case "deliver_supplies", "investigate_anomaly":
		if strings.TrimSpace(quest.TargetRegionID) == "" {
			return false, fmt.Errorf("quest %s missing target region", quest.QuestID)
		}
		if state.Character.LocationRegionID == quest.TargetRegionID {
			fallbackRegion := "main_city"
			if quest.TargetRegionID == "main_city" {
				fallbackRegion = "greenfield_village"
			}
			if err := client.Travel(fallbackRegion); err != nil {
				return false, fmt.Errorf("travel away from target for quest %s: %w", quest.QuestID, err)
			}
			return true, nil
		}
		if err := client.Travel(quest.TargetRegionID); err != nil {
			return false, fmt.Errorf("travel for quest %s: %w", quest.QuestID, err)
		}
		return true, nil
	default:
		return false, fmt.Errorf("unsupported quest template %s", quest.TemplateType)
	}
}

func isFieldQuest(quest Quest) bool {
	switch strings.TrimSpace(quest.TemplateType) {
	case "kill_region_enemies", "collect_materials":
		return true
	default:
		return false
	}
}

func isNonCombatQuest(quest Quest) bool {
	if isFieldQuest(quest) {
		return false
	}
	switch strings.TrimSpace(quest.TemplateType) {
	case "clear_dungeon", "kill_dungeon_elite":
		return false
	default:
		return true
	}
}

func buyStarterCombatPrep(client *Client) error {
	state, err := client.State()
	if err != nil {
		return fmt.Errorf("load state before combat prep: %w", err)
	}
	if state.Character.LocationRegionID != "main_city" {
		if err := client.Travel("main_city"); err != nil {
			return fmt.Errorf("travel to main city for combat prep: %w", err)
		}
	}

	shopItems, err := client.BuildingShopInventory("equipment_shop_main_city")
	if err != nil {
		return fmt.Errorf("load equipment shop inventory: %w", err)
	}
	type prepCandidate struct {
		catalogID string
		priceGold int
	}
	candidates := make([]prepCandidate, 0, len(shopItems))
	for _, item := range shopItems {
		if item.ItemType != "equipment" || item.Slot == "weapon" {
			continue
		}
		candidates = append(candidates, prepCandidate{catalogID: item.CatalogID, priceGold: item.PriceGold})
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].priceGold != candidates[j].priceGold {
			return candidates[i].priceGold < candidates[j].priceGold
		}
		return candidates[i].catalogID < candidates[j].catalogID
	})
	candidateIDs := make([]string, 0, 2)
	remainingGold := state.Character.Gold - 160
	if remainingGold < 0 {
		remainingGold = 0
	}
	for _, item := range candidates {
		if item.priceGold > remainingGold {
			continue
		}
		candidateIDs = append(candidateIDs, item.catalogID)
		remainingGold -= item.priceGold
		if len(candidateIDs) == 2 {
			break
		}
	}
	if len(candidateIDs) == 0 {
		return fmt.Errorf("expected at least one affordable non-weapon item in the equipment shop")
	}
	for _, catalogID := range candidateIDs {
		if err := client.PurchaseBuildingItem("equipment_shop_main_city", catalogID); err != nil {
			return fmt.Errorf("purchase %s from equipment shop: %w", catalogID, err)
		}
	}
	for _, potion := range []struct {
		catalogID string
		priceGold int
	}{
		{catalogID: "potion_hp_t2", priceGold: 22},
		{catalogID: "potion_hp_t2", priceGold: 22},
		{catalogID: "potion_atk_t2", priceGold: 24},
		{catalogID: "potion_atk_t2", priceGold: 24},
	} {
		if potion.priceGold > remainingGold {
			continue
		}
		if err := client.PurchaseBuildingItem("apothecary_main_city", potion.catalogID); err != nil {
			return fmt.Errorf("purchase %s from apothecary: %w", potion.catalogID, err)
		}
		remainingGold -= potion.priceGold
	}

	inventory, err := client.Inventory()
	if err != nil {
		return fmt.Errorf("load inventory after starter combat prep: %w", err)
	}
	for _, item := range inventory.Inventory {
		for _, catalogID := range candidateIDs {
			if item.CatalogID != catalogID {
				continue
			}
			if _, err := client.EquipItem(item.ItemID); err != nil {
				return fmt.Errorf("equip starter combat item %s: %w", item.CatalogID, err)
			}
		}
	}

	return nil
}

func runFirstDayCivilianPveLoop(client *Client) (State, error) {
	if err := completeFirstDayContractBoard(client); err != nil {
		return State{}, fmt.Errorf("complete first-day contract board: %w", err)
	}
	if err := buyStarterCombatPrep(client); err != nil {
		return State{}, fmt.Errorf("buy first-day combat prep: %w", err)
	}

	stateAfterContracts, err := client.State()
	if err != nil {
		return State{}, fmt.Errorf("load state after contracts: %w", err)
	}
	preDungeonPower := stateAfterContracts.CombatPower.PanelPowerScore

	equippedUpgrade, _, err := clearAndClaimNoviceDungeon(client, 2)
	if err != nil {
		return State{}, fmt.Errorf("clear novice dungeon: %w", err)
	}
	if !equippedUpgrade {
		return State{}, fmt.Errorf("novice dungeon did not produce a directly equipable upgrade")
	}

	finalState, err := client.State()
	if err != nil {
		return State{}, fmt.Errorf("load final first-day state: %w", err)
	}
	if finalState.CombatPower.PanelPowerScore <= preDungeonPower {
		return State{}, fmt.Errorf("expected combat power to increase after novice dungeon, before=%d after=%d", preDungeonPower, finalState.CombatPower.PanelPowerScore)
	}
	if finalState.Limits.DungeonEntryUsed < 1 {
		return State{}, fmt.Errorf("expected at least one dungeon reward claim used, got %d", finalState.Limits.DungeonEntryUsed)
	}
	return finalState, nil
}

func clearAndClaimNoviceDungeon(client *Client, maxClaims int) (bool, []Run, error) {
	runs := make([]Run, 0, maxClaims)
	for claimed := 0; claimed < maxClaims; claimed++ {
		run, err := client.EnterDungeon("ancient_catacomb_v1", "normal", []string{"potion_hp_t2", "potion_atk_t2"})
		if err != nil {
			return false, runs, fmt.Errorf("enter novice dungeon run %d: %w", claimed+1, err)
		}
		runs = append(runs, run)
		if run.RunStatus != "cleared" || !run.RewardClaimable {
			return false, runs, fmt.Errorf("novice dungeon run %d did not clear cleanly, status=%s claimable=%v", claimed+1, run.RunStatus, run.RewardClaimable)
		}
		if _, err := client.ClaimDungeonRewards(run.RunID); err != nil {
			return false, runs, fmt.Errorf("claim novice dungeon run %d: %w", claimed+1, err)
		}

		equippedAny, err := equipBestAvailableUpgrades(client)
		if err != nil {
			return false, runs, err
		}
		if equippedAny {
			return true, runs, nil
		}
	}
	return false, runs, nil
}

func equipBestAvailableUpgrades(client *Client) (bool, error) {
	equippedAny := false

	for iteration := 0; iteration < 12; iteration++ {
		view, err := client.Inventory()
		if err != nil {
			return false, fmt.Errorf("load inventory: %w", err)
		}

		bestHint := UpgradeHint{}
		for _, hint := range view.UpgradeHints {
			if hint.Source != "inventory" || !hint.DirectlyEquipable || hint.ScoreDelta <= 0 || strings.TrimSpace(hint.ItemID) == "" {
				continue
			}
			if hint.ScoreDelta > bestHint.ScoreDelta {
				bestHint = hint
			}
		}

		if strings.TrimSpace(bestHint.ItemID) == "" {
			return equippedAny, nil
		}

		if _, err := client.EquipItem(bestHint.ItemID); err != nil {
			return false, fmt.Errorf("equip item %s: %w", bestHint.ItemID, err)
		}
		equippedAny = true
	}

	return equippedAny, nil
}
