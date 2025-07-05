# Logging System

## Overview

Kalycs uses a structured logging system built on top of [Uber's Zap](https://github.com/uber-go/zap) library. The logging system provides high-performance, structured logging with different configurations for development and production environments.

## Architecture

### Core Components

- **Logger Package**: `internal/logging/logging.go`
- **Global Logger**: Singleton instance accessible via `logging.L()`
- **Environment-based Configuration**: Automatic configuration based on `APP_ENV` environment variable

### Logger Initialization

The logger is automatically initialized when first accessed via `logging.L()`. You can also explicitly initialize it using `logging.Init()`.

```go
import "kalycs/internal/logging"

// Implicit initialization (recommended)
logging.L().Info("Application started")

// Explicit initialization (optional)
logging.Init()
logging.L().Info("Application started")
```

## Configuration

### Development Mode
- **Trigger**: `APP_ENV=development`
- **Output**: Human-readable console output
- **Features**: 
  - Colorized output
  - Caller information
  - Stack traces for errors
  - Pretty-printed JSON for structured data

### Production Mode
- **Trigger**: Any value other than "development" or unset `APP_ENV`
- **Output**: Structured JSON logs
- **Features**:
  - ISO8601 timestamps
  - JSON format for log aggregation
  - Performance optimized
  - Suitable for log collectors (ELK, Splunk, etc.)

## Usage Patterns

### Basic Logging

```go
import "kalycs/internal/logging"

// Info level
logging.L().Info("Application started")

// Error level
logging.L().Error("Failed to process request")

// Debug level (only visible in development)
logging.L().Debug("Processing file", "path", "/tmp/file.txt")
```

### Structured Logging

Use the `*w` variants for structured logging with key-value pairs:

```go
// Info with structured data
logging.L().Infow("User logged in", 
    "user_id", userID,
    "ip", request.RemoteAddr,
    "timestamp", time.Now())

// Error with context
logging.L().Errorw("Database connection failed",
    "error", err,
    "database", "postgres",
    "retry_count", 3)

// Fatal with structured data (exits the program)
logging.L().Fatalw("Critical system failure",
    "component", "database",
    "error", err)
```

### Log Levels

The system supports all standard log levels:

- `Debug()` / `Debugw()` - Detailed information for debugging
- `Info()` / `Infow()` - General information about application flow
- `Warn()` / `Warnw()` - Warning conditions that should be addressed
- `Error()` / `Errorw()` - Error conditions that need attention
- `Fatal()` / `Fatalw()` - Critical errors that cause program termination

## Current Usage in Codebase

### Application Lifecycle
- **main.go**: Application startup failure logging
- **app.go:33**: Application startup
- **app.go:38,46,49,54,59**: Fatal errors during initialization
- **app.go:72**: Application shutdown
- **app.go:80,89,93,95**: File import and classification logging

### Database Operations
- **db/db.go**: Database initialization, connection management, and table creation
- **internal/database/transaction.go**: Transaction rollback error logging
- **internal/store/project_repo.go**: Project CRUD operations with detailed logging
- **internal/store/rule_repo.go**: Rule CRUD operations with detailed logging
- **internal/store/file_repo.go**: File upsert and project assignment operations

### File System Operations
- **internal/utils/downloads.go**: Downloads directory detection and OS-specific paths
- **internal/watcher/watcher.go**: File system watching events and file classification
- **internal/classifier/classifier.go**: File classification decisions and rule matching

### Validation and Error Handling
- **internal/validation/validator.go**: Validation failures for projects and rules
- **internal/store/project_repo.go**: Validation and database constraint errors
- **internal/store/rule_repo.go**: Validation and database constraint errors

### File Classification
- **internal/classifier/classifier.go**: Rule compilation, file classification decisions
- **internal/watcher/watcher.go**: File system events and classification triggers
- **app.go**: Bulk file import operations

## Best Practices

### 1. Use Structured Logging
```go
// Good
logging.L().Infow("File processed", 
    "file_path", path,
    "size_bytes", fileSize,
    "duration_ms", processingTime)

// Avoid
logging.L().Info("File processed: " + path)
```

### 2. Include Context Information
```go
// Good
logging.L().Errorw("Failed to save file",
    "error", err,
    "file_path", path,
    "user_id", userID,
    "operation", "save")

// Minimal
logging.L().Error("Failed to save file")
```

### 3. Use Appropriate Log Levels
- **Debug**: Detailed tracing information (validation details, rule compilation)
- **Info**: Normal application flow (file classification, CRUD operations)
- **Warn**: Recoverable issues (validation failures, constraint violations)
- **Error**: Error conditions that need attention (database errors, file system errors)
- **Fatal**: Critical errors requiring immediate shutdown (startup failures)

### 4. Avoid Logging Sensitive Data
```go
// Good
logging.L().Infow("User authenticated", "user_id", userID)

// Avoid
logging.L().Infow("User authenticated", "password", password)
```

## Performance Considerations

- Zap is designed for high performance with minimal allocations
- Use structured logging (`*w` methods) for better performance than string formatting
- Debug logs have minimal overhead in production builds
- JSON output in production is optimized for machine parsing

## Environment Setup

### Development
```bash
export APP_ENV=development
./kalycs
```

### Production
```bash
# APP_ENV defaults to production mode
./kalycs
```

## Logging Coverage Summary

### ✅ **Fully Implemented**
- **Application Lifecycle**: Startup, shutdown, and critical errors
- **Database Operations**: Initialization, transactions, CRUD operations
- **File System**: Watcher events, file classification, downloads directory
- **Validation**: Project and rule validation failures
- **Error Handling**: Database constraints, file system errors, validation errors

### 📊 **Logging Statistics**
- **Files Updated**: 8 core files
- **Log Statements Added**: 50+ structured log entries
- **Coverage Areas**: Application lifecycle, database, file system, validation
- **Log Levels Used**: Debug, Info, Warn, Error, Fatal

### 🔍 **Key Logging Points**
1. **Database Initialization**: Connection setup, table creation, permission issues
2. **CRUD Operations**: Project and rule creation, updates, deletions with context
3. **File Classification**: Rule matching, project assignment, classification decisions
4. **File System Events**: Watcher events, file creation, classification triggers
5. **Validation**: Input validation failures with detailed error context
6. **Error Recovery**: Transaction rollbacks, constraint violations, file system errors

## Future Improvements

1. **Log Rotation**: Implement log rotation for production deployments
2. **Centralized Logging**: Add support for log aggregation services
3. **Metrics Integration**: Add structured metrics alongside logs
4. **Context Propagation**: Implement request/operation context tracing
5. **Performance Monitoring**: Add timing logs for slow operations