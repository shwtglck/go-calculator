package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		header := r.Header.Get("Authorization")

		if header == "" {
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}

		if !strings.HasPrefix(header, "Bearer ") {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(header, "Bearer ")

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})
		
		if err != nil || !token.Valid {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		
		claims, ok := token.Claims.(jwt.MapClaims)
if !ok {
	http.Error(w, "invalid token", http.StatusUnauthorized)
	return
}

userIDFloat, ok := claims["user_id"].(float64)
if !ok {
	http.Error(w, "invalid token", http.StatusUnauthorized)
	return
}

ctx := context.WithValue(r.Context(), "user_id", int(userIDFloat))

next.ServeHTTP(w, r.WithContext(ctx))
})
}