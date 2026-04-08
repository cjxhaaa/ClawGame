package characters

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"clawgame/apps/api/internal/auth"
	"clawgame/apps/api/internal/world"
)

const businessTimezone = "Asia/Shanghai"

var (
	ErrCharacterAlreadyExists     = errors.New("character already exists")
	ErrCharacterNotFound          = errors.New("character not found")
	ErrCharacterInvalidClass      = errors.New("invalid class")
	ErrCharacterInvalidWeapon     = errors.New("invalid weapon style")
	ErrCharacterInvalidProfession = errors.New("invalid profession")
	ErrCharacterProfessionLocked  = errors.New("profession locked")
	ErrCharacterProfessionGold    = errors.New("profession insufficient gold")
	ErrCharacterInvalidRoute      = ErrCharacterInvalidProfession
	ErrCharacterRouteLocked       = ErrCharacterProfessionLocked
	ErrCharacterNameTaken         = errors.New("character name already exists")
	ErrCharacterInvalidName       = errors.New("invalid character name")
	ErrSkillNotFound              = errors.New("skill not found")
	ErrSkillLocked                = errors.New("skill locked")
	ErrSkillMaxLevel              = errors.New("skill max level reached")
	ErrSkillInvalidLoadout        = errors.New("skill loadout invalid")
	ErrGoldInsufficient           = errors.New("gold insufficient")
	ErrMaterialsInsufficient      = errors.New("materials insufficient")
	ErrTravelInsufficientGold     = errors.New("travel insufficient gold")
	ErrTravelRegionNotFound       = errors.New("travel region not found")
	ErrQuestCompletionCap         = errors.New("quest completion cap reached")
	ErrDungeonRewardClaimCap      = errors.New("dungeon reward claim cap reached")
	ErrReputationInsufficient     = errors.New("reputation insufficient")
	ErrActionNotSupported         = errors.New("action not supported")
)

const ProfessionChangeGoldCost = 800

const (
	FreeDungeonRewardClaimsPerDay = 2
	BonusDungeonClaimCostRep      = 50
	DailyQuestBoardSize           = 4
)

type Summary struct {
	CharacterID      string `json:"character_id"`
	Name             string `json:"name"`
	Class            string `json:"class"`
	ProfessionRoute  string `json:"profession_route_id,omitempty"`
	WeaponStyle      string `json:"weapon_style"`
	SeasonLevel      int    `json:"season_level"`
	SeasonXP         int    `json:"season_xp"`
	Reputation       int    `json:"reputation"`
	Gold             int    `json:"gold"`
	LocationRegionID string `json:"location_region_id"`
	Status           string `json:"status"`
}

type StatsSnapshot struct {
	MaxHP           int     `json:"max_hp"`
	PhysicalAttack  int     `json:"physical_attack"`
	MagicAttack     int     `json:"magic_attack"`
	PhysicalDefense int     `json:"physical_defense"`
	MagicDefense    int     `json:"magic_defense"`
	Speed           int     `json:"speed"`
	HealingPower    int     `json:"healing_power"`
	CritRate        float64 `json:"crit_rate"`
	CritDamage      float64 `json:"crit_damage"`
	BlockRate       float64 `json:"block_rate"`
	Precision       float64 `json:"precision"`
	EvasionRate     float64 `json:"evasion_rate"`
	PhysicalMastery float64 `json:"physical_mastery"`
	MagicMastery    float64 `json:"magic_mastery"`
}

type ArenaPowerPreview struct {
	ReferencePower        int    `json:"reference_power"`
	PowerDelta            int    `json:"power_delta"`
	EstimatedWinRateBand  string `json:"estimated_win_rate_band"`
	EstimatedStrengthTier string `json:"estimated_strength_tier"`
}

type DungeonPowerPreview struct {
	DungeonID           string `json:"dungeon_id"`
	DungeonName         string `json:"dungeon_name"`
	RecommendedPowerMin int    `json:"recommended_power_min"`
	RecommendedPowerMax int    `json:"recommended_power_max"`
	CurrentPower        int    `json:"current_power"`
	EstimatedConfidence string `json:"estimated_confidence"`
	EstimatedClearBand  string `json:"estimated_clear_band"`
}

type CombatPowerSummary struct {
	FormulaVersion     string                `json:"formula_version"`
	EffectiveLevel     int                   `json:"effective_level"`
	ProgressionCoeff   float64               `json:"progression_coeff"`
	BaseGrowthScore    int                   `json:"base_growth_score"`
	EquipmentScore     int                   `json:"equipment_score"`
	BuildModifierScore int                   `json:"build_modifier_score"`
	PanelPowerScore    int                   `json:"panel_power_score"`
	PowerTier          string                `json:"power_tier"`
	ArenaPreview       ArenaPowerPreview     `json:"arena_preview"`
	DungeonPreviews    []DungeonPowerPreview `json:"dungeon_previews"`
}

type DailyLimits struct {
	DailyResetAt               string `json:"daily_reset_at"`
	QuestCompletionCap         int    `json:"quest_completion_cap"`
	QuestCompletionUsed        int    `json:"quest_completion_used"`
	DungeonEntryCap            int    `json:"dungeon_entry_cap"`
	DungeonEntryUsed           int    `json:"dungeon_entry_used"`
	FreeDungeonEntryCap        int    `json:"free_dungeon_entry_cap"`
	BonusDungeonEntryPurchased int    `json:"bonus_dungeon_entry_purchased"`
	ReputationPerBonusClaim    int    `json:"reputation_per_bonus_claim"`
}

type QuestSummary struct {
	QuestID          string `json:"quest_id"`
	BoardID          string `json:"board_id"`
	TemplateType     string `json:"template_type"`
	ContractType     string `json:"contract_type,omitempty"`
	Difficulty       string `json:"difficulty,omitempty"`
	FlowKind         string `json:"flow_kind,omitempty"`
	Rarity           string `json:"rarity"`
	Status           string `json:"status"`
	Title            string `json:"title"`
	Description      string `json:"description"`
	TargetRegionID   string `json:"target_region_id,omitempty"`
	ProgressCurrent  int    `json:"progress_current"`
	ProgressTarget   int    `json:"progress_target"`
	RewardGold       int    `json:"reward_gold"`
	RewardReputation int    `json:"reward_reputation"`
}

type ValidAction struct {
	ActionType string         `json:"action_type"`
	Label      string         `json:"label"`
	ArgsSchema map[string]any `json:"args_schema"`
}

