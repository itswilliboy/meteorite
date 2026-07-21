package routes

import (
	"errors"
	"img/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
)

type adminStatistics struct {
	TotalUsers   int   `json:"total_users"`
	ActiveUsers  int   `json:"active_users"`
	TotalMedia   int   `json:"total_media"`
	TotalStorage int64 `json:"total_storage"`
}

func AdminStatistics(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT
			(SELECT COUNT(*) FROM users),
			(SELECT COUNT(*) FROM users WHERE enabled = true),
			(SELECT COUNT(*) FROM media),
			(SELECT COALESCE(SUM(octet_length(COALESCE(data, ''))), 0) FROM media)
	`

	var stats adminStatistics
	err := utils.DB.QueryRow(r.Context(), query).Scan(&stats.TotalUsers, &stats.ActiveUsers, &stats.TotalMedia, &stats.TotalStorage)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}

	utils.WriteJSONBody(w, utils.JSONResponse{Status: http.StatusOK, Data: stats})
}

type adminUserRow struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	CreatedAt   time.Time `json:"created_at"`
	Enabled     bool      `json:"enabled"`
	Admin       bool      `json:"admin"`
	TotalImages int       `json:"total_images"`
	StorageUsed int64     `json:"storage_usage"`
}

func AdminListUsers(w http.ResponseWriter, r *http.Request) {
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 0 {
		page = 0
	}

	const pageSize = 24
	offset := page * pageSize

	rows, err := utils.DB.Query(
		r.Context(),
		`
			SELECT
				u.id, u.name, u.created_at, u.enabled, u.admin,
				COUNT(m.id) AS total_images,
				COALESCE(SUM(octet_length(m.data)), 0) AS storage_usage
			FROM users u
			LEFT JOIN media m ON m.user_id = u.id
			GROUP BY u.id
			ORDER BY u.created_at DESC
			OFFSET $1
			LIMIT $2;
		`,
		offset, pageSize+1,
	)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}
	defer rows.Close()

	users := make([]adminUserRow, 0, pageSize+1)
	for rows.Next() {
		var u adminUserRow
		if err := rows.Scan(&u.ID, &u.Name, &u.CreatedAt, &u.Enabled, &u.Admin, &u.TotalImages, &u.StorageUsed); err != nil {
			utils.InternalServerError(w, err)
			return
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		utils.InternalServerError(w, err)
		return
	}

	hasNext := len(users) > pageSize
	hasPrev := page > 0

	if hasNext {
		users = users[:pageSize]
	}

	utils.WriteJSONBody(w, utils.JSONResponse{
		Status:   http.StatusOK,
		Data:     users,
		Page:     page,
		PageSize: pageSize,
		HasNext:  hasNext,
		HasPrev:  hasPrev,
	})
}

type setEnabledReceive struct {
	Enabled bool `json:"enabled"`
}

func AdminSetUserEnabled(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, "Invalid user id.")
		return
	}

	if id == utils.GetUserID(r) {
		utils.WriteJSONError(w, http.StatusBadRequest, "You cannot change your own enabled status.")
		return
	}

	payload, err := utils.ReadJSONBody[*setEnabledReceive](w, r.Body, 1<<10)
	if err != nil {
		if errors.Is(err, utils.ErrUnknownJSONFields) {
			utils.WriteJSONError(w, http.StatusBadRequest, "Unknown JSON fields.")
			return
		}
		utils.InternalServerError(w, err)
		return
	}

	user, err := utils.GetUserByID(id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			utils.WriteCodeError(w, http.StatusNotFound)
			return
		}
		utils.InternalServerError(w, err)
		return
	}

	if err := user.SetEnabled(payload.Enabled); err != nil {
		utils.InternalServerError(w, err)
		return
	}

	utils.WriteJSONBody(w, utils.JSONResponse{Status: http.StatusOK, Data: user})
}

type setAdminReceive struct {
	Admin bool `json:"admin"`
}

func AdminSetUserAdmin(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, "Invalid user id.")
		return
	}

	if id == utils.GetUserID(r) {
		utils.WriteJSONError(w, http.StatusBadRequest, "You cannot change your own admin status.")
		return
	}

	payload, err := utils.ReadJSONBody[*setAdminReceive](w, r.Body, 1<<10)
	if err != nil {
		if errors.Is(err, utils.ErrUnknownJSONFields) {
			utils.WriteJSONError(w, http.StatusBadRequest, "Unknown JSON fields.")
			return
		}
		utils.InternalServerError(w, err)
		return
	}

	user, err := utils.GetUserByID(id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			utils.WriteCodeError(w, http.StatusNotFound)
			return
		}
		utils.InternalServerError(w, err)
		return
	}

	if err := user.SetAdmin(payload.Admin); err != nil {
		utils.InternalServerError(w, err)
		return
	}

	utils.WriteJSONBody(w, utils.JSONResponse{Status: http.StatusOK, Data: user})
}
