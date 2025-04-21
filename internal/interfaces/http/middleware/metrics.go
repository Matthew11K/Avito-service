package middleware

import (
	"net/http"
	"time"

	"avito/internal/metrics"
)

func Metrics() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rw := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			metrics.RequestsTotal.Inc()

			next.ServeHTTP(rw, r)

			duration := time.Since(start)
			metrics.ResponseTime.Observe(duration.Seconds())

			metrics.ResponseStatus.WithLabelValues(http.StatusText(rw.statusCode)).Inc()
		})
	}
}
