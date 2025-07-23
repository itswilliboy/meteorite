package routes

import (
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
