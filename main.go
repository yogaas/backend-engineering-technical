package main

import (
	"log"
	"net/http"
	"strings"

	"backend-engineering-technical/internal/booking"
)

func main(){
	mux := http.NewServeMux()
	ticketRepo := booking.NewInMemoryTicketRepository()
	ticketRepo.Seed(&booking.Ticket{ID: "VIP-1", Name: "VIP Concert Ticket", Stock: 1})
	bookingSvc := booking.NewService(ticketRepo)
	bookingHandler := booking.NewHandler(bookingSvc)
	mux.HandleFunc("POST /api/v1/tickets/book", bookingHandler.Book)
	
	
	addr := ":8000"
	log.Printf("Running apps di http://localhost%s", addr)
	log.Println(strings.Repeat("-", 60))
	log.Fatal(http.ListenAndServe(addr, mux))
}