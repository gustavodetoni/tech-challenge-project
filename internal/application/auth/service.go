package auth

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	repository "github.com/soat-architecture/tech-challenge-project/internal/application/port"
	domainUser "github.com/soat-architecture/tech-challenge-project/internal/domain/user"
	"github.com/soat-architecture/tech-challenge-project/internal/infra/expections"
)

var (
	ErrInvalidCredentials = expections.New(expections.CodeUnauthorized, "invalid credentials")
	ErrEmailInUse         = expections.New(expections.CodeConflict, "email already in use")
	ErrInvalidInput       = expections.New(expections.CodeValidation, "invalid input")
)

type AuthUseCase struct {
	users repository.UserRepository
	jwt   repository.TokenIssuer
}

func NewAuthUseCase(users repository.UserRepository, jwt repository.TokenIssuer) *AuthUseCase {
	return &AuthUseCase{users: users, jwt: jwt}
}

type RegisterInput struct {
	Name     string
	Email    string
	Password string
}

type LoginInput struct {
	Email    string
	Password string
}

type TokenOutput struct {
	AccessToken string
	ExpiresAt   time.Time
	UserID      string
	Role        domainUser.Role
}

func (s *AuthUseCase) Register(ctx context.Context, in RegisterInput) (*TokenOutput, error) {
	name := strings.TrimSpace(in.Name)
	email := normalizeEmail(in.Email)
	if name == "" || email == "" || !looksLikeEmail(email) {
		return nil, fmt.Errorf("%w: invalid name/email", ErrInvalidInput)
	}
	if len(in.Password) < 8 {
		return nil, fmt.Errorf("%w: password must be at least 8 chars", ErrInvalidInput)
	}

	pwHash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	u := &domainUser.User{
		ID:           uuid.NewString(),
		Name:         name,
		Email:        email,
		PasswordHash: string(pwHash),
		Role:         domainUser.RoleViewer,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.users.Create(ctx, u); err != nil {
		if errors.Is(err, repository.ErrConflict) {
			return nil, ErrEmailInUse
		}
		return nil, err
	}

	return s.issueToken(u.ID, u.Role)
}

func (s *AuthUseCase) Login(ctx context.Context, in LoginInput) (*TokenOutput, error) {
	email := normalizeEmail(in.Email)
	if email == "" || !looksLikeEmail(email) {
		return nil, fmt.Errorf("%w: invalid email", ErrInvalidInput)
	}

	u, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(in.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.issueToken(u.ID, u.Role)
}

func (s *AuthUseCase) issueToken(userID string, role domainUser.Role) (*TokenOutput, error) {
	token, expiresAt, err := s.jwt.NewToken(userID, string(role))
	if err != nil {
		return nil, err
	}
	return &TokenOutput{
		AccessToken: token,
		ExpiresAt:   expiresAt,
		UserID:      userID,
		Role:        role,
	}, nil
}

var emailRe = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

func looksLikeEmail(v string) bool { return emailRe.MatchString(v) }

func normalizeEmail(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return ""
	}
	return strings.ToLower(v)
}
