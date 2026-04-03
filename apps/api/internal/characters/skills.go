package characters

import (
	"slices"
	"sort"
	"strings"

	"clawgame/apps/api/internal/auth"
)

const maxActiveSkillSlots = 4

type skillDefinition struct {
	SkillID        string
	Name           string
	DisplayNameZH  string
	Class          string
	RouteID        string
	Track          string
	Tier           string
	CooldownRounds int
	IsBasic        bool
}

var skillUpgradeCosts = map[int]int{
	0: 120,
	1: 180,
	2: 300,
	3: 480,
	4: 750,
	5: 1140,
	6: 1680,
	7: 2400,
	8: 3360,
	9: 4620,
}

var skillDefinitions = map[string]skillDefinition{
	"Strike":            {SkillID: "Strike", Name: "Strike", DisplayNameZH: "重击", Class: "warrior", RouteID: "basic", Track: "basic", Tier: "basic", CooldownRounds: 0, IsBasic: true},
	"Arc Bolt":          {SkillID: "Arc Bolt", Name: "Arc Bolt", DisplayNameZH: "奥术飞弹", Class: "mage", RouteID: "basic", Track: "basic", Tier: "basic", CooldownRounds: 0, IsBasic: true},
	"Smite":             {SkillID: "Smite", Name: "Smite", DisplayNameZH: "惩击", Class: "priest", RouteID: "basic", Track: "basic", Tier: "basic", CooldownRounds: 0, IsBasic: true},
	"Quickstep":         {SkillID: "Quickstep", Name: "Quickstep", DisplayNameZH: "迅步", Class: "universal", RouteID: "universal", Track: "universal", Tier: "normal", CooldownRounds: 1},
	"Pocket Sand":       {SkillID: "Pocket Sand", Name: "Pocket Sand", DisplayNameZH: "扬沙", Class: "universal", RouteID: "universal", Track: "universal", Tier: "normal", CooldownRounds: 1},
	"Emergency Roll":    {SkillID: "Emergency Roll", Name: "Emergency Roll", DisplayNameZH: "紧急翻滚", Class: "universal", RouteID: "universal", Track: "universal", Tier: "advanced", CooldownRounds: 2},
	"Signal Flare":      {SkillID: "Signal Flare", Name: "Signal Flare", DisplayNameZH: "信号照明弹", Class: "universal", RouteID: "universal", Track: "universal", Tier: "advanced", CooldownRounds: 2},
	"Field Tonic":       {SkillID: "Field Tonic", Name: "Field Tonic", DisplayNameZH: "野战提神剂", Class: "universal", RouteID: "universal", Track: "universal", Tier: "advanced", CooldownRounds: 2},
	"Tripwire Kit":      {SkillID: "Tripwire Kit", Name: "Tripwire Kit", DisplayNameZH: "绊索装置", Class: "universal", RouteID: "universal", Track: "universal", Tier: "ultimate", CooldownRounds: 3},
	"Guard Stance":      {SkillID: "Guard Stance", Name: "Guard Stance", DisplayNameZH: "守御姿态", Class: "warrior", RouteID: "shared", Track: "shared", Tier: "normal", CooldownRounds: 1},
	"War Cry":           {SkillID: "War Cry", Name: "War Cry", DisplayNameZH: "战吼", Class: "warrior", RouteID: "shared", Track: "shared", Tier: "normal", CooldownRounds: 1},
	"Intercept":         {SkillID: "Intercept", Name: "Intercept", DisplayNameZH: "拦截", Class: "warrior", RouteID: "shared", Track: "shared", Tier: "advanced", CooldownRounds: 2},
	"Shield Bash":       {SkillID: "Shield Bash", Name: "Shield Bash", DisplayNameZH: "盾击", Class: "warrior", RouteID: "tank", Track: "tank", Tier: "normal", CooldownRounds: 1},
	"Fortified Slash":   {SkillID: "Fortified Slash", Name: "Fortified Slash", DisplayNameZH: "固守斩", Class: "warrior", RouteID: "tank", Track: "tank", Tier: "advanced", CooldownRounds: 2},
	"Bulwark Field":     {SkillID: "Bulwark Field", Name: "Bulwark Field", DisplayNameZH: "壁垒领域", Class: "warrior", RouteID: "tank", Track: "tank", Tier: "advanced", CooldownRounds: 2},
	"Linebreaker":       {SkillID: "Linebreaker", Name: "Linebreaker", DisplayNameZH: "裂阵斩", Class: "warrior", RouteID: "tank", Track: "tank", Tier: "ultimate", CooldownRounds: 3},
	"Cleave":            {SkillID: "Cleave", Name: "Cleave", DisplayNameZH: "劈斩", Class: "warrior", RouteID: "physical_burst", Track: "physical_burst", Tier: "normal", CooldownRounds: 1},
	"Blood Roar":        {SkillID: "Blood Roar", Name: "Blood Roar", DisplayNameZH: "血怒咆哮", Class: "warrior", RouteID: "physical_burst", Track: "physical_burst", Tier: "advanced", CooldownRounds: 2},
	"Execution Rush":    {SkillID: "Execution Rush", Name: "Execution Rush", DisplayNameZH: "处决突进", Class: "warrior", RouteID: "physical_burst", Track: "physical_burst", Tier: "advanced", CooldownRounds: 2},
	"Rending Arc":       {SkillID: "Rending Arc", Name: "Rending Arc", DisplayNameZH: "裂刃圆弧", Class: "warrior", RouteID: "physical_burst", Track: "physical_burst", Tier: "ultimate", CooldownRounds: 3},
	"Runic Brand":       {SkillID: "Runic Brand", Name: "Runic Brand", DisplayNameZH: "符文烙印", Class: "warrior", RouteID: "magic_burst", Track: "magic_burst", Tier: "normal", CooldownRounds: 1},
	"Arcsteel Surge":    {SkillID: "Arcsteel Surge", Name: "Arcsteel Surge", DisplayNameZH: "奥钢奔流", Class: "warrior", RouteID: "magic_burst", Track: "magic_burst", Tier: "advanced", CooldownRounds: 2},
	"Spellrend Wave":    {SkillID: "Spellrend Wave", Name: "Spellrend Wave", DisplayNameZH: "裂法波", Class: "warrior", RouteID: "magic_burst", Track: "magic_burst", Tier: "advanced", CooldownRounds: 2},
	"Astral Breaker":    {SkillID: "Astral Breaker", Name: "Astral Breaker", DisplayNameZH: "星界破灭", Class: "warrior", RouteID: "magic_burst", Track: "magic_burst", Tier: "ultimate", CooldownRounds: 3},
	"Arc Veil":          {SkillID: "Arc Veil", Name: "Arc Veil", DisplayNameZH: "奥术帷幕", Class: "mage", RouteID: "shared", Track: "shared", Tier: "normal", CooldownRounds: 1},
	"Focus Pulse":       {SkillID: "Focus Pulse", Name: "Focus Pulse", DisplayNameZH: "聚能脉冲", Class: "mage", RouteID: "shared", Track: "shared", Tier: "normal", CooldownRounds: 1},
	"Disrupt Ray":       {SkillID: "Disrupt Ray", Name: "Disrupt Ray", DisplayNameZH: "扰乱射线", Class: "mage", RouteID: "shared", Track: "shared", Tier: "advanced", CooldownRounds: 2},
	"Hex Mark":          {SkillID: "Hex Mark", Name: "Hex Mark", DisplayNameZH: "咒印", Class: "mage", RouteID: "single_burst", Track: "single_burst", Tier: "normal", CooldownRounds: 1},
	"Seal Fracture":     {SkillID: "Seal Fracture", Name: "Seal Fracture", DisplayNameZH: "封印裂解", Class: "mage", RouteID: "single_burst", Track: "single_burst", Tier: "advanced", CooldownRounds: 2},
	"Detonate Sigil":    {SkillID: "Detonate Sigil", Name: "Detonate Sigil", DisplayNameZH: "引爆秘印", Class: "mage", RouteID: "single_burst", Track: "single_burst", Tier: "advanced", CooldownRounds: 2},
	"Star Pierce":       {SkillID: "Star Pierce", Name: "Star Pierce", DisplayNameZH: "星芒穿刺", Class: "mage", RouteID: "single_burst", Track: "single_burst", Tier: "ultimate", CooldownRounds: 3},
	"Flame Burst":       {SkillID: "Flame Burst", Name: "Flame Burst", DisplayNameZH: "烈焰爆裂", Class: "mage", RouteID: "aoe_burst", Track: "aoe_burst", Tier: "normal", CooldownRounds: 1},
	"Meteor Shard":      {SkillID: "Meteor Shard", Name: "Meteor Shard", DisplayNameZH: "陨星碎片", Class: "mage", RouteID: "aoe_burst", Track: "aoe_burst", Tier: "advanced", CooldownRounds: 2},
	"Chain Script":      {SkillID: "Chain Script", Name: "Chain Script", DisplayNameZH: "连锁咒文", Class: "mage", RouteID: "aoe_burst", Track: "aoe_burst", Tier: "advanced", CooldownRounds: 2},
	"Ember Field":       {SkillID: "Ember Field", Name: "Ember Field", DisplayNameZH: "余烬领域", Class: "mage", RouteID: "aoe_burst", Track: "aoe_burst", Tier: "ultimate", CooldownRounds: 3},
	"Frost Bind":        {SkillID: "Frost Bind", Name: "Frost Bind", DisplayNameZH: "寒霜束缚", Class: "mage", RouteID: "control", Track: "control", Tier: "normal", CooldownRounds: 1},
	"Gravity Knot":      {SkillID: "Gravity Knot", Name: "Gravity Knot", DisplayNameZH: "引力结", Class: "mage", RouteID: "control", Track: "control", Tier: "advanced", CooldownRounds: 2},
	"Silencing Prism":   {SkillID: "Silencing Prism", Name: "Silencing Prism", DisplayNameZH: "缄默棱镜", Class: "mage", RouteID: "control", Track: "control", Tier: "advanced", CooldownRounds: 2},
	"Time Lock":         {SkillID: "Time Lock", Name: "Time Lock", DisplayNameZH: "时滞锁", Class: "mage", RouteID: "control", Track: "control", Tier: "ultimate", CooldownRounds: 3},
	"Restore":           {SkillID: "Restore", Name: "Restore", DisplayNameZH: "恢复术", Class: "priest", RouteID: "shared", Track: "shared", Tier: "normal", CooldownRounds: 1},
	"Sanctuary Mark":    {SkillID: "Sanctuary Mark", Name: "Sanctuary Mark", DisplayNameZH: "圣护印记", Class: "priest", RouteID: "shared", Track: "shared", Tier: "normal", CooldownRounds: 1},
	"Purge":             {SkillID: "Purge", Name: "Purge", DisplayNameZH: "净除", Class: "priest", RouteID: "shared", Track: "shared", Tier: "advanced", CooldownRounds: 2},
	"Grace Field":       {SkillID: "Grace Field", Name: "Grace Field", DisplayNameZH: "恩泽领域", Class: "priest", RouteID: "healing_support", Track: "healing_support", Tier: "normal", CooldownRounds: 1},
	"Purifying Wave":    {SkillID: "Purifying Wave", Name: "Purifying Wave", DisplayNameZH: "净化之潮", Class: "priest", RouteID: "healing_support", Track: "healing_support", Tier: "advanced", CooldownRounds: 2},
	"Prayer of Renewal": {SkillID: "Prayer of Renewal", Name: "Prayer of Renewal", DisplayNameZH: "复苏祷言", Class: "priest", RouteID: "healing_support", Track: "healing_support", Tier: "advanced", CooldownRounds: 2},
	"Bless Armor":       {SkillID: "Bless Armor", Name: "Bless Armor", DisplayNameZH: "圣佑护甲", Class: "priest", RouteID: "healing_support", Track: "healing_support", Tier: "ultimate", CooldownRounds: 3},
	"Judged Weakness":   {SkillID: "Judged Weakness", Name: "Judged Weakness", DisplayNameZH: "弱点审判", Class: "priest", RouteID: "curse", Track: "curse", Tier: "normal", CooldownRounds: 1},
	"Seal of Silence":   {SkillID: "Seal of Silence", Name: "Seal of Silence", DisplayNameZH: "沉默封印", Class: "priest", RouteID: "curse", Track: "curse", Tier: "advanced", CooldownRounds: 2},
	"Wither Prayer":     {SkillID: "Wither Prayer", Name: "Wither Prayer", DisplayNameZH: "枯败祷言", Class: "priest", RouteID: "curse", Track: "curse", Tier: "advanced", CooldownRounds: 2},
	"Judgment":          {SkillID: "Judgment", Name: "Judgment", DisplayNameZH: "裁决", Class: "priest", RouteID: "curse", Track: "curse", Tier: "ultimate", CooldownRounds: 3},
	"Sanctified Blow":   {SkillID: "Sanctified Blow", Name: "Sanctified Blow", DisplayNameZH: "圣化打击", Class: "priest", RouteID: "summon", Track: "summon", Tier: "normal", CooldownRounds: 1},
	"Lantern Servitor":  {SkillID: "Lantern Servitor", Name: "Lantern Servitor", DisplayNameZH: "灯灵侍从", Class: "priest", RouteID: "summon", Track: "summon", Tier: "advanced", CooldownRounds: 2},
	"Censer Guardian":   {SkillID: "Censer Guardian", Name: "Censer Guardian", DisplayNameZH: "香炉守卫", Class: "priest", RouteID: "summon", Track: "summon", Tier: "advanced", CooldownRounds: 2},
	"Choir Invocation":  {SkillID: "Choir Invocation", Name: "Choir Invocation", DisplayNameZH: "圣咏召唤", Class: "priest", RouteID: "summon", Track: "summon", Tier: "ultimate", CooldownRounds: 3},
}

