package quests

import (
	"strings"
	"testing"

	"clawgame/apps/api/internal/characters"
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

	character := characters.Summary{
		CharacterID: "char_test",
		Name:        "PersistentRunner",
		Rank:        "low",
	}
	limits := characters.DailyLimits{}

	board := service.ListQuests(character, limits)
	var nightmareQuestID string
	for _, quest := range board.Quests {
		if quest.TemplateType == "clear_dungeon" && quest.Difficulty == "nightmare" {
			nightmareQuestID = quest.QuestID
			break
		}
	}
	if nightmareQuestID == "" {
		t.Fatal("expected nightmare clear_dungeon quest in board")
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

func TestInstantiateQuestAppliesRankOverridesFromYAML(t *testing.T) {
	definition, ok := findQuestDefinition(characters.QuestSummary{
		TemplateType: "deliver_supplies",
		Difficulty:   "normal",
		FlowKind:     "delivery",
	})
	if !ok {
		t.Fatal("expected normal deliver_supplies definition to exist")
	}

	quest := instantiateQuest(definition, characters.Summary{Rank: "mid"}, "board_test")
	if quest.TargetRegionID != "sunscar_desert_outskirts" {
		t.Fatalf("expected mid-rank override target sunscar_desert_outskirts, got %q", quest.TargetRegionID)
	}
	if quest.Title != "Deliver Desert Provisions" {
		t.Fatalf("expected overridden title Deliver Desert Provisions, got %q", quest.Title)
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
