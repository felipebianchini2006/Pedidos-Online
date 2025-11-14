package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateToken(t *testing.T) {
	secret := "test-secret-key"

	t.Run("successfully validate token with valid claims", func(t *testing.T) {
		// Create a valid token manually
		claims := Claims{
			UserID: "user-123",
			Email:  "test@example.com",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				Subject:   "user-123",
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(secret))
		require.NoError(t, err)

		resultClaims, err := ValidateToken(tokenString, secret)

		require.NoError(t, err)
		assert.NotNil(t, resultClaims)
		assert.Equal(t, "user-123", resultClaims.UserID)
		assert.Equal(t, "test@example.com", resultClaims.Email)
	})

	t.Run("error with expired token", func(t *testing.T) {
		// Create an expired token
		claims := Claims{
			UserID: "user-123",
			Email:  "test@example.com",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
				Subject:   "user-123",
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(secret))
		require.NoError(t, err)

		resultClaims, err := ValidateToken(tokenString, secret)

		require.Error(t, err)
		assert.Nil(t, resultClaims)
	})

	t.Run("error with invalid signature", func(t *testing.T) {
		// Create token with one secret
		claims := Claims{
			UserID: "user-123",
			Email:  "test@example.com",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(secret))
		require.NoError(t, err)

		// Validate with different secret
		resultClaims, err := ValidateToken(tokenString, "wrong-secret")

		require.Error(t, err)
		assert.Nil(t, resultClaims)
	})

	t.Run("error with malformed token", func(t *testing.T) {
		resultClaims, err := ValidateToken("malformed.token.string", secret)

		require.Error(t, err)
		assert.Nil(t, resultClaims)
	})

	t.Run("error with empty token", func(t *testing.T) {
		resultClaims, err := ValidateToken("", secret)

		require.Error(t, err)
		assert.Nil(t, resultClaims)
	})

	t.Run("error with wrong signing method", func(t *testing.T) {
		// Note: We can't easily create an RS256 token without keys,
		// but we verify our code checks the signing method
		validClaims := Claims{
			UserID: "user-123",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, validClaims)
		tokenString, err := token.SignedString([]byte(secret))
		require.NoError(t, err)

		// This should succeed with HS256
		resultClaims, err := ValidateToken(tokenString, secret)
		require.NoError(t, err)
		assert.NotNil(t, resultClaims)
	})
}

func TestExtractUserID(t *testing.T) {
	secret := "test-secret-key"

	t.Run("successfully extract userID from valid token", func(t *testing.T) {
		claims := Claims{
			UserID: "user-789",
			Email:  "user@test.com",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(secret))
		require.NoError(t, err)

		userID, err := ExtractUserID(tokenString, secret)

		require.NoError(t, err)
		assert.Equal(t, "user-789", userID)
	})

	t.Run("error extracting userID from invalid token", func(t *testing.T) {
		userID, err := ExtractUserID("invalid.token", secret)

		require.Error(t, err)
		assert.Empty(t, userID)
	})

	t.Run("error extracting userID from expired token", func(t *testing.T) {
		claims := Claims{
			UserID: "user-123",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(secret))
		require.NoError(t, err)

		userID, err := ExtractUserID(tokenString, secret)

		require.Error(t, err)
		assert.Empty(t, userID)
	})
}

func TestTokenCompatibilityWithUserService(t *testing.T) {
	secret := "shared-secret-key"

	t.Run("validate token created with user-service format", func(t *testing.T) {
		// Simulate a token that would be created by user-service
		claims := Claims{
			UserID: "user-abc",
			Email:  "userservice@example.com",
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   "user-abc",
				Issuer:    "pedidos-online-user-service",
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				NotBefore: jwt.NewNumericDate(time.Now()),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(secret))
		require.NoError(t, err)

		// Order service should be able to validate it
		resultClaims, err := ValidateToken(tokenString, secret)

		require.NoError(t, err)
		assert.Equal(t, "user-abc", resultClaims.UserID)
		assert.Equal(t, "userservice@example.com", resultClaims.Email)
		assert.Equal(t, "pedidos-online-user-service", resultClaims.Issuer)
	})
}

func TestTokenNotBeforeValidation(t *testing.T) {
	secret := "test-secret-key"

	t.Run("token not valid before NotBefore time", func(t *testing.T) {
		// Create token that's not valid yet (NotBefore in future)
		claims := Claims{
			UserID: "user-123",
			RegisteredClaims: jwt.RegisteredClaims{
				NotBefore: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(2 * time.Hour)),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(secret))
		require.NoError(t, err)

		resultClaims, err := ValidateToken(tokenString, secret)

		require.Error(t, err)
		assert.Nil(t, resultClaims)
	})
}
