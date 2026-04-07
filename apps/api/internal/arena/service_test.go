package arena

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"clawgame/apps/api/internal/characters"
	"clawgame/apps/api/internal/combat"
	"clawgame/apps/api/internal/world"
)

func TestSignupIsScopedPerDay(t *testing.T) {
	svc := NewService()
	now := time.Date(2026, time.April, 1, 8, 30, 0, 0, time.FixedZone("CST", 8*60*60))
	svc.clock = func() time.Time { return now }

	character := characters.Summary{
		CharacterID: "char_1",
		Name:        "NovaScript",
		Class:       "warrior",
		WeaponStyle: "great_axe",
	}
	status := world.ArenaStatus{Code: "signup_open"}

	if _, err := svc.Signup(character, 6200, 320, status); err != nil {
		t.Fatalf("expected first signup to succeed, got %v", err)
	}
	if _, err := svc.Signup(character, 6200, 320, status); err != ErrAlreadySignedUp {
		t.Fatalf("expected duplicate signup to fail with ErrAlreadySignedUp, got %v", err)
	}

	now = now.Add(24 * time.Hour)
	if _, err := svc.Signup(character, 6400, 325, status); err != nil {
		t.Fatalf("expected signup on next day to succeed, got %v", err)
	}
}

func TestGetCurrentPadsNPCsTo64(t *testing.T) {
	svc := NewService()
	now := time.Date(2026, time.April, 1, 9, 12, 0, 0, time.FixedZone("CST", 8*60*60))
	svc.clock = func() time.Time { return now }

	dayKey := dayKeyFor(now)
	svc.entriesByDay[dayKey] = map[string]Entry{
		"char_1": {CharacterID: "char_1", CharacterName: "NovaScript", Class: "warrior", WeaponStyle: "great_axe", PanelPowerScore: 6100, EquipmentScore: 300, SignedUpAt: now.Add(-3 * time.Minute).Format(time.RFC3339)},
		"char_2": {CharacterID: "char_2", CharacterName: "LyraLoop", Class: "priest", WeaponStyle: "holy_tome", PanelPowerScore: 6400, EquipmentScore: 340, SignedUpAt: now.Add(-2 * time.Minute).Format(time.RFC3339)},
		"char_3": {CharacterID: "char_3", CharacterName: "KiroNode", Class: "mage", WeaponStyle: "spellbook", PanelPowerScore: 9800, EquipmentScore: 380, SignedUpAt: now.Add(-1 * time.Minute).Format(time.RFC3339)},
	}

	view := svc.GetCurrent(world.ArenaStatus{Code: "knockout_in_progress"}, sortedEntries(svc.entriesByDay[dayKey]))
	if view.QualifiedCount != 64 {
		t.Fatalf("expected 64-player main field, got %d", view.QualifiedCount)
	}
	if view.NPCCount != 61 {
		t.Fatalf("expected 61 NPCs, got %d", view.NPCCount)
	}
	if len(view.Rounds) != 6 {
		t.Fatalf("expected 6 main bracket rounds, got %d", len(view.Rounds))
	}
	if view.Rounds[0].Name != "Round of 64" {
		t.Fatalf("expected first round to be Round of 64, got %q", view.Rounds[0].Name)
	}
}

