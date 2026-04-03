package mlb_stats_api_client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetResults(t *testing.T) {
	test_cases := []struct {
		name     string
		date     string
		expected bool
	}{
		{
			name:     "home win",
			date:     "2025-05-21",
			expected: true,
		},
		{
			name:     "away win",
			date:     "2025-05-22",
			expected: true,
		},
		{
			name:     "away loss",
			date:     "2025-05-23",
			expected: false,
		},
		{
			name:     "no game",
			date:     "2025-05-24",
			expected: false,
		},
	}

	for _, test_case := range test_cases {
		t.Run(test_case.name, func(t *testing.T) {
			client := NewMockMLBStatsAPIClient()
			actual, err := client.GetResults(test_case.date)
			assert.Nil(t, err)
			assert.Equal(t, test_case.expected, actual.TotalGames)
		})
	}
}
