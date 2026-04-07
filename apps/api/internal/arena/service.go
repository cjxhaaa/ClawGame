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
	ErrSignupClosed           = errors.New("arena signup closed")
	ErrAlreadySignedUp        = errors.New("arena already signed up")
	ErrArenaMatchNotFound     = errors.New("arena match not found")
	ErrChallengeWindow        = errors.New("arena challenge window closed")
	ErrNoChallengeAttempts    = errors.New("arena no challenge attempts remaining")
	ErrInvalidChallengeTarget = errors.New("arena invalid challenge target")
	ErrPurchaseCapReached     = errors.New("arena purchase cap reached")
)

const mainBracketSize = 64
const (
	freeRatingChallengesPerDay   = 3
	maxPurchasedChallengesPerDay = 10
	basePurchasePriceGold        = 20
	arenaDuelMaxTurns            = 10
)

type arenaDuelMode string

const (
	arenaDuelModeRating     arenaDuelMode = "rating"
	arenaDuelModeKnockout   arenaDuelMode = "knockout"
	arenaDuelModeExhibition arenaDuelMode = "exhibition"
)

type arenaDuelResolution struct {
	Winner       Entry
	Loser        Entry
	WinnerHP     int
	EndReason    string
	Adjudication string
}

type Entry struct {
	CharacterID     string `json:"character_id"`
	CharacterName   string `json:"character_name"`
	Class           string `json:"class"`
	WeaponStyle     string `json:"weapon_style"`
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
	TournamentID        string            `json:"tournament_id"`
	DayKey              string            `json:"day_key"`
	WeekKey             string            `json:"week_key,omitempty"`
	Status              world.ArenaStatus `json:"status"`
	SignupCount         int               `json:"signup_count"`
	QualifiedCount      int               `json:"qualified_count"`
	NPCCount            int               `json:"npc_count"`
	HighestPower        int               `json:"highest_panel_power"`
	LowestPower         int               `json:"lowest_panel_power"`
	MedianPower         int               `json:"median_panel_power"`
	FeaturedEntries     []Entry           `json:"featured_entries"`
	QualifierMatchups   []Matchup         `json:"qualifier_matchups"`
	QualifierRounds     []Round           `json:"qualifier_rounds"`
	Matchups            []Matchup         `json:"matchups"`
	Rounds              []Round           `json:"rounds"`
	Champion            *Entry            `json:"champion,omitempty"`
	NextRoundTime       string            `json:"next_round_time"`
	WeeklyRatingSummary map[string]any    `json:"weekly_rating_summary,omitempty"`
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

type RatingCandidate struct {
	CharacterID     string `json:"character_id"`
	CharacterName   string `json:"character_name"`
	Class           string `json:"class"`
	WeaponStyle     string `json:"weapon_style"`
	Rating          int    `json:"rating"`
	PanelPowerScore int    `json:"panel_power_score"`
	EquipmentScore  int    `json:"equipment_score"`
	IsNPC           bool   `json:"is_npc,omitempty"`
}

type RatingBoardView struct {
	WeekKey               string            `json:"week_key"`
	CharacterID           string            `json:"character_id"`
	Rating                int               `json:"rating"`
	FreeAttemptsRemaining int               `json:"free_attempts_remaining"`
	PurchasedAttemptsUsed int               `json:"purchased_attempts_used"`
	PurchasedAttemptsCap  int               `json:"purchased_attempts_cap"`
	NextPurchasePriceGold int               `json:"next_purchase_price_gold"`
	Candidates            []RatingCandidate `json:"candidates"`
	Leaderboard           []RatingCandidate `json:"leaderboard"`
}

type RatingChallengeResult struct {
	WeekKey         string         `json:"week_key"`
	MatchID         string         `json:"match_id"`
	Result          string         `json:"result"`
	RatingDelta     int            `json:"rating_delta"`
	CharacterRating int            `json:"character_rating"`
	OpponentRating  int            `json:"opponent_rating"`
	BattleReportID  string         `json:"battle_report_id"`
	BattleReport    map[string]any `json:"battle_report"`
}

type ArenaTitleView struct {
	TitleKey      string         `json:"title_key"`
	TitleLabel    string         `json:"title_label"`
	SourceWeekKey string         `json:"source_week_key"`
	GrantedAt     string         `json:"granted_at"`
	ExpiresAt     string         `json:"expires_at"`
	BonusSnapshot map[string]any `json:"bonus_snapshot"`
}

type ratingState struct {
	Rating                     int
	FreeAttemptsRemaining      int
	PurchasedAttemptsBought    int
	PurchasedAttemptsRemaining int
	PurchasePriceStep          int
	DayKey                     string
}

type titleState struct {
	View ArenaTitleView
}

type Service struct {
	mu              sync.Mutex
	clock           func() time.Time
	entriesByDay    map[string]map[string]Entry
	ratingByWeek    map[string]map[string]ratingState
	ratingHistory   map[string][]HistoryDetail
	activeTitles    map[string]titleState
	titlesFinalized map[string]bool
}

func NewService() *Service {
	return &Service{
		clock:           time.Now,
		entriesByDay:    make(map[string]map[string]Entry),
		ratingByWeek:    make(map[string]map[string]ratingState),
		ratingHistory:   make(map[string][]HistoryDetail),
		activeTitles:    make(map[string]titleState),
		titlesFinalized: make(map[string]bool),
	}
}

func (s *Service) SetClock(clock func() time.Time) {
	if clock == nil {
		s.clock = time.Now
		return
	}
	s.clock = clock
}

func (s *Service) GetRatingBoard(characterID string, entries []Entry) (RatingBoardView, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.clock()
	weekKey := weekKeyFor(now)
	entryMap := mapEntriesByCharacter(entries)
	if _, ok := entryMap[characterID]; !ok {
		return RatingBoardView{}, ErrInvalidChallengeTarget
	}
	s.ensureWeekStatesLocked(weekKey, entries)
	s.finalizeTitlesIfNeededLocked(weekKey, entries)

	state := s.ratingByWeek[weekKey][characterID]
	leaderboard := buildRatingLeaderboard(entries, s.ratingByWeek[weekKey], 16)
	candidates := buildRatingCandidates(entries, s.ratingByWeek[weekKey], characterID, weekKey, dayKeyFor(now), 5)

	return RatingBoardView{
		WeekKey:               weekKey,
		CharacterID:           characterID,
		Rating:                state.Rating,
		FreeAttemptsRemaining: state.FreeAttemptsRemaining,
		PurchasedAttemptsUsed: state.PurchasedAttemptsBought,
		PurchasedAttemptsCap:  maxPurchasedChallengesPerDay,
		NextPurchasePriceGold: purchasePriceForStep(state.PurchasePriceStep),
		Candidates:            candidates,
		Leaderboard:           leaderboard,
	}, nil
}

func (s *Service) PurchaseRatingChallenge(characterID string) (int, RatingBoardView, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.clock()
	if !isRatingWeekday(now) {
		return 0, RatingBoardView{}, ErrChallengeWindow
	}
	weekKey := weekKeyFor(now)
	state, ok := s.ratingByWeek[weekKey][characterID]
	if !ok {
		state = newRatingState(dayKeyFor(now))
	}
	if state.PurchasedAttemptsBought >= maxPurchasedChallengesPerDay {
		return 0, RatingBoardView{}, ErrPurchaseCapReached
	}
	price := purchasePriceForStep(state.PurchasePriceStep)

	view := RatingBoardView{
		WeekKey:               weekKey,
		CharacterID:           characterID,
		Rating:                state.Rating,
		FreeAttemptsRemaining: state.FreeAttemptsRemaining,
		PurchasedAttemptsUsed: state.PurchasedAttemptsBought,
		PurchasedAttemptsCap:  maxPurchasedChallengesPerDay,
		NextPurchasePriceGold: purchasePriceForStep(state.PurchasePriceStep),
	}
	return price, view, nil
}

func (s *Service) ConfirmPurchasedRatingChallenge(characterID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.clock()
	if !isRatingWeekday(now) {
		return ErrChallengeWindow
	}
	weekKey := weekKeyFor(now)
	state, ok := s.ratingByWeek[weekKey][characterID]
	if !ok {
		state = newRatingState(dayKeyFor(now))
	}
	if state.PurchasedAttemptsBought >= maxPurchasedChallengesPerDay {
		return ErrPurchaseCapReached
	}
	state.PurchasedAttemptsBought++
	state.PurchasedAttemptsRemaining++
	state.PurchasePriceStep++
	s.ratingByWeek[weekKey][characterID] = state
	return nil
}

func (s *Service) ResolveRatingChallenge(characterID, targetCharacterID string, entries []Entry) (RatingChallengeResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.clock()
	if !isRatingWeekday(now) {
		return RatingChallengeResult{}, ErrChallengeWindow
	}
	weekKey := weekKeyFor(now)
	entryMap := mapEntriesByCharacter(entries)
	challenger, ok := entryMap[characterID]
	if !ok {
		return RatingChallengeResult{}, ErrInvalidChallengeTarget
	}
	target, ok := entryMap[targetCharacterID]
	if !ok || targetCharacterID == characterID {
		return RatingChallengeResult{}, ErrInvalidChallengeTarget
	}

	s.ensureWeekStatesLocked(weekKey, entries)
	state := s.ratingByWeek[weekKey][characterID]
	if state.FreeAttemptsRemaining+remainingPurchasedAttempts(state) <= 0 {
		return RatingChallengeResult{}, ErrNoChallengeAttempts
	}
	candidateIDs := candidateIDSet(buildRatingCandidates(entries, s.ratingByWeek[weekKey], characterID, weekKey, dayKeyFor(now), 5))
	if !candidateIDs[targetCharacterID] {
		return RatingChallengeResult{}, ErrInvalidChallengeTarget
	}

	if state.FreeAttemptsRemaining > 0 {
		state.FreeAttemptsRemaining--
	} else {
		state.PurchasedAttemptsRemaining--
		if state.PurchasedAttemptsRemaining < 0 {
			state.PurchasedAttemptsRemaining = 0
		}
	}

	matchID := fmt.Sprintf("match_weekly_%s_%s_%s", weekKey, characterID, nextChallengeSuffix(now))
	reportID := "report_" + matchID
	result := simulateEntryDuel(challenger, target, matchID)
	resolution := resolveArenaDuel(challenger, target, result, arenaDuelModeRating)
	delta := ratingDeltaForWin(challenger, target)
	outcome := "loss"
	if resolution.Winner.CharacterID == challenger.CharacterID {
		outcome = "win"
		state.Rating += delta
		targetState := s.ratingByWeek[weekKey][targetCharacterID]
		targetState.Rating = maxInt(0, targetState.Rating-delta)
		s.ratingByWeek[weekKey][targetCharacterID] = targetState
	}
	s.ratingByWeek[weekKey][characterID] = state

	challengerDetail, targetDetail := buildRatingHistoryDetails(weekKey, challenger, target, matchID, reportID, result, outcome, delta, now)
	s.ratingHistory[characterID] = prependHistoryDetail(s.ratingHistory[characterID], challengerDetail)
	s.ratingHistory[targetCharacterID] = prependHistoryDetail(s.ratingHistory[targetCharacterID], targetDetail)

	return RatingChallengeResult{
		WeekKey:         weekKey,
		MatchID:         matchID,
		Result:          outcome,
		RatingDelta:     challengerDetail.BattleReport["rating_delta"].(int),
		CharacterRating: s.ratingByWeek[weekKey][characterID].Rating,
		OpponentRating:  s.ratingByWeek[weekKey][targetCharacterID].Rating,
		BattleReportID:  reportID,
		BattleReport:    challengerDetail.BattleReport,
	}, nil
}

func (s *Service) GetArenaTitle(characterID string, entries []Entry) (ArenaTitleView, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	weekKey := weekKeyFor(s.clock())
	s.ensureWeekStatesLocked(weekKey, entries)
	s.finalizeTitlesIfNeededLocked(weekKey, entries)
	title, ok := s.activeTitles[characterID]
	if !ok {
		return ArenaTitleView{}, false
	}
	if expiresAt, err := time.Parse(time.RFC3339, title.View.ExpiresAt); err == nil && s.clock().After(expiresAt) {
		delete(s.activeTitles, characterID)
		return ArenaTitleView{}, false
	}
	return title.View, true
}

func (s *Service) Signup(character characters.Summary, panelPowerScore, equipmentScore int, arenaStatus world.ArenaStatus) (Entry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if arenaStatus.Code != "signup_open" && arenaStatus.Code != "rating_open" {
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
		PanelPowerScore: panelPowerScore,
		EquipmentScore:  equipmentScore,
		SignedUpAt:      s.clock().Format(time.RFC3339),
	}
	s.entriesByDay[dayKey][character.CharacterID] = entry
	return entry, nil
}

func (s *Service) GetCurrent(arenaStatus world.ArenaStatus, entries []Entry) CurrentView {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.clock()
	dayKey := dayKeyFor(now)
	weekKey := weekKeyFor(now)
	s.ensureWeekStatesLocked(weekKey, entries)
	s.finalizeTitlesIfNeededLocked(weekKey, entries)

	displayEntries := featuredEntriesByRating(entries, s.ratingByWeek[weekKey], len(entries))
	featured := featuredEntries(displayEntries, 8)
	snapshot := tournamentSnapshot{}
	npcCount := 0
	switch arenaStatus.Code {
	case "knockout_pending", "knockout_in_progress", "knockout_results_live", "signup_locked", "in_progress", "results_live":
		displayEntries, npcCount = knockoutEntriesForWeekLocked(dayKey, entries, s.ratingByWeek[weekKey])
		snapshot = buildTournamentSnapshot(dayKey, now, displayEntries)
		featured = featuredEntries(displayEntries, 8)
	}

	return CurrentView{
		TournamentID:        tournamentIDForDay(dayKey),
		DayKey:              dayKey,
		WeekKey:             weekKey,
		Status:              arenaStatus,
		SignupCount:         len(displayEntries),
		QualifiedCount:      len(snapshot.mainField),
		NPCCount:            npcCount,
		HighestPower:        highestPanelPower(displayEntries),
		LowestPower:         lowestPanelPower(displayEntries),
		MedianPower:         medianPanelPower(displayEntries),
		FeaturedEntries:     featured,
		QualifierMatchups:   snapshot.currentQualifierMatchups,
		QualifierRounds:     snapshot.qualifierRounds,
		Matchups:            snapshot.currentMatchups,
		Rounds:              snapshot.rounds,
		Champion:            snapshot.champion,
		NextRoundTime:       determineNextRoundTime(now, arenaStatus.Code, snapshot).Format(time.RFC3339),
		WeeklyRatingSummary: buildWeeklyRatingSummary(entries, s.ratingByWeek[weekKey]),
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
	for _, detail := range s.ratingHistory[characterID] {
		items = append(items, detail.HistorySummary)
	}
	items = filterHistory(items, filters)
	return paginateHistory(items, filters.Cursor, filters.Limit)
}

func (s *Service) GetHistoryDetail(characterID, matchID, detailLevel string) (HistoryDetail, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, item := range s.ratingHistory[characterID] {
		if item.MatchID == matchID {
			return historyDetailByLevel(item, detailLevel), nil
		}
	}

	match, round, found := s.findMatchLocked(characterID, matchID)
	if !found {
		return HistoryDetail{}, ErrArenaMatchNotFound
	}
	return buildHistoryDetail(characterID, round, match, detailLevel), nil
}

func (s *Service) GetPublicMatchDetail(matchID, detailLevel string) (PublicMatchDetail, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, items := range s.ratingHistory {
		for _, item := range items {
			if item.MatchID != matchID {
				continue
			}
			detail := historyDetailByLevel(item, detailLevel)
			return PublicMatchDetail{
				MatchID:        detail.MatchID,
				BattleReportID: detail.BattleReportID,
				TournamentID:   detail.TournamentID,
				DayKey:         detail.DayKey,
				Stage:          detail.Stage,
				RoundNumber:    detail.RoundNumber,
				RoundName:      detail.RoundName,
				Status:         "resolved",
				SummaryTag:     detail.SummaryTag,
				StartedAt:      detail.StartedAt,
				ResolvedAt:     detail.ResolvedAt,
				BattleReport:   detail.BattleReport,
				BattleLog:      detail.BattleLog,
			}, nil
		}
	}

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
	result := append([]Entry(nil), entries...)
	for i := len(entries); i < target; i++ {
		result = append(result, Entry{
			CharacterID:     fmt.Sprintf("npc_%s_%02d", dayKey, i+1),
			CharacterName:   fmt.Sprintf("Arena NPC %02d", i+1),
			Class:           npcClassFor(i),
			WeaponStyle:     npcWeaponFor(i),
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
	resolution := resolveArenaDuel(left, right, result, arenaDuelModeKnockout)
	return resolution.Winner, arenaReportID(dayKey, stage, roundNumber, matchNumber)
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
	if statusCode == "rating_open" || statusCode == "signup_open" {
		return nextKnockoutStartTime(now)
	}
	if statusCode == "rest_day" {
		return nextRatingWeekStart(now)
	}
	if statusCode == "knockout_pending" {
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

func nextKnockoutStartTime(now time.Time) time.Time {
	daysUntilSaturday := (int(time.Saturday) - int(now.Weekday()) + 7) % 7
	target := time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, now.Location()).AddDate(0, 0, daysUntilSaturday)
	if !target.After(now) {
		target = target.AddDate(0, 0, 7)
	}
	return target
}

func nextRatingWeekStart(now time.Time) time.Time {
	daysUntilMonday := (int(time.Monday) - int(now.Weekday()) + 7) % 7
	if daysUntilMonday == 0 {
		daysUntilMonday = 7
	}
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, daysUntilMonday)
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

func featuredEntriesByRating(entries []Entry, states map[string]ratingState, limit int) []Entry {
	if len(entries) == 0 {
		return nil
	}
	leaderboard := buildRatingLeaderboard(entries, states, limit)
	if len(leaderboard) == 0 {
		return featuredEntries(entries, limit)
	}
	entryByID := mapEntriesByCharacter(entries)
	items := make([]Entry, 0, len(leaderboard))
	for _, candidate := range leaderboard {
		if entry, ok := entryByID[candidate.CharacterID]; ok {
			items = append(items, entry)
		}
	}
	return items
}

func knockoutEntriesForWeekLocked(dayKey string, entries []Entry, states map[string]ratingState) ([]Entry, int) {
	leaderboard := buildRatingLeaderboard(entries, states, mainBracketSize)
	entryByID := mapEntriesByCharacter(entries)
	qualified := make([]Entry, 0, mainBracketSize)
	for _, candidate := range leaderboard {
		if entry, ok := entryByID[candidate.CharacterID]; ok {
			qualified = append(qualified, entry)
		}
	}
	return fillNPCEntries(dayKey, qualified, mainBracketSize)
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
		MaxTurns:   arenaDuelMaxTurns,
		SideA:      combA,
		SideB:      combB,
	})
}

func resolveArenaDuel(a, b Entry, result combat.BattleResult, mode arenaDuelMode) arenaDuelResolution {
	if arenaBattleTimedOut(result) {
		if mode == arenaDuelModeRating {
			return arenaDuelResolution{
				Winner:       b,
				Loser:        a,
				WinnerHP:     result.SideBFinalHP,
				EndReason:    "round_cap",
				Adjudication: "challenger_loss_at_cap",
			}
		}
		if result.SideAFinalHP > result.SideBFinalHP {
			return arenaDuelResolution{
				Winner:       a,
				Loser:        b,
				WinnerHP:     result.SideAFinalHP,
				EndReason:    "round_cap",
				Adjudication: "higher_remaining_hp",
			}
		}
		if result.SideBFinalHP > result.SideAFinalHP {
			return arenaDuelResolution{
				Winner:       b,
				Loser:        a,
				WinnerHP:     result.SideBFinalHP,
				EndReason:    "round_cap",
				Adjudication: "higher_remaining_hp",
			}
		}
		if a.CharacterID <= b.CharacterID {
			return arenaDuelResolution{
				Winner:       a,
				Loser:        b,
				WinnerHP:     result.SideAFinalHP,
				EndReason:    "round_cap",
				Adjudication: "lower_character_id",
			}
		}
		return arenaDuelResolution{
			Winner:       b,
			Loser:        a,
			WinnerHP:     result.SideBFinalHP,
			EndReason:    "round_cap",
			Adjudication: "lower_character_id",
		}
	}

	if result.SideAWon {
		return arenaDuelResolution{Winner: a, Loser: b, WinnerHP: result.SideAFinalHP, EndReason: "defeat", Adjudication: "direct_victory"}
	}
	if result.SideAFinalHP <= 0 && result.SideBFinalHP > 0 {
		return arenaDuelResolution{Winner: b, Loser: a, WinnerHP: result.SideBFinalHP, EndReason: "defeat", Adjudication: "direct_victory"}
	}
	return arenaDuelResolution{Winner: b, Loser: a, WinnerHP: result.SideBFinalHP, EndReason: "defeat", Adjudication: "direct_victory"}
}

func arenaBattleTimedOut(result combat.BattleResult) bool {
	for i := len(result.Log) - 1; i >= 0; i-- {
		entry := result.Log[i]
		if entry["event_type"] != "room_end" {
			continue
		}
		value, _ := entry["result"].(string)
		return value == "timeout"
	}
	return false
}

func arenaResolutionSnapshot(left, right Entry, result combat.BattleResult, mode arenaDuelMode) arenaDuelResolution {
	return resolveArenaDuel(left, right, result, mode)
}

func buildCombatantFromEntry(entry Entry, team string) combat.Combatant {
	comb := combat.BaselineCombatant(entry.Class)
	comb.EntityID = entry.CharacterID
	comb.Name = entry.CharacterName
	comb.Team = team
	comb.IsPlayerSide = true
	comb.PotionBag = combat.DefaultPotionBag()
	referencePower := 6200.0
	scoreFactor := float64(entry.PanelPowerScore) / referencePower
	scoreFactor = math.Max(0.72, math.Min(1.45, scoreFactor))
	applyFactor := scoreFactor
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
	resolution := arenaResolutionSnapshot(*match.LeftEntry, *match.RightEntry, result, arenaDuelModeKnockout)
	report := map[string]any{
		"outcome":              historyResultForMatch(match, characterID),
		"winner":               match.WinnerEntry,
		"loser":                loserFromMatch(match),
		"winner_final_hp":      winnerFinalHP(match, result),
		"left_final_hp":        result.SideAFinalHP,
		"right_final_hp":       result.SideBFinalHP,
		"end_reason":           resolution.EndReason,
		"adjudication":         resolution.Adjudication,
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
	resolution := arenaResolutionSnapshot(*match.LeftEntry, *match.RightEntry, result, arenaDuelModeKnockout)
	report := map[string]any{
		"outcome":              "resolved",
		"winner":               match.WinnerEntry,
		"loser":                loserFromMatch(match),
		"winner_final_hp":      winnerFinalHP(match, result),
		"left_final_hp":        result.SideAFinalHP,
		"right_final_hp":       result.SideBFinalHP,
		"end_reason":           resolution.EndReason,
		"adjudication":         resolution.Adjudication,
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

func historyDetailByLevel(detail HistoryDetail, detailLevel string) HistoryDetail {
	if detailLevel == "" || detailLevel == "standard" {
		copyDetail := detail
		copyDetail.BattleLog = nil
		return copyDetail
	}
	if detailLevel == "compact" {
		return HistoryDetail{HistorySummary: detail.HistorySummary}
	}
	return detail
}

func weekKeyFor(now time.Time) string {
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -(weekday - 1))
	return start.Format("2006-01-02")
}

func isRatingWeekday(now time.Time) bool {
	switch now.Weekday() {
	case time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday:
		return true
	default:
		return false
	}
}

func isSunday(now time.Time) bool {
	return now.Weekday() == time.Sunday
}

func newRatingState(dayKey string) ratingState {
	return ratingState{
		Rating:                     1000,
		FreeAttemptsRemaining:      freeRatingChallengesPerDay,
		PurchasedAttemptsBought:    0,
		PurchasedAttemptsRemaining: 0,
		PurchasePriceStep:          0,
		DayKey:                     dayKey,
	}
}

func purchasePriceForStep(step int) int {
	if step < 0 {
		step = 0
	}
	return basePurchasePriceGold + step*step*15 + step*20
}

func remainingPurchasedAttempts(state ratingState) int {
	return state.PurchasedAttemptsRemaining
}

func mapEntriesByCharacter(entries []Entry) map[string]Entry {
	items := make(map[string]Entry, len(entries))
	for _, entry := range entries {
		items[entry.CharacterID] = entry
	}
	return items
}

func (s *Service) ensureWeekStatesLocked(weekKey string, entries []Entry) {
	if _, ok := s.ratingByWeek[weekKey]; !ok {
		s.ratingByWeek[weekKey] = make(map[string]ratingState)
	}
	dayKey := dayKeyFor(s.clock())
	for _, entry := range entries {
		state, ok := s.ratingByWeek[weekKey][entry.CharacterID]
		if !ok {
			s.ratingByWeek[weekKey][entry.CharacterID] = newRatingState(dayKey)
			continue
		}
		if state.DayKey != dayKey {
			state.DayKey = dayKey
			state.FreeAttemptsRemaining = freeRatingChallengesPerDay
			state.PurchasedAttemptsBought = 0
			state.PurchasedAttemptsRemaining = 0
			state.PurchasePriceStep = 0
			s.ratingByWeek[weekKey][entry.CharacterID] = state
		}
	}
}

func buildRatingLeaderboard(entries []Entry, states map[string]ratingState, limit int) []RatingCandidate {
	items := make([]RatingCandidate, 0, len(entries))
	for _, entry := range entries {
		state, ok := states[entry.CharacterID]
		if !ok {
			continue
		}
		items = append(items, ratingCandidateFromEntry(entry, state.Rating))
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Rating != items[j].Rating {
			return items[i].Rating > items[j].Rating
		}
		if items[i].PanelPowerScore != items[j].PanelPowerScore {
			return items[i].PanelPowerScore > items[j].PanelPowerScore
		}
		return items[i].CharacterID < items[j].CharacterID
	})
	if limit > 0 && len(items) > limit {
		return items[:limit]
	}
	return items
}

func buildRatingCandidates(entries []Entry, states map[string]ratingState, characterID, weekKey, dayKey string, limit int) []RatingCandidate {
	self, ok := states[characterID]
	if !ok {
		return nil
	}
	items := make([]RatingCandidate, 0, len(entries))
	for _, entry := range entries {
		if entry.CharacterID == characterID {
			continue
		}
		state, ok := states[entry.CharacterID]
		if !ok {
			continue
		}
		candidate := ratingCandidateFromEntry(entry, state.Rating)
		items = append(items, candidate)
	}
	sort.Slice(items, func(i, j int) bool {
		leftGap := absInt(items[i].Rating - self.Rating)
		rightGap := absInt(items[j].Rating - self.Rating)
		if leftGap != rightGap {
			return leftGap < rightGap
		}
		return items[i].CharacterID < items[j].CharacterID
	})
	if len(items) > 15 {
		items = items[:15]
	}
	shuffled := shuffleRatingCandidates(seedForKey(weekKey, "cand-"+characterID+"-"+dayKey), items)
	if limit > 0 && len(shuffled) > limit {
		return shuffled[:limit]
	}
	return shuffled
}

func shuffleRatingCandidates(seed int64, items []RatingCandidate) []RatingCandidate {
	cloned := append([]RatingCandidate(nil), items...)
	rng := rand.New(rand.NewSource(seed))
	rng.Shuffle(len(cloned), func(i, j int) {
		cloned[i], cloned[j] = cloned[j], cloned[i]
	})
	return cloned
}

func ratingCandidateFromEntry(entry Entry, rating int) RatingCandidate {
	return RatingCandidate{
		CharacterID:     entry.CharacterID,
		CharacterName:   entry.CharacterName,
		Class:           entry.Class,
		WeaponStyle:     entry.WeaponStyle,
		Rating:          rating,
		PanelPowerScore: entry.PanelPowerScore,
		EquipmentScore:  entry.EquipmentScore,
		IsNPC:           entry.IsNPC,
	}
}

func candidateIDSet(items []RatingCandidate) map[string]bool {
	set := make(map[string]bool, len(items))
	for _, item := range items {
		set[item.CharacterID] = true
	}
	return set
}

func ratingDeltaForWin(challenger, target Entry) int {
	diff := target.PanelPowerScore - challenger.PanelPowerScore
	switch {
	case diff >= 900:
		return 30
	case diff >= 450:
		return 25
	case diff >= 150:
		return 21
	case diff > -150:
		return 18
	case diff > -450:
		return 15
	default:
		return 12
	}
}

func buildRatingHistoryDetails(weekKey string, challenger, target Entry, matchID, reportID string, result combat.BattleResult, outcome string, delta int, now time.Time) (HistoryDetail, HistoryDetail) {
	report, log := buildRatingBattleReport(challenger, target, result, outcome, delta)
	timestamp := now.Format(time.RFC3339)
	challengerSummary := HistorySummary{
		MatchID:        matchID,
		BattleReportID: reportID,
		TournamentID:   "arena_week_" + weekKey,
		DayKey:         weekKey,
		Stage:          "rating",
		RoundNumber:    0,
		RoundName:      "Rating Challenge",
		Result:         outcome,
		SummaryTag:     "rating_duel",
		StartedAt:      timestamp,
		ResolvedAt:     timestamp,
		Opponent:       historyOpponentFromEntry(&target),
	}
	targetResult := "stable"
	targetDelta := 0
	if outcome == "win" {
		targetResult = "loss"
		targetDelta = -delta
	}
	targetReport := cloneMap(report)
	targetReport["rating_delta"] = targetDelta
	targetSummary := HistorySummary{
		MatchID:        matchID,
		BattleReportID: reportID,
		TournamentID:   "arena_week_" + weekKey,
		DayKey:         weekKey,
		Stage:          "rating",
		RoundNumber:    0,
		RoundName:      "Rating Defense",
		Result:         targetResult,
		SummaryTag:     "rating_duel",
		StartedAt:      timestamp,
		ResolvedAt:     timestamp,
		Opponent:       historyOpponentFromEntry(&challenger),
	}
	return HistoryDetail{HistorySummary: challengerSummary, BattleReport: report, BattleLog: log},
		HistoryDetail{HistorySummary: targetSummary, BattleReport: targetReport, BattleLog: log}
}

func buildRatingBattleReport(challenger, target Entry, result combat.BattleResult, outcome string, delta int) (map[string]any, []map[string]any) {
	resolution := arenaResolutionSnapshot(challenger, target, result, arenaDuelModeRating)
	report := map[string]any{
		"outcome":          outcome,
		"rating_delta":     delta,
		"challenger":       challenger.CharacterName,
		"defender":         target.CharacterName,
		"challenger_power": challenger.PanelPowerScore,
		"defender_power":   target.PanelPowerScore,
		"left_final_hp":    result.SideAFinalHP,
		"right_final_hp":   result.SideBFinalHP,
		"winner_final_hp":  resolution.WinnerHP,
		"end_reason":       resolution.EndReason,
		"adjudication":     resolution.Adjudication,
		"summary_tag":      "rating_duel",
	}
	if outcome != "win" {
		report["rating_delta"] = 0
	}
	return report, result.Log
}

func cloneMap(input map[string]any) map[string]any {
	if input == nil {
		return nil
	}
	out := make(map[string]any, len(input))
	for k, v := range input {
		out[k] = v
	}
	return out
}

func prependHistoryDetail(items []HistoryDetail, item HistoryDetail) []HistoryDetail {
	result := make([]HistoryDetail, 0, len(items)+1)
	result = append(result, item)
	return append(result, items...)
}

func nextChallengeSuffix(now time.Time) string {
	return now.Format("150405")
}

func (s *Service) finalizeTitlesIfNeededLocked(weekKey string, entries []Entry) {
	now := s.clock()
	if !isSunday(now) || s.titlesFinalized[weekKey] {
		return
	}
	s.ensureWeekStatesLocked(weekKey, entries)
	saturday := saturdayReferenceTimeForWeek(weekKey, now.Location())
	dayKey := dayKeyFor(saturday)
	qualified, _ := knockoutEntriesForWeekLocked(dayKey, entries, s.ratingByWeek[weekKey])
	snapshot := buildTournamentSnapshot(dayKey, saturday, qualified)
	titleViews := buildTitleViewsFromSnapshot(weekKey, snapshot)
	if len(titleViews) == 0 {
		return
	}
	grantedAt := now.Format(time.RFC3339)
	expiresAt := now.Add(7 * 24 * time.Hour).Format(time.RFC3339)
	for characterID, view := range titleViews {
		view.GrantedAt = grantedAt
		view.ExpiresAt = expiresAt
		s.activeTitles[characterID] = titleState{View: view}
	}
	s.titlesFinalized[weekKey] = true
}

func saturdayReferenceTimeForWeek(weekKey string, loc *time.Location) time.Time {
	start, err := time.ParseInLocation("2006-01-02", weekKey, loc)
	if err != nil {
		return time.Now().In(loc)
	}
	return start.AddDate(0, 0, 5).Add(23*time.Hour + 59*time.Minute)
}

func buildTitleViewsFromSnapshot(weekKey string, snapshot tournamentSnapshot) map[string]ArenaTitleView {
	views := make(map[string]ArenaTitleView)
	for _, round := range snapshot.rounds {
		for _, match := range round.Matchups {
			if match.Status != "resolved" {
				continue
			}
			loser := loserEntry(match)
			switch round.EntrantCount {
			case 32:
				if loser != nil {
					views[loser.CharacterID] = newArenaTitleView(weekKey, "arena_top_32", "Arena Top 32", titleBonusSnapshot("top_32"))
				}
			case 16:
				if loser != nil {
					views[loser.CharacterID] = newArenaTitleView(weekKey, "arena_top_16", "Arena Top 16", titleBonusSnapshot("top_16"))
				}
			case 8:
				if loser != nil {
					views[loser.CharacterID] = newArenaTitleView(weekKey, "arena_top_8", "Arena Top 8", titleBonusSnapshot("top_8"))
				}
			case 4:
				if loser != nil {
					views[loser.CharacterID] = newArenaTitleView(weekKey, "arena_top_4", "Arena Top 4", titleBonusSnapshot("top_4"))
				}
			case 2:
				if loser != nil {
					views[loser.CharacterID] = newArenaTitleView(weekKey, "arena_runner_up", "Arena Runner-up", titleBonusSnapshot("runner_up"))
				}
				if match.WinnerEntry != nil {
					views[match.WinnerEntry.CharacterID] = newArenaTitleView(weekKey, "arena_champion", "Arena Champion", titleBonusSnapshot("champion"))
				}
			}
		}
	}
	return views
}

func newArenaTitleView(weekKey, titleKey, titleLabel string, bonus map[string]any) ArenaTitleView {
	return ArenaTitleView{
		TitleKey:      titleKey,
		TitleLabel:    titleLabel,
		SourceWeekKey: weekKey,
		BonusSnapshot: bonus,
	}
}

func loserEntry(match Matchup) *Entry {
	if match.WinnerEntry == nil {
		return nil
	}
	if match.LeftEntry != nil && match.LeftEntry.CharacterID != match.WinnerEntry.CharacterID {
		return match.LeftEntry
	}
	if match.RightEntry != nil && match.RightEntry.CharacterID != match.WinnerEntry.CharacterID {
		return match.RightEntry
	}
	return nil
}

func titleBonusSnapshot(tier string) map[string]any {
	switch tier {
	case "champion":
		return map[string]any{"max_hp": 0.09, "physical_attack": 0.09, "magic_attack": 0.09, "physical_defense": 0.09, "magic_defense": 0.09, "healing_power": 0.09, "speed": 0.04, "crit_rate": 0.025, "crit_damage": 0.04, "block_rate": 0.016, "precision": 0.016, "evasion_rate": 0.016, "physical_mastery": 0.04, "magic_mastery": 0.04}
	case "runner_up":
		return map[string]any{"max_hp": 0.065, "physical_attack": 0.065, "magic_attack": 0.065, "physical_defense": 0.065, "magic_defense": 0.065, "healing_power": 0.065, "speed": 0.03, "crit_rate": 0.018, "crit_damage": 0.03, "block_rate": 0.012, "precision": 0.012, "evasion_rate": 0.012, "physical_mastery": 0.03, "magic_mastery": 0.03}
	case "top_4":
		return map[string]any{"max_hp": 0.05, "physical_attack": 0.05, "magic_attack": 0.05, "physical_defense": 0.05, "magic_defense": 0.05, "healing_power": 0.05, "speed": 0.025, "crit_rate": 0.014, "crit_damage": 0.025, "block_rate": 0.01, "precision": 0.01, "evasion_rate": 0.01, "physical_mastery": 0.025, "magic_mastery": 0.025}
	case "top_8":
		return map[string]any{"max_hp": 0.04, "physical_attack": 0.04, "magic_attack": 0.04, "physical_defense": 0.04, "magic_defense": 0.04, "healing_power": 0.04, "speed": 0.02, "crit_rate": 0.011, "crit_damage": 0.02, "block_rate": 0.008, "precision": 0.008, "evasion_rate": 0.008, "physical_mastery": 0.02, "magic_mastery": 0.02}
	case "top_16":
		return map[string]any{"max_hp": 0.03, "physical_attack": 0.03, "magic_attack": 0.03, "physical_defense": 0.03, "magic_defense": 0.03, "healing_power": 0.03, "speed": 0.015, "crit_rate": 0.008, "crit_damage": 0.015, "block_rate": 0.006, "precision": 0.006, "evasion_rate": 0.006, "physical_mastery": 0.015, "magic_mastery": 0.015}
	default:
		return map[string]any{"max_hp": 0.02, "physical_attack": 0.02, "magic_attack": 0.02, "physical_defense": 0.02, "magic_defense": 0.02, "healing_power": 0.02, "speed": 0.01, "crit_rate": 0.005, "crit_damage": 0.01, "block_rate": 0.004, "precision": 0.004, "evasion_rate": 0.004, "physical_mastery": 0.01, "magic_mastery": 0.01}
	}
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func buildWeeklyRatingSummary(entries []Entry, states map[string]ratingState) map[string]any {
	if len(states) == 0 {
		return map[string]any{}
	}
	leaderboard := buildRatingLeaderboard(entries, states, 8)
	highest := 0
	lowest := 0
	if len(leaderboard) > 0 {
		highest = leaderboard[0].Rating
		lowest = leaderboard[len(leaderboard)-1].Rating
	}
	return map[string]any{
		"active_count":   len(leaderboard),
		"highest_rating": highest,
		"lowest_rating":  lowest,
		"featured":       leaderboard,
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
		PanelPowerScore: entry.PanelPowerScore,
		IsNPC:           entry.IsNPC,
	}
}

func detailEntryFromReport(report map[string]any, role string) *Entry {
	if report == nil {
		return nil
	}

	var nameKey string
	switch role {
	case "challenger":
		nameKey = "challenger"
	case "opponent":
		nameKey = "defender"
	case "winner":
		outcome, _ := report["outcome"].(string)
		switch outcome {
		case "win":
			nameKey = "challenger"
		case "loss":
			nameKey = "defender"
		default:
			return nil
		}
	default:
		return nil
	}

	name, _ := report[nameKey].(string)
	if name == "" {
		return nil
	}

	return &Entry{CharacterName: name}
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
// class-baseline stats and the shared starter potion bag.
func (s *Service) SimulateDuel(a, b characters.Summary) DuelResult {
	combA := combat.BaselineCombatant(a.Class)
	combA.EntityID = a.CharacterID
	combA.Name = a.Name
	combA.Team = "a"
	combA.IsPlayerSide = true
	combA.CurrentHP = combA.MaxHP
	combA.PotionBag = combat.DefaultPotionBag()

	combB := combat.BaselineCombatant(b.Class)
	combB.EntityID = b.CharacterID
	combB.Name = b.Name
	combB.Team = "b"
	combB.IsPlayerSide = true
	combB.CurrentHP = combB.MaxHP
	combB.PotionBag = combat.DefaultPotionBag()

	battleID := fmt.Sprintf("duel_%d", s.clock().UnixNano())
	result := combat.SimulateBattle(combat.BattleConfig{
		BattleType: "arena_duel",
		RunID:      battleID,
		RoomIndex:  1,
		MaxTurns:   arenaDuelMaxTurns,
		SideA:      combA,
		SideB:      combB,
	})
	resolution := resolveArenaDuel(
		Entry{CharacterID: a.CharacterID, CharacterName: a.Name},
		Entry{CharacterID: b.CharacterID, CharacterName: b.Name},
		result,
		arenaDuelModeExhibition,
	)

	if resolution.Winner.CharacterID == a.CharacterID {
		return DuelResult{
			BattleID:   battleID,
			WinnerID:   a.CharacterID,
			WinnerName: a.Name,
			LoserID:    b.CharacterID,
			LoserName:  b.Name,
			WinnerHP:   resolution.WinnerHP,
			BattleLog:  result.Log,
		}
	}
	return DuelResult{
		BattleID:   battleID,
		WinnerID:   b.CharacterID,
		WinnerName: b.Name,
		LoserID:    a.CharacterID,
		LoserName:  a.Name,
		WinnerHP:   resolution.WinnerHP,
		BattleLog:  result.Log,
	}
}
