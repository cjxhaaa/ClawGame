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
	ErrSignupClosed       = errors.New("arena signup closed")
	ErrRankNotEligible    = errors.New("arena rank not eligible")
	ErrAlreadySignedUp    = errors.New("arena already signed up")
	ErrArenaMatchNotFound = errors.New("arena match not found")
)

const mainBracketSize = 64

type Entry struct {
	CharacterID     string `json:"character_id"`
	CharacterName   string `json:"character_name"`
	Class           string `json:"class"`
	WeaponStyle     string `json:"weapon_style"`
	Rank            string `json:"rank"`
	PanelPowerScore int    `json:"panel_power_score"`
	EquipmentScore  int    `json:"equipment_score"`
	SignedUpAt      string `json:"signed_up_at"`
	IsNPC           bool   `json:"is_npc,omitempty"`
}

type Matchup struct {
	MatchID        string `json:"match_id"`
	Stage          string `json:"stage"`
	RoundNumber    int    `json:"round_number"`
	MatchNumber    int    `json:"match_number"`
	Status         string `json:"status"`
	ScheduledAt    string `json:"scheduled_at"`
	ResolvedAt     string `json:"resolved_at,omitempty"`
	BattleReportID string `json:"battle_report_id,omitempty"`
	LeftEntry      *Entry `json:"left_entry,omitempty"`
	RightEntry     *Entry `json:"right_entry,omitempty"`
	ByeEntry       *Entry `json:"bye_entry,omitempty"`
	WinnerEntry    *Entry `json:"winner_entry,omitempty"`
}

type Round struct {
	Name         string    `json:"name"`
	Stage        string    `json:"stage"`
	RoundNumber  int       `json:"round_number"`
	EntrantCount int       `json:"entrant_count"`
	Status       string    `json:"status"`
	ScheduledAt  string    `json:"scheduled_at"`
	ResolvedAt   string    `json:"resolved_at,omitempty"`
	Matchups     []Matchup `json:"matchups"`
}

