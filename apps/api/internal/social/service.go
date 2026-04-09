package social

import (
	"errors"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

const businessTimezone = "Asia/Shanghai"

var (
	ErrFollowSelf         = errors.New("cannot follow self")
	ErrChatMessageEmpty   = errors.New("chat message empty")
	ErrChatMessageTooLong = errors.New("chat message too long")
	ErrChatDailyCap       = errors.New("chat daily cap reached")
	ErrChatCooldown       = errors.New("chat cooldown active")
)

type SocialSummary struct {
	FollowingCount              int  `json:"following_count"`
	FollowerCount               int  `json:"follower_count"`
	FriendCount                 int  `json:"friend_count"`
	HasBorrowableAssistTemplate bool `json:"has_borrowable_assist_template"`
}

type AssistTemplate struct {
	BotID        string `json:"bot_id"`
	TemplateName string `json:"template_name"`
	IsSubmitted  bool   `json:"is_submitted"`
	UpdatedAt    string `json:"updated_at"`
}

type ChatMessage struct {
	MessageID   string `json:"message_id"`
	ChannelType string `json:"channel_type"`
	RegionID    string `json:"region_id,omitempty"`
	BotID       string `json:"bot_id"`
	BotName     string `json:"bot_name"`
	MessageType string `json:"message_type"`
	Content     string `json:"content"`
	CreatedAt   string `json:"created_at"`
}

type StoredFollow struct {
	ActorID   string
	TargetID  string
	CreatedAt string
}

type StoredChatQuota struct {
	BotID          string
	ResetDate      string
	WorldCount     int
	RegionCount    int
	WorldCooldown  string
	RegionCooldown string
}

type Repository interface {
	LoadFollows() ([]StoredFollow, error)
	SaveFollow(StoredFollow) error
	DeleteFollow(actorID, targetID string) error
	LoadAssistTemplates() ([]AssistTemplate, error)
	SaveAssistTemplate(AssistTemplate) error
	LoadChatMessages() ([]ChatMessage, error)
	LoadChatQuotas() ([]StoredChatQuota, error)
	SaveChatMessageAndQuota(ChatMessage, StoredChatQuota) error
}

type chatQuota struct {
	ResetDate      string
	WorldCount     int
	RegionCount    int
	WorldCooldown  time.Time
	RegionCooldown time.Time
}

type Service struct {
	mu              sync.RWMutex
	clock           func() time.Time
	loc             *time.Location
	repo            Repository
	follows         map[string]map[string]time.Time
	assistTemplates map[string]AssistTemplate
	worldMessages   []ChatMessage
	regionMessages  map[string][]ChatMessage
	chatQuotaByBot  map[string]chatQuota
	nextMessageID   int64
}

func NewService() *Service {
	service, err := NewServiceWithRepository(nil)
	if err != nil {
		panic(err)
	}
	return service
}

func NewServiceWithRepository(repo Repository) (*Service, error) {
	service := &Service{
		clock:           time.Now,
		loc:             mustLocation(businessTimezone),
		repo:            repo,
		follows:         make(map[string]map[string]time.Time),
		assistTemplates: make(map[string]AssistTemplate),
		worldMessages:   []ChatMessage{},
		regionMessages:  make(map[string][]ChatMessage),
		chatQuotaByBot:  make(map[string]chatQuota),
	}
	if repo == nil {
		return service, nil
	}

	follows, err := repo.LoadFollows()
	if err != nil {
		return nil, err
	}
	for _, stored := range follows {
		actorID := strings.TrimSpace(stored.ActorID)
		targetID := strings.TrimSpace(stored.TargetID)
		if actorID == "" || targetID == "" {
			continue
		}
		if service.follows[actorID] == nil {
			service.follows[actorID] = make(map[string]time.Time)
		}
		service.follows[actorID][targetID] = parseRFC3339(stored.CreatedAt, service.loc)
	}

	templates, err := repo.LoadAssistTemplates()
	if err != nil {
		return nil, err
	}
	for _, template := range templates {
		botID := strings.TrimSpace(template.BotID)
		if botID == "" {
			continue
		}
		service.assistTemplates[botID] = template
	}

	messages, err := repo.LoadChatMessages()
	if err != nil {
		return nil, err
	}
	slices.SortFunc(messages, func(a, b ChatMessage) int { return strings.Compare(b.CreatedAt, a.CreatedAt) })
	for index := len(messages) - 1; index >= 0; index-- {
		message := messages[index]
		if strings.TrimSpace(message.ChannelType) == "region" {
			regionID := strings.TrimSpace(message.RegionID)
			service.regionMessages[regionID] = prependAndTrim(message, service.regionMessages[regionID], 500)
		} else {
			service.worldMessages = prependAndTrim(message, service.worldMessages, 1000)
		}
	}

	quotas, err := repo.LoadChatQuotas()
	if err != nil {
		return nil, err
	}
	for _, stored := range quotas {
		botID := strings.TrimSpace(stored.BotID)
		if botID == "" {
			continue
		}
		service.chatQuotaByBot[botID] = chatQuota{
			ResetDate:      strings.TrimSpace(stored.ResetDate),
			WorldCount:     stored.WorldCount,
			RegionCount:    stored.RegionCount,
			WorldCooldown:  parseRFC3339(stored.WorldCooldown, service.loc),
			RegionCooldown: parseRFC3339(stored.RegionCooldown, service.loc),
		}
	}

	return service, nil
}

func (s *Service) Follow(actorID, targetID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	actorID = strings.TrimSpace(actorID)
	targetID = strings.TrimSpace(targetID)
	if actorID == "" || targetID == "" {
		return nil
	}
	if actorID == targetID {
		return ErrFollowSelf
	}

	createdAt := s.clock().In(s.loc)
	if s.repo != nil {
		if err := s.repo.SaveFollow(StoredFollow{
			ActorID:   actorID,
			TargetID:  targetID,
			CreatedAt: createdAt.Format(time.RFC3339),
		}); err != nil {
			return err
		}
	}

	if s.follows[actorID] == nil {
		s.follows[actorID] = make(map[string]time.Time)
	}
	s.follows[actorID][targetID] = createdAt
	return nil
}

func (s *Service) Unfollow(actorID, targetID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	actorID = strings.TrimSpace(actorID)
	targetID = strings.TrimSpace(targetID)
	if s.repo != nil && actorID != "" && targetID != "" {
		_ = s.repo.DeleteFollow(actorID, targetID)
	}

	if targets := s.follows[actorID]; targets != nil {
		delete(targets, targetID)
		if len(targets) == 0 {
			delete(s.follows, actorID)
		}
	}
}

func (s *Service) ListFollowing(actorID string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.listFollowingLocked(strings.TrimSpace(actorID))
}

func (s *Service) ListFollowers(targetID string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.listFollowersLocked(strings.TrimSpace(targetID))
}

func (s *Service) ListFriends(actorID string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.listFriendsLocked(strings.TrimSpace(actorID))
}

func (s *Service) RelationStatus(viewerID, targetID string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	viewerID = strings.TrimSpace(viewerID)
	targetID = strings.TrimSpace(targetID)
	if viewerID == "" || targetID == "" || viewerID == targetID {
		return "none"
	}
	_, followsTarget := s.follows[viewerID][targetID]
	_, followedByTarget := s.follows[targetID][viewerID]
	switch {
	case followsTarget && followedByTarget:
		return "friends"
	case followsTarget:
		return "following"
	case followedByTarget:
		return "followed_by"
	default:
		return "none"
	}
}

func (s *Service) SocialSummary(botID string) SocialSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()

	botID = strings.TrimSpace(botID)
	followingCount := len(s.follows[botID])
	followerCount := len(s.listFollowersLocked(botID))
	friendCount := len(s.listFriendsLocked(botID))
	template, ok := s.assistTemplates[botID]
	return SocialSummary{
		FollowingCount:              followingCount,
		FollowerCount:               followerCount,
		FriendCount:                 friendCount,
		HasBorrowableAssistTemplate: ok && template.IsSubmitted,
	}
}

