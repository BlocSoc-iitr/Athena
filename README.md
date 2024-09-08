# [Athena](https://github.com/BlocSoc-iitr/Athena/tree/main/athena)

Go package for interacting with starknet. It is a tool which can be used in cli for the following tasks

This is inspired from nethermind/entro but implemented in go and is faster than entro u can see the backfill time taking by both the code bases by this graph.
We used goroutines and worker pools + we coded it in golang to make it faster.


![Graph](https://github.com/user-attachments/assets/f051e42b-9676-48fc-9144-ea4d36bec75b)







## Tasks it can perform

- Backfill full blocks including events , tx & receipts
- Get implementation history
- ABI Decoding
- Decoding of functions & events from a transaction



# Usage/Examples

## Specify this in the following fields
from :- block number 

to :- block number 

jsonRpcUrl - https://rpc.nethermind.io/mainnet-juno?x-apikey=YOUR_API_KEY

output :- your_file_name.csv


## [ ABI Decoder](https://github.com/BlocSoc-iitr/Athena/blob/main/athena/decoder/abi_decoder.go)
abi parser , to parse each event and function ie to get its name and parameters and their datatype to help us in decoding 

```bash
go run cli/get/starknet.go --classHash 0x01a736d6ed154502257f02b1ccdf4d9d1089f80811cd6acad48e6b6a9d1f2003 --jsonRpcUrl "https://rpc.nethermind.io/mainnet-juno?x-apikey=MIkLH4AOTdTH9uqu8PqvSHUBNnAnMU1fXdROa3qc1DsSVxvOcGRrwr6kSj1zsNjT" --decode
```
Result - list of functions and events decoded in that ABI to your console. 
![WhatsApp Image 2024-09-08 at 16 35 46_eddf7e15](https://github.com/user-attachments/assets/03e55c8f-b7d6-4b60-81cf-316b2489b175)


## [ Event Decoder](https://github.com/BlocSoc-iitr/Athena/blob/main/athena/decoder/event_decoder.go)
To decode an event , user provides the contract hash and the name of the event and block range , firstly the abi is fetched and parsed and the data for the particular event and block range is fetched 
The data fetched contains the keys and data required to decode the event , the decoder then initiates a new decoder class for that particular event and the event’s data is decoded with the provided inputs of hash,block range and event name


```bash
go run cli/get/dispatcher.go -event TransactionExecuted -contract 0x005a708f9c84bc709e967086572c6655e2b85eaf5a2ef752d92e24e64c5e392c_1 -from 691000 -to 692000
```
Result - 


## [Function Decoder](https://github.com/BlocSoc-iitr/Athena/blob/main/athena/decoder/function_decoder.go)
Similar as above but instead of keys and data pairs we give call data as input that we will fetch in the provided blockrange



## [Backfill Events](https://github.com/BlocSoc-iitr/Athena/tree/main/athena/backfill/importers)

From - Block Number

To - Block Number


```bash
go run cli/backfill/starknet_events.go --from 67800 --to 67801 --rpc-url https://starknet-mainnet.public.blastapi.io/rpc/v0_7 --output events.csv --chunk-size 100
```

[With Filters](https://github.com/BlocSoc-iitr/Athena/blob/main/athena/backfill/filters.go)
```bash
go run cli/backfill/starknet_filters.go -rpc-url "https://rpc.nethermind.io/mainnet-juno?x-apikey=MIkLH4AOTdTH9uqu8PqvSHUBNnAnMU1fXdROa3qc1DsSVxvOcGRrwr6kSj1zsNjT" -contract-address "0x01a736d6ed154502257f02b1ccdf4d9d1089f80811cd6acad48e6b6a9d1f2003" -from 100000 -to 200000 -output "filtered_events.csv"
```

Result - all events decoded in your excel file

![WhatsApp Image 2024-09-08 at 16 35 46_05437c77](https://github.com/user-attachments/assets/3e2b20cd-f70f-4792-bcf2-dceb7e11d1fe)


## [Backfill full blocks](https://github.com/BlocSoc-iitr/Athena/blob/main/athena/backfill/importers/starknet.go) 

```bash
go run starknet.go --from 110 --to 130 --rpc-url https://rpc.nethermind.io/mainnet-juno?x-apikey=MIkLH4AOTdTH9uqu8PqvSHUBNnAnMU1fXdROa3qc1DsSVxvOcGRrwr6kSj1zsNjT --output blocks_details.csv --transactionhash (for block data)
```
Result - all blocks , transaction hashes , receipts & events of that txns data in your csv files.  

## [Backfill with filters](https://github.com/BlocSoc-iitr/Athena/blob/main/athena/backfill/filters.go)



![WhatsApp Image 2024-09-08 at 16 36 25_74cfd3e6](https://github.com/user-attachments/assets/ea0d53a9-42ab-4394-9963-b059f9b40ed6)
