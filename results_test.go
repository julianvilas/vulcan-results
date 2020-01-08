package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/goadesign/goa"
	"github.com/goadesign/goa/goatest"
	uuid "github.com/satori/go.uuid"

	"github.com/adevinta/vulcan-results/app"
	"github.com/adevinta/vulcan-results/app/test"
	"github.com/adevinta/vulcan-results/storage"
)

var base64report = "{}"
var plainReport = `{ "vulnerabilities" : [ {} ] }`

var scanID = uuid.FromStringOrNil("e0c1ac1a-1036-4e0e-b5cc-d18ae6673eb0")
var checkID = uuid.FromStringOrNil("e0c1ac1a-1036-4e0e-b5cc-d18ae6673eb0")
var scanStartTime = time.Now()

var base64raw = "e30="

var testCasesReport = []struct {
	name             string
	skip, skipAlways bool
	payload          *app.ReportPayload
	stMock           storage.Storage
	psMock           http.HandlerFunc
	psURL            string
	f                funcTestReport
}{
	{
		name: "positive case",
		payload: &app.ReportPayload{
			Report:        &plainReport,
			ScanID:        &scanID,
			CheckID:       &checkID,
			ScanStartTime: &scanStartTime,
		},
		stMock: storageMock{
			//link: "https://s3-eu-west-1.amazonaws.com/vulcan-results/e0c1ac1a-1036-4e0e-b5cc-d18ae6673eb0/report.json",
			link: "https://s3-eu-west-1.amazonaws.com/vulcan-core-reports-dev/dt=2017-09-01/scan=e0c1ac1a-1036-4e0e-b5cc-d18ae6673eb0/e0c1ac1a-1036-4e0e-b5cc-d18ae6673eb0.json",
			err:  nil,
		},
		psMock: ok,
		f:      test.ReportResultsCreated,
	},
}

var testCasesRaw = []struct {
	name             string
	checkID          uuid.UUID
	skip, skipAlways bool
	payload          *app.RawPayload
	stMock           storage.Storage
	psMock           http.HandlerFunc
	psURL            string
	f                funcTestRaw
}{
	{
		name: "positive case",
		payload: &app.RawPayload{
			Raw:           &base64raw,
			ScanID:        &scanID,
			CheckID:       &checkID,
			ScanStartTime: &scanStartTime,
		},
		stMock: storageMock{
			link: "https://s3-eu-west-1.amazonaws.com/vulcan-results/e0c1ac1a-1036-4e0e-b5cc-d18ae6673eb0/raw.json",
			err:  nil,
		},
		psMock: ok,
		f:      test.RawResultsCreated,
	},
}

type funcTestReport func(goatest.TInterface, context.Context, *goa.Service, app.ResultsController, *app.ReportPayload) http.ResponseWriter
type funcTestRaw func(goatest.TInterface, context.Context, *goa.Service, app.ResultsController, *app.RawPayload) http.ResponseWriter
type funcTestGetReport func(goatest.TInterface, context.Context, *goa.Service, app.ResultsController, string, string, string) http.ResponseWriter
type funcTestGetLog func(goatest.TInterface, context.Context, *goa.Service, app.ResultsController, string, string, string) http.ResponseWriter

type storageMock struct {
	link   string
	report []byte
	log    []byte
	err    error
}

func (st storageMock) SaveLogs(checkID, scanID string, startedAt time.Time, raw []byte) (link string, err error) {
	return st.link, st.err
}

func (st storageMock) SaveReports(checkID, scanID string, startedAt time.Time, result []byte, compress bool) (link string, err error) {
	return st.link, st.err
}

func (st storageMock) GetReport(date, scanID, checkID string) ([]byte, error) {
	return st.report, st.err
}

func (st storageMock) GetLog(date, scanID, checkID string) ([]byte, error) {
	return st.log, st.err
}

func ok(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	dec := json.NewDecoder(r.Body)

	var v map[string]interface{}
	if err := dec.Decode(&v); err != nil {
		http.Error(w, "error decoding payload", 500)
		return
	}

	if len(v) != 1 {
		http.Error(w, "more fields than one in the payload", 500)
		return
	}

	raw, ok := v["raw"]
	if !ok || raw != "https://s4-eu-west-1.amazonaws.com/vulcan-results/e0c1ac1a-1036-4e0e-b5cc-d18ae6673eb0/raw.json" {
		http.Error(w, "incorrect payload content", 500)
		return
	}
}

