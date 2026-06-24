package external

import (
	"log"
	"math"
	"math/rand"
	"time"
)

const maxAttempts = 5

type Service struct {
	client  ThirdPartyClient
	cb      *CircuitBreaker
	outbox  *Outbox
	backoff func(attempt int) time.Duration
}

func NewService(client ThirdPartyClient, cb *CircuitBreaker, outbox *Outbox) *Service {
	return &Service{
		client: client,
		cb:     cb,
		outbox: outbox,
		backoff: func(attempt int) time.Duration {
			base := math.Pow(2, float64(attempt)) * 100 
			jitter := rand.Float64() * 100
			return time.Duration(base+jitter) * time.Millisecond
		},
	}
}

func (s *Service) Enqueue(entry *OutboxEntry) {
	s.outbox.Add(entry)
	s.attemptSend(entry)
}

func (s *Service) attemptSend(entry *OutboxEntry) {
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if !s.cb.Allow() {
			log.Printf("[external] circuit open, menunda pengiriman %s", entry.TransactionID)
			return 
		}

		s.outbox.IncrementAttempt(entry.ID)
		err := s.client.SendTransaction(entry.Payload)
		if err == nil {
			s.cb.RecordSuccess()
			s.outbox.MarkSent(entry.ID)
			log.Printf("[external] transaksi %s berhasil terkirim ke accounting service", entry.TransactionID)
			return
		}

		s.cb.RecordFailure()
		log.Printf("[external] gagal kirim %s (attempt %d/%d): %v", entry.TransactionID, attempt, maxAttempts, err)

		if attempt < maxAttempts {
			time.Sleep(s.backoff(attempt))
		}
	}

	log.Printf("[external] transaksi %s masih pending setelah %d attempt, akan dicoba ulang oleh dispatcher", entry.TransactionID, maxAttempts)
}

func (s *Service) RunDispatcher(interval time.Duration, stop <-chan struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			for _, entry := range s.outbox.PendingEntries() {
				s.attemptSend(entry)
			}
		}
	}
}
