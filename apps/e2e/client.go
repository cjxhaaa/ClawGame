package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	token      string
}

type Envelope[T any] struct {
	Data T `json:"data"`
}

type Challenge struct {
	ChallengeID string `json:"challenge_id"`
	PromptText  string `json:"prompt_text"`
}

type Quest struct {
	QuestID          string `json:"quest_id"`
	TemplateType     string `json:"template_type"`
	ContractType     string `json:"contract_type"`
	Difficulty       string `json:"difficulty"`
	FlowKind         string `json:"flow_kind"`
	TargetRegionID   string `json:"target_region_id"`
	Status           string `json:"status"`
	RewardGold       int    `json:"reward_gold"`
	RewardReputation int    `json:"reward_reputation"`
}

type QuestChoice struct {
	ChoiceKey string `json:"choice_key"`
	Label     string `json:"label"`
}

type QuestRuntime struct {
	CurrentStepKey      string         `json:"current_step_key"`
	SuggestedActionType string         `json:"suggested_action_type"`
	SuggestedActionArgs map[string]any `json:"suggested_action_args"`
	AvailableChoices    []QuestChoice  `json:"available_choices"`
	Clues               any            `json:"clues"`
}

type QuestDetail struct {
	Quest   Quest        `json:"quest"`
	Runtime QuestRuntime `json:"runtime"`
}

type Character struct {
	CharacterID      string `json:"character_id"`
	Name             string `json:"name"`
	Class            string `json:"class"`
	WeaponStyle      string `json:"weapon_style"`
	SeasonLevel      int    `json:"season_level"`
	SeasonXP         int    `json:"season_xp"`
	Reputation       int    `json:"reputation"`
	Gold             int    `json:"gold"`
	LocationRegionID string `json:"location_region_id"`
}

type DailyLimits struct {
	QuestCompletionCap         int `json:"quest_completion_cap"`
	QuestCompletionUsed        int `json:"quest_completion_used"`
	DungeonEntryCap            int `json:"dungeon_entry_cap"`
	DungeonEntryUsed           int `json:"dungeon_entry_used"`
	FreeDungeonEntryCap        int `json:"free_dungeon_entry_cap"`
	BonusDungeonEntryPurchased int `json:"bonus_dungeon_entry_purchased"`
	ReputationPerBonusClaim    int `json:"reputation_per_bonus_claim"`
}

type State struct {
	Character   Character         `json:"character"`
	Limits      DailyLimits       `json:"limits"`
	CombatPower StateCombatPower  `json:"combat_power"`
	Objectives  []Quest           `json:"objectives"`
	Materials   []MaterialBalance `json:"materials"`
}

type StateCombatPower struct {
	PanelPowerScore int    `json:"panel_power_score"`
	PowerTier       string `json:"power_tier"`
}

type MaterialBalance struct {
	MaterialKey string `json:"material_key"`
	Quantity    int    `json:"quantity"`
}

type Run struct {
	RunID              string  `json:"run_id"`
	DungeonID          string  `json:"dungeon_id"`
	Difficulty         string  `json:"difficulty"`
	RunStatus          string  `json:"run_status"`
	RewardClaimable    bool    `json:"reward_claimable"`
	HighestRoomCleared int     `json:"highest_room_cleared"`
	CurrentRating      *string `json:"current_rating"`
	RuntimePhase       string  `json:"runtime_phase"`
}

type InventoryItem struct {
	ItemID              string         `json:"item_id"`
	CatalogID           string         `json:"catalog_id"`
	Name                string         `json:"name"`
	Slot                string         `json:"slot"`
	Rarity              string         `json:"rarity"`
	SetID               string         `json:"set_id"`
	RequiredClass       string         `json:"required_class"`
	RequiredWeaponStyle string         `json:"required_weapon_style"`
	Stats               map[string]int `json:"stats"`
}

type UpgradeHint struct {
	Source            string `json:"source"`
	ItemID            string `json:"item_id"`
	CatalogID         string `json:"catalog_id"`
	Name              string `json:"name"`
	Slot              string `json:"slot"`
	ScoreDelta        int    `json:"score_delta"`
	Affordable        bool   `json:"affordable"`
	DirectlyEquipable bool   `json:"directly_equippable"`
}

