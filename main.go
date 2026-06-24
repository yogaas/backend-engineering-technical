package main

import (
	"log"
	"net/http"
	"strings"

	"backend-engineering-technical/internal/booking"
	"backend-engineering-technical/internal/ingestion"
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
	
	
	addr := ":8000"
	log.Printf("Running apps di http://localhost%s", addr)
	log.Println(strings.Repeat("-", 60))
	log.Fatal(http.ListenAndServe(addr, mux))
}