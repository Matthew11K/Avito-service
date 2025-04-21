package handlers

import (
	"encoding/json"
	"net/http"
)

type Logger interface {
	Error(msg string, args ...any)
	Info(msg string, args ...any)
	Debug(msg string, args ...any)
}

type Error struct {
	Message string `json:"message"`
}

func respondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"Ошибка при формировании ответа"}`))

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(response)
}

func respondWithError(w http.ResponseWriter, status int, message string, err error, logger Logger) {
	if logger != nil {
		logger.Error(message, "error", err, "status", status)
	}

	errorResponse := Error{
		Message: message,
	}

	respondWithJSON(w, status, errorResponse)
}