func TestSaturdayKnockoutUsesTop64Seeds(t *testing.T) {
	svc := NewService()
	now := time.Date(2026, time.April, 11, 9, 12, 0, 0, time.FixedZone("CST", 8*60*60))
	svc.clock = func() time.Time { return now }

	entries := make([]Entry, 0, 1000)
	for i := 0; i < 1000; i++ {
		id := fmt.Sprintf("char_%03d", i+1)
		entry := Entry{
			CharacterID:     id,
			CharacterName:   fmt.Sprintf("Bot %03d", i+1),
			Class:           npcClassFor(i),
			WeaponStyle:     npcWeaponFor(i),
			PanelPowerScore: 6200 + ((i % 20) * 40),
			EquipmentScore:  320 + (i % 20),
			SignedUpAt:      now.Add(-time.Duration(i) * time.Second).Format(time.RFC3339),
		}
		entries = append(entries, entry)
	}
	weekKey := weekKeyFor(now)
	svc.ensureWeekStatesLocked(weekKey, entries)
	for i, entry := range entries {
		svc.ratingByWeek[weekKey][entry.CharacterID] = ratingState{
			Rating:                2000 - i,
			FreeAttemptsRemaining: 3,
			DayKey:                dayKeyFor(now),
		}
	}

	view := svc.GetCurrent(world.ArenaStatus{Code: "knockout_in_progress"}, entries)
	if len(view.QualifierRounds) != 0 {
		t.Fatalf("expected seeded weekly knockout without qualifier rounds, got %d", len(view.QualifierRounds))
	}
	if view.QualifiedCount != 64 {
		t.Fatalf("expected weekly knockout to produce 64 qualifiers, got %d", view.QualifiedCount)
	}
	if len(view.Rounds) == 0 || view.Rounds[0].EntrantCount != 64 {
		t.Fatalf("expected main bracket to start at 64 entrants, got %#v", view.Rounds)
	}
	for _, entry := range view.FeaturedEntries {
		if strings.TrimSpace(entry.CharacterID) == "" {
			t.Fatal("expected featured seeded entrants to have character ids")
		}
	}
}

func TestChampionAvailableAfterFinalWindow(t *testing.T) {
	svc := NewService()
	now := time.Date(2026, time.April, 1, 9, 36, 0, 0, time.FixedZone("CST", 8*60*60))
	svc.clock = func() time.Time { return now }

	dayKey := dayKeyFor(now)
	svc.entriesByDay[dayKey] = map[string]Entry{
		"char_1": {CharacterID: "char_1", CharacterName: "NovaScript", Class: "warrior", WeaponStyle: "great_axe", PanelPowerScore: 10200, EquipmentScore: 420, SignedUpAt: now.Add(-5 * time.Minute).Format(time.RFC3339)},
		"char_2": {CharacterID: "char_2", CharacterName: "LyraLoop", Class: "priest", WeaponStyle: "holy_tome", PanelPowerScore: 9900, EquipmentScore: 410, SignedUpAt: now.Add(-4 * time.Minute).Format(time.RFC3339)},
	}

	view := svc.GetCurrent(world.ArenaStatus{Code: "knockout_results_live"}, sortedEntries(svc.entriesByDay[dayKey]))
	if view.Champion == nil {
		t.Fatal("expected champion to be available after the final window")
	}
	if view.Rounds[len(view.Rounds)-1].Name != "Final" {
		t.Fatalf("expected last round to be Final, got %q", view.Rounds[len(view.Rounds)-1].Name)
	}
}

func TestArenaHistoryReturnsResolvedMatches(t *testing.T) {
	svc := NewService()
	now := time.Date(2026, time.April, 1, 10, 0, 0, 0, time.FixedZone("CST", 8*60*60))
	svc.clock = func() time.Time { return now }

	dayKey := dayKeyFor(now)
	svc.entriesByDay[dayKey] = map[string]Entry{
		"char_1": {CharacterID: "char_1", CharacterName: "NovaScript", Class: "warrior", WeaponStyle: "great_axe", PanelPowerScore: 6500, EquipmentScore: 330, SignedUpAt: now.Add(-2 * time.Hour).Format(time.RFC3339)},
		"char_2": {CharacterID: "char_2", CharacterName: "LyraLoop", Class: "priest", WeaponStyle: "holy_tome", PanelPowerScore: 6400, EquipmentScore: 328, SignedUpAt: now.Add(-2 * time.Hour).Format(time.RFC3339)},
		"char_3": {CharacterID: "char_3", CharacterName: "KiroNode", Class: "mage", WeaponStyle: "spellbook", PanelPowerScore: 6300, EquipmentScore: 320, SignedUpAt: now.Add(-2 * time.Hour).Format(time.RFC3339)},
	}

	items := svc.ListHistory("char_1", HistoryFilters{Limit: 20})
	if len(items) == 0 {
		t.Fatal("expected arena history to return resolved matches")
	}
	if items[0].BattleReportID == "" && items[0].Result != "bye" {
		t.Fatal("expected resolved non-bye matches to expose a battle report id")
	}

	detail, err := svc.GetHistoryDetail("char_1", items[0].MatchID, "verbose")
	if err != nil {
		t.Fatalf("expected history detail lookup to succeed, got %v", err)
	}
	if detail.Result == "bye" {
		if detail.BattleReport == nil {
			t.Fatal("expected bye detail to still provide a compact battle report")
		}
	} else {
		if len(detail.BattleLog) == 0 {
			t.Fatal("expected verbose history detail to include battle log")
		}
	}
}

