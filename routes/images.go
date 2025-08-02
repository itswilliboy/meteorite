package routes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"img/utils"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"github.com/jackc/pgx/v5"
)

type imageUploadResponse struct {
	URL string `json:"url"`
}

func ImageUpload(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(100 << 20)

	file, _, err := r.FormFile("file")
	if err != nil {
		log.Printf("Error while retrieving file: %v\n", err)
		utils.WriteCodeError(w, http.StatusBadRequest)
		return
	}
	defer file.Close()

	id, err := utils.GetID(10, false)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}

	data, err := io.ReadAll(file)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}

	mimetype := mimetype.Detect(data)
	userID, _ := r.Context().Value(utils.CtxUserID).(int)

	_, err = utils.DB.Exec(context.Background(), "INSERT INTO images (id, image_data, mimetype, user_id) VALUES ($1, $2, $3, $4)", id, data, mimetype.String(), userID)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}

	URL := fmt.Sprintf(`%s/%s%s`,
		utils.BASE_URL,
		id,
		mimetype.Extension(),
	)

	json, _ := json.Marshal(&imageUploadResponse{URL})
	w.Write(json)
}

// /{id} --redirect--> /{user}/{id}

func ImageRedirect(w http.ResponseWriter, r *http.Request) {
	filename := r.PathValue("id")

	// EeZDFWheuD.png
	imageID := strings.Split(filename, ".")[0]

	var userID int
	err := utils.DB.QueryRow(context.Background(), "SELECT user_id FROM images WHERE id = $1", imageID).Scan(&userID)
	if err != nil {
		utils.WriteCodeError(w, http.StatusNotFound)
		return
	}
	user, err := utils.GetUserByID(userID)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/%s/%s", user.Name, filename), http.StatusTemporaryRedirect)
}

func ImageGet(w http.ResponseWriter, r *http.Request) {
	user := r.PathValue("user")
	id := r.PathValue("id")
	split := strings.Split(id, ".")

	var imageData []byte
	var mimetype string
	err := utils.DB.QueryRow(
		context.Background(),
		`
		with updated AS (
			UPDATE images
			SET views = views + 1
			WHERE id = $1
			RETURNING image_data, mimetype, user_id
		)
		SELECT 
			u.image_data,
			u.mimetype
		FROM updated u 
		JOIN users ON u.user_id = users.id 
			WHERE users.name = $2;
		`,
		split[0], user,
	).Scan(&imageData, &mimetype)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			utils.WriteCodeError(w, http.StatusNotFound)
			return
		}

		utils.InternalServerError(w, err)
		return
	}

	w.Header().Set("Content-Type", mimetype)
	w.Write(imageData)
}
