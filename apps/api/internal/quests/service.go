package quests

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"clawgame/apps/api/internal/characters"
	"clawgame/apps/api/internal/world"
)

const (
	businessTimezone = "Asia/Shanghai"
	rerollCostGold   = 20
)

var (
	ErrQuestNotFound              = errors.New("quest not found")
	ErrQuestInvalidState          = errors.New("quest invalid state")
	ErrQuestCompletionCapReached  = errors.New("quest completion cap reached")
	ErrQuestRerollConfirmRequired = errors.New("quest reroll confirmation required")
	ErrQuestChoiceNotAvailable    = errors.New("quest choice not available")
	ErrQuestInteractionInvalid    = errors.New("quest interaction invalid")
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
	RankOverrides    []questTemplateRankOverride
}

type questTemplateRankOverride struct {
	Ranks            []string `yaml:"ranks"`
	Title            string   `yaml:"title"`
	Description      string   `yaml:"description"`
	TargetRegionID   string   `yaml:"target_region_id"`
	ProgressTarget   int      `yaml:"progress_target"`
	RewardGold       int      `yaml:"reward_gold"`
	RewardReputation int      `yaml:"reward_reputation"`
}

type BoardView struct {
	BoardID          string                    `json:"board_id"`
	Status           string                    `json:"status"`
	RerollCount      int                       `json:"reroll_count"`
	ActiveQuestCount int                       `json:"active_quest_count"`
	Quests           []characters.QuestSummary `json:"quests"`
	Limits           characters.DailyLimits    `json:"limits"`
}

type QuestRewardResult struct {
	Character    characters.Summary     `json:"character"`
	RankChanged  bool                   `json:"rank_changed"`
	PreviousRank string                 `json:"previous_rank"`
	CurrentRank  string                 `json:"current_rank"`
	Limits       characters.DailyLimits `json:"limits"`
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

func RerollCostGold() int {
	return rerollCostGold
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
	return buildBoardView(board, characters.LimitsForRank(character.Rank, s.nextReset(), 0, 0))
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
	if limits.QuestCompletionUsed >= limits.QuestCompletionCap {
		return characters.QuestSummary{}, ErrQuestCompletionCapReached
	}

	quest.Status = "submitted"
	quest = enrichQuestSummary(quest)
	board.quests[index] = quest
	if err := s.saveBoardLocked(character.CharacterID, board); err != nil {
		return characters.QuestSummary{}, err
	}
	s.boardByCharacter[character.CharacterID] = board

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
	if !confirmCost {
		return BoardView{}, ErrQuestRerollConfirmRequired
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	board := s.ensureBoardLocked(character)
	board.rerollCount++

	kept := make([]characters.QuestSummary, 0, len(board.quests))
	for _, quest := range board.quests {
		switch quest.Status {
		case "submitted", "completed":
			kept = append(kept, quest)
		default:
			quest.Status = "expired"
			kept = append(kept, quest)
		}
	}

	replacements := generateQuestTemplates(character, board.boardID, board.rerollCount)
	for index := range replacements {
		replacements[index].BoardID = board.boardID
	}
	board.quests = mergeActiveBoard(kept, replacements)
	if err := s.saveBoardLocked(character.CharacterID, board); err != nil {
		return BoardView{}, err
	}
	s.boardByCharacter[character.CharacterID] = board

	return buildBoardView(board, limits), nil
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
		if quest.Status == "accepted" || quest.Status == "completed" {
			view := buildBoardView(board, limits)
			return view, nil, nil
		}
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

	board := boardRecord{
		boardID:     nextID("board"),
		resetDate:   businessDate,
		status:      "active",
		rerollCount: 0,
		quests:      generateQuestTemplates(character, nextID("boardseed"), 0),
	}

	for index := range board.quests {
		board.quests[index].BoardID = board.boardID
	}

	s.runtimeByQuest[character.CharacterID] = make(map[string]*questRuntimeState)
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

func instantiateQuest(definition questTemplateDefinition, character characters.Summary, boardID string) characters.QuestSummary {
	quest := characters.QuestSummary{
		QuestID:          nextID("quest"),
		BoardID:          boardID,
		TemplateType:     definition.TemplateType,
		Difficulty:       definition.Difficulty,
		FlowKind:         definition.FlowKind,
		Rarity:           definition.Rarity,
		Status:           "available",
		Title:            definition.Title,
		Description:      definition.Description,
		TargetRegionID:   definition.TargetRegionID,
		ProgressCurrent:  0,
		ProgressTarget:   definition.ProgressTarget,
		RewardGold:       definition.RewardGold,
		RewardReputation: definition.RewardReputation,
	}
	applyQuestRankOverrides(&quest, definition.RankOverrides, character.Rank)
	return quest
}

func applyQuestRankOverrides(quest *characters.QuestSummary, overrides []questTemplateRankOverride, rank string) {
	for _, override := range overrides {
		if !rankMatchesOverride(rank, override.Ranks) {
			continue
		}
		if strings.TrimSpace(override.Title) != "" {
			quest.Title = override.Title
		}
		if strings.TrimSpace(override.Description) != "" {
			quest.Description = override.Description
		}
		if strings.TrimSpace(override.TargetRegionID) != "" {
			quest.TargetRegionID = override.TargetRegionID
		}
		if override.ProgressTarget > 0 {
			quest.ProgressTarget = override.ProgressTarget
		}
		if override.RewardGold > 0 {
			quest.RewardGold = override.RewardGold
		}
		if override.RewardReputation > 0 {
			quest.RewardReputation = override.RewardReputation
		}
	}
}

func rankMatchesOverride(rank string, ranks []string) bool {
	if len(ranks) == 0 {
		return false
	}
	for _, candidate := range ranks {
		if strings.EqualFold(strings.TrimSpace(candidate), strings.TrimSpace(rank)) {
			return true
		}
	}
	return false
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

func generateQuestTemplates(character characters.Summary, seed string, rerollCount int) []characters.QuestSummary {
	_ = seed
	definitions := dailyQuestDefinitions()
	quests := make([]characters.QuestSummary, 0, len(definitions))
	for _, definition := range definitions {
		quests = append(quests, instantiateQuest(definition, character, ""))
	}

	if rerollCount%2 == 1 {
		slices.Reverse(quests[:3])
	}

	return quests
}

func mergeActiveBoard(previous []characters.QuestSummary, replacements []characters.QuestSummary) []characters.QuestSummary {
	active := make([]characters.QuestSummary, 0, len(previous)+len(replacements))
	for _, quest := range previous {
		active = append(active, quest)
	}
	for _, quest := range replacements {
		if quest.Status == "available" {
			active = append(active, quest)
		}
	}

	return active
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

func maxInt(left, right int) int {
	if left > right {
		return left
	}
	return right
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
