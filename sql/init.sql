CREATE TABLE IF NOT EXISTS accumulated_fees (
    id                BIGSERIAL PRIMARY KEY,
    block_number            BIGINT NOT NULL,
    total_burnt_base_fee VARCHAR(64) NOT NULL,
    total_stake_reward  VARCHAR(64) NOT NULL
);

ALTER TABLE accumulated_fees ADD COLUMN total_staked_token VARCHAR(64);

CREATE TABLE IF NOT EXISTS backwards_staked_token (
    id                BIGSERIAL PRIMARY KEY,
    block_number            BIGINT NOT NULL,
    amount  VARCHAR(64) NOT NULL
);