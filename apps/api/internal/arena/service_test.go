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

	if _, err := svc.Signup(character, 320, status); err != nil {
		t.Fatalf("expected first signup to succeed, got %v", err)
	}
	if _, err := svc.Signup(character, 320, status); err != ErrAlreadySignedUp {
		t.Fatalf("expected duplicate signup to fail with ErrAlreadySignedUp, got %v", err)
	}

	now = now.Add(24 * time.Hour)
	if _, err := svc.Signup(character, 325, status); err != nil {
		t.Fatalf("expected signup on next day to succeed, got %v", err)
	}
}

func TestGetCurrentPadsNPCsTo64(t *testing.T) {
	svc := NewService()
	now := time.Date(2026, time.April, 1, 9, 12, 0, 0, time.FixedZone("CST", 8*60*60))
	svc.clock = func() time.Time { return now }

	dayKey := dayKeyFor(now)
	svc.entriesByDay[dayKey] = map[string]Entry{
		"char_1": {CharacterID: "char_1", CharacterName: "NovaScript", Class: "warrior", WeaponStyle: "great_axe", Rank: "mid", EquipmentScore: 300, SignedUpAt: now.Add(-3 * time.Minute).Format(time.RFC3339)},
		"char_2": {CharacterID: "char_2", CharacterName: "LyraLoop", Class: "priest", WeaponStyle: "holy_tome", Rank: "mid", EquipmentScore: 340, SignedUpAt: now.Add(-2 * time.Minute).Format(time.RFC3339)},
		"char_3": {CharacterID: "char_3", CharacterName: "KiroNode", Class: "mage", WeaponStyle: "spellbook", Rank: "high", EquipmentScore: 380, SignedUpAt: now.Add(-1 * time.Minute).Format(time.RFC3339)},
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
	now := time.Date(2026, time.April, 1, 9, 2, 0, 0, time.FixedZone("CST", 8*60*60))
	svc.clock = func() time.Time { return now }

	dayKey := dayKeyFor(now)
	entries := make(map[string]Entry)
	for i := 0; i < 80; i++ {
		id := fmt.Sprintf("char_%03d", i+1)
		entries[id] = Entry{
			CharacterID:    id,
			CharacterName:  fmt.Sprintf("Bot %03d", i+1),
			Class:          npcClassFor(i),
			WeaponStyle:    npcWeaponFor(i),
			Rank:           "mid",
			EquipmentScore: 320 + (i % 20),
			SignedUpAt:     now.Add(-time.Duration(i) * time.Second).Format(time.RFC3339),
		}
	}
	svc.entriesByDay[dayKey] = entries

	view := svc.GetCurrent(world.ArenaStatus{Code: "signup_locked"})
	if len(view.QualifierMatchups) == 0 {
		t.Fatal("expected qualifier matchups when signups exceed 64")
	}
	if view.QualifiedCount != 64 {
		t.Fatalf("expected qualifier stage to produce 64 qualifiers, got %d", view.QualifiedCount)
	}
	if len(view.Matchups) != len(view.QualifierMatchups) {
		t.Fatal("expected current matchups to point at qualifier matches during qualifier window")
	}
}

func TestChampionAvailableAfterFinalWindow(t *testing.T) {
	svc := NewService()
	now := time.Date(2026, time.April, 1, 9, 36, 0, 0, time.FixedZone("CST", 8*60*60))
	svc.clock = func() time.Time { return now }

	dayKey := dayKeyFor(now)
	svc.entriesByDay[dayKey] = map[string]Entry{
		"char_1": {CharacterID: "char_1", CharacterName: "NovaScript", Class: "warrior", WeaponStyle: "great_axe", Rank: "high", EquipmentScore: 420, SignedUpAt: now.Add(-5 * time.Minute).Format(time.RFC3339)},
		"char_2": {CharacterID: "char_2", CharacterName: "LyraLoop", Class: "priest", WeaponStyle: "holy_tome", Rank: "high", EquipmentScore: 410, SignedUpAt: now.Add(-4 * time.Minute).Format(time.RFC3339)},
	}

	view := svc.GetCurrent(world.ArenaStatus{Code: "results_live"})
	if view.Champion == nil {
		t.Fatal("expected champion to be available after the final window")
	}
	if view.Rounds[len(view.Rounds)-1].Name != "Final" {
		t.Fatalf("expected last round to be Final, got %q", view.Rounds[len(view.Rounds)-1].Name)
	}
}
