package service

import (
	"circulation-supply-api/config"
	"circulation-supply-api/dao"
	"fmt"
	"math/big"
	"time"

	log "log/slog"
)

var (
	cachedSupply  string
	cachedTime    = time.Now()
	cacheDuration = 30 * time.Second
)

// 1. Get latest block from database
// 2. Get totalBurntBaseFee and totalStakeReward
// 3. Get timestamp of the latest block
// 4. Get the balance of the zero address at the latest block
// 5. Get circulating supply from vesting plan
// 6. Calculate the total supply = circulating supply - totalBurntBaseFee - IPSentToZeroAddress + totalStakeReward
func GetSupply() (string, error) {
	// add cache, if last call is within 30 seconds, return the same result
	if time.Since(cachedTime) < cacheDuration && cachedSupply != "" {
		return cachedSupply, nil
	}

	fee, err := dao.GetLatestAccumulatedFees()
	if err != nil {
		log.Error("Failed to get latest accumulated fees", "error", err)
		return "", err
	}
	blockNumber := fee.BlockNumber
	totalBurntBaseFee, b := new(big.Float).SetString(fee.TotalBurntBaseFee)
	if !b {
		return "", fmt.Errorf("error parsing total burnt base fee")
	}
	totalStakeReward, b := new(big.Float).SetString(fee.TotalStakeReward)
	if !b {
		return "", fmt.Errorf("error parsing total stake reward")
	}

	block, err := getBlockByNumber(fmt.Sprintf("0x%x", blockNumber))
	if err != nil {
		log.Error("Failed to get block by number", "blockNumber", blockNumber, "error", err)
		return "", err
	}
	if block == nil {
		log.Error("Block not found", "blockNumber", blockNumber)
		return "", fmt.Errorf("block not found")
	}
	blockTimestamp, _ := new(big.Int).SetString(block.Time, 0)

	ipSentToZeroAddress := big.NewFloat(0)
	totalBalance := big.NewInt(0)
	for _, zeroAddress := range config.Conf.ZeroAddresses {
		balance, err := getBalanceAtBlock(zeroAddress, blockNumber)
		if err != nil {
			log.Error("Failed to get zero address balance", "address", zeroAddress, "error", err)
			return "", err
		}
		totalBalance.Add(totalBalance, balance)
	}
	ipSentToZeroAddress = new(big.Float).Quo(new(big.Float).SetInt(totalBalance), big.NewFloat(1e18))

	// calculate months passed since genesis
	year, month, _ := time.Unix(blockTimestamp.Int64(), 0).UTC().Date()
	monthsPassed := (year-config.Conf.VestingStartYear)*12 + int(month-time.Month(config.Conf.VestingStartMonth))
	var circulatingSupply *big.Float
	if monthsPassed < 0 {
		// this should never happen in mainnet
		log.Warn("Invalid timestamp", "timestamp", blockTimestamp, "monthsPassed", monthsPassed)
		circulatingSupply = big.NewFloat(config.Conf.Vesting[0])
	} else if monthsPassed >= len(config.Conf.Vesting) {
		circulatingSupply = big.NewFloat(config.Conf.Vesting[len(config.Conf.Vesting)-1])
	} else {
		circulatingSupply = big.NewFloat(config.Conf.Vesting[monthsPassed])
	}
	supply := new(big.Float).Sub(circulatingSupply, totalBurntBaseFee).SetPrec(64)
	supply.Sub(supply, ipSentToZeroAddress)
	supply.Add(supply, totalStakeReward)
	log.Info("Show details - ", "circulatingSupply", circulatingSupply.Text('f', -1), "totalBurntBaseFee", totalBurntBaseFee.Text('f', -1), "IPSentToZeroAddress", ipSentToZeroAddress.Text('f', -1), "totalStakeReward", totalStakeReward.Text('f', -1), "supply", supply.Text('f', -1))

	ret := supply.Text('f', 2)
	cachedSupply = ret
	cachedTime = time.Now()

	return ret, nil
}
