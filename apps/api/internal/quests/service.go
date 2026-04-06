package quests

import (
	"errors"
	"fmt"
	"hash/fnv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"clawgame/apps/api/internal/characters"
	"clawgame/apps/api/internal/world"
)

const (
	businessTimezone = "Asia/Shanghai"
)

var (
	ErrQuestNotFound           = errors.New("quest not found")
	ErrQuestInvalidState       = errors.New("quest invalid state")
	ErrQuestChoiceNotAvailable = errors.New("quest choice not available")
	ErrQuestInteractionInvalid = errors.New("quest interaction invalid")
)

type QuestChoice struct {
	ChoiceKey string `json:"choice_key" yaml:"choice_key"`
	Label     string `json:"label" yaml:"label"`
}

type QuestClue struct {
	ClueKey string `json:"clue_key" yaml:"clue_key"`
	Label   string `json:"label" yaml:"label"`
}

type QuestRuntime struct {
	QuestID             string         `json:"quest_id"`
	CurrentStepKey      string         `json:"current_step_key,omitempty"`
	CurrentStepLabel    string         `json:"current_step_label,omitempty"`
	CurrentStepHint     string         `json:"current_step_hint,omitempty"`
	SuggestedActionType string         `json:"suggested_action_type,omitempty"`
	SuggestedActionArgs map[string]any `json:"suggested_action_args,omitempty"`
	CompletedStepKeys   []string       `json:"completed_step_keys"`
	AvailableChoices    []QuestChoice  `json:"available_choices"`
	Clues               []QuestClue    `json:"clues"`
	State               map[string]any `json:"state_json"`
}

type Trigger struct {
	TriggerType        string
	RegionID           string
	Approach           string
	EnemiesDefeated    int
	MaterialsCollected int
}

type ProgressChange struct {
	Quest          characters.QuestSummary `json:"quest"`
	PreviousStatus string                  `json:"previous_status"`
	CurrentStatus  string                  `json:"current_status"`
	ProgressDelta  int                     `json:"progress_delta"`
	Completed      bool                    `json:"completed"`
}

type questRuntimeState struct {
	SelectedChoiceKey string
	CompletedStepKeys []string
	CurrentStepKey    string
	Clues             []QuestClue
	State             map[string]any
}

type questStepSpec struct {
	Key        string `yaml:"key"`
	Label      string `yaml:"label"`
	Hint       string `yaml:"hint"`
	ActionType string `yaml:"action_type"`
}

type questChoiceSpec struct {
	ChoiceKey   string         `yaml:"choice_key"`
	Label       string         `yaml:"label"`
	NextStepKey string         `yaml:"next_step_key"`
	Clues       []QuestClue    `yaml:"clues"`
	State       map[string]any `yaml:"state"`
}

type questInteractionSpec struct {
	StepKey     string         `yaml:"step_key"`
	Label       string         `yaml:"label"`
	Hint        string         `yaml:"hint"`
	NextStepKey string         `yaml:"next_step_key"`
	Clues       []QuestClue    `yaml:"clues"`
	State       map[string]any `yaml:"state"`
}

type questRuntimeSpec struct {
	InitialStepKey       string
	CompletionStepKey    string
	ChoiceStepKey        string
	InspectStepKey       string
	ProgressTriggerType  string
	ProgressSource       string
	RequiresChoice       bool
	RequiresInspection   bool
	RequiresRouteConfirm bool
	ChoiceSpecs          []questChoiceSpec
	InteractionSpecs     []questInteractionSpec
	BaseClues            []QuestClue
	Steps                []questStepSpec
}

type questTemplateDefinition struct {
	TemplateType     string
	Pool             string
	Order            int
	Difficulty       string
	FlowKind         string
	Rarity           string
	Title            string
	Description      string
	TargetRegionID   string
	ProgressTarget   int
	RewardGold       int
	RewardReputation int
	Spec             questRuntimeSpec
}

type BoardView struct {
	BoardID          string                    `json:"board_id"`
	Status           string                    `json:"status"`
	RerollCount      int                       `json:"reroll_count"`
	ActiveQuestCount int                       `json:"active_quest_count"`
	Quests           []characters.QuestSummary `json:"quests"`
	Limits           characters.DailyLimits    `json:"limits"`
}

type StoredQuestRuntime struct {
	SelectedChoiceKey string         `json:"selected_choice_key,omitempty"`
	CompletedStepKeys []string       `json:"completed_step_keys,omitempty"`
	CurrentStepKey    string         `json:"current_step_key,omitempty"`
	Clues             []QuestClue    `json:"clues,omitempty"`
	State             map[string]any `json:"state_json,omitempty"`
}

type StoredBoard struct {
	CharacterID    string
	BoardID        string
	ResetDate      string
	Status         string
	RerollCount    int
	Quests         []characters.QuestSummary
	RuntimeByQuest map[string]StoredQuestRuntime
}

type Repository interface {
	LoadBoards() ([]StoredBoard, error)
	SaveBoard(StoredBoard) error
}

type boardRecord struct {
	boardID     string
	resetDate   string
	status      string
	rerollCount int
	quests      []characters.QuestSummary
}

type Service struct {
	mu               sync.RWMutex
	clock            func() time.Time
	loc              *time.Location
	repo             Repository
	boardByCharacter map[string]boardRecord
	runtimeByQuest   map[string]map[string]*questRuntimeState
}

var questIDCounter uint64

func NewService() *Service {
	service, err := NewServiceWithRepository(nil)
	if err != nil {
		panic(err)
	}

	return service
}

func NewServiceWithRepository(repo Repository) (*Service, error) {
	if _, err := defaultQuestTemplateCatalog(); err != nil {
		return nil, err
	}

	service := &Service{
		clock:            time.Now,
		loc:              mustLocation(businessTimezone),
		repo:             repo,
		boardByCharacter: make(map[string]boardRecord),
		runtimeByQuest:   make(map[string]map[string]*questRuntimeState),
	}

	if repo == nil {
		return service, nil
	}

	boards, err := repo.LoadBoards()
	if err != nil {
		return nil, err
	}

	for _, stored := range boards {
		quests := make([]characters.QuestSummary, len(stored.Quests))
		copy(quests, stored.Quests)
		service.boardByCharacter[stored.CharacterID] = boardRecord{
			boardID:     stored.BoardID,
			resetDate:   stored.ResetDate,
			status:      stored.Status,
			rerollCount: stored.RerollCount,
			quests:      quests,
		}
		if len(stored.RuntimeByQuest) > 0 {
			service.runtimeByQuest[stored.CharacterID] = make(map[string]*questRuntimeState, len(stored.RuntimeByQuest))
			for questID, runtime := range stored.RuntimeByQuest {
				clues := make([]QuestClue, len(runtime.Clues))
				copy(clues, runtime.Clues)
				completed := make([]string, len(runtime.CompletedStepKeys))
				copy(completed, runtime.CompletedStepKeys)
				state := make(map[string]any, len(runtime.State))
				for key, value := range runtime.State {
					state[key] = value
				}
				service.runtimeByQuest[stored.CharacterID][questID] = &questRuntimeState{
					SelectedChoiceKey: runtime.SelectedChoiceKey,
					CompletedStepKeys: completed,
					CurrentStepKey:    runtime.CurrentStepKey,
					Clues:             clues,
					State:             state,
				}
			}
		}
	}

	return service, nil
}