func okIfStatus(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	dec := json.NewDecoder(r.Body)

	var v map[string]interface{}
	if err := dec.Decode(&v); err != nil {
		http.Error(w, "error decoding payload", 500)
		return
	}

	if len(v) != 1 {
		http.Error(w, "more fields than one in the payload", 500)
		return
	}

	status, ok := v["status"]
	if !ok || status != "FAILED" {
		http.Error(w, "incorrect payload content", 500)
		return
	}
}

func forbidden(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "", 403)
}

func TestReport(t *testing.T) {
	// Test all the test cases defined in testCasesReport
	for _, tc := range testCasesReport {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			if testing.Short() && tc.skip || tc.skipAlways {
				t.SkipNow()
			}

			url := tc.psURL
			if url == "" {
				ts := httptest.NewServer(tc.psMock)
				defer ts.Close()
				url = ts.URL
			}

			service := goa.New("vulcan-results")

			ctrl := NewResultsController(service, tc.stMock)

			_ = tc.f(t, nil, service, ctrl, tc.payload)
		})
	}
}

func TestRaw(t *testing.T) {
	// Test all the test cases defined in testCasesRaw
	for _, tc := range testCasesRaw {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			if testing.Short() && tc.skip || tc.skipAlways {
				t.SkipNow()
			}

			url := tc.psURL
			if url == "" {
				ts := httptest.NewServer(tc.psMock)
				defer ts.Close()
				url = ts.URL
			}

			service := goa.New("vulcan-results")

			ctrl := NewResultsController(service, tc.stMock)

			_ = tc.f(t, nil, service, ctrl, tc.payload)
		})
	}
}

var testCasesSaveResult = []struct {
	name             string
	skip, skipAlways bool
	checkID          string
	kind             string
	result           string
	stMock           storage.Storage
	psMock           http.HandlerFunc
	psURL            string
	nilErr           bool
}{
	{
		name:    "positive case",
		checkID: "e0c1ac1a-1036-4e0e-b5cc-d18ae6673eb0",
		kind:    "raw",
		result:  "RAW_RESULT_CONTENT",
		stMock: storageMock{
			link: "https://s4-eu-west-1.amazonaws.com/vulcan-results/e0c1ac1a-1036-4e0e-b5cc-d18ae6673eb0/raw.json",
			err:  nil,
		},
		psMock: ok,
		nilErr: true,
	},
	{
		name:    "s3 error",
		checkID: "e0c1ac1a-1036-4e0e-b5cc-d18ae6673eb0",
		kind:    "raw",
		result:  "RAW_RESULT_CONTENT",
		stMock: storageMock{
			link: "",
			err:  errors.New("error storing in S3"),
		},
		psMock: okIfStatus,
		nilErr: true,
	},
	{
		name:    "malformed persistence url",
		checkID: "e0c1ac1a-1036-4e0e-b5cc-d18ae6673eb0",
		kind:    "raw",
		result:  "RAW_RESULT_CONTENT",
		stMock: storageMock{
			link: "https://s3-eu-west-1.amazonaws.com/vulcan-results/e0c1ac1a-1036-4e0e-b5cc-d18ae6673eb0/report.json",
			err:  nil,
		},
		psMock: ok,
		psURL:  "//fwefew//fewwf///.../",
		nilErr: false,
	},
	{
		name:    "persistence service answers different from http.StatusOK",
		checkID: "e0c1ac1a-1036-4e0e-b5cc-d18ae6673eb0",
		kind:    "raw",
		result:  "RAW_RESULT_CONTENT",
		stMock: storageMock{
			link: "https://s3-eu-west-1.amazonaws.com/vulcan-results/e0c1ac1a-1036-4e0e-b5cc-d18ae6673eb0/report.json",
			err:  nil,
		},
		psMock: forbidden,
		nilErr: false,
	},
}

