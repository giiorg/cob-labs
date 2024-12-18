package cob

import (
	"container/heap"
)

// Order represents a buy/sell order.
type Order struct {
	ID           string
	Side         string  // "buy" or "sell"
	Price        float64 // Price at which the order is placed
	Quantity     float64 // Quantity to buy/sell
	Timestamp    int64   // Unix time for FIFO ordering
	Provider     string  // "local" or external exchange name
	AvailableBal float64 // Balance available for external exchange orders
}

// OrderQueue represents a priority queue for orders within a price level.
type OrderQueue []*Order

func (oq OrderQueue) Len() int      { return len(oq) }
func (oq OrderQueue) Swap(i, j int) { oq[i], oq[j] = oq[j], oq[i] }
func (oq OrderQueue) Less(i, j int) bool {
	// Custom ordering logic:
	// 1. Local orders are prioritized over external.
	if oq[i].Provider == "local" && oq[j].Provider != "local" {
		return true
	}
	if oq[i].Provider != "local" && oq[j].Provider == "local" {
		return false
	}

	// 2. Higher volume orders are prioritized.
	if oq[i].Quantity != oq[j].Quantity {
		return oq[i].Quantity > oq[j].Quantity
	}

	// 3. Orders with sufficient available balance are prioritized.
	if oq[i].AvailableBal != oq[j].AvailableBal {
		return oq[i].AvailableBal > oq[j].AvailableBal
	}

	// 4. Older orders are prioritized (FIFO).
	return oq[i].Timestamp < oq[j].Timestamp
}

// Push pushes an order onto the queue.
func (oq *OrderQueue) Push(x interface{}) {
	*oq = append(*oq, x.(*Order))
}

// Pop removes and returns the highest-priority order.
func (oq *OrderQueue) Pop() interface{} {
	old := *oq
	n := len(old)
	item := old[n-1]
	*oq = old[0 : n-1]
	return item
}

// RemoveByID removes an order by ID and returns the removed order (if any).
func (oq *OrderQueue) RemoveByID(orderID string) *Order {
	for i, order := range *oq {
		if order.ID == orderID {
			// Remove the order from the slice and return it.
			removed := order
			*oq = append((*oq)[:i], (*oq)[i+1:]...)
			heap.Init(oq) // Reheapify after removal.
			return removed
		}
	}
	return nil
}

// PriceLevel represents a specific price level in the order book.
type PriceLevel struct {
	Price         float64     // Price for this level
	TotalQuantity float64     // Precomputed total quantity for this level
	Orders        *OrderQueue // Priority queue for orders
}

// AddOrder adds an order to the price level and updates TotalQuantity.
func (pl *PriceLevel) AddOrder(order *Order) {
	heap.Push(pl.Orders, order)
	pl.TotalQuantity += order.Quantity
}

// RemoveOrder removes an order and updates TotalQuantity.
func (pl *PriceLevel) RemoveOrder(orderID string) {
	removed := pl.Orders.RemoveByID(orderID)
	if removed != nil {
		pl.TotalQuantity -= removed.Quantity
	}
}

// UpdatePriceLevel recalculates the total quantity for the price level.
func (pl *PriceLevel) UpdatePriceLevel() {
	total := 0.0
	for _, order := range *pl.Orders {
		total += order.Quantity
	}
	pl.TotalQuantity = total
}

// PlaceOrder places an order in the appropriate price level.
func (pl *PriceLevel) PlaceOrder(order *Order) {
	heap.Push(pl.Orders, order)        // Add the order to the priority queue
	pl.TotalQuantity += order.Quantity // Update the total quantity
}

// PriceLevelHeap is a heap for managing price levels.
type PriceLevelHeap []*PriceLevel

func (h PriceLevelHeap) Len() int           { return len(h) }
func (h PriceLevelHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h PriceLevelHeap) Less(i, j int) bool { return h[i].Price < h[j].Price }

func (h *PriceLevelHeap) Push(x interface{}) {
	*h = append(*h, x.(*PriceLevel))
}

func (h *PriceLevelHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[0 : n-1]
	return item
}

// // OrderBook represents the central order book.
// type OrderBook struct {
// 	Bids     *PriceLevelHeap         // Max-Heap for bids
// 	Asks     *PriceLevelHeap         // Min-Heap for asks
// 	PriceMap map[float64]*PriceLevel // Price lookup for quick access
// 	mutex    sync.RWMutex
// }

