package auth

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	businessTimezone     = "Asia/Shanghai"
	accessTokenLifetime  = 24 * time.Hour
	refreshTokenLifetime = 7 * 24 * time.Hour
	challengeLifetime    = 60 * time.Second
)

var (
	ErrBotNameTaken         = errors.New("bot name already exists")
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrSessionNotFound      = errors.New("session not found")
	ErrRefreshTokenExpired  = errors.New("refresh token expired")
	ErrAccessTokenExpired   = errors.New("access token expired")
	ErrInvalidRegisterInput = errors.New("invalid register input")
	ErrChallengeRequired    = errors.New("challenge is required")
	ErrChallengeNotFound    = errors.New("challenge not found")
	ErrChallengeExpired     = errors.New("challenge expired")
	ErrChallengeUsed        = errors.New("challenge already used")
	ErrChallengeInvalid     = errors.New("challenge answer invalid")
)

type Account struct {
	AccountID string `json:"account_id"`
	BotName   string `json:"bot_name"`
	CreatedAt string `json:"created_at"`
}

type TokenPair struct {
	AccessToken           string `json:"access_token"`
	AccessTokenExpiresAt  string `json:"access_token_expires_at"`
	RefreshToken          string `json:"refresh_token"`
	RefreshTokenExpiresAt string `json:"refresh_token_expires_at"`
}

type Challenge struct {
	ChallengeID  string `json:"challenge_id"`
	PromptText   string `json:"prompt_text"`
	AnswerFormat string `json:"answer_format"`
	ExpiresAt    string `json:"expires_at"`
}

type StoredAccount struct {
	Account  Account
	Password string
}

type StoredSession struct {
	SessionID             string
	AccountID             string
	AccessToken           string
	AccessTokenExpiresAt  time.Time
	RefreshToken          string
	RefreshTokenExpiresAt time.Time
}

type StoredChallenge struct {
	ChallengeID    string
	PromptText     string
	AnswerFormat   string
	ExpectedAnswer string
	ExpiresAt      time.Time
	UsedAt         *time.Time
}

type Repository interface {
	LoadAccounts() ([]StoredAccount, error)
	SaveAccount(StoredAccount) error
	LoadSessions(now time.Time) ([]StoredSession, error)
	SaveSession(StoredSession) error
	DeleteSession(sessionID string) error
	LoadChallenges(now time.Time) ([]StoredChallenge, error)
	SaveChallenge(StoredChallenge) error
	MarkChallengeUsed(challengeID string, usedAt time.Time) error
}

type accountRecord struct {
	account  Account
	password string
}

type sessionRecord struct {
	SessionID             string
	AccountID             string
	AccessToken           string
	AccessTokenExpiresAt  time.Time
	RefreshToken          string
	RefreshTokenExpiresAt time.Time
}

type challengeRecord struct {
	challenge      Challenge
	expectedAnswer string
	usedAt         *time.Time
}

type Service struct {
	mu               sync.RWMutex
	clock            func() time.Time
	loc              *time.Location
	rng              *rand.Rand
	repo             Repository
	accountsByID     map[string]accountRecord
	accountIDByName  map[string]string
	sessionByAccess  map[string]sessionRecord
	sessionByRefresh map[string]sessionRecord
	challengesByID   map[string]challengeRecord
}

var authIDCounter uint64

func NewService() *Service {
	service, err := NewServiceWithRepository(nil)
	if err != nil {
		panic(err)
	}

	return service
}

