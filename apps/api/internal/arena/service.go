package arena

import (
	"errors"
	"fmt"
	"hash/fnv"
	"math"
	"math/rand"
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

const (
	mainBracketSize      = 64
	maxQualifierEntrants = 128
)

type Entry struct {
	CharacterID    string `json:"character_id"`
	CharacterName  string `json:"character_name"`
	Class          string `json:"class"`
	WeaponStyle    string `json:"weapon_style"`
	Rank           string `json:"rank"`
	EquipmentScore int    `json:"equipment_score"`
	SignedUpAt     string `json:"signed_up_at"`
	IsNPC          bool   `json:"is_npc,omitempty"`
}

type Matchup struct {
	MatchNumber int    `json:"match_number"`
	Status      string `json:"status"`
	ScheduledAt string `json:"scheduled_at"`
	ResolvedAt  string `json:"resolved_at,omitempty"`
	LeftEntry   *Entry `json:"left_entry,omitempty"`
	RightEntry  *Entry `json:"right_entry,omitempty"`
	WinnerEntry *Entry `json:"winner_entry,omitempty"`
}

type Round struct {
	Name        string    `json:"name"`
	Stage       string    `json:"stage"`
	Status      string    `json:"status"`
	ScheduledAt string    `json:"scheduled_at"`
	ResolvedAt  string    `json:"resolved_at,omitempty"`
	Matchups    []Matchup `json:"matchups"`
}

type CurrentView struct {
	TournamentID      string            `json:"tournament_id"`
	DayKey            string            `json:"day_key"`
	WeekKey           string            `json:"week_key,omitempty"`
	Status            world.ArenaStatus `json:"status"`
	SignupCount       int               `json:"signup_count"`
	QualifiedCount    int               `json:"qualified_count"`
	NPCCount          int               `json:"npc_count"`
	Entries           []Entry           `json:"entries"`
	QualifierMatchups []Matchup         `json:"qualifier_matchups"`
	Matchups          []Matchup         `json:"matchups"`
	Rounds            []Round           `json:"rounds"`
	Champion          *Entry            `json:"champion,omitempty"`
	NextRoundTime     string            `json:"next_round_time"`
}

type LeaderboardEntry struct {
	Rank        int    `json:"rank"`
	CharacterID string `json:"character_id"`
	Name        string `json:"name"`
	Score       int    `json:"score"`
	ScoreLabel  string `json:"score_label"`
}

type Service struct {
	mu           sync.Mutex
	clock        func() time.Time
	entriesByDay map[string]map[string]Entry
}

func NewService() *Service {
	return &Service{
		clock:        time.Now,
		entriesByDay: make(map[string]map[string]Entry),
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

	dayKey := dayKeyFor(s.clock())
	if _, ok := s.entriesByDay[dayKey]; !ok {
		s.entriesByDay[dayKey] = make(map[string]Entry)
	}
	if _, exists := s.entriesByDay[dayKey][character.CharacterID]; exists {
		return Entry{}, ErrAlreadySignedUp
	}

	entry := Entry{
		CharacterID:    character.CharacterID,
		CharacterName:  character.Name,
		Class:          character.Class,
		WeaponStyle:    character.WeaponStyle,
		Rank:           character.Rank,
		EquipmentScore: equipmentScore,
		SignedUpAt:     s.clock().Format(time.RFC3339),
	}
	s.entriesByDay[dayKey][character.CharacterID] = entry
	return entry, nil
}

func (s *Service) GetCurrent(arenaStatus world.ArenaStatus) CurrentView {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.clock()
	dayKey := dayKeyFor(now)
	entriesMap := s.entriesByDay[dayKey]
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

	snapshot := buildTournamentSnapshot(dayKey, now, entries)
	return CurrentView{
		TournamentID:      "tourn_" + dayKey,
		DayKey:            dayKey,
		WeekKey:           dayKey,
		Status:            arenaStatus,
		SignupCount:       len(entries),
		QualifiedCount:    len(snapshot.mainField),
		NPCCount:          snapshot.npcCount,
		Entries:           entries,
		QualifierMatchups: snapshot.qualifierMatchups,
		Matchups:          snapshot.currentMatchups,
		Rounds:            snapshot.rounds,
		Champion:          snapshot.champion,
		NextRoundTime:     nextArenaMilestoneTime(now, arenaStatus.Code).Format(time.RFC3339),
	}
}

func (s *Service) GetLeaderboard() []LeaderboardEntry {
	s.mu.Lock()
	defer s.mu.Unlock()

	dayKey := dayKeyFor(s.clock())
	entriesMap := s.entriesByDay[dayKey]
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

func dayKeyFor(now time.Time) string {
	return now.Format("2006-01-02")
}

type tournamentSnapshot struct {
	mainField         []Entry
	npcCount          int
	qualifierMatchups []Matchup
	currentMatchups   []Matchup
	rounds            []Round
	champion          *Entry
}

func buildTournamentSnapshot(dayKey string, now time.Time, entries []Entry) tournamentSnapshot {
	var snapshot tournamentSnapshot

	qualified, qualifierMatchups := buildQualifiedField(dayKey, now, entries)
	mainField, npcCount := fillNPCEntries(dayKey, qualified, mainBracketSize)
	snapshot.mainField = mainField
	snapshot.npcCount = npcCount
	snapshot.qualifierMatchups = qualifierMatchups
	if len(mainField) == 0 {
		return snapshot
	}

	mainField = shuffleEntries(seedForKey(dayKey, "main-field"), mainField)
	rounds, champion, currentMatchups := buildMainRounds(dayKey, now, mainField)
	snapshot.rounds = rounds
	snapshot.champion = champion
	snapshot.currentMatchups = currentMatchups
	if now.Hour() == 9 && now.Minute() < 5 {
		snapshot.currentMatchups = qualifierMatchups
	}
	return snapshot
}

func buildQualifiedField(dayKey string, now time.Time, entries []Entry) ([]Entry, []Matchup) {
	if len(entries) == 0 {
		return nil, nil
	}

	shuffled := shuffleEntries(seedForKey(dayKey, "qualifier-pool"), entries)
	if len(shuffled) <= mainBracketSize {
		return shuffled, nil
	}

	if len(shuffled) > maxQualifierEntrants {
		shuffled = shuffled[:maxQualifierEntrants]
	}

	byeCount := (2 * mainBracketSize) - len(shuffled)
	if byeCount < 0 {
		byeCount = 0
	}
	if byeCount > len(shuffled) {
		byeCount = len(shuffled)
	}
	qualified := append([]Entry(nil), shuffled[:byeCount]...)
	contestants := shuffled[byeCount:]

	scheduledAt := qualifierStartTime(now)
	resolvedAt := scheduledAt.Add(5 * time.Minute)
	matchups := make([]Matchup, 0, len(contestants)/2)
	for i := 0; i+1 < len(contestants); i += 2 {
		left := contestants[i]
		right := contestants[i+1]
		winner := resolveEntryDuel(left, right)
		match := Matchup{
			MatchNumber: (i / 2) + 1,
			Status:      matchupStatus(now, scheduledAt, resolvedAt),
			ScheduledAt: scheduledAt.Format(time.RFC3339),
			LeftEntry:   cloneEntry(left),
			RightEntry:  cloneEntry(right),
		}
		if !now.Before(resolvedAt) {
			match.ResolvedAt = resolvedAt.Format(time.RFC3339)
			match.WinnerEntry = cloneEntry(winner)
		}
		matchups = append(matchups, match)
		qualified = append(qualified, winner)
	}

	return qualified, matchups
}

func buildMainRounds(dayKey string, now time.Time, mainField []Entry) ([]Round, *Entry, []Matchup) {
	participants := append([]Entry(nil), mainField...)
	scheduledAt := qualifierStartTime(now).Add(5 * time.Minute)
	rounds := make([]Round, 0, 6)
	var champion *Entry
	var current []Matchup

	for len(participants) >= 2 {
		roundName := roundNameFor(len(participants))
		resolvedAt := scheduledAt.Add(5 * time.Minute)
		matchups := make([]Matchup, 0, len(participants)/2)
		nextParticipants := make([]Entry, 0, len(participants)/2)
		for i := 0; i+1 < len(participants); i += 2 {
			left := participants[i]
			right := participants[i+1]
			winner := resolveEntryDuel(left, right)
			match := Matchup{
				MatchNumber: (i / 2) + 1,
				Status:      matchupStatus(now, scheduledAt, resolvedAt),
				ScheduledAt: scheduledAt.Format(time.RFC3339),
				LeftEntry:   cloneEntry(left),
				RightEntry:  cloneEntry(right),
			}
			if !now.Before(resolvedAt) {
				match.ResolvedAt = resolvedAt.Format(time.RFC3339)
				match.WinnerEntry = cloneEntry(winner)
			}
			matchups = append(matchups, match)
			nextParticipants = append(nextParticipants, winner)
		}

		roundStatus := roundStatusFromMatches(matchups)
		round := Round{
			Name:        roundName,
			Stage:       fmt.Sprintf("top_%d", len(participants)),
			Status:      roundStatus,
			ScheduledAt: scheduledAt.Format(time.RFC3339),
			Matchups:    matchups,
		}
		if !now.Before(resolvedAt) {
			round.ResolvedAt = resolvedAt.Format(time.RFC3339)
		}
		rounds = append(rounds, round)
		if roundStatus == "in_progress" {
			current = matchups
		}
		if roundStatus == "scheduled" && len(current) == 0 {
			current = matchups
		}
		participants = nextParticipants
		scheduledAt = scheduledAt.Add(5 * time.Minute)
	}

	if len(participants) == 1 && !now.Before(scheduledAt) {
		champion = cloneEntry(participants[0])
	}
	if len(current) == 0 && len(rounds) > 0 {
		current = rounds[len(rounds)-1].Matchups
	}
	return rounds, champion, current
}

func fillNPCEntries(dayKey string, entries []Entry, target int) ([]Entry, int) {
	if len(entries) >= target {
		return entries[:target], 0
	}

	medianScore := medianEquipment(entries)
	if medianScore == 0 {
		medianScore = 320
	}
	rank := medianRank(entries)
	if rank == "" {
		rank = "mid"
	}

	result := append([]Entry(nil), entries...)
	for i := len(entries); i < target; i++ {
		result = append(result, Entry{
			CharacterID:    fmt.Sprintf("npc_%s_%02d", dayKey, i+1),
			CharacterName:  fmt.Sprintf("Arena NPC %02d", i+1),
			Class:          npcClassFor(i),
			WeaponStyle:    npcWeaponFor(i),
			Rank:           rank,
			EquipmentScore: medianScore,
			SignedUpAt:     "",
			IsNPC:          true,
		})
	}
	return result, target - len(entries)
}

func resolveEntryDuel(left, right Entry) Entry {
	result := simulateEntryDuel(left, right)
	if result.SideAWon {
		return left
	}
	return right
}

func cloneEntry(entry Entry) *Entry {
	item := entry
	return &item
}

func seedForDay(dayKey string) int64 {
	hasher := fnv.New64a()
	_, _ = hasher.Write([]byte(dayKey))
	return int64(hasher.Sum64())
}

func seedForKey(dayKey, suffix string) int64 {
	hasher := fnv.New64a()
	_, _ = hasher.Write([]byte(dayKey))
	_, _ = hasher.Write([]byte("::"))
	_, _ = hasher.Write([]byte(suffix))
	return int64(hasher.Sum64())
}

func nextArenaMilestoneTime(now time.Time, statusCode string) time.Time {
	todayQualifiers := qualifierStartTime(now)

	switch statusCode {
	case "signup_open":
		return todayQualifiers
	case "signup_locked":
		return todayQualifiers.Add(5 * time.Minute)
	case "in_progress":
		minute := now.Minute()
		roundIndex := 1
		if minute >= 5 {
			roundIndex = ((minute - 5) / 5) + 2
		}
		next := todayQualifiers.Add(time.Duration(roundIndex*5) * time.Minute)
		final := todayQualifiers.Add(35 * time.Minute)
		if next.After(final) {
			return final
		}
		return next
	default:
		return todayQualifiers.Add(24 * time.Hour)
	}
}

func qualifierStartTime(now time.Time) time.Time {
	return time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, now.Location())
}

func shuffleEntries(seed int64, entries []Entry) []Entry {
	shuffled := append([]Entry(nil), entries...)
	rng := rand.New(rand.NewSource(seed))
	rng.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})
	return shuffled
}

