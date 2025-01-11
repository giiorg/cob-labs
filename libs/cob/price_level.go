package cob

import "container/heap"

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
