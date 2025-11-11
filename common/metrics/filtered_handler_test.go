package metrics_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.temporal.io/server/common/log"
	"go.temporal.io/server/common/metrics"
	"go.temporal.io/server/common/metrics/metricstest"
)

func TestFilteredMetricsHandler_StartBatchFiltersMetrics(t *testing.T) {
	captureHandler := metricstest.NewCaptureHandler()
	capture := captureHandler.StartCapture()
	t.Cleanup(func() { captureHandler.StopCapture(capture) })

	allowed := map[string]bool{"allowed": true}
	handler := metrics.NewFilteredMetricsHandler(captureHandler, allowed, log.NewNoopLogger())

	batch := handler.StartBatch("test")
	require.NotNil(t, batch)

	batch.Counter("allowed").Record(1)
	batch.Counter("blocked").Record(1)
	require.NoError(t, batch.Close())

	records := capture.Snapshot()
	require.Contains(t, records, "allowed")
	require.NotContains(t, records, "blocked")
}