func matchupStatus(now, scheduledAt, resolvedAt time.Time) string {
	if now.Before(scheduledAt) {
		return "scheduled"
	}
	if now.Before(resolvedAt) {
		return "in_progress"
	}
	return "resolved"
}

func roundStatusFromMatches(matchups []Matchup) string {
	if len(matchups) == 0 {
		return "scheduled"
	}
	status := "resolved"
	for _, match := range matchups {
		if match.Status == "in_progress" {
			return "in_progress"
		}
		if match.Status == "scheduled" {
			status = "scheduled"
		}
	}
	return status
}

func roundNameFor(participants int) string {
	switch participants {
	case 64:
		return "Round of 64"
	case 32:
		return "Round of 32"
	case 16:
		return "Round of 16"
	case 8:
		return "Quarterfinals"
	case 4:
		return "Semifinals"
	case 2:
		return "Final"
	default:
		return fmt.Sprintf("Top %d", participants)
	}
}

func medianEquipment(entries []Entry) int {
	if len(entries) == 0 {
		return 0
	}
	scores := make([]int, 0, len(entries))
	for _, entry := range entries {
		scores = append(scores, entry.EquipmentScore)
	}
	sort.Ints(scores)
	mid := len(scores) / 2
	if len(scores)%2 == 0 {
		return (scores[mid-1] + scores[mid]) / 2
	}
	return scores[mid]
}

