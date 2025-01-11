package cob

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
