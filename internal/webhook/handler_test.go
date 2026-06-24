package webhook

import (
	"sync"
	"testing"
)

func TestSaveIfNotExists_ConcurrentDuplicates(t *testing.T) {
	store := NewPaymentStore()

	rec := PaymentRecord{
		IdempotencyKey: "idem-key-123",
		TransactionID:  "TX-99",
		Amount:         150000,
		Status:         "PAID",
	}

	const concurrentRetries = 10
	var wg sync.WaitGroup
	var newCount int32
	var mu sync.Mutex

	wg.Add(concurrentRetries)
	for i := 0; i < concurrentRetries; i++ {
		go func() {
			defer wg.Done()
			_, isNew, _ := store.SaveIfNotExists(rec)
			if isNew {
				mu.Lock()
				newCount++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	if newCount != 1 {
		t.Fatalf("expected exactly 1 request treated as new, got %d", newCount)
	}
	if store.Count() != 1 {
		t.Fatalf("expected exactly 1 record stored, got %d", store.Count())
	}
}
