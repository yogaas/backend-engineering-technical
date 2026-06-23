package booking

import (
	"fmt"
	"sync/atomic"
)

var txCounter int64

type Service struct {
	repo TicketRepository
}

func NewService(repo TicketRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) BookTicket(ticketID, userID string) (*Transaction, error) {
	ok, err := s.repo.DecrementStockAtomic(ticketID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrSoldOut
	}

	id := atomic.AddInt64(&txCounter, 1)
	return &Transaction{
		ID:       fmt.Sprintf("TX-%d", id),
		TicketID: ticketID,
		UserID:   userID,
	}, nil
}

