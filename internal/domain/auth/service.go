package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repository repositoryPort
	mailer     Mailer
	secret     []byte
	now        func() time.Time
}

type repositoryPort interface {
	FindUserByUsername(context.Context, string) (User, error)
	FindUserByUUID(context.Context, string) (User, error)
	FindUserByEmail(context.Context, string) (User, error)
	Register(context.Context, RegisterRequest, string, string, string) error
	CreateActionToken(context.Context, string, string, string, time.Time) error
	ConsumeActionToken(context.Context, string, string, string) (User, error)
	CreateSecureSession(context.Context, string, string, string, string, string, time.Time) error
	FindRefreshSession(context.Context, string) (string, string, error)
	RotateRefreshSession(context.Context, string, string, string, string) error
	UpdatePassword(context.Context, string, string) error
}

type Mailer interface {
	SendRegistration(fullName, recipient, uid string) error
	SendPasswordReset(fullName, recipient, uid string) error
}

func NewService(repository repositoryPort, mailer Mailer, secret string) *Service {
	return &Service{repository: repository, mailer: mailer, secret: []byte(secret), now: time.Now}
}

func (s *Service) Register(ctx context.Context, request RegisterRequest) (RegisterResponse, error) {
	if request.FirstName == "" || request.LastName == "" || request.PositionCode == "" || request.Email == "" || request.Password == "" {
		return RegisterResponse{}, &BusinessError{Code: "MA00004", Detail: "required field"}
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(request.Password), 8)
	if err != nil {
		return RegisterResponse{}, err
	}
	mobileUUID := uuid.NewString()
	displayName := strings.ToUpper(firstRune(request.FirstName) + firstRune(request.LastName))
	if err := s.repository.Register(ctx, request, mobileUUID, string(hash), displayName); err != nil {
		return RegisterResponse{}, err
	}
	token, err := randomToken()
	if err != nil {
		return RegisterResponse{}, err
	}
	if err := s.repository.CreateActionToken(ctx, mobileUUID, "activation", tokenHash(token), s.now().Add(24*time.Hour)); err != nil {
		return RegisterResponse{}, err
	}
	if s.mailer != nil {
		if err := s.mailer.SendRegistration(request.FirstName+" "+request.LastName, request.Email, token); err != nil {
			return RegisterResponse{}, err
		}
	}
	return RegisterResponse{Username: request.Email, Email: request.Email}, nil
}

func (s *Service) RequestPasswordReset(ctx context.Context, request ResetPasswordRequest) (RegisterResponse, error) {
	if request.Email == "" {
		return RegisterResponse{}, &BusinessError{Code: "MA00004", Detail: "email"}
	}
	user, err := s.repository.FindUserByEmail(ctx, request.Email)
	if errors.Is(err, ErrNotFound) {
		return RegisterResponse{}, &BusinessError{Code: "MA00016", Detail: "user"}
	}
	if err != nil {
		return RegisterResponse{}, err
	}
	token, err := randomToken()
	if err != nil {
		return RegisterResponse{}, err
	}
	if err := s.repository.CreateActionToken(ctx, user.MobileUUID, "reset", tokenHash(token), s.now().Add(30*time.Minute)); err != nil {
		return RegisterResponse{}, err
	}
	if s.mailer != nil {
		if err := s.mailer.SendPasswordReset(user.FirstName+" "+user.LastName, request.Email, token); err != nil {
			return RegisterResponse{}, err
		}
	}
	return RegisterResponse{Username: user.Username, Email: user.Email}, nil
}

func (s *Service) CompletePasswordReset(ctx context.Context, request ForgotPasswordRequest) (RegisterResponse, error) {
	if request.Password == "" || request.ConfirmPassword == "" || request.UID == "" {
		return RegisterResponse{}, &BusinessError{Code: "MA00004", Detail: "required field"}
	}
	if request.Password != request.ConfirmPassword {
		return RegisterResponse{}, &BusinessError{Code: "MA00020", Detail: "password"}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(request.ConfirmPassword), 8)
	if err != nil {
		return RegisterResponse{}, err
	}
	user, err := s.repository.ConsumeActionToken(ctx, tokenHash(request.UID), "reset", string(hash))
	if err != nil {
		return RegisterResponse{}, err
	}
	return RegisterResponse{Username: user.Email, Email: user.Email}, nil
}

func firstRune(value string) string {
	for _, character := range value {
		return string(character)
	}
	return ""
}

