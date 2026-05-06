package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"fmt"
	"time"
)

const CookieName = "keychain_admin_session"

type Service struct {
	db            *sql.DB
	adminUsername string
	adminPassword string
	sessionSecret []byte
	sessionTTL    time.Duration
	now           func() time.Time
}

type Options struct {
	DB            *sql.DB
	AdminUsername string
	AdminPassword string
	SessionSecret string
	SessionTTL    time.Duration
	Now           func() time.Time
}

type Session struct {
	ID        string
	Username  string
	ExpiresAt time.Time
}

func NewService(options Options) (*Service, error) {
	if options.DB == nil {
		return nil, fmt.Errorf("auth database is required")
	}
	if options.AdminUsername == "" {
		return nil, fmt.Errorf("admin username is required")
	}
	if options.AdminPassword == "" {
		return nil, fmt.Errorf("admin password is required")
	}
	if options.SessionSecret == "" {
		return nil, fmt.Errorf("session secret is required")
	}
	if options.SessionTTL == 0 {
		options.SessionTTL = 24 * time.Hour
	}
	if options.Now == nil {
		options.Now = time.Now
	}
	return &Service{
		db:            options.DB,
		adminUsername: options.AdminUsername,
		adminPassword: options.AdminPassword,
		sessionSecret: []byte(options.SessionSecret),
		sessionTTL:    options.SessionTTL,
		now:           options.Now,
	}, nil
}

func (service *Service) AdminUsername() string {
	return service.adminUsername
}

func (service *Service) Authenticate(username string, password string) bool {
	return constantTimeEqual(username, service.adminUsername) && constantTimeEqual(password, service.adminPassword)
}

func (service *Service) CreateSession(ctx context.Context) (string, Session, error) {
	token, err := randomToken()
	if err != nil {
		return "", Session{}, err
	}

	session := Session{
		ID:        randomID(token),
		Username:  service.adminUsername,
		ExpiresAt: service.now().UTC().Add(service.sessionTTL),
	}
	if _, err := service.db.ExecContext(ctx, `
INSERT INTO admin_sessions (id, session_token_hash, expires_at, created_at)
VALUES (?, ?, ?, ?);
`, session.ID, service.hashToken(token), formatTime(session.ExpiresAt), formatTime(service.now().UTC())); err != nil {
		return "", Session{}, fmt.Errorf("create admin session: %w", err)
	}
	return token, session, nil
}

func (service *Service) GetSession(ctx context.Context, token string) (Session, bool, error) {
	if token == "" {
		return Session{}, false, nil
	}

	var session Session
	var expiresAt string
	if err := service.db.QueryRowContext(ctx, `
SELECT id, expires_at FROM admin_sessions WHERE session_token_hash = ?;
`, service.hashToken(token)).Scan(&session.ID, &expiresAt); err != nil {
		if err == sql.ErrNoRows {
			return Session{}, false, nil
		}
		return Session{}, false, fmt.Errorf("get admin session: %w", err)
	}

	parsedExpiresAt, err := time.Parse(time.RFC3339, expiresAt)
	if err != nil {
		return Session{}, false, fmt.Errorf("parse session expiry: %w", err)
	}
	if !service.now().UTC().Before(parsedExpiresAt) {
		if err := service.DeleteSession(ctx, token); err != nil {
			return Session{}, false, err
		}
		return Session{}, false, nil
	}

	session.Username = service.adminUsername
	session.ExpiresAt = parsedExpiresAt
	return session, true, nil
}

func (service *Service) DeleteSession(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	if _, err := service.db.ExecContext(ctx, `
DELETE FROM admin_sessions WHERE session_token_hash = ?;
`, service.hashToken(token)); err != nil {
		return fmt.Errorf("delete admin session: %w", err)
	}
	return nil
}

func (service *Service) DeleteExpiredSessions(ctx context.Context) error {
	if _, err := service.db.ExecContext(ctx, `
DELETE FROM admin_sessions WHERE expires_at <= ?;
`, formatTime(service.now().UTC())); err != nil {
		return fmt.Errorf("delete expired admin sessions: %w", err)
	}
	return nil
}

func (service *Service) hashToken(token string) string {
	mac := hmac.New(sha256.New, service.sessionSecret)
	mac.Write([]byte(token))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func randomToken() (string, error) {
	var bytes [32]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", fmt.Errorf("generate session token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(bytes[:]), nil
}

func randomID(seed string) string {
	sum := sha256.Sum256([]byte(seed))
	return base64.RawURLEncoding.EncodeToString(sum[:16])
}

func constantTimeEqual(a string, b string) bool {
	if len(a) != len(b) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

func formatTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339)
}
