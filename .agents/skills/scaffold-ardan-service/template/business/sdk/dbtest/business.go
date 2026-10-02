package dbtest

import (
	"github.com/jmoiron/sqlx"
	"github.com/tmplorg/tmplsvc/business/domain/productbus"
	"github.com/tmplorg/tmplsvc/business/domain/productbus/extensions/productotel"
	"github.com/tmplorg/tmplsvc/business/domain/productbus/stores/productdb"
	"github.com/tmplorg/tmplsvc/business/domain/userbus"
	"github.com/tmplorg/tmplsvc/business/domain/userbus/extensions/userotel"
	"github.com/tmplorg/tmplsvc/business/domain/userbus/stores/userdb"
	"github.com/tmplorg/tmplsvc/business/sdk/delegate"
	"github.com/tmplorg/tmplsvc/foundation/logger"
)

// BusDomain represents all the business domain apis needed for testing.
type BusDomain struct {
	Delegate *delegate.Delegate
	Product  productbus.ExtBusiness
	User     userbus.ExtBusiness
}

// newBusDomains constructs the business domain apis the same way the
// service main.go constructs them, so tests exercise the real wiring.
func newBusDomains(log *logger.Logger, db *sqlx.DB) BusDomain {
	delegate := delegate.New(log)

	userOtelExt := userotel.NewExtension()
	userStorage := userdb.NewStore(log, db)
	userBus := userbus.NewBusiness(log, delegate, userStorage, userOtelExt)

	productOtelExt := productotel.NewExtension()
	productStorage := productdb.NewStore(log, db)
	productBus := productbus.NewBusiness(log, userBus, delegate, productStorage, productOtelExt)

	return BusDomain{
		Delegate: delegate,
		Product:  productBus,
		User:     userBus,
	}
}