func difficultyFromRarity(rarity string) string {
	switch strings.TrimSpace(rarity) {
	case "challenge":
		return "nightmare"
	case "uncommon":
		return "hard"
	default:
		return "normal"
	}
}

func runtimeSpecForQuest(quest characters.QuestSummary) questRuntimeSpec {
	quest = enrichQuestSummary(quest)
	if definition, ok := findQuestDefinition(quest); ok {
		return definition.Spec
	}

	switch quest.FlowKind {
	case "delivery":
		return questRuntimeSpec{
			InitialStepKey:      "reach_target_region",
			CompletionStepKey:   "reach_target_region",
			ProgressTriggerType: "travel_completed",
			Steps: []questStepSpec{
				{Key: "reach_target_region", Label: "Travel to target", Hint: "Reach the destination region to advance the delivery."},
				{Key: "turn_in_quest", Label: "Turn in contract", Hint: "Return to the quest board and submit the completed contract."},
			},
		}
	case "dungeon":
		return questRuntimeSpec{
			InitialStepKey:      "clear_target_dungeon",
			CompletionStepKey:   "clear_target_dungeon",
			ProgressTriggerType: "dungeon_cleared",
			Steps: []questStepSpec{
				{Key: "clear_target_dungeon", Label: "Clear dungeon", Hint: "Finish the target dungeon objective."},
				{Key: "turn_in_quest", Label: "Turn in contract", Hint: "Return to the quest board and submit the completed contract."},
			},
		}
	default:
		return questRuntimeSpec{
			InitialStepKey:      "accumulate_progress",
			CompletionStepKey:   "accumulate_progress",
			ProgressTriggerType: "field_resolved",
			ProgressSource:      "enemies_defeated",
			Steps: []questStepSpec{
				{Key: "accumulate_progress", Label: "Advance objective", Hint: "Complete the objective progress in the target region."},
				{Key: "turn_in_quest", Label: "Turn in contract", Hint: "Return to the quest board and submit the completed contract."},
			},
		}
	}
}

func enrichQuestSummary(quest characters.QuestSummary) characters.QuestSummary {
	if strings.TrimSpace(quest.Difficulty) == "" {
		quest.Difficulty = difficultyFromRarity(quest.Rarity)
	}
	if strings.TrimSpace(quest.FlowKind) == "" {
		quest.FlowKind = defaultFlowKindForTemplate(quest.TemplateType)
	}
	if definition, ok := findQuestDefinition(quest); ok {
		if strings.TrimSpace(quest.Difficulty) == "" {
			quest.Difficulty = definition.Difficulty
		}
		if strings.TrimSpace(quest.FlowKind) == "" {
			quest.FlowKind = definition.FlowKind
		}
	}
	return quest
}

func defaultFlowKindForTemplate(templateType string) string {
	switch strings.TrimSpace(templateType) {
	case "deliver_supplies", "curio_followup_delivery":
		return "delivery"
	case "investigate_anomaly":
		return "investigation"
	case "kill_dungeon_elite", "clear_dungeon":
		return "dungeon"
	default:
		return "counter"
	}
}

func (s *Service) EnsureDailyQuestBoard(character characters.Summary) BoardView {
	s.mu.Lock()
	defer s.mu.Unlock()

	board := s.ensureBoardLocked(character)
	return buildBoardView(board, characters.BuildDailyLimits(s.nextReset(), 0, 0, 0))
}

func (s *Service) ListQuests(character characters.Summary, limits characters.DailyLimits) BoardView {
	s.mu.Lock()
	defer s.mu.Unlock()

	board := s.ensureBoardLocked(character)
	return buildBoardView(board, limits)
}

func (s *Service) ActiveObjectives(characterID string) []characters.QuestSummary {
	s.mu.Lock()
	defer s.mu.Unlock()

	board, ok := s.boardByCharacter[characterID]
	if !ok {
		return []characters.QuestSummary{}
	}

	objectives := make([]characters.QuestSummary, 0, len(board.quests))
	for _, quest := range board.quests {
		if quest.Status == "accepted" || quest.Status == "completed" {
			objectives = append(objectives, enrichQuestSummary(quest))
		}
	}

	return objectives
}

func (s *Service) ListRuntimeActions(characterID string) []characters.ValidAction {
	s.mu.Lock()
	defer s.mu.Unlock()

	board, ok := s.boardByCharacter[characterID]
	if !ok {
		return nil
	}

	actions := make([]characters.ValidAction, 0, 4)
	for _, quest := range board.quests {
		if quest.Status != "accepted" {
			continue
		}
		quest = enrichQuestSummary(quest)
		runtime := runtimeForQuest(quest)
		if state := s.questRuntimeStateLocked(characterID, quest.QuestID); state != nil {
			runtime = applyRuntimeState(runtime, state)
		}

		switch runtime.SuggestedActionType {
		case "quest_interact":
			argsSchema := map[string]any{
				"quest_id":    "string",
				"interaction": "string",
			}
			if interaction, ok := runtime.SuggestedActionArgs["interaction"]; ok {
				argsSchema["suggested_interaction"] = interaction
			}
			actions = append(actions, characters.ValidAction{
				ActionType: "quest_interact",
				Label:      runtimeActionLabel(runtime, quest.Title),
				ArgsSchema: argsSchema,
			})
		case "quest_choice":
			argsSchema := map[string]any{
				"quest_id":   "string",
				"choice_key": "string",
			}
			if options, ok := runtime.SuggestedActionArgs["choice_options"]; ok {
				argsSchema["choice_options"] = options
			}
			actions = append(actions, characters.ValidAction{
				ActionType: "quest_choice",
				Label:      runtimeActionLabel(runtime, quest.Title),
				ArgsSchema: argsSchema,
			})
		}
	}

	return actions
}

