package userapp

import (
	"net/http"

	"github.com/tmplorg/tmplsvc/app/sdk/auth"
	"github.com/tmplorg/tmplsvc/app/sdk/mid"
	"github.com/tmplorg/tmplsvc/business/domain/userbus"
	"github.com/tmplorg/tmplsvc/foundation/logger"
	"github.com/tmplorg/tmplsvc/foundation/web"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log     *logger.Logger
	UserBus userbus.ExtBusiness
	Auth    *auth.Auth
}

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Bearer(cfg.Auth)
	ruleAdmin := mid.Authorize(cfg.Auth, auth.RuleAdminOnly)
	ruleAuthorizeUser := mid.AuthorizeUser(cfg.Auth, cfg.UserBus, auth.RuleAdminOrSubject)
	ruleAuthorizeAdmin := mid.AuthorizeUser(cfg.Auth, cfg.UserBus, auth.RuleAdminOnly)

	api := newApp(cfg.UserBus)

	app.HandlerFunc(http.MethodGet, version, "/users", api.query, authen, ruleAdmin)
	app.HandlerFunc(http.MethodGet, version, "/users/{user_id}", api.queryByID, authen, ruleAuthorizeUser)
	app.HandlerFunc(http.MethodPost, version, "/users", api.create, authen, ruleAdmin)
	app.HandlerFunc(http.MethodPut, version, "/users/role/{user_id}", api.updateRole, authen, ruleAuthorizeAdmin)
	app.HandlerFunc(http.MethodPut, version, "/users/{user_id}", api.update, authen, ruleAuthorizeUser)
	app.HandlerFunc(http.MethodDelete, version, "/users/{user_id}", api.delete, authen, ruleAuthorizeUser)
}
