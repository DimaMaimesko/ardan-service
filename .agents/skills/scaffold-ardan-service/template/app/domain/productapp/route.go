package productapp

import (
	"net/http"

	"github.com/tmplorg/tmplsvc/app/sdk/auth"
	"github.com/tmplorg/tmplsvc/app/sdk/mid"
	"github.com/tmplorg/tmplsvc/business/domain/productbus"
	"github.com/tmplorg/tmplsvc/foundation/logger"
	"github.com/tmplorg/tmplsvc/foundation/web"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log        *logger.Logger
	ProductBus productbus.ExtBusiness
	Auth       *auth.Auth
}

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Bearer(cfg.Auth)
	ruleAny := mid.Authorize(cfg.Auth, auth.RuleAny)
	ruleUserOnly := mid.Authorize(cfg.Auth, auth.RuleUserOnly)
	ruleAuthorizeProduct := mid.AuthorizeProduct(cfg.Auth, cfg.ProductBus)

	api := newApp(cfg.ProductBus)

	app.HandlerFunc(http.MethodGet, version, "/products", api.query, authen, ruleAny)
	app.HandlerFunc(http.MethodGet, version, "/products/{product_id}", api.queryByID, authen, ruleAuthorizeProduct)
	app.HandlerFunc(http.MethodPost, version, "/products", api.create, authen, ruleUserOnly)
	app.HandlerFunc(http.MethodPut, version, "/products/{product_id}", api.update, authen, ruleAuthorizeProduct)
	app.HandlerFunc(http.MethodDelete, version, "/products/{product_id}", api.delete, authen, ruleAuthorizeProduct)
}
