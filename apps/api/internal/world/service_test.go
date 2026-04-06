package world

import (
	"testing"
	"time"

	"clawgame/apps/api/internal/combat"
)

func TestCurrentArenaStatusDailySchedule(t *testing.T) {
	loc := time.FixedZone("CST", 8*60*60)

	testCases := []struct {
		name string
		now  time.Time
		want string
	}{
		{name: "weekday rating season", now: time.Date(2026, time.April, 1, 8, 59, 0, 0, loc), want: "rating_open"},
		{name: "saturday knockout pending", now: time.Date(2026, time.April, 4, 8, 59, 0, 0, loc), want: "knockout_pending"},
		{name: "saturday bracket in progress", now: time.Date(2026, time.April, 4, 9, 7, 0, 0, loc), want: "knockout_in_progress"},
		{name: "saturday results live after finals", now: time.Date(2026, time.April, 4, 9, 36, 0, 0, loc), want: "knockout_results_live"},
		{name: "sunday rest day", now: time.Date(2026, time.April, 5, 10, 0, 0, 0, loc), want: "rest_day"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := currentArenaStatus(tc.now)
			if got.Code != tc.want {
				t.Fatalf("expected status %q, got %q", tc.want, got.Code)
			}
		})
	}
}

func TestResolveFieldEncounterUsesSharedCombat(t *testing.T) {
	service := NewService()
	player := combat.Combatant{
		EntityID:        "char_test",
		Name:            "FieldTester",
		Team:            "a",
		IsPlayerSide:    true,
		Class:           "warrior",
		MaxHP:           860,
		CurrentHP:       860,
		PhysAtk:         220,
		MagAtk:          20,
		PhysDef:         90,
		MagDef:          70,
		Speed:           30,
		HealPow:         18,
		CritRate:        0.16,
		CritDamage:      0.50,
		PhysicalMastery: 0.24,
		PotionBag:       combat.DefaultPotionBag(),
	}

	result, err := service.ResolveFieldEncounter("whispering_forest", "hunt", player)
	if err != nil {
		t.Fatalf("expected field encounter to resolve, got %v", err)
	}
	if result.BattleType != "field_skirmish" {
		t.Fatalf("expected battle_type field_skirmish, got %q", result.BattleType)
	}
	if result.EncounterID == "" {
		t.Fatal("expected encounter_id to be populated")
	}
	if len(result.BattleLog) == 0 {
		t.Fatal("expected battle log to be populated")
	}
	if !result.Victory {
		t.Fatal("expected tuned player to win the field skirmish")
	}
	if result.RewardGold <= 0 {
		t.Fatal("expected successful field skirmish to grant gold")
	}
}
