package main

import (
	"context"
	"database/sql"
	"errors"
	"img/utils"
	"net/http"
)

func checkAuth(headers http.Header) (auth string) {
	auth = headers.Get("Authorisation")
	if auth == "" {
		auth = headers.Get("Authorization")
	}
	return
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := checkAuth(r.Header)
		if auth == "" {
			utils.WriteCodeError(w, http.StatusUnauthorized)
			return
		}

		var id int
		var enabled bool
		err := utils.DB.QueryRow(context.Background(), `
			SELECT u.id as user_id, u.enabled
				FROM users u
			JOIN tokens t ON u.id = t.user_id
				WHERE t.token = $1
		`, auth).Scan(&id, &enabled)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				utils.WriteCodeError(w, http.StatusUnauthorized)
				return
			}

			utils.InternalServerError(w, err)
			return
		}

		if !enabled {
			utils.WriteCodeError(w, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), utils.CtxUserID, id)
		next.ServeHTTP(w, r.WithContext(ctx))

	})
}

func DashAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := checkAuth(r.Header)
		if token == "" {
			utils.WriteCodeError(w, http.StatusUnauthorized)
			return
		}

		id, err := utils.VerifySessionToken(token)
		if err != nil {
			utils.WriteCodeError(w, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), utils.CtxUserID, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