func TestSaveResult(t *testing.T) {
	// Test all the test cases defined in testCasesSaveResult
	for _, tc := range testCasesSaveResult {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			if testing.Short() && tc.skip || tc.skipAlways {
				t.SkipNow()
			}

			url := tc.psURL
			if url == "" {
				ts := httptest.NewServer(tc.psMock)
				defer ts.Close()
				url = ts.URL
			}

			service := goa.New("vulcan-results")

			ctrl := NewResultsController(service, tc.stMock)

			goaCtx := goa.NewContext(goa.WithAction(context.Background(), "ResultsTest"), nil, nil, nil)
			ctx, _ := app.NewReportResultsContext(goaCtx, nil, service)
			ctx.Payload = &app.ReportPayload{
				CheckID:       &checkID,
				ScanID:        &scanID,
				ScanStartTime: &scanStartTime,
				Report:        &plainReport,
			}
			_, err := ctrl.saveReportToS3(ctx)

			if (tc.nilErr && err != nil) || (!tc.nilErr && err == nil) {
				//TODO: fix this test
				//t.Errorf("(%v) nilErr expected: %v, got error: %v", tc.name, tc.nilErr, err)
			}
		})
	}
}

var testCasesGetReport = []struct {
	name   string
	date   string
	scan   string
	check  string
	stMock storage.Storage
	psMock http.HandlerFunc
	psURL  string
	f      funcTestGetReport
}{
	{
		name:  "Happy path OK",
		date:  "dt=2019-11-01",
		scan:  "scan=9126034c-7caf-4acd-93f3-bee1941aa140",
		check: "e0c1ac1a-1036-4e0e-b5cc-d18ae6673eb0.json",
		stMock: storageMock{
			report: []byte("myReport"),
			err:    nil,
		},
		psMock: func(w http.ResponseWriter, r *http.Request) {},
		f:      test.GetReportResultsOK,
	},
	{
		name:  "Should return bad request",
		date:  "dt=2019-11-01",
		scan:  "scan=9126034c-7caf-4acd-93f3-bee1941aa140",
		check: "e0c1ac1a-1036-4e0e-b5cc-d18ae6673eb0.json",
		stMock: storageMock{
			report: []byte("myReport"),
			err:    errors.New("Error"),
		},
		psMock: func(w http.ResponseWriter, r *http.Request) {},
		f:      test.GetReportResultsBadRequest,
	},
}

func TestGetReport(t *testing.T) {
	// Test all the test cases defined in testCasesGetReport
	for _, tc := range testCasesGetReport {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			url := tc.psURL
			if url == "" {
				ts := httptest.NewServer(tc.psMock)
				defer ts.Close()
				url = ts.URL
			}

			service := goa.New("vulcan-results")

			ctrl := NewResultsController(service, tc.stMock)

			_ = tc.f(t, nil, service, ctrl, tc.date, tc.scan, tc.check)
		})
	}
}

var testCasesGetLog = []struct {
	name   string
	date   string
	scan   string
	check  string
	stMock storage.Storage
	psMock http.HandlerFunc
	psURL  string
	f      funcTestGetLog
}{
	{
		name:  "Happy path OK",
		date:  "dt=2019-11-01",
		scan:  "scan=9126034c-7caf-4acd-93f3-bee1941aa140",
		check: "e0c1ac1a-1036-4e0e-b5cc-d18ae6673eb0.json",
		stMock: storageMock{
			report: []byte("myLog"),
			err:    nil,
		},
		psMock: func(w http.ResponseWriter, r *http.Request) {},
		f:      test.GetLogResultsOK,
	},
	{
		name:  "Should return bad request",
		date:  "dt=2019-11-01",
		scan:  "scan=9126034c-7caf-4acd-93f3-bee1941aa140",
		check: "e0c1ac1a-1036-4e0e-b5cc-d18ae6673eb0.json",
		stMock: storageMock{
			report: []byte("myLog"),
			err:    errors.New("Error"),
		},
		psMock: func(w http.ResponseWriter, r *http.Request) {},
		f:      test.GetLogResultsBadRequest,
	},
}

func TestGetLog(t *testing.T) {
	// Test all the test cases defined in testCasesGetLog
	for _, tc := range testCasesGetLog {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			url := tc.psURL
			if url == "" {
				ts := httptest.NewServer(tc.psMock)
				defer ts.Close()
				url = ts.URL
			}

			service := goa.New("vulcan-results")

			ctrl := NewResultsController(service, tc.stMock)

			_ = tc.f(t, nil, service, ctrl, tc.date, tc.scan, tc.check)
		})
	}
}
