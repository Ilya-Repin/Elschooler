package metrics

import (
	"Elschool-API/internal/config"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	StatusOk       = "ok"
	StatusErr      = "err"
	StatusHit      = "hit"
	StatusMiss     = "miss"
	ServiceUser    = "user"
	ServiceMarks   = "marks"
	ServiceStudent = "student"
	TypeDay        = "day"
	TypeAverage    = "average"
	TypeFinal      = "final"
	MethodAuth     = "auth"
	MethodCheck    = "check"
	ActionWrite    = "write"
	ActionDelete   = "delete"
	ActionRead     = "read"
	ActionUpdate   = "update"
	ActionAdd      = "add"
)

type Metrics struct {
	UserRegistrations     *prometheus.CounterVec
	StudentActions        *prometheus.CounterVec
	MarksRequests         *prometheus.CounterVec
	CacheModifyTotal      *prometheus.CounterVec
	TokenCacheRateTotal   *prometheus.CounterVec
	MarksCacheRateTotal   *prometheus.CounterVec
	StorageRequestsTotal  *prometheus.CounterVec
	ElschoolFetchTotal    *prometheus.CounterVec
	ElschoolFetchDuration *prometheus.HistogramVec
	ElschoolAuthTotal     *prometheus.CounterVec
	ElschoolAuthDuration  *prometheus.HistogramVec
}

func New(config *config.MetricsConfig) (*Metrics, error) {
	m := &Metrics{}

	m.UserRegistrations = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_registrations_total",
			Help: "Total number of user registration attempts",
		},
		[]string{"service", "status"},
	)
	m.StudentActions = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "student_requests_total",
			Help: "Total number of student-related actions",
		},
		[]string{"action", "status"},
	)
	m.MarksRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "marks_requests_total",
			Help: "Total number of marks-related API calls",
		},
		[]string{"type", "status"},
	)
	m.CacheModifyTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_write_total",
			Help: "Total number of cache write",
		},
		[]string{"service", "action", "status"},
	)
	m.TokenCacheRateTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "token_cache_rate_total",
			Help: "Hit rate of tokens cache",
		},
		[]string{"status"},
	)
	m.MarksCacheRateTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "marks_cache_rate_total",
			Help: "Hit rate of marks cache",
		},
		[]string{"type", "status"},
	)
	m.StorageRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "storage_requests_total",
			Help: "Total number of storage requests",
		},
		[]string{"service", "action", "status"},
	)
	m.ElschoolFetchTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "elschool_fetch_total",
			Help: "Total number of elschool requests",
		},
		[]string{"type", "status"},
	)
	m.ElschoolFetchDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "elschool_fetch_duration_seconds",
			Help:    "Duration of fetching elschool",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"type"},
	)
	m.ElschoolAuthTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "elschool_auth_total",
			Help: "Total number of elschool auth requests",
		},
		[]string{"method", "status"},
	)
	m.ElschoolAuthDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "elschool_auth_duration_seconds",
			Help:    "Duration of auth requests to elschool",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)

	prometheus.MustRegister(
		m.UserRegistrations,
		m.StudentActions,
		m.MarksRequests,
		m.CacheModifyTotal,
		m.TokenCacheRateTotal,
		m.MarksCacheRateTotal,
		m.StorageRequestsTotal,
		m.ElschoolFetchTotal,
		m.ElschoolFetchDuration,
		m.ElschoolAuthTotal,
		m.ElschoolAuthDuration,
	)

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(config.Address, nil); err != nil {
			panic("metrics server failed: " + err.Error())
		}
	}()

	return m, nil
}
