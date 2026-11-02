package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/forgeflow/usermanagement/internal/apperror"
	"github.com/forgeflow/usermanagement/internal/model"
)

type errorBody struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

type envelope struct {
	Success bool        `json:"success"`
	Data    any         `json:"data,omitempty"`
	Meta    *model.Meta `json:"meta,omitempty"`
	Error   *errorBody  `json:"error,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, body envelope) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeOK(w http.ResponseWriter, status int, data any, meta *model.Meta) {
	writeJSON(w, status, envelope{Success: true, Data: data, Meta: meta})
}

// writeError is the central error handler: it only ever emits sanitized messages.
func writeError(w http.ResponseWriter, err error) {
	var mbe *http.MaxBytesError
	if errors.As(err, &mbe) {
		writeJSON(w, http.StatusRequestEntityTooLarge, envelope{Error: &errorBody{
			Code: "PAYLOAD_TOO_LARGE", Message: "Request body is too large."}})
		return
	}
	ae := apperror.As(err)
	writeJSON(w, ae.HTTPStatus(), envelope{Error: &errorBody{Code: ae.Code, Message: ae.Message, Details: ae.Details}})
}
