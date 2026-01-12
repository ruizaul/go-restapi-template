package repositories

import (
	"database/sql"
	"errors"

	"tacoshare-delivery-api/internal/auth/models"

	"github.com/google/uuid"
)

// UserRepository handles data access for users in authentication context
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository for authentication
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser creates a new user in the database
func (r *UserRepository) CreateUser(user *models.User) error {
	query := `
		INSERT INTO users (
			id, name, email, phone, password_hash, role,
			first_name, last_name, mother_last_name, birth_date,
			phone_verified, account_status, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`
	_, err := r.db.Exec(
		query,
		user.ID,
		user.Name,
		user.Email,
		user.Phone,
		user.PasswordHash,
		user.Role,
		user.FirstName,
		user.LastName,
		user.MotherLastName,
		user.BirthDate,
		user.PhoneVerified,
		user.AccountStatus,
		user.CreatedAt,
		user.UpdatedAt,
	)
	return err
}

// FindByEmail finds a user by email
func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	query := `
		SELECT id, name, email, phone, password_hash, role,
			first_name, last_name, mother_last_name, birth_date,
			phone_verified, account_status, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	user := &models.User{}
	var phone sql.NullString
	var firstName, lastName, motherLastName sql.NullString
	var birthDate sql.NullTime

	err := r.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&phone,
		&user.PasswordHash,
		&user.Role,
		&firstName,
		&lastName,
		&motherLastName,
		&birthDate,
		&user.PhoneVerified,
		&user.AccountStatus,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if phone.Valid {
		user.Phone = phone.String
	}
	if firstName.Valid {
		user.FirstName = firstName.String
	}
	if lastName.Valid {
		user.LastName = lastName.String
	}
	if motherLastName.Valid {
		user.MotherLastName = motherLastName.String
	}
	if birthDate.Valid {
		user.BirthDate = &birthDate.Time
	}

	return user, nil
}

// FindByID finds a user by ID
func (r *UserRepository) FindByID(id uuid.UUID) (*models.User, error) {
	query := `
		SELECT id, name, email, phone, password_hash, role,
			first_name, last_name, mother_last_name, birth_date,
			phone_verified, account_status, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	user := &models.User{}
	var phone sql.NullString
	var firstName, lastName, motherLastName sql.NullString
	var birthDate sql.NullTime

	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&phone,
		&user.PasswordHash,
		&user.Role,
		&firstName,
		&lastName,
		&motherLastName,
		&birthDate,
		&user.PhoneVerified,
		&user.AccountStatus,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if phone.Valid {
		user.Phone = phone.String
	}
	if firstName.Valid {
		user.FirstName = firstName.String
	}
	if lastName.Valid {
		user.LastName = lastName.String
	}
	if motherLastName.Valid {
		user.MotherLastName = motherLastName.String
	}
	if birthDate.Valid {
		user.BirthDate = &birthDate.Time
	}

	return user, nil
}

// EmailExists checks if an email already exists
func (r *UserRepository) EmailExists(email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	var exists bool
	err := r.db.QueryRow(query, email).Scan(&exists)
	return exists, err
}

