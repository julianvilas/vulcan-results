package metrics

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"testing"
	"time"

	metrics "github.com/adevinta/vulcan-metrics-client"
	"github.com/goadesign/goa"
)

type mockMetricsClient struct {
	metrics.Client
	metrics         []metrics.Metric
	expectedMetrics []metrics.Metric
}

func (c *mockMetricsClient) Push(metric metrics.Metric) {
	c.metrics = append(c.metrics, metric)
}

// Verify verifies the matching between mock client
// expected metrics and the actual pushed metrics.
func (c *mockMetricsClient) Verify() error {
	nMetrics := len(c.metrics)
	nExpectedMetrics := len(c.expectedMetrics)

	if nMetrics != nExpectedMetrics {
		return fmt.Errorf(
			"Number of metrics do not match: Expected %d, but got %d",
			nExpectedMetrics, nMetrics)
	}

	for _, m := range c.metrics {
		var found bool
		for _, em := range c.expectedMetrics {
			if cmpMetrics(m, em) {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("Metrics do not match: Expected %v, but got %v",
				c.expectedMetrics, c.metrics)
		}
	}

	return nil
}

func cmpMetrics(got, exp metrics.Metric) bool {
	if got.Name != exp.Name || got.Typ != exp.Typ ||
		!reflect.DeepEqual(got.Tags, exp.Tags) {
		return false
	}
	if got.Name == metricDuration {
		// If metric is duration metric
		// check got value >= exp value
		return got.Value >= exp.Value
	}
	return got.Value == exp.Value

}

func TestMiddleware(t *testing.T) {
	type input struct {
		ctx context.Context
		rw  http.ResponseWriter
		req *http.Request
		h   goa.Handler
	}

	testCases := []struct {
		name            string
		input           input
		expectedMetrics []metrics.Metric
	}{
		{
			name: "Should push metrics for GetReport action",
			input: input{
				// We have to create context through goa so we can
				// obtain the HTTP response later on in middleware
				// through goa context parsing.
				ctx: goa.NewContext(nil, nil, nil, nil),
				req: &http.Request{
					Method: http.MethodGet,
					URL: &url.URL{
						Path: "/v1/reports/dt=2020-06-01/scan=06b38973-f395-4311-a4af-8f36e1b5b847/b4bda78e-12cc-4975-8362-95a5a10a3bfc.json",
					},
				},
				h: func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
					// Modify goa ctx response to set
					// http status.
					resp := goa.ContextResponse(ctx)
					resp.Status = http.StatusOK

					time.Sleep(50 * time.Millisecond)

					return nil
				},
			},
			expectedMetrics: []metrics.Metric{
				{
					Name:  metricTotal,
					Typ:   metrics.Count,
					Value: 1,
					Tags: []string{
						fmt.Sprint(tagComponent, ":", resultsComponent),
						fmt.Sprint(tagAction, ":", getReportAction),
						fmt.Sprint(tagEntity, ":", reportEntity),
						fmt.Sprint(tagMethod, ":", http.MethodGet),
						fmt.Sprint(tagStatus, ":", http.StatusOK),
					},
				},
				{
					Name:  metricDuration,
					Typ:   metrics.Histogram,
					Value: 50,
					Tags: []string{
						fmt.Sprint(tagComponent, ":", resultsComponent),
						fmt.Sprint(tagAction, ":", getReportAction),
						fmt.Sprint(tagEntity, ":", reportEntity),
						fmt.Sprint(tagMethod, ":", http.MethodGet),
						fmt.Sprint(tagStatus, ":", http.StatusOK),
					},
				},
			},
		},
		{
			name: "Should push metrics for PostReport action",
			input: input{
				ctx: goa.NewContext(nil, nil, nil, nil),
				req: &http.Request{
					Method: http.MethodPost,
					URL: &url.URL{
						Path: "/v1/reportdt=2020-06-01/scan=06b38973-f395-4311-a4af-8f36e1b5b847/b4bda78e-12cc-4975-8362-95a5a10a3bfc",
					},
				},
				h: func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
					resp := goa.ContextResponse(ctx)
					resp.Status = http.StatusConflict

					time.Sleep(20 * time.Millisecond)

					return nil
				},
			},
			expectedMetrics: []metrics.Metric{
				{
					Name:  metricTotal,
					Typ:   metrics.Count,
					Value: 1,
					Tags: []string{
						fmt.Sprint(tagComponent, ":", resultsComponent),
						fmt.Sprint(tagAction, ":", postReportAction),
						fmt.Sprint(tagEntity, ":", reportEntity),
						fmt.Sprint(tagMethod, ":", http.MethodPost),
						fmt.Sprint(tagStatus, ":", http.StatusConflict),
					},
				},
				{
					Name:  metricDuration,
					Typ:   metrics.Histogram,
					Value: 20,
					Tags: []string{
						fmt.Sprint(tagComponent, ":", resultsComponent),
						fmt.Sprint(tagAction, ":", postReportAction),
						fmt.Sprint(tagEntity, ":", reportEntity),
						fmt.Sprint(tagMethod, ":", http.MethodPost),
						fmt.Sprint(tagStatus, ":", http.StatusConflict),
					},
				},
				{
					Name:  metricFailed,
					Typ:   metrics.Count,
					Value: 1,
					Tags: []string{
						fmt.Sprint(tagComponent, ":", resultsComponent),
						fmt.Sprint(tagAction, ":", postReportAction),
						fmt.Sprint(tagEntity, ":", reportEntity),
						fmt.Sprint(tagMethod, ":", http.MethodPost),
						fmt.Sprint(tagStatus, ":", http.StatusConflict),
					},
				},
			},
		},
		{
			name: "Should push metrics for GetLog action",
			input: input{
				ctx: goa.NewContext(nil, nil, nil, nil),
				req: &http.Request{
					Method: http.MethodGet,
					URL: &url.URL{
						Path: "/v1/logs",
					},
				},
				h: func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
					resp := goa.ContextResponse(ctx)
					resp.Status = http.StatusOK

					time.Sleep(35 * time.Millisecond)

					return nil
				},
			},
			expectedMetrics: []metrics.Metric{
				{
					Name:  metricTotal,
					Typ:   metrics.Count,
					Value: 1,
					Tags: []string{
						fmt.Sprint(tagComponent, ":", resultsComponent),
						fmt.Sprint(tagAction, ":", getLogAction),
						fmt.Sprint(tagEntity, ":", logEntity),
						fmt.Sprint(tagMethod, ":", http.MethodGet),
						fmt.Sprint(tagStatus, ":", http.StatusOK),
					},
				},
				{
					Name:  metricDuration,
					Typ:   metrics.Histogram,
					Value: 35,
					Tags: []string{
						fmt.Sprint(tagComponent, ":", resultsComponent),
						fmt.Sprint(tagAction, ":", getLogAction),
						fmt.Sprint(tagEntity, ":", logEntity),
						fmt.Sprint(tagMethod, ":", http.MethodGet),
						fmt.Sprint(tagStatus, ":", http.StatusOK),
					},
				},
			},
		},
		{
			name: "Should push metrics for PostLog action",
			input: input{
				ctx: goa.NewContext(nil, nil, nil, nil),
				req: &http.Request{
					Method: http.MethodPost,
					URL: &url.URL{
						Path: "/v1/raw",
					},
				},
				h: func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
					resp := goa.ContextResponse(ctx)
					resp.Status = http.StatusInternalServerError

					time.Sleep(35 * time.Millisecond)

					return nil
				},
			},
			expectedMetrics: []metrics.Metric{
				{
					Name:  metricTotal,
					Typ:   metrics.Count,
					Value: 1,
					Tags: []string{
						fmt.Sprint(tagComponent, ":", resultsComponent),
						fmt.Sprint(tagAction, ":", postLogAction),
						fmt.Sprint(tagEntity, ":", logEntity),
						fmt.Sprint(tagMethod, ":", http.MethodPost),
						fmt.Sprint(tagStatus, ":", http.StatusInternalServerError),
					},
				},
				{
					Name:  metricDuration,
					Typ:   metrics.Histogram,
					Value: 35,
					Tags: []string{
						fmt.Sprint(tagComponent, ":", resultsComponent),
						fmt.Sprint(tagAction, ":", postLogAction),
						fmt.Sprint(tagEntity, ":", logEntity),
						fmt.Sprint(tagMethod, ":", http.MethodPost),
						fmt.Sprint(tagStatus, ":", http.StatusInternalServerError),
					},
				},
				{
					Name:  metricFailed,
					Typ:   metrics.Count,
					Value: 1,
					Tags: []string{
						fmt.Sprint(tagComponent, ":", resultsComponent),
						fmt.Sprint(tagAction, ":", postLogAction),
						fmt.Sprint(tagEntity, ":", logEntity),
						fmt.Sprint(tagMethod, ":", http.MethodPost),
						fmt.Sprint(tagStatus, ":", http.StatusInternalServerError),
					},
				},
			},
		},
		{
			name: "Should NOT push metrics for healthcheck endpoint",
			input: input{
				ctx: goa.NewContext(nil, nil, nil, nil),
				req: &http.Request{
					Method: http.MethodGet,
					URL: &url.URL{
						Path: "/healthcheck",
					},
				},
				h: func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
					resp := goa.ContextResponse(ctx)
					resp.Status = http.StatusOK
					return nil
				},
			},
			expectedMetrics: []metrics.Metric{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			metricsClient := &mockMetricsClient{
				expectedMetrics: tc.expectedMetrics,
			}

			middleware := NewMiddleware(metricsClient)

			middleware(tc.input.h)(tc.input.ctx, tc.input.rw, tc.input.req)

			if err := metricsClient.Verify(); err != nil {
				t.Fatalf("Error verifying pushed metrics: %v", err)
			}
		})
	}
}
