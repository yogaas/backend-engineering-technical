package external

import "sync"

type OutboxStatus string

const (
	StatusPending OutboxStatus = "pending"
	StatusSent    OutboxStatus = "sent"
	StatusFailed  OutboxStatus = "failed"
)

type OutboxEntry struct {
	ID            string
	TransactionID string
	Payload       Payload
	Status        OutboxStatus
	Attempts      int
}

type Outbox struct {
	mu      sync.Mutex
	entries map[string]*OutboxEntry
}

func NewOutbox() *Outbox {
	return &Outbox{entries: make(map[string]*OutboxEntry)}
}

func (o *Outbox) Add(entry *OutboxEntry) {
	o.mu.Lock()
	defer o.mu.Unlock()
	entry.Status = StatusPending
	o.entries[entry.ID] = entry
}

func (o *Outbox) Get(id string) (*OutboxEntry, bool) {
	o.mu.Lock()
	defer o.mu.Unlock()
	e, ok := o.entries[id]
	return e, ok
}

func (o *Outbox) PendingEntries() []*OutboxEntry {
	o.mu.Lock()
	defer o.mu.Unlock()
	var result []*OutboxEntry
	for _, e := range o.entries {
		if e.Status == StatusPending {
			result = append(result, e)
		}
	}
	return result
}

func (o *Outbox) MarkSent(id string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if e, ok := o.entries[id]; ok {
		e.Status = StatusSent
	}
}

func (o *Outbox) MarkFailed(id string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if e, ok := o.entries[id]; ok {
		e.Status = StatusFailed
	}
}

func (o *Outbox) IncrementAttempt(id string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if e, ok := o.entries[id]; ok {
		e.Attempts++
	}
}
