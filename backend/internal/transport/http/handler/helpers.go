package handler

import (
	"encoding/json"
	"net/http"

	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/transport/http/httpctx"
	"gitlab.com/libs-artifex/wrapper/v2"
)

// ReadJSON читает JSON из тела запроса.
func ReadJSON(r *http.Request, v any) error {
	err := json.NewDecoder(r.Body).Decode(v)
	if err != nil {
		return wrapper.Wrap(err)
	}

	return nil
}

// WriteJSON пишет JSON-ответ.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		// игнорируем ошибку записи ответа (стандартная практика для хелперов)
		_ = err
	}
}

// GetUserIDFromContext — заглушка (потом будет из JWT).
func GetUserIDFromContext(r *http.Request) domain.UUID {
	userID, ok := httpctx.UserID(r.Context())
	if !ok {
		return domain.UUID{}
	}

	return userID
}
