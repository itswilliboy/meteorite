package main

import (
	"context"
	"database/sql"
	"errors"
	"img/utils"
	"net/http"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorisation")
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

		ctx := context.WithValue(r.Context(), "userId", id)
		next.ServeHTTP(w, r.WithContext(ctx))

	})
}
