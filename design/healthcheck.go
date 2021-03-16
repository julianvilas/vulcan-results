/*
Copyright 2019 Adevinta
*/

package design

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/apidsl"
)

var _ = Resource("healthcheck", func() {
	BasePath("/healthcheck")

	Action("show", func() {
		Routing(GET(""))
		Description("Get the health status for the application")
		Response(OK)
	})
})
