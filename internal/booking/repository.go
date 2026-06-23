package booking

import "sync"


type TicketRepository interface {
	Get(id string) (*Ticket, error)
	DecrementStockAtomic(id string) (bool, error)
	Seed(t *Ticket)
}

type inMemoryTicket struct {
	mu     sync.Mutex
	ticket Ticket
}

type InMemoryTicketRepository struct {
	mapMu sync.RWMutex 
	data  map[string]*inMemoryTicket
}

func NewInMemoryTicketRepository() *InMemoryTicketRepository {
	return &InMemoryTicketRepository{
		data: make(map[string]*inMemoryTicket),
	}
}

func (r *InMemoryTicketRepository) Seed(t *Ticket) {
	r.mapMu.Lock()
	defer r.mapMu.Unlock()
	r.data[t.ID] = &inMemoryTicket{ticket: *t}
}

func (r *InMemoryTicketRepository) Get(id string) (*Ticket, error) {
	r.mapMu.RLock()
	entry, ok := r.data[id]
	r.mapMu.RUnlock()
	if !ok {
		return nil, ErrTicketNotFound
	}

	entry.mu.Lock()
	defer entry.mu.Unlock()
	cp := entry.ticket
	return &cp, nil
}

func (r *InMemoryTicketRepository) DecrementStockAtomic(id string) (bool, error) {
	r.mapMu.RLock()
	entry, ok := r.data[id]
	r.mapMu.RUnlock()
	if !ok {
		return false, ErrTicketNotFound
	}

	entry.mu.Lock()
	defer entry.mu.Unlock()

	if entry.ticket.Stock <= 0 {
		return false, nil
	}
	entry.ticket.Stock--
	return true, nil
}
