CREATE TABLE IF NOT EXISTS history_accumulated_fees (
    id                BIGSERIAL PRIMARY KEY,
    block_number            BIGINT NOT NULL,
    total_burnt_base_fee VARCHAR(64) NOT NULL,
    total_stake_reward  VARCHAR(64) NOT NULL,
    total_staked_token VARCHAR(64) NOT NULL,
    block_timestamp BIGINT NOT NULL
);


CREATE INDEX IF NOT EXISTS idx_history_accumulated_fees_block_number
    ON history_accumulated_fees (block_number);

CREATE INDEX IF NOT EXISTS idx_history_accumulated_fees_block_timestamp
    ON history_accumulated_fees (block_timestamp);
