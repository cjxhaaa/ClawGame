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
)

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

type StoredBoard struct {
	CharacterID string
	BoardID     string
	ResetDate   string
	Status      string
	RerollCount int
	Quests      []characters.QuestSummary
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
	service := &Service{
		clock:            time.Now,
		loc:              mustLocation(businessTimezone),
		repo:             repo,
		boardByCharacter: make(map[string]boardRecord),
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
	}

	return service, nil
}

func RerollCostGold() int {
	return rerollCostGold
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
			objectives = append(objectives, quest)
		}
	}

	return objectives
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
	board.quests[index] = quest
	if err := s.saveBoardLocked(character.CharacterID, board); err != nil {
		return characters.QuestSummary{}, err
	}
	s.boardByCharacter[character.CharacterID] = board

	return quest, nil
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
	board.quests = mergeActiveBoard(kept, replacements)
	if err := s.saveBoardLocked(character.CharacterID, board); err != nil {
		return BoardView{}, err
	}
	s.boardByCharacter[character.CharacterID] = board

	return buildBoardView(board, limits), nil
}

func (s *Service) ProgressTravelQuests(character characters.Summary, targetRegionID string, limits characters.DailyLimits) (BoardView, []characters.QuestSummary) {
	s.mu.Lock()
	defer s.mu.Unlock()

	board := s.ensureBoardLocked(character)
	completed := make([]characters.QuestSummary, 0, 2)

	for index, quest := range board.quests {
		if quest.Status != "accepted" {
			continue
		}
		if quest.TemplateType != "deliver_supplies" {
			continue
		}
		if quest.TargetRegionID != targetRegionID {
			continue
		}

		quest.ProgressCurrent = quest.ProgressTarget
		quest.Status = "completed"
		board.quests[index] = quest
		completed = append(completed, quest)
	}

	_ = s.saveBoardLocked(character.CharacterID, board)
	s.boardByCharacter[character.CharacterID] = board
	return buildBoardView(board, limits), completed
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
		Quests:           quests,
		Limits:           limits,
	}
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
		CharacterID: characterID,
		BoardID:     board.boardID,
		ResetDate:   board.resetDate,
		Status:      board.status,
		RerollCount: board.rerollCount,
		Quests:      quests,
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

func generateQuestTemplates(character characters.Summary, seed string, rerollCount int) []characters.QuestSummary {
	_ = seed
	lowRank := []characters.QuestSummary{
		{
			QuestID:          nextID("quest"),
			TemplateType:     "kill_region_enemies",
			Rarity:           "common",
			Status:           "available",
			Title:            "Clear 6 Forest Enemies",
			Description:      "Defeat 6 enemies in Whispering Forest.",
			TargetRegionID:   "whispering_forest",
			ProgressCurrent:  0,
			ProgressTarget:   6,
			RewardGold:       120,
			RewardReputation: 24,
		},
		{
			QuestID:          nextID("quest"),
			TemplateType:     "collect_materials",
			Rarity:           "common",
			Status:           "available",
			Title:            "Gather 5 Forest Reagents",
			Description:      "Collect 5 reagents from Whispering Forest encounter routes.",
			TargetRegionID:   "whispering_forest",
			ProgressCurrent:  0,
			ProgressTarget:   5,
			RewardGold:       105,
			RewardReputation: 20,
		},
		{
			QuestID:          nextID("quest"),
			TemplateType:     "deliver_supplies",
			Rarity:           "common",
			Status:           "available",
			Title:            "Deliver Guild Supplies",
			Description:      "Carry guild supplies from Main City to Greenfield Village.",
			TargetRegionID:   "greenfield_village",
			ProgressCurrent:  0,
			ProgressTarget:   1,
			RewardGold:       95,
			RewardReputation: 18,
		},
		{
			QuestID:          nextID("quest"),
			TemplateType:     "deliver_supplies",
			Rarity:           "uncommon",
			Status:           "available",
			Title:            "Reinforce the Forest Outpost",
			Description:      "Deliver higher-value supplies to Whispering Forest.",
			TargetRegionID:   "whispering_forest",
			ProgressCurrent:  0,
			ProgressTarget:   1,
			RewardGold:       165,
			RewardReputation: 34,
		},
		{
			QuestID:          nextID("quest"),
			TemplateType:     "kill_dungeon_elite",
			Rarity:           "uncommon",
			Status:           "available",
			Title:            "Cull the Catacomb Elite",
			Description:      "Defeat the elite guarding the Ancient Catacomb corridors.",
			TargetRegionID:   "ancient_catacomb",
			ProgressCurrent:  0,
			ProgressTarget:   1,
			RewardGold:       210,
			RewardReputation: 42,
		},
		{
			QuestID:          nextID("quest"),
			TemplateType:     "clear_dungeon",
			Rarity:           "challenge",
			Status:           "available",
			Title:            "Cleanse Ancient Catacomb",
			Description:      "Clear Ancient Catacomb without being defeated.",
			TargetRegionID:   "ancient_catacomb",
			ProgressCurrent:  0,
			ProgressTarget:   1,
			RewardGold:       320,
			RewardReputation: 70,
		},
	}

	if character.Rank == "mid" || character.Rank == "high" {
		lowRank[0].TargetRegionID = "sunscar_desert_outskirts"
		lowRank[0].Description = "Defeat 6 enemies in Sunscar Desert Outskirts."
		lowRank[0].Title = "Clear 6 Desert Enemies"
		lowRank[0].RewardGold = 180
		lowRank[0].RewardReputation = 30

		lowRank[2].TargetRegionID = "sunscar_desert_outskirts"
		lowRank[2].Description = "Carry supplies from Main City to Sunscar Desert Outskirts."
		lowRank[2].Title = "Deliver Desert Provisions"
		lowRank[2].RewardGold = 145
		lowRank[2].RewardReputation = 26
	}

	if rerollCount%2 == 1 {
		slices.Reverse(lowRank[:3])
	}

	return lowRank
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
