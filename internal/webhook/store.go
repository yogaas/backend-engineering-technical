package webhook

import (
	"sync"
)

type PaymentRecord struct {
	IdempotencyKey string
	TransactionID  string
	Amount         float64
	Status         string
}

type PaymentStore struct {
	mu   sync.Mutex
	data map[string]PaymentRecord
}

func NewPaymentStore() *PaymentStore {
	return &PaymentStore{data: make(map[string]PaymentRecord)}
}

func (s *PaymentStore) SaveIfNotExists(rec PaymentRecord) (PaymentRecord, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if existing, ok := s.data[rec.IdempotencyKey]; ok {
		return existing, false, nil
	}
	s.data[rec.IdempotencyKey] = rec
	return rec, true, nil
}

func (s *PaymentStore) Count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.data)
}