func cloneSkillLevels(source map[string]int) map[string]int {
	if len(source) == 0 {
		return map[string]int{}
	}
	cloned := make(map[string]int, len(source))
	for key, value := range source {
		if _, ok := skillDefinitions[key]; !ok {
			continue
		}
		if value < 0 {
			value = 0
		}
		if value > 10 {
			value = 10
		}
		cloned[key] = value
	}
	return cloned
}

func cloneSkillLoadout(source []string) []string {
	if len(source) == 0 {
		return []string{}
	}
	cloned := make([]string, 0, len(source))
	seen := map[string]struct{}{}
	for _, skillID := range source {
		if _, ok := skillDefinitions[skillID]; !ok {
			continue
		}
		if _, exists := seen[skillID]; exists {
			continue
		}
		seen[skillID] = struct{}{}
		cloned = append(cloned, skillID)
		if len(cloned) >= maxActiveSkillSlots {
			break
		}
	}
	return cloned
}

func (s *Service) SkillsState(account auth.Account) (SkillsStateView, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.characterByAccountID[account.AccountID]
	if !ok {
		return SkillsStateView{}, ErrCharacterNotFound
	}

	return buildSkillsState(entry.summary, entry.skillLevels, entry.skillLoadout), nil
}

func (s *Service) UpgradeSkill(account auth.Account, skillID string) (SkillsStateView, Summary, error) {
	skillID = strings.TrimSpace(skillID)
	definition, ok := skillDefinitions[skillID]
	if !ok || definition.IsBasic {
		return SkillsStateView{}, Summary{}, ErrSkillNotFound
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.characterByAccountID[account.AccountID]
	if !ok {
		return SkillsStateView{}, Summary{}, ErrCharacterNotFound
	}

	if !canAccessSkill(entry.summary, definition) {
		return SkillsStateView{}, Summary{}, ErrSkillLocked
	}

	if entry.skillLevels == nil {
		entry.skillLevels = map[string]int{}
	}

	currentLevel := entry.skillLevels[skillID]
	if currentLevel >= 10 {
		return SkillsStateView{}, Summary{}, ErrSkillMaxLevel
	}

	cost, ok := skillUpgradeCosts[currentLevel]
	if !ok {
		return SkillsStateView{}, Summary{}, ErrSkillMaxLevel
	}
	if entry.summary.Gold < cost {
		return SkillsStateView{}, Summary{}, ErrGoldInsufficient
	}

	entry.summary.Gold -= cost
	entry.skillLevels[skillID] = currentLevel + 1
	if err := s.saveRecordLocked(account.AccountID, entry); err != nil {
		return SkillsStateView{}, Summary{}, err
	}
	s.characterByAccountID[account.AccountID] = entry

	return buildSkillsState(entry.summary, entry.skillLevels, entry.skillLoadout), entry.summary, nil
}

func (s *Service) SetSkillLoadout(account auth.Account, skillIDs []string) (SkillsStateView, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.characterByAccountID[account.AccountID]
	if !ok {
		return SkillsStateView{}, ErrCharacterNotFound
	}

	loadout, err := validateSkillLoadout(entry.summary, entry.skillLevels, skillIDs)
	if err != nil {
		return SkillsStateView{}, err
	}
	entry.skillLoadout = loadout
	if err := s.saveRecordLocked(account.AccountID, entry); err != nil {
		return SkillsStateView{}, err
	}
	s.characterByAccountID[account.AccountID] = entry

	return buildSkillsState(entry.summary, entry.skillLevels, entry.skillLoadout), nil
}

func buildSkillsState(summary Summary, skillLevels map[string]int, skillLoadout []string) SkillsStateView {
	levels := cloneSkillLevels(skillLevels)
	loadout := cloneSkillLoadout(skillLoadout)
	universal := make([]SkillView, 0, 6)
	classSkills := make([]SkillView, 0, 16)

	for _, definition := range orderedSkillDefinitions() {
		if definition.IsBasic {
			continue
		}
		view := skillViewFromDefinition(summary, definition, levels[definition.SkillID])
		if definition.Class == "universal" {
			universal = append(universal, view)
			continue
		}
		if definition.Class == summary.Class {
			classSkills = append(classSkills, view)
		}
	}

	return SkillsStateView{
		BasicAttack:    basicAttackView(summary),
		Universal:      universal,
		ClassSkills:    classSkills,
		ActiveLoadout:  loadout,
		MaxActiveSlots: maxActiveSkillSlots,
	}
}

func orderedSkillDefinitions() []skillDefinition {
	items := make([]skillDefinition, 0, len(skillDefinitions))
	for _, definition := range skillDefinitions {
		items = append(items, definition)
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Class != items[j].Class {
			return items[i].Class < items[j].Class
		}
		if items[i].RouteID != items[j].RouteID {
			return items[i].RouteID < items[j].RouteID
		}
		return items[i].SkillID < items[j].SkillID
	})
	return items
}