type InventoryView struct {
	EquipmentScore int             `json:"equipment_score"`
	Equipped       []InventoryItem `json:"equipped"`
	Inventory      []InventoryItem `json:"inventory"`
	UpgradeHints   []UpgradeHint   `json:"upgrade_hints"`
}

type ShopInventoryItem struct {
	CatalogID string         `json:"catalog_id"`
	Name      string         `json:"name"`
	ItemType  string         `json:"item_type"`
	Slot      string         `json:"slot"`
	Rarity    string         `json:"rarity"`
	PriceGold int            `json:"price_gold"`
	Stats     map[string]int `json:"stats"`
}

type FieldEncounterResult struct {
	ActionResult struct {
		ActionType string `json:"action_type"`
		Victory    bool   `json:"victory"`
	} `json:"action_result"`
	State State `json:"state"`
}

type RunsResponse struct {
	Items []Run `json:"items"`
}

type WorldBossConfig struct {
	BossID            string `json:"boss_id"`
	Name              string `json:"name"`
	RequiredPartySize int    `json:"required_party_size"`
}

type WorldBossQueueStatus struct {
	BossID             string   `json:"boss_id"`
	CharacterID        string   `json:"character_id"`
	Queued             bool     `json:"queued"`
	CurrentQueuedCount int      `json:"current_queued_count"`
	RequiredPartySize  int      `json:"required_party_size"`
	LastRaidID         string   `json:"last_raid_id"`
	LastRewardTier     string   `json:"last_reward_tier"`
	PendingRaidIDs     []string `json:"pending_raid_ids"`
}

type WorldBossRaidMember struct {
	CharacterID string `json:"character_id"`
	Name        string `json:"name"`
	Class       string `json:"class"`
	DamageDealt int    `json:"damage_dealt"`
	Survived    bool   `json:"survived"`
	PowerScore  int    `json:"power_score"`
}

type WorldBossRewardTier struct {
	Tier           string           `json:"tier"`
	RequiredDamage int              `json:"required_damage"`
	RewardGold     int              `json:"reward_gold"`
	MaterialDrops  []map[string]any `json:"material_drops"`
}

type WorldBossRaid struct {
	RaidID        string                `json:"raid_id"`
	BossID        string                `json:"boss_id"`
	BossName      string                `json:"boss_name"`
	Status        string                `json:"status"`
	TotalDamage   int                   `json:"total_damage"`
	BossMaxHP     int                   `json:"boss_max_hp"`
	RewardTier    string                `json:"reward_tier"`
	RewardPackage WorldBossRewardTier   `json:"reward_package"`
	Members       []WorldBossRaidMember `json:"members"`
	ResolvedAt    string                `json:"resolved_at"`
}

type WorldBossJoinResult struct {
	Status              WorldBossQueueStatus `json:"status"`
	MatchedCharacterIDs []string             `json:"matched_character_ids"`
	ResolvedRaid        *WorldBossRaid       `json:"resolved_raid"`
}

func NewClient(baseURL string, httpClient *http.Client) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: httpClient,
	}
}

func (c *Client) Clone() *Client {
	return &Client{
		baseURL:    c.baseURL,
		httpClient: c.httpClient,
		token:      c.token,
	}
}

func (c *Client) Register(botName, password string) error {
	challenge, err := c.IssueChallenge()
	if err != nil {
		return err
	}
	return c.doJSON(http.MethodPost, "/api/v1/auth/register", map[string]any{
		"bot_name":         botName,
		"password":         password,
		"challenge_id":     challenge.ChallengeID,
		"challenge_answer": solveChallengePrompt(challenge.PromptText),
	}, nil)
}

func (c *Client) Login(botName, password string) error {
	challenge, err := c.IssueChallenge()
	if err != nil {
		return err
	}

	var response Envelope[struct {
		AccessToken string `json:"access_token"`
	}]
	if err := c.doJSON(http.MethodPost, "/api/v1/auth/login", map[string]any{
		"bot_name":         botName,
		"password":         password,
		"challenge_id":     challenge.ChallengeID,
		"challenge_answer": solveChallengePrompt(challenge.PromptText),
	}, &response); err != nil {
		return err
	}
	c.token = response.Data.AccessToken
	return nil
}

func (c *Client) CreateCharacter(name, class, weaponStyle string) error {
	return c.doJSON(http.MethodPost, "/api/v1/characters", map[string]any{
		"name": name,
	}, nil)
}

