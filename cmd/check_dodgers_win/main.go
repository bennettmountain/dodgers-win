package main

import (
	"fmt"
	"log"
	"time"

	"dodgers-win/internal/mlb_stats_api_client"
)

func main() {
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	fmt.Printf("Checking if Dodgers won a home game on %s...\n", yesterday)

	client := mlb_stats_api_client.NewMLBClient()
	defer client.Close()

	resp, err := client.GetResults(yesterday)
	if err != nil {
		log.Fatalf("Error querying MLB API: %v", err)
	}

	for _, dateData := range resp.Dates {
		for _, game := range dateData.Games {
			homeTeam := game.Teams.Home.Team.Id
			awayTeam := game.Teams.Away.Team.Id
			status := game.Status.StatusCode
			homeWin := game.Teams.Home.IsWinner

			fmt.Printf("Game: Away(%d) @ Home(%d) | Status: %s | Home Win: %v\n",
				awayTeam, homeTeam, status, homeWin)

			if status == mlb_stats_api_client.FINISHED &&
				homeTeam == mlb_stats_api_client.DODGERS_ID &&
				homeWin {
				fmt.Println("✅ Dodgers won at home! Deal is ACTIVE.")
				return
			}
		}
	}

	fmt.Println("❌ No Dodgers home win yesterday. Deal is NOT active.")
}
