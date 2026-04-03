package mlb_stats_api_client

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/stretchr/testify/mock"
)

// MockMLBStatsAPIClient implements MLBStatsAPIClient for testing using testify/mock
type MockMLBStatsAPIClient struct {
	mock.Mock
}

func NewMockMLBStatsAPIClient() *MockMLBStatsAPIClient {
	mock := new(MockMLBStatsAPIClient)
	mock.
		On("GetResults", "2025-05-21").
		Return(CreateMockHomeWinResponse()).
		On("GetResults", "2025-05-22").
		Return(CreateMockAwayWinResponse()).
		On("GetResults", "2025-05-23").
		Return(CreateMockAwayLossResponse()).
		On("GetResults", "2025-05-24").
		Return(CreateMockNoGameResponse())
	return mock
}

func (m *MockMLBStatsAPIClient) GetResults(date string) (*MLBAPIResponse, error) {
	args := m.Called(date)
	return args.Get(0).(*MLBAPIResponse), args.Error(1)
}

// LoadMockResponse loads a JSON response from file
func LoadMockResponse(filename string) (*MLBAPIResponse, error) {
	// Get the directory of this file
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// Construct path to the JSON file
	jsonPath := filepath.Join(currentDir, "internal", "mlb_stats_api", filename)

	// Read the JSON file
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, err
	}

	// Parse the JSON
	var response MLBAPIResponse
	err = json.Unmarshal(data, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// Helper functions to create mock responses from actual JSON files
func CreateMockHomeWinResponse() (*MLBAPIResponse, error) {
	return LoadMockResponse("home_win.json")
}

func CreateMockAwayWinResponse() (*MLBAPIResponse, error) {
	return LoadMockResponse("away_win.json")
}

func CreateMockAwayLossResponse() (*MLBAPIResponse, error) {
	return LoadMockResponse("away_loss.json")
}

func CreateMockNoGameResponse() (*MLBAPIResponse, error) {
	return LoadMockResponse("no_game.json")
}
