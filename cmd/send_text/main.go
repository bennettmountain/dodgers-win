package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/joho/godotenv"

	"dodgers-win/internal/consts"
	"dodgers-win/internal/db"
	"dodgers-win/internal/twilio_client"
)

func main() {
	phoneNumber := flag.String("phone-number", "", "phone number to send the text to (e.g. +11234567890)")
	allSubscribers := flag.Bool("all-subscribers", false, "send the text to all subscribers in the remote DB")
	textType := flag.String("text", "", "type of text to send: \"win\" or \"welcome\"")
	flag.Parse()

	if *textType == "" {
		flag.Usage()
		log.Fatal("--text is required")
	}

	if *phoneNumber == "" && !*allSubscribers {
		flag.Usage()
		log.Fatal("either --phone-number or --all-subscribers is required")
	}

	if *phoneNumber != "" && *allSubscribers {
		log.Fatal("--phone-number and --all-subscribers are mutually exclusive")
	}

	var body string
	var mediaUrl []string
	switch *textType {
	case "win":
		body = consts.DODGERS_WIN_TEXT
		mediaUrl = []string{consts.DODGERS_WIN_GIF}
	case "welcome":
		body = consts.WELCOME_TEXT
	default:
		log.Fatalf("invalid --text value %q: must be \"win\" or \"welcome\"", *textType)
	}

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	ctx := context.Background()

	var recipients []string
	if *allSubscribers {
		database, err := db.NewDB(ctx)
		if err != nil {
			log.Fatalf("Error connecting to database: %v", err)
		}
		defer database.Close(ctx)

		recipients, err = database.GetSubscribers(ctx)
		if err != nil {
			log.Fatalf("Error getting subscribers: %v", err)
		}
		fmt.Printf("Found %d active subscribers\n", len(recipients))
	} else {
		recipients = []string{*phoneNumber}
	}

	client, err := twilio_client.NewClient()
	if err != nil {
		log.Fatalf("Error creating Twilio client: %v", err)
	}

	for _, number := range recipients {
		fmt.Printf("Sending %q text to %s...\n", *textType, number)
		err = client.SendMessage(ctx, number, body, mediaUrl...)
		if err != nil {
			log.Fatalf("Error sending message to %s: %v", number, err)
		}
	}

	fmt.Printf("Successfully sent %d text(s)!\n", len(recipients))
}
