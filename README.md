# Circulating Supply API

This repo is for querying circulating IP supply.

# API Interface

- URL: GET `/circulating-supply`
- Description:  return the circulating IP supply at a relatively recent block height
- Request: nil
- Response: {”result”:”10000000”} // unit is IP

# Detailed Design

## Infra setup

As local test shows,  both service consume few resources. 
Recommend: 1 core CPU &  256m MEM 

## Services

### Scanner Service

- description
    1. Single instance
    2. Index blocks and calculate accumulated burntBaseFee and stakeReward
    
- config
    
    rpcEndpoints
    
- flow
    1. Fetch info of last iterated block from database. Iterate from last iterated block. If no such block, iterate from genesis block.
    2. Call `eth_getBlockByNumber` . The response contains `baseFeePerGas` , `gasUsed`, `withdrawals`
    3.  `burntBaseFee`  = `baseFeePerGas` * `gasUsed`
        
        `stakeReward`  =  sum of `withdrawal.amount` ( `withdrawal.Validator` MUST equal “0x1” or “0x2” )
        
    4. Save blockNumber & totalBurntBaseFee & totalStakeReward into a snapshot file every 100 blocks.

### API Service

- description
    1. Multiple instances
    2. Provide an HTTP interface `/circulating-supply` for users to query circulation volume.

- config
    
    rpcEndpoints
    
    zero addresses
    
    vesting schedule
    
- flow
    1. User calls API `/circulating-supply`
        1. If last call is within 5 seconds, return cached result
        2. get `totalBurntBaseFee` and `totalStakeReward` and `blockNumber` from last iterated block from database
        3. fetch `IPSentToZeroAddresses` in last iterated block through JSON RPC eth_getBalance
        4. get `blockTimestamp`  by `blockNumber`  through JSON RPC eth_getBlockByNumber
        5. calculate year & month of `blockTimestamp` → get `vestedCirculatingSupply` from configuration
        6. `CirculatingSupply` = `vestedCirculatingSupply` - (`totalBurntBaseFee` + `IPSentToZeroAddresses`) + `totalStakeReward`
        7. cache and return
