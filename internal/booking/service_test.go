package booking

import (
	"strconv"
	"sync"
	"testing"
)


func TestBookTicket_RaceCondition(t *testing.T) {
	repo := NewInMemoryTicketRepository()
	// mencoba memasukan 1 data tiket
	var ticketStock int32 = 5;
	repo.Seed(&Ticket{ID: "VIP-1", Name: "VIP", Stock: int(ticketStock)})
	svc := NewService(repo)

	// mengumpamakan ada 3 user request
	const concurrentUsers = 10
	var wg sync.WaitGroup
	var successCount int32
	var mu sync.Mutex

	wg.Add(concurrentUsers)
	for i := 0; i < concurrentUsers; i++ {
		go func(idx int) {
			defer wg.Done()
			_, err := svc.BookTicket("VIP-1", "user-" + strconv.Itoa(idx))
			if err == nil {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}(i)    
	}
	wg.Wait()

	// hanya untuk memastikan jika booking hanya terjadi sama dengan jumlah tiket
	if successCount != ticketStock {
		t.Fatalf("expected exactly 1 successful booking, got %d", successCount)
	}

	// pengecekan jika jumlah tiket tidak 0
	ticket, _ := repo.Get("VIP-1")
	if ticket.Stock != 0 {
		t.Fatalf("expected remaining stock 0, got %d", ticket.Stock)
	}
}
