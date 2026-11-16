package repository

import (
	"context"
	"errors"

	"hafiz-pintar-store/backend/api/internal/database/sqlc"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrCartNotFound              = errors.New("cart not found")
	ErrCartItemNotFound          = errors.New("cart item not found")
	ErrInvalidCartQuantity       = errors.New("cart quantity must be greater than zero")
	ErrProductVariantUnavailable = errors.New("product variant unavailable")
)

type CartRepository struct {
	queries *sqlc.Queries
	pool    *pgxpool.Pool
}

func NewCartRepository(
	queries *sqlc.Queries,
	pool *pgxpool.Pool,
) *CartRepository {
	return &CartRepository{
		queries: queries,
		pool:    pool,
	}
}

// GetOrCreate mengambil keranjang customer atau membuatnya jika belum ada.
func (r *CartRepository) GetOrCreate(
	ctx context.Context,
	customerID int64,
) (sqlc.Cart, error) {
	return r.queries.GetOrCreateCart(ctx, customerID)
}

// GetItems mengambil seluruh item dalam keranjang.
func (r *CartRepository) GetItems(
	ctx context.Context,
	cartID int64,
) ([]sqlc.GetCartItemsRow, error) {
	return r.queries.GetCartItems(ctx, cartID)
}

// AddItem menambahkan varian produk ke keranjang.
// Jika varian sudah ada, quantity akan ditambahkan.
func (r *CartRepository) AddItem(
	ctx context.Context,
	cartID int64,
	productID int64,
	variantID int64,
	quantity int32,
) (sqlc.CartItem, error) {
	if quantity <= 0 {
		return sqlc.CartItem{}, ErrInvalidCartQuantity
	}

	item, err := r.queries.AddCartItem(
		ctx,
		sqlc.AddCartItemParams{
			CartID:    cartID,
			ProductID: productID,
			ID:        variantID,
			Quantity:  quantity,
		},
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return sqlc.CartItem{}, ErrProductVariantUnavailable
	}

	return item, err
}

// UpdateQuantity mengubah quantity item dalam keranjang.
func (r *CartRepository) UpdateQuantity(
	ctx context.Context,
	cartID int64,
	itemID int64,
	quantity int32,
) (sqlc.CartItem, error) {
	if quantity <= 0 {
		return sqlc.CartItem{}, ErrInvalidCartQuantity
	}

	item, err := r.queries.UpdateCartItemQuantity(
		ctx,
		sqlc.UpdateCartItemQuantityParams{
			ID:       itemID,
			CartID:   cartID,
			Quantity: quantity,
		},
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return sqlc.CartItem{}, ErrCartItemNotFound
	}

	return item, err
}

// RemoveItem menghapus satu item dari keranjang.
func (r *CartRepository) RemoveItem(
	ctx context.Context,
	cartID int64,
	itemID int64,
) error {
	rowsAffected, err := r.queries.RemoveCartItem(
		ctx,
		sqlc.RemoveCartItemParams{
			ID:     itemID,
			CartID: cartID,
		},
	)

	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrCartItemNotFound
	}

	return nil
}

// Clear menghapus seluruh item dalam keranjang.
func (r *CartRepository) Clear(
	ctx context.Context,
	cartID int64,
) error {
	_, err := r.queries.ClearCart(ctx, cartID)
	return err
}
