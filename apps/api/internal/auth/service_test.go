package auth

import (
	"errors"
	"regexp"
	"strconv"
	"testing"
	"time"
)

type memoryRepo struct {
	accounts   map[string]StoredAccount
	sessions   map[string]StoredSession
	challenges map[string]StoredChallenge
}

func newMemoryRepo() *memoryRepo {
	return &memoryRepo{
		accounts:   make(map[string]StoredAccount),
		sessions:   make(map[string]StoredSession),
		challenges: make(map[string]StoredChallenge),
	}
}

func (r *memoryRepo) LoadAccounts() ([]StoredAccount, error) {
	items := make([]StoredAccount, 0, len(r.accounts))
	for _, item := range r.accounts {
		items = append(items, item)
	}
	return items, nil
}

func (r *memoryRepo) SaveAccount(item StoredAccount) error {
	r.accounts[item.Account.AccountID] = item
	return nil
}

func (r *memoryRepo) LoadSessions(now time.Time) ([]StoredSession, error) {
	items := make([]StoredSession, 0, len(r.sessions))
	for _, item := range r.sessions {
		if now.After(item.RefreshTokenExpiresAt) {
			continue
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *memoryRepo) SaveSession(item StoredSession) error {
	r.sessions[item.SessionID] = item
	return nil
}

func (r *memoryRepo) DeleteSession(sessionID string) error {
	delete(r.sessions, sessionID)
	return nil
}

func (r *memoryRepo) LoadChallenges(now time.Time) ([]StoredChallenge, error) {
	items := make([]StoredChallenge, 0, len(r.challenges))
	for _, item := range r.challenges {
		if now.After(item.ExpiresAt) {
			continue
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *memoryRepo) SaveChallenge(item StoredChallenge) error {
	r.challenges[item.ChallengeID] = item
	return nil
}

func (r *memoryRepo) MarkChallengeUsed(challengeID string, usedAt time.Time) error {
	item, ok := r.challenges[challengeID]
	if !ok {
		return nil
	}
	item.UsedAt = &usedAt
	r.challenges[challengeID] = item
	return nil
}

func TestServiceLoadsPersistedSessions(t *testing.T) {
	repo := newMemoryRepo()
	service, err := NewServiceWithRepository(repo)
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	registerChallenge, err := service.IssueChallenge()
	if err != nil {
		t.Fatalf("failed to issue register challenge: %v", err)
	}

	account, err := service.RegisterAccount(
		"bot-session",
		"verysecure",
		registerChallenge.ChallengeID,
		solveChallengePrompt(t, registerChallenge.PromptText),
	)
	if err != nil {
		t.Fatalf("failed to register account: %v", err)
	}

	loginChallenge, err := service.IssueChallenge()
	if err != nil {
		t.Fatalf("failed to issue login challenge: %v", err)
	}

	tokens, err := service.Login(
		"bot-session",
		"verysecure",
		loginChallenge.ChallengeID,
		solveChallengePrompt(t, loginChallenge.PromptText),
	)
	if err != nil {
		t.Fatalf("failed to login: %v", err)
	}

	reloaded, err := NewServiceWithRepository(repo)
	if err != nil {
		t.Fatalf("failed to reload service: %v", err)
	}

	authenticated, err := reloaded.Authenticate(tokens.AccessToken)
	if err != nil {
		t.Fatalf("expected persisted session to authenticate, got error: %v", err)
	}

	if authenticated.AccountID != account.AccountID {
		t.Fatalf("expected account %q, got %q", account.AccountID, authenticated.AccountID)
	}
}

func TestChallengeSingleUseAndValidation(t *testing.T) {
	repo := newMemoryRepo()
	service, err := NewServiceWithRepository(repo)
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	challenge, err := service.IssueChallenge()
	if err != nil {
		t.Fatalf("failed to issue challenge: %v", err)
	}

	if _, err := service.RegisterAccount("bot-invalid", "verysecure", challenge.ChallengeID, "999999"); !errors.Is(err, ErrChallengeInvalid) {
		t.Fatalf("expected ErrChallengeInvalid, got %v", err)
	}

	challenge, err = service.IssueChallenge()
	if err != nil {
		t.Fatalf("failed to issue second challenge: %v", err)
	}

	answer := solveChallengePrompt(t, challenge.PromptText)
	if _, err := service.RegisterAccount("bot-valid", "verysecure", challenge.ChallengeID, answer); err != nil {
		t.Fatalf("expected register to succeed, got %v", err)
	}

	if _, err := service.Login("bot-valid", "verysecure", challenge.ChallengeID, answer); !errors.Is(err, ErrChallengeUsed) {
		t.Fatalf("expected ErrChallengeUsed on reused challenge, got %v", err)
	}
}

func TestRegisterStoresHashedPassword(t *testing.T) {
	repo := newMemoryRepo()
	service, err := NewServiceWithRepository(repo)
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	challenge, err := service.IssueChallenge()
	if err != nil {
		t.Fatalf("failed to issue challenge: %v", err)
	}

	account, err := service.RegisterAccount(
		"bot-hash",
		"verysecure",
		challenge.ChallengeID,
		solveChallengePrompt(t, challenge.PromptText),
	)
	if err != nil {
		t.Fatalf("failed to register account: %v", err)
	}

	stored, ok := repo.accounts[account.AccountID]
	if !ok {
		t.Fatalf("expected stored account for %q", account.AccountID)
	}

	if stored.Password == "verysecure" {
		t.Fatal("expected stored password to be hashed, got plaintext")
	}

	if valid, needsUpgrade := verifyPassword(stored.Password, "verysecure"); !valid || needsUpgrade {
		t.Fatalf("expected stored hash to validate without upgrade, valid=%v needsUpgrade=%v", valid, needsUpgrade)
	}
}

func TestLoginUpgradesLegacyPlaintextPassword(t *testing.T) {
	repo := newMemoryRepo()
	account := Account{
		AccountID: "acct_legacy",
		BotName:   "bot-legacy",
		CreatedAt: time.Now().Format(time.RFC3339),
	}
	repo.accounts[account.AccountID] = StoredAccount{
		Account:  account,
		Password: "verysecure",
	}

	service, err := NewServiceWithRepository(repo)
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	challenge, err := service.IssueChallenge()
	if err != nil {
		t.Fatalf("failed to issue challenge: %v", err)
	}

	if _, err := service.Login(
		"bot-legacy",
		"verysecure",
		challenge.ChallengeID,
		solveChallengePrompt(t, challenge.PromptText),
	); err != nil {
		t.Fatalf("expected legacy plaintext password to still login, got %v", err)
	}

	stored := repo.accounts[account.AccountID]
	if stored.Password == "verysecure" {
		t.Fatal("expected legacy plaintext password to be upgraded to a hash")
	}

	if valid, needsUpgrade := verifyPassword(stored.Password, "verysecure"); !valid || needsUpgrade {
		t.Fatalf("expected upgraded password hash to validate without upgrade, valid=%v needsUpgrade=%v", valid, needsUpgrade)
	}
}

func TestExpiredAccessTokenCanStillRefresh(t *testing.T) {
	repo := newMemoryRepo()
	service, err := NewServiceWithRepository(repo)
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	baseTime := time.Date(2026, time.March, 30, 10, 0, 0, 0, time.FixedZone("CST", 8*60*60))
	service.clock = func() time.Time { return baseTime }

	registerChallenge, err := service.IssueChallenge()
	if err != nil {
		t.Fatalf("failed to issue register challenge: %v", err)
	}

	if _, err := service.RegisterAccount(
		"bot-refresh",
		"verysecure",
		registerChallenge.ChallengeID,
		solveChallengePrompt(t, registerChallenge.PromptText),
	); err != nil {
		t.Fatalf("failed to register account: %v", err)
	}

	loginChallenge, err := service.IssueChallenge()
	if err != nil {
		t.Fatalf("failed to issue login challenge: %v", err)
	}

	tokens, err := service.Login(
		"bot-refresh",
		"verysecure",
		loginChallenge.ChallengeID,
		solveChallengePrompt(t, loginChallenge.PromptText),
	)
	if err != nil {
		t.Fatalf("failed to login: %v", err)
	}

	service.clock = func() time.Time { return baseTime.Add(25 * time.Hour) }

	if _, err := service.Authenticate(tokens.AccessToken); !errors.Is(err, ErrAccessTokenExpired) {
		t.Fatalf("expected ErrAccessTokenExpired, got %v", err)
	}

	refreshed, err := service.RefreshSession(tokens.RefreshToken)
	if err != nil {
		t.Fatalf("expected refresh token to remain valid after access token expiry, got %v", err)
	}

	if refreshed.AccessToken == "" || refreshed.RefreshToken == "" {
		t.Fatal("expected refreshed token pair to be returned")
	}

	if refreshed.AccessToken == tokens.AccessToken {
		t.Fatal("expected refreshed access token to be rotated")
	}
}

func TestReloadedServiceCanRefreshExpiredAccessSession(t *testing.T) {
	repo := newMemoryRepo()
	service, err := NewServiceWithRepository(repo)
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	baseTime := time.Now().Add(-25 * time.Hour)
	service.clock = func() time.Time { return baseTime }

	registerChallenge, err := service.IssueChallenge()
	if err != nil {
		t.Fatalf("failed to issue register challenge: %v", err)
	}

	if _, err := service.RegisterAccount(
		"bot-reload",
		"verysecure",
		registerChallenge.ChallengeID,
		solveChallengePrompt(t, registerChallenge.PromptText),
	); err != nil {
		t.Fatalf("failed to register account: %v", err)
	}

	loginChallenge, err := service.IssueChallenge()
	if err != nil {
		t.Fatalf("failed to issue login challenge: %v", err)
	}

	tokens, err := service.Login(
		"bot-reload",
		"verysecure",
		loginChallenge.ChallengeID,
		solveChallengePrompt(t, loginChallenge.PromptText),
	)
	if err != nil {
		t.Fatalf("failed to login: %v", err)
	}

	reloaded, err := NewServiceWithRepository(repo)
	if err != nil {
		t.Fatalf("failed to reload service: %v", err)
	}

	if _, err := reloaded.Authenticate(tokens.AccessToken); !errors.Is(err, ErrAccessTokenExpired) {
		t.Fatalf("expected ErrAccessTokenExpired after reload, got %v", err)
	}

	refreshed, err := reloaded.RefreshSession(tokens.RefreshToken)
	if err != nil {
		t.Fatalf("expected refresh after reload to succeed, got %v", err)
	}

	if refreshed.AccessToken == "" || refreshed.RefreshToken == "" {
		t.Fatal("expected refreshed tokens after reload")
	}
}

func solveChallengePrompt(t *testing.T, prompt string) string {
	t.Helper()

	matcher := regexp.MustCompile(`ember=(\d+).+frost=(\d+).+moss=(\d+).+factor=(\d+)`)
	matches := matcher.FindStringSubmatch(prompt)
	if len(matches) != 5 {
		t.Fatalf("unexpected challenge prompt format: %q", prompt)
	}

	ember, _ := strconv.Atoi(matches[1])
	frost, _ := strconv.Atoi(matches[2])
	moss, _ := strconv.Atoi(matches[3])
	factor, _ := strconv.Atoi(matches[4])

	return strconv.Itoa(((ember + frost) - moss) * factor)
}
