package kraken_ws_client

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
)

type KrakenWsChannel string

const (
	TickerChannel     KrakenWsChannel = "ticker"
	InstrumentChannel KrakenWsChannel = "instrument"
	BalancesChannel   KrakenWsChannel = "balances"
	ExecutionsChannel KrakenWsChannel = "executions"
)

type SubscribeRequestParams struct {
	Channel      KrakenWsChannel `json:"channel"`
	EventTrigger string          `json:"event_trigger"`
	Symbol       []string        `json:"symbol"`
	Snapshot     bool            `json:"snapshot"`
}

type SubscribeRequest struct {
	Method string                 `json:"method"`
	Params SubscribeRequestParams `json:"params"`
}

type SubscribeRequestToPrivateParams struct {
	SubscribeRequestParams
	Token string `json:"token"`
}

type SubscribeRequestToPrivate struct {
	Method string                          `json:"method"`
	Params SubscribeRequestToPrivateParams `json:"params"`
}

type ResponseMessage struct {
	Channel  KrakenWsChannel `json:"channel"`
	Type     string          `json:"type"`
	Data     json.RawMessage `json:"data"`
	Sequence int64           `json:"sequence"`
}

type Ticker struct {
	Symbol    string  `json:"symbol"`
	Bid       float64 `json:"bid"`
	BidQty    float64 `json:"bid_qty"`
	Ask       float64 `json:"ask"`
	AskQty    float64 `json:"ask_qty"`
	Last      float64 `json:"last"`
	Volume    float64 `json:"volume"`
	VWAP      float64 `json:"vwap"`
	Low       float64 `json:"low"`
	High      float64 `json:"high"`
	Change    float64 `json:"change"`
	ChangePct float64 `json:"change_pct"`
}

func (t Ticker) GetBid() float64 {
	return t.Bid
}

func (t Ticker) GetAsk() float64 {
	return t.Ask
}

func (t Ticker) GetSymbol() string {
	return t.Symbol
}

type Asset struct {
	Borrowable       bool    `json:"borrowable"`
	CollateralValue  float64 `json:"collateral_value"`
	ID               string  `json:"id"`
	MarginRate       float64 `json:"margin_rate"`
	Precision        int     `json:"precision"`
	PrecisionDisplay int     `json:"precision_display"`
	Status           string  `json:"status"`
}

type Pair struct {
	Base               string  `json:"base"`
	Quote              string  `json:"quote"`
	CostMin            float64 `json:"cost_min"`
	CostPrecision      int     `json:"cost_precision"`
	HasIndex           bool    `json:"has_index"`
	MarginInitial      float64 `json:"margin_initial"`
	Marginable         bool    `json:"marginable"`
	PositionLimitLong  int     `json:"position_limit_long"`
	PositionLimitShort int     `json:"position_limit_short"`
	PriceIncrement     float64 `json:"price_increment"`
	PricePrecision     int     `json:"price_precision"`
	QtyIncrement       float64 `json:"qty_increment"`
	QtyMin             float64 `json:"qty_min"`
	QtyPrecision       int     `json:"qty_precision"`
	Status             string  `json:"status"`
	Symbol             string  `json:"symbol"`
	TickSize           float64 `json:"tick_size"` // Deprecated
}

type InstrumentData struct {
	Assets []Asset `json:"assets"`
	Pairs  []Pair  `json:"pairs"`
}

type BalanceAsset struct {
	Asset      string   `json:"asset"`
	AssetClass string   `json:"asset_class"`
	Balance    float64  `json:"balance"`
	Wallets    []Wallet `json:"wallets"`
}

type Wallet struct {
	Balance float64 `json:"balance"`
	Type    string  `json:"type"`
	ID      string  `json:"id"`
}

type BalanceSnapshot struct {
	Balances []BalanceAsset `json:"data"`
}

type LedgerTransaction struct {
	Asset      string    `json:"asset"`
	AssetClass string    `json:"asset_class"`
	Amount     float64   `json:"amount"`
	Balance    float64   `json:"balance"`
	Fee        float64   `json:"fee"`
	LedgerID   string    `json:"ledger_id"`
	RefID      string    `json:"ref_id"`
	Timestamp  time.Time `json:"timestamp"`
	Type       string    `json:"type"`
	Subtype    string    `json:"subtype"`
	Category   string    `json:"category"`
	WalletType string    `json:"wallet_type"`
	WalletID   string    `json:"wallet_id"`
}

type BalanceUpdate struct {
	Transactions []LedgerTransaction `json:"data"`
}

type TokenResponse struct {
	Error  []string `json:"error"`
	Result struct {
		Token string `json:"token"`
	} `json:"result"`
}