func TestListEntriesPaginatesSortedField(t *testing.T) {
	svc := NewService()
	now := time.Date(2026, time.April, 1, 8, 45, 0, 0, time.FixedZone("CST", 8*60*60))
	svc.clock = func() time.Time { return now }

	dayKey := dayKeyFor(now)
	svc.entriesByDay[dayKey] = map[string]Entry{
		"char_1": {CharacterID: "char_1", CharacterName: "Alpha", PanelPowerScore: 7000, EquipmentScore: 300, SignedUpAt: now.Add(-3 * time.Minute).Format(time.RFC3339)},
		"char_2": {CharacterID: "char_2", CharacterName: "Beta", PanelPowerScore: 6800, EquipmentScore: 290, SignedUpAt: now.Add(-2 * time.Minute).Format(time.RFC3339)},
		"char_3": {CharacterID: "char_3", CharacterName: "Gamma", PanelPowerScore: 6600, EquipmentScore: 280, SignedUpAt: now.Add(-1 * time.Minute).Format(time.RFC3339)},
	}

	page1 := svc.ListEntries(EntryListFilters{Limit: 2})
	if len(page1) != 2 {
		t.Fatalf("expected 2 entries on first page, got %d", len(page1))
	}
	if page1[0].CharacterID != "char_1" || page1[1].CharacterID != "char_2" {
		t.Fatalf("expected sorted first page, got %q then %q", page1[0].CharacterID, page1[1].CharacterID)
	}

	page2 := svc.ListEntries(EntryListFilters{Cursor: page1[len(page1)-1].CharacterID, Limit: 2})
	if len(page2) != 1 || page2[0].CharacterID != "char_3" {
		t.Fatalf("expected remaining entry on second page, got %#v", page2)
	}
}

func TestRatingBoardAndChallengeFlow(t *testing.T) {
	svc := NewService()
	now := time.Date(2026, time.April, 6, 10, 0, 0, 0, time.FixedZone("CST", 8*60*60))
	svc.clock = func() time.Time { return now }

	entries := []Entry{
		{CharacterID: "char_1", CharacterName: "NovaScript", Class: "warrior", WeaponStyle: "great_axe", PanelPowerScore: 6800, EquipmentScore: 340},
		{CharacterID: "char_2", CharacterName: "LyraLoop", Class: "priest", WeaponStyle: "holy_tome", PanelPowerScore: 7000, EquipmentScore: 345},
		{CharacterID: "char_3", CharacterName: "KiroNode", Class: "mage", WeaponStyle: "spellbook", PanelPowerScore: 6600, EquipmentScore: 330},
		{CharacterID: "char_4", CharacterName: "RuneFork", Class: "warrior", WeaponStyle: "great_axe", PanelPowerScore: 6400, EquipmentScore: 320},
		{CharacterID: "char_5", CharacterName: "VelaByte", Class: "mage", WeaponStyle: "spellbook", PanelPowerScore: 7200, EquipmentScore: 350},
		{CharacterID: "char_6", CharacterName: "PatchLeaf", Class: "priest", WeaponStyle: "holy_tome", PanelPowerScore: 6900, EquipmentScore: 338},
	}

	board, err := svc.GetRatingBoard("char_1", entries)
	if err != nil {
		t.Fatalf("expected rating board to load, got %v", err)
	}
	if board.Rating != 1000 {
		t.Fatalf("expected starting rating 1000, got %d", board.Rating)
	}
	if board.FreeAttemptsRemaining != 3 {
		t.Fatalf("expected 3 free attempts, got %d", board.FreeAttemptsRemaining)
	}
	if len(board.Candidates) == 0 {
		t.Fatal("expected nearby rating candidates")
	}

	targetID := board.Candidates[0].CharacterID
	result, err := svc.ResolveRatingChallenge("char_1", targetID, entries)
	if err != nil {
		t.Fatalf("expected rating challenge to resolve, got %v", err)
	}
	if result.MatchID == "" {
		t.Fatal("expected rating challenge match id")
	}
	if result.Result == "win" && result.RatingDelta <= 0 {
		t.Fatalf("expected positive rating delta on win, got %d", result.RatingDelta)
	}

	updatedBoard, err := svc.GetRatingBoard("char_1", entries)
	if err != nil {
		t.Fatalf("expected updated rating board to load, got %v", err)
	}
	if updatedBoard.FreeAttemptsRemaining != 2 {
		t.Fatalf("expected one free attempt consumed, got %d", updatedBoard.FreeAttemptsRemaining)
	}

	history := svc.ListHistory("char_1", HistoryFilters{Limit: 10})
	found := false
	for _, item := range history {
		if item.Stage == "rating" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected rating challenge to appear in arena history")
	}
}

