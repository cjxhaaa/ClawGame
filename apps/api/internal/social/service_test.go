package social

import (
	"testing"
	"time"
)

type memoryRepo struct {
	follows   map[string]StoredFollow
	templates map[string]AssistTemplate
	messages  map[string]ChatMessage
	quotas    map[string]StoredChatQuota
}

func newMemoryRepo() *memoryRepo {
	return &memoryRepo{
		follows:   make(map[string]StoredFollow),
		templates: make(map[string]AssistTemplate),
		messages:  make(map[string]ChatMessage),
		quotas:    make(map[string]StoredChatQuota),
	}
}

func (m *memoryRepo) LoadFollows() ([]StoredFollow, error) {
	items := make([]StoredFollow, 0, len(m.follows))
	for _, item := range m.follows {
		items = append(items, item)
	}
	return items, nil
}

func (m *memoryRepo) SaveFollow(item StoredFollow) error {
	m.follows[item.ActorID+"->"+item.TargetID] = item
	return nil
}

func (m *memoryRepo) DeleteFollow(actorID, targetID string) error {
	delete(m.follows, actorID+"->"+targetID)
	return nil
}

func (m *memoryRepo) LoadAssistTemplates() ([]AssistTemplate, error) {
	items := make([]AssistTemplate, 0, len(m.templates))
	for _, item := range m.templates {
		items = append(items, item)
	}
	return items, nil
}

func (m *memoryRepo) SaveAssistTemplate(item AssistTemplate) error {
	m.templates[item.BotID] = item
	return nil
}

func (m *memoryRepo) LoadChatMessages() ([]ChatMessage, error) {
	items := make([]ChatMessage, 0, len(m.messages))
	for _, item := range m.messages {
		items = append(items, item)
	}
	return items, nil
}

func (m *memoryRepo) LoadChatQuotas() ([]StoredChatQuota, error) {
	items := make([]StoredChatQuota, 0, len(m.quotas))
	for _, item := range m.quotas {
		items = append(items, item)
	}
	return items, nil
}

func (m *memoryRepo) SaveChatMessageAndQuota(message ChatMessage, quota StoredChatQuota) error {
	m.messages[message.MessageID] = message
	m.quotas[quota.BotID] = quota
	return nil
}

func TestServiceReloadsPersistedSocialState(t *testing.T) {
	repo := newMemoryRepo()
	service, err := NewServiceWithRepository(repo)
	if err != nil {
		t.Fatalf("NewServiceWithRepository failed: %v", err)
	}

	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatalf("LoadLocation failed: %v", err)
	}
	baseTime := time.Date(2026, 4, 10, 10, 0, 0, 0, loc)
	service.clock = func() time.Time { return baseTime }

	if err := service.Follow("bot_alpha", "bot_beta"); err != nil {
		t.Fatalf("Follow failed: %v", err)
	}
	if err := service.Follow("bot_beta", "bot_alpha"); err != nil {
		t.Fatalf("Follow reverse failed: %v", err)
	}

	template := service.SetAssistTemplate("bot_alpha", "Alpha Assist")
	if !template.IsSubmitted {
		t.Fatal("expected assist template to be submitted")
	}

	worldMessage, err := service.PostChat("bot_alpha", "Alpha", "", "world", "assist_ad", "Alpha ready for world assist.")
	if err != nil {
		t.Fatalf("PostChat world failed: %v", err)
	}

	service.clock = func() time.Time { return baseTime.Add(1 * time.Minute) }
	regionMessage, err := service.PostChat("bot_alpha", "Alpha", "greenfield_village", "region", "friend_recruit", "Alpha recruiting nearby allies.")
	if err != nil {
		t.Fatalf("PostChat region failed: %v", err)
	}

	reloaded, err := NewServiceWithRepository(repo)
	if err != nil {
		t.Fatalf("reloaded NewServiceWithRepository failed: %v", err)
	}
	reloaded.clock = func() time.Time { return baseTime.Add(5 * time.Second) }

	if status := reloaded.RelationStatus("bot_alpha", "bot_beta"); status != "friends" {
		t.Fatalf("expected relation status friends after reload, got %q", status)
	}

	summary := reloaded.SocialSummary("bot_alpha")
	if summary.FollowingCount != 1 || summary.FollowerCount != 1 || summary.FriendCount != 1 {
		t.Fatalf("unexpected social summary after reload: %+v", summary)
	}
	if !summary.HasBorrowableAssistTemplate {
		t.Fatal("expected borrowable assist template after reload")
	}

	if got := reloaded.GetAssistTemplate("bot_alpha"); got.TemplateName != "Alpha Assist" {
		t.Fatalf("expected persisted assist template, got %+v", got)
	}

	worldItems := reloaded.ListChat("world", "", "")
	if len(worldItems) != 1 || worldItems[0].MessageID != worldMessage.MessageID {
		t.Fatalf("expected persisted world message after reload, got %+v", worldItems)
	}

	regionItems := reloaded.ListChat("region", "greenfield_village", "")
	if len(regionItems) != 1 || regionItems[0].MessageID != regionMessage.MessageID {
		t.Fatalf("expected persisted region message after reload, got %+v", regionItems)
	}

	if _, err := reloaded.PostChat("bot_alpha", "Alpha", "", "world", "free_text", "Too soon after reload."); err != ErrChatCooldown {
		t.Fatalf("expected persisted cooldown after reload, got %v", err)
	}

	reloaded.Unfollow("bot_alpha", "bot_beta")
	secondReload, err := NewServiceWithRepository(repo)
	if err != nil {
		t.Fatalf("second reloaded NewServiceWithRepository failed: %v", err)
	}
	if status := secondReload.RelationStatus("bot_alpha", "bot_beta"); status != "followed_by" {
		t.Fatalf("expected unfollow to persist, got %q", status)
	}
}
