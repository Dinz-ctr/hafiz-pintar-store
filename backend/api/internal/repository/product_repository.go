package repository

import (
	"context"
	"errors"

	"hafiz-pintar-store/backend/api/internal/database/sqlc"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrInvalidDefaultVariant = errors.New(
		"product must have exactly one default variant",
	)

	ErrProductVariantNotFound = errors.New(
		"product variant not found or cannot be deleted",
	)
)

type ProductRepository struct {
	queries *sqlc.Queries
	pool    *pgxpool.Pool
}

// =========================================
// INPUT STRUCTS
// =========================================

type CreateProductInput struct {
	CategoryID  *int64
	Name        string
	Slug        string
	Description *string
	Variants    []CreateProductVariantInput
}

type CreateProductVariantInput struct {
	Sku        string
	Name       string
	Price      int64
	StockQty   int32
	WeightGram int32
	IsDefault  bool
	Attributes map[string]string
}

// =========================================
// RESPONSE STRUCTS
// =========================================

type ProductVariantDetail struct {
	ID         int64             `json:"id"`
	ProductID  int64             `json:"product_id"`
	Sku        string            `json:"sku"`
	Name       string            `json:"name"`
	Price      int64             `json:"price"`
	StockQty   int32             `json:"stock_qty"`
	WeightGram int32             `json:"weight_gram"`
	IsDefault  bool              `json:"is_default"`
	Attributes map[string]string `json:"attributes"`
}

// =========================================
// CONSTRUCTOR
// =========================================

func NewProductRepository(
	queries *sqlc.Queries,
	pool *pgxpool.Pool,
) *ProductRepository {
	return &ProductRepository{
		queries: queries,
		pool:    pool,
	}
}

// =========================================
// PRODUCT: READ
// =========================================

// ListActive mengambil daftar produk aktif.
func (r *ProductRepository) ListActive(
	ctx context.Context,
	limit int32,
	offset int32,
) ([]sqlc.ListActiveProductsRow, error) {
	return r.queries.ListActiveProducts(
		ctx,
		sqlc.ListActiveProductsParams{
			Limit:  limit,
			Offset: offset,
		},
	)
}

// GetBySlug mengambil detail produk berdasarkan slug.
func (r *ProductRepository) GetBySlug(
	ctx context.Context,
	slug string,
) (sqlc.GetProductBySlugRow, error) {
	return r.queries.GetProductBySlug(ctx, slug)
}

// Search mencari produk aktif berdasarkan keyword dan kategori.
func (r *ProductRepository) Search(
	ctx context.Context,
	keyword string,
	categoryID *int64,
	limit int32,
	offset int32,
) ([]sqlc.SearchActiveProductsRow, error) {
	return r.queries.SearchActiveProducts(
		ctx,
		sqlc.SearchActiveProductsParams{
			Keyword:     keyword,
			CategoryID:  categoryID,
			LimitCount:  limit,
			OffsetCount: offset,
		},
	)
}

// =========================================
// PRODUCT: CREATE
// =========================================

// Create membuat produk beserta seluruh variannya
// dalam satu transaksi PostgreSQL.
func (r *ProductRepository) Create(
	ctx context.Context,
	input CreateProductInput,
) (sqlc.CreateProductBaseRow, error) {

	// Pastikan hanya ada satu varian default.
	var defaultVariant *CreateProductVariantInput

	for i := range input.Variants {
		if input.Variants[i].IsDefault {
			if defaultVariant != nil {
				return sqlc.CreateProductBaseRow{},
					ErrInvalidDefaultVariant
			}

			defaultVariant = &input.Variants[i]
		}
	}

	if defaultVariant == nil {
		return sqlc.CreateProductBaseRow{},
			ErrInvalidDefaultVariant
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return sqlc.CreateProductBaseRow{}, err
	}

	defer func() {
		_ = tx.Rollback(context.Background())
	}()

	qtx := r.queries.WithTx(tx)

	// 1. Buat identitas produk.
	// Kolom legacy mengikuti varian default.
	product, err := qtx.CreateProductBase(
		ctx,
		sqlc.CreateProductBaseParams{
			CategoryID:  input.CategoryID,
			Sku:         defaultVariant.Sku,
			Name:        input.Name,
			Slug:        input.Slug,
			Description: input.Description,
			Price:       defaultVariant.Price,
			StockQty:    defaultVariant.StockQty,
			WeightGram:  defaultVariant.WeightGram,
		},
	)
	if err != nil {
		return sqlc.CreateProductBaseRow{}, err
	}

	attributeIDs := make(map[string]int64)

	// 2. Buat seluruh varian.
	for _, variant := range input.Variants {
		createdVariant, err := qtx.CreateProductVariant(
			ctx,
			sqlc.CreateProductVariantParams{
				ProductID:  product.ID,
				Sku:        variant.Sku,
				Name:       variant.Name,
				Price:      variant.Price,
				StockQty:   variant.StockQty,
				WeightGram: variant.WeightGram,
				IsDefault:  variant.IsDefault,
			},
		)
		if err != nil {
			return sqlc.CreateProductBaseRow{}, err
		}

		// 3. Buat atribut dan nilai atribut.
		for attributeName, attributeValue := range variant.Attributes {
			attributeID, exists := attributeIDs[attributeName]

			if !exists {
				attribute, err := qtx.CreateVariantAttribute(
					ctx,
					sqlc.CreateVariantAttributeParams{
						ProductID: product.ID,
						Name:      attributeName,
					},
				)
				if err != nil {
					return sqlc.CreateProductBaseRow{}, err
				}

				attributeID = attribute.ID
				attributeIDs[attributeName] = attributeID
			}

			_, err = qtx.CreateVariantAttributeValue(
				ctx,
				sqlc.CreateVariantAttributeValueParams{
					VariantID:   createdVariant.ID,
					ProductID:   product.ID,
					AttributeID: attributeID,
					Value:       attributeValue,
				},
			)
			if err != nil {
				return sqlc.CreateProductBaseRow{}, err
			}
		}
	}

	// 4. Commit transaksi.
	if err := tx.Commit(ctx); err != nil {
		return sqlc.CreateProductBaseRow{}, err
	}

	return product, nil
}