func TestSundayFinalizesWeeklyArenaTitlesForOneWeek(t *testing.T) {
	svc := NewService()
	now := time.Date(2026, time.April, 12, 10, 0, 0, 0, time.FixedZone("CST", 8*60*60))
	svc.clock = func() time.Time { return now }

	entries := make([]Entry, 0, 40)
	for i := 0; i < 40; i++ {
		entries = append(entries, Entry{
			CharacterID:     fmt.Sprintf("char_%02d", i+1),
			CharacterName:   fmt.Sprintf("Bot %02d", i+1),
			Class:           npcClassFor(i),
			WeaponStyle:     npcWeaponFor(i),
			PanelPowerScore: 7000 - i*10,
			EquipmentScore:  330,
		})
	}

	board, err := svc.GetRatingBoard("char_01", entries)
	if err != nil {
		t.Fatalf("expected rating board to load, got %v", err)
	}
	if board.WeekKey == "" {
		t.Fatal("expected week key")
	}

	svc.ratingByWeek[board.WeekKey]["char_01"] = ratingState{Rating: 1320, FreeAttemptsRemaining: 3, DayKey: dayKeyFor(now)}
	svc.ratingByWeek[board.WeekKey]["char_02"] = ratingState{Rating: 1280, FreeAttemptsRemaining: 3, DayKey: dayKeyFor(now)}
	for i := 2; i < len(entries); i++ {
		svc.ratingByWeek[board.WeekKey][entries[i].CharacterID] = ratingState{Rating: 1200 - i, FreeAttemptsRemaining: 3, DayKey: dayKeyFor(now)}
	}

	title, found := svc.GetArenaTitle("char_01", entries)
	if !found {
		t.Fatal("expected title to be granted on Sunday finalization")
	}
	if title.TitleKey != "arena_champion" {
		t.Fatalf("expected champion title, got %q", title.TitleKey)
	}
	expiresAt, err := time.Parse(time.RFC3339, title.ExpiresAt)
	if err != nil {
		t.Fatalf("expected valid expires_at, got %v", err)
	}
	if hours := expiresAt.Sub(now).Hours(); hours < 24*6.5 || hours > 24*7.5 {
		t.Fatalf("expected one-week duration, got %.2f hours", hours)
	}
}

func TestResolveArenaDuelRatingTimeoutCountsAsChallengerLoss(t *testing.T) {
	left := Entry{CharacterID: "char_a", CharacterName: "Challenger"}
	right := Entry{CharacterID: "char_b", CharacterName: "Defender"}
	result := timeoutBattleResult(48, 52)

	resolution := resolveArenaDuel(left, right, result, arenaDuelModeRating)
	if resolution.Winner.CharacterID != right.CharacterID {
		t.Fatalf("expected defender to win rating timeout, got %q", resolution.Winner.CharacterID)
	}
	if resolution.WinnerHP != result.SideBFinalHP {
		t.Fatalf("expected defender HP %d, got %d", result.SideBFinalHP, resolution.WinnerHP)
	}
	if resolution.EndReason != "round_cap" {
		t.Fatalf("expected round_cap end reason, got %q", resolution.EndReason)
	}
	if resolution.Adjudication != "challenger_loss_at_cap" {
		t.Fatalf("expected challenger_loss_at_cap adjudication, got %q", resolution.Adjudication)
	}
}

