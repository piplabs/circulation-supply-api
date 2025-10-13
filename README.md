# Circulating Supply API

This repo is for tracing circulating/total IP supply.
The supply is calculated based on a combination of on-chain metrics and the off-chain vesting schedule. 

# API Interface

| Entrypoint            | Action | Request parameters | Response example                     | Comment                                             | 
|------------------------|--------|--------------------|--------------------------------------|-----------------------------------------------------|
| /circulating-supply    | GET    | nil                | {"result":"24900000.687"}            | return the circulating IP supply in string format   |
| /total-supply          | GET    | nil                | {"result":"9900000.687"}             | return the total IP supply in string format         |
| /cs                    | GET    | nil                | 24900001                             | return the circulating IP supply in whole number format | 
| /ts                    | GET    | nil                | 9900001                              | return the total IP supply in whole number format   | 
| /estimate-supply?date=2025-10-13       | GET    | date(format:yyyy-mm-dd) | {"result":"24900000.687"}            | return the estimated circulating IP supply at the given date in format yyyy-mm-dd|
| /supplydelta?from=2025-10-13&to=2025-11-13 | GET    | from(format:yyyy-mm-dd), to(format:yyyy-mm-dd) | {"startTime":"2025-10-13","endTime":"2025-11-13","totalDelta":"12345.67","inflationDelta":"2345.67","vestingDelta":"10000.00"} | return the supply delta between two dates, including total delta, inflation delta and vesting delta |

# Sources

- **Vested Supply**: Based on the official vesting plan.
- **Burnt IP**: Includes gas burnt and IP sent to the zero address.
- **Minted IP**: Rewards from staking 

# Detailed Design

## Services

In order to trace the circulating supply, we need to run two services:
1. **Scanner Service**: This service is responsible for iterating through the blockchain, calculating values like accumulated burnt IP and minted IP, and saving these values to a database.
2. **API Service**: This service provides an HTTP interface for users to query the circulating supply.

### Infra setup

As our test shows,  both service consume few resources. 
Recommend: 1 core CPU &  256m MEM 


### Scanner Service

#### description
1. Single instance
2. Index blocks and calculate accumulated burnt IP and minted IP
    
#### entry point
    go run /circulation-supply-api/scanner/main.go --config=/circulation-supply-api/config/mainnet.yaml

#### flow
1. Fetch info of last iterated block from database. Iterate from last iterated block. If no such block, iterate from genesis block.
2. Call `eth_getBlockByNumber` . The response contains `baseFeePerGas` , `gasUsed`, `withdrawals`
3.  `burntBaseFee`  = `baseFeePerGas` * `gasUsed`
    `stakeReward`  =  sum of `withdrawal.amount` ( `withdrawal.Validator` MUST equal "0x1" or "0x2" )
    `stakedToken` = sum of transfers that transfer over 1024 IP from Stake contract to Stake receiver contract
4. Save blockNumber & totalBurntBaseFee & totalStakedToken & totalStakeReward into database every 100 blocks.

### API Service

#### description
1. Multiple instances
2. Provide HTTP interfaces for users to query circulating and total supply.

#### entry point
    go run /circulation-supply-api/api/main.go --config=/circulation-supply-api/config/mainnet.yaml

#### flow
1. `/circulating-supply`  & `/cs`
    a. Get latest block from database
    b.  Get totalBurntBaseFee, totalStakedToken and totalStakeReward
    c.  Get timestamp of the latest block
    d.  Get the balance of the zero address at the latest block
    e.  Get accumulated vested supply from vesting plan
    f.  Calculate the total circulating supply = accumulated vested supply - totalBurntBaseFee - IPSentToZeroAddress + totalStakedToken + totalStakeReward
2. `/total-supply`  & `/cs`
    a. Get latest block from database
    b.  Get totalBurntBaseFee, totalStakedToken and totalStakeReward
    c.  Get timestamp of the latest block
    d.  Get the balance of the zero address at the latest block
    e. Get genesis total supply from vesting plan
    f. Calculate the total supply = genesis total supply - totalBurntBaseFee - IPSentToZeroAddress +totalStakedToken + totalStakeReward
3. `/estimate-supply`
    a. Get the date parameter from request
    b. Convert date to timestamp
    c. Return error if the date is before time.now
    d. Get the vested supply from vesting plan based on the number of months passed
    e. Get the recent record within 12 hours 
    f. Calculate the estimated minted tokens based on the number of blocks passed since the recent record
    g. Calculate the estimated circulating supply = vested supply - currentBurnt + estimatedMinted

## Notes
1. How to estimate minted tokens:
    - Reward per block: 1.929 (20,000,000/10,368,000)
    - The number of blocks after the singularity: CurrentBlockNumber-1,580,851, where 1,580,851 is the singularity height
    - Total reward = 1.929 * (CurrentBlockNumber-1,580,851)

2. API service will output key information in the log, for example: 
```
2025/07/30 05:02:06 INFO Show details -  vestedSupply=284760417 totalBurntBaseFee=12.313564381151027358 IPSentToZeroAddress=535319295.33350471519 totalStakedToken=535318403.28723771457 totalStakeReward=10248222.271689286413 circulatingSupply=295007734.91185790463 monthsPassed=5

```

Above log shows the following information on 2025/07/30 05:02:06:
- vested supply = 284,760,417
- burnt IP = 12.313564 + 892.046267
    - baseFeeBurnt = 12.313564
    - sendToZeroAddressBurnt = 892.046267
- minted = 10,248,222.271689

Thus circulating supply = 284,760,417 - 12.313564 - 892.046267 + 535,318,403.28723771457 + 10,248,222.271689286413 = 295,007,734.911858
