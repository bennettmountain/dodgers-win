package db

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/jackc/pgx/v5"
)

type Database struct {
	conn *pgx.Conn
}

func NewDB(ctx context.Context) (Database, error) {
	var err error

	// for local development the DATABASE_URL environment variable is set.
	// on supabase the DATABASE_URL environment variable is not set so we
	// must construct the DATABASE_URL string from the individual environment variables
	// supabase sets
	dbURL := os.Getenv("DATABASE_URL")

	if dbURL == "" {
		host := os.Getenv("DB_HOST")
		port := os.Getenv("DB_PORT")
		user := os.Getenv("DB_USER")
		name := os.Getenv("DB_NAME")
		password := os.Getenv("DB_PASSWORD")

		encodedPassword := url.QueryEscape(password)
		fmt.Printf("connecting to database postgresql://%s:****@%s:%s/%s\n", user, host, port, name)
		dbURL = fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", user, encodedPassword, host, port, name)
	}

	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		return Database{}, fmt.Errorf("unable to connect to database: %w", err)
	}

	err = conn.Ping(ctx)
	if err != nil {
		return Database{}, fmt.Errorf("unable to ping database: %w", err)
	}

	return Database{conn: conn}, nil
}

func (db *Database) GetSubscribers(ctx context.Context) ([]string, error) {
	rows, err := db.conn.Query(ctx, "SELECT phone_number FROM subscribers WHERE unsubscribed = false")
	if err != nil {
		return nil, fmt.Errorf("unable to get subscribers: %w", err)
	}
	defer rows.Close()

	subscribers := []string{}
	for rows.Next() {
		var phoneNumber string
		err := rows.Scan(&phoneNumber)
		if err != nil {
			return nil, fmt.Errorf("unable to scan subscribers table: %w", err)
		}
		subscribers = append(subscribers, phoneNumber)
	}

	return subscribers, nil
}

func (db *Database) AddSubscriber(ctx context.Context, phoneNumber string) error {
	_, err := db.conn.Exec(ctx, "INSERT INTO subscribers (phone_number) VALUES ($1)", phoneNumber)
	if err != nil {
		return fmt.Errorf("unable to add %s to subscribers table: %w", phoneNumber, err)
	}
	return nil
}

func (db *Database) Unsubscribe(ctx context.Context, phoneNumber string) error {
	_, err := db.conn.Exec(ctx, "UPDATE subscribers SET unsubscribed = true WHERE phone_number = $1", phoneNumber)
	if err != nil {
		return fmt.Errorf("unable to unsubscribe %s from subscribers table: %w", phoneNumber, err)
	}
	return nil
}

func (db *Database) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return db.conn.Query(ctx, sql, args...)
}

func (db *Database) Close(ctx context.Context) error {
	return db.conn.Close(ctx)
}