type SkillView struct {
	SkillID           string `json:"skill_id"`
	Name              string `json:"name"`
	DisplayNameZH     string `json:"display_name_zh"`
	Class             string `json:"class"`
	RouteID           string `json:"route_id"`
	Track             string `json:"skill_track"`
	Tier              string `json:"skill_tier"`
	CooldownRounds    int    `json:"cooldown_rounds"`
	IsBasic           bool   `json:"is_basic"`
	IsUnlocked        bool   `json:"is_unlocked"`
	Level             int    `json:"level"`
	MaxLevel          int    `json:"max_level"`
	CurrentMultiplier int    `json:"current_multiplier_pct"`
	NextLevelCost     int    `json:"next_level_cost_gold,omitempty"`
}

type SkillsStateView struct {
	BasicAttack    SkillView   `json:"basic_attack"`
	Universal      []SkillView `json:"universal_skills"`
	ClassSkills    []SkillView `json:"class_skills"`
	ActiveLoadout  []string    `json:"active_loadout"`
	MaxActiveSlots int         `json:"max_active_slots"`
}

type ProfessionChangeResultView struct {
	RequestedClass            string   `json:"requested_class"`
	FromClass                 string   `json:"from_class"`
	ToClass                   string   `json:"to_class"`
	GoldCost                  int      `json:"gold_cost"`
	GoldBefore                int      `json:"gold_before"`
	GoldAfter                 int      `json:"gold_after"`
	SkillLevelsPreserved      bool     `json:"skill_levels_preserved"`
	ActiveLoadoutBefore       []string `json:"active_loadout_before"`
	ActiveLoadoutAfter        []string `json:"active_loadout_after"`
	RemovedActiveSkillIDs     []string `json:"removed_active_skill_ids,omitempty"`
	WeaponAutoUnequipped      bool     `json:"weapon_auto_unequipped"`
	UnequippedWeaponItemID    string   `json:"unequipped_weapon_item_id,omitempty"`
	UnequippedWeaponCatalogID string   `json:"unequipped_weapon_catalog_id,omitempty"`
	UnequippedWeaponName      string   `json:"unequipped_weapon_name,omitempty"`
	StarterWeaponGranted      bool     `json:"starter_weapon_granted"`
	StarterWeaponItemID       string   `json:"starter_weapon_item_id,omitempty"`
	StarterWeaponCatalogID    string   `json:"starter_weapon_catalog_id,omitempty"`
	StarterWeaponEquipped     bool     `json:"starter_weapon_equipped"`
	Warnings                  []string `json:"warnings,omitempty"`
}

type StateView struct {
	ServerTime             string                      `json:"server_time"`
	Account                auth.Account                `json:"account"`
	Character              Summary                     `json:"character"`
	Stats                  StatsSnapshot               `json:"stats"`
	CombatPower            CombatPowerSummary          `json:"combat_power"`
	Skills                 SkillsStateView             `json:"skills"`
	Limits                 DailyLimits                 `json:"limits"`
	Materials              []MaterialBalance           `json:"materials"`
	SlotEnhancements       []SlotEnhancementView       `json:"slot_enhancements"`
	Objectives             []QuestSummary              `json:"objectives"`
	DungeonDaily           DungeonDailyHint            `json:"dungeon_daily"`
	RecentEvents           []world.WorldEvent          `json:"recent_events"`
	ValidActions           []ValidAction               `json:"valid_actions"`
	ProfessionChangeResult *ProfessionChangeResultView `json:"profession_change_result,omitempty"`
}

type MaterialBalance struct {
	MaterialKey string `json:"material_key"`
	Quantity    int    `json:"quantity"`
}

type SlotEnhancementView struct {
	Slot                  string  `json:"slot"`
	EnhancementLevel      int     `json:"enhancement_level"`
	EnhancementPreviewPct float64 `json:"enhancement_preview_pct"`
	MaxEnhancementLevel   int     `json:"max_enhancement_level"`
}

type DungeonDailyHint struct {
	HasRemainingQuota  bool     `json:"has_remaining_quota"`
	RemainingClaims    int      `json:"remaining_claims"`
	HasClaimableRun    bool     `json:"has_claimable_run"`
	PendingClaimRunIDs []string `json:"pending_claim_run_ids"`
}

type TravelResult struct {
	FromRegionID   string    `json:"from_region_id"`
	ToRegionID     string    `json:"to_region_id"`
	TravelCostGold int       `json:"travel_cost_gold"`
	Character      Summary   `json:"character"`
	State          StateView `json:"state"`
}

type StoredCharacter struct {
	AccountID             string
	Summary               Summary
	Stats                 StatsSnapshot
	SkillLevels           map[string]int
	SkillLoadout          []string
	QuestCompletionUsed   int
	DungeonEntryUsed      int
	DungeonBonusPurchased int
	DailyLimitsResetDate  string
}

type Repository interface {
	LoadCharacters() ([]StoredCharacter, error)
	LoadRecentEvents(limitPerCharacter int) ([]world.WorldEvent, error)
	SaveCharacter(StoredCharacter) error
	AppendWorldEvents(accountID string, characterID string, events []world.WorldEvent) error
}

type record struct {
	summary               Summary
	stats                 StatsSnapshot
	skillLevels           map[string]int
	skillLoadout          []string
	questCompletionUsed   int
	dungeonEntryUsed      int
	dungeonBonusPurchased int
	materials             map[string]int
	dailyLimitsResetDate  string
	recentEvents          []world.WorldEvent
}

type RuntimeSnapshot struct {
	Characters []Summary
	Events     []world.WorldEvent
}

type Service struct {
	mu                   sync.RWMutex
	clock                func() time.Time
	loc                  *time.Location
	repo                 Repository
	characterByAccountID map[string]record
	accountIDByName      map[string]string
}

var characterIDCounter uint64

func NewService() *Service {
	service, err := NewServiceWithRepository(nil)
	if err != nil {
		panic(err)
	}

	return service
}

func NewServiceWithRepository(repo Repository) (*Service, error) {
	service := &Service{
		clock:                time.Now,
		loc:                  mustLocation(businessTimezone),
		repo:                 repo,
		characterByAccountID: make(map[string]record),
		accountIDByName:      make(map[string]string),
	}

	if repo == nil {
		return service, nil
	}

	storedCharacters, err := repo.LoadCharacters()
	if err != nil {
		return nil, err
	}

	for _, stored := range storedCharacters {
		summary := normalizeSummary(stored.Summary)
		service.characterByAccountID[stored.AccountID] = record{
			summary:               summary,
			stats:                 baseStatsForSummary(summary),
			skillLevels:           cloneSkillLevels(stored.SkillLevels),
			skillLoadout:          cloneSkillLoadout(stored.SkillLoadout),
			questCompletionUsed:   stored.QuestCompletionUsed,
			dungeonEntryUsed:      stored.DungeonEntryUsed,
			dungeonBonusPurchased: stored.DungeonBonusPurchased,
			materials:             map[string]int{},
			dailyLimitsResetDate:  strings.TrimSpace(stored.DailyLimitsResetDate),
			recentEvents:          []world.WorldEvent{},
		}
		service.accountIDByName[strings.ToLower(summary.Name)] = stored.AccountID
	}

	events, err := repo.LoadRecentEvents(10)
	if err != nil {
		return nil, err
	}

	for _, event := range events {
		accountID, entry, ok := service.lookupByCharacterIDLocked(event.ActorCharacterID)
		if !ok {
			continue
		}

		if len(entry.recentEvents) < 10 {
			entry.recentEvents = append(entry.recentEvents, event)
		}
		service.characterByAccountID[accountID] = entry
	}

	return service, nil
}

