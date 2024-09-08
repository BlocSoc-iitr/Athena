go run starknet.go --from 110 --to 130 --rpc-url  https://free-rpc.nethermind.io/mainnet-juno/ --output blocks_details.csv
(if the file name is main.go)
go run starknet.go --from 110 --to 130 --rpc-url  https://rpc.nethermind.io/mainnet-juno?x-apikey=MIkLH4AOTdTH9uqu8PqvSHUBNnAnMU1fXdROa3qc1DsSVxvOcGRrwr6kSj1zsNjT --output blocks_details.csv --transactionhash
(for block data)
go run starknet_events.go --from 67800 --to 67801 --rpc-url https://starknet-mainnet.public.blastapi.io/rpc/v0_7 --output events.csv --chunk-size 100