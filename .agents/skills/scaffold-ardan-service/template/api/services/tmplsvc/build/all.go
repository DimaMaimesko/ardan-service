// Package build binds the set of routes the service exposes. Add another
// file with a build tag and its own RouteAdder when a deployment needs a
// different subset of routes from the same codebase.
package build

import (
	"github.com/tmplorg/tmplsvc/app/domain/authapp"
	"github.com/tmplorg/tmplsvc/app/domain/checkapp"
	"github.com/tmplorg/tmplsvc/app/domain/productapp"
	"github.com/tmplorg/tmplsvc/app/domain/userapp"
	"github.com/tmplorg/tmplsvc/app/sdk/mux"
	"github.com/tmplorg/tmplsvc/foundation/web"
)

// Routes binds all the routes for the service.
func Routes() all {
	return all{}
}

type all struct{}

// Add implements the RouteAdder interface.
func (all) Add(app *web.App, cfg mux.Config) {
	checkapp.Routes(app, checkapp.Config{
		Build: cfg.Build,
		Log:   cfg.Log,
		DB:    cfg.DB,
	})

	authapp.Routes(app, authapp.Config{
		UserBus: cfg.BusConfig.UserBus,
		Auth:    cfg.Auth,
	})

	userapp.Routes(app, userapp.Config{
		Log:     cfg.Log,
		UserBus: cfg.BusConfig.UserBus,
		Auth:    cfg.Auth,
	})

	productapp.Routes(app, productapp.Config{
		Log:        cfg.Log,
		ProductBus: cfg.BusConfig.ProductBus,
		Auth:       cfg.Auth,
	})
}
