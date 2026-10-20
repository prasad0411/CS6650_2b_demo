package main

import (
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Product struct {
	ProductID    int    `json:"product_id"`
	SKU          string `json:"sku"`
	Manufacturer string `json:"manufacturer"`
	CategoryID   int    `json:"category_id"`
	Weight       int    `json:"weight"`
	SomeOtherID  int    `json:"some_other_id"`
	Name         string `json:"name"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

const (
	totalProducts = 100000
	checkLimit    = 100
	maxResults    = 20
)

var (
	store    sync.Map
	logger   = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	brands   = []string{"Alpha", "Beta", "Gamma", "Delta", "Omega"}
	cats     = []string{"Electronics", "Books", "Home", "Toys", "Sports", "Garden"}
	catName  = map[int]string{}

	reqTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "http_requests_total", Help: "Total HTTP requests"},
		[]string{"method", "endpoint", "status"})
	errTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "http_errors_total", Help: "Total non-2xx responses"},
		[]string{"method", "endpoint", "status"})
	reqDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{Name: "http_request_duration_seconds", Help: "Request duration", Buckets: prometheus.DefBuckets},
		[]string{"method", "endpoint"})
)

func init() {
	prometheus.MustRegister(reqTotal, errTotal, reqDuration)
	for i, c := range cats {
		catName[i+1] = c
	}
}

func seed() {
	for i := 1; i <= totalProducts; i++ {
		brand := brands[i%len(brands)]
		catID := (i % len(cats)) + 1
		store.Store(i, Product{
			ProductID:    i,
			SKU:          "SKU-" + strconv.Itoa(i),
			Manufacturer: brand,
			CategoryID:   catID,
			Weight:       100 + i%500,
			SomeOtherID:  i,
			Name:         "Product " + brand + " " + strconv.Itoa(i),
		})
	}
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
		logger.Info("request", "method", c.Request.Method, "endpoint", endpoint, "status", status, "duration_ms", dur*1000)
	}
}

func sendErr(c *gin.Context, code int, ec, msg, det string) {
	c.JSON(code, ErrorResponse{Error: ec, Message: msg, Details: det})
}

func searchProducts(c *gin.Context) {
	q := c.Query("q")
	if len(q) < 1 || len(q) > 100 {
		sendErr(c, http.StatusBadRequest, "INVALID_INPUT", "Invalid input data", "q must be 1 to 100 characters")
		return
	}
	ql := strings.ToLower(q)
	results := []Product{}
	checked := 0
	for i := 1; i <= totalProducts && checked < checkLimit; i++ {
		v, ok := store.Load(i)
		if !ok {
			continue
		}
		checked++
		p := v.(Product)
		cn := strings.ToLower(catName[p.CategoryID])
		if strings.Contains(strings.ToLower(p.Name), ql) || strings.Contains(cn, ql) {
			if len(results) < maxResults {
				results = append(results, p)
			}
		}
	}
	c.JSON(http.StatusOK, results)
}

func health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func main() {
	seed()
	r := gin.New()
	r.Use(gin.Recovery(), metrics())
	r.GET("/health", health)
	r.GET("/products/search", searchProducts)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	r.Run(":8080")
}