func medianRank(entries []Entry) string {
	if len(entries) == 0 {
		return ""
	}
	high := 0
	for _, entry := range entries {
		if entry.Rank == "high" {
			high++
		}
	}
	if high*2 >= len(entries) {
		return "high"
	}
	return "mid"
}

func npcClassFor(index int) string {
	classes := []string{"warrior", "mage", "priest"}
	return classes[index%len(classes)]
}

func npcWeaponFor(index int) string {
	weapons := []string{"great_axe", "spellbook", "holy_tome"}
	return weapons[index%len(weapons)]
}

func simulateEntryDuel(a, b Entry) combat.BattleResult {
	combA := buildCombatantFromEntry(a, "a")
	combB := buildCombatantFromEntry(b, "b")
	return combat.SimulateBattle(combat.BattleConfig{
		BattleType: "arena_duel",
		RunID:      fmt.Sprintf("duel_%s_%s", a.CharacterID, b.CharacterID),
		RoomIndex:  1,
		SideA:      combA,
		SideB:      combB,
	})
}

func buildCombatantFromEntry(entry Entry, team string) combat.Combatant {
	comb := combat.BaselineCombatant(entry.Class)
	comb.EntityID = entry.CharacterID
	comb.Name = entry.CharacterName
	comb.Team = team
	comb.IsPlayerSide = true
	comb.PotionBag = combat.DefaultPotionBag(entry.Rank)
	scoreFactor := 1 + ((float64(entry.EquipmentScore) - 320) / 1000)
	scoreFactor = math.Max(0.75, math.Min(1.45, scoreFactor))
	rankFactor := 1.0
	if entry.Rank == "high" {
		rankFactor = 1.08
	}
	applyFactor := scoreFactor * rankFactor
	comb.MaxHP = maxInt(1, int(float64(comb.MaxHP)*applyFactor))
	comb.PhysAtk = maxInt(1, int(float64(comb.PhysAtk)*applyFactor))
	comb.MagAtk = maxInt(1, int(float64(comb.MagAtk)*applyFactor))
	comb.PhysDef = maxInt(1, int(float64(comb.PhysDef)*applyFactor))
	comb.MagDef = maxInt(1, int(float64(comb.MagDef)*applyFactor))
	comb.Speed = maxInt(1, int(float64(comb.Speed)*(0.9+applyFactor*0.1)))
	comb.HealPow = maxInt(1, int(float64(comb.HealPow)*applyFactor))
	comb.CurrentHP = comb.MaxHP
	return comb
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
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
