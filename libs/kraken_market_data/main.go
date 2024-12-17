package kraken_market_data

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	krakenwsclient "bitnet/kraken_ws_client"

	// "github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
)

type KrakenMarketDataProvider struct {
	natsClient *nats.Conn
}

func New(natsClient *nats.Conn) *KrakenMarketDataProvider {
	return &KrakenMarketDataProvider{
		natsClient: natsClient,
	}
}

func (k *KrakenMarketDataProvider) Run(ctx context.Context, enabledPairs []string) {
	config := krakenwsclient.KrakenWsClientConfig{
		Url: os.Getenv("KRAKEN_PUBLIC_WS_URL"),
	}
	krakenWsClient := krakenwsclient.NewKrakenWsClient(config)

	updates, err := krakenWsClient.Subscribe(
		krakenwsclient.SubscribeRequestParams{
			Channel:      "ticker",
			EventTrigger: "bbo",
			Symbol:       enabledPairs,
			Snapshot:     true,
		},
		krakenwsclient.SubscribeRequestParams{
			Channel:  krakenwsclient.InstrumentChannel,
			Snapshot: true,
		},
	)
	if err != nil {
		log.Printf("can not subscribe to ticker channel: %v\n", err)
	}

	for update := range updates {
		switch update.Channel {
		case krakenwsclient.TickerChannel:
			var tickersData []krakenwsclient.Ticker
			if err = json.Unmarshal(update.Data, &tickersData); err != nil {
				log.Printf("error unmarshalling ticker message: %v\n", err)
				continue
			}

			for _, ticker := range tickersData {
				encoded, err := json.Marshal(ticker)
				if err != nil {
					log.Printf("failed to marshal ticker %+v: %+v\n", ticker, err)
				} else {
					err := k.natsClient.Publish(fmt.Sprintf("market.kraken.%s", ticker.Symbol), encoded)
					if err != nil {
						log.Printf("enable to publish ticker %+v: %+v\n", ticker, err)
					}
				}
			}
		case krakenwsclient.InstrumentChannel:
			var instrumentData krakenwsclient.InstrumentData
			if err = json.Unmarshal(update.Data, &instrumentData); err != nil {
				log.Printf("error unmarshalling ticker message: %v\n", err)
				continue
			}

			for _, pair := range instrumentData.Pairs {
				pairBytes := new(bytes.Buffer)
				err := json.NewEncoder(pairBytes).Encode(pair)
				if err != nil {
					log.Printf("failed to marshal pair info %+v: %+v\n", pair, err)
				} else {
					err := k.natsClient.Publish(fmt.Sprintf("market_info.kraken.%s", pair.Symbol), pairBytes.Bytes())
					if err != nil {
						log.Printf("enable to publish pair info %+v: %+v\n", pair, err)
					}
				}
			}
		default:
			//
		}
	}
}

func getEnabledPairsFromEnv() []string {
	enabledPairsStr := os.Getenv("ENABLED_PAIRS")

	return strings.Split(enabledPairsStr, ",")
}

// func main() {
// 	ctx, cancel := context.WithCancel(context.Background())
// 	defer cancel()

// 	err := godotenv.Load()
// 	if err != nil {
// 		log.Fatalf("error loading .env file: %+v", err)
// 	}

// 	cacheManager := cachemanager.NewCacheManager(os.Getenv("REDIS_ADDRESS"))

// 	enabledPairs := getEnabledPairsFromEnv()

// 	runKrakenWs(ctx, enabledPairs, cacheManager)
// }
