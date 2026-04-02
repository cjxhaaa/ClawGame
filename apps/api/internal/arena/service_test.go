package arena

import (
	"fmt"
	"testing"
	"time"

	"clawgame/apps/api/internal/characters"
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
		Rank:        "mid",
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
		"char_1": {CharacterID: "char_1", CharacterName: "NovaScript", Class: "warrior", WeaponStyle: "great_axe", Rank: "mid", PanelPowerScore: 6100, EquipmentScore: 300, SignedUpAt: now.Add(-3 * time.Minute).Format(time.RFC3339)},
		"char_2": {CharacterID: "char_2", CharacterName: "LyraLoop", Class: "priest", WeaponStyle: "holy_tome", Rank: "mid", PanelPowerScore: 6400, EquipmentScore: 340, SignedUpAt: now.Add(-2 * time.Minute).Format(time.RFC3339)},
		"char_3": {CharacterID: "char_3", CharacterName: "KiroNode", Class: "mage", WeaponStyle: "spellbook", Rank: "high", PanelPowerScore: 9800, EquipmentScore: 380, SignedUpAt: now.Add(-1 * time.Minute).Format(time.RFC3339)},
	}

	view := svc.GetCurrent(world.ArenaStatus{Code: "in_progress"})
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

func TestQualifierRoundReducesFieldTo64(t *testing.T) {
	svc := NewService()
	now := time.Date(2026, time.April, 1, 9, 12, 0, 0, time.FixedZone("CST", 8*60*60))
	svc.clock = func() time.Time { return now }

	dayKey := dayKeyFor(now)
	entries := make(map[string]Entry)
	for i := 0; i < 1000; i++ {
		id := fmt.Sprintf("char_%03d", i+1)
		entries[id] = Entry{
			CharacterID:     id,
			CharacterName:   fmt.Sprintf("Bot %03d", i+1),
			Class:           npcClassFor(i),
			WeaponStyle:     npcWeaponFor(i),
			Rank:            "mid",
			PanelPowerScore: 6200 + ((i % 20) * 40),
			EquipmentScore:  320 + (i % 20),
			SignedUpAt:      now.Add(-time.Duration(i) * time.Second).Format(time.RFC3339),
		}
	}
	svc.entriesByDay[dayKey] = entries

	view := svc.GetCurrent(world.ArenaStatus{Code: "in_progress"})
	if len(view.QualifierRounds) < 2 {
		t.Fatalf("expected multiple qualifier rounds for a large signup pool, got %d", len(view.QualifierRounds))
	}
	if view.QualifiedCount != 64 {
		t.Fatalf("expected qualifier stage to produce 64 qualifiers, got %d", view.QualifiedCount)
	}
	firstRoundEntrants := 0
	for _, match := range view.QualifierRounds[0].Matchups {
		if match.ByeEntry != nil {
			firstRoundEntrants++
		}
		if match.LeftEntry != nil {
			firstRoundEntrants++
		}
		if match.RightEntry != nil {
			firstRoundEntrants++
		}
	}
	if firstRoundEntrants != 1000 {
		t.Fatalf("expected first qualifier round to cover all 1000 entrants, got %d", firstRoundEntrants)
	}
}

func TestChampionAvailableAfterFinalWindow(t *testing.T) {
	svc := NewService()
	now := time.Date(2026, time.April, 1, 9, 36, 0, 0, time.FixedZone("CST", 8*60*60))
	svc.clock = func() time.Time { return now }

	dayKey := dayKeyFor(now)
	svc.entriesByDay[dayKey] = map[string]Entry{
		"char_1": {CharacterID: "char_1", CharacterName: "NovaScript", Class: "warrior", WeaponStyle: "great_axe", Rank: "high", PanelPowerScore: 10200, EquipmentScore: 420, SignedUpAt: now.Add(-5 * time.Minute).Format(time.RFC3339)},
		"char_2": {CharacterID: "char_2", CharacterName: "LyraLoop", Class: "priest", WeaponStyle: "holy_tome", Rank: "high", PanelPowerScore: 9900, EquipmentScore: 410, SignedUpAt: now.Add(-4 * time.Minute).Format(time.RFC3339)},
	}

	view := svc.GetCurrent(world.ArenaStatus{Code: "results_live"})
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
		"char_1": {CharacterID: "char_1", CharacterName: "NovaScript", Class: "warrior", WeaponStyle: "great_axe", Rank: "mid", PanelPowerScore: 6500, EquipmentScore: 330, SignedUpAt: now.Add(-2 * time.Hour).Format(time.RFC3339)},
		"char_2": {CharacterID: "char_2", CharacterName: "LyraLoop", Class: "priest", WeaponStyle: "holy_tome", Rank: "mid", PanelPowerScore: 6400, EquipmentScore: 328, SignedUpAt: now.Add(-2 * time.Hour).Format(time.RFC3339)},
		"char_3": {CharacterID: "char_3", CharacterName: "KiroNode", Class: "mage", WeaponStyle: "spellbook", Rank: "mid", PanelPowerScore: 6300, EquipmentScore: 320, SignedUpAt: now.Add(-2 * time.Hour).Format(time.RFC3339)},
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
