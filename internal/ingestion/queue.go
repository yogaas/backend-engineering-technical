package ingestion

import (
	"errors"
	"log"
	"sync"
)

var ErrQueueFull = errors.New("antrian penuh, silakan coba lagi")

type Job struct {
	ID     string
	UserID string
	Amount float64
}

type Store interface {
	Save(j Job) error
	Get(id string) (Job, bool)
}

type InMemoryStore struct {
	mu   sync.RWMutex
	data map[string]Job
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{data: make(map[string]Job)}
}

func (s *InMemoryStore) Save(j Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[j.ID] = j
	return nil
}

func (s *InMemoryStore) Get(id string) (Job, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	j, ok := s.data[id]
	return j, ok
}

type Queue struct {
	ch          chan Job
	store       Store
	workerCount int
	wg          sync.WaitGroup

	deadLetterMu sync.Mutex
	deadLetter   []Job 
}

func NewQueue(bufferSize, workerCount int, store Store) *Queue {
	return &Queue{
		ch:          make(chan Job, bufferSize),
		store:       store,
		workerCount: workerCount,
	}
}

func (q *Queue) Start() {
	for i := 0; i < q.workerCount; i++ {
		q.wg.Add(1)
		go q.worker(i)
	}
}

func (q *Queue) Stop() {
	close(q.ch)
	q.wg.Wait()
}

func (q *Queue) Enqueue(j Job) error {
	select {
	case q.ch <- j:
		return nil
	default:
		return ErrQueueFull
	}
}

const maxRetry = 3

func (q *Queue) worker(id int) {
	defer q.wg.Done()
	for job := range q.ch {
		var lastErr error
		for attempt := 1; attempt <= maxRetry; attempt++ {
			if err := q.store.Save(job); err != nil {
				lastErr = err
				log.Printf("[worker-%d] gagal simpan job %s (attempt %d): %v", id, job.ID, attempt, err)
				continue
			}
			lastErr = nil
			break
		}
		if lastErr != nil {
			q.deadLetterMu.Lock()
			q.deadLetter = append(q.deadLetter, job)
			q.deadLetterMu.Unlock()
		}
	}
}

func (q *Queue) DeadLetterCount() int {
	q.deadLetterMu.Lock()
	defer q.deadLetterMu.Unlock()
	return len(q.deadLetter)
}
