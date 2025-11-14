package jwt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateToken(t *testing.T) {
	secret := "test-secret-key"
	userID := "user-123"
	email := "test@example.com"
	expiration := 24 * time.Hour

	t.Run("successfully generate token with valid data", func(t *testing.T) {
		token, err := GenerateToken(userID, email, secret, expiration)

		require.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.Contains(t, token, ".")
	})

	t.Run("generated token contains correct claims", func(t *testing.T) {
		token, err := GenerateToken(userID, email, secret, expiration)
		require.NoError(t, err)

		// Validate and extract claims
		claims, err := ValidateToken(token, secret)
		require.NoError(t, err)

		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, email, claims.Email)
		assert.Equal(t, userID, claims.Subject)
		assert.Equal(t, "pedidos-online-user-service", claims.Issuer)
		assert.NotNil(t, claims.ExpiresAt)
		assert.NotNil(t, claims.IssuedAt)
		assert.NotNil(t, claims.NotBefore)
	})

	t.Run("use default expiration when zero", func(t *testing.T) {
		token, err := GenerateToken(userID, email, secret, 0)
		require.NoError(t, err)

		claims, err := ValidateToken(token, secret)
		require.NoError(t, err)

		// Check that expiration is approximately 24 hours from now
		expectedExpiry := time.Now().Add(24 * time.Hour)
		actualExpiry := claims.ExpiresAt.Time
		timeDiff := actualExpiry.Sub(expectedExpiry).Abs()
		assert.Less(t, timeDiff, 5*time.Second, "Expiration should be approximately 24 hours")
	})

	t.Run("error when secret is empty string", func(t *testing.T) {
		// Unset env var if it exists
		t.Setenv("JWT_SECRET", "")

		token, err := GenerateToken(userID, email, "", expiration)

		require.Error(t, err)
		assert.Empty(t, token)
		assert.Contains(t, err.Error(), "JWT_SECRET is not set")
	})

	t.Run("use environment variable when secret is empty", func(t *testing.T) {
		t.Setenv("JWT_SECRET", "env-secret")

		token, err := GenerateToken(userID, email, "", expiration)

		require.NoError(t, err)
		assert.NotEmpty(t, token)

		// Validate with env secret
		claims, err := ValidateToken(token, "env-secret")
		require.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
	})
}

