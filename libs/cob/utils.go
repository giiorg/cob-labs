package cob

import "container/heap"

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
