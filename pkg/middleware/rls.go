// Package middleware provides HTTP middleware for RLS (Row-Level Security)
package middleware

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

// RLSContext holds RLS session variables for PostgreSQL
type RLSContext struct {
	UserID     uuid.UUID
	UserRole   string
	UserIP     string
	UserAgent  string
	SystemRead bool // Allow system to read all data (e.g., assignment algorithm)
}

// SetRLSVariables sets PostgreSQL session variables for Row-Level Security
// Call this before any query that should be restricted by RLS policies
func SetRLSVariables(ctx context.Context, db *sql.DB, rlsCtx *RLSContext) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil && rollbackErr != sql.ErrTxDone {
			// Log rollback error but don't fail - in production, use proper logger
			_ = rollbackErr
		}
	}()

	// Set user_id
	if _, err := tx.ExecContext(ctx, "SET LOCAL app.user_id = $1", rlsCtx.UserID.String()); err != nil {
		return fmt.Errorf("failed to set app.user_id: %w", err)
	}

	// Set user_role
	if _, err := tx.ExecContext(ctx, "SET LOCAL app.user_role = $1", rlsCtx.UserRole); err != nil {
		return fmt.Errorf("failed to set app.user_role: %w", err)
	}

	// Set user_ip (optional)
	if rlsCtx.UserIP != "" {
		if _, err := tx.ExecContext(ctx, "SET LOCAL app.user_ip = $1", rlsCtx.UserIP); err != nil {
			return fmt.Errorf("failed to set app.user_ip: %w", err)
		}
	}

	// Set user_agent (optional)
	if rlsCtx.UserAgent != "" {
		if _, err := tx.ExecContext(ctx, "SET LOCAL app.user_agent = $1", rlsCtx.UserAgent); err != nil {
			return fmt.Errorf("failed to set app.user_agent: %w", err)
		}
	}

	// Set system_read flag (for system operations that need to bypass RLS)
	if rlsCtx.SystemRead {
		if _, err := tx.ExecContext(ctx, "SET LOCAL app.system_read = 'true'"); err != nil {
			return fmt.Errorf("failed to set app.system_read: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// WithRLS middleware extracts user context and prepares RLS variables
// It adds RLSContext to the request context for use in handlers
func WithRLS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract user info from context (set by RequireAuth middleware)
		userID, ok := r.Context().Value(UserIDKey).(uuid.UUID)
		if !ok {
			// If no user context, this is a public endpoint - skip RLS
			next.ServeHTTP(w, r)
			return
		}

		userRole, _ := r.Context().Value(UserRoleKey).(string)

		// Build RLS context
		rlsCtx := &RLSContext{
			UserID:    userID,
			UserRole:  userRole,
			UserIP:    getClientIP(r),
			UserAgent: r.UserAgent(),
		}

		// Add RLS context to request context
		ctx := context.WithValue(r.Context(), contextKey("rls_context"), rlsCtx)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRLSContext retrieves RLS context from request context
func GetRLSContext(ctx context.Context) (*RLSContext, bool) {
	rlsCtx, ok := ctx.Value(contextKey("rls_context")).(*RLSContext)
	return rlsCtx, ok
}

// getClientIP extracts the real client IP from headers (considering proxies)
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (set by proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the list (original client)
		return xff
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fallback to RemoteAddr
	return r.RemoteAddr
}

// ExecuteWithRLS executes a query with RLS variables set
// Use this helper in repositories to automatically apply RLS
func ExecuteWithRLS(ctx context.Context, db *sql.DB, fn func(*sql.Tx) error) error {
	// Get RLS context from request context
	rlsCtx, ok := GetRLSContext(ctx)
	if !ok {
		// No RLS context - execute without RLS (system operation)
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer func() {
			if rollbackErr := tx.Rollback(); rollbackErr != nil && rollbackErr != sql.ErrTxDone {
				// Log error - in production, use proper logger
				_ = rollbackErr
			}
		}()

		if err := fn(tx); err != nil {
			return err
		}

		return tx.Commit()
	}

	// Begin transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil && rollbackErr != sql.ErrTxDone {
			// Log error - in production, use proper logger
			_ = rollbackErr
		}
	}()

	// Set RLS variables
	if _, err := tx.ExecContext(ctx, "SET LOCAL app.user_id = $1", rlsCtx.UserID.String()); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "SET LOCAL app.user_role = $1", rlsCtx.UserRole); err != nil {
		return err
	}
	if rlsCtx.UserIP != "" {
		if _, err := tx.ExecContext(ctx, "SET LOCAL app.user_ip = $1", rlsCtx.UserIP); err != nil {
			return err
		}
	}
	if rlsCtx.UserAgent != "" {
		if _, err := tx.ExecContext(ctx, "SET LOCAL app.user_agent = $1", rlsCtx.UserAgent); err != nil {
			return err
		}
	}
	if rlsCtx.SystemRead {
		if _, err := tx.ExecContext(ctx, "SET LOCAL app.system_read = 'true'"); err != nil {
			return err
		}
	}

	// Execute user function
	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit()
}
