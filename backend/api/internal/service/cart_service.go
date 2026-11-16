package service

import (
	"context"
	"errors"

	"hafiz-pintar-store/backend/api/internal/database/sqlc"
	"hafiz-pintar-store/backend/api/internal/repository"
)

var (
	ErrInvalidCartCustomerID  = errors.New("customer ID must be greater than zero")
	ErrInvalidCartProductID   = errors.New("product ID must be greater than zero")
	ErrInvalidCartVariantID   = errors.New("variant ID must be greater than zero")
	ErrInvalidCartItemID      = errors.New("cart item ID must be greater than zero")
	ErrInvalidCartQuantity    = errors.New("cart quantity must be greater than zero")
	ErrCartInsufficientStock  = errors.New("requested quantity exceeds available stock")
	ErrCartVariantUnavailable = errors.New("product variant is unavailable")
	ErrCartItemNotFound       = errors.New("cart item not found")
)

type CartService struct {
	repository *repository.CartRepository
}

type CartDetail struct {
	Cart      sqlc.Cart              `json:"cart"`
	Items     []sqlc.GetCartItemsRow `json:"items"`
	TotalItem int32                  `json:"total_item"`
	Subtotal  int64                  `json:"subtotal"`
}

func NewCartService(
	repo *repository.CartRepository,
) *CartService {
	return &CartService{
		repository: repo,
	}
}

// Get mengambil keranjang beserta seluruh item milik customer.
func (s *CartService) Get(
	ctx context.Context,
	customerID int64,
) (CartDetail, error) {
	if customerID < 1 {
		return CartDetail{}, ErrInvalidCartCustomerID
	}

	cart, err := s.repository.GetOrCreate(ctx, customerID)
	if err != nil {
		return CartDetail{}, err
	}

	items, err := s.repository.GetItems(ctx, cart.ID)
	if err != nil {
		return CartDetail{}, err
	}

	result := CartDetail{
		Cart:      cart,
		Items:     items,
		TotalItem: 0,
		Subtotal:  0,
	}

	for _, item := range items {
		result.TotalItem += item.Quantity
		result.Subtotal += item.Subtotal
	}

	return result, nil
}

// AddItem menambahkan produk varian ke keranjang.
func (s *CartService) AddItem(
	ctx context.Context,
	customerID int64,
	productID int64,
	variantID int64,
	quantity int32,
) (sqlc.CartItem, error) {
	if customerID < 1 {
		return sqlc.CartItem{}, ErrInvalidCartCustomerID
	}

	if productID < 1 {
		return sqlc.CartItem{}, ErrInvalidCartProductID
	}

	if variantID < 1 {
		return sqlc.CartItem{}, ErrInvalidCartVariantID
	}

	if quantity < 1 {
		return sqlc.CartItem{}, ErrInvalidCartQuantity
	}

	cart, err := s.repository.GetOrCreate(ctx, customerID)
	if err != nil {
		return sqlc.CartItem{}, err
	}

	item, err := s.repository.AddItem(
		ctx,
		cart.ID,
		productID,
		variantID,
		quantity,
	)

	if errors.Is(err, repository.ErrProductVariantUnavailable) {
		return sqlc.CartItem{}, ErrCartVariantUnavailable
	}

	if err != nil {
		return sqlc.CartItem{}, err
	}

	return item, nil
}

// UpdateQuantity memperbarui jumlah item dalam keranjang.
func (s *CartService) UpdateQuantity(
	ctx context.Context,
	customerID int64,
	itemID int64,
	quantity int32,
) (sqlc.CartItem, error) {
	if customerID < 1 {
		return sqlc.CartItem{}, ErrInvalidCartCustomerID
	}

	if itemID < 1 {
		return sqlc.CartItem{}, ErrInvalidCartItemID
	}

	if quantity < 1 {
		return sqlc.CartItem{}, ErrInvalidCartQuantity
	}

	cart, err := s.repository.GetOrCreate(ctx, customerID)
	if err != nil {
		return sqlc.CartItem{}, err
	}

	items, err := s.repository.GetItems(ctx, cart.ID)
	if err != nil {
		return sqlc.CartItem{}, err
	}

	found := false

	for _, item := range items {
		if item.ID != itemID {
			continue
		}

		found = true

		if !item.VariantIsActive {
			return sqlc.CartItem{}, ErrCartVariantUnavailable
		}

		if quantity > item.StockQty {
			return sqlc.CartItem{}, ErrCartInsufficientStock
		}

		break
	}

	if !found {
		return sqlc.CartItem{}, ErrCartItemNotFound
	}

	return s.repository.UpdateQuantity(
		ctx,
		cart.ID,
		itemID,
		quantity,
	)
}

// RemoveItem menghapus satu item dari keranjang.
func (s *CartService) RemoveItem(
	ctx context.Context,
	customerID int64,
	itemID int64,
) error {
	if customerID < 1 {
		return ErrInvalidCartCustomerID
	}

	if itemID < 1 {
		return ErrInvalidCartItemID
	}

	cart, err := s.repository.GetOrCreate(ctx, customerID)
	if err != nil {
		return err
	}

	err = s.repository.RemoveItem(ctx, cart.ID, itemID)
	if errors.Is(err, repository.ErrCartItemNotFound) {
		return ErrCartItemNotFound
	}

	return err
}

// Clear menghapus seluruh item dalam keranjang.
func (s *CartService) Clear(
	ctx context.Context,
	customerID int64,
) error {
	if customerID < 1 {
		return ErrInvalidCartCustomerID
	}

	cart, err := s.repository.GetOrCreate(ctx, customerID)
	if err != nil {
		return err
	}

	return s.repository.Clear(ctx, cart.ID)
}
