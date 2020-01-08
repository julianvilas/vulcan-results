package api

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	report "github.com/adevinta/vulcan-report"
	"github.com/adevinta/vulcan-results/app"
	"github.com/adevinta/vulcan-results/storage"
	"github.com/goadesign/goa"
)

type Check struct {
	Scan *Scan `json:"scan"`
}

type Scan struct {
	ID        *string   `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

type CheckPatchReport struct {
	Report *string `json:"report"`
}

type CheckPatchLogs struct {
	Raw *string `json:"raw"`
}

// ResultsController implements the Results resource.
type ResultsController struct {
	*goa.Controller
	storage storage.Storage
}

// NewResultsController creates a Results controller.
func NewResultsController(service *goa.Service, s storage.Storage) *ResultsController {
	return &ResultsController{Controller: service.NewController("ResultsController"), storage: s}
}

// Report runs the report action.
func (c *ResultsController) Report(ctx *app.ReportResultsContext) error {
	goa.LogInfo(ctx, "Uploading report to S3", "scan_id", ctx.Payload.ScanID, "check_id", ctx.Payload.CheckID, "scan_started_at", ctx.Payload.ScanStartTime)
	link, err := c.saveReportToS3(ctx)
	if err == nil {
		goa.LogInfo(ctx, "Report uploaded to S3", "link", link)
		ctx.ResponseData.Header().Add("Location", link)
		return ctx.Created()
	}
	goa.LogError(ctx, err.Error())
	return ctx.BadRequest()
}

// Raw runs the raw action.
func (c *ResultsController) Raw(ctx *app.RawResultsContext) error {
	goa.LogInfo(ctx, "Uploading raw logs to S3", "scan_id", ctx.Payload.ScanID, "check_id", ctx.Payload.CheckID, "scan_started_at", ctx.Payload.ScanStartTime)
	link, err := c.saveLogsToS3(ctx)

	if err == nil {
		goa.LogInfo(ctx, "Raw logs uploaded to S3", "link", link)
		ctx.ResponseData.Header().Add("Location", link)
		return ctx.Created()
	}

	goa.LogError(ctx, err.Error())
	return ctx.BadRequest()
}

// GetReport runs the getReport action.
func (c *ResultsController) GetReport(ctx *app.GetReportResultsContext) error {
	goa.LogInfo(ctx, "Downloading report from S3",
		"date", ctx.Date, "scan", ctx.Scan, "check", ctx.Check)

	report, err := c.storage.GetReport(ctx.Date, ctx.Scan, ctx.Check)
	if err == nil {
		goa.LogInfo(ctx, "Report downloaded from S3")
		return ctx.OK(report)
	}

	goa.LogError(ctx, err.Error())
	return ctx.BadRequest()
}

// GetLog runs the getLog action.
func (c *ResultsController) GetLog(ctx *app.GetLogResultsContext) error {
	goa.LogInfo(ctx, "Downloading log from S3",
		"date", ctx.Date, "scan", ctx.Scan, "check", ctx.Check)

	log, err := c.storage.GetLog(ctx.Date, ctx.Scan, ctx.Check)
	if err == nil {
		goa.LogInfo(ctx, "Log downloaded from S3")
		return ctx.OK(log)
	}

	goa.LogError(ctx, err.Error())
	return ctx.BadRequest()
}

// saveReportToS3 must perform the following actions:
// - upload vulnerable reports to vulcan-core-vulnerable-reports-{env} bucket
// - upload vulnerable reports to vulcan-core-reports-{env} bucket
// - returns a link to the report
func (c *ResultsController) saveReportToS3(ctx context.Context) (link string, err error) {
	payload := ctx.(*app.ReportResultsContext).Payload

	missingRequired := false
	if payload.CheckID == nil {
		missingRequired = true
		goa.LogError(ctx, "check_id is missing")
	}

	if payload.ScanID == nil {
		missingRequired = true
		goa.LogError(ctx, "scan_id is missing")
	}

	if payload.ScanStartTime == nil {
		missingRequired = true
		goa.LogError(ctx, "scan_start_time is missing")
	}

	if payload.Report == nil {
		missingRequired = true
		goa.LogError(ctx, "report is missing")
	}

	if missingRequired {
		return "", fmt.Errorf("Missing required parameters")
	}

	checkID := payload.CheckID.String()
	scanID := payload.ScanID.String()
	scanStartTime := *payload.ScanStartTime
	notParsedReport := *payload.Report

	// if there are vulnerabilities mark the report to be uploaded to
	// the vulnerable reports bucket
	var parsedReport report.Report
	if err := json.Unmarshal([]byte(notParsedReport), &parsedReport); err != nil {
		return "", fmt.Errorf("the report can not be unmarshaled correctly: %v", err)
	}
	vulnerable := len(parsedReport.Vulnerabilities) > 0

	// Prepare the report to be stored for Athena.
	marshaledReport, err := parsedReport.MarshalJSONTimeAsString()
	if err != nil {
		return "", fmt.Errorf("the report can not be marshaled again: %v", err)
	}

	// save the report on report bucket
	return c.storage.SaveReports(scanID, checkID, scanStartTime, marshaledReport, vulnerable)
}

// saveLogsToS3 must perform the following actions:
// - upload the logs to vulcan-core-logs-{env} bucket
// - returns a link to the raw logs
func (c *ResultsController) saveLogsToS3(ctx context.Context) (link string, err error) {
	payload := ctx.(*app.RawResultsContext).Payload

	missingRequired := false
	if payload.CheckID == nil {
		missingRequired = true
		goa.LogError(ctx, "check_id is missing")
	}

	if payload.ScanID == nil {
		missingRequired = true
		goa.LogError(ctx, "scan_id is missing")
	}

	if payload.ScanStartTime == nil {
		missingRequired = true
		goa.LogError(ctx, "scan_start_time is missing")
	}

	if payload.Raw == nil {
		goa.LogInfo(ctx, "raw is missing")
	}

	if missingRequired {
		return "", fmt.Errorf("Missing required parameters. Received payload is %+v", ctx.(*app.RawResultsContext).Body)
	}

	checkID := payload.CheckID.String()
	scanID := payload.ScanID.String()
	scanStartTime := *payload.ScanStartTime
	raw := *payload.Raw

	dataRaw, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return "", err
	}

	//save the report on report bucket
	return c.storage.SaveLogs(scanID, checkID, scanStartTime, dataRaw)
}
