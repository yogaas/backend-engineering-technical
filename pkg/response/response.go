// Package response menyediakan standardized response envelope
// agar seluruh endpoint memiliki format yang konsisten.
package response

import (
	"encoding/json"
	"net/http"
)

// Envelope adalah bentuk standar response untuk semua API.
type Envelope struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// JSON menulis envelope sebagai JSON response dengan status code tertentu.
func JSON(w http.ResponseWriter, status int, success bool, message string, data interface{}, errMsg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Envelope{
		Success: success,
		Message: message,
		Data:    data,
		Error:   errMsg,
	})
}

// Success menulis response sukses standar.
func Success(w http.ResponseWriter, status int, message string, data interface{}) {
	JSON(w, status, true, message, data, "")
}

// Fail menulis response gagal standar.
func Fail(w http.ResponseWriter, status int, message string, err error) {
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	JSON(w, status, false, message, nil, errMsg)
}
