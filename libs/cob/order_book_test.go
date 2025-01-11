package cob

// func TestUpdatePriceLevel(t *testing.T) {
// 	orderBook := NewOrderBook()

// 	// Update price levels
// 	orderBook.UpdatePriceLevel("bid", 100.0, 10.0) // Add new bid level
// 	orderBook.UpdatePriceLevel("bid", 101.0, 20.0) // Add new bid level
// 	orderBook.UpdatePriceLevel("ask", 102.0, 15.0) // Add new ask level
// 	orderBook.UpdatePriceLevel("ask", 103.0, 25.0) // Add new ask level

// 	// Update existing price level
// 	orderBook.UpdatePriceLevel("bid", 100.0, 15.0) // Update quantity at bid 100.0

// 	// Assertions for bids
// 	if len(*orderBook.Bids) != 2 {
// 		t.Errorf("Expected 2 bid levels, got %d", len(*orderBook.Bids))
// 	}
// 	if (*orderBook.Bids)[0].Price != 100.0 || (*orderBook.Bids)[0].TotalQuantity != 15.0 {
// 		t.Errorf("Expected bid totalQty at price 100.0 to be 15.0, got %.2f", (*orderBook.Bids)[0].TotalQuantity)
// 	}

// 	// Assertions for asks
// 	if len(*orderBook.Asks) != 2 {
// 		t.Errorf("Expected 2 ask levels, got %d", len(*orderBook.Asks))
// 	}
// 	if (*orderBook.Asks)[0].Price != 102.0 || (*orderBook.Asks)[0].TotalQuantity != 15.0 {
// 		t.Errorf("Expected ask totalQty at price 102.0 to be 15.0, got %.2f", (*orderBook.Asks)[0].TotalQuantity)
// 	}
// }

// func TestParallelUpdatePriceLevel(t *testing.T) {
// 	ob := NewOrderBook()
// 	var wg sync.WaitGroup

// 	// Simulated updates from multiple providers
// 	providers := []struct {
// 		provider  string
// 		side      string
// 		price     float64
// 		totalQty  float64
// 		newOrders []*Order
// 	}{
// 		{"Kraken", "bid", 100.0, 10.0, []*Order{{ID: "k1", Side: "buy", Price: 100.0, Quantity: 10.0, Provider: "Kraken"}}},
// 		{"Bybit", "ask", 101.0, 15.0, []*Order{{ID: "b1", Side: "sell", Price: 101.0, Quantity: 15.0, Provider: "Bybit"}}},
// 		{"Whitebit", "bid", 100.0, 5.0, []*Order{{ID: "w1", Side: "buy", Price: 100.0, Quantity: 5.0, Provider: "Whitebit"}}},
// 		{"Kraken", "ask", 101.0, 5.0, []*Order{{ID: "k2", Side: "sell", Price: 101.0, Quantity: 5.0, Provider: "Kraken"}}},
// 	}

// 	// Run updates in parallel
// 	for _, update := range providers {
// 		wg.Add(1)
// 		go func(update struct {
// 			provider  string
// 			side      string
// 			price     float64
// 			totalQty  float64
// 			newOrders []*Order
// 		}) {
// 			defer wg.Done()
// 			pl := &PriceLevel{
// 				Price:         update.price,
// 				TotalQuantity: update.totalQty,
// 				Orders:        &OrderQueue{},
// 			}
// 			for _, order := range update.newOrders {
// 				pl.AddOrder(order)
// 			}
// 			ob.UpdatePriceLevel(update.side, pl)
// 		}(update)
// 	}

// 	wg.Wait()

// 	// Validate the final state of the order book
// 	expectedBids := map[float64]float64{
// 		100.0: 15.0, // Consolidated from Kraken (10) and Whitebit (5)
// 	}
// 	expectedAsks := map[float64]float64{
// 		101.0: 20.0, // Consolidated from Bybit (15) and Kraken (5)
// 	}

// 	for price, qty := range expectedBids {
// 		found := false
// 		for _, pl := range *ob.Bids {
// 			if pl.Price == price {
// 				found = true
// 				if pl.TotalQuantity != qty {
// 					t.Errorf("Expected bid totalQty at price %f to be %f, got %f", price, qty, pl.TotalQuantity)
// 				}
// 				break
// 			}
// 		}
// 		if !found {
// 			t.Errorf("Expected bid price level %f not found", price)
// 		}
// 	}

// 	for price, qty := range expectedAsks {
// 		found := false
// 		for _, pl := range *ob.Asks {
// 			if pl.Price == price {
// 				found = true
// 				if pl.TotalQuantity != qty {
// 					t.Errorf("Expected ask totalQty at price %f to be %f, got %f", price, qty, pl.TotalQuantity)
// 				}
// 				break
// 			}
// 		}
// 		if !found {
// 			t.Errorf("Expected ask price level %f not found", price)
// 		}
// 	}
// }
