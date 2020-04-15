package users

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

type Token struct {
	UserID   uint
	Name     string
	Username string
	*jwt.StandardClaims
}

// JwtVerify check the the JWT token provided in the Headers
// is valid and current
func JwtVerify(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var header = r.Header.Get("x-access-token")
		var resp map[string]interface{}
		header = strings.TrimSpace(header)

		if header == "" {
			// No token was provided, so immediate fail
			w.WriteHeader(http.StatusForbidden)
			resp = map[string]interface{}{"status": false, "message": "Missing auth token"}
			json.NewEncoder(w).Encode(resp)
			return
		}

		tk := &Token{}

		_, err := jwt.ParseWithClaims(header, tk, func(token *jwt.Token) (interface{}, error) {
			return []byte("secret"), nil
		})
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			resp = map[string]interface{}{"status": false, "message": err.Error()}
			json.NewEncoder(w).Encode(resp)
			return
		}

		ctx := context.WithValue(r.Context(), "user", tk)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