func (s *Service) AcceptQuest(character characters.Summary, questID string, limits characters.DailyLimits) (BoardView, characters.QuestSummary, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	board := s.ensureBoardLocked(character)
	index := findQuestIndex(board.quests, questID)
	if index < 0 {
		return BoardView{}, characters.QuestSummary{}, ErrQuestNotFound
	}

	quest := board.quests[index]
	if quest.Status == "accepted" || quest.Status == "completed" {
		return buildBoardView(board, limits), enrichQuestSummary(quest), nil
	}
	if quest.Status != "available" {
		return BoardView{}, characters.QuestSummary{}, ErrQuestInvalidState
	}

	quest.Status = "accepted"
	quest = enrichQuestSummary(quest)
	board.quests[index] = quest
	if err := s.saveBoardLocked(character.CharacterID, board); err != nil {
		return BoardView{}, characters.QuestSummary{}, err
	}
	s.boardByCharacter[character.CharacterID] = board

	return buildBoardView(board, limits), quest, nil
}

func (s *Service) SubmitQuest(character characters.Summary, questID string, limits characters.DailyLimits) (characters.QuestSummary, error) {
	quest, err := s.PrepareQuestSubmission(character, questID, limits)
	if err != nil {
		return characters.QuestSummary{}, err
	}
	return s.FinalizeQuestSubmission(character.CharacterID, quest.QuestID)
}

func (s *Service) PrepareQuestSubmission(character characters.Summary, questID string, limits characters.DailyLimits) (characters.QuestSummary, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	board := s.ensureBoardLocked(character)
	index := findQuestIndex(board.quests, questID)
	if index < 0 {
		return characters.QuestSummary{}, ErrQuestNotFound
	}

	quest := board.quests[index]
	if quest.Status != "completed" {
		return characters.QuestSummary{}, ErrQuestInvalidState
	}

	return enrichQuestSummary(quest), nil
}

func (s *Service) FinalizeQuestSubmission(characterID, questID string) (characters.QuestSummary, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	board, ok := s.boardByCharacter[characterID]
	if !ok {
		return characters.QuestSummary{}, ErrQuestNotFound
	}

	index := findQuestIndex(board.quests, questID)
	if index < 0 {
		return characters.QuestSummary{}, ErrQuestNotFound
	}

	quest := board.quests[index]
	if quest.Status != "completed" {
		return characters.QuestSummary{}, ErrQuestInvalidState
	}

	quest.Status = "submitted"
	quest = enrichQuestSummary(quest)
	board.quests[index] = quest
	if err := s.saveBoardLocked(characterID, board); err != nil {
		return characters.QuestSummary{}, err
	}
	s.boardByCharacter[characterID] = board

	return quest, nil
}

func (s *Service) GetQuestRuntime(characterID string, questID string) (characters.QuestSummary, QuestRuntime, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	board, ok := s.boardByCharacter[characterID]
	if !ok {
		return characters.QuestSummary{}, QuestRuntime{}, ErrQuestNotFound
	}

	index := findQuestIndex(board.quests, questID)
	if index < 0 {
		return characters.QuestSummary{}, QuestRuntime{}, ErrQuestNotFound
	}

	quest := enrichQuestSummary(board.quests[index])
	runtime := runtimeForQuest(quest)
	if state := s.questRuntimeStateLocked(characterID, quest.QuestID); state != nil {
		runtime = applyRuntimeState(runtime, state)
	}
	return quest, runtime, nil
}

func (s *Service) ApplyQuestChoice(character characters.Summary, questID string, choiceKey string) (characters.QuestSummary, QuestRuntime, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	board := s.ensureBoardLocked(character)
	index := findQuestIndex(board.quests, questID)
	if index < 0 {
		return characters.QuestSummary{}, QuestRuntime{}, ErrQuestNotFound
	}

	quest := enrichQuestSummary(board.quests[index])
	runtime := runtimeForQuest(quest)
	if state := s.questRuntimeStateLocked(character.CharacterID, quest.QuestID); state != nil {
		runtime = applyRuntimeState(runtime, state)
	}

	available := false
	for _, choice := range runtime.AvailableChoices {
		if choice.ChoiceKey == strings.TrimSpace(choiceKey) {
			available = true
			break
		}
	}
	if !available {
		return characters.QuestSummary{}, QuestRuntime{}, ErrQuestChoiceNotAvailable
	}

	spec := runtimeSpecForQuest(quest)
	state := s.ensureQuestRuntimeStateLocked(character.CharacterID, quest.QuestID)
	selectedChoiceKey := strings.TrimSpace(choiceKey)
	state.SelectedChoiceKey = selectedChoiceKey
	state.State["selected_choice_key"] = state.SelectedChoiceKey
	state.CompletedStepKeys = appendUniqueString(state.CompletedStepKeys, spec.ChoiceStepKey)
	choiceSpec := findChoiceSpec(spec, selectedChoiceKey)
	state.CurrentStepKey = spec.CompletionStepKey
	if choiceSpec != nil && strings.TrimSpace(choiceSpec.NextStepKey) != "" {
		state.CurrentStepKey = choiceSpec.NextStepKey
	}
	if choiceSpec != nil {
		state.State["selected_choice_label"] = choiceSpec.Label
		for key, value := range choiceSpec.State {
			state.State[key] = value
		}
		for _, clue := range choiceSpec.Clues {
			state.Clues = appendQuestClue(state.Clues, clue)
		}
	}

	if err := s.saveBoardLocked(character.CharacterID, board); err != nil {
		return characters.QuestSummary{}, QuestRuntime{}, err
	}
	return quest, applyRuntimeState(runtimeForQuest(quest), state), nil
}

