package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"pedidos-online/user-service/internal/model"
	"pedidos-online/user-service/internal/repository"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByID(ctx context.Context, id string) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, limit, offset int) ([]*model.User, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.User), args.Error(1)
}

func (m *MockUserRepository) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func TestRegister(t *testing.T) {
	jwtSecret := "test-secret"
	jwtExpiration := 24 * time.Hour

	t.Run("successfully register valid user", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo, jwtSecret, jwtExpiration)

		ctx := context.Background()
		email := "test@example.com"
		password := "strongpassword123"
		name := "Test User"
		phone := "1234567890"

		mockRepo.On("FindByEmail", ctx, email).Return(nil, repository.ErrUserNotFound)

		passwordWasHashed := false
		mockRepo.On("Create", ctx, mock.MatchedBy(func(u *model.User) bool {
			if u.Email == email && u.Name == name && u.Phone == phone {
				passwordWasHashed = u.Password != password && len(u.Password) > 20
				return true
			}
			return false
		})).Return(nil)

		user, err := service.Register(ctx, email, password, name, phone)

		require.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, email, user.Email)
		assert.Equal(t, name, user.Name)
		assert.Equal(t, phone, user.Phone)
		assert.Empty(t, user.Password, "password should be cleared from returned user")
		assert.NotEqual(t, uuid.Nil, user.ID)
		assert.True(t, passwordWasHashed, "password should be hashed before saving to database")
		mockRepo.AssertExpectations(t)
	})

	t.Run("error for invalid email format", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo, jwtSecret, jwtExpiration)

		ctx := context.Background()
		user, err := service.Register(ctx, "invalid-email", "strongpassword123", "Test User", "1234567890")

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidEmail)
		assert.Nil(t, user)
		mockRepo.AssertExpectations(t)
	})

	t.Run("error for weak password", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo, jwtSecret, jwtExpiration)

		ctx := context.Background()
		user, err := service.Register(ctx, "test@example.com", "weak", "Test User", "1234567890")

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrWeakPassword)
		assert.Nil(t, user)
	})

	t.Run("error when email already exists", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo, jwtSecret, jwtExpiration)

		ctx := context.Background()
		email := "existing@example.com"
		existingUser := &model.User{ID: uuid.New(), Email: email}

		mockRepo.On("FindByEmail", ctx, email).Return(existingUser, nil)

		user, err := service.Register(ctx, email, "strongpassword123", "Test User", "1234567890")

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrEmailAlreadyExists)
		assert.Nil(t, user)
		mockRepo.AssertExpectations(t)
	})
}

func TestLogin(t *testing.T) {
	jwtSecret := "test-secret"
	jwtExpiration := 24 * time.Hour

	t.Run("successfully login with valid credentials", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo, jwtSecret, jwtExpiration)

		ctx := context.Background()
		email := "test@example.com"
		password := "correctpassword"

		user := &model.User{
			ID:       uuid.New(),
			Email:    email,
			Password: password,
			Name:     "Test User",
			Phone:    "1234567890",
		}
		_ = user.HashPassword()

		mockRepo.On("FindByEmail", ctx, email).Return(user, nil)

		token, err := service.Login(ctx, email, password)

		require.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.Contains(t, token, ".")
		mockRepo.AssertExpectations(t)
	})

	t.Run("token contains correct format", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo, jwtSecret, jwtExpiration)

		ctx := context.Background()
		email := "test@example.com"
		password := "correctpassword"

		user := &model.User{ID: uuid.New(), Email: email, Password: password}
		_ = user.HashPassword()

		mockRepo.On("FindByEmail", ctx, email).Return(user, nil)

		token, err := service.Login(ctx, email, password)

		require.NoError(t, err)

		userID, err := service.ValidateToken(token)
		require.NoError(t, err)
		assert.Equal(t, user.ID.String(), userID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("error for non-existent email", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo, jwtSecret, jwtExpiration)

		ctx := context.Background()
		mockRepo.On("FindByEmail", ctx, "nonexistent@example.com").Return(nil, repository.ErrUserNotFound)

		token, err := service.Login(ctx, "nonexistent@example.com", "somepassword")

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidCredentials)
		assert.Empty(t, token)
		mockRepo.AssertExpectations(t)
	})

	t.Run("error for incorrect password", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo, jwtSecret, jwtExpiration)

		ctx := context.Background()
		email := "test@example.com"

		user := &model.User{ID: uuid.New(), Email: email, Password: "correctpassword"}
		_ = user.HashPassword()

		mockRepo.On("FindByEmail", ctx, email).Return(user, nil)

		token, err := service.Login(ctx, email, "wrongpassword")

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidCredentials)
		assert.Empty(t, token)
		mockRepo.AssertExpectations(t)
	})

	t.Run("error for empty credentials", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo, jwtSecret, jwtExpiration)

		ctx := context.Background()

		token1, err1 := service.Login(ctx, "", "password")
		assert.Error(t, err1)
		assert.ErrorIs(t, err1, ErrInvalidCredentials)
		assert.Empty(t, token1)

		token2, err2 := service.Login(ctx, "test@example.com", "")
		assert.Error(t, err2)
		assert.ErrorIs(t, err2, ErrInvalidCredentials)
		assert.Empty(t, token2)
	})
}