func (c *Client) IssueChallenge() (Challenge, error) {
	var response Envelope[struct {
		Challenge Challenge `json:"challenge"`
	}]
	if err := c.doJSON(http.MethodPost, "/api/v1/auth/challenge", map[string]any{}, &response); err != nil {
		return Challenge{}, err
	}
	return response.Data.Challenge, nil
}

func (c *Client) State() (State, error) {
	var response Envelope[State]
	if err := c.doJSON(http.MethodGet, "/api/v1/me/state", nil, &response); err != nil {
		return State{}, err
	}
	return response.Data, nil
}

func (c *Client) Quests() ([]Quest, error) {
	var response Envelope[struct {
		Quests []Quest `json:"quests"`
	}]
	if err := c.doJSON(http.MethodGet, "/api/v1/me/quests", nil, &response); err != nil {
		return nil, err
	}
	return response.Data.Quests, nil
}

func (c *Client) QuestDetail(questID string) (QuestDetail, error) {
	var response Envelope[QuestDetail]
	if err := c.doJSON(http.MethodGet, "/api/v1/me/quests/"+questID, nil, &response); err != nil {
		return QuestDetail{}, err
	}
	return response.Data, nil
}

func (c *Client) QuestInteract(questID, interaction string) error {
	return c.doJSON(http.MethodPost, "/api/v1/me/quests/"+questID+"/interact", map[string]any{
		"interaction": interaction,
	}, nil)
}

func (c *Client) QuestChoice(questID, choiceKey string) error {
	return c.doJSON(http.MethodPost, "/api/v1/me/quests/"+questID+"/choice", map[string]any{
		"choice_key": choiceKey,
	}, nil)
}

func (c *Client) ChooseProfession(className, weaponStyle string) error {
	routeID, ok := legacyRouteID(className, weaponStyle)
	if !ok {
		return fmt.Errorf("unsupported class/weapon style combination %s/%s", className, weaponStyle)
	}
	return c.doJSON(http.MethodPost, "/api/v1/me/profession-route", map[string]any{
		"route_id": routeID,
	}, nil)
}

func (c *Client) SubmitQuest(questID string) error {
	return c.doJSON(http.MethodPost, "/api/v1/me/quests/"+questID+"/submit", map[string]any{}, nil)
}

func (c *Client) Travel(regionID string) error {
	return c.doJSON(http.MethodPost, "/api/v1/me/travel", map[string]any{
		"region_id": regionID,
	}, nil)
}

func (c *Client) FieldEncounter(approach string) (FieldEncounterResult, error) {
	var response Envelope[FieldEncounterResult]
	if err := c.doJSON(http.MethodPost, "/api/v1/me/field-encounter", map[string]any{
		"approach": approach,
	}, &response); err != nil {
		return FieldEncounterResult{}, err
	}
	return response.Data, nil
}

func (c *Client) EnterDungeon(dungeonID, difficulty string, potionLoadout []string) (Run, error) {
	var response Envelope[Run]
	if err := c.doJSON(http.MethodPost, "/api/v1/dungeons/"+dungeonID+"/enter", map[string]any{
		"difficulty":     difficulty,
		"potion_loadout": potionLoadout,
	}, &response); err != nil {
		return Run{}, err
	}
	return response.Data, nil
}

func (c *Client) ClaimDungeonRewards(runID string) (Run, error) {
	var response Envelope[Run]
	if err := c.doJSON(http.MethodPost, "/api/v1/me/runs/"+runID+"/claim", nil, &response); err != nil {
		return Run{}, err
	}
	return response.Data, nil
}

func (c *Client) ListRuns() ([]Run, error) {
	var response Envelope[RunsResponse]
	if err := c.doJSON(http.MethodGet, "/api/v1/me/runs", nil, &response); err != nil {
		return nil, err
	}
	return response.Data.Items, nil
}

func (c *Client) ExchangeDungeonRewardClaims(quantity int) (State, error) {
	var response Envelope[struct {
		Character Character   `json:"character"`
		Limits    DailyLimits `json:"limits"`
	}]
	if err := c.doJSON(http.MethodPost, "/api/v1/me/dungeons/reward-claims/exchange", map[string]any{
		"quantity": quantity,
	}, &response); err != nil {
		return State{}, err
	}
	return State{
		Character: response.Data.Character,
		Limits:    response.Data.Limits,
	}, nil
}