func NewServiceWithRepository(repo Repository) (*Service, error) {
	service := &Service{
		clock:            time.Now,
		loc:              mustLocation(businessTimezone),
		rng:              rand.New(rand.NewSource(time.Now().UnixNano())),
		repo:             repo,
		accountsByID:     make(map[string]accountRecord),
		accountIDByName:  make(map[string]string),
		sessionByAccess:  make(map[string]sessionRecord),
		sessionByRefresh: make(map[string]sessionRecord),
		challengesByID:   make(map[string]challengeRecord),
	}

	if repo == nil {
		return service, nil
	}

	accounts, err := repo.LoadAccounts()
	if err != nil {
		return nil, err
	}

	for _, stored := range accounts {
		service.accountsByID[stored.Account.AccountID] = accountRecord{
			account:  stored.Account,
			password: stored.Password,
		}
		service.accountIDByName[strings.ToLower(strings.TrimSpace(stored.Account.BotName))] = stored.Account.AccountID
	}

	sessions, err := repo.LoadSessions(service.clock().In(service.loc))
	if err != nil {
		return nil, err
	}

	for _, stored := range sessions {
		session := sessionRecord{
			SessionID:             stored.SessionID,
			AccountID:             stored.AccountID,
			AccessToken:           stored.AccessToken,
			AccessTokenExpiresAt:  stored.AccessTokenExpiresAt.In(service.loc),
			RefreshToken:          stored.RefreshToken,
			RefreshTokenExpiresAt: stored.RefreshTokenExpiresAt.In(service.loc),
		}
		service.sessionByAccess[session.AccessToken] = session
		service.sessionByRefresh[session.RefreshToken] = session
	}

	challenges, err := repo.LoadChallenges(service.clock().In(service.loc))
	if err != nil {
		return nil, err
	}

	for _, stored := range challenges {
		challenge := Challenge{
			ChallengeID:  stored.ChallengeID,
			PromptText:   stored.PromptText,
			AnswerFormat: stored.AnswerFormat,
			ExpiresAt:    stored.ExpiresAt.In(service.loc).Format(time.RFC3339),
		}
		service.challengesByID[challenge.ChallengeID] = challengeRecord{
			challenge:      challenge,
			expectedAnswer: stored.ExpectedAnswer,
			usedAt:         stored.UsedAt,
		}
	}

	return service, nil
}

func (s *Service) IssueChallenge() (Challenge, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.clock().In(s.loc)
	challenge, expectedAnswer := s.generateChallengeLocked(now)
	record := challengeRecord{
		challenge:      challenge,
		expectedAnswer: expectedAnswer,
	}

	if s.repo != nil {
		if err := s.repo.SaveChallenge(StoredChallenge{
			ChallengeID:    challenge.ChallengeID,
			PromptText:     challenge.PromptText,
			AnswerFormat:   challenge.AnswerFormat,
			ExpectedAnswer: expectedAnswer,
			ExpiresAt:      parseRFC3339(challenge.ExpiresAt, s.loc),
			UsedAt:         nil,
		}); err != nil {
			return Challenge{}, err
		}
	}

	s.challengesByID[challenge.ChallengeID] = record
	return challenge, nil
}

func (s *Service) RegisterAccount(botName, password, challengeID, challengeAnswer string) (Account, error) {
	botName = strings.TrimSpace(botName)
	if len(botName) < 3 || len(botName) > 32 || len(password) < 8 || len(password) > 128 {
		return Account{}, ErrInvalidRegisterInput
	}

	lookupKey := strings.ToLower(botName)

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.consumeChallengeLocked(challengeID, challengeAnswer); err != nil {
		return Account{}, err
	}

	if _, exists := s.accountIDByName[lookupKey]; exists {
		return Account{}, ErrBotNameTaken
	}

	account := Account{
		AccountID: nextID("acct"),
		BotName:   botName,
		CreatedAt: s.clock().In(s.loc).Format(time.RFC3339),
	}

	passwordHash, err := hashPassword(password)
	if err != nil {
		return Account{}, err
	}

	if s.repo != nil {
		if err := s.repo.SaveAccount(StoredAccount{
			Account:  account,
			Password: passwordHash,
		}); err != nil {
			return Account{}, err
		}
	}

	s.accountsByID[account.AccountID] = accountRecord{
		account:  account,
		password: passwordHash,
	}
	s.accountIDByName[lookupKey] = account.AccountID

	return account, nil
}

