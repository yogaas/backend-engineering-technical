package datasync

import "sync"

type Service struct {
	mu   sync.Mutex
	data map[string]Availability
}

func NewService() *Service {
	return &Service{data: make(map[string]Availability)}
}


func (s *Service) ApplyUpdate(update Availability) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	current, exists := s.data[update.TicketID]
	if exists && update.Version <= current.Version {
		return false
	}
	s.data[update.TicketID] = update
	return true
}

func (s *Service) Get(ticketID string) (Availability, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	a, ok := s.data[ticketID]
	return a, ok
}
