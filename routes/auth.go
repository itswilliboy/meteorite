package routes

import (
	"context"
	"encoding/json"
	"errors"
	"img/utils"
	"io"
	"net/http"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type registerUserReceive struct {
	Username string `json:"username"`
}

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var payload registerUserReceive
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()

	err := d.Decode(&payload)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}

	ctx := context.Background()

	var userId int
	var createdAt time.Time
	err = utils.DB.QueryRow(
		ctx,
		"INSERT INTO users (name) VALUES ($1) RETURNING id, created_at",
		payload.Username,
	).Scan(&userId, &createdAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			utils.WriteJSONError(w, http.StatusBadRequest, "Username already in use.")
			return
		}
		utils.InternalServerError(w, err)
		return
	}

	token, _ := utils.CreateToken(userId, createdAt)

	utils.DB.Exec(ctx, "INSERT INTO tokens VALUES ($1, $2)", userId, []byte(token))
	io.WriteString(w, "Token: "+token)

}
