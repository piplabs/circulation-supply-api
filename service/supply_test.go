package service

import (
	"circulation-supply-api/config"
	"testing"
)

func setConf() {
	config.Conf = &config.Config{
		VestingStartYear:  2025,
		VestingStartMonth: 2,
		VestingStartDay:   13,
	}
}

func TestMonthsPassedSinceGenesis(t *testing.T) {
	setConf()
	tests := []struct {
		blockTime string
		expected  int
	}{

		// On vesting start date (2025-02-13)
		{"1739404800", 0}, // 2025-02-13 00:00:00 UTC

		// Just before the next month's 13th (2025-03-12)
		{"1741737600", 0}, // 2025-03-12 00:00:00 UTC

		// One month after vesting start (2025-03-13)
		{"1741824000", 1}, // 2025-03-13 00:00:00 UTC

		// Two months after vesting start (2025-04-13)
		{"1744502400", 2}, // 2025-04-13 00:00:00 UTC
	}

	for _, tt := range tests {
		t.Run(tt.blockTime, func(t *testing.T) {
			if got := MonthsPassedSinceGenesis(tt.blockTime); got != tt.expected {
				t.Errorf("MonthsPassedSinceGenesis(%v) = %v; want %v", tt.blockTime, got, tt.expected)
			}
		})
	}
}