// func NewOrderBook() *OrderBook {
// 	bids := &PriceLevelHeap{}
// 	asks := &PriceLevelHeap{}
// 	heap.Init(bids)
// 	heap.Init(asks)

// 	return &OrderBook{
// 		Bids:     bids,
// 		Asks:     asks,
// 		PriceMap: make(map[float64]*PriceLevel),
// 	}
// }

type OrderBook struct {
	Bids map[float64]*PriceLevel // Map of bid price levels
	Asks map[float64]*PriceLevel // Map of ask price levels
}

// NewOrderBook creates a new, empty order book.
func NewOrderBook() *OrderBook {
	return &OrderBook{
		Bids: make(map[float64]*PriceLevel),
		Asks: make(map[float64]*PriceLevel),
	}
}

// UpdatePriceLevel updates the specified price level in the order book.
// It adds, modifies, or removes the price level based on the total quantity.
func (ob *OrderBook) UpdatePriceLevel(side string, price float64) {
	priceLevels := ob.Bids
	if side == "sell" {
		priceLevels = ob.Asks
	}

	if pl, exists := priceLevels[price]; exists {
		if pl.TotalQuantity == 0 {
			// Remove the price level if no orders remain
			delete(priceLevels, price)
		}
	} else {
		// Add a new price level if it doesn't exist
		priceLevels[price] = &PriceLevel{
			Price:         price,
			TotalQuantity: 0,
			Orders:        &OrderQueue{},
		}
		heap.Init(priceLevels[price].Orders)
	}
}

// PlaceOrder adds an order to the order book.
func (ob *OrderBook) PlaceOrder(order *Order) {
	priceLevels := ob.Bids
	if order.Side == "sell" {
		priceLevels = ob.Asks
	}

	// Ensure the price level exists
	if _, exists := priceLevels[order.Price]; !exists {
		priceLevels[order.Price] = &PriceLevel{
			Price:         order.Price,
			TotalQuantity: 0,
			Orders:        &OrderQueue{},
		}
		heap.Init(priceLevels[order.Price].Orders)
	}

	// Place the order in the priority queue
	priceLevel := priceLevels[order.Price]
	priceLevel.PlaceOrder(order) // Uses the PriceLevel.PlaceOrder method

	// Update the price level in the order book
	ob.UpdatePriceLevel(order.Side, order.Price)
}

// MatchOrder matches an incoming order against existing orders in the price level.
// Returns the remaining unmatched quantity.
func (pl *PriceLevel) MatchOrder(order *Order) float64 {
	remaining := order.Quantity

	for pl.Orders.Len() > 0 && remaining > 0 {
		// Peek the highest-priority order
		bestOrder := heap.Pop(pl.Orders).(*Order)

		if remaining >= bestOrder.Quantity {
			// Fully match the best order
			remaining -= bestOrder.Quantity
			pl.TotalQuantity -= bestOrder.Quantity
		} else {
			// Partially match the best order
			bestOrder.Quantity -= remaining
			pl.TotalQuantity -= remaining
			remaining = 0

			// Push the partially filled order back into the queue
			heap.Push(pl.Orders, bestOrder)
		}
	}

	return remaining
}

// // CancelOrder removes an order from the order book.
// func (ob *OrderBook) CancelOrder(orderID string, price float64) {
// 	ob.mutex.Lock()
// 	defer ob.mutex.Unlock()

// 	if priceLevel, exists := ob.PriceMap[price]; exists {
// 		priceLevel.RemoveOrder(orderID)
// 		if priceLevel.TotalQuantity == 0 {
// 			delete(ob.PriceMap, price)
// 			if priceLevel.Price == price {
// 				removeFromHeap(ob.Bids, priceLevel)
// 				removeFromHeap(ob.Asks, priceLevel)
// 			}
// 		}
// 	}
// }

// Utility functions.
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func oppositeSide(side string) string {
	if side == "buy" {
		return "sell"
	}
	return "buy"
}

func removeFromHeap(h *PriceLevelHeap, pl *PriceLevel) {
	for i, item := range *h {
		if item == pl {
			heap.Remove(h, i)
			break
		}
	}
}
