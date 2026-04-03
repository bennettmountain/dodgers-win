package db

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testDBName = "dodgers_win_test"
)

// setupTestDB creates a test database and returns a cleanup function
func setupTestDB(t *testing.T) (Database, func()) {
	ctx := context.Background()

	// Get database connection details from environment
	host := getEnvOrDefault("DB_HOST", "localhost")
	port := getEnvOrDefault("DB_PORT", "5432")
	user := getEnvOrDefault("DB_USER", "postgres")
	password := getEnvOrDefault("DB_PASSWORD", "postgres")

	// Connect to postgres database to create test database
	adminURL := fmt.Sprintf("postgresql://%s:%s@%s:%s/postgres", user, password, host, port)
	adminConn, err := pgx.Connect(ctx, adminURL)
	if err != nil {
		t.Skipf("Unable to connect to PostgreSQL for testing: %v", err)
	}

	// Create test database
	_, err = adminConn.Exec(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s", testDBName))
	require.NoError(t, err)

	_, err = adminConn.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s", testDBName))
	require.NoError(t, err)

	adminConn.Close(ctx)

	// Connect to test database
	testURL := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", user, password, host, port, testDBName)
	testConn, err := pgx.Connect(ctx, testURL)
	require.NoError(t, err)

	// Create subscribers table
	_, err = testConn.Exec(ctx, `
		CREATE TABLE subscribers (
			id SERIAL PRIMARY KEY,
			phone_number VARCHAR(255) UNIQUE NOT NULL,
			unsubscribed BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT NOW()
		)
	`)
	require.NoError(t, err)

	db := Database{conn: testConn}

	cleanup := func() {
		testConn.Close(ctx)
		// Reconnect to postgres to drop test database
		adminConn, err := pgx.Connect(ctx, adminURL)
		if err == nil {
			adminConn.Exec(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s", testDBName))
			adminConn.Close(ctx)
		}
	}

	return db, cleanup
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func TestNewDB_WithDatabaseURL(t *testing.T) {
	// Skip if no database available
	if os.Getenv("DATABASE_URL") == "" && os.Getenv("DB_HOST") == "" {
		t.Skip("No database configuration available for testing")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Set DATABASE_URL for this test
	originalURL := os.Getenv("DATABASE_URL")
	host := getEnvOrDefault("DB_HOST", "localhost")
	port := getEnvOrDefault("DB_PORT", "5432")
	user := getEnvOrDefault("DB_USER", "postgres")
	password := getEnvOrDefault("DB_PASSWORD", "postgres")
	testURL := fmt.Sprintf("postgresql://%s:%s@%s:%s/postgres", user, password, host, port)

	os.Setenv("DATABASE_URL", testURL)
	defer func() {
		if originalURL != "" {
			os.Setenv("DATABASE_URL", originalURL)
		} else {
			os.Unsetenv("DATABASE_URL")
		}
	}()

	db, err := NewDB(ctx)

	if err != nil {
		t.Skipf("Unable to connect to database: %v", err)
	}

	assert.NoError(t, err)
	assert.NotNil(t, db.conn)

	// Test that connection is working
	err = db.conn.Ping(ctx)
	assert.NoError(t, err)

	// Cleanup
	db.Close(ctx)
}

func TestNewDB_WithIndividualEnvVars(t *testing.T) {
	// Skip if no database available
	if os.Getenv("DB_HOST") == "" {
		t.Skip("No DB_HOST configuration available for testing")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Clear DATABASE_URL and set individual vars
	originalURL := os.Getenv("DATABASE_URL")
	os.Unsetenv("DATABASE_URL")
	defer func() {
		if originalURL != "" {
			os.Setenv("DATABASE_URL", originalURL)
		}
	}()

	os.Setenv("DB_HOST", getEnvOrDefault("DB_HOST", "localhost"))
	os.Setenv("DB_PORT", getEnvOrDefault("DB_PORT", "5432"))
	os.Setenv("DB_USER", getEnvOrDefault("DB_USER", "postgres"))
	os.Setenv("DB_PASSWORD", getEnvOrDefault("DB_PASSWORD", "postgres"))
	os.Setenv("DB_NAME", "postgres")

	db, err := NewDB(ctx)

	if err != nil {
		t.Skipf("Unable to connect to database: %v", err)
	}

	assert.NoError(t, err)
	assert.NotNil(t, db.conn)

	// Test that connection is working
	err = db.conn.Ping(ctx)
	assert.NoError(t, err)

	// Cleanup
	db.Close(ctx)
}

func TestNewDB_InvalidConnection(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Set invalid DATABASE_URL
	originalURL := os.Getenv("DATABASE_URL")
	os.Setenv("DATABASE_URL", "postgresql://invalid:invalid@nonexistent:5432/invalid")
	defer func() {
		if originalURL != "" {
			os.Setenv("DATABASE_URL", originalURL)
		} else {
			os.Unsetenv("DATABASE_URL")
		}
	}()

	_, err := NewDB(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unable to connect to database")
}

func TestGetSubscribers_EmptyTable(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	subscribers, err := db.GetSubscribers(ctx)

	assert.NoError(t, err)
	assert.Empty(t, subscribers)
}

func TestGetSubscribers_WithData(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Insert test data
	testPhones := []string{"+1234567890", "+1987654321", "+1555666777"}
	for _, phone := range testPhones {
		_, err := db.conn.Exec(ctx, "INSERT INTO subscribers (phone_number, unsubscribed) VALUES ($1, false)", phone)
		require.NoError(t, err)
	}

	// Insert unsubscribed user (should not be returned)
	_, err := db.conn.Exec(ctx, "INSERT INTO subscribers (phone_number, unsubscribed) VALUES ($1, true)", "+1000000000")
	require.NoError(t, err)

	subscribers, err := db.GetSubscribers(ctx)

	assert.NoError(t, err)
	assert.Len(t, subscribers, 3)
	assert.ElementsMatch(t, testPhones, subscribers)
}

func TestAddSubscriber_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	phone := "+1234567890"

	err := db.AddSubscriber(ctx, phone)
	assert.NoError(t, err)

	// Verify the subscriber was added
	var count int
	err = db.conn.QueryRow(ctx, "SELECT COUNT(*) FROM subscribers WHERE phone_number = $1", phone).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	// Verify default values
	var unsubscribed bool
	err = db.conn.QueryRow(ctx, "SELECT unsubscribed FROM subscribers WHERE phone_number = $1", phone).Scan(&unsubscribed)
	assert.NoError(t, err)
	assert.False(t, unsubscribed)
}

func TestAddSubscriber_Duplicate(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	phone := "+1234567890"

	// Add subscriber first time
	err := db.AddSubscriber(ctx, phone)
	assert.NoError(t, err)

	// Try to add same subscriber again
	err = db.AddSubscriber(ctx, phone)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unable to add")
}

func TestUnsubscribe_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	phone := "+1234567890"

	// Add subscriber first
	err := db.AddSubscriber(ctx, phone)
	require.NoError(t, err)

	// Unsubscribe
	err = db.Unsubscribe(ctx, phone)
	assert.NoError(t, err)

	// Verify unsubscribed status
	var unsubscribed bool
	err = db.conn.QueryRow(ctx, "SELECT unsubscribed FROM subscribers WHERE phone_number = $1", phone).Scan(&unsubscribed)
	assert.NoError(t, err)
	assert.True(t, unsubscribed)

	// Verify subscriber is not returned by GetSubscribers
	subscribers, err := db.GetSubscribers(ctx)
	assert.NoError(t, err)
	assert.NotContains(t, subscribers, phone)
}

func TestUnsubscribe_NonexistentUser(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	phone := "+1234567890"

	// Try to unsubscribe non-existent user
	err := db.Unsubscribe(ctx, phone)
	assert.NoError(t, err) // PostgreSQL UPDATE doesn't error when no rows are affected
}

func TestUnsubscribe_AlreadyUnsubscribed(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	phone := "+1234567890"

	// Add and immediately unsubscribe
	err := db.AddSubscriber(ctx, phone)
	require.NoError(t, err)

	err = db.Unsubscribe(ctx, phone)
	require.NoError(t, err)

	// Unsubscribe again
	err = db.Unsubscribe(ctx, phone)
	assert.NoError(t, err) // Should still succeed

	// Verify still unsubscribed
	var unsubscribed bool
	err = db.conn.QueryRow(ctx, "SELECT unsubscribed FROM subscribers WHERE phone_number = $1", phone).Scan(&unsubscribed)
	assert.NoError(t, err)
	assert.True(t, unsubscribed)
}

func TestClose_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	err := db.Close(ctx)
	assert.NoError(t, err)

	// Verify connection is closed by trying to ping
	err = db.conn.Ping(ctx)
	assert.Error(t, err)
}

func TestDatabase_IntegrationWorkflow(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Start with empty table
	subscribers, err := db.GetSubscribers(ctx)
	assert.NoError(t, err)
	assert.Empty(t, subscribers)

	// Add multiple subscribers
	phones := []string{"+1234567890", "+1987654321", "+1555666777"}
	for _, phone := range phones {
		err := db.AddSubscriber(ctx, phone)
		assert.NoError(t, err)
	}

	// Verify all subscribers are returned
	subscribers, err = db.GetSubscribers(ctx)
	assert.NoError(t, err)
	assert.Len(t, subscribers, 3)
	assert.ElementsMatch(t, phones, subscribers)

	// Unsubscribe one user
	err = db.Unsubscribe(ctx, phones[0])
	assert.NoError(t, err)

	// Verify only 2 subscribers remain
	subscribers, err = db.GetSubscribers(ctx)
	assert.NoError(t, err)
	assert.Len(t, subscribers, 2)
	assert.NotContains(t, subscribers, phones[0])
	assert.Contains(t, subscribers, phones[1])
	assert.Contains(t, subscribers, phones[2])
}
