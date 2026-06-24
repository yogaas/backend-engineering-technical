package datasync

import (
	"encoding/json"
	"net/http"

	"backend-engineering-technical/pkg/response"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

type updateRequest struct {
	TicketID string `json:"ticket_id"`
	Quantity int    `json:"quantity"`
	Version  int64  `json:"version"`
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	var req updateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Fail(w, http.StatusBadRequest, "payload tidak valid", err)
		return
	}

	applied := h.svc.ApplyUpdate(Availability{
		TicketID: req.TicketID,
		Quantity: req.Quantity,
		Version:  req.Version,
	})

	current, _ := h.svc.Get(req.TicketID)

	if !applied {
		response.Success(w, http.StatusOK, "update diabaikan karena versi lebih lama dari data saat ini (out-of-order)", current)
		return
	}
	response.Success(w, http.StatusOK, "update berhasil diterapkan", current)
}
