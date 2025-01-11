package cob

type Order struct {
	ID        string
	Side      string
	Price     float64
	Quantity  float64
	Timestamp int64
	Provider  string // Use a string enum for provider ("kraken", "whitebit", "bybit", "local", etc.)
}
