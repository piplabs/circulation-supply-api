package service

import (
	"circulation-supply-api/config"
	"circulation-supply-api/dao"
	"errors"
	"fmt"
	"math/big"
	"time"

	log "log/slog"

	"gorm.io/gorm"
)

var (
	cachedTotalSupply     *big.Float
	cachedTotalSupplyTime = time.Now()

	cachedCirculatingSupply     *big.Float
	cachedCirculatingSupplyTime = time.Now()

	cacheDuration = 30 * time.Second
)

// 1. Get latest block from database
// 2. Get totalBurntBaseFee and totalStakeReward
// 3. Get timestamp of the latest block
// 4. Get the balance of the zero address at the latest block
// 5. Get accumulated vested supply from vesting plan
// 6. Calculate the total circulating supply = accumulated vested supply - totalBurntBaseFee - IPSentToZeroAddress + totalStakeReward
func GetCirculatingSupply() (*big.Float, error) {
	// add cache, if last call is within 30 seconds, return the same result
	if time.Since(cachedCirculatingSupplyTime) < cacheDuration && cachedCirculatingSupply != nil {
		return cachedCirculatingSupply, nil
	}

	totalBurntBaseFee, totalStakeReward, totalStakedToken, ipSentToZeroAddress, blockTime, err := GetAccumulatedFees()
	if err != nil {
		return nil, err
	}

	var vestedSupply *big.Float
	monthsPassed := MonthsPassedSinceGenesis(blockTime)
	if monthsPassed < 0 {
		// this should never happen in mainnet
		log.Warn("Invalid timestamp", "blockTime", blockTime, "monthsPassed", monthsPassed)
		vestedSupply = big.NewFloat(config.Conf.Vesting[0])
	} else if monthsPassed >= len(config.Conf.Vesting) {
		vestedSupply = big.NewFloat(config.Conf.Vesting[len(config.Conf.Vesting)-1])
	} else {
		vestedSupply = big.NewFloat(config.Conf.Vesting[monthsPassed])
	}
	circulatingSupply := new(big.Float).Sub(vestedSupply, totalBurntBaseFee).SetPrec(64)
	circulatingSupply.Sub(circulatingSupply, ipSentToZeroAddress)
	circulatingSupply.Add(circulatingSupply, totalStakeReward)

	// ipSentToZeroAddress contains totalStakedToken, so we need to add it back
	circulatingSupply.Add(circulatingSupply, totalStakedToken)
	log.Info("Show details - ", "vestedSupply", vestedSupply.Text('f', -1), "totalBurntBaseFee",
		totalBurntBaseFee.Text('f', -1), "IPSentToZeroAddress", ipSentToZeroAddress.Text('f', -1), "totalStakedToken", totalStakedToken.Text('f', -1),
		"totalStakeReward", totalStakeReward.Text('f', -1), "circulatingSupply", circulatingSupply.Text('f', -1), "monthsPassed", monthsPassed)

	cachedCirculatingSupply = circulatingSupply
	cachedCirculatingSupplyTime = time.Now()

	return circulatingSupply, nil
}

// 1. Get latest block from database
// 2. Get totalBurntBaseFee and totalStakeReward
// 3. Get timestamp of the latest block
// 4. Get the balance of the zero address at the latest block
// 5. Get genesis total supply from vesting plan
// 6. Calculate the total supply = genesis total supply - totalBurntBaseFee - IPSentToZeroAddress + totalStakeReward
func GetTotalSupply() (*big.Float, error) {
	// add cache, if last call is within 30 seconds, return the same result
	if time.Since(cachedTotalSupplyTime) < cacheDuration && cachedTotalSupply != nil {
		return cachedTotalSupply, nil
	}

	totalBurntBaseFee, totalStakeReward, totalStakedToken, ipSentToZeroAddress, _, err := GetAccumulatedFees()
	if err != nil {
		return nil, err
	}

	totalSupply := big.NewFloat(config.Conf.GenesisTotalSupply).SetPrec(64)
	totalSupply.Sub(totalSupply, totalBurntBaseFee)
	totalSupply.Sub(totalSupply, ipSentToZeroAddress)
	totalSupply.Add(totalSupply, totalStakeReward)
	// ipSentToZeroAddress contains totalStakedToken, so we need to add it back
	totalSupply.Add(totalSupply, totalStakedToken)

	log.Info("Show details - ", "genesisTotalSupply", config.Conf.GenesisTotalSupply, "totalBurntBaseFee", totalBurntBaseFee.Text('f', -1),
		"IPSentToZeroAddress", ipSentToZeroAddress.Text('f', -1), "totalStakedToken", totalStakedToken.Text('f', -1),
		"totalStakeReward", totalStakeReward.Text('f', -1), "totalSupply", totalSupply.Text('f', -1))

	cachedTotalSupply = totalSupply
	cachedTotalSupplyTime = time.Now()

	return totalSupply, nil
}

