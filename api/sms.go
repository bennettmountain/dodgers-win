package dodgers_win_alerter

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/joho/godotenv"

	"dodgers-win/pkg/consts"
	"dodgers-win/pkg/db"
	"dodgers-win/pkg/twilio_client"
)

func Sms(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	err = r.ParseForm()
	if err != nil {
		log.Printf("Error parsing form: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	from := r.FormValue("From")
	body := strings.TrimSpace(strings.ToLower(r.FormValue("Body")))

	ctx := context.Background()

	database, err := db.NewDB(ctx)
	if err != nil {
		log.Printf("Error connecting to database: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer database.Close(ctx)

	twilioClient, err := twilio_client.NewClient()
	if err != nil {
		log.Printf("Error creating twilio client: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	switch body {
	case "dodgerswin":
		err = database.AddSubscriber(ctx, from)
		if err != nil {
			log.Printf("Error adding subscriber %s: %v", from, err)
			// still try to respond even if DB insert fails (e.g. duplicate)
		}

		err = twilioClient.SendMessage(ctx, from, consts.WELCOME_TEXT)
		if err != nil {
			log.Printf("Error sending welcome text to %s: %v", from, err)
		}

		err = twilioClient.SendMessageWithContactCard(ctx, from)
		if err != nil {
			log.Printf("Error sending contact card to %s: %v", from, err)
		}
	case "stop":
		err = database.Unsubscribe(ctx, from)
		if err != nil {
			log.Printf("Error unsubscribing %s: %v", from, err)
		}
	default:
		log.Printf("Unknown message from %s: %q", from, body)
	}

	// Respond with empty TwiML so Twilio doesn't retry
	w.Header().Set("Content-Type", "text/xml")
	fmt.Fprint(w, "<?xml version=\"1.0\" encoding=\"UTF-8\"?><Response></Response>")
}
