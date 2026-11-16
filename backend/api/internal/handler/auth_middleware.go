package handler

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const customerIDContextKey contextKey = "customer_id"

// JWTAuth memvalidasi JWT sebelum request diteruskan ke handler.
func JWTAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		secret := os.Getenv("JWT_SECRET")

		if len(secret) < 32 {
			customerError(
				w,
				http.StatusInternalServerError,
				"authentication service is not configured",
			)
			return
		}

		// Ambil Authorization header.
		authHeader := r.Header.Get("Authorization")
		parts := strings.Fields(authHeader)

		if len(parts) != 2 ||
			!strings.EqualFold(parts[0], "Bearer") {
			customerError(
				w,
				http.StatusUnauthorized,
				"authorization token is required",
			)
			return
		}

		tokenString := parts[1]

		claims := &jwt.RegisteredClaims{}

		token, err := jwt.ParseWithClaims(
			tokenString,
			claims,
			func(token *jwt.Token) (any, error) {
				return []byte(secret), nil
			},
			jwt.WithValidMethods([]string{
				jwt.SigningMethodHS256.Alg(),
			}),
			jwt.WithIssuer("hafiz-pintar-api"),
		)

		if err != nil || token == nil || !token.Valid {
			customerError(
				w,
				http.StatusUnauthorized,
				"invalid or expired access token",
			)
			return
		}

		// Ambil customer ID dari JWT subject.
		customerID, err := strconv.ParseInt(
			claims.Subject,
			10,
			64,
		)

		if err != nil || customerID <= 0 {
			customerError(
				w,
				http.StatusUnauthorized,
				"invalid access token subject",
			)
			return
		}

		// Simpan customer ID ke request context.
		ctx := context.WithValue(
			r.Context(),
			customerIDContextKey,
			customerID,
		)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// CustomerIDFromContext mengambil customer ID dari JWT context.
func CustomerIDFromContext(ctx context.Context) (int64, bool) {
	customerID, ok := ctx.Value(customerIDContextKey).(int64)
	return customerID, ok
}
