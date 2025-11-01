package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"pedidos-online/user-service/internal/model"
)

// Common repository errors
var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user with this email already exists")
	ErrInvalidUserID     = errors.New("invalid user ID")
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	// Create inserts a new user into the database
	Create(ctx context.Context, user *model.User) error

	// FindByID retrieves a user by their ID
	FindByID(ctx context.Context, id string) (*model.User, error)

	// FindByEmail retrieves a user by their email (includes password for authentication)
	FindByEmail(ctx context.Context, email string) (*model.User, error)

	// Update updates an existing user's information
	Update(ctx context.Context, user *model.User) error

	// Delete soft or hard deletes a user by their ID
	Delete(ctx context.Context, id string) error

	// List retrieves a paginated list of users
	List(ctx context.Context, limit, offset int) ([]*model.User, error)

	// Count returns the total number of users
	Count(ctx context.Context) (int64, error)
}

// userRepository implements the UserRepository interface
type userRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new instance of UserRepository
func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

// Create inserts a new user into the database
// It validates that the email is unique before inserting
func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	// Check if email already exists
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	err := r.db.QueryRowContext(ctx, checkQuery, user.Email).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check email uniqueness: %w", err)
	}
	if exists {
		return ErrUserAlreadyExists
	}

	// Insert new user
	query := `
		INSERT INTO users (id, email, password, name, phone, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err = r.db.ExecContext(
		ctx,
		query,
		user.ID,
		user.Email,
		user.Password,
		user.Name,
		user.Phone,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// FindByID retrieves a user by their ID (without password)
func (r *userRepository) FindByID(ctx context.Context, id string) (*model.User, error) {
	if id == "" {
		return nil, ErrInvalidUserID
	}

	query := `
		SELECT id, email, name, phone, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	user := &model.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Phone,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user by id: %w", err)
	}

	return user, nil
}

// FindByEmail retrieves a user by their email (includes password for authentication)
// Note: This method returns the password hash for login verification
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	if email == "" {
		return nil, fmt.Errorf("email cannot be empty")
	}

	query := `
		SELECT id, email, password, name, phone, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	user := &model.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Name,
		&user.Phone,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}

	return user, nil
}

// Update updates an existing user's information
// Only updates name and phone (email and password have separate update methods)
func (r *userRepository) Update(ctx context.Context, user *model.User) error {
	if user.ID.String() == "" {
		return ErrInvalidUserID
	}

	query := `
		UPDATE users
		SET name = $1, phone = $2, updated_at = $3
		WHERE id = $4
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		user.Name,
		user.Phone,
		user.UpdatedAt,
		user.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// Delete removes a user from the database by their ID
// This is a hard delete. Consider implementing soft delete in production
func (r *userRepository) Delete(ctx context.Context, id string) error {
	if id == "" {
		return ErrInvalidUserID
	}

	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// List retrieves a paginated list of users (without passwords)
// Ordered by created_at descending (newest first)
func (r *userRepository) List(ctx context.Context, limit, offset int) ([]*model.User, error) {
	// Set default and maximum limits
	if limit <= 0 {
		limit = 10 // default
	}
	if limit > 100 {
		limit = 100 // maximum
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT id, email, name, phone, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	users := make([]*model.User, 0)
	for rows.Next() {
		user := &model.User{}
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Name,
			&user.Phone,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}

// Count returns the total number of users in the database
func (r *userRepository) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM users`

	var count int64
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return count, nil
}

// UpdatePassword updates a user's password
// This is a separate method to ensure password updates are explicit
func (r *userRepository) UpdatePassword(ctx context.Context, userID, hashedPassword string) error {
	if userID == "" {
		return ErrInvalidUserID
	}

	query := `
		UPDATE users
		SET password = $1, updated_at = NOW()
		WHERE id = $2
	`

	result, err := r.db.ExecContext(ctx, query, hashedPassword, userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// EmailExists checks if an email already exists in the database
// Useful for validation before updates
func (r *userRepository) EmailExists(ctx context.Context, email string, excludeUserID string) (bool, error) {
	var query string
	var args []interface{}

	if excludeUserID != "" {
		// Exclude current user when checking (for updates)
		query = `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND id != $2)`
		args = []interface{}{email, excludeUserID}
	} else {
		query = `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
		args = []interface{}{email}
	}

	var exists bool
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}

	return exists, nil
}

// Transaction helper methods

// BeginTx starts a new database transaction
func (r *userRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	return tx, nil
}

// CreateWithTx inserts a new user within a transaction
func (r *userRepository) CreateWithTx(ctx context.Context, tx *sql.Tx, user *model.User) error {
	// Check if email already exists
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	err := tx.QueryRowContext(ctx, checkQuery, user.Email).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check email uniqueness: %w", err)
	}
	if exists {
		return ErrUserAlreadyExists
	}

	// Insert new user
	query := `
		INSERT INTO users (id, email, password, name, phone, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err = tx.ExecContext(
		ctx,
		query,
		user.ID,
		user.Email,
		user.Password,
		user.Name,
		user.Phone,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}