func (c *Client) Inventory() (InventoryView, error) {
	var response Envelope[InventoryView]
	if err := c.doJSON(http.MethodGet, "/api/v1/me/inventory", nil, &response); err != nil {
		return InventoryView{}, err
	}
	return response.Data, nil
}

func (c *Client) EquipItem(itemID string) (InventoryView, error) {
	var response Envelope[InventoryView]
	if err := c.doJSON(http.MethodPost, "/api/v1/me/equipment/equip", map[string]any{
		"item_id": itemID,
	}, &response); err != nil {
		return InventoryView{}, err
	}
	return response.Data, nil
}

func (c *Client) BuildingShopInventory(buildingID string) ([]ShopInventoryItem, error) {
	var response Envelope[struct {
		Items []ShopInventoryItem `json:"items"`
	}]
	if err := c.doJSON(http.MethodGet, "/api/v1/buildings/"+buildingID+"/shop-inventory", nil, &response); err != nil {
		return nil, err
	}
	return response.Data.Items, nil
}

func (c *Client) PurchaseBuildingItem(buildingID, catalogID string) error {
	return c.doJSON(http.MethodPost, "/api/v1/buildings/"+buildingID+"/purchase", map[string]any{
		"catalog_id": catalogID,
	}, nil)
}

func (c *Client) WorldBossCurrent() (WorldBossConfig, error) {
	var response Envelope[WorldBossConfig]
	if err := c.doJSON(http.MethodGet, "/api/v1/world-boss/current", nil, &response); err != nil {
		return WorldBossConfig{}, err
	}
	return response.Data, nil
}

func (c *Client) WorldBossQueueStatus() (WorldBossQueueStatus, error) {
	var response Envelope[WorldBossQueueStatus]
	if err := c.doJSON(http.MethodGet, "/api/v1/world-boss/queue-status", nil, &response); err != nil {
		return WorldBossQueueStatus{}, err
	}
	return response.Data, nil
}

func (c *Client) JoinWorldBossQueue() (WorldBossJoinResult, error) {
	var response Envelope[WorldBossJoinResult]
	if err := c.doJSON(http.MethodPost, "/api/v1/world-boss/queue", nil, &response); err != nil {
		return WorldBossJoinResult{}, err
	}
	return response.Data, nil
}

func (c *Client) WorldBossRaid(raidID string) (WorldBossRaid, error) {
	var response Envelope[WorldBossRaid]
	if err := c.doJSON(http.MethodGet, "/api/v1/world-boss/raids/"+raidID, nil, &response); err != nil {
		return WorldBossRaid{}, err
	}
	return response.Data, nil
}

func (c *Client) doJSON(method, path string, body any, target any) error {
	var payload io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return err
		}
		payload = bytes.NewReader(raw)
	}

	request, err := http.NewRequest(method, c.baseURL+path, payload)
	if err != nil {
		return err
	}
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if strings.TrimSpace(c.token) != "" {
		request.Header.Set("Authorization", "Bearer "+c.token)
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("%s %s returned %d: %s", method, path, response.StatusCode, strings.TrimSpace(string(data)))
	}
	if target == nil || len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, target)
}

func solveChallengePrompt(prompt string) string {
	matcher := regexp.MustCompile(`ember=(\d+).+frost=(\d+).+moss=(\d+).+factor=(\d+)`)
	matches := matcher.FindStringSubmatch(prompt)
	if len(matches) != 5 {
		return ""
	}

	ember, _ := strconv.Atoi(matches[1])
	frost, _ := strconv.Atoi(matches[2])
	moss, _ := strconv.Atoi(matches[3])
	factor, _ := strconv.Atoi(matches[4])
	return strconv.Itoa(((ember + frost) - moss) * factor)
}

func legacyRouteID(className, weaponStyle string) (string, bool) {
	switch {
	case className == "warrior" && weaponStyle == "sword_shield":
		return "tank", true
	case className == "warrior" && weaponStyle == "great_axe":
		return "physical_burst", true
	case className == "mage" && weaponStyle == "staff":
		return "aoe_burst", true
	case className == "mage" && weaponStyle == "spellbook":
		return "single_burst", true
	case className == "priest" && weaponStyle == "scepter":
		return "curse", true
	case className == "priest" && weaponStyle == "holy_tome":
		return "healing_support", true
	default:
		return "", false
	}
}
