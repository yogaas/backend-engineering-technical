package ingestion

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync/atomic"

	"backend-engineering-technical/pkg/response"
)

type Service struct {
	queue   *Queue
	store   Store
	counter int64
}

func NewService(queue *Queue, store Store) *Service {
	return &Service{queue: queue, store: store}
}

func (s *Service) Submit(userID string, amount float64) (Job, error) {
	id := atomic.AddInt64(&s.counter, 1)
	job := Job{ID: fmt.Sprintf("ING-%d", id), UserID: userID, Amount: amount}
	if err := s.queue.Enqueue(job); err != nil {
		return Job{}, err
	}
	return job, nil
}

func (s *Service) Status(id string) (Job, bool) {
	return s.store.Get(id)
}

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

type submitRequest struct {
	UserID string  `json:"user_id"`
	Amount float64 `json:"amount"`
}

func (h *Handler) Submit(w http.ResponseWriter, r *http.Request) {
	var req submitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Fail(w, http.StatusBadRequest, "payload tidak valid", err)
		return
	}

	job, err := h.svc.Submit(req.UserID, req.Amount)
	if err != nil {
		if errors.Is(err, ErrQueueFull) {
			response.Fail(w, http.StatusServiceUnavailable, "sistem sedang sibuk", err)
			return
		}
		response.Fail(w, http.StatusInternalServerError, "terjadi kesalahan internal", err)
		return
	}

	response.Success(w, http.StatusAccepted, "transaksi diterima dan akan diproses", job)
}

func (h *Handler) Status(w http.ResponseWriter, r *http.Request, id string) {
	job, ok := h.svc.Status(id)
	if !ok {
		response.Fail(w, http.StatusNotFound, "status transaksi belum tersedia (masih diproses atau tidak ditemukan)", nil)
		return
	}
	response.Success(w, http.StatusOK, "status transaksi ditemukan", job)
}
