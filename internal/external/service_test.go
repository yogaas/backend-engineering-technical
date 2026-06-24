package external

import (
	"testing"
	"time"
)

func TestService_RetryEventuallySucceeds(t *testing.T) {
	client := &MockFlakyClient{FailFirstNCalls: 3}
	cb := NewCircuitBreaker(10, time.Second) 
	outbox := NewOutbox()
	svc := NewService(client, cb, outbox)

	svc.backoff = func(attempt int) time.Duration { return time.Millisecond }

	entry := &OutboxEntry{
		ID:            "OB-1",
		TransactionID: "TX-1",
		Payload:       Payload{TransactionID: "TX-1", Amount: 100},
	}
	svc.Enqueue(entry)

	got, _ := outbox.Get("OB-1")
	if got.Status != StatusSent {
		t.Fatalf("expected status sent after retries, got %s", got.Status)
	}
}

func TestService_DispatcherRetriesAfterCircuitRecovers(t *testing.T) {
	client := &MockFlakyClient{FailFirstNCalls: 100}
	cb := NewCircuitBreaker(2, 50*time.Millisecond)
	outbox := NewOutbox()
	svc := NewService(client, cb, outbox)
	svc.backoff = func(attempt int) time.Duration { return time.Millisecond }

	entry := &OutboxEntry{
		ID:            "OB-2",
		TransactionID: "TX-2",
		Payload:       Payload{TransactionID: "TX-2", Amount: 50},
	}
	svc.Enqueue(entry)

	got, _ := outbox.Get("OB-2")
	if got.Status == StatusSent {
		t.Fatalf("expected entry still pending while third-party is down")
	}

	client.FailFirstNCalls = 0

	stop := make(chan struct{})
	go svc.RunDispatcher(20*time.Millisecond, stop)
	defer close(stop)

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		got, _ = outbox.Get("OB-2")
		if got.Status == StatusSent {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	if got.Status != StatusSent {
		t.Fatalf("expected dispatcher to eventually send entry, final status: %s", got.Status)
	}
}