// PhoneExists checks if a phone number already exists
func (r *UserRepository) PhoneExists(phone string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE phone = $1)`
	var exists bool
	err := r.db.QueryRow(query, phone).Scan(&exists)
	return exists, err
}

// SaveOTP stores OTP code and expiration for a phone number
func (r *UserRepository) SaveOTP(phone, otpCode string, expiresAt sql.NullTime) error {
	query := `
		UPDATE users
		SET otp_code = $1, otp_expires_at = $2, updated_at = NOW()
		WHERE phone = $3
	`
	_, err := r.db.Exec(query, otpCode, expiresAt, phone)
	return err
}

// FindByPhoneWithOTP finds a user by phone and returns OTP data
func (r *UserRepository) FindByPhoneWithOTP(phone string) (*models.User, error) {
	query := `
		SELECT id, name, email, phone, password_hash, role,
			first_name, last_name, mother_last_name, birth_date,
			phone_verified, account_status, otp_code, otp_expires_at,
			created_at, updated_at
		FROM users
		WHERE phone = $1
	`

	user := &models.User{}
	var name, email, passwordHash, role sql.NullString
	var firstName, lastName, motherLastName, otpCode sql.NullString
	var birthDate, otpExpiresAt sql.NullTime

	err := r.db.QueryRow(query, phone).Scan(
		&user.ID,
		&name,
		&email,
		&user.Phone,
		&passwordHash,
		&role,
		&firstName,
		&lastName,
		&motherLastName,
		&birthDate,
		&user.PhoneVerified,
		&user.AccountStatus,
		&otpCode,
		&otpExpiresAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if name.Valid {
		user.Name = name.String
	}
	if email.Valid {
		user.Email = email.String
	}
	if passwordHash.Valid {
		user.PasswordHash = passwordHash.String
	}
	if role.Valid {
		user.Role = role.String
	}
	if firstName.Valid {
		user.FirstName = firstName.String
	}
	if lastName.Valid {
		user.LastName = lastName.String
	}
	if motherLastName.Valid {
		user.MotherLastName = motherLastName.String
	}
	if birthDate.Valid {
		user.BirthDate = &birthDate.Time
	}
	if otpCode.Valid {
		user.OTPCode = otpCode.String
	}
	if otpExpiresAt.Valid {
		user.OTPExpiresAt = &otpExpiresAt.Time
	}

	return user, nil
}

// MarkPhoneAsVerified marks a phone number as verified
func (r *UserRepository) MarkPhoneAsVerified(phone string) error {
	query := `
		UPDATE users
		SET phone_verified = TRUE, otp_code = NULL, otp_expires_at = NULL, updated_at = NOW()
		WHERE phone = $1
	`
	_, err := r.db.Exec(query, phone)
	return err
}

// CreatePendingUser creates a user with pending status (only phone, no email/password yet)
func (r *UserRepository) CreatePendingUser(phone, otpCode string, expiresAt sql.NullTime) error {
	query := `
		INSERT INTO users (
			id, phone, otp_code, otp_expires_at,
			phone_verified, account_status, role,
			name, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, FALSE, 'pending', 'customer', '', NOW(), NOW())
		ON CONFLICT (phone) DO UPDATE
		SET otp_code = EXCLUDED.otp_code,
			otp_expires_at = EXCLUDED.otp_expires_at,
			updated_at = NOW()
	`
	_, err := r.db.Exec(query, uuid.New(), phone, otpCode, expiresAt)
	return err
}

// CompleteRegistration completes user registration after OTP verification
func (r *UserRepository) CompleteRegistration(user *models.User) error {
	query := `
		UPDATE users
		SET first_name = $1, last_name = $2, mother_last_name = $3, birth_date = $4,
			email = $5, password_hash = $6, name = $7, role = $8,
			account_status = 'active', updated_at = NOW()
		WHERE phone = $9 AND phone_verified = TRUE
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(
		query,
		user.FirstName,
		user.LastName,
		user.MotherLastName,
		user.BirthDate,
		user.Email,
		user.PasswordHash,
		user.Name,
		user.Role,
		user.Phone,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	return err
}

// SaveOTPHash stores OTP hash (SHA-256) and expiration for a phone number
// This replaces SaveOTP to use secure hash instead of plaintext
func (r *UserRepository) SaveOTPHash(phone, otpHash string, expiresAt sql.NullTime) error {
	query := `
		UPDATE users
		SET otp_hash = $1, otp_expires_at = $2, otp_attempts = 0, otp_locked_until = NULL, updated_at = NOW()
		WHERE phone = $3
	`
	_, err := r.db.Exec(query, otpHash, expiresAt, phone)
	return err
}