func (s *Service) Login(botName, password, challengeID, challengeAnswer string) (TokenPair, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.consumeChallengeLocked(challengeID, challengeAnswer); err != nil {
		return TokenPair{}, err
	}

	record, ok := s.lookupAccountByNameLocked(botName)
	if !ok {
		return TokenPair{}, ErrInvalidCredentials
	}

	passwordValid, needsUpgrade := verifyPassword(record.password, password)
	if !passwordValid {
		return TokenPair{}, ErrInvalidCredentials
	}

	if needsUpgrade {
		passwordHash, err := hashPassword(password)
		if err != nil {
			return TokenPair{}, err
		}

		if s.repo != nil {
			if err := s.repo.SaveAccount(StoredAccount{
				Account:  record.account,
				Password: passwordHash,
			}); err != nil {
				return TokenPair{}, err
			}
		}

		record.password = passwordHash
		s.accountsByID[record.account.AccountID] = record
	}

	return s.issueTokenPairLocked(record.account.AccountID)
}

func (s *Service) RefreshSession(refreshToken string) (TokenPair, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, ok := s.sessionByRefresh[refreshToken]
	if !ok {
		return TokenPair{}, ErrSessionNotFound
	}

	now := s.clock().In(s.loc)
	if now.After(session.RefreshTokenExpiresAt) {
		_ = s.deleteSessionLocked(session)
		delete(s.sessionByAccess, session.AccessToken)
		delete(s.sessionByRefresh, session.RefreshToken)
		return TokenPair{}, ErrRefreshTokenExpired
	}

	if err := s.deleteSessionLocked(session); err != nil {
		return TokenPair{}, err
	}
	delete(s.sessionByAccess, session.AccessToken)
	delete(s.sessionByRefresh, session.RefreshToken)

	return s.issueTokenPairLocked(session.AccountID)
}

func (s *Service) Authenticate(accessToken string) (Account, error) {
	s.mu.RLock()
	session, ok := s.sessionByAccess[accessToken]
	if !ok {
		s.mu.RUnlock()
		return Account{}, ErrSessionNotFound
	}

	record, ok := s.accountsByID[session.AccountID]
	s.mu.RUnlock()
	if !ok {
		return Account{}, ErrSessionNotFound
	}

	now := s.clock().In(s.loc)
	if now.After(session.AccessTokenExpiresAt) {
		return Account{}, ErrAccessTokenExpired
	}

	return record.account, nil
}

func (s *Service) GetAccount(accountID string) (Account, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	record, ok := s.accountsByID[accountID]
	if !ok {
		return Account{}, false
	}

	return record.account, true
}

func (s *Service) generateChallengeLocked(now time.Time) (Challenge, string) {
	ember := 12 + s.rng.Intn(18)
	frost := 5 + s.rng.Intn(11)
	moss := 2 + s.rng.Intn(7)
	factor := 2 + s.rng.Intn(3)
	answer := ((ember + frost) - moss) * factor
	expiresAt := now.Add(challengeLifetime)

	return Challenge{
			ChallengeID: nextID("challenge"),
			PromptText: fmt.Sprintf(
				"Cipher slate // ember=%d // frost=%d // moss=%d // factor=%d. Compute ((ember + frost) - moss) * factor. Reply with digits only.",
				ember,
				frost,
				moss,
				factor,
			),
			AnswerFormat: "digits_only",
			ExpiresAt:    expiresAt.Format(time.RFC3339),
		},
		strconv.Itoa(answer)
}

