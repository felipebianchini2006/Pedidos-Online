package repository

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pedidos-online/user-service/internal/model"
)

// setupTestDB creates a mock database connection and sqlmock
func setupTestDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, UserRepository) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	repo := NewUserRepository(db)

	return db, mock, repo
}

func TestCreate(t *testing.T) {
	t.Run("successfully create user", func(t *testing.T) {
		db, mock, repo := setupTestDB(t)
		defer db.Close()

		ctx := context.Background()
		user := &model.User{
			ID:        uuid.New(),
			Email:     "test@example.com",
			Password:  "hashedpassword123",
			Name:      "Test User",
			Phone:     "1234567890",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Expect email uniqueness check
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`)).
			WithArgs(user.Email).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		// Expect insert query
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO users (id, email, password, name, phone, created_at, updated_at)`)).
			WithArgs(user.ID, user.Email, user.Password, user.Name, user.Phone, user.CreatedAt, user.UpdatedAt).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Create(ctx, user)

		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error when email already exists", func(t *testing.T) {
		db, mock, repo := setupTestDB(t)
		defer db.Close()

		ctx := context.Background()
		user := &model.User{
			ID:        uuid.New(),
			Email:     "existing@example.com",
			Password:  "hashedpassword123",
			Name:      "Test User",
			Phone:     "1234567890",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Expect email uniqueness check to return true (email exists)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`)).
			WithArgs(user.Email).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		err := repo.Create(ctx, user)

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrUserAlreadyExists)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error when database insert fails", func(t *testing.T) {
		db, mock, repo := setupTestDB(t)
		defer db.Close()

		ctx := context.Background()
		user := &model.User{
			ID:        uuid.New(),
			Email:     "test@example.com",
			Password:  "hashedpassword123",
			Name:      "Test User",
			Phone:     "1234567890",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Expect email uniqueness check
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`)).
			WithArgs(user.Email).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		// Expect insert query to fail
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO users`)).
			WithArgs(user.ID, user.Email, user.Password, user.Name, user.Phone, user.CreatedAt, user.UpdatedAt).
			WillReturnError(sql.ErrConnDone)

		err := repo.Create(ctx, user)

		require.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error when checking email uniqueness fails", func(t *testing.T) {
		db, mock, repo := setupTestDB(t)
		defer db.Close()

		ctx := context.Background()
		user := &model.User{
			ID:    uuid.New(),
			Email: "test@example.com",
		}

		// Expect email uniqueness check to fail
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`)).
			WithArgs(user.Email).
			WillReturnError(sql.ErrConnDone)

		err := repo.Create(ctx, user)

		require.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestFindByEmail(t *testing.T) {
	t.Run("successfully find user by email", func(t *testing.T) {
		db, mock, repo := setupTestDB(t)
		defer db.Close()

		ctx := context.Background()
		expectedUser := &model.User{
			ID:        uuid.New(),
			Email:     "test@example.com",
			Password:  "hashedpassword123",
			Name:      "Test User",
			Phone:     "1234567890",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		rows := sqlmock.NewRows([]string{"id", "email", "password", "name", "phone", "created_at", "updated_at"}).
			AddRow(expectedUser.ID, expectedUser.Email, expectedUser.Password, expectedUser.Name, expectedUser.Phone, expectedUser.CreatedAt, expectedUser.UpdatedAt)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, email, password, name, phone, created_at, updated_at FROM users WHERE email = $1`)).
			WithArgs(expectedUser.Email).
			WillReturnRows(rows)

		user, err := repo.FindByEmail(ctx, expectedUser.Email)

		require.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, expectedUser.ID, user.ID)
		assert.Equal(t, expectedUser.Email, user.Email)
		assert.Equal(t, expectedUser.Password, user.Password)
		assert.Equal(t, expectedUser.Name, user.Name)
		assert.Equal(t, expectedUser.Phone, user.Phone)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("return error when user not found", func(t *testing.T) {
		db, mock, repo := setupTestDB(t)
		defer db.Close()

		ctx := context.Background()
		email := "nonexistent@example.com"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, email, password, name, phone, created_at, updated_at FROM users WHERE email = $1`)).
			WithArgs(email).
			WillReturnError(sql.ErrNoRows)

		user, err := repo.FindByEmail(ctx, email)

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrUserNotFound)
		assert.Nil(t, user)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error when email is empty", func(t *testing.T) {
		db, mock, repo := setupTestDB(t)
		defer db.Close()

		ctx := context.Background()

		user, err := repo.FindByEmail(ctx, "")

		require.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "email cannot be empty")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error when database query fails", func(t *testing.T) {
		db, mock, repo := setupTestDB(t)
		defer db.Close()

		ctx := context.Background()
		email := "test@example.com"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, email, password, name, phone, created_at, updated_at FROM users WHERE email = $1`)).
			WithArgs(email).
			WillReturnError(sql.ErrConnDone)

		user, err := repo.FindByEmail(ctx, email)

		require.Error(t, err)
		assert.Nil(t, user)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestFindByID(t *testing.T) {
	t.Run("successfully find user by ID", func(t *testing.T) {
		db, mock, repo := setupTestDB(t)
		defer db.Close()

		ctx := context.Background()
		expectedUser := &model.User{
			ID:        uuid.New(),
			Email:     "test@example.com",
			Name:      "Test User",
			Phone:     "1234567890",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		rows := sqlmock.NewRows([]string{"id", "email", "name", "phone", "created_at", "updated_at"}).
			AddRow(expectedUser.ID, expectedUser.Email, expectedUser.Name, expectedUser.Phone, expectedUser.CreatedAt, expectedUser.UpdatedAt)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, email, name, phone, created_at, updated_at FROM users WHERE id = $1`)).
			WithArgs(expectedUser.ID.String()).
			WillReturnRows(rows)

		user, err := repo.FindByID(ctx, expectedUser.ID.String())

		require.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, expectedUser.ID, user.ID)
		assert.Equal(t, expectedUser.Email, user.Email)
		assert.Equal(t, expectedUser.Name, user.Name)
		assert.Equal(t, expectedUser.Phone, user.Phone)
		assert.Empty(t, user.Password) // Password should not be returned
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("return error when user not found", func(t *testing.T) {
		db, mock, repo := setupTestDB(t)
		defer db.Close()

		ctx := context.Background()
		userID := uuid.New().String()

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, email, name, phone, created_at, updated_at FROM users WHERE id = $1`)).
			WithArgs(userID).
			WillReturnError(sql.ErrNoRows)

		user, err := repo.FindByID(ctx, userID)

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrUserNotFound)
		assert.Nil(t, user)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error for invalid user ID", func(t *testing.T) {
		db, mock, repo := setupTestDB(t)
		defer db.Close()

		ctx := context.Background()

		user, err := repo.FindByID(ctx, "")

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidUserID)
		assert.Nil(t, user)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error when database query fails", func(t *testing.T) {
		db, mock, repo := setupTestDB(t)
		defer db.Close()

		ctx := context.Background()
		userID := uuid.New().String()

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, email, name, phone, created_at, updated_at FROM users WHERE id = $1`)).
			WithArgs(userID).
			WillReturnError(sql.ErrConnDone)

		user, err := repo.FindByID(ctx, userID)

		require.Error(t, err)
		assert.Nil(t, user)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUpdate(t *testing.T) {
	t.Run("successfully update user", func(t *testing.T) {
		db, mock, repo := setupTestDB(t)
		defer db.Close()

		ctx := context.Background()
		user := &model.User{
			ID:        uuid.New(),
			Name:      "Updated Name",
			Phone:     "9876543210",
			UpdatedAt: time.Now(),
		}

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE users SET name = $1, phone = $2, updated_at = $3 WHERE id = $4`)).
			WithArgs(user.Name, user.Phone, user.UpdatedAt, user.ID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Update(ctx, user)

		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error when user not found", func(t *testing.T) {
		db, mock, repo := setupTestDB(t)
		defer db.Close()

		ctx := context.Background()
		user := &model.User{
			ID:        uuid.New(),
			Name:      "Updated Name",
			Phone:     "9876543210",
			UpdatedAt: time.Now(),
		}

		// No rows affected
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE users SET name = $1, phone = $2, updated_at = $3 WHERE id = $4`)).
			WithArgs(user.Name, user.Phone, user.UpdatedAt, user.ID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Update(ctx, user)

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrUserNotFound)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error for invalid user ID", func(t *testing.T) {
		db, mock, repo := setupTestDB(t)
		defer db.Close()

		ctx := context.Background()
		user := &model.User{
			ID:        uuid.Nil,
			Name:      "Updated Name",
			Phone:     "1234567890",
			UpdatedAt: time.Now(),
		}

		// uuid.Nil.String() returns "00000000-0000-0000-0000-000000000000", not empty string
		// So the update will be attempted but return 0 rows affected
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE users SET name = $1, phone = $2, updated_at = $3 WHERE id = $4`)).
			WithArgs(user.Name, user.Phone, user.UpdatedAt, user.ID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Update(ctx, user)

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrUserNotFound)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error when database update fails", func(t *testing.T) {
		db, mock, repo := setupTestDB(t)
		defer db.Close()

		ctx := context.Background()
		user := &model.User{
			ID:        uuid.New(),
			Name:      "Updated Name",
			Phone:     "9876543210",
			UpdatedAt: time.Now(),
		}

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE users SET name = $1, phone = $2, updated_at = $3 WHERE id = $4`)).
			WithArgs(user.Name, user.Phone, user.UpdatedAt, user.ID).
			WillReturnError(sql.ErrConnDone)

		err := repo.Update(ctx, user)

		require.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDelete(t *testing.T) {
	t.Run("successfully delete user", func(t *testing.T) {
		db, mock, repo := setupTestDB(t)
		defer db.Close()

		ctx := context.Background()
		userID := uuid.New().String()

		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM users WHERE id = $1`)).
			WithArgs(userID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Delete(ctx, userID)

		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error when user not found", func(t *testing.T) {
		db, mock, repo := setupTestDB(t)
		defer db.Close()

		ctx := context.Background()
		userID := uuid.New().String()

		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM users WHERE id = $1`)).
			WithArgs(userID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Delete(ctx, userID)

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrUserNotFound)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error for empty user ID", func(t *testing.T) {
		db, mock, repo := setupTestDB(t)
		defer db.Close()

		ctx := context.Background()

		err := repo.Delete(ctx, "")

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidUserID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestList(t *testing.T) {
	t.Run("successfully list users", func(t *testing.T) {
		db, mock, repo := setupTestDB(t)
		defer db.Close()

		ctx := context.Background()
		limit := 10
		offset := 0

		user1 := &model.User{
			ID:        uuid.New(),
			Email:     "user1@example.com",
			Name:      "User 1",
			Phone:     "1111111111",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		user2 := &model.User{
			ID:        uuid.New(),
			Email:     "user2@example.com",
			Name:      "User 2",
			Phone:     "2222222222",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		rows := sqlmock.NewRows([]string{"id", "email", "name", "phone", "created_at", "updated_at"}).
			AddRow(user1.ID, user1.Email, user1.Name, user1.Phone, user1.CreatedAt, user1.UpdatedAt).
			AddRow(user2.ID, user2.Email, user2.Name, user2.Phone, user2.CreatedAt, user2.UpdatedAt)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, email, name, phone, created_at, updated_at FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`)).
			WithArgs(limit, offset).
			WillReturnRows(rows)

		users, err := repo.List(ctx, limit, offset)

		require.NoError(t, err)
		require.Len(t, users, 2)
		assert.Equal(t, user1.Email, users[0].Email)
		assert.Equal(t, user2.Email, users[1].Email)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("use default limit when zero", func(t *testing.T) {
		db, mock, repo := setupTestDB(t)
		defer db.Close()

		ctx := context.Background()

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, email, name, phone, created_at, updated_at FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`)).
			WithArgs(10, 0). // Default limit is 10
			WillReturnRows(sqlmock.NewRows([]string{"id", "email", "name", "phone", "created_at", "updated_at"}))

		users, err := repo.List(ctx, 0, 0)

		require.NoError(t, err)
		require.NotNil(t, users)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("cap limit at maximum", func(t *testing.T) {
		db, mock, repo := setupTestDB(t)
		defer db.Close()

		ctx := context.Background()

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, email, name, phone, created_at, updated_at FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`)).
			WithArgs(100, 0). // Maximum limit is 100
			WillReturnRows(sqlmock.NewRows([]string{"id", "email", "name", "phone", "created_at", "updated_at"}))

		users, err := repo.List(ctx, 200, 0)

		require.NoError(t, err)
		require.NotNil(t, users)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("return empty list when no users", func(t *testing.T) {
		db, mock, repo := setupTestDB(t)
		defer db.Close()

		ctx := context.Background()

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, email, name, phone, created_at, updated_at FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`)).
			WithArgs(10, 0).
			WillReturnRows(sqlmock.NewRows([]string{"id", "email", "name", "phone", "created_at", "updated_at"}))

		users, err := repo.List(ctx, 10, 0)

		require.NoError(t, err)
		require.NotNil(t, users)
		assert.Len(t, users, 0)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCount(t *testing.T) {
	t.Run("successfully count users", func(t *testing.T) {
		db, mock, repo := setupTestDB(t)
		defer db.Close()

		ctx := context.Background()
		expectedCount := int64(42)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM users`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))

		count, err := repo.Count(ctx)

		require.NoError(t, err)
		assert.Equal(t, expectedCount, count)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("return zero when no users", func(t *testing.T) {
		db, mock, repo := setupTestDB(t)
		defer db.Close()

		ctx := context.Background()

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM users`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

		count, err := repo.Count(ctx)

		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error when database query fails", func(t *testing.T) {
		db, mock, repo := setupTestDB(t)
		defer db.Close()

		ctx := context.Background()

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM users`)).
			WillReturnError(sql.ErrConnDone)

		count, err := repo.Count(ctx)

		require.Error(t, err)
		assert.Equal(t, int64(0), count)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
