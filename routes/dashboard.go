package routes

import (
	"context"
	"errors"
	"img/utils"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5"
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

	token, err := utils.CreateUser(payload.Username, payload.Password)
	if err != nil {
		if errors.Is(err, utils.ErrUsernameAlreadyExists) {
			utils.WriteJSONError(w, http.StatusConflict, "Username already exists")
			return
		}

		log.Println("error creating user")
		utils.InternalServerError(w, err)
		return
	}

	resp := utils.JSONResponse{Status: 200, Data: token}
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

		utils.InternalServerError(w, err)
		return
	}

	token, _ := utils.GetToken(user)

	resp := utils.JSONResponse{Status: 200, Data: token}
	utils.WriteJSONBody(w, resp)
}

type dashboardStatistics struct {
	TotalImages    int `json:"total_images"`
	StorageUsage   int `json:"storage_usage"`
	MonthlyUploads int `json:"monthly_uploads"`
	UserBandwidth  int `json:"user_bandwidth"`
}

func DashboardStatistics(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT
			COUNT(*),
			SUM(octet_length(image_data))
		FROM images i
	`
	var stats dashboardStatistics
	err := utils.DB.QueryRow(context.Background(), query).Scan(&stats.TotalImages, &stats.StorageUsage)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}

	p := utils.JSONResponse{Status: http.StatusOK, Data: stats}
	utils.WriteJSONBody(w, p)
}