func (s *Service) CreateCharacter(account auth.Account, name, class, weaponStyle string, worldService *world.Service) (StateView, error) {
	name = strings.TrimSpace(name)

	if len(name) < 3 || len(name) > 32 {
		return StateView{}, ErrCharacterInvalidName
	}

	s.mu.Lock()
	if _, exists := s.characterByAccountID[account.AccountID]; exists {
		s.mu.Unlock()
		return StateView{}, ErrCharacterAlreadyExists
	}
	if _, exists := s.accountIDByName[strings.ToLower(name)]; exists {
		s.mu.Unlock()
		return StateView{}, ErrCharacterNameTaken
	}

	summary := Summary{
		CharacterID:      nextID("char"),
		Name:             name,
		Class:            "civilian",
		ProfessionRoute:  "",
		WeaponStyle:      "",
		SeasonLevel:      1,
		SeasonXP:         0,
		Reputation:       0,
		Gold:             100,
		LocationRegionID: "main_city",
		Status:           "active",
	}

	entry := record{
		summary:              summary,
		stats:                baseStatsForSummary(summary),
		skillLevels:          map[string]int{},
		skillLoadout:         []string{},
		materials:            map[string]int{},
		dailyLimitsResetDate: s.businessDate(),
		recentEvents:         []world.WorldEvent{s.newEvent(summary, "character.created", "main_city", fmt.Sprintf("%s entered Main City as a civilian adventurer.", name), map[string]any{"class": summary.Class, "profession_route_id": "", "weapon_style": ""})},
	}

	if err := s.saveRecordLocked(account.AccountID, entry); err != nil {
		s.mu.Unlock()
		return StateView{}, err
	}
	if err := s.appendEventsLocked(account.AccountID, entry.summary.CharacterID, entry.recentEvents); err != nil {
		s.mu.Unlock()
		return StateView{}, err
	}

	s.characterByAccountID[account.AccountID] = entry
	s.accountIDByName[strings.ToLower(name)] = account.AccountID
	s.mu.Unlock()

	return s.GetState(account, worldService)
}

func (s *Service) GrantSeasonXP(characterID string, amount int) (Summary, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	accountID, entry, ok := s.lookupByCharacterIDLocked(characterID)
	if !ok {
		return Summary{}, ErrCharacterNotFound
	}
	entry, _, err := s.normalizeDailyLimitsLocked(accountID, entry)
	if err != nil {
		return Summary{}, err
	}
	grantSeasonXP(&entry.summary, amount)
	entry.stats = baseStatsForSummary(entry.summary)
	if err := s.saveRecordLocked(accountID, entry); err != nil {
		return Summary{}, err
	}
	s.characterByAccountID[accountID] = entry
	return entry.summary, nil
}

func (s *Service) ChooseProfessionRoute(account auth.Account, routeID string, worldService *world.Service) (StateView, error) {
	professionID, ok := NormalizeProfessionTarget(routeID)
	if !ok {
		return StateView{}, ErrCharacterInvalidProfession
	}

	s.mu.Lock()
	entry, ok := s.characterByAccountID[account.AccountID]
	if !ok {
		s.mu.Unlock()
		return StateView{}, ErrCharacterNotFound
	}
	entry, _, err := s.normalizeDailyLimitsLocked(account.AccountID, entry)
	if err != nil {
		s.mu.Unlock()
		return StateView{}, err
	}
	if entry.summary.SeasonLevel < 10 {
		s.mu.Unlock()
		return StateView{}, ErrCharacterProfessionLocked
	}
	currentClass := normalizeClassID(entry.summary.Class)
	if currentClass == professionID {
		s.mu.Unlock()
		return StateView{}, ErrCharacterProfessionLocked
	}
	if entry.summary.Gold < ProfessionChangeGoldCost {
		s.mu.Unlock()
		return StateView{}, ErrCharacterProfessionGold
	}

	entry.summary.Gold -= ProfessionChangeGoldCost
	entry.summary.Class = professionID
	if professionID == "civilian" {
		entry.summary.ProfessionRoute = ""
		entry.summary.WeaponStyle = ""
	} else {
		entry.summary.ProfessionRoute = professionID
		entry.summary.WeaponStyle = defaultStarterWeaponStyle(professionID)
	}
	entry.stats = baseStatsForSummary(entry.summary)
	entry.skillLoadout = activeLoadoutForSummary(entry.summary, entry.skillLevels, entry.skillLoadout)

	message := fmt.Sprintf("%s changed profession from %s to %s.", entry.summary.Name, currentClass, professionID)
	if currentClass == "civilian" && professionID != "civilian" {
		message = fmt.Sprintf("%s chose the %s profession.", entry.summary.Name, professionID)
	} else if professionID == "civilian" {
		message = fmt.Sprintf("%s returned to the civilian class.", entry.summary.Name)
	}
	event := s.newEvent(entry.summary, "character.profession_changed", entry.summary.LocationRegionID, message, map[string]any{
		"class":               professionID,
		"profession_route_id": entry.summary.ProfessionRoute,
		"weapon_style":        entry.summary.WeaponStyle,
		"from_class":          currentClass,
		"to_class":            professionID,
		"gold_cost":           ProfessionChangeGoldCost,
		"skills_preserved":    true,
	})
	entry.recentEvents = prependEvent(entry.recentEvents, event)

	if err := s.saveRecordLocked(account.AccountID, entry); err != nil {
		s.mu.Unlock()
		return StateView{}, err
	}
	if err := s.appendEventsLocked(account.AccountID, entry.summary.CharacterID, []world.WorldEvent{event}); err != nil {
		s.mu.Unlock()
		return StateView{}, err
	}
	s.characterByAccountID[account.AccountID] = entry
	s.mu.Unlock()

	return s.GetState(account, worldService)
}

func (s *Service) GetSummary(accountID string) (Summary, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.characterByAccountID[accountID]
	if !ok {
		return Summary{}, false
	}

	entry, _, _ = s.normalizeDailyLimitsLocked(accountID, entry)

	return entry.summary, true
}

