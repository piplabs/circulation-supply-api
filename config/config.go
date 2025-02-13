package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

var Conf *Config

var defaultMetrics = Metric{
	Listen: "0.0.0.0",
	Port:   9111,
}

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

	// should be older than first block contains stake transaction, so that the backward scan end here, or it will scan till the genesis block
	BackwardsStartBlock int `yaml:"backwardsStartBlock"`

	// we check all internal transactions of StakeReceiverAddress(0x00000..000), and sum up the value of the transactions that are from StakeContractAddress and value >= StakeThreshold
	StakeContractAddress string  `yaml:"stakeContractAddress"`
	StakeReceiverAddress string  `yaml:"stakeReceiverAddress"`
	StakeThreshold       float64 `yaml:"stakeThreshold"` // usually 1024 IP

	RealTimeDataAvailableAt int64 `yaml:"realTimeDataAvailableAt"`

	Metric Metric `yaml:"_"`
}

type Metric struct {
	Listen string
	Port   int
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

	if len(config.StakeContractAddress) == 0 {
		return nil, fmt.Errorf("stakeContractAddress is empty")
	}

	if len(config.StakeReceiverAddress) == 0 {
		return nil, fmt.Errorf("stakeReceiverAddress is empty")
	}

	if config.StakeThreshold == 0 {
		return nil, fmt.Errorf("stakeThreshold is empty")
	}

	config.Metric = defaultMetrics
	return &config, nil
}