func (s *Service) AdvanceQuestInteraction(character characters.Summary, questID string, interaction string) (characters.QuestSummary, QuestRuntime, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	board := s.ensureBoardLocked(character)
	index := findQuestIndex(board.quests, questID)
	if index < 0 {
		return characters.QuestSummary{}, QuestRuntime{}, ErrQuestNotFound
	}

	quest := enrichQuestSummary(board.quests[index])
	if quest.Status != "accepted" && quest.Status != "completed" {
		return characters.QuestSummary{}, QuestRuntime{}, ErrQuestInvalidState
	}

	step := strings.TrimSpace(interaction)
	if step == "" {
		return characters.QuestSummary{}, QuestRuntime{}, ErrQuestInteractionInvalid
	}
	runtime := runtimeForQuest(quest)
	if existing := s.questRuntimeStateLocked(character.CharacterID, quest.QuestID); existing != nil {
		runtime = applyRuntimeState(runtime, existing)
	}
	if runtime.CurrentStepKey != step || !isInteractiveStep(runtimeSpecForQuest(quest), step) {
		return characters.QuestSummary{}, QuestRuntime{}, ErrQuestInteractionInvalid
	}

	state := s.ensureQuestRuntimeStateLocked(character.CharacterID, quest.QuestID)
	state.CompletedStepKeys = appendUniqueString(state.CompletedStepKeys, step)
	state.State["last_interaction"] = step
	if interactionSpec := findInteractionSpec(runtimeSpecForQuest(quest), step); interactionSpec != nil {
		for _, clue := range interactionSpec.Clues {
			state.Clues = appendQuestClue(state.Clues, clue)
		}
		for key, value := range interactionSpec.State {
			state.State[key] = value
		}
		state.CurrentStepKey = interactionSpec.NextStepKey
	} else if step == "inspect_clue" {
		state.Clues = appendQuestClue(state.Clues, QuestClue{
			ClueKey: "inspected_runtime_clue",
			Label:   "A runtime interaction revealed an additional clue for this quest.",
		})
	}
	if quest.Status == "completed" {
		state.CurrentStepKey = "turn_in_quest"
	} else if strings.TrimSpace(state.CurrentStepKey) == "" {
		state.CurrentStepKey = nextInteractionStepKey(quest, step)
	}

	if err := s.saveBoardLocked(character.CharacterID, board); err != nil {
		return characters.QuestSummary{}, QuestRuntime{}, err
	}
	return quest, applyRuntimeState(runtimeForQuest(quest), state), nil
}

func (s *Service) ApplyQuestTrigger(character characters.Summary, trigger Trigger, limits characters.DailyLimits) (BoardView, []ProgressChange) {
	s.mu.Lock()
	defer s.mu.Unlock()

	board := s.ensureBoardLocked(character)
	changes := make([]ProgressChange, 0, 2)

	for index, quest := range board.quests {
		if quest.Status != "accepted" {
			continue
		}

		state := s.questRuntimeStateLocked(character.CharacterID, quest.QuestID)
		updatedQuest, change, changed := applyTriggerToQuest(quest, trigger, state)
		if !changed {
			continue
		}

		updatedQuest = enrichQuestSummary(updatedQuest)
		board.quests[index] = updatedQuest
		if change.Completed {
			runtimeState := s.ensureQuestRuntimeStateLocked(character.CharacterID, quest.QuestID)
			runtimeState.CompletedStepKeys = appendUniqueString(runtimeState.CompletedStepKeys, runtimeStepKeyForCompletion(updatedQuest))
			runtimeState.CurrentStepKey = "turn_in_quest"
			runtimeState.State["auto_completed_by_trigger"] = trigger.TriggerType
		}
		change.Quest = updatedQuest
		changes = append(changes, change)
	}

	_ = s.saveBoardLocked(character.CharacterID, board)
	s.boardByCharacter[character.CharacterID] = board
	return buildBoardView(board, limits), changes
}

func (s *Service) RerollQuestBoard(character characters.Summary, limits characters.DailyLimits, confirmCost bool) (BoardView, error) {
	return BoardView{}, ErrQuestInvalidState
}

func (s *Service) ProgressTravelQuests(character characters.Summary, targetRegionID string, limits characters.DailyLimits) (BoardView, []characters.QuestSummary) {
	board, changes := s.ApplyQuestTrigger(character, Trigger{
		TriggerType: "travel_completed",
		RegionID:    targetRegionID,
	}, limits)
	return board, completedQuestsFromChanges(changes)
}

func (s *Service) EnsureCurioFollowupQuest(character characters.Summary, seed world.CurioQuestSeed, limits characters.DailyLimits) (BoardView, *characters.QuestSummary, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	board := s.ensureBoardLocked(character)
	if len(board.quests) >= characters.DailyQuestBoardSize {
		view := buildBoardView(board, limits)
		return view, nil, nil
	}

	for _, quest := range board.quests {
		if quest.TemplateType != seed.TemplateType {
			continue
		}
		if quest.TargetRegionID != seed.TargetRegionID {
			continue
		}
		if quest.Title != seed.Title {
			continue
		}
		if quest.Status == "accepted" || quest.Status == "completed" || quest.Status == "submitted" {
			view := buildBoardView(board, limits)
			return view, nil, nil
		}
	}

	activeCount := 0
	for _, quest := range board.quests {
		if quest.Status == "accepted" || quest.Status == "completed" {
			activeCount++
		}
	}
	if activeCount >= characters.DailyQuestBoardSize {
		view := buildBoardView(board, limits)
		return view, nil, nil
	}

	quest := characters.QuestSummary{
		QuestID:          nextID("quest"),
		BoardID:          board.boardID,
		TemplateType:     seed.TemplateType,
		Difficulty:       "hard",
		FlowKind:         "delivery",
		Rarity:           "rare",
		Status:           "accepted",
		Title:            seed.Title,
		Description:      seed.Description,
		TargetRegionID:   seed.TargetRegionID,
		ProgressCurrent:  0,
		ProgressTarget:   1,
		RewardGold:       seed.RewardGold,
		RewardReputation: seed.RewardReputation,
	}
	board.quests = append([]characters.QuestSummary{quest}, board.quests...)

	if err := s.saveBoardLocked(character.CharacterID, board); err != nil {
		return BoardView{}, nil, err
	}

	s.boardByCharacter[character.CharacterID] = board
	view := buildBoardView(board, limits)
	return view, &quest, nil
}

func (s *Service) ProgressDungeonQuests(character characters.Summary, dungeonRegionID string, limits characters.DailyLimits) (BoardView, []characters.QuestSummary) {
	board, changes := s.ApplyQuestTrigger(character, Trigger{
		TriggerType: "dungeon_cleared",
		RegionID:    dungeonRegionID,
	}, limits)
	return board, completedQuestsFromChanges(changes)
}