func (s *Service) SetAssistTemplate(botID, templateName string) AssistTemplate {
	s.mu.Lock()
	defer s.mu.Unlock()

	template := AssistTemplate{
		BotID:        strings.TrimSpace(botID),
		TemplateName: strings.TrimSpace(templateName),
		IsSubmitted:  true,
		UpdatedAt:    s.clock().In(s.loc).Format(time.RFC3339),
	}
	if s.repo != nil {
		if err := s.repo.SaveAssistTemplate(template); err != nil {
			if existing, ok := s.assistTemplates[template.BotID]; ok {
				return existing
			}
			return AssistTemplate{
				BotID:        template.BotID,
				TemplateName: "",
				IsSubmitted:  false,
				UpdatedAt:    "",
			}
		}
	}
	s.assistTemplates[template.BotID] = template
	return template
}

func (s *Service) GetAssistTemplate(botID string) AssistTemplate {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if template, ok := s.assistTemplates[strings.TrimSpace(botID)]; ok {
		return template
	}
	return AssistTemplate{
		BotID:        strings.TrimSpace(botID),
		TemplateName: "",
		IsSubmitted:  false,
		UpdatedAt:    "",
	}
}

func (s *Service) PostChat(botID, botName, regionID, channelType, messageType, content string) (ChatMessage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	botID = strings.TrimSpace(botID)
	botName = strings.TrimSpace(botName)
	content = strings.TrimSpace(content)
	if content == "" {
		return ChatMessage{}, ErrChatMessageEmpty
	}
	if len([]rune(content)) > 120 {
		return ChatMessage{}, ErrChatMessageTooLong
	}
	if channelType != "region" {
		channelType = "world"
		regionID = ""
	} else {
		regionID = strings.TrimSpace(regionID)
	}

	quota := s.normalizeQuotaLocked(botID)
	now := s.clock().In(s.loc)
	if channelType == "world" {
		if now.Before(quota.WorldCooldown) {
			return ChatMessage{}, ErrChatCooldown
		}
		if quota.WorldCount >= 100 {
			return ChatMessage{}, ErrChatDailyCap
		}
		quota.WorldCount++
		quota.WorldCooldown = now.Add(10 * time.Second)
	} else {
		if now.Before(quota.RegionCooldown) {
			return ChatMessage{}, ErrChatCooldown
		}
		if quota.RegionCount >= 200 {
			return ChatMessage{}, ErrChatDailyCap
		}
		quota.RegionCount++
		quota.RegionCooldown = now.Add(5 * time.Second)
	}

	s.nextMessageID++
	message := ChatMessage{
		MessageID:   "chat_runtime_" + now.Format("20060102150405") + "_" + botID + "_" + itoa(s.nextMessageID),
		ChannelType: channelType,
		RegionID:    regionID,
		BotID:       botID,
		BotName:     botName,
		MessageType: normalizeMessageType(messageType),
		Content:     content,
		CreatedAt:   now.Format(time.RFC3339),
	}

	if s.repo != nil {
		if err := s.repo.SaveChatMessageAndQuota(message, StoredChatQuota{
			BotID:          botID,
			ResetDate:      quota.ResetDate,
			WorldCount:     quota.WorldCount,
			RegionCount:    quota.RegionCount,
			WorldCooldown:  quota.WorldCooldown.Format(time.RFC3339),
			RegionCooldown: quota.RegionCooldown.Format(time.RFC3339),
		}); err != nil {
			return ChatMessage{}, err
		}
	}

	s.chatQuotaByBot[botID] = quota
	if channelType == "world" {
		s.worldMessages = prependAndTrim(message, s.worldMessages, 1000)
	} else {
		s.regionMessages[message.RegionID] = prependAndTrim(message, s.regionMessages[message.RegionID], 500)
	}

	return message, nil
}

