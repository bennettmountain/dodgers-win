package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/joho/godotenv"

	"dodgers-win/internal/db"
)

type subscriber struct {
	id           int
	phoneNumber  string
	unsubscribed bool
	created      time.Time
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	ctx := context.Background()
	database, err := db.NewDB(ctx)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer database.Close(ctx)

	rows, err := database.Query(ctx, "SELECT id, phone_number, unsubscribed, created FROM subscribers ORDER BY id")
	if err != nil {
		log.Fatalf("Error querying subscribers: %v", err)
	}
	defer rows.Close()

	fmt.Printf("%-5s %-16s %-14s %s\n", "ID", "Phone Number", "Unsubscribed", "Created")
	fmt.Println("---   ---------------  ------------   -------------------")

	count := 0
	for rows.Next() {
		var s subscriber
		err := rows.Scan(&s.id, &s.phoneNumber, &s.unsubscribed, &s.created)
		if err != nil {
			log.Fatalf("Error scanning row: %v", err)
		}
		fmt.Printf("%-5d %-16s %-14v %s\n", s.id, s.phoneNumber, s.unsubscribed, s.created.Format("2006-01-02 15:04:05"))
		count++
	}

	fmt.Printf("\nTotal: %d subscribers\n", count)
}