func (s *Service) GetCharacterByID(characterID string) (Summary, StatsSnapshot, SkillsStateView, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, entry, ok := s.lookupByCharacterIDLocked(characterID)
	if !ok {
		return Summary{}, StatsSnapshot{}, SkillsStateView{}, false
	}
	return entry.summary, entry.stats, buildSkillsState(entry.summary, entry.skillLevels, entry.skillLoadout), true
}

func (s *Service) GetMe(account auth.Account) map[string]any {
	character, ok := s.GetSummary(account.AccountID)
	if !ok {
		return map[string]any{
			"account":   account,
			"character": nil,
		}
	}

	return map[string]any{
		"account":   account,
		"character": character,
	}
}

func (s *Service) GetState(account auth.Account, worldService *world.Service) (StateView, error) {
	s.mu.Lock()
	entry, ok := s.characterByAccountID[account.AccountID]
	if ok {
		var err error
		entry, _, err = s.normalizeDailyLimitsLocked(account.AccountID, entry)
		if err != nil {
			s.mu.Unlock()
			return StateView{}, err
		}
	}
	s.mu.Unlock()
	if !ok {
		return StateView{}, ErrCharacterNotFound
	}

	now := s.clock().In(s.loc)
	limits := BuildDailyLimits(nextDailyReset(now), entry.questCompletionUsed, entry.dungeonEntryUsed, entry.dungeonBonusPurchased)
	validActions := s.listValidActions(entry.summary.LocationRegionID, worldService)

	recentEvents := make([]world.WorldEvent, len(entry.recentEvents))
	copy(recentEvents, entry.recentEvents)

	return StateView{
		ServerTime:   now.Format(time.RFC3339),
		Account:      account,
		Character:    entry.summary,
		Stats:        entry.stats,
		Skills:       buildSkillsState(entry.summary, entry.skillLevels, entry.skillLoadout),
		Limits:       limits,
		Materials:    materialBalancesView(entry.materials),
		Objectives:   []QuestSummary{},
		DungeonDaily: DungeonDailyHint{HasRemainingQuota: limits.DungeonEntryUsed < limits.DungeonEntryCap, RemainingClaims: max(0, limits.DungeonEntryCap-limits.DungeonEntryUsed), PendingClaimRunIDs: []string{}},
		RecentEvents: recentEvents,
		ValidActions: validActions,
	}, nil
}

func (s *Service) ListValidActions(account auth.Account, worldService *world.Service) ([]ValidAction, error) {
	s.mu.RLock()
	entry, ok := s.characterByAccountID[account.AccountID]
	s.mu.RUnlock()
	if !ok {
		return nil, ErrCharacterNotFound
	}

	return s.listValidActions(entry.summary.LocationRegionID, worldService), nil
}

func (s *Service) Travel(account auth.Account, targetRegionID string, worldService *world.Service) (TravelResult, error) {
	targetRegionID = strings.TrimSpace(targetRegionID)

	region, ok := worldService.GetRegion(targetRegionID)
	if !ok {
		return TravelResult{}, ErrTravelRegionNotFound
	}

	s.mu.Lock()
	entry, ok := s.characterByAccountID[account.AccountID]
	if !ok {
		s.mu.Unlock()
		return TravelResult{}, ErrCharacterNotFound
	}
	entry, _, err := s.normalizeDailyLimitsLocked(account.AccountID, entry)
	if err != nil {
		s.mu.Unlock()
		return TravelResult{}, err
	}
	if entry.summary.Gold < region.Region.TravelCostGold {
		s.mu.Unlock()
		return TravelResult{}, ErrTravelInsufficientGold
	}

	fromRegionID := entry.summary.LocationRegionID
	entry.summary.Gold -= region.Region.TravelCostGold
	entry.summary.LocationRegionID = targetRegionID
	entry.recentEvents = prependEvent(entry.recentEvents, s.newEvent(entry.summary, "travel.completed", targetRegionID, fmt.Sprintf("%s travelled to %s.", entry.summary.Name, region.Region.Name), map[string]any{
		"from_region_id":   fromRegionID,
		"to_region_id":     targetRegionID,
		"travel_cost_gold": region.Region.TravelCostGold,
	}))
	if err := s.saveRecordLocked(account.AccountID, entry); err != nil {
		s.mu.Unlock()
		return TravelResult{}, err
	}
	if err := s.appendEventsLocked(account.AccountID, entry.summary.CharacterID, entry.recentEvents[:1]); err != nil {
		s.mu.Unlock()
		return TravelResult{}, err
	}
	s.characterByAccountID[account.AccountID] = entry
	s.mu.Unlock()

	state, err := s.GetState(account, worldService)
	if err != nil {
		return TravelResult{}, err
	}

	return TravelResult{
		FromRegionID:   fromRegionID,
		ToRegionID:     targetRegionID,
		TravelCostGold: region.Region.TravelCostGold,
		Character:      state.Character,
		State:          state,
	}, nil
}

func (s *Service) ExecuteAction(account auth.Account, actionType string, actionArgs map[string]any, worldService *world.Service) (map[string]any, error) {
	switch strings.TrimSpace(actionType) {
	case "travel":
		rawRegionID, _ := actionArgs["region_id"].(string)
		result, err := s.Travel(account, rawRegionID, worldService)
		if err != nil {
			return nil, err
		}

		return map[string]any{
			"action_result": map[string]any{
				"action_type":      "travel",
				"from_region_id":   result.FromRegionID,
				"to_region_id":     result.ToRegionID,
				"travel_cost_gold": result.TravelCostGold,
			},
			"state": result.State,
		}, nil
	default:
		return nil, ErrActionNotSupported
	}
}

func (s *Service) listValidActions(regionID string, worldService *world.Service) []ValidAction {
	region, ok := worldService.GetRegion(regionID)
	if !ok {
		return nil
	}

	actions := make([]ValidAction, 0, len(region.TravelOptions)+len(region.Buildings))
	if region.Region.Type == "dungeon" {
		actions = append(actions, ValidAction{
			ActionType: "enter_dungeon",
			Label:      fmt.Sprintf("Enter %s", region.Region.Name),
			ArgsSchema: map[string]any{
				"dungeon_id":     "string",
				"potion_loadout": "string[0..2]",
			},
		})
	}
	if region.Region.Type == "field" {
		actions = append(actions, ValidAction{
			ActionType: "resolve_field_encounter:hunt",
			Label:      fmt.Sprintf("Hunt in %s", region.Region.Name),
			ArgsSchema: map[string]any{},
		}, ValidAction{
			ActionType: "resolve_field_encounter:gather",
			Label:      fmt.Sprintf("Gather in %s", region.Region.Name),
			ArgsSchema: map[string]any{},
		}, ValidAction{
			ActionType: "resolve_field_encounter:curio",
			Label:      fmt.Sprintf("Investigate curios in %s", region.Region.Name),
			ArgsSchema: map[string]any{},
		})
	}
	for _, option := range region.TravelOptions {
		actions = append(actions, ValidAction{
			ActionType: "travel",
			Label:      fmt.Sprintf("Travel to %s", option.Name),
			ArgsSchema: map[string]any{
				"region_id":           "string",
				"suggested_region_id": option.RegionID,
			},
		})
	}
	for _, building := range region.Buildings {
		actions = append(actions, ValidAction{
			ActionType: "enter_building",
			Label:      fmt.Sprintf("Enter %s", building.Name),
			ArgsSchema: map[string]any{
				"building_id":           "string",
				"suggested_building_id": building.ID,
			},
		})
	}

	return actions
}

