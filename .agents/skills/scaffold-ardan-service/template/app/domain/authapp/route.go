package authapp

import (
	"net/http"

	"github.com/tmplorg/tmplsvc/app/sdk/auth"
	"github.com/tmplorg/tmplsvc/app/sdk/mid"
	"github.com/tmplorg/tmplsvc/business/domain/userbus"
	"github.com/tmplorg/tmplsvc/foundation/web"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	UserBus userbus.ExtBusiness
	Auth    *auth.Auth
}

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	basic := mid.Basic(cfg.Auth, cfg.UserBus)

	api := newApp(cfg.Auth)

	app.HandlerFunc(http.MethodGet, version, "/auth/token/{kid}", api.token, basic)
}
