package main

import (
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Product struct {
	ProductID    *int   `json:"product_id"`
	SKU          string `json:"sku"`
	Manufacturer string `json:"manufacturer"`
	CategoryID   *int   `json:"category_id"`
	Weight       *int   `json:"weight"`
	SomeOtherID  *int   `json:"some_other_id"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

var (
	mu       sync.RWMutex
	products = map[int]Product{}
	logger   = slog.New(slog.NewJSONHandler(os.Stdout, nil))

	reqTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "http_requests_total", Help: "Total HTTP requests"},
		[]string{"method", "endpoint", "status"},
	)
	errTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "http_errors_total", Help: "Total non-2xx responses"},
		[]string{"method", "endpoint", "status"},
	)
	reqDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{Name: "http_request_duration_seconds", Help: "Request duration", Buckets: prometheus.DefBuckets},
		[]string{"method", "endpoint"},
	)
)

func init() {
	prometheus.MustRegister(reqTotal, errTotal, reqDuration)
}

func metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		endpoint := c.FullPath()
		if endpoint == "" {
			endpoint = "unknown"
		}
		status := c.Writer.Status()
		statusStr := strconv.Itoa(status)
		dur := time.Since(start).Seconds()
		reqTotal.WithLabelValues(c.Request.Method, endpoint, statusStr).Inc()
		reqDuration.WithLabelValues(c.Request.Method, endpoint).Observe(dur)
		if status >= 400 {
			errTotal.WithLabelValues(c.Request.Method, endpoint, statusStr).Inc()
		}
		logger.Info("request",
			"method", c.Request.Method,
			"endpoint", endpoint,
			"status", status,
			"duration_ms", dur*1000,
		)
	}
}

func sendErr(c *gin.Context, code int, errCode, msg, details string) {
	c.JSON(code, ErrorResponse{Error: errCode, Message: msg, Details: details})
}

func getProduct(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("productId"))
	if err != nil || id < 1 {
		sendErr(c, http.StatusNotFound, "NOT_FOUND", "Product not found", "Product ID must be a positive integer")
		return
	}
	mu.RLock()
	p, ok := products[id]
	mu.RUnlock()
	if !ok {
		sendErr(c, http.StatusNotFound, "NOT_FOUND", "Product not found", "No product with the given ID")
		return
	}
	c.JSON(http.StatusOK, p)
}

func validProduct(p Product) bool {
	if p.ProductID == nil || *p.ProductID < 1 {
		return false
	}
	if len(p.SKU) < 1 || len(p.SKU) > 100 {
		return false
	}
	if len(p.Manufacturer) < 1 || len(p.Manufacturer) > 200 {
		return false
	}
	if p.CategoryID == nil || *p.CategoryID < 1 {
		return false
	}
	if p.Weight == nil || *p.Weight < 0 {
		return false
	}
	if p.SomeOtherID == nil || *p.SomeOtherID < 1 {
		return false
	}
	return true
}

func addProductDetails(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("productId"))
	if err != nil || id < 1 {
		sendErr(c, http.StatusBadRequest, "INVALID_INPUT", "Invalid input data", "Product ID must be a positive integer")
		return
	}
	var p Product
	if err := c.ShouldBindJSON(&p); err != nil {
		sendErr(c, http.StatusBadRequest, "INVALID_INPUT", "Invalid input data", err.Error())
		return
	}
	if !validProduct(p) {
		sendErr(c, http.StatusBadRequest, "INVALID_INPUT", "Invalid input data", "One or more fields are missing or invalid")
		return
	}
	mu.Lock()
	products[id] = p
	mu.Unlock()
	c.Status(http.StatusNoContent)
}

func main() {
	r := gin.New()
	r.Use(gin.Recovery(), metrics())
	r.GET("/products/:productId", getProduct)
	r.POST("/products/:productId/details", addProductDetails)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	r.Run(":8080")
}