func (s *Service) newEvent(summary Summary, eventType, regionID, message string, payload map[string]any) world.WorldEvent {
	return world.WorldEvent{
		EventID:          nextID("evt"),
		EventType:        eventType,
		Visibility:       "public",
		ActorCharacterID: summary.CharacterID,
		ActorName:        summary.Name,
		RegionID:         regionID,
		Summary:          message,
		Payload:          payload,
		OccurredAt:       s.clock().In(s.loc).Format(time.RFC3339),
	}
}

func (s *Service) businessDate() string {
	now := s.clock().In(s.loc)
	if now.Hour() < 4 {
		now = now.Add(-24 * time.Hour)
	}

	return now.Format("2006-01-02")
}

func (s *Service) normalizeDailyLimitsLocked(accountID string, entry record) (record, bool, error) {
	currentDate := s.businessDate()
	if strings.TrimSpace(entry.dailyLimitsResetDate) == "" {
		entry.dailyLimitsResetDate = currentDate
	}
	if entry.dailyLimitsResetDate == currentDate {
		return entry, false, nil
	}

	entry.questCompletionUsed = 0
	entry.dungeonEntryUsed = 0
	entry.dungeonBonusPurchased = 0
	entry.dailyLimitsResetDate = currentDate

	if err := s.saveRecordLocked(accountID, entry); err != nil {
		return record{}, false, err
	}

	s.characterByAccountID[accountID] = entry
	return entry, true, nil
}

func (s *Service) saveRecordLocked(accountID string, entry record) error {
	if s.repo == nil {
		return nil
	}

	return s.repo.SaveCharacter(StoredCharacter{
		AccountID:             accountID,
		Summary:               entry.summary,
		Stats:                 entry.stats,
		SkillLevels:           cloneSkillLevels(entry.skillLevels),
		SkillLoadout:          cloneSkillLoadout(entry.skillLoadout),
		QuestCompletionUsed:   entry.questCompletionUsed,
		DungeonEntryUsed:      entry.dungeonEntryUsed,
		DungeonBonusPurchased: entry.dungeonBonusPurchased,
		DailyLimitsResetDate:  entry.dailyLimitsResetDate,
	})
}

func (s *Service) appendEventsLocked(accountID string, characterID string, events []world.WorldEvent) error {
	if s.repo == nil || len(events) == 0 {
		return nil
	}

	return s.repo.AppendWorldEvents(accountID, characterID, events)
}

func prependEvent(events []world.WorldEvent, event world.WorldEvent) []world.WorldEvent {
	items := make([]world.WorldEvent, 0, len(events)+1)
	items = append(items, event)
	items = append(items, events...)
	if len(items) > 10 {
		items = items[:10]
	}

	return items
}

func BuildDailyLimits(resetAt time.Time, questCompletionUsed, dungeonEntryUsed, dungeonBonusPurchased int) DailyLimits {
	return DailyLimits{
		DailyResetAt:               resetAt.Format(time.RFC3339),
		QuestCompletionCap:         DailyQuestBoardSize,
		QuestCompletionUsed:        questCompletionUsed,
		DungeonEntryCap:            FreeDungeonRewardClaimsPerDay + max(0, dungeonBonusPurchased),
		DungeonEntryUsed:           dungeonEntryUsed,
		FreeDungeonEntryCap:        FreeDungeonRewardClaimsPerDay,
		BonusDungeonEntryPurchased: max(0, dungeonBonusPurchased),
		ReputationPerBonusClaim:    BonusDungeonClaimCostRep,
	}
}

func (s *Service) SpendGold(characterID string, amount int) (Summary, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	accountID, entry, ok := s.lookupByCharacterIDLocked(characterID)
	if !ok {
		return Summary{}, ErrCharacterNotFound
	}
	entry, _, err := s.normalizeDailyLimitsLocked(accountID, entry)
	if err != nil {
		return Summary{}, err
	}
	if amount < 0 || entry.summary.Gold < amount {
		return Summary{}, ErrGoldInsufficient
	}

	entry.summary.Gold -= amount
	if err := s.saveRecordLocked(accountID, entry); err != nil {
		return Summary{}, err
	}
	s.characterByAccountID[accountID] = entry
	return entry.summary, nil
}