type CurrentView struct {
	TournamentID      string            `json:"tournament_id"`
	DayKey            string            `json:"day_key"`
	WeekKey           string            `json:"week_key,omitempty"`
	Status            world.ArenaStatus `json:"status"`
	SignupCount       int               `json:"signup_count"`
	QualifiedCount    int               `json:"qualified_count"`
	NPCCount          int               `json:"npc_count"`
	HighestPower      int               `json:"highest_panel_power"`
	LowestPower       int               `json:"lowest_panel_power"`
	MedianPower       int               `json:"median_panel_power"`
	FeaturedEntries   []Entry           `json:"featured_entries"`
	QualifierMatchups []Matchup         `json:"qualifier_matchups"`
	QualifierRounds   []Round           `json:"qualifier_rounds"`
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

type HistoryFilters struct {
	Result       string
	TournamentID string
	Stage        string
	Cursor       string
	Limit        int
}

type HistoryOpponent struct {
	CharacterID     string `json:"character_id"`
	CharacterName   string `json:"character_name"`
	Class           string `json:"class"`
	WeaponStyle     string `json:"weapon_style"`
	Rank            string `json:"rank"`
	PanelPowerScore int    `json:"panel_power_score"`
	IsNPC           bool   `json:"is_npc,omitempty"`
}

type HistorySummary struct {
	MatchID        string           `json:"match_id"`
	BattleReportID string           `json:"battle_report_id,omitempty"`
	TournamentID   string           `json:"tournament_id"`
	DayKey         string           `json:"day_key"`
	Stage          string           `json:"stage"`
	RoundNumber    int              `json:"round_number"`
	RoundName      string           `json:"round_name"`
	Result         string           `json:"result"`
	SummaryTag     string           `json:"summary_tag"`
	StartedAt      string           `json:"started_at"`
	ResolvedAt     string           `json:"resolved_at,omitempty"`
	Opponent       *HistoryOpponent `json:"opponent_summary,omitempty"`
}

type HistoryDetail struct {
	HistorySummary
	BattleReport map[string]any   `json:"battle_report,omitempty"`
	BattleLog    []map[string]any `json:"battle_log,omitempty"`
}

type PublicMatchDetail struct {
	MatchID        string           `json:"match_id"`
	BattleReportID string           `json:"battle_report_id,omitempty"`
	TournamentID   string           `json:"tournament_id"`
	DayKey         string           `json:"day_key"`
	Stage          string           `json:"stage"`
	RoundNumber    int              `json:"round_number"`
	RoundName      string           `json:"round_name"`
	Status         string           `json:"status"`
	SummaryTag     string           `json:"summary_tag"`
	StartedAt      string           `json:"started_at"`
	ResolvedAt     string           `json:"resolved_at,omitempty"`
	LeftEntry      *HistoryOpponent `json:"left_entry,omitempty"`
	RightEntry     *HistoryOpponent `json:"right_entry,omitempty"`
	ByeEntry       *HistoryOpponent `json:"bye_entry,omitempty"`
	WinnerEntry    *HistoryOpponent `json:"winner_entry,omitempty"`
	BattleReport   map[string]any   `json:"battle_report,omitempty"`
	BattleLog      []map[string]any `json:"battle_log,omitempty"`
}

type EntryListFilters struct {
	Cursor string
	Limit  int
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

func (s *Service) SetClock(clock func() time.Time) {
	if clock == nil {
		s.clock = time.Now
		return
	}
	s.clock = clock
}

func (s *Service) Signup(character characters.Summary, panelPowerScore, equipmentScore int, arenaStatus world.ArenaStatus) (Entry, error) {
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
		CharacterID:     character.CharacterID,
		CharacterName:   character.Name,
		Class:           character.Class,
		WeaponStyle:     character.WeaponStyle,
		Rank:            character.Rank,
		PanelPowerScore: panelPowerScore,
		EquipmentScore:  equipmentScore,
		SignedUpAt:      s.clock().Format(time.RFC3339),
	}
	s.entriesByDay[dayKey][character.CharacterID] = entry
	return entry, nil
}

func (s *Service) GetCurrent(arenaStatus world.ArenaStatus) CurrentView {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.clock()
	dayKey := dayKeyFor(now)
	entries := sortedEntries(s.entriesByDay[dayKey])
	snapshot := buildTournamentSnapshot(dayKey, now, entries)

	return CurrentView{
		TournamentID:      tournamentIDForDay(dayKey),
		DayKey:            dayKey,
		WeekKey:           dayKey,
		Status:            arenaStatus,
		SignupCount:       len(entries),
		QualifiedCount:    len(snapshot.mainField),
		NPCCount:          snapshot.npcCount,
		HighestPower:      highestPanelPower(entries),
		LowestPower:       lowestPanelPower(entries),
		MedianPower:       medianPanelPower(entries),
		FeaturedEntries:   featuredEntries(entries, 8),
		QualifierMatchups: snapshot.currentQualifierMatchups,
		QualifierRounds:   snapshot.qualifierRounds,
		Matchups:          snapshot.currentMatchups,
		Rounds:            snapshot.rounds,
		Champion:          snapshot.champion,
		NextRoundTime:     determineNextRoundTime(now, arenaStatus.Code, snapshot).Format(time.RFC3339),
	}
}

func (s *Service) ListEntries(filters EntryListFilters) []Entry {
	s.mu.Lock()
	defer s.mu.Unlock()

	dayKey := dayKeyFor(s.clock())
	entries := sortedEntries(s.entriesByDay[dayKey])
	if filters.Limit <= 0 {
		filters.Limit = 20
	}
	if filters.Cursor != "" {
		start := -1
		for i, entry := range entries {
			if entry.CharacterID == filters.Cursor {
				start = i + 1
				break
			}
		}
		if start >= 0 {
			entries = entries[start:]
		}
	}
	if len(entries) > filters.Limit {
		return entries[:filters.Limit]
	}
	return entries
}

func (s *Service) GetLeaderboard() []LeaderboardEntry {
	s.mu.Lock()
	defer s.mu.Unlock()

	dayKey := dayKeyFor(s.clock())
	entries := sortedEntries(s.entriesByDay[dayKey])
	result := make([]LeaderboardEntry, 0, len(entries))
	for i, entry := range entries {
		result = append(result, LeaderboardEntry{
			Rank:        i + 1,
			CharacterID: entry.CharacterID,
			Name:        entry.CharacterName,
			Score:       entry.PanelPowerScore,
			ScoreLabel:  "panel_power_score",
		})
	}
	return result
}

func (s *Service) ListHistory(characterID string, filters HistoryFilters) []HistorySummary {
	s.mu.Lock()
	defer s.mu.Unlock()

	items := s.buildHistoryLocked(characterID)
	items = filterHistory(items, filters)
	return paginateHistory(items, filters.Cursor, filters.Limit)
}

func (s *Service) GetHistoryDetail(characterID, matchID, detailLevel string) (HistoryDetail, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	match, round, found := s.findMatchLocked(characterID, matchID)
	if !found {
		return HistoryDetail{}, ErrArenaMatchNotFound
	}
	return buildHistoryDetail(characterID, round, match, detailLevel), nil
}

func (s *Service) GetPublicMatchDetail(matchID, detailLevel string) (PublicMatchDetail, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	match, round, found := s.findPublicMatchLocked(matchID)
	if !found {
		return PublicMatchDetail{}, ErrArenaMatchNotFound
	}
	return buildPublicMatchDetail(round, match, detailLevel), nil
}

func dayKeyFor(now time.Time) string {
	return now.Format("2006-01-02")
}

type tournamentSnapshot struct {
	mainField                []Entry
	npcCount                 int
	qualifierRounds          []Round
	currentQualifierMatchups []Matchup
	currentMatchups          []Matchup
	rounds                   []Round
	champion                 *Entry
}

func buildTournamentSnapshot(dayKey string, now time.Time, entries []Entry) tournamentSnapshot {
	var snapshot tournamentSnapshot

	qualified, qualifierRounds := buildQualifierRounds(dayKey, now, entries)
	mainField, npcCount := fillNPCEntries(dayKey, qualified, mainBracketSize)
	snapshot.mainField = mainField
	snapshot.npcCount = npcCount
	snapshot.qualifierRounds = qualifierRounds
	snapshot.currentQualifierMatchups = currentRoundMatchups(qualifierRounds)
	if len(mainField) == 0 {
		snapshot.currentMatchups = snapshot.currentQualifierMatchups
		return snapshot
	}

	mainField = shuffleEntries(seedForKey(dayKey, "main-field"), mainField)
	rounds, champion := buildMainRounds(dayKey, now, mainField, len(qualifierRounds))
	snapshot.rounds = rounds
	snapshot.champion = champion
	if len(snapshot.currentQualifierMatchups) > 0 {
		snapshot.currentMatchups = snapshot.currentQualifierMatchups
	} else {
		snapshot.currentMatchups = currentRoundMatchups(rounds)
	}
	return snapshot
}

func buildQualifierRounds(dayKey string, now time.Time, entries []Entry) ([]Entry, []Round) {
	if len(entries) == 0 {
		return nil, nil
	}

	participants := append([]Entry(nil), entries...)
	scheduledAt := qualifierStartTime(now)
	roundNumber := 1
	rounds := make([]Round, 0, 8)

	for len(participants) > mainBracketSize {
		entrantCount := len(participants)
		shuffled := shuffleEntries(seedForKey(dayKey, fmt.Sprintf("qualifier-round-%02d", roundNumber)), participants)
		resolvedAt := scheduledAt.Add(5 * time.Minute)
		matchups := make([]Matchup, 0, (len(shuffled)+1)/2)
		nextParticipants := make([]Entry, 0, (len(shuffled)+1)/2)
		matchNumber := 1

		if len(shuffled)%2 == 1 {
			bye := shuffled[len(shuffled)-1]
			match := Matchup{
				MatchID:     arenaMatchID(dayKey, "qualifier", roundNumber, matchNumber),
				Stage:       "qualifier",
				RoundNumber: roundNumber,
				MatchNumber: matchNumber,
				Status:      matchupStatus(now, scheduledAt, resolvedAt),
				ScheduledAt: scheduledAt.Format(time.RFC3339),
				ByeEntry:    cloneEntry(bye),
			}
			if !now.Before(resolvedAt) {
				match.ResolvedAt = resolvedAt.Format(time.RFC3339)
				match.WinnerEntry = cloneEntry(bye)
			}
			matchups = append(matchups, match)
			nextParticipants = append(nextParticipants, bye)
			shuffled = shuffled[:len(shuffled)-1]
			matchNumber++
		}

		for i := 0; i+1 < len(shuffled); i += 2 {
			left := shuffled[i]
			right := shuffled[i+1]
			winner, reportID := resolveArenaMatch(dayKey, "qualifier", roundNumber, matchNumber, left, right)
			match := Matchup{
				MatchID:        arenaMatchID(dayKey, "qualifier", roundNumber, matchNumber),
				Stage:          "qualifier",
				RoundNumber:    roundNumber,
				MatchNumber:    matchNumber,
				Status:         matchupStatus(now, scheduledAt, resolvedAt),
				ScheduledAt:    scheduledAt.Format(time.RFC3339),
				BattleReportID: reportID,
				LeftEntry:      cloneEntry(left),
				RightEntry:     cloneEntry(right),
			}
			if !now.Before(resolvedAt) {
				match.ResolvedAt = resolvedAt.Format(time.RFC3339)
				match.WinnerEntry = cloneEntry(winner)
			}
			matchups = append(matchups, match)
			nextParticipants = append(nextParticipants, winner)
			matchNumber++
		}

		round := Round{
			Name:         fmt.Sprintf("Qualifier Round %d", roundNumber),
			Stage:        "qualifier",
			RoundNumber:  roundNumber,
			EntrantCount: entrantCount,
			Status:       roundStatusFromMatches(matchups),
			ScheduledAt:  scheduledAt.Format(time.RFC3339),
			Matchups:     matchups,
		}
		if !now.Before(resolvedAt) {
			round.ResolvedAt = resolvedAt.Format(time.RFC3339)
		}
		rounds = append(rounds, round)
		participants = nextParticipants
		scheduledAt = scheduledAt.Add(5 * time.Minute)
		roundNumber++
	}

	return participants, rounds
}

func buildMainRounds(dayKey string, now time.Time, mainField []Entry, qualifierRounds int) ([]Round, *Entry) {
	participants := append([]Entry(nil), mainField...)
	scheduledAt := mainBracketStartTime(now, qualifierRounds)
	rounds := make([]Round, 0, 6)

	for len(participants) >= 2 {
		entrantCount := len(participants)
		stage := fmt.Sprintf("top_%d", len(participants))
		roundNumber := len(rounds) + 1
		resolvedAt := scheduledAt.Add(5 * time.Minute)
		matchups := make([]Matchup, 0, len(participants)/2)
		nextParticipants := make([]Entry, 0, len(participants)/2)

		for i := 0; i+1 < len(participants); i += 2 {
			matchNumber := (i / 2) + 1
			left := participants[i]
			right := participants[i+1]
			winner, reportID := resolveArenaMatch(dayKey, stage, roundNumber, matchNumber, left, right)
			match := Matchup{
				MatchID:        arenaMatchID(dayKey, stage, roundNumber, matchNumber),
				Stage:          stage,
				RoundNumber:    roundNumber,
				MatchNumber:    matchNumber,
				Status:         matchupStatus(now, scheduledAt, resolvedAt),
				ScheduledAt:    scheduledAt.Format(time.RFC3339),
				BattleReportID: reportID,
				LeftEntry:      cloneEntry(left),
				RightEntry:     cloneEntry(right),
			}
			if !now.Before(resolvedAt) {
				match.ResolvedAt = resolvedAt.Format(time.RFC3339)
				match.WinnerEntry = cloneEntry(winner)
			}
			matchups = append(matchups, match)
			nextParticipants = append(nextParticipants, winner)
		}

		round := Round{
			Name:         roundNameFor(len(participants)),
			Stage:        stage,
			RoundNumber:  roundNumber,
			EntrantCount: entrantCount,
			Status:       roundStatusFromMatches(matchups),
			ScheduledAt:  scheduledAt.Format(time.RFC3339),
			Matchups:     matchups,
		}
		if !now.Before(resolvedAt) {
			round.ResolvedAt = resolvedAt.Format(time.RFC3339)
		}
		rounds = append(rounds, round)
		participants = nextParticipants
		scheduledAt = scheduledAt.Add(5 * time.Minute)
	}

	if len(participants) == 1 && len(rounds) > 0 && rounds[len(rounds)-1].Status == "resolved" {
		return rounds, cloneEntry(participants[0])
	}
	return rounds, nil
}

func fillNPCEntries(dayKey string, entries []Entry, target int) ([]Entry, int) {
	if len(entries) >= target {
		return entries[:target], 0
	}

	medianScore := medianPanelPower(entries)
	if medianScore == 0 {
		medianScore = 6200
	}
	medianEquipmentScore := medianEquipment(entries)
	if medianEquipmentScore == 0 {
		medianEquipmentScore = 320
	}
	rank := medianRank(entries)
	if rank == "" {
		rank = "mid"
	}

	result := append([]Entry(nil), entries...)
	for i := len(entries); i < target; i++ {
		result = append(result, Entry{
			CharacterID:     fmt.Sprintf("npc_%s_%02d", dayKey, i+1),
			CharacterName:   fmt.Sprintf("Arena NPC %02d", i+1),
			Class:           npcClassFor(i),
			WeaponStyle:     npcWeaponFor(i),
			Rank:            rank,
			PanelPowerScore: medianScore,
			EquipmentScore:  medianEquipmentScore,
			SignedUpAt:      "",
			IsNPC:           true,
		})
	}
	return result, target - len(entries)
}

func resolveArenaMatch(dayKey, stage string, roundNumber, matchNumber int, left, right Entry) (Entry, string) {
	result := simulateEntryDuel(left, right, arenaMatchID(dayKey, stage, roundNumber, matchNumber))
	if result.SideAWon {
		return left, arenaReportID(dayKey, stage, roundNumber, matchNumber)
	}
	return right, arenaReportID(dayKey, stage, roundNumber, matchNumber)
}

func cloneEntry(entry Entry) *Entry {
	item := entry
	return &item
}

func tournamentIDForDay(dayKey string) string {
	return "tourn_" + dayKey
}

func arenaMatchID(dayKey, stage string, roundNumber, matchNumber int) string {
	return fmt.Sprintf("match_%s_%s_r%02d_m%03d", dayKey, stage, roundNumber, matchNumber)
}

func arenaReportID(dayKey, stage string, roundNumber, matchNumber int) string {
	return fmt.Sprintf("report_%s_%s_r%02d_m%03d", dayKey, stage, roundNumber, matchNumber)
}

func seedForKey(dayKey, suffix string) int64 {
	hasher := fnv.New64a()
	_, _ = hasher.Write([]byte(dayKey))
	_, _ = hasher.Write([]byte("::"))
	_, _ = hasher.Write([]byte(suffix))
	return int64(hasher.Sum64())
}

func determineNextRoundTime(now time.Time, statusCode string, snapshot tournamentSnapshot) time.Time {
	if statusCode == "signup_open" {
		return qualifierStartTime(now)
	}

	for _, round := range append(append([]Round(nil), snapshot.qualifierRounds...), snapshot.rounds...) {
		if round.Status == "scheduled" {
			if ts, err := time.Parse(time.RFC3339, round.ScheduledAt); err == nil {
				return ts
			}
		}
		if round.Status == "in_progress" && round.ResolvedAt != "" {
			if ts, err := time.Parse(time.RFC3339, round.ResolvedAt); err == nil {
				return ts
			}
		}
	}

	return qualifierStartTime(now).Add(24 * time.Hour)
}

func mainBracketStartTime(now time.Time, qualifierRounds int) time.Time {
	start := qualifierStartTime(now)
	if qualifierRounds == 0 {
		return start.Add(5 * time.Minute)
	}
	return start.Add(time.Duration(qualifierRounds) * 5 * time.Minute)
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

func sortedEntries(entriesMap map[string]Entry) []Entry {
	entries := make([]Entry, 0, len(entriesMap))
	for _, entry := range entriesMap {
		entries = append(entries, entry)
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].PanelPowerScore != entries[j].PanelPowerScore {
			return entries[i].PanelPowerScore > entries[j].PanelPowerScore
		}
		if entries[i].EquipmentScore != entries[j].EquipmentScore {
			return entries[i].EquipmentScore > entries[j].EquipmentScore
		}
		if entries[i].SignedUpAt != entries[j].SignedUpAt {
			return entries[i].SignedUpAt < entries[j].SignedUpAt
		}
		return entries[i].CharacterID < entries[j].CharacterID
	})
	return entries
}

