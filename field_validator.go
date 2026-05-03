// Package gpagorm provides field name validation to prevent SQL injection
package gpagorm

import (
	"regexp"
	"sync"
)

var (
	// safeFieldPattern matches valid SQL field names
	// - Must start with a letter or underscore
	// - Can contain letters, numbers, underscores, and dots (for table.column syntax)
	safeFieldPattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_.]*$`)
	fieldCache       = make(map[string]bool)
	fieldCacheMu     sync.RWMutex
)

// isValidFieldName validates that a field name is safe to use in SQL queries.
// It checks against a whitelist pattern to prevent SQL injection through
// malicious field names.
func isValidFieldName(field string) bool {
	// Check cache first (read lock)
	fieldCacheMu.RLock()
	if cached, ok := fieldCache[field]; ok {
		fieldCacheMu.RUnlock()
		return cached
	}
	fieldCacheMu.RUnlock()

	// Validate against pattern
	valid := safeFieldPattern.MatchString(field)

	// Cache the result (write lock)
	fieldCacheMu.Lock()
	fieldCache[field] = valid
	fieldCacheMu.Unlock()

	return valid
}

// isValidTableName validates that a table name is safe to use in SQL queries.
// Table names follow the same rules as field names.
func isValidTableName(table string) bool {
	return isValidFieldName(table)
}

// validateFieldName validates a field name and returns an error if invalid.
// This is useful for providing detailed error messages.
func validateFieldName(field string) error {
	if !isValidFieldName(field) {
		return &FieldValidationError{
			Field: field,
			Reason: "field name contains invalid characters or doesn't follow naming rules",
		}
	}
	return nil
}

// FieldValidationError represents an error that occurs when a field name
// fails validation.
type FieldValidationError struct {
	Field  string
	Reason string
}

// Error returns the error message for FieldValidationError.
func (e *FieldValidationError) Error() string {
	return e.Reason + ": " + e.Field
}

// Type ensures FieldValidationError implements the error interface.
var _ error = (*FieldValidationError)(nil)
