// Package gpagorm provides structured logging for entity hooks
package gpagorm

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"runtime"
)

// HookLogger provides structured logging for entity hook errors
type HookLogger struct {
	logger *slog.Logger
}

// NewHookLogger creates a new hook logger
func NewHookLogger(logger *slog.Logger) *HookLogger {
	if logger == nil {
		logger = slog.Default()
	}
	return &HookLogger{logger: logger}
}

// LogHookError logs an error that occurred during hook execution
func (h *HookLogger) LogHookError(ctx context.Context, entity interface{}, hookType string, hookName string, err error) {
	if err == nil {
		return
	}

	// Get caller information
	pc, file, line, _ := runtime.Caller(2)
	caller := runtime.FuncForPC(pc)
	funcName := "unknown"
	if caller != nil {
		funcName = caller.Name()
	}

	// Get entity type name
	entityType := "unknown"
	if entity != nil {
		t := reflect.TypeOf(entity)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		entityType = t.Name()
	}

	h.logger.LogAttrs(ctx, slog.LevelError,
		"entity hook failed",
		slog.String("hook_type", hookType),
		slog.String("hook_name", hookName),
		slog.String("entity_type", entityType),
		slog.String("error", err.Error()),
		slog.String("caller", funcName),
		slog.String("file", file),
		slog.Int("line", line),
	)
}

// LogValidationError logs a validation error
func (h *HookLogger) LogValidationError(ctx context.Context, entity interface{}, err error) {
	if err == nil {
		return
	}

	// Get entity type name
	entityType := "unknown"
	if entity != nil {
		t := reflect.TypeOf(entity)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		entityType = t.Name()
	}

	h.logger.LogAttrs(ctx, slog.LevelWarn,
		"entity validation failed",
		slog.String("entity_type", entityType),
		slog.String("error", err.Error()),
	)
}

// LogHookSuccess logs successful hook execution (debug level)
func (h *HookLogger) LogHookSuccess(ctx context.Context, entity interface{}, hookType string, hookName string) {
	if !h.logger.Enabled(ctx, slog.LevelDebug) {
		return
	}

	// Get entity type name
	entityType := "unknown"
	if entity != nil {
		t := reflect.TypeOf(entity)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		entityType = t.Name()
	}

	h.logger.LogAttrs(ctx, slog.LevelDebug,
		"entity hook executed",
		slog.String("hook_type", hookType),
		slog.String("hook_name", hookName),
		slog.String("entity_type", entityType),
	)
}

// DefaultHookLogger is the default hook logger instance
var DefaultHookLogger = NewHookLogger(nil)

// Helper functions for quick logging without a HookLogger instance

// LogAfterCreateError logs an error from AfterCreate hook
func LogAfterCreateError(ctx context.Context, entity interface{}, err error) {
	DefaultHookLogger.LogHookError(ctx, entity, "AfterCreate", "AfterCreate", err)
}

// LogAfterUpdateError logs an error from AfterUpdate hook
func LogAfterUpdateError(ctx context.Context, entity interface{}, err error) {
	DefaultHookLogger.LogHookError(ctx, entity, "AfterUpdate", "AfterUpdate", err)
}

// LogAfterDeleteError logs an error from AfterDelete hook
func LogAfterDeleteError(ctx context.Context, entity interface{}, err error) {
	DefaultHookLogger.LogHookError(ctx, entity, "AfterDelete", "AfterDelete", err)
}

// LogAfterFindError logs an error from AfterFind hook
func LogAfterFindError(ctx context.Context, entity interface{}, err error) {
	DefaultHookLogger.LogHookError(ctx, entity, "AfterFind", "AfterFind", err)
}

// LogBeforeCreateError logs an error from BeforeCreate hook
func LogBeforeCreateError(ctx context.Context, entity interface{}, err error) {
	DefaultHookLogger.LogHookError(ctx, entity, "BeforeCreate", "BeforeCreate", err)
}

// LogBeforeUpdateError logs an error from BeforeUpdate hook
func LogBeforeUpdateError(ctx context.Context, entity interface{}, err error) {
	DefaultHookLogger.LogHookError(ctx, entity, "BeforeUpdate", "BeforeUpdate", err)
}

// LogBeforeDeleteError logs an error from BeforeDelete hook
func LogBeforeDeleteError(ctx context.Context, entity interface{}, err error) {
	DefaultHookLogger.LogHookError(ctx, entity, "BeforeDelete", "BeforeDelete", err)
}

// LogValidationError logs a validation error
func LogValidationError(ctx context.Context, entity interface{}, err error) {
	DefaultHookLogger.LogValidationError(ctx, entity, err)
}

// HookError represents an error that occurred during hook execution
// It wraps the original error with context about which hook failed
type HookError struct {
	HookType   string // e.g., "AfterCreate", "BeforeUpdate"
	EntityType string // e.g., "User", "Post"
	Err        error  // The original error
}

// Error returns the error message
func (e *HookError) Error() string {
	return fmt.Sprintf("%s hook failed for %s: %v", e.HookType, e.EntityType, e.Err)
}

// Unwrap returns the underlying error
func (e *HookError) Unwrap() error {
	return e.Err
}

// NewHookError creates a new HookError
func NewHookError(hookType, entityType string, err error) *HookError {
	return &HookError{
		HookType:   hookType,
		EntityType: entityType,
		Err:        err,
	}
}
