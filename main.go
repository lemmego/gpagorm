// Package gpagorm provides a GORM adapter for the Go Persistence API (GPA)
package gpagorm

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/lemmego/gpa"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// =====================================
// Provider Implementation
// =====================================

// Provider implements gpa.Provider and gpa.SQLProvider using GORM
type Provider struct {
	db     *gorm.DB
	config gpa.Config
}

// NewProvider creates a new GORM provider instance
func NewProvider(config gpa.Config) (*Provider, error) {
	provider := &Provider{config: config}

	// Configure GORM
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: false,
		},
	}

	// Apply custom configurations from options
	if options, ok := config.Options["gorm"]; ok {
		if gormOpts, ok := options.(map[string]interface{}); ok {
			if logLevel, ok := gormOpts["log_level"].(string); ok {
				switch logLevel {
				case "silent":
					gormConfig.Logger = logger.Default.LogMode(logger.Silent)
				case "error":
					gormConfig.Logger = logger.Default.LogMode(logger.Error)
				case "warn":
					gormConfig.Logger = logger.Default.LogMode(logger.Warn)
				case "info":
					gormConfig.Logger = logger.Default.LogMode(logger.Info)
				}
			}

			if singularTable, ok := gormOpts["singular_table"].(bool); ok {
				gormConfig.NamingStrategy = schema.NamingStrategy{
					SingularTable: singularTable,
				}
			}
		}
	}

	// Initialize database connection
	var dialector gorm.Dialector

	switch strings.ToLower(config.Driver) {
	case "postgres", "postgresql":
		dialector = postgres.Open(buildPostgresDSN(config))
	case "mysql":
		dialector = mysql.Open(buildMySQLDSN(config))
	case "sqlite", "sqlite3":
		dialector = sqlite.Open(config.Database)
	case "sqlserver", "mssql":
		dialector = sqlserver.Open(buildSQLServerDSN(config))
	default:
		return nil, fmt.Errorf("unsupported driver: %s", config.Driver)
	}

	db, err := gorm.Open(dialector, gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if config.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	}
	if config.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	}
	if config.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)
	}
	if config.ConnMaxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(config.ConnMaxIdleTime)
	}

	provider.db = db
	return provider, nil
}

// Configure applies configuration to the provider
func (p *Provider) Configure(config gpa.Config) error {
	p.config = config
	return nil
}

// Health checks the database connection health
func (p *Provider) Health() error {
	sqlDB, err := p.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return sqlDB.PingContext(ctx)
}

// Close closes the database connection
func (p *Provider) Close() error {
	sqlDB, err := p.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// SupportedFeatures returns the list of supported features
func (p *Provider) SupportedFeatures() []gpa.Feature {
	return []gpa.Feature{
		gpa.FeatureTransactions,
		gpa.FeatureJSONQueries,
		gpa.FeatureIndexing,
		gpa.FeatureAggregation,
		gpa.FeatureMigration,
		gpa.FeatureRawSQL,
	}
}

// ProviderInfo returns information about this provider
func (p *Provider) ProviderInfo() gpa.ProviderInfo {
	return gpa.ProviderInfo{
		Name:         "GORM",
		Version:      "1.0.0",
		DatabaseType: gpa.DatabaseTypeSQL,
		Features:     p.SupportedFeatures(),
	}
}

// GetRepository returns a type-safe repository for any entity type T
// If no instanceName provided, uses default instance
// Usage: 
//   userRepo := gpagorm.GetRepository[User]()           // default
//   userRepo := gpagorm.GetRepository[User]("primary")  // named
func GetRepository[T any](instanceName ...string) gpa.Repository[T] {
	provider := gpa.MustGet[*Provider](instanceName...)
	return NewRepository[T](provider.db, provider)
}

// =====================================
// SQLProvider Implementation
// =====================================

// DB returns the underlying database/sql.DB instance
func (p *Provider) DB() interface{} {
	sqlDB, _ := p.db.DB()
	return sqlDB
}

// BeginTx starts a transaction with specific isolation level
func (p *Provider) BeginTx(ctx context.Context, opts *gpa.TxOptions) (interface{}, error) {
	if opts == nil {
		return p.db.WithContext(ctx).Begin(), nil
	}
	
	// Convert GPA isolation level to sql.IsolationLevel
	sqlOpts := &sql.TxOptions{
		ReadOnly: opts.ReadOnly,
	}
	
	switch opts.IsolationLevel {
	case gpa.IsolationReadUncommitted:
		sqlOpts.Isolation = sql.LevelReadUncommitted
	case gpa.IsolationReadCommitted:
		sqlOpts.Isolation = sql.LevelReadCommitted
	case gpa.IsolationRepeatableRead:
		sqlOpts.Isolation = sql.LevelRepeatableRead
	case gpa.IsolationSerializable:
		sqlOpts.Isolation = sql.LevelSerializable
	default:
		sqlOpts.Isolation = sql.LevelDefault
	}
	
	sqlDB, err := p.db.DB()
	if err != nil {
		return nil, err
	}
	
	return sqlDB.BeginTx(ctx, sqlOpts)
}

// Migrate runs database migrations
func (p *Provider) Migrate(models ...interface{}) error {
	return p.db.AutoMigrate(models...)
}

// RawQuery executes raw SQL and returns results
func (p *Provider) RawQuery(ctx context.Context, query string, args ...interface{}) (interface{}, error) {
	var results []map[string]interface{}
	err := p.db.WithContext(ctx).Raw(query, args...).Scan(&results).Error
	return results, err
}

// RawExec executes raw SQL without returning results
func (p *Provider) RawExec(ctx context.Context, query string, args ...interface{}) (gpa.Result, error) {
	result := p.db.WithContext(ctx).Exec(query, args...)
	if result.Error != nil {
		return nil, result.Error
	}
	return &SQLResult{
		rowsAffected: result.RowsAffected,
	}, nil
}


// =====================================
// Helper Functions
// =====================================

// buildPostgresDSN builds a PostgreSQL DSN
func buildPostgresDSN(config gpa.Config) string {
	if config.ConnectionURL != "" {
		return config.ConnectionURL
	}

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s",
		config.Host, config.Port, config.Username, config.Password, config.Database)

	if config.SSL.Enabled {
		dsn += " sslmode=" + config.SSL.Mode
		if config.SSL.CertFile != "" {
			dsn += " sslcert=" + config.SSL.CertFile
		}
		if config.SSL.KeyFile != "" {
			dsn += " sslkey=" + config.SSL.KeyFile
		}
		if config.SSL.CAFile != "" {
			dsn += " sslrootcert=" + config.SSL.CAFile
		}
	} else {
		dsn += " sslmode=disable"
	}

	return dsn
}

// buildMySQLDSN builds a MySQL DSN
func buildMySQLDSN(config gpa.Config) string {
	if config.ConnectionURL != "" {
		return config.ConnectionURL
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.Username, config.Password, config.Host, config.Port, config.Database)

	if config.SSL.Enabled {
		dsn += "&tls=" + config.SSL.Mode
	}

	return dsn
}

// buildSQLServerDSN builds a SQL Server DSN
func buildSQLServerDSN(config gpa.Config) string {
	if config.ConnectionURL != "" {
		return config.ConnectionURL
	}

	return fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s",
		config.Username, config.Password, config.Host, config.Port, config.Database)
}

// SupportedDrivers returns the list of supported database drivers
func SupportedDrivers() []string {
	return []string{"postgres", "postgresql", "mysql", "sqlite", "sqlite3", "sqlserver", "mssql"}
}