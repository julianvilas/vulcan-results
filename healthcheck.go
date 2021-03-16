/*
Copyright 2019 Adevinta
*/

package api

import (
	"github.com/goadesign/goa"
	"github.com/adevinta/vulcan-results/app"
)

// HealthcheckController implements the healthcheck resource.
type HealthcheckController struct {
	*goa.Controller
}

// NewHealthcheckController creates a healthcheck controller.
func NewHealthcheckController(service *goa.Service) *HealthcheckController {
	return &HealthcheckController{Controller: service.NewController("HealthcheckController")}
}

// Show runs the show action.
func (c *HealthcheckController) Show(ctx *app.ShowHealthcheckContext) error {
	return ctx.OK([]byte{})
}
