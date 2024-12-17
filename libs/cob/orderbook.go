package cob

import (
	"container/heap"
	"fmt"
	"sync"
)

// Order represents a single order in the order book
type Order struct {
	ID       int
	Price    float64
	Quantity int
	Next     *Order
	Prev     *Order
}

// OrderQueue represents orders at the same price level
type OrderQueue struct {
	Head *Order
	Tail *Order
	Size int
}

func NewOrderQueue() *OrderQueue {
	return &OrderQueue{}
}

// AddOrder adds an order to the queue
func (q *OrderQueue) AddOrder(order *Order) {
	if q.Head == nil {
		q.Head = order
		q.Tail = order
	} else {
		q.Tail.Next = order
		order.Prev = q.Tail
		q.Tail = order
	}
	q.Size++
}

// RemoveOrder removes an order from the queue
func (q *OrderQueue) RemoveOrder(order *Order) {
	if order.Prev != nil {
		order.Prev.Next = order.Next
	} else {
		q.Head = order.Next
	}
	if order.Next != nil {
		order.Next.Prev = order.Prev
	} else {
		q.Tail = order.Prev
	}
	q.Size--
}

type PriceHeap []float64

func (h PriceHeap) Len() int           { return len(h) }
func (h PriceHeap) Less(i, j int) bool { return h[i] < h[j] } // Min-Heap for asks
func (h PriceHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *PriceHeap) Push(x interface{}) {
	*h = append(*h, x.(float64))
}

func (h *PriceHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

type OrderBook struct {
	Bids       map[float64]*OrderQueue // Price → OrderQueue (HashMap)
	Asks       map[float64]*OrderQueue
	BidPrices  *PriceHeap // Max-Heap for bids
	AskPrices  *PriceHeap // Min-Heap for asks
	OrderMutex sync.Mutex
	OrderID    int
}

func NewOrderBook() *OrderBook {
	return &OrderBook{
		Bids:      make(map[float64]*OrderQueue),
		Asks:      make(map[float64]*OrderQueue),
		BidPrices: &PriceHeap{},
		AskPrices: &PriceHeap{},
	}
}

// PlaceOrder adds a new order to the order book
func (ob *OrderBook) PlaceOrder(price float64, quantity int, isBuy bool) int {
	ob.OrderMutex.Lock()
	defer ob.OrderMutex.Unlock()

	ob.OrderID++
	order := &Order{
		ID:       ob.OrderID,
		Price:    price,
		Quantity: quantity,
	}

	if isBuy {
		if _, exists := ob.Bids[price]; !exists {
			ob.Bids[price] = NewOrderQueue()
			heap.Push(ob.BidPrices, -price) // Use negative for max-heap
		}
		ob.Bids[price].AddOrder(order)
	} else {
		if _, exists := ob.Asks[price]; !exists {
			ob.Asks[price] = NewOrderQueue()
			heap.Push(ob.AskPrices, price)
		}
		ob.Asks[price].AddOrder(order)
	}
	return ob.OrderID
}

// MatchOrders matches buy and sell orders
func (ob *OrderBook) MatchOrders() {
	ob.OrderMutex.Lock()
	defer ob.OrderMutex.Unlock()

	for len(*ob.BidPrices) > 0 && len(*ob.AskPrices) > 0 {
		bestBid := -(*ob.BidPrices)[0] // Max bid price
		bestAsk := (*ob.AskPrices)[0]  // Min ask price

		if bestBid < bestAsk {
			break // No matching possible
		}

		bidQueue := ob.Bids[bestBid]
		askQueue := ob.Asks[bestAsk]

		bidOrder := bidQueue.Head
		askOrder := askQueue.Head

		matchQuantity := min(bidOrder.Quantity, askOrder.Quantity)

		fmt.Printf("Matched: Bid ID %d with Ask ID %d for %d units at price %.2f\n",
			bidOrder.ID, askOrder.ID, matchQuantity, bestAsk)

		// Update quantities
		bidOrder.Quantity -= matchQuantity
		askOrder.Quantity -= matchQuantity

		if bidOrder.Quantity == 0 {
			bidQueue.RemoveOrder(bidOrder)
		}
		if askOrder.Quantity == 0 {
			askQueue.RemoveOrder(askOrder)
		}

		// Clean up empty price levels
		if bidQueue.Size == 0 {
			delete(ob.Bids, bestBid)
			heap.Pop(ob.BidPrices)
		}
		if askQueue.Size == 0 {
			delete(ob.Asks, bestAsk)
			heap.Pop(ob.AskPrices)
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
