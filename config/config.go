package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

var Conf *Config

type Config struct {
	// archive node rpc endpoint
	RpcEndpoint string `yaml:"rpcEndpoint"`

	// zero addresses are where the tokens are sent to and are considered as burned
	ZeroAddresses []string `yaml:"zeroAddresses"`

	// year when the vesting starts
	VestingStartYear int `yaml:"vestingStartYear"`

	// month when the vesting starts
	VestingStartMonth int `yaml:"vestingStartMonth"`

	// each element represents the circulating supply for a month starting from "VestingStartYear/VestingStartMonth"
	// e.g. Vesting = [100, 200, 300] and VestingStartYear = 2025 and VestingStartMonth = 1, then the circulating supply
	// for 2025/1 is 100, for 2025/2 is 200, for 2025/3 is 300.
	Vesting []float64 `yaml:"vesting"`

	GenesisTotalSupply float64 `yaml:"genesisTotalSupply"`
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
	if config.VestingStartYear == 0 {
		return nil, fmt.Errorf("vestingStartYear is empty")
	}
	if config.VestingStartMonth == 0 {
		return nil, fmt.Errorf("vestingStartMonth is empty")
	}
	if len(config.RpcEndpoint) == 0 {
		return nil, fmt.Errorf("rpcEndpoint is empty")
	}
	if len(config.ZeroAddresses) == 0 {
		return nil, fmt.Errorf("zeroAddresses is empty")
	}
	if config.GenesisTotalSupply == 0 {
		return nil, fmt.Errorf("genesis is empty")
	}
	return &config, nil
}