// CreatePendingUserWithHash creates a user with pending status using OTP hash
func (r *UserRepository) CreatePendingUserWithHash(phone, otpHash string, expiresAt sql.NullTime) error {
	query := `
		INSERT INTO users (
			id, phone, otp_hash, otp_expires_at, otp_attempts,
			phone_verified, account_status, role,
			name, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, 0, FALSE, 'pending', 'customer', '', NOW(), NOW())
		ON CONFLICT (phone) DO UPDATE
		SET otp_hash = EXCLUDED.otp_hash,
			otp_expires_at = EXCLUDED.otp_expires_at,
			otp_attempts = 0,
			otp_locked_until = NULL,
			updated_at = NOW()
	`
	_, err := r.db.Exec(query, uuid.New(), phone, otpHash, expiresAt)
	return err
}

// FindByPhoneWithOTPHash finds a user by phone and returns OTP hash data
func (r *UserRepository) FindByPhoneWithOTPHash(phone string) (*models.User, error) {
	query := `
		SELECT id, name, email, phone, password_hash, role,
			first_name, last_name, mother_last_name, birth_date,
			phone_verified, account_status, otp_hash, otp_expires_at,
			otp_attempts, otp_locked_until, created_at, updated_at
		FROM users
		WHERE phone = $1
	`

	user := &models.User{}
	var name, email, passwordHash, role sql.NullString
	var firstName, lastName, motherLastName, otpHash sql.NullString
	var birthDate, otpExpiresAt, otpLockedUntil sql.NullTime
	var otpAttempts sql.NullInt32

	err := r.db.QueryRow(query, phone).Scan(
		&user.ID,
		&name,
		&email,
		&user.Phone,
		&passwordHash,
		&role,
		&firstName,
		&lastName,
		&motherLastName,
		&birthDate,
		&user.PhoneVerified,
		&user.AccountStatus,
		&otpHash,
		&otpExpiresAt,
		&otpAttempts,
		&otpLockedUntil,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if name.Valid {
		user.Name = name.String
	}
	if email.Valid {
		user.Email = email.String
	}
	if passwordHash.Valid {
		user.PasswordHash = passwordHash.String
	}
	if role.Valid {
		user.Role = role.String
	}
	if firstName.Valid {
		user.FirstName = firstName.String
	}
	if lastName.Valid {
		user.LastName = lastName.String
	}
	if motherLastName.Valid {
		user.MotherLastName = motherLastName.String
	}
	if birthDate.Valid {
		user.BirthDate = &birthDate.Time
	}
	if otpHash.Valid {
		user.OTPHash = otpHash.String
	}
	if otpExpiresAt.Valid {
		user.OTPExpiresAt = &otpExpiresAt.Time
	}
	if otpAttempts.Valid {
		user.OTPAttempts = int(otpAttempts.Int32)
	}
	if otpLockedUntil.Valid {
		user.OTPLockedUntil = &otpLockedUntil.Time
	}

	return user, nil
}

// IncrementOTPAttempts increments the OTP verification attempt counter
func (r *UserRepository) IncrementOTPAttempts(phone string) error {
	query := `
		UPDATE users
		SET otp_attempts = otp_attempts + 1, updated_at = NOW()
		WHERE phone = $1
	`
	_, err := r.db.Exec(query, phone)
	return err
}

// LockOTPAccount locks the account for OTP verification for specified duration
func (r *UserRepository) LockOTPAccount(phone string, lockedUntil sql.NullTime) error {
	query := `
		UPDATE users
		SET otp_locked_until = $1, updated_at = NOW()
		WHERE phone = $2
	`
	_, err := r.db.Exec(query, lockedUntil, phone)
	return err
}

// ClearOTPData clears OTP hash and resets attempts after successful verification
func (r *UserRepository) ClearOTPData(phone string) error {
	query := `
		UPDATE users
		SET otp_hash = NULL, otp_expires_at = NULL, otp_attempts = 0,
		    otp_locked_until = NULL, phone_verified = TRUE, updated_at = NOW()
		WHERE phone = $1
	`
	_, err := r.db.Exec(query, phone)
	return err
}
