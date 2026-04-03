package mlb_stats_api_client

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

const (
	DODGERS_ID = 119
	FINISHED   = "F"
	BASE_URL   = "https://statsapi.mlb.com/api/v1/schedule"
)

type MLBAPIResponse struct {
	TotalGames int    `json:"totalGames"`
	Dates      []Date `json:"dates"`
}

type Date struct {
	Date  string `json:"date"`
	Games []Game `json:"games"`
}

type Game struct {
	Status Status `json:"status"`
	Teams  struct {
		Away TeamDetails `json:"away"`
		Home TeamDetails `json:"home"`
	}
}

type Status struct {
	StatusCode string `json:"statusCode"`
}

type TeamDetails struct {
	Team struct {
		Id int `json:"id"`
	}
	IsWinner bool `json:"isWinner"`
}

type MLBStatsAPIClient interface {
	GetResults(date string) (*MLBAPIResponse, error)
	Close()
}

type Client struct {
	httpClient *http.Client
}

func NewMLBClient() MLBStatsAPIClient {
	return &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) GetResults(date string) (*MLBAPIResponse, error) {
	resp, err := c.httpClient.Get(BASE_URL + fmt.Sprintf("?sportId=1&teamId=%d&date=%s", DODGERS_ID, date))
	if err != nil {
		log.Printf("Error making request to MLB Stats API: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Bad status code from MLB Stats API: %d", resp.StatusCode)
		return nil, fmt.Errorf("bad status code from MLB Stats API: %d", resp.StatusCode)
	}

	var mlbAPIResponse MLBAPIResponse
	err = json.NewDecoder(resp.Body).Decode(&mlbAPIResponse)
	if err != nil {
		log.Printf("Error unmarshalling response body from MLB Stats API: %v\n", err)
		return nil, fmt.Errorf("error unmarshalling response body from MLB Stats API: %v", err)
	}

	return &mlbAPIResponse, nil
}

func (c *Client) Close() {
	c.httpClient.CloseIdleConnections()
}
