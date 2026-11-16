package repository

import (
	"context"

	"hafiz-pintar-store/backend/api/internal/database/sqlc"
)

type CustomerRepository struct {
	queries *sqlc.Queries
}

func NewCustomerRepository(
	queries *sqlc.Queries,
) *CustomerRepository {
	return &CustomerRepository{
		queries: queries,
	}
}

// =========================================
// CUSTOMER: AUTHENTICATION
// =========================================

func (r *CustomerRepository) GetByEmail(
	ctx context.Context,
	email *string,
) (sqlc.Customer, error) {
	return r.queries.GetCustomerByEmail(ctx, email)
}

func (r *CustomerRepository) GetByPhone(
	ctx context.Context,
	phone *string,
) (sqlc.Customer, error) {
	return r.queries.GetCustomerByPhone(ctx, phone)
}

func (r *CustomerRepository) GetByEmailOrPhone(
	ctx context.Context,
	identifier *string,
) (sqlc.Customer, error) {
	return r.queries.GetCustomerByEmailOrPhone(ctx, identifier)
}

// =========================================
// CUSTOMER: REGISTRATION
// =========================================

func (r *CustomerRepository) Create(
	ctx context.Context,
	fullName string,
	email *string,
	phone *string,
	passwordHash *string,
) (sqlc.CreateCustomerRow, error) {
	return r.queries.CreateCustomer(
		ctx,
		sqlc.CreateCustomerParams{
			FullName:     fullName,
			Email:        email,
			Phone:        phone,
			PasswordHash: passwordHash,
		},
	)
}

// =========================================
// CUSTOMER: PROFILE
// =========================================

func (r *CustomerRepository) GetByID(
	ctx context.Context,
	id int64,
) (sqlc.GetCustomerByIDRow, error) {
	return r.queries.GetCustomerByID(ctx, id)
}

func (r *CustomerRepository) UpdateProfile(
	ctx context.Context,
	id int64,
	fullName string,
	email *string,
	phone *string,
) (sqlc.UpdateCustomerProfileRow, error) {
	return r.queries.UpdateCustomerProfile(
		ctx,
		sqlc.UpdateCustomerProfileParams{
			ID:       id,
			FullName: fullName,
			Email:    email,
			Phone:    phone,
		},
	)
}

// =========================================
// CUSTOMER: PASSWORD
// =========================================

func (r *CustomerRepository) UpdatePassword(
	ctx context.Context,
	id int64,
	passwordHash *string,
) (int64, error) {
	return r.queries.UpdateCustomerPassword(
		ctx,
		sqlc.UpdateCustomerPasswordParams{
			ID:           id,
			PasswordHash: passwordHash,
		},
	)
}

// =========================================
// CUSTOMER: ACCOUNT STATUS
// =========================================

func (r *CustomerRepository) Activate(
	ctx context.Context,
	id int64,
) (int64, error) {
	return r.queries.ActivateCustomer(ctx, id)
}

func (r *CustomerRepository) Deactivate(
	ctx context.Context,
	id int64,
) (int64, error) {
	return r.queries.DeactivateCustomer(ctx, id)
}

// =========================================
// CUSTOMER: ADMIN MANAGEMENT
// =========================================

func (r *CustomerRepository) List(
	ctx context.Context,
	limit int32,
	offset int32,
) ([]sqlc.ListCustomersRow, error) {
	return r.queries.ListCustomers(
		ctx,
		sqlc.ListCustomersParams{
			Limit:  limit,
			Offset: offset,
		},
	)
}

func (r *CustomerRepository) Search(
	ctx context.Context,
	keyword string,
	limit int32,
	offset int32,
) ([]sqlc.SearchCustomersRow, error) {
	return r.queries.SearchCustomers(
		ctx,
		sqlc.SearchCustomersParams{
			Keyword:     keyword,
			LimitCount:  limit,
			OffsetCount: offset,
		},
	)
}

func (r *CustomerRepository) Count(
	ctx context.Context,
) (int64, error) {
	return r.queries.CountCustomers(ctx)
}
