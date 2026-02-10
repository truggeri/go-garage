package services

import (
	"context"

	"github.com/truggeri/go-garage/internal/auth"
	"github.com/truggeri/go-garage/internal/models"
	"github.com/truggeri/go-garage/internal/repositories"
	"golang.org/x/crypto/bcrypt"
)

// AuthenticationService defines the interface for authentication operations
type AuthenticationService interface {
	// Register creates a new user account and returns authentication tokens
	Register(ctx context.Context, registration RegistrationRequest) (*AuthenticationResult, error)

	// Authenticate validates credentials and returns authentication tokens
	Authenticate(ctx context.Context, identifier, password string) (*AuthenticationResult, error)

	// RefreshSession generates new tokens using a valid refresh token
	RefreshSession(ctx context.Context, refreshToken string) (*AuthenticationResult, error)
}

// RegistrationRequest contains the data needed to register a new user
type RegistrationRequest struct {
	Username  string
	Email     string
	Password  string
	FirstName string
	LastName  string
}

// AuthenticationResult contains the tokens returned after successful authentication
type AuthenticationResult struct {
	AccessToken      string
	RefreshToken     string
	AccessExpiresAt  int64
	RefreshExpiresAt int64
	AccountID        string
	AccountName      string
}

// GarageAuthService implements AuthenticationService
type GarageAuthService struct {
	userRepo     repositories.UserRepository
	tokenManager *auth.TokenManager
}

// BuildAuthService creates a new GarageAuthService
func BuildAuthService(userRepo repositories.UserRepository, tokenMgr *auth.TokenManager) *GarageAuthService {
	return &GarageAuthService{
		userRepo:     userRepo,
		tokenManager: tokenMgr,
	}
}

// Register creates a new user account and returns authentication tokens
func (s *GarageAuthService) Register(ctx context.Context, registration RegistrationRequest) (*AuthenticationResult, error) {
	if err := models.ValidatePassword(registration.Password); err != nil {
		return nil, err
	}

	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(registration.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, models.NewDatabaseError("hash password", err)
	}

	newUser := &models.User{
		Username:     registration.Username,
		Email:        registration.Email,
		PasswordHash: string(hashedPwd),
		FirstName:    registration.FirstName,
		LastName:     registration.LastName,
	}

	if createErr := s.userRepo.Create(ctx, newUser); createErr != nil {
		return nil, createErr
	}

	bundle, err := s.tokenManager.GenerateTokenBundle(auth.TokenPayload{
		AccountID:   newUser.ID,
		AccountName: newUser.Username,
	})
	if err != nil {
		return nil, err
	}

	// UpdateLastLogin failure is non-critical - the user is already registered
	// and has valid tokens. We intentionally ignore any error here.
	_ = s.userRepo.UpdateLastLogin(ctx, newUser.ID)

	return &AuthenticationResult{
		AccessToken:      bundle.AccessToken,
		RefreshToken:     bundle.RefreshToken,
		AccessExpiresAt:  bundle.AccessExpiresAt.Unix(),
		RefreshExpiresAt: bundle.RefreshExpiresAt.Unix(),
		AccountID:        newUser.ID,
		AccountName:      newUser.Username,
	}, nil
}

// Authenticate validates credentials and returns authentication tokens
func (s *GarageAuthService) Authenticate(ctx context.Context, identifier, password string) (*AuthenticationResult, error) {
	// Try to find user by email first, then by username
	user, err := s.userRepo.FindByEmail(ctx, identifier)
	if err != nil {
		var notFoundErr *models.NotFoundError
		if isNotFound := models.IsNotFoundError(err, &notFoundErr); isNotFound {
			// Try username
			user, err = s.userRepo.FindByUsername(ctx, identifier)
			if err != nil {
				return nil, models.NewValidationError("credentials", "invalid email/username or password")
			}
		} else {
			return nil, err
		}
	}

	if pwdErr := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); pwdErr != nil {
		return nil, models.NewValidationError("credentials", "invalid email/username or password")
	}

	bundle, err := s.tokenManager.GenerateTokenBundle(auth.TokenPayload{
		AccountID:   user.ID,
		AccountName: user.Username,
	})
	if err != nil {
		return nil, err
	}

	// UpdateLastLogin failure is non-critical - the user is already authenticated
	// and has valid tokens. We intentionally ignore any error here.
	_ = s.userRepo.UpdateLastLogin(ctx, user.ID)

	return &AuthenticationResult{
		AccessToken:      bundle.AccessToken,
		RefreshToken:     bundle.RefreshToken,
		AccessExpiresAt:  bundle.AccessExpiresAt.Unix(),
		RefreshExpiresAt: bundle.RefreshExpiresAt.Unix(),
		AccountID:        user.ID,
		AccountName:      user.Username,
	}, nil
}

// RefreshSession generates new tokens using a valid refresh token
func (s *GarageAuthService) RefreshSession(ctx context.Context, refreshToken string) (*AuthenticationResult, error) {
	bundle, err := s.tokenManager.RefreshAccessToken(refreshToken)
	if err != nil {
		return nil, models.NewValidationError("refresh_token", "invalid or expired refresh token")
	}

	verified, err := s.tokenManager.ValidateToken(bundle.AccessToken)
	if err != nil {
		return nil, err
	}

	return &AuthenticationResult{
		AccessToken:      bundle.AccessToken,
		RefreshToken:     bundle.RefreshToken,
		AccessExpiresAt:  bundle.AccessExpiresAt.Unix(),
		RefreshExpiresAt: bundle.RefreshExpiresAt.Unix(),
		AccountID:        verified.AccountID,
		AccountName:      verified.AccountName,
	}, nil
}