func (s *Service) GrantGold(characterID string, amount int) (Summary, error) {
	if amount < 0 {
		return Summary{}, ErrGoldInsufficient
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	accountID, entry, ok := s.lookupByCharacterIDLocked(characterID)
	if !ok {
		return Summary{}, ErrCharacterNotFound
	}
	entry, _, err := s.normalizeDailyLimitsLocked(accountID, entry)
	if err != nil {
		return Summary{}, err
	}

	entry.summary.Gold += amount
	if err := s.saveRecordLocked(accountID, entry); err != nil {
		return Summary{}, err
	}
	s.characterByAccountID[accountID] = entry
	return entry.summary, nil
}

func (s *Service) GrantMaterials(characterID string, materialDrops []map[string]any) ([]MaterialBalance, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	accountID, entry, ok := s.lookupByCharacterIDLocked(characterID)
	if !ok {
		return nil, ErrCharacterNotFound
	}
	entry, _, err := s.normalizeDailyLimitsLocked(accountID, entry)
	if err != nil {
		return nil, err
	}
	if entry.materials == nil {
		entry.materials = map[string]int{}
	}
	applyMaterialDrops(entry.materials, materialDrops)
	if err := s.saveRecordLocked(accountID, entry); err != nil {
		return nil, err
	}
	s.characterByAccountID[accountID] = entry
	return materialBalancesView(entry.materials), nil
}

func (s *Service) SpendMaterials(characterID string, materialCosts []map[string]any) ([]MaterialBalance, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	accountID, entry, ok := s.lookupByCharacterIDLocked(characterID)
	if !ok {
		return nil, ErrCharacterNotFound
	}
	entry, _, err := s.normalizeDailyLimitsLocked(accountID, entry)
	if err != nil {
		return nil, err
	}
	if entry.materials == nil {
		entry.materials = map[string]int{}
	}
	for _, drop := range materialCosts {
		key := stringPayload(drop, "material_key")
		qty := intPayload(drop, "quantity")
		if key == "" || qty <= 0 {
			continue
		}
		if entry.materials[key] < qty {
			return nil, ErrMaterialsInsufficient
		}
	}
	for _, drop := range materialCosts {
		key := stringPayload(drop, "material_key")
		qty := intPayload(drop, "quantity")
		if key == "" || qty <= 0 {
			continue
		}
		entry.materials[key] -= qty
	}
	if err := s.saveRecordLocked(accountID, entry); err != nil {
		return nil, err
	}
	s.characterByAccountID[accountID] = entry
	return materialBalancesView(entry.materials), nil
}

func (s *Service) ApplyFieldEncounter(characterID string, rewardGold int, materialDrops []map[string]any, grantXP bool, event world.WorldEvent) (Summary, []MaterialBalance, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	accountID, entry, ok := s.lookupByCharacterIDLocked(characterID)
	if !ok {
		return Summary{}, nil, ErrCharacterNotFound
	}

	entry, _, err := s.normalizeDailyLimitsLocked(accountID, entry)
	if err != nil {
		return Summary{}, nil, err
	}

	if rewardGold > 0 {
		entry.summary.Gold += rewardGold
	}
	if grantXP {
		grantSeasonXP(&entry.summary, 100)
	}
	if entry.materials == nil {
		entry.materials = map[string]int{}
	}
	applyMaterialDrops(entry.materials, materialDrops)
	entry.recentEvents = prependEvent(entry.recentEvents, event)

	if err := s.saveRecordLocked(accountID, entry); err != nil {
		return Summary{}, nil, err
	}
	if err := s.appendEventsLocked(accountID, entry.summary.CharacterID, []world.WorldEvent{event}); err != nil {
		return Summary{}, nil, err
	}

	s.characterByAccountID[accountID] = entry
	return entry.summary, materialBalancesView(entry.materials), nil
}

func (s *Service) ApplyDungeonRewardClaim(characterID, runID, dungeonID string, rewardGold int, rating string, rewardItemCatalogIDs []string, materialDrops []map[string]any) (Summary, DailyLimits, world.WorldEvent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	accountID, entry, ok := s.lookupByCharacterIDLocked(characterID)
	if !ok {
		return Summary{}, DailyLimits{}, world.WorldEvent{}, ErrCharacterNotFound
	}

	entry, _, err := s.normalizeDailyLimitsLocked(accountID, entry)
	if err != nil {
		return Summary{}, DailyLimits{}, world.WorldEvent{}, err
	}

	resetAt := nextDailyReset(s.clock().In(s.loc))

	if rewardGold > 0 {
		entry.summary.Gold += rewardGold
	}
	grantSeasonXP(&entry.summary, xpForDungeonRating(rating))
	if entry.materials == nil {
		entry.materials = map[string]int{}
	}
	applyMaterialDrops(entry.materials, materialDrops)
	entry.dungeonEntryUsed++

	event := world.WorldEvent{
		EventID:          nextID("evt"),
		EventType:        "dungeon.loot_granted",
		Visibility:       "public",
		ActorCharacterID: entry.summary.CharacterID,
		ActorName:        entry.summary.Name,
		RegionID:         entry.summary.LocationRegionID,
		Summary:          fmt.Sprintf("%s claimed dungeon rewards from %s.", entry.summary.Name, dungeonID),
		Payload: map[string]any{
			"run_id":                  runID,
			"dungeon_id":              dungeonID,
			"reward_gold":             rewardGold,
			"rating":                  rating,
			"reward_item_catalog_ids": rewardItemCatalogIDs,
			"material_drops":          materialDrops,
		},
		OccurredAt: s.clock().In(s.loc).Format(time.RFC3339),
	}

	entry.recentEvents = prependEvent(entry.recentEvents, event)

	if err := s.saveRecordLocked(accountID, entry); err != nil {
		return Summary{}, DailyLimits{}, world.WorldEvent{}, err
	}
	if err := s.appendEventsLocked(accountID, entry.summary.CharacterID, []world.WorldEvent{event}); err != nil {
		return Summary{}, DailyLimits{}, world.WorldEvent{}, err
	}

	s.characterByAccountID[accountID] = entry
	updatedLimits := BuildDailyLimits(resetAt, entry.questCompletionUsed, entry.dungeonEntryUsed, entry.dungeonBonusPurchased)
	return entry.summary, updatedLimits, event, nil
}

func (s *Service) ApplyQuestSubmission(characterID string, quest QuestSummary) (Summary, DailyLimits, []world.WorldEvent, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	accountID, entry, ok := s.lookupByCharacterIDLocked(characterID)
	if !ok {
		return Summary{}, DailyLimits{}, nil, false, ErrCharacterNotFound
	}
	entry, _, err := s.normalizeDailyLimitsLocked(accountID, entry)
	if err != nil {
		return Summary{}, DailyLimits{}, nil, false, err
	}

	resetAt := nextDailyReset(s.clock().In(s.loc))
	if entry.questCompletionUsed >= DailyQuestBoardSize {
		return Summary{}, DailyLimits{}, nil, false, ErrQuestCompletionCap
	}

	entry.summary.Gold += quest.RewardGold
	entry.summary.Reputation += quest.RewardReputation
	grantSeasonXP(&entry.summary, xpForQuest(quest))
	entry.questCompletionUsed++

	submitEvent := world.WorldEvent{
		EventID:          nextID("evt"),
		EventType:        "quest.submitted",
		Visibility:       "public",
		ActorCharacterID: entry.summary.CharacterID,
		ActorName:        entry.summary.Name,
		RegionID:         entry.summary.LocationRegionID,
		Summary:          fmt.Sprintf("%s submitted %s.", entry.summary.Name, quest.Title),
		Payload: map[string]any{
			"quest_id":          quest.QuestID,
			"quest_title":       quest.Title,
			"reward_gold":       quest.RewardGold,
			"reward_reputation": quest.RewardReputation,
		},
		OccurredAt: s.clock().In(s.loc).Format(time.RFC3339),
	}

	events := []world.WorldEvent{submitEvent}

	for _, event := range events {
		entry.recentEvents = prependEvent(entry.recentEvents, event)
	}

	if err := s.saveRecordLocked(accountID, entry); err != nil {
		return Summary{}, DailyLimits{}, nil, false, err
	}
	if err := s.appendEventsLocked(accountID, entry.summary.CharacterID, events); err != nil {
		return Summary{}, DailyLimits{}, nil, false, err
	}
	s.characterByAccountID[accountID] = entry
	updatedLimits := BuildDailyLimits(resetAt, entry.questCompletionUsed, entry.dungeonEntryUsed, entry.dungeonBonusPurchased)
	return entry.summary, updatedLimits, events, false, nil
}

func (s *Service) PurchaseDungeonRewardClaims(characterID string, quantity int) (Summary, DailyLimits, error) {
	if quantity <= 0 {
		quantity = 1
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	accountID, entry, ok := s.lookupByCharacterIDLocked(characterID)
	if !ok {
		return Summary{}, DailyLimits{}, ErrCharacterNotFound
	}
	entry, _, err := s.normalizeDailyLimitsLocked(accountID, entry)
	if err != nil {
		return Summary{}, DailyLimits{}, err
	}

	totalCost := quantity * BonusDungeonClaimCostRep
	if entry.summary.Reputation < totalCost {
		return Summary{}, DailyLimits{}, ErrReputationInsufficient
	}

	entry.summary.Reputation -= totalCost
	entry.dungeonBonusPurchased += quantity

	if err := s.saveRecordLocked(accountID, entry); err != nil {
		return Summary{}, DailyLimits{}, err
	}

	s.characterByAccountID[accountID] = entry
	resetAt := nextDailyReset(s.clock().In(s.loc))
	return entry.summary, BuildDailyLimits(resetAt, entry.questCompletionUsed, entry.dungeonEntryUsed, entry.dungeonBonusPurchased), nil
}

func (s *Service) AppendEvents(characterID string, events ...world.WorldEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	accountID, entry, ok := s.lookupByCharacterIDLocked(characterID)
	if !ok {
		return ErrCharacterNotFound
	}

	for _, event := range events {
		entry.recentEvents = prependEvent(entry.recentEvents, event)
	}

	if err := s.appendEventsLocked(accountID, characterID, events); err != nil {
		return err
	}
	s.characterByAccountID[accountID] = entry
	return nil
}

func (s *Service) GetCharacterByAccount(account auth.Account) (Summary, bool) {
	return s.GetSummary(account.AccountID)
}

func (s *Service) GetRuntimeDetailByCharacterID(characterID string) (Summary, StatsSnapshot, DailyLimits, []world.WorldEvent, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	accountID, entry, ok := s.lookupByCharacterIDLocked(characterID)
	if !ok {
		return Summary{}, StatsSnapshot{}, DailyLimits{}, nil, false
	}

	entry, _, err := s.normalizeDailyLimitsLocked(accountID, entry)
	if err != nil {
		return Summary{}, StatsSnapshot{}, DailyLimits{}, nil, false
	}

	resetAt := nextDailyReset(s.clock().In(s.loc))
	limits := BuildDailyLimits(resetAt, entry.questCompletionUsed, entry.dungeonEntryUsed, entry.dungeonBonusPurchased)
	events := make([]world.WorldEvent, len(entry.recentEvents))
	copy(events, entry.recentEvents)

	return entry.summary, entry.stats, limits, events, true
}

func (s *Service) SnapshotRuntime() RuntimeSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	characters := make([]Summary, 0, len(s.characterByAccountID))
	events := make([]world.WorldEvent, 0, len(s.characterByAccountID)*4)

	for _, entry := range s.characterByAccountID {
		characters = append(characters, entry.summary)
		events = append(events, entry.recentEvents...)
	}

	return RuntimeSnapshot{
		Characters: characters,
		Events:     events,
	}
}

func (s *Service) lookupByCharacterIDLocked(characterID string) (string, record, bool) {
	for accountID, entry := range s.characterByAccountID {
		if entry.summary.CharacterID == characterID {
			return accountID, entry, true
		}
	}

	return "", record{}, false
}

func applyMaterialDrops(balance map[string]int, drops []map[string]any) {
	if balance == nil || len(drops) == 0 {
		return
	}

	for _, drop := range drops {
		key, _ := drop["material_key"].(string)
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}

		quantity := intFromAny(drop["quantity"])
		if quantity <= 0 {
			continue
		}

		balance[key] += quantity
	}
}

