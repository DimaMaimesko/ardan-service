package mid

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/tmplorg/tmplsvc/app/sdk/auth"
	"github.com/tmplorg/tmplsvc/app/sdk/errs"
	"github.com/tmplorg/tmplsvc/business/domain/userbus"
	"github.com/tmplorg/tmplsvc/business/types/role"
	"github.com/tmplorg/tmplsvc/foundation/web"
)

// Bearer processes JWT authentication logic. The token is validated in
// process against the keystore and the OPA authentication policy.
func Bearer(ath *auth.Auth) web.MidFunc {
	m := func(next web.HandlerFunc) web.HandlerFunc {
		h := func(ctx context.Context, r *http.Request) web.Encoder {
			claims, err := ath.Authenticate(ctx, r.Header.Get("authorization"))
			if err != nil {
				return errs.New(errs.Unauthenticated, err)
			}

			if claims.Subject == "" {
				return errs.Errorf(errs.Unauthenticated, "authorize: you are not authorized for that action, no claims")
			}

			subjectID, err := uuid.Parse(claims.Subject)
			if err != nil {
				return errs.Errorf(errs.Unauthenticated, "parsing subject: %s", err)
			}

			ctx = setUserID(ctx, subjectID)
			ctx = setClaims(ctx, claims)

			return next(ctx, r)
		}

		return h
	}

	return m
}

// Basic processes basic authentication logic. It is used to exchange an
// email and password for a set of claims that can be turned into a token.
func Basic(ath *auth.Auth, userBus userbus.ExtBusiness) web.MidFunc {
	m := func(next web.HandlerFunc) web.HandlerFunc {
		h := func(ctx context.Context, r *http.Request) web.Encoder {
			email, pass, ok := parseBasicAuth(r.Header.Get("authorization"))
			if !ok {
				return errs.Errorf(errs.Unauthenticated, "invalid Basic auth")
			}

			addr, err := mail.ParseAddress(email)
			if err != nil {
				return errs.New(errs.Unauthenticated, err)
			}

			usr, err := userBus.Authenticate(ctx, *addr, pass)
			if err != nil {
				return errs.New(errs.Unauthenticated, err)
			}

			claims := auth.Claims{
				RegisteredClaims: jwt.RegisteredClaims{
					Subject:   usr.ID.String(),
					Issuer:    ath.Issuer(),
					ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(8760 * time.Hour)),
					IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
				},
				Roles: role.ParseToString(usr.Roles),
			}

			ctx = setUserID(ctx, usr.ID)
			ctx = setClaims(ctx, claims)

			return next(ctx, r)
		}

		return h
	}

	return m
}

func parseBasicAuth(auth string) (string, string, bool) {
	parts := strings.Split(auth, " ")
	if len(parts) != 2 || parts[0] != "Basic" {
		return "", "", false
	}

	c, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", "", false
	}

	username, password, ok := strings.Cut(string(c), ":")
	if !ok {
		return "", "", false
	}

	return username, password, true
}
