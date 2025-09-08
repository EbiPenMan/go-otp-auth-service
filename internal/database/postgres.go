package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/ebipenman/go-otp-auth-service/internal/model"

	"github.com/google/uuid"
	"github.com/lib/pq" // PostgreSQL driver
)

// PostgresStore holds the database connection pool.
type PostgresStore struct {
	db *sql.DB
}

// NewPostgresStore creates a new PostgreSQL store, connects to the database,
// and runs initial migrations.
func NewPostgresStore(dataSourceName string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Ping the database to verify the connection is alive.
	if err := db.Ping(); err != nil {
		db.Close() // Close the connection if ping fails
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Successfully connected to PostgreSQL database.")

	store := &PostgresStore{db: db}

	// Run migrations to ensure tables are created.
	if err := store.runMigrations(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return store, nil
}

// runMigrations executes the SQL statements to create the necessary tables if they don't exist.
func (s *PostgresStore) runMigrations() error {
	createUsersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		phone_number VARCHAR(20) UNIQUE NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);`

	// --- THIS IS THE CHANGE ---
	createOTPsTable := `
	CREATE TABLE IF NOT EXISTS otps (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		-- Add the UNIQUE constraint to this column
		phone_number VARCHAR(20) UNIQUE NOT NULL,
		otp_code VARCHAR(6) NOT NULL,
		expires_at TIMESTAMPTZ NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	-- This index is now automatically created by the UNIQUE constraint, but leaving it is fine.
	CREATE INDEX IF NOT EXISTS idx_otps_phone_number ON otps (phone_number);
	`

	_, err := s.db.Exec(createUsersTable)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	_, err = s.db.Exec(createOTPsTable)
	if err != nil {
		return fmt.Errorf("failed to create otps table: %w", err)
	}

	log.Println("Database migrations completed successfully.")
	return nil
}

// --- UserStore Implementation ---

func (s *PostgresStore) CreateUser(user model.User) (model.User, error) {
	query := `
		INSERT INTO users (phone_number)
		VALUES ($1)
		RETURNING id, created_at, updated_at;
	`
	row := s.db.QueryRow(query, user.PhoneNumber)
	err := row.Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		// Check for unique constraint violation
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return model.User{}, fmt.Errorf("%w: user with phone number %s", ErrAlreadyExists, user.PhoneNumber)
		}
		return model.User{}, fmt.Errorf("failed to create user: %w", err)
	}
	return user, nil
}

func (s *PostgresStore) GetUserByID(id uuid.UUID) (model.User, error) {
	var user model.User
	query := `SELECT id, phone_number, created_at, updated_at FROM users WHERE id = $1;`
	row := s.db.QueryRow(query, id)
	err := row.Scan(&user.ID, &user.PhoneNumber, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.User{}, fmt.Errorf("%w: user with ID %s", ErrNotFound, id)
		}
		return model.User{}, fmt.Errorf("failed to get user by ID: %w", err)
	}
	return user, nil
}

func (s *PostgresStore) GetUserByPhoneNumber(phoneNumber string) (model.User, error) {
	var user model.User
	query := `SELECT id, phone_number, created_at, updated_at FROM users WHERE phone_number = $1;`
	row := s.db.QueryRow(query, phoneNumber)
	err := row.Scan(&user.ID, &user.PhoneNumber, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.User{}, fmt.Errorf("%w: user with phone number %s", ErrNotFound, phoneNumber)
		}
		return model.User{}, fmt.Errorf("failed to get user by phone number: %w", err)
	}
	return user, nil
}

func (s *PostgresStore) ListUsers(limit, offset int, search string) ([]model.User, int, error) {
	var users []model.User
	var total int

	// Base query for listing users
	baseQuery := `FROM users`
	var args []interface{}
	argID := 1

	// Add search filter if provided
	if search != "" {
		baseQuery += fmt.Sprintf(" WHERE phone_number LIKE $%d", argID)
		args = append(args, "%"+search+"%")
		argID++
	}

	// Query to get the total count of users matching the filter
	countQuery := `SELECT COUNT(*) ` + baseQuery
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Query to get the paginated list of users
	listQuery := `SELECT id, phone_number, created_at, updated_at ` + baseQuery +
		fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argID, argID+1)
	args = append(args, limit, offset)

	rows, err := s.db.Query(listQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var user model.User
		if err := rows.Scan(&user.ID, &user.PhoneNumber, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan user row: %w", err)
		}
		users = append(users, user)
	}

	return users, total, nil
}

// --- OTPStore Implementation ---

// StoreOTP uses an "UPSERT" operation to either insert a new OTP or update an existing one for a given phone number.
func (s *PostgresStore) StoreOTP(otp model.OTP) error {
	query := `
		INSERT INTO otps (phone_number, otp_code, expires_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (phone_number) DO UPDATE
		SET otp_code = EXCLUDED.otp_code, expires_at = EXCLUDED.expires_at, created_at = NOW();
	`
	_, err := s.db.Exec(query, otp.PhoneNumber, otp.OTPCode, otp.ExpiresAt)
	if err != nil {
		return fmt.Errorf("failed to store OTP: %w", err)
	}
	return nil
}

func (s *PostgresStore) GetOTP(phoneNumber string) (model.OTP, error) {
	var otp model.OTP
	query := `SELECT id, phone_number, otp_code, created_at, expires_at FROM otps WHERE phone_number = $1;`
	row := s.db.QueryRow(query, phoneNumber)
	err := row.Scan(&otp.ID, &otp.PhoneNumber, &otp.OTPCode, &otp.CreatedAt, &otp.ExpiresAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.OTP{}, fmt.Errorf("%w: OTP for phone number %s", ErrNotFound, phoneNumber)
		}
		return model.OTP{}, fmt.Errorf("failed to get OTP: %w", err)
	}
	return otp, nil
}

func (s *PostgresStore) DeleteOTP(phoneNumber string) error {
	query := `DELETE FROM otps WHERE phone_number = $1;`
	_, err := s.db.Exec(query, phoneNumber)
	if err != nil {
		// It's safe to ignore "not found" errors on delete
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("failed to delete OTP: %w", err)
	}
	return nil
}
