package booking

import "errors"


var (
	ErrTicketNotFound = errors.New("tiket tidak ditemukan")
	ErrSoldOut        = errors.New("tiket sudah habis terjual")
)

type Ticket struct {
	ID    string
	Name  string
	Stock int
}

type Transaction struct {
	ID       string
	TicketID string
	UserID   string
}
