package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

var Conf *Config

type Config struct {
	RpcEndpoint   string   `yaml:"rpcEndpoint"`
	ZeroAddresses []string `yaml:"zeroAddresses"`

	// Vesting schedule
	VestingStartYear  int       `yaml:"vestingStartYear"`
	VestingStartMonth int       `yaml:"vestingStartMonth"`
	Vesting           []float64 `yaml:"vesting"` // each element represents the circulating supply for a month starting from "VestingStartYear/VestingStartMonth"
}

func LoadConfig(filePath string) (*Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %v", err)
	}
	defer file.Close()

	var config Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %v", err)
	}

	// Validate config
	if len(config.Vesting) == 0 {
		return nil, fmt.Errorf("vesting schedule is empty")
	}
	if len(config.RpcEndpoint) == 0 {
		return nil, fmt.Errorf("rpcEndpoint is empty")
	}
	if len(config.ZeroAddresses) == 0 {
		return nil, fmt.Errorf("zeroAddresses is empty")
	}
	return &config, nil
}
