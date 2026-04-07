package quests

import (
	"strings"
	"testing"
	"time"

	"clawgame/apps/api/internal/characters"
	"clawgame/apps/api/internal/world"
)

type stubRepo struct {
	boards map[string]StoredBoard
}

func (s *stubRepo) LoadBoards() ([]StoredBoard, error) {
	items := make([]StoredBoard, 0, len(s.boards))
	for _, board := range s.boards {
		items = append(items, board)
	}
	return items, nil
}

func (s *stubRepo) SaveBoard(board StoredBoard) error {
	if s.boards == nil {
		s.boards = make(map[string]StoredBoard)
	}
	s.boards[board.CharacterID] = board
	return nil
}

func TestQuestRuntimePersistsAcrossServiceReload(t *testing.T) {
	repo := &stubRepo{boards: make(map[string]StoredBoard)}

	service, err := NewServiceWithRepository(repo)
	if err != nil {
		t.Fatalf("NewServiceWithRepository failed: %v", err)
	}

	limits := characters.DailyLimits{}

	var character characters.Summary
	var nightmareQuestID string
	for attempt := 0; attempt < 24; attempt++ {
		candidate := characters.Summary{
			CharacterID: "char_test_" + strings.Repeat("x", attempt+1),
			Name:        "PersistentRunner",
		}
		board := service.ListQuests(candidate, limits)
		for _, quest := range board.Quests {
			if quest.TemplateType == "clear_dungeon" && quest.Difficulty == "nightmare" {
				character = candidate
				nightmareQuestID = quest.QuestID
				break
			}
		}
		if nightmareQuestID != "" {
			break
		}
	}
	if nightmareQuestID == "" {
		t.Fatal("expected nightmare clear_dungeon quest in sampled boards")
	}

	if _, _, err := service.AcceptQuest(character, nightmareQuestID, limits); err != nil {
		t.Fatalf("AcceptQuest failed: %v", err)
	}
	if _, _, err := service.AdvanceQuestInteraction(character, nightmareQuestID, "inspect_clue"); err != nil {
		t.Fatalf("AdvanceQuestInteraction failed: %v", err)
	}
	if _, _, err := service.ApplyQuestChoice(character, nightmareQuestID, "follow_standard_brief"); err != nil {
		t.Fatalf("ApplyQuestChoice failed: %v", err)
	}

	reloaded, err := NewServiceWithRepository(repo)
	if err != nil {
		t.Fatalf("reloaded NewServiceWithRepository failed: %v", err)
	}

	quest, runtime, err := reloaded.GetQuestRuntime(character.CharacterID, nightmareQuestID)
	if err != nil {
		t.Fatalf("GetQuestRuntime failed after reload: %v", err)
	}
	if quest.Difficulty != "nightmare" {
		t.Fatalf("expected difficulty nightmare after reload, got %q", quest.Difficulty)
	}
	if runtime.CurrentStepKey != "clear_target_dungeon" {
		t.Fatalf("expected current_step_key clear_target_dungeon after reload, got %q", runtime.CurrentStepKey)
	}
	if runtime.CurrentStepLabel == "" || runtime.CurrentStepHint == "" {
		t.Fatal("expected runtime step metadata after reload")
	}
	if runtime.State["selected_choice_key"] != "follow_standard_brief" {
		t.Fatalf("expected selected_choice_key to persist, got %#v", runtime.State["selected_choice_key"])
	}
	if runtime.State["selected_choice_label"] == nil {
		t.Fatal("expected selected_choice_label to persist")
	}
	foundInspect := false
	for _, step := range runtime.CompletedStepKeys {
		if step == "inspect_clue" {
			foundInspect = true
			break
		}
	}
	if !foundInspect {
		t.Fatal("expected inspect_clue to persist in completed steps")
	}
	if len(runtime.Clues) < 2 {
		t.Fatal("expected runtime clues to include both base and choice/interaction clues after reload")
	}
}

func TestQuestTemplateCatalogLoadsFromYAML(t *testing.T) {
	catalog, err := defaultQuestTemplateCatalog()
	if err != nil {
		t.Fatalf("defaultQuestTemplateCatalog failed: %v", err)
	}
	if len(catalog.daily) != 6 {
		t.Fatalf("expected 6 daily quest templates, got %d", len(catalog.daily))
	}
	if len(catalog.supplemental) != 2 {
		t.Fatalf("expected 2 supplemental quest templates, got %d", len(catalog.supplemental))
	}
	if catalog.daily[0].TemplateType != "kill_region_enemies" {
		t.Fatalf("expected first daily template kill_region_enemies, got %q", catalog.daily[0].TemplateType)
	}
	if catalog.daily[4].TemplateType != "investigate_anomaly" {
		t.Fatalf("expected investigate_anomaly template in daily catalog, got %q", catalog.daily[4].TemplateType)
	}
	if len(catalog.daily[4].Spec.ChoiceSpecs) == 0 {
		t.Fatal("expected investigation template to carry choice specs from yaml")
	}
}

