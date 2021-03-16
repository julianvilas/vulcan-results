/*
Copyright 2019 Adevinta
*/

package design

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/apidsl"
)

var _ = API("vulcan-results", func() {
	Title("Vulcan Persistence Results Uploader")
	Description("A component to handle persistence service results storage")
	Scheme("http")
	Host("localhost:8080")
	Consumes("application/json")
	ResponseTemplate(Created, func(pattern string) {
		Description("Resource created")
		Status(201)
		Headers(func() {
			Header("Location", String, "href to created resource", func() {
				Pattern(pattern)
			})
		})
	})
})