func materialBalancesView(balance map[string]int) []MaterialBalance {
	if len(balance) == 0 {
		return []MaterialBalance{}
	}

	items := make([]MaterialBalance, 0, len(balance))
	for key, quantity := range balance {
		if quantity <= 0 {
			continue
		}
		items = append(items, MaterialBalance{MaterialKey: key, Quantity: quantity})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].MaterialKey < items[j].MaterialKey
	})

	return items
}

func intFromAny(value any) int {
	switch cast := value.(type) {
	case int:
		return cast
	case int32:
		return int(cast)
	case int64:
		return int(cast)
	case float64:
		return int(cast)
	default:
		return 0
	}
}

func stringPayload(payload map[string]any, key string) string {
	if payload == nil {
		return ""
	}
	value, _ := payload[key].(string)
	return strings.TrimSpace(value)
}

func intPayload(payload map[string]any, key string) int {
	if payload == nil {
		return 0
	}
	return intFromAny(payload[key])
}

func nextDailyReset(now time.Time) time.Time {
	resetToday := time.Date(now.Year(), now.Month(), now.Day(), 4, 0, 0, 0, now.Location())
	if now.Before(resetToday) {
		return resetToday
	}

	return resetToday.Add(24 * time.Hour)
}

func isWeaponCompatible(class, weaponStyle string) bool {
	allowedStyles, ok := allowedWeaponStyles[class]
	if !ok {
		return false
	}

	for _, item := range allowedStyles {
		if item == weaponStyle {
			return true
		}
	}

	return false
}

func mustLocation(name string) *time.Location {
	location, err := time.LoadLocation(name)
	if err != nil {
		panic(err)
	}

	return location
}

func nextID(prefix string) string {
	return fmt.Sprintf("%s_%d_%06d", prefix, time.Now().UnixNano(), atomic.AddUint64(&characterIDCounter, 1))
}

var allowedWeaponStyles = map[string][]string{
	"warrior": {"sword_shield", "great_axe"},
	"mage":    {"staff", "spellbook"},
	"priest":  {"scepter", "holy_tome"},
}

type growthProfile struct {
	Base              StatsSnapshot
	MaxHPGrowth       int
	PhysicalAtkGrowth int
	MagicAtkGrowth    int
	PhysicalDefGrowth int
	MagicDefGrowth    int
	SpeedGrowth       int
	HealingGrowth     int
}

