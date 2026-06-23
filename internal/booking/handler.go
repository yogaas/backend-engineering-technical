package booking

import (
	"encoding/json"
	"errors"
	"net/http"

	"backend-engineering-technical/pkg/response"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

type bookRequest struct {
	TicketID string `json:"ticket_id"`
	UserID   string `json:"user_id"`
}

func (h *Handler) Book(w http.ResponseWriter, r *http.Request) {
	var req bookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Fail(w, http.StatusBadRequest, "payload tidak valid", err)
		return
	}
	if req.TicketID == "" || req.UserID == "" {
		response.Fail(w, http.StatusBadRequest, "ticket_id dan user_id wajib diisi", nil)
		return
	}

	tx, err := h.svc.BookTicket(req.TicketID, req.UserID)
	if err != nil {
		switch {
		case errors.Is(err, ErrSoldOut):
			response.Fail(w, http.StatusConflict, "gagal memesan tiket", err)
		case errors.Is(err, ErrTicketNotFound):
			response.Fail(w, http.StatusNotFound, "gagal memesan tiket", err)
		default:
			response.Fail(w, http.StatusInternalServerError, "terjadi kesalahan internal", err)
		}
		return
	}

	response.Success(w, http.StatusCreated, "tiket berhasil dipesan", tx)
}
