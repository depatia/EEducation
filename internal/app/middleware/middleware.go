package middleware

import (
	"APIGateway/pkg/tools/jwt"
	"context"
	"net/http"

	"strings"
)

func AuthMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := strings.Split(r.Header.Get("Authorization"), "Bearer ")

		if len(authHeader) != 2 {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("malformed token"))
			return
		}
		jwtToken := authHeader[1]

		claims, _ := jwt.ValidateToken(jwtToken)

		if claims.Uid == 0 {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("invalid token"))
			return
		}

		ctx := context.WithValue(r.Context(), "user_id", claims.Uid)
		ctx = context.WithValue(ctx, "permission_lvl", claims.Permissions)

		h(w, r.WithContext(ctx))
	}
}
