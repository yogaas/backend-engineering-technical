package main

import (
	"log"
	"net/http"
	"strings"
	"time"

	"backend-engineering-technical/internal/booking"
	"backend-engineering-technical/internal/external"
	"backend-engineering-technical/internal/ingestion"
	"backend-engineering-technical/internal/webhook"
)

func main(){
	mux := http.NewServeMux()
	
	// Section 1
	ticketRepo := booking.NewInMemoryTicketRepository()
	ticketRepo.Seed(&booking.Ticket{ID: "VIP-1", Name: "VIP Concert Ticket", Stock: 1})
	bookingSvc := booking.NewService(ticketRepo)
	bookingHandler := booking.NewHandler(bookingSvc)
	mux.HandleFunc("POST /api/v1/tickets/book", bookingHandler.Book)
	
	// Section 2
	ingestionStore := ingestion.NewInMemoryStore()
	queue := ingestion.NewQueue(20000, 50, ingestionStore)
	queue.Start()
	ingestionSvc := ingestion.NewService(queue, ingestionStore)
	ingestionHandler := ingestion.NewHandler(ingestionSvc)
	mux.HandleFunc("POST /api/v1/transactions", ingestionHandler.Submit)
	mux.HandleFunc("GET /api/v1/transactions/{id}/status", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		ingestionHandler.Status(w, r, id)
	})
	
	// Section 3
	thirdPartyClient := &external.MockFlakyClient{FailFirstNCalls: 2} // simulasi 2x gagal lalu sukses
	circuitBreaker := external.NewCircuitBreaker(5, 10*time.Second)
	outbox := external.NewOutbox()
	externalSvc := external.NewService(thirdPartyClient, circuitBreaker, outbox)
	stopDispatcher := make(chan struct{})
	go externalSvc.RunDispatcher(5*time.Second, stopDispatcher)

	// Endpoint demo: simulasikan "transaksi sukses" yang lalu dikirim ke accounting service
	mux.HandleFunc("POST /api/v1/transactions/{id}/send-to-accounting", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		entry := &external.OutboxEntry{
			ID:            "OB-" + id,
			TransactionID: id,
			Payload:       external.Payload{TransactionID: id, Amount: 100},
		}
		externalSvc.Enqueue(entry)
		got, _ := outbox.Get(entry.ID)
		w.Write([]byte("status pengiriman ke accounting service: " + string(got.Status) + "\n"))
	})
	
	// Section 4
	paymentStore := webhook.NewPaymentStore()
	webhookHandler := webhook.NewHandler(paymentStore)
	mux.HandleFunc("POST /api/v1/webhook/payment", webhookHandler.Handle)
	
	
	addr := ":8000"
	log.Printf("Running apps di http://localhost%s", addr)
	log.Println(strings.Repeat("-", 60))
	log.Fatal(http.ListenAndServe(addr, mux))
}