package metrics

import (
    "net/http"
    "strconv"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
    HTTPRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Количество HTTP-запросов",
        },
        []string{"method", "path", "status"},
    )
    HTTPRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "Время обработки HTTP-запросов в секундах",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "path"},
    )

    PVZCreatedTotal = prometheus.NewCounter(
        prometheus.CounterOpts{
            Name: "pvz_created_total",
            Help: "Количество созданных ПВЗ",
        },
    )
    ReceptionsCreatedTotal = prometheus.NewCounter(
        prometheus.CounterOpts{
            Name: "receptions_created_total",
            Help: "Количество созданных приёмок заказов",
        },
    )
    ProductsCreatedTotal = prometheus.NewCounter(
        prometheus.CounterOpts{
            Name: "products_created_total",
            Help: "Количество добавленных товаров",
        },
    )
)

func init() {
    // регистрируем метрики в глобальном реестре
    prometheus.MustRegister(
        HTTPRequestsTotal,
        HTTPRequestDuration,
        PVZCreatedTotal,
        ReceptionsCreatedTotal,
        ProductsCreatedTotal,
    )
}

// middleware для Gin, считает кол‑во запросов и время
func GinMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        c.Next()
        duration := time.Since(start).Seconds()
        status := strconv.Itoa(c.Writer.Status())
        path := c.FullPath()
        if path == "" {
            path = c.Request.URL.Path
        }
        HTTPRequestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
        HTTPRequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
    }
}

func RunMetricsServer() {
    http.Handle("/metrics", promhttp.Handler())
    logErr := http.ListenAndServe(":9000", nil)
    if logErr != nil {
        panic("Metrics server failed: " + logErr.Error())
    }
}