func (s *Service) ListChat(channelType, regionID, messageType string) []ChatMessage {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var source []ChatMessage
	if channelType == "region" {
		source = s.regionMessages[strings.TrimSpace(regionID)]
	} else {
		source = s.worldMessages
	}
	messageType = normalizeMessageTypeFilter(messageType)
	if messageType == "" {
		items := make([]ChatMessage, len(source))
		copy(items, source)
		return items
	}
	items := make([]ChatMessage, 0, len(source))
	for _, item := range source {
		if item.MessageType == messageType {
			items = append(items, item)
		}
	}
	return items
}

func (s *Service) ListChatByBot(botID string, limit int) []ChatMessage {
	s.mu.RLock()
	defer s.mu.RUnlock()

	botID = strings.TrimSpace(botID)
	if limit <= 0 {
		limit = 10
	}
	items := make([]ChatMessage, 0, limit)
	appendFrom := func(source []ChatMessage) {
		for _, item := range source {
			if item.BotID != botID {
				continue
			}
			items = append(items, item)
			if len(items) >= limit {
				return
			}
		}
	}
	appendFrom(s.worldMessages)
	if len(items) < limit {
		for _, source := range s.regionMessages {
			appendFrom(source)
			if len(items) >= limit {
				break
			}
		}
	}
	slices.SortFunc(items, func(a, b ChatMessage) int { return strings.Compare(b.CreatedAt, a.CreatedAt) })
	if len(items) > limit {
		items = items[:limit]
	}
	return items
}

