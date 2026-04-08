package characters

import (
	"errors"
	"testing"
	"time"

	"clawgame/apps/api/internal/auth"
	"clawgame/apps/api/internal/world"
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

func TestSeasonXPToNextLevelCurve(t *testing.T) {
	tests := []struct {
		level int
		want  int
	}{
		{level: 1, want: 420},
		{level: 5, want: 500},
		{level: 9, want: 580},
		{level: 10, want: 880},
		{level: 20, want: 905},
		{level: 40, want: 985},
		{level: 60, want: 1105},
		{level: 80, want: 1265},
		{level: 99, want: 1454},
	}
	for _, tt := range tests {
		if got := seasonXPToNextLevel(tt.level); got != tt.want {
			t.Fatalf("seasonXPToNextLevel(%d) = %d, want %d", tt.level, got, tt.want)
		}
	}
}

func TestSeasonLevelForXPUsesUpdatedCurve(t *testing.T) {
	tests := []struct {
		xp   int
		want int
	}{
		{xp: 0, want: 1},
		{xp: 4499, want: 9},
		{xp: 4500, want: 10},
		{xp: 13402, want: 19},
		{xp: 13403, want: 20},
		{xp: 103646, want: 99},
		{xp: 103647, want: 100},
		{xp: 200000, want: 100},
	}
	for _, tt := range tests {
		if got := seasonLevelForXP(tt.xp); got != tt.want {
			t.Fatalf("seasonLevelForXP(%d) = %d, want %d", tt.xp, got, tt.want)
		}
	}
}

func TestGrantSeasonXPRecomputesCivilianGrowth(t *testing.T) {
	service := NewService()
	accountID := "acct_growth"
	service.characterByAccountID[accountID] = record{
		summary: Summary{
			CharacterID:      "char_growth",
			Name:             "Growth",
			Class:            "civilian",
			SeasonLevel:      1,
			SeasonXP:         0,
			LocationRegionID: "main_city",
			Status:           "active",
		},
		stats:                civilianBaseStats,
		skillLevels:          map[string]int{},
		skillLoadout:         []string{},
		materials:            map[string]int{},
		dailyLimitsResetDate: service.businessDate(),
	}

	summary, err := service.GrantSeasonXP("char_growth", 4500)
	if err != nil {
		t.Fatalf("GrantSeasonXP failed: %v", err)
	}
	if summary.SeasonLevel != 10 {
		t.Fatalf("expected level 10 after granting xp, got %d", summary.SeasonLevel)
	}
	_, stats, _, ok := service.GetCharacterByID("char_growth")
	if !ok {
		t.Fatal("expected character after granting xp")
	}
	if stats.MaxHP != 150 || stats.PhysicalAttack != 23 || stats.MagicAttack != 23 {
		t.Fatalf("expected civilian level-10 balanced growth, got %#v", stats)
	}
}

func TestChooseProfessionUsesClassGrowthAtCurrentLevel(t *testing.T) {
	service := NewService()
	worldService := world.NewService()
	account := auth.Account{AccountID: "acct_prof"}
	service.characterByAccountID[account.AccountID] = record{
		summary: Summary{
			CharacterID:      "char_prof",
			Name:             "Promote",
			Class:            "civilian",
			SeasonLevel:      10,
			SeasonXP:         4500,
			Gold:             ProfessionChangeGoldCost + 100,
			LocationRegionID: "main_city",
			Status:           "active",
		},
		stats:                baseStatsForClassLevel("civilian", 10),
		skillLevels:          map[string]int{"Quickstep": 2},
		skillLoadout:         []string{"Quickstep"},
		materials:            map[string]int{},
		dailyLimitsResetDate: service.businessDate(),
	}

	state, err := service.ChooseProfessionRoute(account, "mage", worldService)
	if err != nil {
		t.Fatalf("ChooseProfessionRoute failed: %v", err)
	}
	if state.Character.Class != "mage" {
		t.Fatalf("expected class mage, got %q", state.Character.Class)
	}
	if state.Character.ProfessionRoute != "mage" {
		t.Fatalf("expected profession id mage, got %q", state.Character.ProfessionRoute)
	}
	if state.Character.WeaponStyle != "spellbook" {
		t.Fatalf("expected starter weapon spellbook, got %q", state.Character.WeaponStyle)
	}
	if state.Character.Gold != 100 {
		t.Fatalf("expected remaining gold 100 after profession change cost, got %d", state.Character.Gold)
	}
	if state.Stats.MaxHP != 132 || state.Stats.MagicAttack != 45 || state.Stats.Speed != 17 {
		t.Fatalf("expected mage level-10 growth profile, got %#v", state.Stats)
	}
	if state.Skills.BasicAttack.SkillID == "" {
		t.Fatal("expected skills state to remain populated")
	}
	if len(state.Skills.CivilianSkills) == 0 {
		t.Fatal("expected civilian skills to remain available after profession change")
	}
	if len(state.Skills.ClassCommonSkills) == 0 {
		t.Fatal("expected mage class-common skills after profession change")
	}
	if len(state.Skills.ActiveLoadout) != 0 {
		t.Fatalf("expected profession change to clear active loadout, got %#v", state.Skills.ActiveLoadout)
	}
}

func TestChooseProfessionAcceptsLegacyRouteAlias(t *testing.T) {
	service := NewService()
	worldService := world.NewService()
	account := auth.Account{AccountID: "acct_legacy"}
	service.characterByAccountID[account.AccountID] = record{
		summary: Summary{
			CharacterID:      "char_legacy",
			Name:             "Legacy",
			Class:            "civilian",
			SeasonLevel:      10,
			SeasonXP:         4500,
			Gold:             ProfessionChangeGoldCost,
			LocationRegionID: "main_city",
			Status:           "active",
		},
		stats:                baseStatsForClassLevel("civilian", 10),
		skillLevels:          map[string]int{},
		skillLoadout:         []string{},
		materials:            map[string]int{},
		dailyLimitsResetDate: service.businessDate(),
	}

	state, err := service.ChooseProfessionRoute(account, "control", worldService)
	if err != nil {
		t.Fatalf("ChooseProfessionRoute with legacy alias failed: %v", err)
	}
	if state.Character.Class != "mage" || state.Character.ProfessionRoute != "mage" {
		t.Fatalf("expected legacy route alias to normalize to mage, got class=%q profession=%q", state.Character.Class, state.Character.ProfessionRoute)
	}
}

func TestChooseProfessionAllowsClassSwapAndKeepsLearnedSkills(t *testing.T) {
	service := NewService()
	worldService := world.NewService()
	account := auth.Account{AccountID: "acct_respec"}
	service.characterByAccountID[account.AccountID] = record{
		summary: Summary{
			CharacterID:      "char_respec",
			Name:             "Respec",
			Class:            "warrior",
			ProfessionRoute:  "warrior",
			WeaponStyle:      "great_axe",
			SeasonLevel:      14,
			SeasonXP:         9000,
			Gold:             ProfessionChangeGoldCost + 75,
			LocationRegionID: "main_city",
			Status:           "active",
		},
		stats: baseStatsForClassLevel("warrior", 14),
		skillLevels: map[string]int{
			"Quickstep":    2,
			"Guard Stance": 3,
			"Arc Veil":     1,
		},
		skillLoadout:         []string{"Guard Stance", "Quickstep"},
		materials:            map[string]int{},
		dailyLimitsResetDate: service.businessDate(),
	}

	state, err := service.ChooseProfessionRoute(account, "mage", worldService)
	if err != nil {
		t.Fatalf("ChooseProfessionRoute class swap failed: %v", err)
	}
	if state.Character.Class != "mage" || state.Character.WeaponStyle != "spellbook" {
		t.Fatalf("expected switch to mage spellbook profile, got class=%q weapon=%q", state.Character.Class, state.Character.WeaponStyle)
	}
	if state.Character.Gold != 75 {
		t.Fatalf("expected remaining gold 75 after class swap, got %d", state.Character.Gold)
	}
	if state.Skills.ClassSkills == nil {
		t.Fatal("expected class skills view after class swap")
	}
	if len(state.Skills.CivilianSkills) == 0 || len(state.Skills.ClassCommonSkills) == 0 {
		t.Fatal("expected class swap to expose civilian and current-class common skills")
	}
	if len(state.Skills.ActiveLoadout) != 0 {
		t.Fatalf("expected class swap to clear active loadout, got %#v", state.Skills.ActiveLoadout)
	}
	accountID, entry, ok := service.lookupByCharacterIDLocked("char_respec")
	if !ok || accountID != account.AccountID {
		t.Fatal("expected stored character after class swap")
	}
	if entry.skillLevels["Guard Stance"] != 3 || entry.skillLevels["Arc Veil"] != 1 {
		t.Fatalf("expected learned skill levels to be preserved, got %#v", entry.skillLevels)
	}
}

func TestChooseProfessionAllowsReturningToCivilian(t *testing.T) {
	service := NewService()
	worldService := world.NewService()
	account := auth.Account{AccountID: "acct_return"}
	service.characterByAccountID[account.AccountID] = record{
		summary: Summary{
			CharacterID:      "char_return",
			Name:             "Return",
			Class:            "priest",
			ProfessionRoute:  "priest",
			WeaponStyle:      "holy_tome",
			SeasonLevel:      12,
			SeasonXP:         6260,
			Gold:             ProfessionChangeGoldCost + 20,
			LocationRegionID: "main_city",
			Status:           "active",
		},
		stats:                baseStatsForClassLevel("priest", 12),
		skillLevels:          map[string]int{"Quickstep": 1, "Restore": 2},
		skillLoadout:         []string{"Restore", "Quickstep"},
		materials:            map[string]int{},
		dailyLimitsResetDate: service.businessDate(),
	}

	state, err := service.ChooseProfessionRoute(account, "civilian", worldService)
	if err != nil {
		t.Fatalf("ChooseProfessionRoute civilian return failed: %v", err)
	}
	if state.Character.Class != "civilian" || state.Character.ProfessionRoute != "" || state.Character.WeaponStyle != "" {
		t.Fatalf("expected return to civilian, got %#v", state.Character)
	}
	if state.Character.Gold != 20 {
		t.Fatalf("expected remaining gold 20 after return, got %d", state.Character.Gold)
	}
	if len(state.Skills.ClassCommonSkills) == 0 {
		t.Fatal("expected civilian class to access profession-common skills")
	}
	if len(state.Skills.ActiveLoadout) != 0 {
		t.Fatalf("expected return to civilian to clear active loadout, got %#v", state.Skills.ActiveLoadout)
	}
}

func TestClassSkillMetadataUsesRouteLabelsWithoutSharedLayer(t *testing.T) {
	tests := []struct {
		skillID  string
		routeID  string
		track    string
		tier     string
		cooldown int
	}{
		{skillID: "Guard Stance", routeID: "tank", track: "tank", tier: "advanced", cooldown: 2},
		{skillID: "War Cry", routeID: "physical_burst", track: "physical_burst", tier: "ultimate", cooldown: 3},
		{skillID: "Intercept", routeID: "tank", track: "tank", tier: "normal", cooldown: 1},
		{skillID: "Arc Veil", routeID: "control", track: "control", tier: "advanced", cooldown: 2},
		{skillID: "Focus Pulse", routeID: "single_burst", track: "single_burst", tier: "normal", cooldown: 1},
		{skillID: "Restore", routeID: "healing_support", track: "healing_support", tier: "normal", cooldown: 1},
		{skillID: "Sanctuary Mark", routeID: "healing_support", track: "healing_support", tier: "advanced", cooldown: 2},
	}

	for _, tt := range tests {
		definition, ok := skillDefinitions[tt.skillID]
		if !ok {
			t.Fatalf("expected skill definition for %q", tt.skillID)
		}
		if definition.RouteID != tt.routeID || definition.Track != tt.track {
			t.Fatalf("expected %s route/track %s/%s, got %s/%s", tt.skillID, tt.routeID, tt.track, definition.RouteID, definition.Track)
		}
		if definition.RouteID == "shared" || definition.Track == "shared" {
			t.Fatalf("expected %s to avoid shared skill-layer tags, got %#v", tt.skillID, definition)
		}
		if definition.Tier != tt.tier || definition.CooldownRounds != tt.cooldown {
			t.Fatalf("expected %s tier/cooldown %s/%d, got %s/%d", tt.skillID, tt.tier, tt.cooldown, definition.Tier, definition.CooldownRounds)
		}
	}
}

func TestChooseProfessionRequiresGoldCost(t *testing.T) {
	service := NewService()
	worldService := world.NewService()
	account := auth.Account{AccountID: "acct_cost"}
	service.characterByAccountID[account.AccountID] = record{
		summary: Summary{
			CharacterID:      "char_cost",
			Name:             "Cost",
			Class:            "civilian",
			SeasonLevel:      10,
			SeasonXP:         4500,
			Gold:             ProfessionChangeGoldCost - 1,
			LocationRegionID: "main_city",
			Status:           "active",
		},
		stats:                baseStatsForClassLevel("civilian", 10),
		skillLevels:          map[string]int{},
		skillLoadout:         []string{},
		materials:            map[string]int{},
		dailyLimitsResetDate: service.businessDate(),
	}

	if _, err := service.ChooseProfessionRoute(account, "warrior", worldService); !errors.Is(err, ErrCharacterProfessionGold) {
		t.Fatalf("expected ErrCharacterProfessionGold, got %v", err)
	}
}
