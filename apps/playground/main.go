package main

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/joho/godotenv"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"

	krakenMarketDataProvider "bitnet/kraken_market_data"
)

func runEmbeddedNatsServer(inProcess bool, enableLogging bool) (*server.Server, error) {
	opts := &server.Options{
		ServerName:      "embedded_server",
		DontListen:      inProcess,
		JetStream:       true,
		JetStreamDomain: "embedded",
	}

	natsServer, err := server.NewServer(opts)
	if err != nil {
		return nil, err
	}

	if enableLogging {
		natsServer.ConfigureLogger()
	}

	go natsServer.Start()
	if !natsServer.ReadyForConnections(5 * time.Second) {
		return nil, errors.New("NATS server timeout")
	}

	return natsServer, nil
}

func runEmbeddedNatsClient(inProcess bool, natsServer *server.Server) (*nats.Conn, error) {
	clientOpts := []nats.Option{}
	if inProcess {
		clientOpts = append(clientOpts, nats.InProcessServer(natsServer))
	}

	natsClient, err := nats.Connect(natsServer.ClientURL(), clientOpts...)
	if err != nil {
		return nil, err
	}

	return natsClient, nil
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("error loading .env file: %+v", err)
	}

	natsServer, err := runEmbeddedNatsServer(true, true)
	if err != nil {
		log.Fatal(err)
	}

	natsClient1, err := runEmbeddedNatsClient(true, natsServer)
	if err != nil {
		log.Fatal(err)
	}

	natsClient1.Subscribe("market.*.BTC/USDT", func(msg *nats.Msg) {
		log.Printf("got data: %+v\n", string(msg.Data))
		msg.Respond([]byte("Hello there"))
	})

	natsClient2, err := runEmbeddedNatsClient(true, natsServer)
	if err != nil {
		log.Fatal(err)
	}

	krakenMarketDataProvider := krakenMarketDataProvider.New(natsClient2)
	// runKrakenWs(ctx, enabledPairs, cacheManager)
	krakenMarketDataProvider.Run(ctx, []string{"BTC/USDT"})

	natsServer.WaitForShutdown()
}
