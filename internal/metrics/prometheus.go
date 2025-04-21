package metrics

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type PrometheusMetrics struct {
	RequestsTotal      prometheus.Counter
	RequestDuration    prometheus.Histogram
	PVZsCreated        prometheus.Counter
	ReceptionsCreated  prometheus.Counter
	ProductsAddedTotal prometheus.Counter
}

func NewPrometheusMetrics(namespace string) *PrometheusMetrics {
	return &PrometheusMetrics{
		RequestsTotal: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "requests_total",
			Help:      "Общее количество запросов",
		}),
		RequestDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "request_duration_seconds",
			Help:      "Продолжительность запросов в секундах",
			Buckets:   prometheus.DefBuckets,
		}),

		PVZsCreated: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "pvzs_created_total",
			Help:      "Общее количество созданных ПВЗ",
		}),
		ReceptionsCreated: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "receptions_created_total",
			Help:      "Общее количество созданных приемок",
		}),
		ProductsAddedTotal: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "products_added_total",
			Help:      "Общее количество добавленных товаров",
		}),
	}
}

func StartServer(addr string) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		_ = server.ListenAndServe()
	}()

	return server
}

var (
	RequestsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "app_requests_total",
		Help: "Общее количество HTTP запросов",
	})

	ResponseTime = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "app_response_time_seconds",
		Help:    "Время ответа в секундах",
		Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 2},
	})

	ResponseStatus = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "app_response_status_total",
			Help: "Количество HTTP ответов по статусам",
		},
		[]string{"status"},
	)

	PVZCreatedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "app_pvz_created_total",
		Help: "Общее количество созданных ПВЗ",
	})

	ReceptionsCreatedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "app_receptions_created_total",
		Help: "Общее количество созданных приемок",
	})

	ProductsAddedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "app_products_added_total",
		Help: "Общее количество добавленных товаров",
	})
)
