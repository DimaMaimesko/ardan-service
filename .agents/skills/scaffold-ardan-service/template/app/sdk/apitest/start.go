package apitest

import (
	"testing"

	"github.com/tmplorg/tmplsvc/api/services/tmplsvc/build"
	"github.com/tmplorg/tmplsvc/app/sdk/auth"
	"github.com/tmplorg/tmplsvc/app/sdk/mux"
	"github.com/tmplorg/tmplsvc/business/sdk/dbtest"
)

// New initialized the system to run a test.
func New(t *testing.T, testName string) *Test {
	db := dbtest.New(t, testName)

	// -------------------------------------------------------------------------

	ath := auth.New(auth.Config{
		Log:       db.Log,
		UserBus:   db.BusDomain.User,
		KeyLookup: &KeyStore{},
	})

	// -------------------------------------------------------------------------

	mux := mux.WebAPI(mux.Config{
		Log:  db.Log,
		DB:   db.DB,
		Auth: ath,
		BusConfig: mux.BusConfig{
			UserBus:    db.BusDomain.User,
			ProductBus: db.BusDomain.Product,
		},
	}, build.Routes())

	return &Test{
		DB:   db,
		Auth: ath,
		mux:  mux,
	}
}
