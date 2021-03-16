/*
Copyright 2019 Adevinta
*/

package design

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/apidsl"
)

var _ = Resource("Results", func() {
	BasePath("v1")

	Action("report", func() {
		Routing(POST("/report"))
		Description("Update the Report of a Check")
		Payload(ReportPayload)
		Response(Created)
		Response(BadRequest)
	})

	// TODO: we should modify this endpoint from 'raw' to 'logs'
	// for now, lets just keep it for compability reasons
	Action("raw", func() {
		Routing(POST("/raw"))
		Description("Update the Raw of a Check")
		Payload(RawPayload)
		Response(Created)
		Response(BadRequest)
	})

	Action("getReport", func() {
		Routing(GET("/reports/:date/:scan/:check"))
		Description("Download a report")
		Params(func() {
			Param("date", String, "Report date")
			Param("scan", String, "Scan ID")
			Param("check", String, "Check ID")
		})
		Response(OK)
		Response(BadRequest)
	})

	Action("getLog", func() {
		Routing(GET("/logs/:date/:scan/:check"))
		Description("Download a log")
		Params(func() {
			Param("date", String, "Report date")
			Param("scan", String, "Scan ID")
			Param("check", String, "Check ID")
		})
		Response(OK)
		Response(BadRequest)
	})
})

var ReportPayload = Type("ReportPayload", func() {
	Attribute("check_id", UUID, "Check UUID")
	Attribute("scan_id", UUID, "Scan UUID")
	Attribute("scan_start_time", DateTime, "Scan start time")
	Attribute("report", String, func() {
		MinLength(2)
		Example(`{ report : "{"report":"{\"check_id\":\"aabbccdd-abcd-0123-4567-abcdef012345\", .....}}" }`)
		Pattern("^[[:print:]]+")
		Description("Report of a Check. It's a JSON containing the value of the report")
	})
})

var RawPayload = Type("RawPayload", func() {
	Attribute("check_id", UUID, "Check UUID")
	Attribute("scan_id", UUID, "Scan UUID")
	Attribute("scan_start_time", DateTime, "Scan start time")
	Attribute("raw", String, func() {
		Example(`{ raw : "BASE_64_FORMAT" }`)
		Description("Raw result of a Check. It's a JSON with a BASE64 encoded value of the raw result")
	})
})