var classGrowthProfiles = map[string]growthProfile{
	"civilian": {
		Base: StatsSnapshot{
			MaxHP:           96,
			PhysicalAttack:  14,
			MagicAttack:     14,
			PhysicalDefense: 10,
			MagicDefense:    10,
			Speed:           12,
			HealingPower:    6,
			CritRate:        0.20,
			CritDamage:      0.50,
			BlockRate:       0.05,
		},
		MaxHPGrowth:       600,
		PhysicalAtkGrowth: 100,
		MagicAtkGrowth:    100,
		PhysicalDefGrowth: 75,
		MagicDefGrowth:    75,
		SpeedGrowth:       11,
		HealingGrowth:     44,
	},
	"warrior": {
		Base: StatsSnapshot{
			MaxHP:           118,
			PhysicalAttack:  20,
			MagicAttack:     5,
			PhysicalDefense: 14,
			MagicDefense:    10,
			Speed:           10,
			HealingPower:    3,
			CritRate:        0.20, CritDamage: 0.50, BlockRate: 0.05,
		},
		MaxHPGrowth:       622,
		PhysicalAtkGrowth: 167,
		MagicAtkGrowth:    33,
		PhysicalDefGrowth: 89,
		MagicDefGrowth:    56,
		SpeedGrowth:       11,
		HealingGrowth:     11,
	},
	"mage": {
		Base: StatsSnapshot{
			MaxHP:           90,
			PhysicalAttack:  9,
			MagicAttack:     29,
			PhysicalDefense: 8,
			MagicDefense:    16,
			Speed:           15,
			HealingPower:    8,
			CritRate:        0.20, CritDamage: 0.50, BlockRate: 0.05,
		},
		MaxHPGrowth:       467,
		PhysicalAtkGrowth: 33,
		MagicAtkGrowth:    178,
		PhysicalDefGrowth: 44,
		MagicDefGrowth:    67,
		SpeedGrowth:       22,
		HealingGrowth:     22,
	},
	"priest": {
		Base: StatsSnapshot{
			MaxHP:           104,
			PhysicalAttack:  9,
			MagicAttack:     20,
			PhysicalDefense: 10,
			MagicDefense:    16,
			Speed:           13,
			HealingPower:    16,
			CritRate:        0.20, CritDamage: 0.50, BlockRate: 0.05,
		},
		MaxHPGrowth:       511,
		PhysicalAtkGrowth: 22,
		MagicAtkGrowth:    144,
		PhysicalDefGrowth: 44,
		MagicDefGrowth:    78,
		SpeedGrowth:       11,
		HealingGrowth:     89,
	},
}

var civilianBaseStats = classGrowthProfiles["civilian"].Base

func normalizeSummary(summary Summary) Summary {
	className := normalizeClassID(summary.Class)
	if className == "" {
		if mapped, ok := normalizeProfessionChoice(summary.ProfessionRoute); ok {
			className = mapped
		} else {
			className = "civilian"
		}
	}
	summary.Class = className
	if className == "civilian" {
		summary.ProfessionRoute = ""
		summary.WeaponStyle = ""
		return summary
	}
	summary.ProfessionRoute = className
	if strings.TrimSpace(summary.WeaponStyle) == "" {
		summary.WeaponStyle = defaultStarterWeaponStyle(className)
	}
	return summary
}

func normalizeClassID(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "civilian":
		return "civilian"
	case "warrior":
		return "warrior"
	case "mage":
		return "mage"
	case "priest":
		return "priest"
	default:
		return ""
	}
}

func normalizeProfessionChoice(value string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "warrior", "tank", "physical_burst", "magic_burst":
		return "warrior", true
	case "mage", "single_burst", "aoe_burst", "control":
		return "mage", true
	case "priest", "healing_support", "curse", "summon":
		return "priest", true
	default:
		return "", false
	}
}

func NormalizeProfessionTarget(value string) (string, bool) {
	if className := normalizeClassID(value); className != "" {
		return className, true
	}
	return normalizeProfessionChoice(value)
}

func defaultStarterWeaponStyle(className string) string {
	switch normalizeClassID(className) {
	case "warrior":
		return "sword_shield"
	case "mage":
		return "spellbook"
	case "priest":
		return "holy_tome"
	default:
		return ""
	}
}

func baseStatsForSummary(summary Summary) StatsSnapshot {
	return baseStatsForClassLevel(summary.Class, summary.SeasonLevel)
}

func baseStatsForClassLevel(className string, level int) StatsSnapshot {
	profile, ok := classGrowthProfiles[normalizeClassID(className)]
	if !ok {
		profile = classGrowthProfiles["civilian"]
	}
	if level < 1 {
		level = 1
	}
	if level > 100 {
		level = 100
	}
	steps := level - 1
	stats := profile.Base
	stats.MaxHP = growStat(stats.MaxHP, profile.MaxHPGrowth, steps)
	stats.PhysicalAttack = growStat(stats.PhysicalAttack, profile.PhysicalAtkGrowth, steps)
	stats.MagicAttack = growStat(stats.MagicAttack, profile.MagicAtkGrowth, steps)
	stats.PhysicalDefense = growStat(stats.PhysicalDefense, profile.PhysicalDefGrowth, steps)
	stats.MagicDefense = growStat(stats.MagicDefense, profile.MagicDefGrowth, steps)
	stats.Speed = growStat(stats.Speed, profile.SpeedGrowth, steps)
	stats.HealingPower = growStat(stats.HealingPower, profile.HealingGrowth, steps)
	return stats
}

func growStat(base, growthScaled, steps int) int {
	return base + (steps*growthScaled+50)/100
}

func grantSeasonXP(summary *Summary, amount int) {
	if amount <= 0 {
		return
	}
	summary.SeasonXP += amount
	if summary.SeasonXP < 0 {
		summary.SeasonXP = 0
	}
	summary.SeasonLevel = seasonLevelForXP(summary.SeasonXP)
}

func seasonLevelForXP(xp int) int {
	if xp <= 0 {
		return 1
	}
	level := 1
	remaining := xp
	for level < 100 {
		cost := seasonXPToNextLevel(level)
		if remaining < cost {
			return level
		}
		remaining -= cost
		level++
	}
	return 100
}

func seasonXPToNextLevel(level int) int {
	switch {
	case level < 1 || level >= 100:
		return 0
	case level < 10:
		return 420 + 20*(level-1)
	default:
		n := level - 10
		scaled := 88000 + 200*n + 5*n*n
		return (scaled + 50) / 100
	}
}

func xpForQuest(quest QuestSummary) int {
	switch strings.ToLower(strings.TrimSpace(quest.Difficulty)) {
	case "nightmare":
		return 520
	case "hard":
		return 320
	case "normal":
		return 220
	}
	switch strings.ToLower(strings.TrimSpace(quest.Rarity)) {
	case "challenge":
		return 520
	case "uncommon":
		return 320
	default:
		return 220
	}
}

func xpForDungeonRating(rating string) int {
	switch strings.ToUpper(strings.TrimSpace(rating)) {
	case "S":
		return 520
	case "A":
		return 460
	case "B":
		return 400
	case "C":
		return 340
	default:
		return 280
	}
}
