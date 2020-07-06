package metrics

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	metrics "github.com/adevinta/vulcan-metrics-client"
	"github.com/goadesign/goa"
)

const (
	// Endpoint path prefixes
	getReportPathPrefix  = "/v1/reports"
	postReportPathPrefix = "/v1/report"
	getLogPathPrefix     = "/v1/logs"
	postLogPathPrefix    = "/v1/raw"

	// Endpoint actions
	postReportAction = "PostReport"
	getReportAction  = "GetReport"
	postLogAction    = "PostLog"
	getLogAction     = "GetLog"

	unknownAction = "unknown"

	// Metric names
	metricTotal    = "vulcan.request.total"
	metricDuration = "vulcan.request.duration"
	metricFailed   = "vulcan.request.failed"

	// Metric tags
	resultsComponent = "results"

	tagComponent = "component"
	tagAction    = "action"
	tagEntity    = "entity"
	tagMethod    = "method"
	tagStatus    = "status"

	reportEntity = "report"
	logEntity    = "log"
)

var (
	actionToEntity = map[string]string{
		postReportAction: reportEntity,
		getReportAction:  reportEntity,
		postLogAction:    logEntity,
		getLogAction:     logEntity,
	}
)

// NewMiddleware builds and returns a new metrics middleware for the API.
func NewMiddleware(metricsClient metrics.Client) goa.Middleware {
	return func(h goa.Handler) goa.Handler {
		return func(ctx context.Context, rw http.ResponseWriter, req *http.Request) (err error) {
			// Do not push metrics for healtchcheck
			if req.URL.Path == "/healthcheck" {
				return h(ctx, rw, req)
			}

			// Time and execute request
			reqStart := time.Now()
			err = h(ctx, rw, req)
			reqEnd := time.Now()

			// Collect metrics
			resp := goa.ContextResponse(ctx)

			httpMethod := req.Method
			path := req.URL.Path
			action := parseAction(httpMethod, path)
			httpStatus := resp.Status
			duration := reqEnd.Sub(reqStart).Milliseconds()
			failed := httpStatus >= 400

			// Build tags
			tags := []string{
				fmt.Sprint(tagComponent, ":", resultsComponent),
				fmt.Sprint(tagAction, ":", action),
				fmt.Sprint(tagEntity, ":", actionToEntity[action]),
				fmt.Sprint(tagMethod, ":", httpMethod),
				fmt.Sprint(tagStatus, ":", httpStatus),
			}

			// Push metrics
			mm := buildMetrics(httpMethod, duration, failed, tags)

			for _, met := range mm {
				metricsClient.Push(met)
			}

			return err
		}
	}
}

func buildMetrics(httpMethod string, duration int64, failed bool, tags []string) []metrics.Metric {
	mm := []metrics.Metric{
		{
			Name:  metricTotal,
			Typ:   metrics.Count,
			Value: 1,
			Tags:  tags,
		},
		{
			Name:  metricDuration,
			Typ:   metrics.Histogram,
			Value: float64(duration),
			Tags:  tags,
		},
	}

	if failed {
		mm = append(mm, metrics.Metric{
			Name:  metricFailed,
			Typ:   metrics.Count,
			Value: 1,
			Tags:  tags,
		})
	}

	return mm
}

func parseAction(httpMethod, path string) string {
	if httpMethod == http.MethodGet {
		if strings.HasPrefix(path, getReportPathPrefix) {
			return getReportAction
		}
		if strings.HasPrefix(path, getLogPathPrefix) {
			return getLogAction
		}
	} else if httpMethod == http.MethodPost {
		if strings.HasPrefix(path, postReportPathPrefix) {
			return postReportAction
		}
		if strings.HasPrefix(path, postLogPathPrefix) {
			return postLogAction
		}
	}

	return unknownAction
}
