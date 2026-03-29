package arena

import (
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"clawgame/apps/api/internal/characters"
	"clawgame/apps/api/internal/combat"
	"clawgame/apps/api/internal/world"
)

var (
	ErrSignupClosed    = errors.New("arena signup closed")
	ErrRankNotEligible = errors.New("arena rank not eligible")
	ErrAlreadySignedUp = errors.New("arena already signed up")
)

type Entry struct {
	CharacterID    string `json:"character_id"`
	CharacterName  string `json:"character_name"`
	Rank           string `json:"rank"`
	EquipmentScore int    `json:"equipment_score"`
	SignedUpAt     string `json:"signed_up_at"`
}

type CurrentView struct {
	TournamentID  string            `json:"tournament_id"`
	WeekKey       string            `json:"week_key"`
	Status        world.ArenaStatus `json:"status"`
	SignupCount   int               `json:"signup_count"`
	Entries       []Entry           `json:"entries"`
	NextRoundTime string            `json:"next_round_time"`
}

type LeaderboardEntry struct {
	Rank        int    `json:"rank"`
	CharacterID string `json:"character_id"`
	Name        string `json:"name"`
	Score       int    `json:"score"`
	ScoreLabel  string `json:"score_label"`
}

type Service struct {
	mu            sync.Mutex
	clock         func() time.Time
	entriesByWeek map[string]map[string]Entry
}

func NewService() *Service {
	return &Service{
		clock:         time.Now,
		entriesByWeek: make(map[string]map[string]Entry),
	}
}

func (s *Service) Signup(character characters.Summary, equipmentScore int, arenaStatus world.ArenaStatus) (Entry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if character.Rank != "mid" && character.Rank != "high" {
		return Entry{}, ErrRankNotEligible
	}
	if arenaStatus.Code != "signup_open" {
		return Entry{}, ErrSignupClosed
	}

	weekKey := weekKeyFor(s.clock())
	if _, ok := s.entriesByWeek[weekKey]; !ok {
		s.entriesByWeek[weekKey] = make(map[string]Entry)
	}
	if _, exists := s.entriesByWeek[weekKey][character.CharacterID]; exists {
		return Entry{}, ErrAlreadySignedUp
	}

	entry := Entry{
		CharacterID:    character.CharacterID,
		CharacterName:  character.Name,
		Rank:           character.Rank,
		EquipmentScore: equipmentScore,
		SignedUpAt:     s.clock().Format(time.RFC3339),
	}
	s.entriesByWeek[weekKey][character.CharacterID] = entry
	return entry, nil
}

func (s *Service) GetCurrent(arenaStatus world.ArenaStatus) CurrentView {
	s.mu.Lock()
	defer s.mu.Unlock()

	weekKey := weekKeyFor(s.clock())
	entriesMap := s.entriesByWeek[weekKey]
	entries := make([]Entry, 0, len(entriesMap))
	for _, entry := range entriesMap {
		entries = append(entries, entry)
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].EquipmentScore != entries[j].EquipmentScore {
			return entries[i].EquipmentScore > entries[j].EquipmentScore
		}
		return entries[i].SignedUpAt < entries[j].SignedUpAt
	})

	nextRound := s.clock().Add(5 * time.Minute).Format(time.RFC3339)
	return CurrentView{
		TournamentID:  "tourn_" + weekKey,
		WeekKey:       weekKey,
		Status:        arenaStatus,
		SignupCount:   len(entries),
		Entries:       entries,
		NextRoundTime: nextRound,
	}
}

func (s *Service) GetLeaderboard() []LeaderboardEntry {
	s.mu.Lock()
	defer s.mu.Unlock()

	weekKey := weekKeyFor(s.clock())
	entriesMap := s.entriesByWeek[weekKey]
	entries := make([]Entry, 0, len(entriesMap))
	for _, entry := range entriesMap {
		entries = append(entries, entry)
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].EquipmentScore != entries[j].EquipmentScore {
			return entries[i].EquipmentScore > entries[j].EquipmentScore
		}
		return entries[i].CharacterName < entries[j].CharacterName
	})

	result := make([]LeaderboardEntry, 0, len(entries))
	for i, entry := range entries {
		result = append(result, LeaderboardEntry{
			Rank:        i + 1,
			CharacterID: entry.CharacterID,
			Name:        entry.CharacterName,
			Score:       entry.EquipmentScore,
			ScoreLabel:  "equipment_score",
		})
	}
	return result
}

func weekKeyFor(now time.Time) string {
	year, week := now.ISOWeek()
	return fmt.Sprintf("%04dW%02d", year, week)
}

// DuelResult is the outcome of a simulated arena duel between two characters.
type DuelResult struct {
	BattleID   string           `json:"battle_id"`
	WinnerID   string           `json:"winner_id"`
	WinnerName string           `json:"winner_name"`
	LoserID    string           `json:"loser_id"`
	LoserName  string           `json:"loser_name"`
	WinnerHP   int              `json:"winner_hp"`
	BattleLog  []map[string]any `json:"battle_log"`
}

// SimulateDuel runs a one-off auto-resolved duel between two characters using
// class-baseline stats and rank-appropriate potion bags.
func (s *Service) SimulateDuel(a, b characters.Summary) DuelResult {
	combA := combat.BaselineCombatant(a.Class)
	combA.EntityID = a.CharacterID
	combA.Name = a.Name
	combA.Team = "a"
	combA.IsPlayerSide = true
	combA.CurrentHP = combA.MaxHP
	combA.PotionBag = combat.DefaultPotionBag(a.Rank)

	combB := combat.BaselineCombatant(b.Class)
	combB.EntityID = b.CharacterID
	combB.Name = b.Name
	combB.Team = "b"
	combB.IsPlayerSide = true
	combB.CurrentHP = combB.MaxHP
	combB.PotionBag = combat.DefaultPotionBag(b.Rank)

	battleID := fmt.Sprintf("duel_%d", s.clock().UnixNano())
	result := combat.SimulateBattle(combat.BattleConfig{
		BattleType: "arena_duel",
		RunID:      battleID,
		RoomIndex:  1,
		SideA:      combA,
		SideB:      combB,
	})

	if result.SideAWon {
		return DuelResult{
			BattleID:   battleID,
			WinnerID:   a.CharacterID,
			WinnerName: a.Name,
			LoserID:    b.CharacterID,
			LoserName:  b.Name,
			WinnerHP:   result.SideAFinalHP,
			BattleLog:  result.Log,
		}
	}
	return DuelResult{
		BattleID:   battleID,
		WinnerID:   b.CharacterID,
		WinnerName: b.Name,
		LoserID:    a.CharacterID,
		LoserName:  a.Name,
		WinnerHP:   result.SideBFinalHP,
		BattleLog:  result.Log,
	}
}
