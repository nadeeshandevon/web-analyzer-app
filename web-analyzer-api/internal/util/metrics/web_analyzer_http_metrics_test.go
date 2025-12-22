package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestRecordHttpRequestTotal(t *testing.T) {
	RegisterAPIMetrics()

	method := "POST"
	path := "/api/v1/analyze"
	status := "200"

	initialCount := testutil.ToFloat64(httpRequestsTotal.WithLabelValues(method, path, status))

	RecordHttpRequestTotal(method, path, status)

	count1 := testutil.ToFloat64(httpRequestsTotal.WithLabelValues(method, path, status))
	if count1 != initialCount+1 {
		t.Errorf("Expected count to be %f, got %f", initialCount+1, count1)
	}
	RecordHttpRequestTotal(method, path, status)
	count2 := testutil.ToFloat64(httpRequestsTotal.WithLabelValues(method, path, status))
	if count2 != initialCount+2 {
		t.Errorf("Expected count to be %f, got %f", initialCount+2, count2)
	}
}