func (s *Service) listFollowingLocked(actorID string) []string {
	targets := s.follows[actorID]
	if len(targets) == 0 {
		return []string{}
	}
	items := make([]string, 0, len(targets))
	for targetID := range targets {
		items = append(items, targetID)
	}
	slices.Sort(items)
	return items
}

func (s *Service) listFollowersLocked(targetID string) []string {
	items := make([]string, 0, len(s.follows))
	for actorID, targets := range s.follows {
		if _, ok := targets[targetID]; ok {
			items = append(items, actorID)
		}
	}
	slices.Sort(items)
	return items
}

func (s *Service) listFriendsLocked(actorID string) []string {
	following := s.follows[actorID]
	if len(following) == 0 {
		return []string{}
	}
	items := make([]string, 0, len(following))
	for targetID := range following {
		if _, ok := s.follows[targetID][actorID]; ok {
			items = append(items, targetID)
		}
	}
	slices.Sort(items)
	return items
}

func (s *Service) normalizeQuotaLocked(botID string) chatQuota {
	quota := s.chatQuotaByBot[botID]
	today := businessDayKey(s.clock().In(s.loc))
	if quota.ResetDate != today {
		quota = chatQuota{ResetDate: today}
	}
	return quota
}

func businessDayKey(now time.Time) string {
	if now.Hour() < 4 {
		now = now.Add(-24 * time.Hour)
	}
	return now.Format("2006-01-02")
}

func normalizeMessageType(value string) string {
	switch strings.TrimSpace(value) {
	case "friend_recruit", "assist_ad":
		return strings.TrimSpace(value)
	default:
		return "free_text"
	}
}

func normalizeMessageTypeFilter(value string) string {
	switch strings.TrimSpace(value) {
	case "free_text", "friend_recruit", "assist_ad":
		return strings.TrimSpace(value)
	default:
		return ""
	}
}

func prependAndTrim(item ChatMessage, source []ChatMessage, max int) []ChatMessage {
	items := make([]ChatMessage, 0, len(source)+1)
	items = append(items, item)
	items = append(items, source...)
	if len(items) > max {
		items = items[:max]
	}
	return items
}

func itoa(v int64) string {
	return strconv.FormatInt(v, 10)
}

func parseRFC3339(raw string, loc *time.Location) time.Time {
	if strings.TrimSpace(raw) == "" {
		return time.Time{}
	}
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}
	}
	return parsed.In(loc)
}

func mustLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic(err)
	}
	return loc
}
