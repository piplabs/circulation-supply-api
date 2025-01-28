package dao

type AccumulatedFees struct {
	ID                uint
	BlockNumber       uint64
	TotalBurntBaseFee string
	TotalStakeReward  string
}

func SaveAccumulatedFees(block uint64, totalBurntBaseFee string, totalStakeReward string) error {
	return db.Save(&AccumulatedFees{
		ID:                1,
		BlockNumber:       block,
		TotalBurntBaseFee: totalBurntBaseFee,
		TotalStakeReward:  totalStakeReward,
	}).Error
}

func GetLatestAccumulatedFees() (*AccumulatedFees, error) {
	var f AccumulatedFees
	err := db.Table("public.accumulated_fees").First(&f).Error
	return &f, err
}