func MonthsPassedSinceGenesis(blockTime string) int {
	blockTimestamp, _ := new(big.Int).SetString(blockTime, 0)
	year, month, day := time.Unix(blockTimestamp.Int64(), 0).UTC().Date()
	monthsPassed := (year-config.Conf.VestingStartYear)*12 + int(month-time.Month(config.Conf.VestingStartMonth))

	// do not adjust according to vesting schedule until 13th of each month
	if monthsPassed > 0 && day < config.Conf.VestingStartDay {
		monthsPassed--
	}
	return monthsPassed
}

func GetAccumulatedFees() (totalBurntBaseFee, totalStakeReward, totalStakedToken, ipSentToZeroAddress *big.Float, blockTime string, err error) {
	var fee *dao.AccumulatedFees
	var block *Block
	var b bool

	fee, err = dao.GetLatestAccumulatedFees()
	if err != nil || fee == nil {
		log.Error("Failed to get latest accumulated fees", "error", err)
		return
	}
	blockNumber := fee.BlockNumber
	totalBurntBaseFee, b = new(big.Float).SetString(fee.TotalBurntBaseFee)
	if !b {
		err = fmt.Errorf("error parsing total burnt base fee")
		return
	}
	totalStakeReward, b = new(big.Float).SetString(fee.TotalStakeReward)
	if !b {
		err = fmt.Errorf("error parsing total stake reward")
		return
	}

	// newly added column could be nil
	if len(fee.TotalStakedToken) > 0 {
		totalStakedToken, b = new(big.Float).SetString(fee.TotalStakedToken)
		if !b {
			err = fmt.Errorf("error parsing total staked token")
			return
		}
	} else {
		totalStakedToken = big.NewFloat(0)
	}

	// get backwards staked token
	_, amount, err := dao.GetBackwardsStakedToken()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			amount = "0"
		} else {
			log.Error("Failed to get backwards staked token", "error", err)
			return
		}
	}
	backwardsStakedToken, b := new(big.Float).SetString(amount)
	if !b {
		err = fmt.Errorf("error parsing backwards staked token")
		return
	}
	totalStakedToken.Add(totalStakedToken, backwardsStakedToken)

	block, err = getBlockByNumber(fmt.Sprintf("0x%x", blockNumber))
	if err != nil {
		log.Error("Failed to get block by number", "blockNumber", blockNumber, "error", err)
		return
	}
	if block == nil {
		err = fmt.Errorf("block not found")
		log.Error("Block not found", "blockNumber", blockNumber)
		return
	}
	blockTime = block.Time

	ipSentToZeroAddress = big.NewFloat(0)
	totalBalance := big.NewInt(0)
	for _, zeroAddress := range config.Conf.ZeroAddresses {
		var balance *big.Int
		balance, err = getBalanceAtBlock(zeroAddress, blockNumber)
		if err != nil {
			log.Error("Failed to get zero address balance", "address", zeroAddress, "error", err)
			return
		}
		totalBalance.Add(totalBalance, balance)
	}
	ipSentToZeroAddress = new(big.Float).Quo(new(big.Float).SetInt(totalBalance), big.NewFloat(1e18))
	return
}
