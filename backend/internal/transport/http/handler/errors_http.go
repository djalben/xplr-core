package handler

import (
	"net/http"

	"gitlab.com/libs-artifex/wrapper/v2"
)

// WrapAndWriteError wraps internal error (for logging/trace) but never exposes it to client.
func WrapAndWriteError(w http.ResponseWriter, err error, status int, publicMessage string) {
	if err != nil {
		_ = wrapper.Wrap(err)
	}

	http.Error(w, publicMessage, status)
}

func WriteBadRequest(w http.ResponseWriter, publicMessage string) {
	http.Error(w, publicMessage, http.StatusBadRequest)
}

func WriteInternalServerError(w http.ResponseWriter, err error) {
	WrapAndWriteError(w, err, http.StatusInternalServerError, "Ошибка сервера. Попробуйте позже.")
}