func basicAttackView(summary Summary) SkillView {
	className := strings.ToLower(strings.TrimSpace(summary.Class))
	basicID := "Strike"
	switch className {
	case "mage":
		basicID = "Arc Bolt"
	case "priest":
		basicID = "Smite"
	}
	return skillViewFromDefinition(summary, skillDefinitions[basicID], 1)
}

func skillViewFromDefinition(summary Summary, definition skillDefinition, level int) SkillView {
	isUnlocked := definition.IsBasic
	if !definition.IsBasic {
		isUnlocked = canAccessSkill(summary, definition) && level > 0
	}
	currentMultiplier := 100 + level*2
	if definition.IsBasic {
		currentMultiplier = 100
	}
	nextCost := 0
	if !definition.IsBasic && level < 10 && canAccessSkill(summary, definition) {
		nextCost = skillUpgradeCosts[level]
	}
	return SkillView{
		SkillID:           definition.SkillID,
		Name:              definition.Name,
		DisplayNameZH:     definition.DisplayNameZH,
		Class:             definition.Class,
		RouteID:           definition.RouteID,
		Track:             definition.Track,
		Tier:              definition.Tier,
		CooldownRounds:    definition.CooldownRounds,
		IsBasic:           definition.IsBasic,
		IsUnlocked:        isUnlocked,
		Level:             level,
		MaxLevel:          10,
		CurrentMultiplier: currentMultiplier,
		NextLevelCost:     nextCost,
	}
}

