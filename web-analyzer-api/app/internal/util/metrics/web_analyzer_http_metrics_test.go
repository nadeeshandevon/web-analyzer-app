package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestRecordHttpRequestTotal(t *testing.T) {
	RegisterAPIMetrics()

	method := "POST"
	path := "/api/v1/analyze"
	status := "200"

	initialCount := testutil.ToFloat64(httpRequestsTotal.WithLabelValues(method, path, status))

	RecordHttpRequestTotal(method, path, status)

	count1 := testutil.ToFloat64(httpRequestsTotal.WithLabelValues(method, path, status))
	assert.Equal(t, initialCount+1, count1)

	RecordHttpRequestTotal(method, path, status)
	count2 := testutil.ToFloat64(httpRequestsTotal.WithLabelValues(method, path, status))
	assert.Equal(t, initialCount+2, count2)
}
