package cob

import (
	"container/heap"
	"testing"
	"time"
)

func TestOrderQueue_Push(t *testing.T) {
	oq := &OrderQueue{}
	heap.Init(oq)

	order1 := &Order{ID: "1", Provider: "local", Quantity: 100, Timestamp: time.Now().Unix()}
	order2 := &Order{ID: "2", Provider: "kraken", Quantity: 120, Timestamp: time.Now().Unix()}
	heap.Push(oq, order1)
	heap.Push(oq, order2)

	if oq.Len() != 2 {
		t.Errorf("Expected queue length to be 2, but got %d", oq.Len())
	}

	order3 := &Order{ID: "3", Provider: "local", Quantity: 110, Timestamp: time.Now().Unix()}
	heap.Push(oq, order3)

	if oq.Len() != 3 {
		t.Errorf("Expected queue length to be 3, but got %d", oq.Len())
	}
}

func TestOrderQueue_Pop(t *testing.T) {
	oq := &OrderQueue{}
	heap.Init(oq)

	order1 := &Order{ID: "1", Provider: "kraken", Quantity: 100, Timestamp: time.Now().Unix()}
	order2 := &Order{ID: "2", Provider: "local", Quantity: 120, Timestamp: time.Now().Unix()}
	order3 := &Order{ID: "3", Provider: "local", Quantity: 110, Timestamp: time.Now().Unix()}
	heap.Push(oq, order1)
	heap.Push(oq, order2)
	heap.Push(oq, order3)

	popped1 := heap.Pop(oq).(*Order)
	if popped1.ID != order2.ID {
		t.Errorf("Expected first popped order to be order with ID '2', but got '%s'", popped1.ID)
	}
	if oq.Len() != 2 {
		t.Errorf("Expected queue length to be 2, but got %d", oq.Len())
	}

	popped2 := heap.Pop(oq).(*Order)
	if popped2.ID != order3.ID {
		t.Errorf("Expected second popped order to be order with ID '3', but got '%s'", popped2.ID)
	}
	if oq.Len() != 1 {
		t.Errorf("Expected queue length to be 1, but got %d", oq.Len())
	}

	popped3 := heap.Pop(oq).(*Order)
	if popped3.ID != order1.ID {
		t.Errorf("Expected third popped order to be order with ID '1', but got '%s'", popped3.ID)
	}

	if oq.Len() != 0 {
		t.Errorf("Expected queue length to be 0, but got %d", oq.Len())
	}
}

func TestOrderQueue_RemoveByID(t *testing.T) {
	oq := &OrderQueue{}
	heap.Init(oq)

	order1 := &Order{ID: "1", Provider: "local", Quantity: 100, Timestamp: time.Now().Unix()}
	order2 := &Order{ID: "2", Provider: "kraken", Quantity: 120, Timestamp: time.Now().Unix()}
	order3 := &Order{ID: "3", Provider: "local", Quantity: 110, Timestamp: time.Now().Unix()}

	heap.Push(oq, order1)
	heap.Push(oq, order2)
	heap.Push(oq, order3)

	removed := oq.RemoveByID("2")

	if removed == nil || removed.ID != "2" {
		t.Errorf("Expected to remove order with ID '2' and get it, but got %+v", removed)
	}

	if oq.Len() != 2 {
		t.Errorf("Expected queue length to be 2, but got %d", oq.Len())
	}

	removed = oq.RemoveByID("4")
	if removed != nil {
		t.Errorf("Expected not to remove order with ID '4' and get nil, but got %+v", removed)
	}

	if oq.Len() != 2 {
		t.Errorf("Expected queue length to be 2, but got %d", oq.Len())
	}
}

func TestOrderQueue_Ordering(t *testing.T) {
	oq := &OrderQueue{}
	heap.Init(oq)

	// Orders to verify ordering rules:

	localOrder1 := &Order{ID: "local1", Provider: "local", Quantity: 100, Timestamp: time.Now().Unix() - 3}
	localOrder2 := &Order{ID: "local2", Provider: "local", Quantity: 120, Timestamp: time.Now().Unix() - 2}
	localOrder3 := &Order{ID: "local3", Provider: "local", Quantity: 120, Timestamp: time.Now().Unix() - 1}
	externalOrder1 := &Order{ID: "external1", Provider: "kraken", Quantity: 120, Timestamp: time.Now().Unix() - 4}
	externalOrder2 := &Order{ID: "external2", Provider: "kraken", Quantity: 100, Timestamp: time.Now().Unix() - 5}
	externalOrder3 := &Order{ID: "external3", Provider: "kraken", Quantity: 120, Timestamp: time.Now().Unix() - 6}
	heap.Push(oq, externalOrder1)
	heap.Push(oq, localOrder1)
	heap.Push(oq, externalOrder2)
	heap.Push(oq, localOrder2)
	heap.Push(oq, externalOrder3)
	heap.Push(oq, localOrder3)

	expectedOrder := []*Order{localOrder3, localOrder2, localOrder1, externalOrder1, externalOrder2, externalOrder3}

	for i := 0; i < len(expectedOrder); i++ {
		popped := heap.Pop(oq).(*Order)
		found := false
		for j, expected := range expectedOrder {
			if popped.ID == expected.ID {
				//Remove expected from the expected order
				expectedOrder = append(expectedOrder[:j], expectedOrder[j+1:]...)
				found = true
				break
			}
		}

		if !found {
			t.Errorf("Expected popped order %d to be one of the expected orders, but got %+v", i, popped)
		}
	}
}
