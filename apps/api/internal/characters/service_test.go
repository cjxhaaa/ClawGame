package characters

import (
	"errors"
	"testing"
	"time"
)

func TestPurchaseDungeonRewardClaimsSpendsReputation(t *testing.T) {
	service := NewService()
	now := time.Date(2026, 4, 6, 10, 0, 0, 0, time.FixedZone("CST", 8*3600))
	service.clock = func() time.Time { return now }

	accountID := "acct_test"
	service.characterByAccountID[accountID] = record{
		summary: Summary{
			CharacterID:      "char_test",
			Name:             "RepBuyer",
			Class:            "civilian",
			SeasonLevel:      1,
			Reputation:       130,
			Gold:             100,
			LocationRegionID: "main_city",
			Status:           "active",
		},
		stats:                civilianBaseStats,
		skillLevels:          map[string]int{},
		skillLoadout:         []string{},
		materials:            map[string]int{},
		dailyLimitsResetDate: service.businessDate(),
	}

	summary, limits, err := service.PurchaseDungeonRewardClaims("char_test", 2)
	if err != nil {
		t.Fatalf("PurchaseDungeonRewardClaims failed: %v", err)
	}
	if summary.Reputation != 30 {
		t.Fatalf("expected remaining reputation 30, got %d", summary.Reputation)
	}
	if limits.DungeonEntryCap != FreeDungeonRewardClaimsPerDay+2 {
		t.Fatalf("expected dungeon cap %d, got %d", FreeDungeonRewardClaimsPerDay+2, limits.DungeonEntryCap)
	}
}

func TestApplyQuestSubmissionEnforcesDailyCompletionCap(t *testing.T) {
	service := NewService()
	now := time.Date(2026, 4, 6, 10, 0, 0, 0, time.FixedZone("CST", 8*3600))
	service.clock = func() time.Time { return now }

	accountID := "acct_cap"
	service.characterByAccountID[accountID] = record{
		summary: Summary{
			CharacterID:      "char_cap",
			Name:             "CapRunner",
			Class:            "civilian",
			SeasonLevel:      1,
			Reputation:       0,
			Gold:             100,
			LocationRegionID: "main_city",
			Status:           "active",
		},
		stats:                civilianBaseStats,
		skillLevels:          map[string]int{},
		skillLoadout:         []string{},
		materials:            map[string]int{},
		questCompletionUsed:  DailyQuestBoardSize,
		dailyLimitsResetDate: service.businessDate(),
	}

	_, _, _, _, err := service.ApplyQuestSubmission("char_cap", QuestSummary{
		QuestID:          "quest_cap",
		Title:            "Cap Test Contract",
		RewardGold:       100,
		RewardReputation: 20,
	})
	if !errors.Is(err, ErrQuestCompletionCap) {
		t.Fatalf("expected ErrQuestCompletionCap, got %v", err)
	}
}
