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
	ErrCharacterAlreadyExists = errors.New("character already exists")
	ErrCharacterNotFound      = errors.New("character not found")
	ErrCharacterInvalidClass  = errors.New("invalid class")
	ErrCharacterInvalidWeapon = errors.New("invalid weapon style")
	ErrCharacterNameTaken     = errors.New("character name already exists")
	ErrCharacterInvalidName   = errors.New("invalid character name")
	ErrGoldInsufficient       = errors.New("gold insufficient")
	ErrMaterialsInsufficient  = errors.New("materials insufficient")
	ErrTravelRankLocked       = errors.New("travel rank locked")
	ErrTravelInsufficientGold = errors.New("travel insufficient gold")
	ErrTravelRegionNotFound   = errors.New("travel region not found")
	ErrQuestCompletionCap     = errors.New("quest completion cap reached")
	ErrDungeonRewardClaimCap  = errors.New("dungeon reward claim cap reached")
	ErrActionNotSupported     = errors.New("action not supported")
)

type Summary struct {
	CharacterID      string `json:"character_id"`
	Name             string `json:"name"`
	Class            string `json:"class"`
	WeaponStyle      string `json:"weapon_style"`
	Rank             string `json:"rank"`
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

type DailyLimits struct {
	DailyResetAt        string `json:"daily_reset_at"`
	QuestCompletionCap  int    `json:"quest_completion_cap"`
	QuestCompletionUsed int    `json:"quest_completion_used"`
	DungeonEntryCap     int    `json:"dungeon_entry_cap"`
	DungeonEntryUsed    int    `json:"dungeon_entry_used"`
}

type QuestSummary struct {
	QuestID          string `json:"quest_id"`
	BoardID          string `json:"board_id"`
	TemplateType     string `json:"template_type"`
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

type StateView struct {
	ServerTime       string                `json:"server_time"`
	Account          auth.Account          `json:"account"`
	Character        Summary               `json:"character"`
	Stats            StatsSnapshot         `json:"stats"`
	Limits           DailyLimits           `json:"limits"`
	Materials        []MaterialBalance     `json:"materials"`
	SlotEnhancements []SlotEnhancementView `json:"slot_enhancements"`
	Objectives       []QuestSummary        `json:"objectives"`
	DungeonDaily     DungeonDailyHint      `json:"dungeon_daily"`
	RecentEvents     []world.WorldEvent    `json:"recent_events"`
	ValidActions     []ValidAction         `json:"valid_actions"`
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
	AccountID            string
	Summary              Summary
	Stats                StatsSnapshot
	QuestCompletionUsed  int
	DungeonEntryUsed     int
	DailyLimitsResetDate string
}

type Repository interface {
	LoadCharacters() ([]StoredCharacter, error)
	LoadRecentEvents(limitPerCharacter int) ([]world.WorldEvent, error)
	SaveCharacter(StoredCharacter) error
	AppendWorldEvents(accountID string, characterID string, events []world.WorldEvent) error
}

type record struct {
	summary              Summary
	stats                StatsSnapshot
	questCompletionUsed  int
	dungeonEntryUsed     int
	materials            map[string]int
	dailyLimitsResetDate string
	recentEvents         []world.WorldEvent
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
		service.characterByAccountID[stored.AccountID] = record{
			summary:              stored.Summary,
			stats:                stored.Stats,
			questCompletionUsed:  stored.QuestCompletionUsed,
			dungeonEntryUsed:     stored.DungeonEntryUsed,
			materials:            map[string]int{},
			dailyLimitsResetDate: strings.TrimSpace(stored.DailyLimitsResetDate),
			recentEvents:         []world.WorldEvent{},
		}
		service.accountIDByName[strings.ToLower(stored.Summary.Name)] = stored.AccountID
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
	class = strings.TrimSpace(strings.ToLower(class))
	weaponStyle = strings.TrimSpace(strings.ToLower(weaponStyle))

	if len(name) < 3 || len(name) > 32 {
		return StateView{}, ErrCharacterInvalidName
	}
	if _, ok := allowedWeaponStyles[class]; !ok {
		return StateView{}, ErrCharacterInvalidClass
	}
	if !isWeaponCompatible(class, weaponStyle) {
		return StateView{}, ErrCharacterInvalidWeapon
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
		Class:            class,
		WeaponStyle:      weaponStyle,
		Rank:             "low",
		Reputation:       0,
		Gold:             100,
		LocationRegionID: "main_city",
		Status:           "active",
	}

	entry := record{
		summary:              summary,
		stats:                baseStatsByStyle[weaponStyle],
		materials:            map[string]int{},
		dailyLimitsResetDate: s.businessDate(),
		recentEvents:         []world.WorldEvent{s.newEvent(summary, "character.created", "main_city", fmt.Sprintf("%s completed character creation and entered Main City.", name), map[string]any{"class": class, "weapon_style": weaponStyle})},
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
	limits := LimitsForRank(entry.summary.Rank, nextDailyReset(now), entry.questCompletionUsed, entry.dungeonEntryUsed)
	validActions := s.listValidActions(entry.summary.LocationRegionID, worldService)

	recentEvents := make([]world.WorldEvent, len(entry.recentEvents))
	copy(recentEvents, entry.recentEvents)

	return StateView{
		ServerTime:   now.Format(time.RFC3339),
		Account:      account,
		Character:    entry.summary,
		Stats:        entry.stats,
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
	if !rankAllows(entry.summary.Rank, region.Region.MinRank) {
		s.mu.Unlock()
		return TravelResult{}, ErrTravelRankLocked
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
			ArgsSchema: map[string]any{"region_id": "string"},
		})
	}
	for _, building := range region.Buildings {
		actions = append(actions, ValidAction{
			ActionType: "enter_building",
			Label:      fmt.Sprintf("Enter %s", building.Name),
			ArgsSchema: map[string]any{"building_id": "string"},
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
		AccountID:            accountID,
		Summary:              entry.summary,
		Stats:                entry.stats,
		QuestCompletionUsed:  entry.questCompletionUsed,
		DungeonEntryUsed:     entry.dungeonEntryUsed,
		DailyLimitsResetDate: entry.dailyLimitsResetDate,
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

func LimitsForRank(rank string, resetAt time.Time, questCompletionUsed, dungeonEntryUsed int) DailyLimits {
	switch rank {
	case "mid":
		return DailyLimits{
			DailyResetAt:        resetAt.Format(time.RFC3339),
			QuestCompletionCap:  6,
			QuestCompletionUsed: questCompletionUsed,
			DungeonEntryCap:     4,
			DungeonEntryUsed:    dungeonEntryUsed,
		}
	case "high":
		return DailyLimits{
			DailyResetAt:        resetAt.Format(time.RFC3339),
			QuestCompletionCap:  8,
			QuestCompletionUsed: questCompletionUsed,
			DungeonEntryCap:     6,
			DungeonEntryUsed:    dungeonEntryUsed,
		}
	default:
		return DailyLimits{
			DailyResetAt:        resetAt.Format(time.RFC3339),
			QuestCompletionCap:  4,
			QuestCompletionUsed: questCompletionUsed,
			DungeonEntryCap:     2,
			DungeonEntryUsed:    dungeonEntryUsed,
		}
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

func (s *Service) ApplyFieldEncounter(characterID string, rewardGold int, materialDrops []map[string]any, event world.WorldEvent) (Summary, []MaterialBalance, error) {
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
	updatedLimits := LimitsForRank(entry.summary.Rank, resetAt, entry.questCompletionUsed, entry.dungeonEntryUsed)
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
	limits := LimitsForRank(entry.summary.Rank, resetAt, entry.questCompletionUsed, entry.dungeonEntryUsed)
	if entry.questCompletionUsed >= limits.QuestCompletionCap {
		return Summary{}, DailyLimits{}, nil, false, ErrQuestCompletionCap
	}

	previousRank := entry.summary.Rank
	entry.summary.Gold += quest.RewardGold
	entry.summary.Reputation += quest.RewardReputation
	entry.questCompletionUsed++
	entry.summary.Rank = rankForReputation(entry.summary.Reputation)

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
	if entry.summary.Rank != previousRank {
		events = append([]world.WorldEvent{{
			EventID:          nextID("evt"),
			EventType:        "character.rank_up",
			Visibility:       "public",
			ActorCharacterID: entry.summary.CharacterID,
			ActorName:        entry.summary.Name,
			RegionID:         entry.summary.LocationRegionID,
			Summary:          fmt.Sprintf("%s advanced from %s-rank to %s-rank.", entry.summary.Name, previousRank, entry.summary.Rank),
			Payload: map[string]any{
				"previous_rank": previousRank,
				"current_rank":  entry.summary.Rank,
			},
			OccurredAt: s.clock().In(s.loc).Format(time.RFC3339),
		}}, events...)
	}

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
	updatedLimits := LimitsForRank(entry.summary.Rank, resetAt, entry.questCompletionUsed, entry.dungeonEntryUsed)
	return entry.summary, updatedLimits, events, entry.summary.Rank != previousRank, nil
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
	limits := LimitsForRank(entry.summary.Rank, resetAt, entry.questCompletionUsed, entry.dungeonEntryUsed)
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

func rankForReputation(reputation int) string {
	switch {
	case reputation >= 600:
		return "high"
	case reputation >= 200:
		return "mid"
	default:
		return "low"
	}
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

func rankAllows(currentRank, requiredRank string) bool {
	return rankOrder[currentRank] >= rankOrder[requiredRank]
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

var rankOrder = map[string]int{
	"low":  1,
	"mid":  2,
	"high": 3,
}

var allowedWeaponStyles = map[string][]string{
	"warrior": {"sword_shield", "great_axe"},
	"mage":    {"staff", "spellbook"},
	"priest":  {"scepter", "holy_tome"},
}

var baseStatsByStyle = map[string]StatsSnapshot{
	"sword_shield": {
		MaxHP:           132,
		PhysicalAttack:  24,
		MagicAttack:     6,
		PhysicalDefense: 18,
		MagicDefense:    10,
		Speed:           10,
		HealingPower:    4,
		CritRate:        0.20,
		CritDamage:      0.50,
		BlockRate:       0.05,
	},
	"great_axe": {
		MaxHP:           120,
		PhysicalAttack:  30,
		MagicAttack:     4,
		PhysicalDefense: 14,
		MagicDefense:    8,
		Speed:           11,
		HealingPower:    3,
		CritRate:        0.20,
		CritDamage:      0.50,
		BlockRate:       0.05,
	},
	"staff": {
		MaxHP:           92,
		PhysicalAttack:  12,
		MagicAttack:     34,
		PhysicalDefense: 9,
		MagicDefense:    18,
		Speed:           16,
		HealingPower:    8,
		CritRate:        0.20,
		CritDamage:      0.50,
		BlockRate:       0.05,
	},
	"spellbook": {
		MaxHP:           88,
		PhysicalAttack:  10,
		MagicAttack:     36,
		PhysicalDefense: 8,
		MagicDefense:    16,
		Speed:           15,
		HealingPower:    10,
		CritRate:        0.20,
		CritDamage:      0.50,
		BlockRate:       0.05,
	},
	"scepter": {
		MaxHP:           104,
		PhysicalAttack:  10,
		MagicAttack:     26,
		PhysicalDefense: 11,
		MagicDefense:    17,
		Speed:           14,
		HealingPower:    20,
		CritRate:        0.20,
		CritDamage:      0.50,
		BlockRate:       0.05,
	},
	"holy_tome": {
		MaxHP:           98,
		PhysicalAttack:  8,
		MagicAttack:     22,
		PhysicalDefense: 10,
		MagicDefense:    20,
		Speed:           13,
		HealingPower:    24,
		CritRate:        0.20,
		CritDamage:      0.50,
		BlockRate:       0.05,
	},
}
