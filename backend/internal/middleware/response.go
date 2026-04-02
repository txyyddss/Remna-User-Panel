package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

// APIResponse is the standardized API response format
type APIResponse struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	RequestID string      `json:"request_id"`
}

// WriteJSON writes a JSON response with the given status code
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	resp := APIResponse{
		Code:      status,
		Message:   "success",
		Data:      data,
		RequestID: uuid.New().String(),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

// WriteError writes an error JSON response
func WriteError(w http.ResponseWriter, status int, message string) {
	resp := APIResponse{
		Code:      status,
		Message:   message,
		RequestID: uuid.New().String(),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

// WriteSuccess writes a success JSON response with 200 status
func WriteSuccess(w http.ResponseWriter, data interface{}) {
	WriteJSON(w, http.StatusOK, data)
}

// DecodeJSON decodes JSON request body into the given interface
func DecodeJSON(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}
