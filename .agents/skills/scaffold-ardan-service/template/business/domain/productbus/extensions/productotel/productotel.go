// Package productotel provides an extension for productbus that adds
// otel tracking.
package productotel

import (
	"context"

	"github.com/google/uuid"
	"github.com/tmplorg/tmplsvc/business/domain/productbus"
	"github.com/tmplorg/tmplsvc/business/sdk/order"
	"github.com/tmplorg/tmplsvc/business/sdk/page"
	"github.com/tmplorg/tmplsvc/business/sdk/sqldb"
	"github.com/tmplorg/tmplsvc/foundation/otel"
)

// Extension provides a wrapper for otel functionality around the productbus.
type Extension struct {
	bus productbus.ExtBusiness
}

// NewExtension constructs a new extension that wraps the productbus with otel.
func NewExtension() productbus.Extension {
	return func(bus productbus.ExtBusiness) productbus.ExtBusiness {
		return &Extension{
			bus: bus,
		}
	}
}

// NewWithTx does not apply otel.
func (ext *Extension) NewWithTx(tx sqldb.CommitRollbacker) (productbus.ExtBusiness, error) {
	return ext.bus.NewWithTx(tx)
}

// Create applies otel to the product creation process.
func (ext *Extension) Create(ctx context.Context, np productbus.NewProduct) (productbus.Product, error) {
	ctx, span := otel.AddSpan(ctx, "business.productbus.create")
	defer span.End()

	usr, err := ext.bus.Create(ctx, np)
	if err != nil {
		return productbus.Product{}, err
	}

	return usr, nil
}

// Update applies otel to the product update process.
func (ext *Extension) Update(ctx context.Context, prd productbus.Product, up productbus.UpdateProduct) (productbus.Product, error) {
	ctx, span := otel.AddSpan(ctx, "business.productbus.update")
	defer span.End()

	usr, err := ext.bus.Update(ctx, prd, up)
	if err != nil {
		return productbus.Product{}, err
	}

	return usr, nil
}

// Delete applies otel to the product deletion process.
func (ext *Extension) Delete(ctx context.Context, prd productbus.Product) error {
	ctx, span := otel.AddSpan(ctx, "business.productbus.delete")
	defer span.End()

	if err := ext.bus.Delete(ctx, prd); err != nil {
		return err
	}

	return nil
}

// Query applies otel to the product query process.
func (ext *Extension) Query(ctx context.Context, filter productbus.QueryFilter, orderBy order.By, page page.Page) ([]productbus.Product, error) {
	ctx, span := otel.AddSpan(ctx, "business.productbus.query")
	defer span.End()

	return ext.bus.Query(ctx, filter, orderBy, page)
}

// Count applies otel to the product count process.
func (ext *Extension) Count(ctx context.Context, filter productbus.QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.productbus.count")
	defer span.End()

	return ext.bus.Count(ctx, filter)
}

// QueryByID applies otel to the product query by ID process.
func (ext *Extension) QueryByID(ctx context.Context, userID uuid.UUID) (productbus.Product, error) {
	ctx, span := otel.AddSpan(ctx, "business.productbus.querybyid")
	defer span.End()

	return ext.bus.QueryByID(ctx, userID)
}

func (ext *Extension) QueryByUserID(ctx context.Context, userID uuid.UUID) ([]productbus.Product, error) {
	ctx, span := otel.AddSpan(ctx, "business.productbus.querybyuserid")
	defer span.End()

	return ext.bus.QueryByUserID(ctx, userID)
}
