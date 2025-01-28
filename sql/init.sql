CREATE TABLE IF NOT EXISTS accumulated_fees (
    id                BIGSERIAL PRIMARY KEY,
    block_number            BIGINT NOT NULL,
    total_burnt_base_fee VARCHAR(64) NOT NULL,
    total_stake_reward  VARCHAR(64) NOT NULL
);