func (s *Service) ProgressFieldQuests(character characters.Summary, regionID string, enemiesDefeated, materialsCollected int, limits characters.DailyLimits) (BoardView, []characters.QuestSummary) {
	board, changes := s.ApplyQuestTrigger(character, Trigger{
		TriggerType:        "field_resolved",
		RegionID:           regionID,
		EnemiesDefeated:    enemiesDefeated,
		MaterialsCollected: materialsCollected,
	}, limits)
	return board, completedQuestsFromChanges(changes)
}

func buildBoardView(board boardRecord, limits characters.DailyLimits) BoardView {
	quests := make([]characters.QuestSummary, len(board.quests))
	copy(quests, board.quests)

	activeCount := 0
	for _, quest := range quests {
		if quest.Status == "accepted" || quest.Status == "completed" {
			activeCount++
		}
	}

	return BoardView{
		BoardID:          board.boardID,
		Status:           board.status,
		RerollCount:      board.rerollCount,
		ActiveQuestCount: activeCount,
		Quests:           enrichQuestSummaries(quests),
		Limits:           limits,
	}
}

func enrichQuestSummaries(quests []characters.QuestSummary) []characters.QuestSummary {
	items := make([]characters.QuestSummary, len(quests))
	for index, quest := range quests {
		items[index] = enrichQuestSummary(quest)
	}
	return items
}

func (s *Service) ensureBoardLocked(character characters.Summary) boardRecord {
	businessDate := s.businessDate()
	if board, ok := s.boardByCharacter[character.CharacterID]; ok && board.resetDate == businessDate {
		return board
	}

	board, exists := s.boardByCharacter[character.CharacterID]
	if !exists {
		board = boardRecord{
			boardID:     nextID("board"),
			status:      "active",
			rerollCount: 0,
			quests:      []characters.QuestSummary{},
		}
	}

	carried := carryOverDailyQuests(board.quests)
	deficit := maxInt(0, characters.DailyQuestBoardSize-len(carried))
	if deficit > 0 {
		carried = append(carried, generateQuestTemplates(character, board.boardID, businessDate, deficit, carried)...)
	}

	board.resetDate = businessDate
	board.quests = carried

	if _, ok := s.runtimeByQuest[character.CharacterID]; !ok {
		s.runtimeByQuest[character.CharacterID] = make(map[string]*questRuntimeState)
	}
	_ = s.saveBoardLocked(character.CharacterID, board)
	s.boardByCharacter[character.CharacterID] = board
	return board
}

func (s *Service) saveBoardLocked(characterID string, board boardRecord) error {
	if s.repo == nil {
		return nil
	}

	quests := make([]characters.QuestSummary, len(board.quests))
	copy(quests, board.quests)

	return s.repo.SaveBoard(StoredBoard{
		CharacterID:    characterID,
		BoardID:        board.boardID,
		ResetDate:      board.resetDate,
		Status:         board.status,
		RerollCount:    board.rerollCount,
		Quests:         quests,
		RuntimeByQuest: s.snapshotRuntimeByQuestLocked(characterID, quests),
	})
}

func (s *Service) businessDate() string {
	now := s.clock().In(s.loc)
	if now.Hour() < 4 {
		now = now.Add(-24 * time.Hour)
	}

	return now.Format("2006-01-02")
}

func (s *Service) nextReset() time.Time {
	now := s.clock().In(s.loc)
	resetToday := time.Date(now.Year(), now.Month(), now.Day(), 4, 0, 0, 0, now.Location())
	if now.Before(resetToday) {
		return resetToday
	}

	return resetToday.Add(24 * time.Hour)
}

func dailyQuestDefinitions() []questTemplateDefinition {
	catalog := mustQuestTemplateCatalog()
	items := make([]questTemplateDefinition, len(catalog.daily))
	copy(items, catalog.daily)
	return items
}

func supplementalQuestDefinitions() []questTemplateDefinition {
	catalog := mustQuestTemplateCatalog()
	items := make([]questTemplateDefinition, len(catalog.supplemental))
	copy(items, catalog.supplemental)
	return items
}

func instantiateQuest(definition questTemplateDefinition, _ characters.Summary, boardID string, seed string) characters.QuestSummary {
	rewardGold, rewardReputation := rollQuestRewards(definition, seed)
	contractType := "normal"
	progressTarget := definition.ProgressTarget
	if isBountyContract(seed) {
		contractType = "bounty"
		rewardGold *= 2
		rewardReputation *= 2
		if definition.FlowKind == "counter" {
			progressTarget += maxInt(1, progressTarget/2)
		}
	}

	return characters.QuestSummary{
		QuestID:          nextID("quest"),
		BoardID:          boardID,
		TemplateType:     definition.TemplateType,
		ContractType:     contractType,
		Difficulty:       definition.Difficulty,
		FlowKind:         definition.FlowKind,
		Rarity:           definition.Rarity,
		Status:           "accepted",
		Title:            definition.Title,
		Description:      definition.Description,
		TargetRegionID:   definition.TargetRegionID,
		ProgressCurrent:  0,
		ProgressTarget:   progressTarget,
		RewardGold:       rewardGold,
		RewardReputation: rewardReputation,
	}
}

func findQuestDefinition(quest characters.QuestSummary) (questTemplateDefinition, bool) {
	templateType := strings.TrimSpace(quest.TemplateType)
	difficulty := strings.TrimSpace(quest.Difficulty)
	flowKind := strings.TrimSpace(quest.FlowKind)
	if difficulty == "" {
		difficulty = difficultyFromRarity(quest.Rarity)
	}
	if flowKind == "" {
		flowKind = defaultFlowKindForTemplate(templateType)
	}

	for _, definition := range allQuestDefinitions() {
		if definition.TemplateType != templateType {
			continue
		}
		if definition.Difficulty != "" && definition.Difficulty != difficulty {
			continue
		}
		if definition.FlowKind != "" && definition.FlowKind != flowKind {
			continue
		}
		return definition, true
	}
	return questTemplateDefinition{}, false
}

func allQuestDefinitions() []questTemplateDefinition {
	catalog := mustQuestTemplateCatalog()
	items := make([]questTemplateDefinition, len(catalog.all))
	copy(items, catalog.all)
	return items
}

