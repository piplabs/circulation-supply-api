package service

import (
	"circulation-supply-api/config"
	"math/big"
	"testing"
	"time"
)

func setConf() {
	config.Conf = &config.Config{
		VestingStartYear:  2025,
		VestingStartMonth: 2,
		VestingStartDay:   13,
		Vesting: []float64{
			250000000, 256952083, 263904167, 270856250, 277808333, 284760417,
			291712500, 298664581, 305616664, 312568747, 319520830, 326472913,
			427499996, 445655899, 463811802, 481967705, 500123608, 518279512,
			536435415, 554591318, 572747221, 590903124, 609059027, 627214930,
			645370833, 663526736, 681682640, 699838543, 717994446, 736150349,
			754306252, 772462155, 790618058, 808773961, 826929864, 845085767,
			863241670, 874638198, 886034725, 897431253, 908827780, 920224308,
			931620836, 943017363, 954413891, 965810418, 977206946, 988603473,
			1000000000,
		},
	}
	mintPerSec, _ = new(big.Float).SetPrec(64).Quo(
		new(big.Float).SetFloat64(BlockReward),
		new(big.Float).SetFloat64(2.364),
	).Float64()
}

func TestMonthsPassedSinceGenesis(t *testing.T) {
	setConf()
	tests := []struct {
		blockTime *big.Int
		expected  int
	}{

		// On vesting start date (2025-02-13)
		{big.NewInt(1739404800), 0}, // 2025-02-13 00:00:00 UTC

		// Just before the next month's 13th (2025-03-12)
		{big.NewInt(1741737600), 0}, // 2025-03-12 00:00:00 UTC

		// One month after vesting start (2025-03-13)
		{big.NewInt(1741824000), 1}, // 2025-03-13 00:00:00 UTC

		// Two months after vesting start (2025-04-13)
		{big.NewInt(1744502400), 2}, // 2025-04-13 00:00:00 UTC
	}

	for _, tt := range tests {
		t.Run("test", func(t *testing.T) {
			if got := MonthsPassedSinceGenesis(tt.blockTime); got != tt.expected {
				t.Errorf("MonthsPassedSinceGenesis(%v) = %v; want %v", tt.blockTime, got, tt.expected)
			}
		})
	}
}

func TestGetSupplyDelta(t *testing.T) {
	setConf()

	expectedDelta := &SupplyDelta{
		StartTime:      "2026-01-13",
		EndTime:        "2026-02-13",
		InflationDelta: new(big.Float).Mul(new(big.Float).SetFloat64(mintPerSec), new(big.Float).SetInt64(31*24*3600)).Text('f', 2),
		VestingDelta:   new(big.Float).SetFloat64(427499996-326472913).Text('f', 2),
	}

	tests := []struct {
		startTimestamp int64
		endTimestamp   int64
		expectedDelta  *SupplyDelta
	}{
		{startTimestamp: timeStringToTimestamp("2026-01-13"), endTimestamp: timeStringToTimestamp("2026-02-13"), expectedDelta: expectedDelta},
	}

	for _, tt := range tests {
		t.Run("test", func(t *testing.T) {
			r, err := GetSupplyDelta(tt.startTimestamp, tt.endTimestamp)
			if err != nil {
				t.Error("GetSupplyDelta() error =", err)
			}
			if r.StartTime != tt.expectedDelta.StartTime {
				t.Errorf("GetSupplyDelta() StartTime = %v; want %v", r.StartTime, tt.expectedDelta.StartTime)
			}
			if r.EndTime != tt.expectedDelta.EndTime {
				t.Errorf("GetSupplyDelta() EndTime = %v; want %v", r.EndTime, tt.expectedDelta.EndTime)
			}
			if r.InflationDelta != tt.expectedDelta.InflationDelta {
				t.Errorf("GetSupplyDelta() InflationDelta = %v; want %v", r.InflationDelta, tt.expectedDelta.InflationDelta)
			}
			if r.VestingDelta != tt.expectedDelta.VestingDelta {
				t.Errorf("GetSupplyDelta() VestingDelta = %v; want %v", r.VestingDelta, tt.expectedDelta.VestingDelta)
			}
		})
	}
}

func timeStringToTimestamp(t string) int64 {
	layout := "2006-01-02"
	tm, _ := time.ParseInLocation(layout, t, time.UTC)
	return tm.Unix()
}
