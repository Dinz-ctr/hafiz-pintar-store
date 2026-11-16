package repository

import (
	"context"
	"errors"

	"hafiz-pintar-store/backend/api/internal/database/sqlc"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrCustomerAddressNotFound = errors.New("customer address not found")

type CustomerAddressRepository struct {
	queries *sqlc.Queries
	pool    *pgxpool.Pool
}

func NewCustomerAddressRepository(
	queries *sqlc.Queries,
	pool *pgxpool.Pool,
) *CustomerAddressRepository {
	return &CustomerAddressRepository{
		queries: queries,
		pool:    pool,
	}
}

// List seluruh alamat milik customer.
func (r *CustomerAddressRepository) List(
	ctx context.Context,
	customerID int64,
) ([]sqlc.CustomerAddress, error) {
	return r.queries.ListCustomerAddresses(ctx, customerID)
}

// Ambil alamat berdasarkan ID dan pemiliknya.
func (r *CustomerAddressRepository) GetByID(
	ctx context.Context,
	id int64,
	customerID int64,
) (sqlc.CustomerAddress, error) {
	return r.queries.GetCustomerAddressByID(
		ctx,
		sqlc.GetCustomerAddressByIDParams{
			ID:         id,
			CustomerID: customerID,
		},
	)
}

// Ambil alamat utama customer.
func (r *CustomerAddressRepository) GetDefault(
	ctx context.Context,
	customerID int64,
) (sqlc.CustomerAddress, error) {
	return r.queries.GetDefaultCustomerAddress(ctx, customerID)
}

// Tambahkan alamat baru.
func (r *CustomerAddressRepository) Create(
	ctx context.Context,
	arg sqlc.CreateCustomerAddressParams,
) (sqlc.CustomerAddress, error) {
	return r.queries.CreateCustomerAddress(ctx, arg)
}

// Update alamat milik customer.
func (r *CustomerAddressRepository) Update(
	ctx context.Context,
	arg sqlc.UpdateCustomerAddressParams,
) (sqlc.CustomerAddress, error) {
	return r.queries.UpdateCustomerAddress(ctx, arg)
}

// Hapus alamat milik customer.
func (r *CustomerAddressRepository) Delete(
	ctx context.Context,
	id int64,
	customerID int64,
) (int64, error) {
	return r.queries.DeleteCustomerAddress(
		ctx,
		sqlc.DeleteCustomerAddressParams{
			ID:         id,
			CustomerID: customerID,
		},
	)
}

// Hitung jumlah alamat customer.
func (r *CustomerAddressRepository) Count(
	ctx context.Context,
	customerID int64,
) (int64, error) {
	return r.queries.CountCustomerAddresses(ctx, customerID)
}

// Atur alamat utama menggunakan transaksi.
func (r *CustomerAddressRepository) SetDefault(
	ctx context.Context,
	id int64,
	customerID int64,
) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	// Pastikan alamat memang milik customer.
	_, err = qtx.GetCustomerAddressByID(
		ctx,
		sqlc.GetCustomerAddressByIDParams{
			ID:         id,
			CustomerID: customerID,
		},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrCustomerAddressNotFound
		}
		return err
	}

	// Hapus status default sebelumnya.
	err = qtx.ClearDefaultCustomerAddress(ctx, customerID)
	if err != nil {
		return err
	}

	// Jadikan alamat pilihan sebagai default.
	affected, err := qtx.SetDefaultCustomerAddress(
		ctx,
		sqlc.SetDefaultCustomerAddressParams{
			ID:         id,
			CustomerID: customerID,
		},
	)
	if err != nil {
		return err
	}

	if affected == 0 {
		return ErrCustomerAddressNotFound
	}

	return tx.Commit(ctx)
}
