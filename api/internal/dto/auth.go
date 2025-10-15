package dto

import "time"

// GoogleLoginRequest represents the request to initiate Google OAuth login
type GoogleLoginRequest struct {
	RedirectURL string `json:"redirect_url,omitempty"`
}

// GoogleLoginResponse represents the response containing the OAuth URL
type GoogleLoginResponse struct {
	AuthURL string `json:"auth_url"`
}

// GoogleCallbackRequest represents the OAuth callback parameters
type GoogleCallbackRequest struct {
	Code  string `json:"code" binding:"required"`
	State string `json:"state" binding:"required"`
}

// AuthResponse represents the authentication response with tokens
type AuthResponse struct {
	AccessToken  string    `json:"access_token"`            // Google OAuth access token
	RefreshToken string    `json:"refresh_token,omitempty"` // Google OAuth refresh token
	ExpiresAt    time.Time `json:"expires_at"`              // Google OAuth token expiry
	JWTToken     string    `json:"jwt_token"`               // Our internal JWT token
	JWTExpiresAt time.Time `json:"jwt_expires_at"`          // JWT token expiry
	User         UserInfo  `json:"user"`
}

// UserInfo represents basic user information
type UserInfo struct {
	ID          uint      `json:"id"`
	Email       string    `json:"email"`
	Name        string    `json:"name"`
	Picture     string    `json:"picture"`
	LastLoginAt time.Time `json:"last_login_at"`
}

// RefreshTokenRequest represents a token refresh request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// LogoutRequest represents a logout request
type LogoutRequest struct {
	// No body needed, will use authorization header
}

// LogoutResponse represents a logout response
type LogoutResponse struct {
	Message string `json:"message"`
}
