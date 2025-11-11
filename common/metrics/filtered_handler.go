package metrics

import (
	"time"

	"go.temporal.io/server/common/log"
)

// FilteredMetricsHandler wraps another Handler and filters out unwanted metrics
type FilteredMetricsHandler struct {
	delegate       Handler
	allowedMetrics map[string]bool
	logger         log.Logger
}

// NewFilteredMetricsHandler creates a new filtered metrics handler
// allowedMetrics: map of metric names that should be kept (if nil or empty, allows all)
func NewFilteredMetricsHandler(delegate Handler, allowedMetrics map[string]bool, logger log.Logger) Handler {
	return &FilteredMetricsHandler{
		delegate:       delegate,
		allowedMetrics: allowedMetrics,
		logger:         logger,
	}
}

func (h *FilteredMetricsHandler) isAllowed(metricName string) bool {
	if h.allowedMetrics == nil || len(h.allowedMetrics) == 0 {
		return true // If no filter specified, allow all
	}
	return h.allowedMetrics[metricName]
}

func (h *FilteredMetricsHandler) WithTags(tags ...Tag) Handler {
	return &FilteredMetricsHandler{
		delegate:       h.delegate.WithTags(tags...),
		allowedMetrics: h.allowedMetrics,
		logger:         h.logger,
	}
}

func (h *FilteredMetricsHandler) Counter(name string) CounterIface {
	if !h.isAllowed(name) {
		return noopCounter{}
	}
	return h.delegate.Counter(name)
}

func (h *FilteredMetricsHandler) Gauge(name string) GaugeIface {
	if !h.isAllowed(name) {
		return noopGauge{}
	}
	return h.delegate.Gauge(name)
}

func (h *FilteredMetricsHandler) Timer(name string) TimerIface {
	if !h.isAllowed(name) {
		return noopTimer{}
	}
	return h.delegate.Timer(name)
}

func (h *FilteredMetricsHandler) Histogram(name string, unit MetricUnit) HistogramIface {
	if !h.isAllowed(name) {
		return noopHistogram{}
	}
	return h.delegate.Histogram(name, unit)
}

func (h *FilteredMetricsHandler) Stop(logger log.Logger) {
	h.delegate.Stop(logger)
}

func (h *FilteredMetricsHandler) StartBatch(name string) BatchHandler {
	delegateBatch := h.delegate.StartBatch(name)
	if delegateBatch == nil {
		return nil
	}

	return &filteredBatchHandler{
		FilteredMetricsHandler: &FilteredMetricsHandler{
			delegate:       delegateBatch,
			allowedMetrics: h.allowedMetrics,
			logger:         h.logger,
		},
		delegate: delegateBatch,
	}
}

// Noop implementations for filtered-out metrics
type noopCounter struct{}

func (noopCounter) Record(int64, ...Tag) {}

type noopGauge struct{}

func (noopGauge) Record(float64, ...Tag) {}

type noopTimer struct{}

func (noopTimer) Record(time.Duration, ...Tag) {}

type noopHistogram struct{}

func (noopHistogram) Record(int64, ...Tag) {}

type filteredBatchHandler struct {
	*FilteredMetricsHandler
	delegate BatchHandler
}

func (b *filteredBatchHandler) Close() error {
	if b.delegate == nil {
		return nil
	}
	return b.delegate.Close()
}
