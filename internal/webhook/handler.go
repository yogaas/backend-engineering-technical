package webhook

import (
	"encoding/json"
	"net/http"

	"backend-engineering-technical/pkg/response"
)

type Handler struct {
	store *PaymentStore
}

func NewHandler(store *PaymentStore) *Handler {
	return &Handler{store: store}
}

type paymentWebhookRequest struct {
	IdempotencyKey string  `json:"idempotency_key"`
	TransactionID  string  `json:"transaction_id"`
	Amount         float64 `json:"amount"`
	Status         string  `json:"status"`
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	var req paymentWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Fail(w, http.StatusBadRequest, "payload tidak valid", err)
		return
	}
	if req.IdempotencyKey == "" {
		response.Fail(w, http.StatusBadRequest, "idempotency_key wajib diisi", nil)
		return
	}

	rec := PaymentRecord{
		IdempotencyKey: req.IdempotencyKey,
		TransactionID:  req.TransactionID,
		Amount:         req.Amount,
		Status:         req.Status,
	}

	saved, isNew, err := h.store.SaveIfNotExists(rec)
	if err != nil {
		response.Fail(w, http.StatusInternalServerError, "terjadi kesalahan internal", err)
		return
	}

	if !isNew {
		response.Success(w, http.StatusOK, "webhook sudah pernah diterima sebelumnya (duplicate, diabaikan)", saved)
		return
	}

	response.Success(w, http.StatusCreated, "data payment berhasil disimpan", saved)
}
