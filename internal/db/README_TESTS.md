# Database Tests

This directory contains comprehensive tests for all database functions in `db.go`.

## Test Coverage

The tests cover all database functions:

- ✅ **NewDB()** - Database connection creation with both `DATABASE_URL` and individual environment variables
- ✅ **GetSubscribers()** - Retrieving active subscribers from the database
- ✅ **AddSubscriber()** - Adding new subscribers and handling duplicates
- ✅ **Unsubscribe()** - Marking subscribers as unsubscribed
- ✅ **Close()** - Properly closing database connections

## Test Types

### Unit Tests
- Each function is tested in isolation
- Both success and error scenarios are covered
- Edge cases like duplicate entries, non-existent users, etc.

### Integration Tests
- Full workflow testing combining multiple operations
- End-to-end scenarios mimicking real usage

## Prerequisites

To run these tests, you need:

1. **PostgreSQL database** running locally or accessible remotely
2. **Database permissions** to create and drop test databases
3. **Environment variables** for database connection

## Environment Variables

The tests use these environment variables (with defaults):

```bash
# Option 1: Single DATABASE_URL
DATABASE_URL=postgresql://user:password@host:port/dbname

# Option 2: Individual variables (used if DATABASE_URL is not set)
DB_HOST=localhost        # default: localhost
DB_PORT=5432            # default: 5432
DB_USER=postgres        # default: postgres
DB_PASSWORD=postgres    # default: postgres
DB_NAME=postgres        # for admin operations
```

## Running the Tests

### Run all database tests:
```bash
go test ./internal/db/
```

### Run with verbose output:
```bash
go test -v ./internal/db/
```

### Run a specific test:
```bash
go test -v ./internal/db/ -run TestGetSubscribers_WithData
```

### Run tests with race detection:
```bash
go test -race ./internal/db/
```

## Test Database Setup

The tests automatically:
1. **Create** a temporary test database (`dodgers_win_test`)
2. **Set up** the required `subscribers` table schema
3. **Run** the tests with isolated data
4. **Clean up** by dropping the test database

Each test gets a fresh database to ensure isolation.

## Test Schema

The tests create this table structure:

```sql
CREATE TABLE subscribers (
    id SERIAL PRIMARY KEY,
    phone_number VARCHAR(255) UNIQUE NOT NULL,
    unsubscribed BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW()
);
```

## Troubleshooting

### Tests are skipped
If you see tests being skipped, it means the database connection couldn't be established. Check:
- PostgreSQL is running
- Environment variables are set correctly
- Database user has CREATE/DROP privileges

### Permission errors
The tests need to create and drop databases. Make sure your database user has:
```sql
GRANT CREATE ON DATABASE postgres TO your_user;
-- Or make the user a superuser for testing
ALTER USER your_user WITH SUPERUSER;
```

### Connection timeouts
Tests have built-in timeouts. If you're seeing timeout errors:
- Check if your database is accessible
- Verify network connectivity
- Consider increasing timeout in test setup if needed

## Coverage Report

To see test coverage:
```bash
go test -cover ./internal/db/
```

For detailed coverage report:
```bash
go test -coverprofile=coverage.out ./internal/db/
go tool cover -html=coverage.out
``` 