func (s *Service) Login(ctx context.Context, request LoginRequest, correlationID string) (LoginResponse, error) {
	if request.Username == "" || request.Password == "" {
		return LoginResponse{}, &BusinessError{Code: "MA00004", Detail: "username and password"}
	}
	user, err := s.repository.FindUserByUsername(ctx, request.Username)
	if errors.Is(err, ErrNotFound) {
		return LoginResponse{}, &BusinessError{Code: "MA00027", Detail: "username"}
	}
	if err != nil {
		return LoginResponse{}, err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password)) != nil {
		return LoginResponse{}, &BusinessError{Code: "MA00003", Detail: "password"}
	}
	if user.Status == "D" {
		return LoginResponse{}, &BusinessError{Code: "MA00023", Detail: "user"}
	}
	if user.Status != "A" {
		return LoginResponse{}, &BusinessError{Code: "MA00002", Detail: "user"}
	}

	refreshToken, err := randomToken()
	if err != nil {
		return LoginResponse{}, err
	}
	sessionID := uuid.NewString()
	token, now, err := s.issueToken(user.Username, sessionID)
	if err != nil {
		return LoginResponse{}, err
	}
	if err := s.repository.CreateSecureSession(ctx, sessionID, user.MobileUUID, token, tokenHash(refreshToken), correlationID, now); err != nil {
		return LoginResponse{}, err
	}
	return LoginResponse{
		Token: token, RefreshToken: refreshToken, Username: user.Username,
		LastLoginDateTime: now.Format("2006-01-02 15:04:05"),
	}, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (RefreshTokenResponse, error) {
	if refreshToken == "" {
		return RefreshTokenResponse{}, &BusinessError{Code: "MA00018", Detail: "refToken"}
	}
	sessionID, userID, err := s.repository.FindRefreshSession(ctx, tokenHash(refreshToken))
	if errors.Is(err, ErrNotFound) {
		return RefreshTokenResponse{}, &BusinessError{Code: "MA00009", Detail: "refToken"}
	}
	if err != nil {
		return RefreshTokenResponse{}, err
	}
	user, err := s.repository.FindUserByUUID(ctx, userID)
	if errors.Is(err, ErrNotFound) {
		return RefreshTokenResponse{}, &BusinessError{Code: "MA00002", Detail: "username"}
	}
	if err != nil {
		return RefreshTokenResponse{}, err
	}
	if user.Status != "A" {
		return RefreshTokenResponse{}, ErrNotFound
	}
	nextRefresh, err := randomToken()
	if err != nil {
		return RefreshTokenResponse{}, err
	}
	token, _, err := s.issueToken(user.Username, sessionID)
	if err != nil {
		return RefreshTokenResponse{}, err
	}
	if err := s.repository.RotateRefreshSession(ctx, sessionID, tokenHash(refreshToken), tokenHash(nextRefresh), token); err != nil {
		return RefreshTokenResponse{}, err
	}
	return RefreshTokenResponse{Token: token, RefreshToken: nextRefresh}, nil
}

func (s *Service) ChangePassword(ctx context.Context, username string, request ChangePasswordRequest) (RegisterResponse, error) {
	if request.OldPassword == "" || request.NewPassword == "" || request.ConfirmPassword == "" {
		return RegisterResponse{}, &BusinessError{Code: "MA00004", Detail: "password"}
	}
	if request.NewPassword != request.ConfirmPassword {
		return RegisterResponse{}, &BusinessError{Code: "MA00020", Detail: "password"}
	}
	user, err := s.repository.FindUserByUsername(ctx, username)
	if errors.Is(err, ErrNotFound) || (err == nil && bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.OldPassword)) != nil) {
		return RegisterResponse{}, &BusinessError{Code: "MA00003", Detail: "password"}
	}
	if err != nil {
		return RegisterResponse{}, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(request.ConfirmPassword), 8)
	if err != nil {
		return RegisterResponse{}, err
	}
	if err := s.repository.UpdatePassword(ctx, user.MobileUUID, string(hash)); err != nil {
		return RegisterResponse{}, err
	}
	return RegisterResponse{Username: user.Email, Email: user.Email}, nil
}

func (s *Service) Activate(ctx context.Context, uid string) (RegisterResponse, error) {
	user, err := s.repository.ConsumeActionToken(ctx, tokenHash(uid), "activation", "")
	if err != nil {
		return RegisterResponse{}, err
	}
	return RegisterResponse{Username: user.Username, Email: user.Email}, nil
}

func (s *Service) issueToken(username, tokenID string) (string, time.Time, error) {
	now := s.now()
	claims := jwt.RegisteredClaims{
		Issuer: "smart-seaman-mobile-api", Subject: username, ID: tokenID,
		IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(30 * time.Minute)),
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign token: %w", err)
	}
	return token, now, nil
}

func randomToken() (string, error) {
	var value [32]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(value[:]), nil
}
func tokenHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
