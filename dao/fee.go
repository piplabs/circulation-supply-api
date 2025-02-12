package dao

type AccumulatedFees struct {
	ID                uint
	BlockNumber       uint64
	TotalBurntBaseFee string
	TotalStakeReward  string
	TotalStakedToken  string
}

type BackwardsStakedToken struct {
	ID          uint
	BlockNumber uint64
	Amount      string
}

func SaveAccumulatedFees(block uint64, totalBurntBaseFee string, totalStakeReward string, totalStakedToken string) error {
	return db.Save(&AccumulatedFees{
		ID:                1,
		BlockNumber:       block,
		TotalBurntBaseFee: totalBurntBaseFee,
		TotalStakeReward:  totalStakeReward,
		TotalStakedToken:  totalStakedToken,
	}).Error
}

func GetLatestAccumulatedFees() (*AccumulatedFees, error) {
	var f AccumulatedFees
	err := db.Table("public.accumulated_fees").First(&f).Error
	return &f, err
}

func SaveBackwardsStakedToken(block uint64, amount string) error {
	return db.Table("public.backwards_staked_token").Save(&BackwardsStakedToken{
		ID:          1,
		BlockNumber: block,
		Amount:      amount,
	}).Error
}

func GetBackwardsStakedToken() (uint64, string, error) {
	var f BackwardsStakedToken
	err := db.Table("public.backwards_staked_token").First(&f).Error
	return f.BlockNumber, f.Amount, err
}
