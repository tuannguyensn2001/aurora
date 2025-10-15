package service

import (
	"api/config"
	"api/internal/dto"
	"api/internal/middleware"
	"api/internal/model"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
)

// GoogleUserInfo represents the user info returned by Google
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

// GetGoogleOAuthConfig returns the OAuth2 config for Google
func (s *service) GetGoogleOAuthConfig(cfg *config.Config) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     cfg.OAuth.Google.ClientID,
		ClientSecret: cfg.OAuth.Google.ClientSecret,
		RedirectURL:  cfg.OAuth.Google.RedirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
}

// GenerateStateToken generates a random state token for OAuth
func (s *service) GenerateStateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GetGoogleLoginURL generates the Google OAuth login URL
func (s *service) GetGoogleLoginURL(ctx context.Context, cfg *config.Config) (string, string, error) {
	logger := log.Ctx(ctx).With().Str("service", "auth").Str("method", "GetGoogleLoginURL").Logger()

	oauthConfig := s.GetGoogleOAuthConfig(cfg)
	state, err := s.GenerateStateToken()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to generate state token")
		return "", "", err
	}

	url := oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	return url, state, nil
}

// HandleGoogleCallback handles the OAuth callback from Google
func (s *service) HandleGoogleCallback(ctx context.Context, cfg *config.Config, code string, state string) (*dto.AuthResponse, error) {
	logger := log.Ctx(ctx).With().Str("service", "auth").Str("method", "HandleGoogleCallback").Logger()

	oauthConfig := s.GetGoogleOAuthConfig(cfg)

	// Exchange code for token
	token, err := oauthConfig.Exchange(ctx, code)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to exchange code for token")
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	// Get user info from Google
	userInfo, err := s.GetGoogleUserInfo(ctx, token.AccessToken)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get user info")
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// Check if user exists, create if not
	user, err := s.repo.GetUserByGoogleID(ctx, userInfo.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create new user
			now := time.Now()
			user = &model.User{
				Email:        userInfo.Email,
				Name:         userInfo.Name,
				Picture:      userInfo.Picture,
				GoogleID:     userInfo.ID,
				AccessToken:  token.AccessToken,
				RefreshToken: token.RefreshToken,
				TokenExpiry:  &token.Expiry,
				LastLoginAt:  &now,
			}

			if err := s.repo.CreateUser(ctx, user); err != nil {
				logger.Error().Err(err).Msg("Failed to create user")
				return nil, fmt.Errorf("failed to create user: %w", err)
			}
			logger.Info().Str("email", user.Email).Msg("Created new user")
		} else {
			logger.Error().Err(err).Msg("Failed to get user by Google ID")
			return nil, fmt.Errorf("failed to get user: %w", err)
		}
	} else {
		// Update existing user
		user.AccessToken = token.AccessToken
		user.RefreshToken = token.RefreshToken
		user.TokenExpiry = &token.Expiry
		user.Name = userInfo.Name
		user.Picture = userInfo.Picture
		now := time.Now()
		user.LastLoginAt = &now

		if err := s.repo.UpdateUser(ctx, user); err != nil {
			logger.Error().Err(err).Msg("Failed to update user")
			return nil, fmt.Errorf("failed to update user: %w", err)
		}
		logger.Info().Str("email", user.Email).Msg("Updated existing user")
	}

	// Generate JWT token
	jwtToken, err := middleware.GenerateJWT(cfg, user.ID, user.Email, user.Name)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to generate JWT token")
		return nil, fmt.Errorf("failed to generate JWT token: %w", err)
	}

	jwtExpiresAt := time.Now().Add(time.Duration(cfg.JWT.ExpireHour) * time.Hour)

	// Create response
	response := &dto.AuthResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresAt:    token.Expiry,
		JWTToken:     jwtToken,
		JWTExpiresAt: jwtExpiresAt,
		User: dto.UserInfo{
			ID:          user.ID,
			Email:       user.Email,
			Name:        user.Name,
			Picture:     user.Picture,
			LastLoginAt: *user.LastLoginAt,
		},
	}

	return response, nil
}

// GetGoogleUserInfo fetches user info from Google
func (s *service) GetGoogleUserInfo(ctx context.Context, accessToken string) (*GoogleUserInfo, error) {
	logger := log.Ctx(ctx).With().Str("service", "auth").Str("method", "GetGoogleUserInfo").Logger()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get user info from Google")
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.Error().Int("status", resp.StatusCode).Str("body", string(body)).Msg("Google API returned non-200 status")
		return nil, fmt.Errorf("google API returned status %d", resp.StatusCode)
	}

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		logger.Error().Err(err).Msg("Failed to decode user info")
		return nil, err
	}

	return &userInfo, nil
}

// GetUserByID retrieves a user by ID
func (s *service) GetUserByID(ctx context.Context, id uint) (*model.User, error) {
	return s.repo.GetUserByID(ctx, id)
}

// RefreshToken refreshes an OAuth token
func (s *service) RefreshToken(ctx context.Context, cfg *config.Config, refreshToken string) (*dto.AuthResponse, error) {
	logger := log.Ctx(ctx).With().Str("service", "auth").Str("method", "RefreshToken").Logger()

	oauthConfig := s.GetGoogleOAuthConfig(cfg)

	token := &oauth2.Token{
		RefreshToken: refreshToken,
	}

	tokenSource := oauthConfig.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to refresh token")
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	// Get user info to find the user
	userInfo, err := s.GetGoogleUserInfo(ctx, newToken.AccessToken)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get user info")
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// Update user's tokens
	user, err := s.repo.GetUserByGoogleID(ctx, userInfo.ID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get user")
		return nil, fmt.Errorf("user not found: %w", err)
	}

	user.AccessToken = newToken.AccessToken
	if newToken.RefreshToken != "" {
		user.RefreshToken = newToken.RefreshToken
	}
	user.TokenExpiry = &newToken.Expiry

	if err := s.repo.UpdateUser(ctx, user); err != nil {
		logger.Error().Err(err).Msg("Failed to update user tokens")
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Generate new JWT token
	jwtToken, err := middleware.GenerateJWT(cfg, user.ID, user.Email, user.Name)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to generate JWT token")
		return nil, fmt.Errorf("failed to generate JWT token: %w", err)
	}

	jwtExpiresAt := time.Now().Add(time.Duration(cfg.JWT.ExpireHour) * time.Hour)

	response := &dto.AuthResponse{
		AccessToken:  newToken.AccessToken,
		RefreshToken: user.RefreshToken,
		ExpiresAt:    newToken.Expiry,
		JWTToken:     jwtToken,
		JWTExpiresAt: jwtExpiresAt,
		User: dto.UserInfo{
			ID:          user.ID,
			Email:       user.Email,
			Name:        user.Name,
			Picture:     user.Picture,
			LastLoginAt: *user.LastLoginAt,
		},
	}

	return response, nil
}
