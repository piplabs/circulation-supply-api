package service

import (
	"circulation-supply-api/dao"
	"errors"
	"fmt"
	"math/big"

	"gorm.io/gorm"
)

func saveAccumulatedFees(block uint64, totalBurntBaseFee *big.Float, totalStakeReward *big.Float) error {
	return dao.SaveAccumulatedFees(block, totalBurntBaseFee.Text('f', -1), totalStakeReward.Text('f', -1))
}

func load() (uint64, *big.Float, *big.Float, error) {
	fee, err := dao.GetLatestAccumulatedFees()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, big.NewFloat(0).SetPrec(64), big.NewFloat(0).SetPrec(64), nil
		}
		return 0, big.NewFloat(0).SetPrec(64), big.NewFloat(0).SetPrec(64), err
	}
	totalBurntBaseFee, b := new(big.Float).SetString(fee.TotalBurntBaseFee)
	if !b {
		return 0, big.NewFloat(0).SetPrec(64), big.NewFloat(0).SetPrec(64), fmt.Errorf("error parsing total burnt base fee")
	}
	totalStakeReward, b := new(big.Float).SetString(fee.TotalStakeReward)
	if !b {
		return 0, big.NewFloat(0).SetPrec(64), big.NewFloat(0).SetPrec(64), fmt.Errorf("error parsing total stake reward")
	}
	return fee.BlockNumber + 1, totalBurntBaseFee, totalStakeReward, nil
}
