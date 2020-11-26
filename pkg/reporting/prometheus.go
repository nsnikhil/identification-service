package reporters

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	attempt = "attempt"
	success = "success"
	failure = "failure"

	apiCounterName = "identification_api_call"
	apiCounterHelp = "total number of api calls"

	apiResponseTimeName = "identification_api_response_time"
	apiResponseTimeHelp = "total time taken by the api"
)

type Prometheus interface {
	ReportAttempt(bucket string)
	ReportSuccess(bucket string)
	ReportFailure(bucket string)
	Observe(bucket string, value float64)
}

//TODO: REMOVE (REMOVE DEFAULT)
type defaultPrometheus struct {
	apiCounter        *prometheus.CounterVec
	responseHistogram *prometheus.HistogramVec
}

func (dp *defaultPrometheus) ReportAttempt(bucket string) {
	incCounter(dp.apiCounter, attempt, bucket)
}

func (dp *defaultPrometheus) ReportSuccess(bucket string) {
	incCounter(dp.apiCounter, success, bucket)
}

func (dp *defaultPrometheus) ReportFailure(bucket string) {
	incCounter(dp.apiCounter, failure, bucket)
}

func incCounter(counter *prometheus.CounterVec, call, api string) {
	counter.WithLabelValues(call, api).Inc()
}

func newCounter() *prometheus.CounterVec {
	return prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: apiCounterName,
		Help: apiCounterHelp,
	}, []string{"call", "api"})
}

func (dp *defaultPrometheus) Observe(bucket string, value float64) {
	dp.responseHistogram.WithLabelValues(bucket).Observe(value)
}

func newHistogram() *prometheus.HistogramVec {
	return prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: apiResponseTimeName,
		Help: apiResponseTimeHelp,
	}, []string{"api"})
}

func NewPrometheus() Prometheus {
	ct := newCounter()
	ht := newHistogram()

	//prometheus.MustRegister(ct, ht)

	prometheus.Register(ct)
	prometheus.Register(ht)

	return &defaultPrometheus{
		apiCounter:        ct,
		responseHistogram: ht,
	}
}