func TestResolveArenaDuelKnockoutTimeoutUsesRemainingHP(t *testing.T) {
	left := Entry{CharacterID: "char_a", CharacterName: "Left"}
	right := Entry{CharacterID: "char_b", CharacterName: "Right"}
	result := timeoutBattleResult(61, 33)

	resolution := resolveArenaDuel(left, right, result, arenaDuelModeKnockout)
	if resolution.Winner.CharacterID != left.CharacterID {
		t.Fatalf("expected higher-HP left entrant to win knockout timeout, got %q", resolution.Winner.CharacterID)
	}
	if resolution.WinnerHP != result.SideAFinalHP {
		t.Fatalf("expected winner HP %d, got %d", result.SideAFinalHP, resolution.WinnerHP)
	}
	if resolution.EndReason != "round_cap" {
		t.Fatalf("expected round_cap end reason, got %q", resolution.EndReason)
	}
	if resolution.Adjudication != "higher_remaining_hp" {
		t.Fatalf("expected higher_remaining_hp adjudication, got %q", resolution.Adjudication)
	}
}

func TestResolveArenaDuelKnockoutTimeoutUsesStableIDTiebreak(t *testing.T) {
	left := Entry{CharacterID: "char_a", CharacterName: "Left"}
	right := Entry{CharacterID: "char_b", CharacterName: "Right"}
	result := timeoutBattleResult(40, 40)

	resolution := resolveArenaDuel(left, right, result, arenaDuelModeKnockout)
	if resolution.Winner.CharacterID != left.CharacterID {
		t.Fatalf("expected lower character_id to win exact timeout tie, got %q", resolution.Winner.CharacterID)
	}
	if resolution.EndReason != "round_cap" {
		t.Fatalf("expected round_cap end reason, got %q", resolution.EndReason)
	}
	if resolution.Adjudication != "lower_character_id" {
		t.Fatalf("expected lower_character_id adjudication, got %q", resolution.Adjudication)
	}
}

func TestBuildRatingBattleReportIncludesArenaResolutionFields(t *testing.T) {
	challenger := Entry{CharacterID: "char_a", CharacterName: "Challenger", PanelPowerScore: 6400}
	defender := Entry{CharacterID: "char_b", CharacterName: "Defender", PanelPowerScore: 6500}
	result := timeoutBattleResult(41, 55)

	report, _ := buildRatingBattleReport(challenger, defender, result, "loss", 18)

	if report["end_reason"] != "round_cap" {
		t.Fatalf("expected round_cap end reason in report, got %#v", report["end_reason"])
	}
	if report["adjudication"] != "challenger_loss_at_cap" {
		t.Fatalf("expected challenger_loss_at_cap adjudication in report, got %#v", report["adjudication"])
	}
	if report["winner_final_hp"] != 55 {
		t.Fatalf("expected winner_final_hp 55, got %#v", report["winner_final_hp"])
	}
	if report["rating_delta"] != 0 {
		t.Fatalf("expected non-win report delta to be zeroed, got %#v", report["rating_delta"])
	}
}

func TestSimulateEntryDuelUsesTenTurnCap(t *testing.T) {
	left := Entry{CharacterID: "char_a", CharacterName: "Left", Class: "warrior", WeaponStyle: "great_axe", PanelPowerScore: 1000}
	right := Entry{CharacterID: "char_b", CharacterName: "Right", Class: "warrior", WeaponStyle: "great_axe", PanelPowerScore: 1000}

	result := simulateEntryDuel(left, right, "arena-turn-cap-test")

	turnStarts := 0
	for _, entry := range result.Log {
		if entry["event_type"] == "turn_start" {
			turnStarts++
		}
	}
	if turnStarts > arenaDuelMaxTurns {
		t.Fatalf("expected arena duel to stop by %d turns, got %d", arenaDuelMaxTurns, turnStarts)
	}
}

func timeoutBattleResult(sideAFinalHP, sideBFinalHP int) combat.BattleResult {
	return combat.BattleResult{
		SideAWon:     false,
		SideAFinalHP: sideAFinalHP,
		SideBFinalHP: sideBFinalHP,
		Log: []map[string]any{
			{
				"event_type": "room_end",
				"result":     "timeout",
			},
		},
	}
}
