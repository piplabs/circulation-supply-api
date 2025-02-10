# Circulating Supply API

This repo is for querying circulating/total IP supply.

# API Interface

## Get circulating supply

- URL: GET `/circulating-supply`
- Description:  return the circulating IP supply at a relatively recent block height
- Request: nil
- Response: {”result”:”10000000”} // unit is IP

## Get total supply

- URL: GET `/total-supply`
- Description:  return the total IP supply at a relatively recent block height
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
    
- entry point
    go run /circulation-supply-api/scanner/main.go --config=/circulation-supply-api/config/odyssey.yaml

- config
    1.rpcEndpoints

    
- flow
    1. Fetch info of last iterated block from database. Iterate from last iterated block. If no such block, iterate from genesis block.
    2. Call `eth_getBlockByNumber` . The response contains `baseFeePerGas` , `gasUsed`, `withdrawals`
    3.  `burntBaseFee`  = `baseFeePerGas` * `gasUsed`
        
        `stakeReward`  =  sum of `withdrawal.amount` ( `withdrawal.Validator` MUST equal "0x1" or "0x2" or "0x3" )
        
    4. Save blockNumber & totalBurntBaseFee & totalStakeReward into a snapshot file every 100 blocks.

### API Service

- description
    1. Multiple instances
    2. Provide an HTTP interface `/circulating-supply` for users to query circulation volume.

- entry point
    go run /circulation-supply-api/api/main.go --config=/circulation-supply-api/config/odyssey.yaml

- config
    1. rpcEndpoints
    2. zero addresses
    3. vesting schedule
    4. vesting start year
    5. vesting start month
    6. genesis total supply
    
- flow
    1. User calls API `/circulating-supply` or `/total-supply`
    2. If last call is within 30 seconds, return cached result
    3. get `totalBurntBaseFee` and `totalStakeReward` and `blockNumber` from last iterated block from database
    4. fetch `IPSentToZeroAddresses` in last iterated block through JSON RPC eth_getBalance
    5. get `blockTimestamp`  by `blockNumber`  through JSON RPC eth_getBlockByNumber
    7. compute result
        - if `/circulating-supply`:  calculate year & month of `blockTimestamp`, get `vestedCirculatingSupply` from configuration according to year & month, then result=  `vestedCirculatingSupply` - (`totalBurntBaseFee` + `IPSentToZeroAddresses`) + `totalStakeReward`
        - if `/total-supply`: get `genesisTotalSupply` from configuration, then result = `genesisTotalSupply` - (`totalBurntBaseFee` + `IPSentToZeroAddresses`) + `totalStakeReward`
