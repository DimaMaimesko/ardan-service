// Package authapp maintains the web based api for auth access.
package authapp

import (
	"context"
	"errors"
	"net/http"

	"github.com/tmplorg/tmplsvc/app/sdk/auth"
	"github.com/tmplorg/tmplsvc/app/sdk/errs"
	"github.com/tmplorg/tmplsvc/app/sdk/mid"
	"github.com/tmplorg/tmplsvc/foundation/web"
)

type app struct {
	auth *auth.Auth
}

func newApp(ath *auth.Auth) *app {
	return &app{
		auth: ath,
	}
}

// token generates a JWT for the claims the Basic middleware produced after
// it verified the caller's email and password.
func (a *app) token(ctx context.Context, r *http.Request) web.Encoder {
	kid := web.Param(r, "kid")
	if kid == "" {
		return errs.NewFieldErrors("kid", errors.New("missing kid"))
	}

	claims := mid.GetClaims(ctx)

	tkn, err := a.auth.GenerateToken(kid, claims)
	if err != nil {
		return errs.New(errs.Internal, err)
	}

	return token{Token: tkn}
}
