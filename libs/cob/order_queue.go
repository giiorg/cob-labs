package cob

import "container/heap"

// OrderQueue is a heap-based priority queue of Orders.
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

	// 3. Older orders are prioritized (FIFO).
	return oq[i].Timestamp < oq[j].Timestamp
}

func (oq *OrderQueue) Push(x interface{}) {
	*oq = append(*oq, x.(*Order))
}

func (oq *OrderQueue) Pop() interface{} {
	old := *oq
	n := len(old)
	item := old[n-1]
	*oq = old[0 : n-1]
	return item
}

func (oq *OrderQueue) RemoveByID(orderID string) *Order {
	for i, order := range *oq {
		if order.ID == orderID {
			removed := order
			*oq = append((*oq)[:i], (*oq)[i+1:]...)
			heap.Init(oq) // Reheapify after removal.
			return removed
		}
	}
	return nil
}