func TestValidateToken(t *testing.T) {
	secret := "test-secret-key"
	userID := "user-123"
	email := "test@example.com"

	t.Run("successfully validate valid token", func(t *testing.T) {
		token, err := GenerateToken(userID, email, secret, 24*time.Hour)
		require.NoError(t, err)

		claims, err := ValidateToken(token, secret)

		require.NoError(t, err)
		require.NotNil(t, claims)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, email, claims.Email)
	})

	t.Run("error with expired token", func(t *testing.T) {
		// Generate token with negative expiration (already expired)
		token, err := GenerateToken(userID, email, secret, -1*time.Hour)
		require.NoError(t, err)

		claims, err := ValidateToken(token, secret)

		require.Error(t, err)
		assert.Nil(t, claims)
		assert.Contains(t, err.Error(), "expired")
	})

	t.Run("error with invalid signature", func(t *testing.T) {
		token, err := GenerateToken(userID, email, secret, 24*time.Hour)
		require.NoError(t, err)

		// Try to validate with different secret
		claims, err := ValidateToken(token, "wrong-secret")

		require.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("error with malformed token", func(t *testing.T) {
		claims, err := ValidateToken("malformed.token.string", secret)

		require.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("error with empty token", func(t *testing.T) {
		claims, err := ValidateToken("", secret)

		require.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("error when secret is empty and env var not set", func(t *testing.T) {
		t.Setenv("JWT_SECRET", "")

		token, err := GenerateToken(userID, email, secret, 24*time.Hour)
		require.NoError(t, err)

		claims, err := ValidateToken(token, "")

		require.Error(t, err)
		assert.Nil(t, claims)
		assert.Contains(t, err.Error(), "JWT_SECRET is not set")
	})

	t.Run("error with tampered token", func(t *testing.T) {
		token, err := GenerateToken(userID, email, secret, 24*time.Hour)
		require.NoError(t, err)

		// Tamper with the token by changing a character
		tamperedToken := token[:len(token)-1] + "x"

		claims, err := ValidateToken(tamperedToken, secret)

		require.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("verify HMAC signing method is accepted", func(t *testing.T) {
		// Create a valid token and verify our method check works with HMAC
		validToken, _ := GenerateToken(userID, email, secret, 24*time.Hour)
		result, err := ValidateToken(validToken, secret)

		// This should succeed with HMAC (HS256)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, userID, result.UserID)
		assert.Equal(t, email, result.Email)
	})
}

func TestExtractTokenFromHeader(t *testing.T) {
	t.Run("successfully extract token from valid header", func(t *testing.T) {
		authHeader := "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test.token"

		token, err := ExtractTokenFromHeader(authHeader)

		require.NoError(t, err)
		assert.Equal(t, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test.token", token)
	})

	t.Run("error with empty header", func(t *testing.T) {
		token, err := ExtractTokenFromHeader("")

		require.Error(t, err)
		assert.Empty(t, token)
		assert.Contains(t, err.Error(), "authorization header is empty")
	})

	t.Run("error with invalid header format", func(t *testing.T) {
		authHeader := "InvalidFormat token"

		token, err := ExtractTokenFromHeader(authHeader)

		require.Error(t, err)
		assert.Empty(t, token)
		assert.Contains(t, err.Error(), "invalid authorization header format")
	})

	t.Run("error with missing Bearer prefix", func(t *testing.T) {
		authHeader := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test.token"

		token, err := ExtractTokenFromHeader(authHeader)

		require.Error(t, err)
		assert.Empty(t, token)
		assert.Contains(t, err.Error(), "invalid authorization header format")
	})

	t.Run("error with Bearer prefix but empty token", func(t *testing.T) {
		authHeader := "Bearer "

		token, err := ExtractTokenFromHeader(authHeader)

		require.Error(t, err)
		assert.Empty(t, token)
		assert.Contains(t, err.Error(), "token is empty")
	})

	t.Run("successfully extract token with spaces", func(t *testing.T) {
		authHeader := "Bearer   eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test.token"

		token, err := ExtractTokenFromHeader(authHeader)

		require.NoError(t, err)
		// Note: our implementation doesn't trim spaces after "Bearer ", so this includes leading spaces
		assert.Equal(t, "  eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test.token", token)
	})
}

func TestTokenExpiration(t *testing.T) {
	secret := "test-secret-key"
	userID := "user-123"
	email := "test@example.com"

	t.Run("token expires after specified duration", func(t *testing.T) {
		// Create token that expires in 1 second
		token, err := GenerateToken(userID, email, secret, 1*time.Second)
		require.NoError(t, err)

		// Validate immediately (should succeed)
		claims, err := ValidateToken(token, secret)
		require.NoError(t, err)
		assert.NotNil(t, claims)

		// Wait for token to expire
		time.Sleep(2 * time.Second)

		// Validate after expiration (should fail)
		claims, err = ValidateToken(token, secret)
		require.Error(t, err)
		assert.Nil(t, claims)
		assert.Contains(t, err.Error(), "expired")
	})
}

func TestTokenClaims(t *testing.T) {
	secret := "test-secret-key"

	t.Run("token contains all required claims", func(t *testing.T) {
		userID := "user-456"
		email := "another@example.com"

		token, err := GenerateToken(userID, email, secret, 24*time.Hour)
		require.NoError(t, err)

		claims, err := ValidateToken(token, secret)
		require.NoError(t, err)

		// Verify all claims
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, email, claims.Email)
		assert.Equal(t, userID, claims.Subject)
		assert.Equal(t, "pedidos-online-user-service", claims.Issuer)

		// Verify timestamps are set and reasonable
		now := time.Now()
		assert.True(t, claims.IssuedAt.Time.Before(now) || claims.IssuedAt.Time.Equal(now))
		assert.True(t, claims.NotBefore.Time.Before(now) || claims.NotBefore.Time.Equal(now))
		assert.True(t, claims.ExpiresAt.Time.After(now))
	})
}
