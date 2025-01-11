package cob

import (
	"container/heap"
	"sync"
)

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
	Bids *PriceLevelHeap
	Asks *PriceLevelHeap
	mu   sync.RWMutex
}

func NewOrderBook() *OrderBook {
	bids := &PriceLevelHeap{}
	asks := &PriceLevelHeap{}
	heap.Init(bids)
	heap.Init(asks)

	return &OrderBook{
		Bids: bids,
		Asks: asks,
		mu:   sync.RWMutex{},
	}
}

func (ob *OrderBook) UpdatePriceLevel(side string, price float64, quantity float64) {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	var heapToUpdate *PriceLevelHeap
	if side == "bid" {
		heapToUpdate = ob.Bids
	} else if side == "ask" {
		heapToUpdate = ob.Asks
	}

	// Check if the price level already exists
	for i, existing := range *heapToUpdate {
		if existing.Price == price {
			// Update the total quantity
			existing.TotalQuantity = quantity
			heap.Fix(heapToUpdate, i) // Restore heap property
			return
		}
	}

	// If the price level does not exist, add a new one
	newLevel := &PriceLevel{
		Price:         price,
		TotalQuantity: quantity,
		Orders:        &OrderQueue{}, // Initially empty
	}
	heap.Push(heapToUpdate, newLevel)
}

// // UpdatePriceLevel updates the specified price level in the order book.
// // It adds, modifies, or removes the price level based on the total quantity.
// func (ob *OrderBook) UpdatePriceLevel(side string, price float64) {
// 	priceLevels := ob.Bids
// 	if side == "sell" {
// 		priceLevels = ob.Asks
// 	}

// 	if pl, exists := priceLevels[price]; exists {
// 		if pl.TotalQuantity == 0 {
// 			// Remove the price level if no orders remain
// 			delete(priceLevels, price)
// 		}
// 	} else {
// 		// Add a new price level if it doesn't exist
// 		priceLevels[price] = &PriceLevel{
// 			Price:         price,
// 			TotalQuantity: 0,
// 			Orders:        &OrderQueue{},
// 		}
// 		heap.Init(priceLevels[price].Orders)
// 	}
// }

// // PlaceOrder adds an order to the order book.
// func (ob *OrderBook) PlaceOrder(order *Order) {
// 	priceLevels := ob.Bids
// 	if order.Side == "sell" {
// 		priceLevels = ob.Asks
// 	}

// 	// Ensure the price level exists
// 	if _, exists := priceLevels[order.Price]; !exists {
// 		priceLevels[order.Price] = &PriceLevel{
// 			Price:         order.Price,
// 			TotalQuantity: 0,
// 			Orders:        &OrderQueue{},
// 		}
// 		heap.Init(priceLevels[order.Price].Orders)
// 	}

// 	// Place the order in the priority queue
// 	priceLevel := priceLevels[order.Price]
// 	priceLevel.PlaceOrder(order) // Uses the PriceLevel.PlaceOrder method

// 	// Update the price level in the order book
// 	ob.UpdatePriceLevel(order.Side, order.Price)
// }

// // MatchOrder matches an incoming order against existing orders in the price level.
// // Returns the remaining unmatched quantity.
// func (pl *PriceLevel) MatchOrder(order *Order) float64 {
// 	remaining := order.Quantity

// 	for pl.Orders.Len() > 0 && remaining > 0 {
// 		// Peek the highest-priority order
// 		bestOrder := heap.Pop(pl.Orders).(*Order)

// 		if remaining >= bestOrder.Quantity {
// 			// Fully match the best order
// 			remaining -= bestOrder.Quantity
// 			pl.TotalQuantity -= bestOrder.Quantity
// 		} else {
// 			// Partially match the best order
// 			bestOrder.Quantity -= remaining
// 			pl.TotalQuantity -= remaining
// 			remaining = 0

// 			// Push the partially filled order back into the queue
// 			heap.Push(pl.Orders, bestOrder)
// 		}
// 	}

// 	return remaining
// }

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
