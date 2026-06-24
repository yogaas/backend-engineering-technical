package ingestion

import (
	"sync"
	"testing"
	"time"
)

func TestIngestion_HighTraffic(t *testing.T) {
	store := NewInMemoryStore()
	queue := NewQueue(20000, 50, store) // buffer besar + 50 worker
	queue.Start()
	defer queue.Stop()

	svc := NewService(queue, store)

	const totalTx = 10000
	var wg sync.WaitGroup
	var enqueueFailed int32
	var enqueueSuccess int32
	var mu sync.Mutex

	wg.Add(totalTx)
	for i := 0; i < totalTx; i++ {
		go func(idx int) {
			defer wg.Done()
			_, err := svc.Submit("user", float64(idx))
			if err != nil {
				mu.Lock()
				enqueueFailed++
				mu.Unlock()
			}else{
				mu.Lock()
				enqueueSuccess++
				mu.Unlock()
			}
		}(i)
	}
	wg.Wait()

	if enqueueFailed > 0 {
		t.Fatalf("expected 0 enqueue failures with sufficient buffer, got %d", enqueueFailed)
	}
	
	if enqueueSuccess > 0 {
		t.Logf("Berhasil, got %d", enqueueSuccess)
	}

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		store.mu.RLock()
		count := len(store.data)
		store.mu.RUnlock()
		if count == totalTx {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	store.mu.RLock()
	finalCount := len(store.data)
	store.mu.RUnlock()

	if finalCount != totalTx {
		t.Fatalf("expected all %d transactions saved, got %d (dead-letter: %d)",
			totalTx, finalCount, queue.DeadLetterCount())
	}
}