type KrakenWsClientConfigCredentials struct {
	ApiKey    string
	ApiSecret string
}

type KrakenWsClientConfig struct {
	Url         string
	Credentials *KrakenWsClientConfigCredentials
}

type KrakenWsClient struct {
	config    KrakenWsClientConfig
	Conn      *websocket.Conn
	isPrivate bool
	token     string
	Db        *pgxpool.Pool
}

func createSignature(urlPath string, data interface{}, secret string) (string, error) {
	var encodedData string

	switch v := data.(type) {
	case string:
		var jsonData map[string]interface{}
		if err := json.Unmarshal([]byte(v), &jsonData); err != nil {
			return "", err
		}
		encodedData = jsonData["nonce"].(string) + v
	case map[string]interface{}:
		dataMap := url.Values{}
		for key, value := range v {
			dataMap.Set(key, fmt.Sprintf("%v", value))
		}
		encodedData = v["nonce"].(string) + dataMap.Encode()
	default:
		return "", fmt.Errorf("invalid data type")
	}
	sha := sha256.New()
	sha.Write([]byte(encodedData))
	shaSum := sha.Sum(nil)

	message := append([]byte(urlPath), shaSum...)
	decodedSecret, _ := base64.StdEncoding.DecodeString(secret)
	mac := hmac.New(sha512.New, decodedSecret)
	mac.Write(message)
	macSum := mac.Sum(nil)
	sigDigest := base64.StdEncoding.EncodeToString(macSum)
	return sigDigest, nil
}

func getWebSocketToken(credentials KrakenWsClientConfigCredentials) (string, error) {
	nonce := fmt.Sprintf("%d", time.Now().UnixNano())
	data := "nonce=" + nonce

	payload := map[string]interface{}{
		"nonce": nonce,
	}

	signature, err := createSignature("/0/private/GetWebSocketsToken", payload, credentials.ApiSecret)
	if err != nil {
		return "", err
	}

	wsGetTokenUrl := os.Getenv("KRAKEN_REST_API_URL") + os.Getenv("KRAKEN_GET_WEBSOCKET_TOKEN_PATH")
	req, err := http.NewRequest("POST", wsGetTokenUrl, bytes.NewBufferString(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("API-Key", credentials.ApiKey)
	req.Header.Set("API-Sign", signature)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var tokenResponse TokenResponse
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return "", err
	}

	if len(tokenResponse.Error) > 0 {
		return "", fmt.Errorf("kraken API error: %v", tokenResponse.Error)
	}

	return tokenResponse.Result.Token, nil
}

func NewKrakenWsClient(config KrakenWsClientConfig) *KrakenWsClient {
	conn := reconnect(config.Url)
	krakenWsClient := KrakenWsClient{
		config: config,
		Conn:   conn,
	}

	if config.Credentials != nil {
		krakenWsClient.isPrivate = true
		token, err := getWebSocketToken(*config.Credentials)
		if err != nil {
			return nil
		}

		krakenWsClient.token = token
	}

	return &krakenWsClient
}

func reconnect(url string) *websocket.Conn {
	var wsConn *websocket.Conn
	var err error

	for {
		wsConn, _, err = websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			fmt.Printf("websocket connection failed: %v, retrying...\n", err)
			time.Sleep(5 * time.Second) // TODO: implement exponential backoff (?)
			continue
		}
		break
	}

	return wsConn
}

func (k *KrakenWsClient) Subscribe(paramsSet ...SubscribeRequestParams) (chan ResponseMessage, error) {
	responseMessages := make(chan ResponseMessage, 20)

	for _, params := range paramsSet {
		var subscribeRequest any

		if k.isPrivate {
			subscribeRequest = SubscribeRequestToPrivate{
				Method: "subscribe",
				Params: SubscribeRequestToPrivateParams{
					SubscribeRequestParams: params,
					Token:                  k.token,
				},
			}
		} else {
			subscribeRequest = SubscribeRequest{
				Method: "subscribe",
				Params: params,
			}
		}

		if err := k.Conn.WriteJSON(subscribeRequest); err != nil {
			return nil, err
		}
	}

	go func() {
		for {
			_, mesasge, err := k.Conn.ReadMessage()
			if err != nil {
				fmt.Printf("error reading message: %v..\n", err)
				continue
			}

			var responseMessage ResponseMessage
			if err = json.Unmarshal(mesasge, &responseMessage); err != nil {
				fmt.Printf("Error unmarshalling ticker message: %v\n", err)
				continue
			}

			responseMessages <- responseMessage
		}
	}()

	return responseMessages, nil
}
