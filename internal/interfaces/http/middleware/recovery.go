package middleware

import (
	"net/http"
	"runtime/debug"
)

func Recovery(logger Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("Внутренняя ошибка сервера",
						"error", err,
						"stack", string(debug.Stack()),
						"path", r.URL.Path,
						"method", r.Method,
					)

					w.WriteHeader(http.StatusInternalServerError)
					w.Header().Set("Content-Type", "application/json")
					w.Write([]byte(`{"message":"Внутренняя ошибка сервера"}`))
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