func canAccessSkill(summary Summary, definition skillDefinition) bool {
	if definition.IsBasic {
		return true
	}
	if definition.Class == "universal" {
		return true
	}
	if strings.EqualFold(summary.Class, "civilian") {
		return false
	}
	return strings.EqualFold(summary.Class, definition.Class)
}

func validateSkillLoadout(summary Summary, skillLevels map[string]int, skillIDs []string) ([]string, error) {
	if len(skillIDs) > maxActiveSkillSlots {
		return nil, ErrSkillInvalidLoadout
	}
	loadout := make([]string, 0, len(skillIDs))
	seen := map[string]struct{}{}
	for _, skillID := range skillIDs {
		skillID = strings.TrimSpace(skillID)
		definition, ok := skillDefinitions[skillID]
		if !ok || definition.IsBasic {
			return nil, ErrSkillInvalidLoadout
		}
		if _, exists := seen[skillID]; exists {
			return nil, ErrSkillInvalidLoadout
		}
		if !canAccessSkill(summary, definition) || skillLevels[skillID] <= 0 {
			return nil, ErrSkillLocked
		}
		seen[skillID] = struct{}{}
		loadout = append(loadout, skillID)
	}
	return loadout, nil
}

func SkillBuildingActions(actions []string) []string {
	items := slices.Clone(actions)
	for _, action := range []string{"view_skills", "upgrade_skill", "set_skill_loadout", "choose_profession_route"} {
		if !slices.Contains(items, action) {
			items = append(items, action)
		}
	}
	return items
}
