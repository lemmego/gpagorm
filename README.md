# GPAGorm - GORM Adapter for Go Persistence API

A type-safe GORM adapter implementation for the Go Persistence API (GPA), providing compile-time type safety and unified database operations across multiple SQL databases.

## Features

- **Type-Safe Operations**: Leverages Go generics for compile-time type safety
- **Multi-Database Support**: PostgreSQL, MySQL, SQLite, and SQL Server
- **Unified API**: Consistent interface across different database drivers
- **Transaction Support**: Safe transaction handling with automatic rollback
- **Migration Support**: Schema migration and table management
- **Raw SQL Support**: Direct SQL execution when needed
- **Relationship Support**: Preload relationships with type safety
- **Connection Pooling**: Configurable connection pool settings

## Supported Databases

- **PostgreSQL** (`postgres`, `postgresql`)
- **MySQL** (`mysql`)
- **SQLite** (`sqlite`, `sqlite3`)
- **SQL Server** (`sqlserver`, `mssql`)

## Installation

```bash
go get github.com/lemmego/gpagorm
```

## Quick Start

```go
package main

import (
    "context"
    "log"

    "github.com/lemmego/gpa"
    "github.com/lemmego/gpagorm"
)

type User struct {
    ID    uint   `gorm:"primaryKey"`
    Name  string `gorm:"size:100"`
    Email string `gorm:"uniqueIndex"`
}

func main() {
    // Configure database connection
    config := gpa.Config{
        Driver:   "postgres",
        Host:     "localhost",
        Port:     5432,
        Database: "myapp",
        Username: "user",
        Password: "password",
    }

    // Create provider
    provider, err := gpagorm.NewProvider(config)
    if err != nil {
        log.Fatal(err)
    }
    defer provider.Close()

    // Get type-safe repository
    userRepo := gpagorm.GetRepository[User](provider)

    // Create user
    user := &User{Name: "John Doe", Email: "john@example.com"}
    err = userRepo.Create(context.Background(), user)
    if err != nil {
        log.Fatal(err)
    }

    // Find user by ID
    foundUser, err := userRepo.FindByID(context.Background(), user.ID)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Found user: %+v", foundUser)
}
```

## Configuration

### Basic Configuration

```go
config := gpa.Config{
    Driver:   "postgres",
    Host:     "localhost",
    Port:     5432,
    Database: "myapp",
    Username: "user",
    Password: "password",
}
```

### Advanced Configuration

```go
config := gpa.Config{
    Driver:          "postgres",
    Host:            "localhost",
    Port:            5432,
    Database:        "myapp",
    Username:        "user",
    Password:        "password",
    MaxOpenConns:    25,
    MaxIdleConns:    5,
    ConnMaxLifetime: time.Hour,
    ConnMaxIdleTime: time.Minute * 30,
    SSL: gpa.SSLConfig{
        Enabled:  true,
        Mode:     "require",
        CertFile: "/path/to/cert.pem",
        KeyFile:  "/path/to/key.pem",
        CAFile:   "/path/to/ca.pem",
    },
    Options: map[string]interface{}{
        "gorm": map[string]interface{}{
            "log_level":      "info",
            "singular_table": false,
        },
    },
}
```

## API Reference

### Repository Operations

```go
// Create operations
err := repo.Create(ctx, entity)
err := repo.CreateBatch(ctx, entities)

// Read operations
entity, err := repo.FindByID(ctx, id)
entities, err := repo.FindAll(ctx, opts...)
entities, err := repo.Query(ctx, opts...)
entity, err := repo.QueryOne(ctx, opts...)

// Update operations
err := repo.Update(ctx, entity)
err := repo.UpdatePartial(ctx, id, updates)

// Delete operations
err := repo.Delete(ctx, id)
err := repo.DeleteByCondition(ctx, condition)

// Aggregation
count, err := repo.Count(ctx, opts...)
exists, err := repo.Exists(ctx, opts...)
```

### Query Options

```go
// Filtering
entities, err := repo.Query(ctx,
    gpa.Where("name", gpa.OpEqual, "John"),
    gpa.Where("age", gpa.OpGreaterThan, 18),
)

// Ordering
entities, err := repo.Query(ctx,
    gpa.OrderBy("name", gpa.ASC),
    gpa.OrderBy("created_at", gpa.DESC),
)

// Pagination
entities, err := repo.Query(ctx,
    gpa.Limit(10),
    gpa.Offset(20),
)

// Field selection
entities, err := repo.Query(ctx,
    gpa.Select("id", "name", "email"),
)

// Preloading relationships
entities, err := repo.FindWithRelations(ctx, []string{"Profile", "Orders"})
```

### Transactions

```go
err := repo.Transaction(ctx, func(txRepo gpa.Transaction[User]) error {
    // All operations within this function are part of the transaction
    user := &User{Name: "Jane", Email: "jane@example.com"}
    if err := txRepo.Create(ctx, user); err != nil {
        return err // Transaction will be rolled back
    }

    // More operations...
    return nil // Transaction will be committed
})
```

### Raw SQL

```go
// Raw query
users, err := repo.RawQuery(ctx, "SELECT * FROM users WHERE age > ?", []interface{}{18})

// Raw execution
result, err := repo.RawExec(ctx, "UPDATE users SET status = ? WHERE active = ?", []interface{}{"verified", true})
```

### Schema Management

```go
// Create table
err := repo.CreateTable(ctx)

// Migrate table
err := repo.MigrateTable(ctx)

// Drop table
err := repo.DropTable(ctx)

// Create index
err := repo.CreateIndex(ctx, []string{"email"}, true) // unique index

// Get table info
info, err := repo.GetTableInfo(ctx)
```

## Error Handling

GPAGorm converts GORM errors to GPA errors for consistent error handling:

```go
user, err := repo.FindByID(ctx, 999)
if err != nil {
    if gpaErr, ok := err.(gpa.GPAError); ok {
        switch gpaErr.Type {
        case gpa.ErrorTypeNotFound:
            log.Println("User not found")
        case gpa.ErrorTypeDuplicate:
            log.Println("Duplicate key violation")
        case gpa.ErrorTypeDatabase:
            log.Println("Database error:", gpaErr.Cause)
        }
    }
}
```

## Testing

Run the test suite:

```bash
go test -v
```

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details.

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Dependencies

- [GORM](https://gorm.io/) - The fantastic ORM library for Golang
- [Go Persistence API](https://github.com/lemmego/gpa) - Unified persistence layer for Go
- Database drivers for PostgreSQL, MySQL, SQLite, and SQL Server
