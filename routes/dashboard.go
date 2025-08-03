package routes

import (
	"context"
	"errors"
	"img/utils"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type userReceive struct {
	Username string `json:"username"`
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
		if errors.Is(err, pgx.ErrNoRows) {
			utils.WriteCodeError(w, http.StatusNotFound)
			return
		}

		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
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
	userID := r.Context().Value(utils.CtxUserID)
	query := `
		SELECT
			COUNT(*),
			COALESCE(SUM(octet_length(COALESCE(image_data, ''))), 0),
			COUNT(*) FILTER (
				WHERE date >= date_trunc('month', CURRENT_DATE)
			),
			COALESCE(SUM(COALESCE(octet_length(image_data), 0) * COALESCE(views, 0)), 0)
		FROM images
		WHERE user_id = $1
	`

	var stats dashboardStatistics
	err := utils.DB.QueryRow(context.Background(), query, userID).Scan(&stats.TotalImages, &stats.StorageUsage, &stats.MonthlyUploads, &stats.UserBandwidth)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}

	p := utils.JSONResponse{Status: http.StatusOK, Data: stats}
	utils.WriteJSONBody(w, p)
}

func DashboardPing(w http.ResponseWriter, r *http.Request) {
	p := utils.JSONResponse{Status: http.StatusOK, Data: "pong"}
	utils.WriteJSONBody(w, p)
}

func ResetToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, err := strconv.Atoi(ctx.Value(utils.CtxUserID).(string))
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}

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
	ctx := r.Context()
	userID, err := strconv.Atoi(ctx.Value(utils.CtxUserID).(string))
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}

	user, err := utils.GetUserByID(userID)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}

	p := utils.JSONResponse{Status: http.StatusOK, Data: user}
	utils.WriteJSONBody(w, p)
}