func (s *Service) consumeChallengeLocked(challengeID, challengeAnswer string) error {
	challengeID = strings.TrimSpace(challengeID)
	challengeAnswer = normalizeChallengeAnswer(challengeAnswer)
	if challengeID == "" || challengeAnswer == "" {
		return ErrChallengeRequired
	}

	record, ok := s.challengesByID[challengeID]
	if !ok {
		return ErrChallengeNotFound
	}

	if record.usedAt != nil {
		return ErrChallengeUsed
	}

	now := s.clock().In(s.loc)
	expiresAt := parseRFC3339(record.challenge.ExpiresAt, s.loc)
	if now.After(expiresAt) {
		return ErrChallengeExpired
	}

	if challengeAnswer != normalizeChallengeAnswer(record.expectedAnswer) {
		return ErrChallengeInvalid
	}

	record.usedAt = &now
	s.challengesByID[challengeID] = record
	if err := s.markChallengeUsedLocked(challengeID, now); err != nil {
		return err
	}

	return nil
}

func (s *Service) markChallengeUsedLocked(challengeID string, usedAt time.Time) error {
	if s.repo == nil {
		return nil
	}

	return s.repo.MarkChallengeUsed(challengeID, usedAt)
}

func (s *Service) issueTokenPairLocked(accountID string) (TokenPair, error) {
	now := s.clock().In(s.loc)
	session := sessionRecord{
		SessionID:             nextID("sess"),
		AccountID:             accountID,
		AccessToken:           nextID("access"),
		AccessTokenExpiresAt:  now.Add(accessTokenLifetime),
		RefreshToken:          nextID("refresh"),
		RefreshTokenExpiresAt: now.Add(refreshTokenLifetime),
	}

	if s.repo != nil {
		if err := s.repo.SaveSession(StoredSession{
			SessionID:             session.SessionID,
			AccountID:             session.AccountID,
			AccessToken:           session.AccessToken,
			AccessTokenExpiresAt:  session.AccessTokenExpiresAt,
			RefreshToken:          session.RefreshToken,
			RefreshTokenExpiresAt: session.RefreshTokenExpiresAt,
		}); err != nil {
			return TokenPair{}, err
		}
	}

	s.sessionByAccess[session.AccessToken] = session
	s.sessionByRefresh[session.RefreshToken] = session

	return TokenPair{
		AccessToken:           session.AccessToken,
		AccessTokenExpiresAt:  session.AccessTokenExpiresAt.Format(time.RFC3339),
		RefreshToken:          session.RefreshToken,
		RefreshTokenExpiresAt: session.RefreshTokenExpiresAt.Format(time.RFC3339),
	}, nil
}

func (s *Service) deleteSessionLocked(session sessionRecord) error {
	if s.repo == nil {
		return nil
	}

	return s.repo.DeleteSession(session.SessionID)
}

func (s *Service) lookupAccountByNameLocked(botName string) (accountRecord, bool) {
	accountID, ok := s.accountIDByName[strings.ToLower(strings.TrimSpace(botName))]
	if !ok {
		return accountRecord{}, false
	}

	record, ok := s.accountsByID[accountID]
	if !ok {
		return accountRecord{}, false
	}

	return record, true
}

func mustLocation(name string) *time.Location {
	location, err := time.LoadLocation(name)
	if err != nil {
		panic(err)
	}

	return location
}

func normalizeChallengeAnswer(answer string) string {
	return strings.TrimSpace(strings.ToLower(answer))
}

func hashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashed), nil
}

func verifyPassword(storedPassword, providedPassword string) (valid bool, needsUpgrade bool) {
	if isPasswordHash(storedPassword) {
		return bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(providedPassword)) == nil, false
	}

	return subtle.ConstantTimeCompare([]byte(storedPassword), []byte(providedPassword)) == 1, true
}

func isPasswordHash(storedPassword string) bool {
	_, err := bcrypt.Cost([]byte(storedPassword))
	return err == nil
}

func parseRFC3339(raw string, loc *time.Location) time.Time {
	if raw == "" {
		return time.Now().In(loc)
	}

	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Now().In(loc)
	}

	return parsed.In(loc)
}

func nextID(prefix string) string {
	return fmt.Sprintf("%s_%d_%06d", prefix, time.Now().UnixNano(), atomic.AddUint64(&authIDCounter, 1))
}
