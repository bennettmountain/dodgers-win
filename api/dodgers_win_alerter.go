package dodgers_win_alerter

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"

	"dodgers-win/pkg/consts"
	"dodgers-win/pkg/db"
	"dodgers-win/pkg/mlb_stats_api_client"
	"dodgers-win/pkg/twilio_client"
)

// Alerter is the entry point for the Vercel Serverless Function
// it checks if the Dodgers won yesterday and sends a text message to all subscribers
// if the Dodgers won
func Alerter(w http.ResponseWriter, r *http.Request) {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	dealActive, err := checkDodgersWin()
	if err != nil {
		log.Printf("Error checking Dodgers win: %v", err)
		return
	}

	if dealActive {
		err = sendDodgersWinTexts(context.Background())
		if err != nil {
			log.Printf("Error sending SMS messages: %v", err)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func checkDodgersWin() (bool, error) {
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	client := mlb_stats_api_client.NewMLBClient()
	defer client.Close()

	mlbAPIResponse, err := client.GetResults(yesterday)
	if err != nil {
		return false, err
	}

	for _, dateData := range mlbAPIResponse.Dates {
		gameDate, err := time.Parse("2006-01-02", dateData.Date)
		if err != nil {
			return false, err
		}

		// the MLB API shouldn't return games from any other day we request
		// but let's still check just in case
		if gameDate.Format("2006-01-02") != yesterday {
			continue
		}

		// iterate over all games played on the date
		// to take into account double headers
		for _, game := range dateData.Games {
			if game.Status.StatusCode == mlb_stats_api_client.FINISHED &&
				game.Teams.Home.Team.Id == mlb_stats_api_client.DODGERS_ID &&
				game.Teams.Home.IsWinner {
				return true, nil
			}
		}
	}

	return false, nil
}

func sendDodgersWinTexts(ctx context.Context) error {
	db, err := db.NewDB(ctx)
	if err != nil {
		return fmt.Errorf("error creating db: %w", err)
	}

	defer db.Close(ctx)

	subscribers, err := db.GetSubscribers(ctx)
	if err != nil {
		return fmt.Errorf("error getting subscribers: %w", err)
	}

	twilioClient, err := twilio_client.NewClient()

	if err != nil {
		return fmt.Errorf("error creating twilio client: %w", err)
	}

	for _, subscriber := range subscribers {
		err = twilioClient.SendMessage(ctx, subscriber, consts.DODGERS_WIN_TEXT, os.Getenv("DODGERS_WIN_GIF_URL"))
		if err != nil {
			return fmt.Errorf("error sending message to %s: %w", subscriber, err)
		}
	}

	return nil
}
