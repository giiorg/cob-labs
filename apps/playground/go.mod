module cob/playground

replace bitnet/kraken_market_data => ../../libs/kraken_market_data

replace bitnet/kraken_ws_client => ../../libs/kraken_ws_client

go 1.23.1

require (
	bitnet/kraken_market_data v0.0.0-00010101000000-000000000000
	github.com/joho/godotenv v1.5.1
	github.com/nats-io/nats-server/v2 v2.10.23
	github.com/nats-io/nats.go v1.37.0
)

require (
	bitnet/kraken_ws_client v0.0.0-00010101000000-000000000000 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.7.1 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/minio/highwayhash v1.0.3 // indirect
	github.com/nats-io/jwt/v2 v2.5.8 // indirect
	github.com/nats-io/nkeys v0.4.8 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	golang.org/x/crypto v0.30.0 // indirect
	golang.org/x/sync v0.10.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	golang.org/x/time v0.8.0 // indirect
)