func featuredEntries(entries []Entry, limit int) []Entry {
	if limit <= 0 || len(entries) == 0 {
		return nil
	}
	if len(entries) <= limit {
		return append([]Entry(nil), entries...)
	}
	return append([]Entry(nil), entries[:limit]...)
}

func currentRoundMatchups(rounds []Round) []Matchup {
	for _, round := range rounds {
		if round.Status == "in_progress" {
			return round.Matchups
		}
	}
	for _, round := range rounds {
		if round.Status == "scheduled" {
			return round.Matchups
		}
	}
	return nil
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

func medianPanelPower(entries []Entry) int {
	if len(entries) == 0 {
		return 0
	}
	scores := make([]int, 0, len(entries))
	for _, entry := range entries {
		scores = append(scores, entry.PanelPowerScore)
	}
	sort.Ints(scores)
	mid := len(scores) / 2
	if len(scores)%2 == 0 {
		return (scores[mid-1] + scores[mid]) / 2
	}
	return scores[mid]
}

func highestPanelPower(entries []Entry) int {
	if len(entries) == 0 {
		return 0
	}
	return entries[0].PanelPowerScore
}

func lowestPanelPower(entries []Entry) int {
	if len(entries) == 0 {
		return 0
	}
	return entries[len(entries)-1].PanelPowerScore
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

func simulateEntryDuel(a, b Entry, runID string) combat.BattleResult {
	combA := buildCombatantFromEntry(a, "a")
	combB := buildCombatantFromEntry(b, "b")
	return combat.SimulateBattle(combat.BattleConfig{
		BattleType: "arena_duel",
		RunID:      runID,
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
	referencePower := 6200.0
	if entry.Rank == "high" {
		referencePower = 9800.0
	}
	scoreFactor := float64(entry.PanelPowerScore) / referencePower
	scoreFactor = math.Max(0.72, math.Min(1.45, scoreFactor))
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

func buildHistoryDetail(characterID string, round Round, match Matchup, detailLevel string) HistoryDetail {
	summary := historySummaryFromMatch(characterID, round, match)
	detail := HistoryDetail{HistorySummary: summary}
	if detailLevel == "" {
		detailLevel = "standard"
	}
	report, log := buildBattleArtifacts(match, characterID)
	switch detailLevel {
	case "compact":
		return detail
	case "verbose":
		detail.BattleReport = report
		detail.BattleLog = log
		return detail
	default:
		detail.BattleReport = report
		return detail
	}
}

func buildPublicMatchDetail(round Round, match Matchup, detailLevel string) PublicMatchDetail {
	detail := PublicMatchDetail{
		MatchID:        match.MatchID,
		BattleReportID: match.BattleReportID,
		TournamentID:   tournamentIDForDay(dayKeyFromMatchID(match.MatchID)),
		DayKey:         dayKeyFromMatchID(match.MatchID),
		Stage:          match.Stage,
		RoundNumber:    match.RoundNumber,
		RoundName:      round.Name,
		Status:         match.Status,
		SummaryTag:     publicSummaryTag(match),
		StartedAt:      match.ScheduledAt,
		ResolvedAt:     match.ResolvedAt,
		LeftEntry:      historyOpponentFromEntry(match.LeftEntry),
		RightEntry:     historyOpponentFromEntry(match.RightEntry),
		ByeEntry:       historyOpponentFromEntry(match.ByeEntry),
		WinnerEntry:    historyOpponentFromEntry(match.WinnerEntry),
	}
	if detailLevel == "" {
		detailLevel = "standard"
	}
	report, log := buildPublicBattleArtifacts(match)
	switch detailLevel {
	case "compact":
		return detail
	case "verbose":
		detail.BattleReport = report
		detail.BattleLog = log
		return detail
	default:
		detail.BattleReport = report
		return detail
	}
}

func buildBattleArtifacts(match Matchup, characterID string) (map[string]any, []map[string]any) {
	if match.ByeEntry != nil && match.LeftEntry == nil && match.RightEntry == nil {
		report := map[string]any{
			"outcome":        "bye",
			"winner":         match.WinnerEntry,
			"summary_tag":    "advanced_by_bye",
			"decisive_turns": []any{},
		}
		return report, nil
	}
	if match.LeftEntry == nil || match.RightEntry == nil {
		return nil, nil
	}

	result := simulateEntryDuel(*match.LeftEntry, *match.RightEntry, match.MatchID)
	report := map[string]any{
		"outcome":              historyResultForMatch(match, characterID),
		"winner":               match.WinnerEntry,
		"loser":                loserFromMatch(match),
		"winner_final_hp":      winnerFinalHP(match, result),
		"player_power_delta":   match.LeftEntry.PanelPowerScore - match.RightEntry.PanelPowerScore,
		"battle_length_events": len(result.Log),
		"summary_tag":          summaryTag(match, result),
	}
	return report, result.Log
}

func buildPublicBattleArtifacts(match Matchup) (map[string]any, []map[string]any) {
	if match.ByeEntry != nil && match.LeftEntry == nil && match.RightEntry == nil {
		report := map[string]any{
			"outcome":        "bye",
			"winner":         match.WinnerEntry,
			"summary_tag":    "advanced_by_bye",
			"decisive_turns": []any{},
		}
		return report, nil
	}
	if match.LeftEntry == nil || match.RightEntry == nil {
		return nil, nil
	}

	result := simulateEntryDuel(*match.LeftEntry, *match.RightEntry, match.MatchID)
	report := map[string]any{
		"outcome":              "resolved",
		"winner":               match.WinnerEntry,
		"loser":                loserFromMatch(match),
		"winner_final_hp":      winnerFinalHP(match, result),
		"power_delta":          match.LeftEntry.PanelPowerScore - match.RightEntry.PanelPowerScore,
		"battle_length_events": len(result.Log),
		"summary_tag":          publicSummaryTag(match),
	}
	return report, result.Log
}

func loserFromMatch(match Matchup) *Entry {
	if match.LeftEntry == nil || match.RightEntry == nil || match.WinnerEntry == nil {
		return nil
	}
	if match.WinnerEntry.CharacterID == match.LeftEntry.CharacterID {
		return cloneEntry(*match.RightEntry)
	}
	return cloneEntry(*match.LeftEntry)
}

func winnerFinalHP(match Matchup, result combat.BattleResult) int {
	if match.WinnerEntry == nil || match.LeftEntry == nil || match.RightEntry == nil {
		return 0
	}
	if match.WinnerEntry.CharacterID == match.LeftEntry.CharacterID {
		return result.SideAFinalHP
	}
	return result.SideBFinalHP
}

func summaryTag(match Matchup, result combat.BattleResult) string {
	if match.ByeEntry != nil && match.LeftEntry == nil && match.RightEntry == nil {
		return "advanced_by_bye"
	}
	if match.WinnerEntry == nil || match.LeftEntry == nil || match.RightEntry == nil {
		return "unresolved"
	}
	winner := buildCombatantFromEntry(*match.WinnerEntry, "a")
	winnerHP := winnerFinalHP(match, result)
	if winnerHP >= int(float64(winner.MaxHP)*0.7) {
		return "won_clean"
	}
	if winnerHP > 0 {
		return "won_close"
	}
	return "lost"
}

func historySummaryFromMatch(characterID string, round Round, match Matchup) HistorySummary {
	var opponent *HistoryOpponent
	result := historyResultForMatch(match, characterID)
	if match.ByeEntry == nil {
		opponent = historyOpponentFor(match, characterID)
	}
	return HistorySummary{
		MatchID:        match.MatchID,
		BattleReportID: match.BattleReportID,
		TournamentID:   tournamentIDForDay(dayKeyFromMatchID(match.MatchID)),
		DayKey:         dayKeyFromMatchID(match.MatchID),
		Stage:          match.Stage,
		RoundNumber:    match.RoundNumber,
		RoundName:      round.Name,
		Result:         result,
		SummaryTag:     historySummaryTag(result),
		StartedAt:      match.ScheduledAt,
		ResolvedAt:     match.ResolvedAt,
		Opponent:       opponent,
	}
}

func historySummaryTag(result string) string {
	switch result {
	case "win":
		return "advanced"
	case "loss":
		return "eliminated"
	case "bye":
		return "advanced_by_bye"
	default:
		return "pending"
	}
}

func historyResultForMatch(match Matchup, characterID string) string {
	if match.ByeEntry != nil && match.ByeEntry.CharacterID == characterID {
		return "bye"
	}
	if match.WinnerEntry == nil {
		return "pending"
	}
	if match.WinnerEntry.CharacterID == characterID {
		return "win"
	}
	return "loss"
}

func historyOpponentFor(match Matchup, characterID string) *HistoryOpponent {
	var opponent *Entry
	switch {
	case match.LeftEntry != nil && match.LeftEntry.CharacterID == characterID:
		opponent = match.RightEntry
	case match.RightEntry != nil && match.RightEntry.CharacterID == characterID:
		opponent = match.LeftEntry
	}
	if opponent == nil {
		return nil
	}
	return &HistoryOpponent{
		CharacterID:     opponent.CharacterID,
		CharacterName:   opponent.CharacterName,
		Class:           opponent.Class,
		WeaponStyle:     opponent.WeaponStyle,
		Rank:            opponent.Rank,
		PanelPowerScore: opponent.PanelPowerScore,
		IsNPC:           opponent.IsNPC,
	}
}

func historyOpponentFromEntry(entry *Entry) *HistoryOpponent {
	if entry == nil {
		return nil
	}
	return &HistoryOpponent{
		CharacterID:     entry.CharacterID,
		CharacterName:   entry.CharacterName,
		Class:           entry.Class,
		WeaponStyle:     entry.WeaponStyle,
		Rank:            entry.Rank,
		PanelPowerScore: entry.PanelPowerScore,
		IsNPC:           entry.IsNPC,
	}
}

func (s *Service) buildHistoryLocked(characterID string) []HistorySummary {
	now := s.clock()
	dayKeys := make([]string, 0, len(s.entriesByDay))
	for dayKey := range s.entriesByDay {
		dayKeys = append(dayKeys, dayKey)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(dayKeys)))

	items := make([]HistorySummary, 0, 32)
	for _, dayKey := range dayKeys {
		effectiveNow := effectiveNowForDay(now, dayKey)
		snapshot := buildTournamentSnapshot(dayKey, effectiveNow, sortedEntries(s.entriesByDay[dayKey]))
		for _, round := range append(append([]Round(nil), snapshot.qualifierRounds...), snapshot.rounds...) {
			for _, match := range round.Matchups {
				if match.ResolvedAt == "" || !matchContainsCharacter(match, characterID) {
					continue
				}
				items = append(items, historySummaryFromMatch(characterID, round, match))
			}
		}
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].ResolvedAt != items[j].ResolvedAt {
			return items[i].ResolvedAt > items[j].ResolvedAt
		}
		return items[i].MatchID > items[j].MatchID
	})
	return items
}

func (s *Service) findMatchLocked(characterID, matchID string) (Matchup, Round, bool) {
	now := s.clock()
	for dayKey := range s.entriesByDay {
		effectiveNow := effectiveNowForDay(now, dayKey)
		snapshot := buildTournamentSnapshot(dayKey, effectiveNow, sortedEntries(s.entriesByDay[dayKey]))
		for _, round := range append(append([]Round(nil), snapshot.qualifierRounds...), snapshot.rounds...) {
			for _, match := range round.Matchups {
				if match.MatchID == matchID && match.ResolvedAt != "" && matchContainsCharacter(match, characterID) {
					return match, round, true
				}
			}
		}
	}
	return Matchup{}, Round{}, false
}

func (s *Service) findPublicMatchLocked(matchID string) (Matchup, Round, bool) {
	now := s.clock()
	for dayKey := range s.entriesByDay {
		effectiveNow := effectiveNowForDay(now, dayKey)
		snapshot := buildTournamentSnapshot(dayKey, effectiveNow, sortedEntries(s.entriesByDay[dayKey]))
		for _, round := range append(append([]Round(nil), snapshot.qualifierRounds...), snapshot.rounds...) {
			for _, match := range round.Matchups {
				if match.MatchID == matchID && match.ResolvedAt != "" {
					return match, round, true
				}
			}
		}
	}
	return Matchup{}, Round{}, false
}

func publicSummaryTag(match Matchup) string {
	if match.ByeEntry != nil && match.LeftEntry == nil && match.RightEntry == nil {
		return "advanced_by_bye"
	}
	if match.WinnerEntry == nil || match.LeftEntry == nil || match.RightEntry == nil {
		return "pending"
	}
	result := simulateEntryDuel(*match.LeftEntry, *match.RightEntry, match.MatchID)
	return summaryTag(match, result)
}

func filterHistory(items []HistorySummary, filters HistoryFilters) []HistorySummary {
	filtered := make([]HistorySummary, 0, len(items))
	for _, item := range items {
		if filters.Result != "" && item.Result != filters.Result {
			continue
		}
		if filters.TournamentID != "" && item.TournamentID != filters.TournamentID {
			continue
		}
		if filters.Stage != "" && item.Stage != filters.Stage {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}

func paginateHistory(items []HistorySummary, cursor string, limit int) []HistorySummary {
	if limit <= 0 {
		limit = 20
	}
	if cursor != "" {
		start := -1
		for i, item := range items {
			if item.MatchID == cursor {
				start = i + 1
				break
			}
		}
		if start >= 0 {
			items = items[start:]
		}
	}
	if len(items) > limit {
		return items[:limit]
	}
	return items
}

func effectiveNowForDay(now time.Time, dayKey string) time.Time {
	if dayKey == dayKeyFor(now) {
		return now
	}
	parsed, err := time.ParseInLocation("2006-01-02", dayKey, now.Location())
	if err != nil {
		return now
	}
	return parsed.Add(24 * time.Hour)
}

func matchContainsCharacter(match Matchup, characterID string) bool {
	return (match.LeftEntry != nil && match.LeftEntry.CharacterID == characterID) ||
		(match.RightEntry != nil && match.RightEntry.CharacterID == characterID) ||
		(match.ByeEntry != nil && match.ByeEntry.CharacterID == characterID)
}

func dayKeyFromMatchID(matchID string) string {
	if len(matchID) >= 16 {
		return matchID[6:16]
	}
	return ""
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