func TestGetProfile(t *testing.T) {
	jwtSecret := "test-secret"
	jwtExpiration := 24 * time.Hour

	t.Run("successfully get user profile", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo, jwtSecret, jwtExpiration)

		ctx := context.Background()
		userID := uuid.New().String()

		expectedUser := &model.User{
			ID:    uuid.MustParse(userID),
			Email: "test@example.com",
			Name:  "Test User",
			Phone: "1234567890",
		}

		mockRepo.On("FindByID", ctx, userID).Return(expectedUser, nil)

		user, err := service.GetProfile(ctx, userID)

		require.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, expectedUser.ID, user.ID)
		assert.Equal(t, expectedUser.Email, user.Email)
		assert.Equal(t, expectedUser.Name, user.Name)
		assert.Equal(t, expectedUser.Phone, user.Phone)
		assert.Empty(t, user.Password)
		mockRepo.AssertExpectations(t)
	})

	t.Run("error for non-existent user", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo, jwtSecret, jwtExpiration)

		ctx := context.Background()
		userID := uuid.New().String()

		mockRepo.On("FindByID", ctx, userID).Return(nil, repository.ErrUserNotFound)

		user, err := service.GetProfile(ctx, userID)

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrUserNotFound)
		assert.Nil(t, user)
		mockRepo.AssertExpectations(t)
	})

	t.Run("error for invalid user ID", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo, jwtSecret, jwtExpiration)

		ctx := context.Background()

		user, err := service.GetProfile(ctx, "")

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidUserID)
		assert.Nil(t, user)
	})
}

func TestUpdateProfile(t *testing.T) {
	jwtSecret := "test-secret"
	jwtExpiration := 24 * time.Hour

	t.Run("successfully update user profile", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo, jwtSecret, jwtExpiration)

		ctx := context.Background()
		userID := uuid.New().String()

		existingUser := &model.User{
			ID:    uuid.MustParse(userID),
			Email: "test@example.com",
			Name:  "Old Name",
			Phone: "1234567890",
		}

		mockRepo.On("FindByID", ctx, userID).Return(existingUser, nil)
		mockRepo.On("Update", ctx, mock.MatchedBy(func(u *model.User) bool {
			return u.ID.String() == userID && u.Name == "Updated Name" && u.Phone == "9876543210"
		})).Return(nil)

		user, err := service.UpdateProfile(ctx, userID, "Updated Name", "9876543210")

		require.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, "Updated Name", user.Name)
		assert.Equal(t, "9876543210", user.Phone)
		mockRepo.AssertExpectations(t)
	})

	t.Run("return current user when no changes", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo, jwtSecret, jwtExpiration)

		ctx := context.Background()
		userID := uuid.New().String()

		existingUser := &model.User{
			ID:    uuid.MustParse(userID),
			Email: "test@example.com",
			Name:  "Test User",
			Phone: "1234567890",
		}

		mockRepo.On("FindByID", ctx, userID).Return(existingUser, nil)

		user, err := service.UpdateProfile(ctx, userID, "Test User", "1234567890")

		require.NoError(t, err)
		assert.Equal(t, existingUser.Name, user.Name)
		assert.Equal(t, existingUser.Phone, user.Phone)
		mockRepo.AssertExpectations(t)
	})

	t.Run("error for non-existent user", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo, jwtSecret, jwtExpiration)

		ctx := context.Background()
		userID := uuid.New().String()

		mockRepo.On("FindByID", ctx, userID).Return(nil, repository.ErrUserNotFound)

		user, err := service.UpdateProfile(ctx, userID, "New Name", "1234567890")

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrUserNotFound)
		assert.Nil(t, user)
		mockRepo.AssertExpectations(t)
	})
}

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		email string
		valid bool
	}{
		{"test@example.com", true},
		{"user.name@example.co.uk", true},
		{"user+tag@example.com", true},
		{"invalid-email", false},
		{"@example.com", false},
		{"user@", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			result := isValidEmail(tt.email)
			assert.Equal(t, tt.valid, result)
		})
	}
}

func TestIsValidPassword(t *testing.T) {
	tests := []struct {
		password string
		valid    bool
	}{
		{"12345678", true},
		{"strongpassword", true},
		{"Pass123!", true},
		{"short", false},
		{"1234567", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.password, func(t *testing.T) {
			result := isValidPassword(tt.password)
			assert.Equal(t, tt.valid, result)
		})
	}
}
