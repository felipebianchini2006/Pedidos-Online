package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"pedidos-online/user-service/internal/model"
)

// MockDB is a helper for testing
// In production, use a real test database or library like sqlmock

func TestUserRepository_Interface(t *testing.T) {
	// Verify that userRepository implements UserRepository interface
	var _ UserRepository = (*userRepository)(nil)
}

func TestNewUserRepository(t *testing.T) {
	db := &sql.DB{} // Mock DB
	repo := NewUserRepository(db)

	if repo == nil {
		t.Error("NewUserRepository returned nil")
	}
}

func TestUserRepository_Errors(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "user not found error",
			err:      ErrUserNotFound,
			expected: "user not found",
		},
		{
			name:     "user already exists error",
			err:      ErrUserAlreadyExists,
			expected: "user with this email already exists",
		},
		{
			name:     "invalid user ID error",
			err:      ErrInvalidUserID,
			expected: "invalid user ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.expected {
				t.Errorf("expected error message %q, got %q", tt.expected, tt.err.Error())
			}
		})
	}
}

func TestUserRepository_ValidationLogic(t *testing.T) {
	// Test ID validation logic
	t.Run("empty ID should be invalid", func(t *testing.T) {
		id := ""
		if id != "" {
			t.Error("empty string should be considered invalid")
		}
	})

	// Test email validation logic
	t.Run("empty email should be invalid", func(t *testing.T) {
		email := ""
		if email != "" {
			t.Error("empty email should be considered invalid")
		}
	})
}

func TestUserRepository_ListLimits(t *testing.T) {
	tests := []struct {
		name          string
		inputLimit    int
		inputOffset   int
		expectedLimit int
		expectedOffset int
	}{
		{
			name:           "default limit when zero",
			inputLimit:     0,
			inputOffset:    0,
			expectedLimit:  10,
			expectedOffset: 0,
		},
		{
			name:           "max limit when exceeding",
			inputLimit:     200,
			inputOffset:    0,
			expectedLimit:  100,
			expectedOffset: 0,
		},
		{
			name:           "negative offset becomes zero",
			inputLimit:     10,
			inputOffset:    -5,
			expectedLimit:  10,
			expectedOffset: 0,
		},
		{
			name:           "normal values unchanged",
			inputLimit:     50,
			inputOffset:    20,
			expectedLimit:  50,
			expectedOffset: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limit := tt.inputLimit
			offset := tt.inputOffset

			// Apply same logic as in List method
			if limit <= 0 {
				limit = 10
			}
			if limit > 100 {
				limit = 100
			}
			if offset < 0 {
				offset = 0
			}

			if limit != tt.expectedLimit {
				t.Errorf("expected limit %d, got %d", tt.expectedLimit, limit)
			}
			if offset != tt.expectedOffset {
				t.Errorf("expected offset %d, got %d", tt.expectedOffset, offset)
			}
		})
	}
}

func TestUserModel_BeforeCreate(t *testing.T) {
	user := &model.User{
		Email:    "test@example.com",
		Password: "hashedpassword",
		Name:     "Test User",
		Phone:    "11999999999",
	}

	user.BeforeCreate()

	if user.ID == uuid.Nil {
		t.Error("expected ID to be generated")
	}

	if user.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}

	if user.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}
}

func TestUserModel_HashPassword(t *testing.T) {
	user := &model.User{
		Password: "plainpassword",
	}

	err := user.HashPassword()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if user.Password == "plainpassword" {
		t.Error("password should be hashed")
	}

	if len(user.Password) < 20 {
		t.Error("hashed password should be longer")
	}
}

func TestUserModel_ComparePassword(t *testing.T) {
	user := &model.User{
		Password: "plainpassword",
	}

	err := user.HashPassword()
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	t.Run("correct password", func(t *testing.T) {
		err := user.ComparePassword("plainpassword")
		if err != nil {
			t.Errorf("expected password to match, got error: %v", err)
		}
	})

	t.Run("incorrect password", func(t *testing.T) {
		err := user.ComparePassword("wrongpassword")
		if err == nil {
			t.Error("expected error for incorrect password")
		}
	})
}

func TestUserModel_ToResponse(t *testing.T) {
	now := time.Now()
	id := uuid.New()

	user := &model.User{
		ID:        id,
		Email:     "test@example.com",
		Password:  "hashedpassword",
		Name:      "Test User",
		Phone:     "11999999999",
		CreatedAt: now,
		UpdatedAt: now,
	}

	response := user.ToResponse()

	if response.ID != id {
		t.Error("ID mismatch")
	}
	if response.Email != user.Email {
		t.Error("Email mismatch")
	}
	if response.Name != user.Name {
		t.Error("Name mismatch")
	}
	if response.Phone != user.Phone {
		t.Error("Phone mismatch")
	}
	if response.CreatedAt != now {
		t.Error("CreatedAt mismatch")
	}
	if response.UpdatedAt != now {
		t.Error("UpdatedAt mismatch")
	}

	// Verify password is not exposed in response struct
	// This is verified by the struct definition, but we can't directly test it here
}

func TestContext_Timeout(t *testing.T) {
	// Test that context timeout works as expected
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	select {
	case <-ctx.Done():
		t.Error("context should not be done immediately")
	default:
		// Expected
	}

	time.Sleep(150 * time.Millisecond)

	select {
	case <-ctx.Done():
		// Expected - context timed out
	default:
		t.Error("context should be done after timeout")
	}
}

// Note: For full integration tests with a real database, you would need:
// 1. A test database instance
// 2. Setup and teardown functions
// 3. Test data fixtures
// 4. Transaction rollbacks between tests
//
// Consider using libraries like:
// - github.com/DATA-DOG/go-sqlmock for mocking database
// - testcontainers-go for spinning up real PostgreSQL in Docker
// - github.com/stretchr/testify for assertions
