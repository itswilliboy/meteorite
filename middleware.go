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
		err := utils.DB.QueryRow(context.Background(), "SELECT user_id FROM tokens WHERE token = $1", auth).Scan(&id)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				utils.WriteCodeError(w, http.StatusUnauthorized)
				return
			}
			utils.InternalServerError(w, err)
			return
		}

		ctx := context.WithValue(r.Context(), "userId", id)
		next.ServeHTTP(w, r.WithContext(ctx))

	})
}