// =========================================
// PRODUCT: UPDATE
// =========================================

// UpdateDetails memperbarui informasi dasar produk.
func (r *ProductRepository) UpdateDetails(
	ctx context.Context,
	productID int64,
	name string,
	slug string,
	description *string,
) (sqlc.UpdateProductDetailsRow, error) {
	return r.queries.UpdateProductDetails(
		ctx,
		sqlc.UpdateProductDetailsParams{
			ID:          productID,
			Name:        name,
			Slug:        slug,
			Description: description,
		},
	)
}

// UpdateVariant memperbarui data varian produk.
// Jika varian default diperbarui, kolom legacy ikut disinkronkan.
func (r *ProductRepository) UpdateVariant(
	ctx context.Context,
	productID int64,
	variantID int64,
	name string,
	price int64,
	stockQty int32,
	weightGram int32,
) (sqlc.UpdateProductVariantRow, error) {

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return sqlc.UpdateProductVariantRow{}, err
	}

	defer func() {
		_ = tx.Rollback(context.Background())
	}()

	qtx := r.queries.WithTx(tx)

	variant, err := qtx.UpdateProductVariant(
		ctx,
		sqlc.UpdateProductVariantParams{
			ID:         variantID,
			ProductID:  productID,
			Name:       name,
			Price:      price,
			StockQty:   stockQty,
			WeightGram: weightGram,
		},
	)
	if err != nil {
		return sqlc.UpdateProductVariantRow{}, err
	}

	if variant.IsDefault {
		_, err = tx.Exec(
			ctx,
			`UPDATE products
			SET price = $2,
			    stock_qty = $3,
			    weight_gram = $4,
			    updated_at = NOW()
			WHERE id = $1`,
			productID,
			price,
			stockQty,
			weightGram,
		)
		if err != nil {
			return sqlc.UpdateProductVariantRow{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return sqlc.UpdateProductVariantRow{}, err
	}

	return variant, nil
}

// =========================================
// PRODUCT: DELETE
// =========================================

// DeleteVariant hanya menghapus varian non-default.
func (r *ProductRepository) DeleteVariant(
	ctx context.Context,
	productID int64,
	variantID int64,
) error {

	rowsAffected, err := r.queries.DeleteProductVariant(
		ctx,
		sqlc.DeleteProductVariantParams{
			ID:        variantID,
			ProductID: productID,
		},
	)
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrProductVariantNotFound
	}

	return nil
}

// Delete menghapus produk berdasarkan ID.
func (r *ProductRepository) Delete(
	ctx context.Context,
	productID int64,
) error {

	rowsAffected, err := r.queries.DeleteProduct(
		ctx,
		productID,
	)
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// =========================================
// PRODUCT VARIANTS: READ
// =========================================

// ListVariants mengambil seluruh varian aktif suatu produk.
func (r *ProductRepository) ListVariants(
	ctx context.Context,
	productID int64,
) ([]sqlc.ListProductVariantsRow, error) {
	return r.queries.ListProductVariants(ctx, productID)
}

// ListVariantAttributes mengambil atribut suatu varian.
func (r *ProductRepository) ListVariantAttributes(
	ctx context.Context,
	variantID int64,
) ([]sqlc.ListVariantAttributesRow, error) {
	return r.queries.ListVariantAttributes(ctx, variantID)
}

// ListVariantsWithAttributes mengambil seluruh varian aktif
// beserta atribut masing-masing varian.
func (r *ProductRepository) ListVariantsWithAttributes(
	ctx context.Context,
	productID int64,
) ([]ProductVariantDetail, error) {

	variants, err := r.ListVariants(ctx, productID)
	if err != nil {
		return nil, err
	}

	result := make(
		[]ProductVariantDetail,
		0,
		len(variants),
	)

	for _, variant := range variants {
		attributes, err := r.ListVariantAttributes(
			ctx,
			variant.ID,
		)
		if err != nil {
			return nil, err
		}

		attributeMap := make(
			map[string]string,
			len(attributes),
		)

		for _, attribute := range attributes {
			attributeMap[attribute.AttributeName] =
				attribute.AttributeValue
		}

		result = append(result, ProductVariantDetail{
			ID:         variant.ID,
			ProductID:  variant.ProductID,
			Sku:        variant.Sku,
			Name:       variant.Name,
			Price:      variant.Price,
			StockQty:   variant.StockQty,
			WeightGram: variant.WeightGram,
			IsDefault:  variant.IsDefault,
			Attributes: attributeMap,
		})
	}

	return result, nil
}