func generateQuestTemplates(character characters.Summary, boardID, dayKey string, count int, existing []characters.QuestSummary) []characters.QuestSummary {
	definitions := dailyQuestDefinitions()
	if count <= 0 || len(definitions) == 0 {
		return []characters.QuestSummary{}
	}

	used := make(map[string]struct{}, len(existing))
	for _, quest := range existing {
		used[quest.TemplateType] = struct{}{}
	}

	available := make([]questTemplateDefinition, 0, len(definitions))
	for _, definition := range definitions {
		if _, ok := used[definition.TemplateType]; ok {
			continue
		}
		available = append(available, definition)
	}
	if len(available) == 0 {
		available = definitions
	}

	quests := make([]characters.QuestSummary, 0, minInt(count, len(available)))
	baseSeed := fmt.Sprintf("%s|%s|%s", character.CharacterID, boardID, dayKey)
	for slot := 0; slot < count && len(available) > 0; slot++ {
		index := deterministicIndex(fmt.Sprintf("%s|template|%d", baseSeed, slot), len(available))
		definition := available[index]
		available = append(available[:index], available[index+1:]...)
		quests = append(quests, instantiateQuest(definition, character, boardID, fmt.Sprintf("%s|quest|%d|%s", baseSeed, slot, definition.TemplateType)))
	}

	return quests
}

func carryOverDailyQuests(quests []characters.QuestSummary) []characters.QuestSummary {
	active := make([]characters.QuestSummary, 0, len(quests))
	for _, quest := range quests {
		switch quest.Status {
		case "accepted", "completed", "available":
			if quest.Status == "available" {
				quest.Status = "accepted"
			}
			active = append(active, quest)
		case "submitted":
			continue
		default:
			continue
		}
	}

	return active
}

type questRewardRange struct {
	GoldMin       int
	GoldMax       int
	ReputationMin int
	ReputationMax int
}

func rewardRangeForTemplate(definition questTemplateDefinition) questRewardRange {
	switch strings.TrimSpace(definition.TemplateType) {
	case "kill_region_enemies":
		return questRewardRange{GoldMin: 90, GoldMax: 130, ReputationMin: 16, ReputationMax: 20}
	case "collect_materials":
		return questRewardRange{GoldMin: 80, GoldMax: 120, ReputationMin: 14, ReputationMax: 18}
	case "deliver_supplies":
		if strings.EqualFold(strings.TrimSpace(definition.Difficulty), "hard") {
			return questRewardRange{GoldMin: 110, GoldMax: 150, ReputationMin: 20, ReputationMax: 24}
		}
		return questRewardRange{GoldMin: 85, GoldMax: 125, ReputationMin: 15, ReputationMax: 19}
	case "investigate_anomaly":
		return questRewardRange{GoldMin: 125, GoldMax: 170, ReputationMin: 22, ReputationMax: 28}
	case "clear_dungeon":
		return questRewardRange{GoldMin: 155, GoldMax: 210, ReputationMin: 28, ReputationMax: 36}
	default:
		return questRewardRange{GoldMin: 90, GoldMax: 130, ReputationMin: 16, ReputationMax: 20}
	}
}

func rollQuestRewards(definition questTemplateDefinition, seed string) (int, int) {
	rewardRange := rewardRangeForTemplate(definition)
	gold := rewardRange.GoldMin + deterministicIndex(seed+"|gold", rewardRange.GoldMax-rewardRange.GoldMin+1)
	reputation := rewardRange.ReputationMin + deterministicIndex(seed+"|reputation", rewardRange.ReputationMax-rewardRange.ReputationMin+1)
	return gold, reputation
}

func isBountyContract(seed string) bool {
	return deterministicUnitFloat(seed+"|bounty") < 0.15
}

func deterministicUnitFloat(seed string) float64 {
	hasher := fnv.New32a()
	_, _ = hasher.Write([]byte(seed))
	value := hasher.Sum32()
	return float64(value%1000000) / 1000000.0
}

