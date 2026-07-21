package routes

import (
	"errors"
	"img/utils"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type userReceive struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type changePasswordReceive struct {
	Password string `json:"password"`
}

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	payload, err := utils.ReadJSONBody[*userReceive](w, r.Body, 1<<20)
	if err != nil {
		if errors.Is(err, utils.ErrUnknownJSONFields) {
			utils.WriteJSONError(w, http.StatusBadRequest, "Unknown JSON fields.")
			return
		}

		utils.InternalServerError(w, err)
		return
	}

	user, err := utils.CreateUser(payload.Username, payload.Password)
	if err != nil {
		if errors.Is(err, utils.ErrUsernameAlreadyExists) {
			utils.WriteJSONError(w, http.StatusConflict, "Username already exists")
			return
		}

		utils.InternalServerError(w, err)
		return
	}

	cookie := utils.CreateDashSessionCookie(user)
	http.SetCookie(w, &cookie)

	resp := utils.JSONResponse{Status: 200, Data: user}
	utils.WriteJSONBody(w, resp)
}

func ChangePassword(w http.ResponseWriter, r *http.Request) {
	payload, err := utils.ReadJSONBody[*changePasswordReceive](w, r.Body, 1<<20)
	if err != nil {
		if errors.Is(err, utils.ErrUnknownJSONFields) {
			utils.WriteJSONError(w, http.StatusBadRequest, "Unknown JSON fields.")
			return
		}

		utils.InternalServerError(w, err)
		return
	}

	userID := r.Context().Value(utils.CtxUserID)
	user, err := utils.GetUserByID(userID.(int))
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}

	user.SetPassword(payload.Password)

	resp := utils.JSONResponse{Status: 200, Data: "ok"}
	utils.WriteJSONBody(w, resp)
}

func LoginUser(w http.ResponseWriter, r *http.Request) {
	payload, err := utils.ReadJSONBody[*userReceive](w, r.Body, 1<<20)
	if err != nil {
		if errors.Is(err, utils.ErrUnknownJSONFields) {
			utils.WriteJSONError(w, http.StatusBadRequest, "Unknown JSON fields.")
			return
		}

		utils.InternalServerError(w, err)
		return
	}

	user, err := utils.LoginUser(payload.Username, payload.Password)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			utils.WriteCodeError(w, http.StatusUnauthorized)
			return
		}

		utils.InternalServerError(w, err)
		return
	}

	cookie := utils.CreateDashSessionCookie(user)
	http.SetCookie(w, &cookie)

	resp := utils.JSONResponse{Status: 200, Data: user}
	utils.WriteJSONBody(w, resp)
}

type dashboardStatistics struct {
	TotalImages    int `json:"total_images"`
	StorageUsage   int `json:"storage_usage"`
	MonthlyUploads int `json:"monthly_uploads"`
	UserBandwidth  int `json:"user_bandwidth"`
}

func DashboardStatistics(w http.ResponseWriter, r *http.Request) {
	userID := utils.GetUserID(r)
	query := `
		SELECT
			COUNT(*),
			COALESCE(SUM(octet_length(COALESCE(data, ''))), 0),
			COUNT(*) FILTER (
				WHERE date >= date_trunc('month', CURRENT_DATE)
			),
			COALESCE(SUM(COALESCE(octet_length(data), 0) * COALESCE(views, 0)), 0)
		FROM media
		WHERE user_id = $1
	`

	var stats dashboardStatistics
	err := utils.DB.QueryRow(r.Context(), query, userID).Scan(&stats.TotalImages, &stats.StorageUsage, &stats.MonthlyUploads, &stats.UserBandwidth)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}

	p := utils.JSONResponse{Status: http.StatusOK, Data: stats}
	utils.WriteJSONBody(w, p)
}

type dailyStat struct {
	Date    string `json:"date"`
	Uploads int    `json:"uploads"`
	Bytes   int64  `json:"bytes"`
}

type timeseriesResponse struct {
	Days          []dailyStat `json:"days"`
	BaselineBytes int64       `json:"baseline_bytes"`
}

func DashboardTimeseries(w http.ResponseWriter, r *http.Request) {
	userID := utils.GetUserID(r)

	const windowDays = 29

	rows, err := utils.DB.Query(
		r.Context(),
		`
			WITH days AS (
				SELECT generate_series(
					CURRENT_DATE - $2::int,
					CURRENT_DATE,
					'1 day'
				)::date AS day
			),
			daily AS (
				SELECT
					date_trunc('day', date)::date AS day,
					COUNT(*) AS uploads,
					COALESCE(SUM(octet_length(data)), 0) AS bytes
				FROM media
				WHERE user_id = $1 AND date >= CURRENT_DATE - $2::int
				GROUP BY 1
			)
			SELECT d.day, COALESCE(daily.uploads, 0), COALESCE(daily.bytes, 0)
			FROM days d
			LEFT JOIN daily ON daily.day = d.day
			ORDER BY d.day;
		`,
		userID, windowDays,
	)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}
	defer rows.Close()

	days := make([]dailyStat, 0, windowDays+1)
	for rows.Next() {
		var day time.Time
		var stat dailyStat
		if err := rows.Scan(&day, &stat.Uploads, &stat.Bytes); err != nil {
			utils.InternalServerError(w, err)
			return
		}
		stat.Date = day.Format("2006-01-02")
		days = append(days, stat)
	}
	if err := rows.Err(); err != nil {
		utils.InternalServerError(w, err)
		return
	}

	var windowBytes, totalBytes int64
	for _, d := range days {
		windowBytes += d.Bytes
	}
	err = utils.DB.QueryRow(
		r.Context(),
		`SELECT COALESCE(SUM(octet_length(COALESCE(data, ''))), 0) FROM media WHERE user_id = $1`,
		userID,
	).Scan(&totalBytes)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}

	resp := timeseriesResponse{Days: days, BaselineBytes: totalBytes - windowBytes}
	utils.WriteJSONBody(w, utils.JSONResponse{Status: http.StatusOK, Data: resp})
}

func DashboardPing(w http.ResponseWriter, r *http.Request) {
	p := utils.JSONResponse{Status: http.StatusOK, Data: "pong"}
	utils.WriteJSONBody(w, p)
}

func ResetToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := utils.GetUserID(r)

	token, err := utils.CreateToken(userID)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}

	_, err = utils.DB.Exec(ctx, `
		INSERT INTO tokens (user_id, token) 
			VALUES ($1, $2)
		ON CONFLICT (user_id) 
			DO UPDATE SET token = $2
	`, userID, token)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}

	p := utils.JSONResponse{Status: http.StatusOK, Data: token}
	utils.WriteJSONBody(w, p)
}

func LogoutUser(w http.ResponseWriter, r *http.Request) {
	cookie := utils.CreateInvalidDashSessionCookie()
	http.SetCookie(w, &cookie)

	// TODO: perhaps logging?

	resp := utils.JSONResponse{Status: 200, Data: "ok"}
	utils.WriteJSONBody(w, resp)
}

func AdminStatistics(w http.ResponseWriter, r *http.Request) {
	user, err := utils.GetUserByID(utils.GetUserID(r))
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}

	p := utils.JSONResponse{Status: http.StatusOK, Data: user}
	utils.WriteJSONBody(w, p)
}
