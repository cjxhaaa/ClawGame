package auth

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	businessTimezone     = "Asia/Shanghai"
	accessTokenLifetime  = 15 * time.Minute
	refreshTokenLifetime = 7 * 24 * time.Hour
)

var (
	ErrBotNameTaken         = errors.New("bot name already exists")
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrSessionNotFound      = errors.New("session not found")
	ErrRefreshTokenExpired  = errors.New("refresh token expired")
	ErrAccessTokenExpired   = errors.New("access token expired")
	ErrInvalidRegisterInput = errors.New("invalid register input")
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

type Service struct {
	mu               sync.RWMutex
	clock            func() time.Time
	loc              *time.Location
	accountsByID     map[string]accountRecord
	accountIDByName  map[string]string
	sessionByAccess  map[string]sessionRecord
	sessionByRefresh map[string]sessionRecord
}

var authIDCounter uint64

func NewService() *Service {
	return &Service{
		clock:            time.Now,
		loc:              mustLocation(businessTimezone),
		accountsByID:     make(map[string]accountRecord),
		accountIDByName:  make(map[string]string),
		sessionByAccess:  make(map[string]sessionRecord),
		sessionByRefresh: make(map[string]sessionRecord),
	}
}

func (s *Service) RegisterAccount(botName, password string) (Account, error) {
	botName = strings.TrimSpace(botName)
	if len(botName) < 3 || len(botName) > 32 || len(password) < 8 || len(password) > 128 {
		return Account{}, ErrInvalidRegisterInput
	}

	lookupKey := strings.ToLower(botName)

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.accountIDByName[lookupKey]; exists {
		return Account{}, ErrBotNameTaken
	}

	account := Account{
		AccountID: nextID("acct"),
		BotName:   botName,
		CreatedAt: s.clock().In(s.loc).Format(time.RFC3339),
	}

	s.accountsByID[account.AccountID] = accountRecord{
		account:  account,
		password: password,
	}
	s.accountIDByName[lookupKey] = account.AccountID

	return account, nil
}

func (s *Service) Login(botName, password string) (TokenPair, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.lookupAccountByNameLocked(botName)
	if !ok || record.password != password {
		return TokenPair{}, ErrInvalidCredentials
	}

	return s.issueTokenPairLocked(record.account.AccountID), nil
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
		delete(s.sessionByAccess, session.AccessToken)
		delete(s.sessionByRefresh, session.RefreshToken)
		return TokenPair{}, ErrRefreshTokenExpired
	}

	delete(s.sessionByAccess, session.AccessToken)
	delete(s.sessionByRefresh, session.RefreshToken)

	return s.issueTokenPairLocked(session.AccountID), nil
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
		s.mu.Lock()
		delete(s.sessionByAccess, accessToken)
		delete(s.sessionByRefresh, session.RefreshToken)
		s.mu.Unlock()
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

func (s *Service) issueTokenPairLocked(accountID string) TokenPair {
	now := s.clock().In(s.loc)
	session := sessionRecord{
		SessionID:             nextID("sess"),
		AccountID:             accountID,
		AccessToken:           nextID("access"),
		AccessTokenExpiresAt:  now.Add(accessTokenLifetime),
		RefreshToken:          nextID("refresh"),
		RefreshTokenExpiresAt: now.Add(refreshTokenLifetime),
	}

	s.sessionByAccess[session.AccessToken] = session
	s.sessionByRefresh[session.RefreshToken] = session

	return TokenPair{
		AccessToken:           session.AccessToken,
		AccessTokenExpiresAt:  session.AccessTokenExpiresAt.Format(time.RFC3339),
		RefreshToken:          session.RefreshToken,
		RefreshTokenExpiresAt: session.RefreshTokenExpiresAt.Format(time.RFC3339),
	}
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

func nextID(prefix string) string {
	return fmt.Sprintf("%s_%06d", prefix, atomic.AddUint64(&authIDCounter, 1))
}
