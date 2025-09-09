package service

import (
	"circulation-supply-api/config"
	"testing"
)

func TestCalculateAverageBlockTime(t *testing.T) {
	config.Conf = &config.Config{}
	config.Conf.RpcEndpoint = "https://mainnet.storyrpc.io"
	err := CalculateAverageBlockTime()
	if err != nil {
		t.Errorf("CalculateAverageBlockTime() error = %v", err)
	}
}