func TestInstantiateQuestAutoAcceptsAndRollsRewardFromTemplate(t *testing.T) {
	definition, ok := findQuestDefinition(characters.QuestSummary{
		TemplateType: "deliver_supplies",
		Difficulty:   "normal",
		FlowKind:     "delivery",
	})
	if !ok {
		t.Fatal("expected normal deliver_supplies definition to exist")
	}

	quest := instantiateQuest(definition, characters.Summary{}, "board_test", "board_test|seed")
	if quest.Status != "accepted" {
		t.Fatalf("expected generated quest to be auto-accepted, got %q", quest.Status)
	}
	if quest.RewardGold < 85 || quest.RewardGold > 250 {
		t.Fatalf("expected rolled deliver reward in supported range, got %d", quest.RewardGold)
	}
}

func TestValidateQuestTemplateDefinitionRejectsBrokenStepRefs(t *testing.T) {
	err := validateQuestTemplateDefinition(questTemplateDefinition{
		TemplateType: "broken_quest",
		Spec: questRuntimeSpec{
			InitialStepKey:    "start",
			CompletionStepKey: "missing_finish",
			Steps: []questStepSpec{
				{Key: "start"},
			},
		},
	})
	if err == nil {
		t.Fatal("expected validateQuestTemplateDefinition to fail for missing completion step")
	}
	if !strings.Contains(err.Error(), "missing step") {
		t.Fatalf("expected missing step error, got %v", err)
	}
}

func TestQuestBoardRefillsToFourAcrossDailyReset(t *testing.T) {
	service := NewService()
	now := time.Date(2026, 4, 6, 10, 0, 0, 0, time.FixedZone("CST", 8*3600))
	service.clock = func() time.Time { return now }

	character := characters.Summary{
		CharacterID: "char_daily",
		Name:        "DailyRunner",
	}

	board := service.ListQuests(character, characters.DailyLimits{})
	if len(board.Quests) != characters.DailyQuestBoardSize {
		t.Fatalf("expected %d quests on first board, got %d", characters.DailyQuestBoardSize, len(board.Quests))
	}

	for i := 0; i < 2; i++ {
		board.Quests[i].Status = "submitted"
	}
	service.boardByCharacter[character.CharacterID] = boardRecord{
		boardID:     board.BoardID,
		resetDate:   service.businessDate(),
		status:      "active",
		rerollCount: 0,
		quests:      board.Quests,
	}

	now = now.Add(24 * time.Hour)
	refilled := service.ListQuests(character, characters.DailyLimits{})
	if len(refilled.Quests) != characters.DailyQuestBoardSize {
		t.Fatalf("expected board to refill back to %d, got %d", characters.DailyQuestBoardSize, len(refilled.Quests))
	}

	active := 0
	for _, quest := range refilled.Quests {
		if quest.Status == "accepted" || quest.Status == "completed" {
			active++
		}
	}
	if active != characters.DailyQuestBoardSize {
		t.Fatalf("expected %d active quests after refill, got %d", characters.DailyQuestBoardSize, active)
	}
}

func TestCurioFollowupDoesNotOverflowFixedDailyBoard(t *testing.T) {
	service := NewService()
	now := time.Date(2026, 4, 6, 10, 0, 0, 0, time.FixedZone("CST", 8*3600))
	service.clock = func() time.Time { return now }

	character := characters.Summary{
		CharacterID: "char_curio_fixed_board",
		Name:        "CurioFixedBoard",
	}

	board := service.ListQuests(character, characters.DailyLimits{})
	if len(board.Quests) != characters.DailyQuestBoardSize {
		t.Fatalf("expected fixed board size %d, got %d", characters.DailyQuestBoardSize, len(board.Quests))
	}

	board.Quests[0].Status = "submitted"
	service.boardByCharacter[character.CharacterID] = boardRecord{
		boardID:     board.BoardID,
		resetDate:   service.businessDate(),
		status:      "active",
		rerollCount: 0,
		quests:      board.Quests,
	}

	view, followup, err := service.EnsureCurioFollowupQuest(character, world.CurioQuestSeed{
		TemplateType:     "curio_followup_delivery",
		Title:            "Escort the Shrine Witness",
		Description:      "Take the witness back to the village.",
		TargetRegionID:   "greenfield_village",
		RewardGold:       120,
		RewardReputation: 20,
	}, characters.DailyLimits{})
	if err != nil {
		t.Fatalf("EnsureCurioFollowupQuest failed: %v", err)
	}
	if followup != nil {
		t.Fatal("expected curio followup not to create a fifth same-day quest")
	}
	if len(view.Quests) != characters.DailyQuestBoardSize {
		t.Fatalf("expected board to remain at %d quests, got %d", characters.DailyQuestBoardSize, len(view.Quests))
	}
}

func TestCivilianBoardUsesFourNonCombatContracts(t *testing.T) {
	service := NewService()
	service.clock = func() time.Time {
		return time.Date(2026, 4, 6, 10, 0, 0, 0, time.FixedZone("CST", 8*3600))
	}

	character := characters.Summary{
		CharacterID: "char_civilian_day_one",
		Name:        "CivilianDayOne",
		Class:       "civilian",
	}

	board := service.ListQuests(character, characters.DailyLimits{})
	if len(board.Quests) != characters.DailyQuestBoardSize {
		t.Fatalf("expected civilian board to fill to %d quests, got %d", characters.DailyQuestBoardSize, len(board.Quests))
	}
	for _, quest := range board.Quests {
		switch quest.TemplateType {
		case "kill_region_enemies", "collect_materials", "clear_dungeon":
			t.Fatalf("expected civilian day-one board to exclude combat contract %s", quest.TemplateType)
		}
	}
}
