package mid

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/tmplorg/tmplsvc/app/sdk/auth"
	"github.com/tmplorg/tmplsvc/app/sdk/errs"
	"github.com/tmplorg/tmplsvc/business/domain/productbus"
	"github.com/tmplorg/tmplsvc/business/domain/userbus"
	"github.com/tmplorg/tmplsvc/foundation/web"
)

// ErrInvalidID represents a condition where the id is not a uuid.
var ErrInvalidID = errors.New("ID is not in its proper form")

// Authorize validates that the authenticated user satisfies the specified
// OPA rule.
func Authorize(ath *auth.Auth, rule string) web.MidFunc {
	m := func(next web.HandlerFunc) web.HandlerFunc {
		h := func(ctx context.Context, r *http.Request) web.Encoder {
			userID, err := GetUserID(ctx)
			if err != nil {
				return errs.New(errs.Unauthenticated, err)
			}

			if err := authorize(ctx, ath, userID, rule); err != nil {
				return err
			}

			return next(ctx, r)
		}

		return h
	}

	return m
}

// AuthorizeUser executes the specified rule and extracts the specified
// user from the DB if a user id is specified in the call. Depending on the rule
// specified, the userid from the claims may be compared with the specified
// user id.
func AuthorizeUser(ath *auth.Auth, userBus userbus.ExtBusiness, rule string) web.MidFunc {
	m := func(next web.HandlerFunc) web.HandlerFunc {
		h := func(ctx context.Context, r *http.Request) web.Encoder {
			id := web.Param(r, "user_id")

			var userID uuid.UUID

			if id != "" {
				var err error
				userID, err = uuid.Parse(id)
				if err != nil {
					return errs.New(errs.Unauthenticated, ErrInvalidID)
				}

				usr, err := userBus.QueryByID(ctx, userID)
				if err != nil {
					switch {
					case errors.Is(err, userbus.ErrNotFound):
						return errs.New(errs.Unauthenticated, err)
					default:
						return errs.Errorf(errs.Unauthenticated, "querybyid: userID[%s]: %s", userID, err)
					}
				}

				ctx = setUser(ctx, usr)
			}

			if err := authorize(ctx, ath, userID, rule); err != nil {
				return err
			}

			return next(ctx, r)
		}

		return h
	}

	return m
}

// AuthorizeProduct extracts the specified product from the DB if a product
// id is specified in the call. The owner of the product is compared with
// the subject in the claims unless the caller is an admin.
func AuthorizeProduct(ath *auth.Auth, productBus productbus.ExtBusiness) web.MidFunc {
	m := func(next web.HandlerFunc) web.HandlerFunc {
		h := func(ctx context.Context, r *http.Request) web.Encoder {
			id := web.Param(r, "product_id")

			var userID uuid.UUID

			if id != "" {
				productID, err := uuid.Parse(id)
				if err != nil {
					return errs.New(errs.Unauthenticated, ErrInvalidID)
				}

				prd, err := productBus.QueryByID(ctx, productID)
				if err != nil {
					switch {
					case errors.Is(err, productbus.ErrNotFound):
						return errs.New(errs.Unauthenticated, err)
					default:
						return errs.Errorf(errs.Unauthenticated, "querybyid: productID[%s]: %s", productID, err)
					}
				}

				userID = prd.UserID
				ctx = setProduct(ctx, prd)
			}

			if err := authorize(ctx, ath, userID, auth.RuleAdminOrSubject); err != nil {
				return err
			}

			return next(ctx, r)
		}

		return h
	}

	return m
}

// authorize evaluates the rule against the claims in the context. The OPA
// details stay out of the response so policy internals are not exposed to
// the client.
func authorize(ctx context.Context, ath *auth.Auth, userID uuid.UUID, rule string) *errs.Error {
	claims := GetClaims(ctx)

	if err := ath.Authorize(ctx, claims, userID, rule); err != nil {
		return errs.Errorf(errs.Unauthenticated, "authorize: you are not authorized for that action, claims[%v] rule[%v]", claims.Roles, rule)
	}

	return nil
}