func deterministicIndex(seed string, size int) int {
	if size <= 1 {
		return 0
	}
	hasher := fnv.New32a()
	_, _ = hasher.Write([]byte(seed))
	return int(hasher.Sum32() % uint32(size))
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func findQuestIndex(quests []characters.QuestSummary, questID string) int {
	for index, quest := range quests {
		if quest.QuestID == strings.TrimSpace(questID) {
			return index
		}
	}

	return -1
}

func runtimeForQuest(quest characters.QuestSummary) QuestRuntime {
	quest = enrichQuestSummary(quest)
	spec := runtimeSpecForQuest(quest)

	runtime := QuestRuntime{
		QuestID:           quest.QuestID,
		CompletedStepKeys: []string{},
		AvailableChoices:  []QuestChoice{},
		Clues:             append([]QuestClue{}, spec.BaseClues...),
		State: map[string]any{
			"template_type": quest.TemplateType,
			"status":        quest.Status,
			"difficulty":    quest.Difficulty,
			"flow_kind":     quest.FlowKind,
			"rarity":        quest.Rarity,
			"progress":      map[string]any{"current": quest.ProgressCurrent, "target": quest.ProgressTarget},
			"target_hint":   quest.TargetRegionID,
			"step_keys":     questRuntimeStepKeys(spec.Steps),
			"choice_defs":   encodeQuestChoices(choiceSpecsToChoices(spec.ChoiceSpecs)),
		},
	}
	runtime.CurrentStepKey = spec.InitialStepKey

	if quest.Status == "submitted" {
		for _, step := range spec.Steps {
			runtime.CompletedStepKeys = appendUniqueString(runtime.CompletedStepKeys, step.Key)
		}
		runtime.CurrentStepKey = ""
		runtime.AvailableChoices = []QuestChoice{}
	} else if quest.Status == "completed" {
		if spec.RequiresInspection {
			runtime.CompletedStepKeys = appendUniqueString(runtime.CompletedStepKeys, spec.InspectStepKey)
		}
		if spec.RequiresChoice {
			runtime.CompletedStepKeys = appendUniqueString(runtime.CompletedStepKeys, spec.ChoiceStepKey)
		}
		if spec.RequiresRouteConfirm {
			runtime.CompletedStepKeys = appendUniqueString(runtime.CompletedStepKeys, "confirm_route")
		}
		runtime.CompletedStepKeys = appendUniqueString(runtime.CompletedStepKeys, spec.CompletionStepKey)
		runtime.CurrentStepKey = "turn_in_quest"
		runtime.AvailableChoices = []QuestChoice{}
	}
	return finalizeRuntime(runtime, quest, spec)
}

func applyTriggerToQuest(quest characters.QuestSummary, trigger Trigger, state *questRuntimeState) (characters.QuestSummary, ProgressChange, bool) {
	quest = enrichQuestSummary(quest)
	spec := runtimeSpecForQuest(quest)
	change := ProgressChange{
		Quest:          quest,
		PreviousStatus: quest.Status,
		CurrentStatus:  quest.Status,
	}

	if !canApplyTrigger(spec, quest, trigger, state) {
		return quest, change, false
	}
	change.ProgressDelta = progressDeltaForTrigger(spec, quest, trigger)
	if change.ProgressDelta <= 0 {
		return quest, change, false
	}
	quest.ProgressCurrent += change.ProgressDelta
	if quest.ProgressCurrent >= quest.ProgressTarget {
		quest.ProgressCurrent = quest.ProgressTarget
		quest.Status = "completed"
	}

	change.CurrentStatus = quest.Status
	change.Completed = change.PreviousStatus != "completed" && quest.Status == "completed"
	return quest, change, true
}

func completedQuestsFromChanges(changes []ProgressChange) []characters.QuestSummary {
	completed := make([]characters.QuestSummary, 0, len(changes))
	for _, change := range changes {
		if change.Completed {
			completed = append(completed, change.Quest)
		}
	}
	return completed
}

func (s *Service) questRuntimeStateLocked(characterID, questID string) *questRuntimeState {
	byCharacter, ok := s.runtimeByQuest[characterID]
	if !ok {
		return nil
	}
	return byCharacter[questID]
}

func (s *Service) ensureQuestRuntimeStateLocked(characterID, questID string) *questRuntimeState {
	byCharacter, ok := s.runtimeByQuest[characterID]
	if !ok {
		byCharacter = make(map[string]*questRuntimeState)
		s.runtimeByQuest[characterID] = byCharacter
	}
	state, ok := byCharacter[questID]
	if !ok {
		state = &questRuntimeState{
			CompletedStepKeys: []string{},
			Clues:             []QuestClue{},
			State:             map[string]any{},
		}
		byCharacter[questID] = state
	}
	if state.State == nil {
		state.State = map[string]any{}
	}
	return state
}

func (s *Service) snapshotRuntimeByQuestLocked(characterID string, quests []characters.QuestSummary) map[string]StoredQuestRuntime {
	byCharacter, ok := s.runtimeByQuest[characterID]
	if !ok || len(byCharacter) == 0 {
		return nil
	}

	result := make(map[string]StoredQuestRuntime)
	for _, quest := range quests {
		state, ok := byCharacter[quest.QuestID]
		if !ok || state == nil {
			continue
		}
		clues := make([]QuestClue, len(state.Clues))
		copy(clues, state.Clues)
		completed := make([]string, len(state.CompletedStepKeys))
		copy(completed, state.CompletedStepKeys)
		payload := make(map[string]any, len(state.State))
		for key, value := range state.State {
			payload[key] = value
		}
		result[quest.QuestID] = StoredQuestRuntime{
			SelectedChoiceKey: state.SelectedChoiceKey,
			CompletedStepKeys: completed,
			CurrentStepKey:    state.CurrentStepKey,
			Clues:             clues,
			State:             payload,
		}
	}

	if len(result) == 0 {
		return nil
	}
	return result
}

func applyRuntimeState(runtime QuestRuntime, state *questRuntimeState) QuestRuntime {
	spec := runtimeSpecForQuest(characters.QuestSummary{
		QuestID:      runtime.QuestID,
		TemplateType: asString(runtime.State["template_type"]),
		Difficulty:   asString(runtime.State["difficulty"]),
		FlowKind:     asString(runtime.State["flow_kind"]),
		Rarity:       asString(runtime.State["rarity"]),
		Status:       asString(runtime.State["status"]),
	})
	if state == nil {
		return finalizeRuntime(runtime, characters.QuestSummary{
			QuestID:      runtime.QuestID,
			TemplateType: asString(runtime.State["template_type"]),
			Difficulty:   asString(runtime.State["difficulty"]),
			FlowKind:     asString(runtime.State["flow_kind"]),
			Rarity:       asString(runtime.State["rarity"]),
			Status:       asString(runtime.State["status"]),
		}, spec)
	}
	if state.CurrentStepKey != "" {
		runtime.CurrentStepKey = state.CurrentStepKey
	}
	for _, step := range state.CompletedStepKeys {
		runtime.CompletedStepKeys = appendUniqueString(runtime.CompletedStepKeys, step)
	}
	for _, clue := range state.Clues {
		runtime.Clues = appendQuestClue(runtime.Clues, clue)
	}
	if runtime.State == nil {
		runtime.State = map[string]any{}
	}
	for key, value := range state.State {
		runtime.State[key] = value
	}
	if state.SelectedChoiceKey != "" {
		runtime.State["selected_choice_key"] = state.SelectedChoiceKey
	}
	quest := characters.QuestSummary{
		QuestID:      runtime.QuestID,
		TemplateType: asString(runtime.State["template_type"]),
		Difficulty:   asString(runtime.State["difficulty"]),
		FlowKind:     asString(runtime.State["flow_kind"]),
		Rarity:       asString(runtime.State["rarity"]),
		Status:       asString(runtime.State["status"]),
	}
	return finalizeRuntime(runtime, quest, spec)
}

func appendUniqueString(values []string, value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return values
	}
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func appendQuestClue(values []QuestClue, clue QuestClue) []QuestClue {
	for _, existing := range values {
		if existing.ClueKey == clue.ClueKey {
			return values
		}
	}
	return append(values, clue)
}

func runtimeStateHasStep(state *questRuntimeState, stepKey string) bool {
	if state == nil {
		return false
	}
	for _, step := range state.CompletedStepKeys {
		if step == stepKey {
			return true
		}
	}
	return false
}

func finalizeRuntime(runtime QuestRuntime, quest characters.QuestSummary, spec questRuntimeSpec) QuestRuntime {
	stepSpec := findStepSpec(spec, runtime.CurrentStepKey)
	runtime.CurrentStepLabel = stepSpec.Label
	runtime.CurrentStepHint = stepSpec.Hint
	runtime.SuggestedActionType = stepSpec.ActionType
	runtime.SuggestedActionArgs = suggestedActionArgsForStep(quest, runtime.CurrentStepKey, stepSpec.ActionType)

	if strings.TrimSpace(asString(runtime.State["selected_choice_key"])) != "" {
		runtime.AvailableChoices = []QuestChoice{}
	} else if runtime.CurrentStepKey == spec.ChoiceStepKey {
		runtime.AvailableChoices = choiceSpecsToChoices(spec.ChoiceSpecs)
	} else {
		runtime.AvailableChoices = []QuestChoice{}
	}
	return runtime
}

func suggestedActionArgsForStep(quest characters.QuestSummary, stepKey, actionType string) map[string]any {
	if strings.TrimSpace(actionType) == "" {
		return nil
	}
	args := map[string]any{
		"quest_id": quest.QuestID,
	}
	switch strings.TrimSpace(actionType) {
	case "quest_interact":
		args["interaction"] = stepKey
	case "quest_choice":
		args["choice_options"] = encodeQuestChoices(choiceSpecsToChoices(defaultChoiceSpecsForQuest(quest)))
	}
	return args
}

func runtimeActionLabel(runtime QuestRuntime, questTitle string) string {
	if strings.TrimSpace(runtime.CurrentStepLabel) != "" {
		return fmt.Sprintf("%s for %s", runtime.CurrentStepLabel, questTitle)
	}
	switch runtime.SuggestedActionType {
	case "quest_choice":
		return fmt.Sprintf("Choose next step for %s", questTitle)
	case "quest_interact":
		return fmt.Sprintf("Advance quest step for %s", questTitle)
	default:
		return questTitle
	}
}

func findStepSpec(spec questRuntimeSpec, stepKey string) questStepSpec {
	for _, step := range spec.Steps {
		if step.Key == strings.TrimSpace(stepKey) {
			return step
		}
	}
	return questStepSpec{Key: stepKey}
}

func findChoiceSpec(spec questRuntimeSpec, choiceKey string) *questChoiceSpec {
	for _, choice := range spec.ChoiceSpecs {
		if choice.ChoiceKey == strings.TrimSpace(choiceKey) {
			item := choice
			return &item
		}
	}
	return nil
}

func choiceSpecsToChoices(specs []questChoiceSpec) []QuestChoice {
	choices := make([]QuestChoice, 0, len(specs))
	for _, choice := range specs {
		choices = append(choices, QuestChoice{
			ChoiceKey: choice.ChoiceKey,
			Label:     choice.Label,
		})
	}
	return choices
}

func findInteractionSpec(spec questRuntimeSpec, stepKey string) *questInteractionSpec {
	for _, interaction := range spec.InteractionSpecs {
		if interaction.StepKey == strings.TrimSpace(stepKey) {
			item := interaction
			return &item
		}
	}
	return nil
}

func defaultChoiceSpecsForQuest(quest characters.QuestSummary) []questChoiceSpec {
	return runtimeSpecForQuest(quest).ChoiceSpecs
}

func defaultChoicesForRuntime(state map[string]any) []QuestChoice {
	rawChoices, ok := state["choice_defs"]
	if !ok {
		return nil
	}
	switch typed := rawChoices.(type) {
	case []map[string]string:
		choices := make([]QuestChoice, 0, len(typed))
		for _, item := range typed {
			choices = append(choices, QuestChoice{
				ChoiceKey: item["choice_key"],
				Label:     item["label"],
			})
		}
		return choices
	case []any:
		choices := make([]QuestChoice, 0, len(typed))
		for _, item := range typed {
			payload, ok := item.(map[string]any)
			if !ok {
				continue
			}
			choices = append(choices, QuestChoice{
				ChoiceKey: asString(payload["choice_key"]),
				Label:     asString(payload["label"]),
			})
		}
		return choices
	}
	return nil
}

func nextInteractionStepKey(quest characters.QuestSummary, interaction string) string {
	spec := runtimeSpecForQuest(quest)
	if interactionSpec := findInteractionSpec(spec, interaction); interactionSpec != nil && strings.TrimSpace(interactionSpec.NextStepKey) != "" {
		return interactionSpec.NextStepKey
	}
	switch strings.TrimSpace(interaction) {
	case "inspect_clue", "confirm_route":
		return spec.CompletionStepKey
	default:
		return spec.CompletionStepKey
	}
}

func runtimeStepKeyForCompletion(quest characters.QuestSummary) string {
	return runtimeSpecForQuest(quest).CompletionStepKey
}

func isInteractiveStep(spec questRuntimeSpec, step string) bool {
	step = strings.TrimSpace(step)
	if step == "" {
		return false
	}
	if step == spec.InspectStepKey || step == "confirm_route" {
		return true
	}
	return findInteractionSpec(spec, step) != nil
}

func canApplyTrigger(spec questRuntimeSpec, quest characters.QuestSummary, trigger Trigger, state *questRuntimeState) bool {
	if strings.TrimSpace(spec.ProgressTriggerType) != strings.TrimSpace(trigger.TriggerType) {
		return false
	}
	if strings.TrimSpace(quest.TargetRegionID) != "" && quest.TargetRegionID != trigger.RegionID {
		return false
	}
	if spec.RequiresRouteConfirm && !runtimeStateHasStep(state, "confirm_route") {
		return false
	}
	if spec.RequiresInspection && !runtimeStateHasStep(state, spec.InspectStepKey) {
		return false
	}
	if spec.RequiresChoice && (state == nil || strings.TrimSpace(state.SelectedChoiceKey) == "") {
		return false
	}
	return true
}

func progressDeltaForTrigger(spec questRuntimeSpec, quest characters.QuestSummary, trigger Trigger) int {
	switch strings.TrimSpace(spec.ProgressSource) {
	case "enemies_defeated":
		return trigger.EnemiesDefeated
	case "materials_collected":
		return trigger.MaterialsCollected
	default:
		return maxInt(0, quest.ProgressTarget-quest.ProgressCurrent)
	}
}

func questRuntimeStepKeys(steps []questStepSpec) []string {
	keys := make([]string, 0, len(steps))
	for _, step := range steps {
		if strings.TrimSpace(step.Key) == "" {
			continue
		}
		keys = append(keys, step.Key)
	}
	return keys
}

func encodeQuestChoices(choices []QuestChoice) []map[string]string {
	if len(choices) == 0 {
		return nil
	}
	encoded := make([]map[string]string, 0, len(choices))
	for _, choice := range choices {
		encoded = append(encoded, map[string]string{
			"choice_key": choice.ChoiceKey,
			"label":      choice.Label,
		})
	}
	return encoded
}

func asString(value any) string {
	text, _ := value.(string)
	return text
}

func mustLocation(name string) *time.Location {
	location, err := time.LoadLocation(name)
	if err != nil {
		panic(err)
	}

	return location
}

func nextID(prefix string) string {
	return fmt.Sprintf("%s_%d_%06d", prefix, time.Now().UnixNano(), atomic.AddUint64(&questIDCounter, 1))
}
