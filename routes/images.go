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
	"strconv"
	"strings"
	"time"

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
		utils.BaseURL,
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
	isDashboard := (r.URL.Query().Get("d") == "true")

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

	format := "/%s/%s"
	if isDashboard {
		format += "?d=true"
	}
	http.Redirect(w, r, fmt.Sprintf(format, user.Name, filename), http.StatusTemporaryRedirect)
}

func ImageGet(w http.ResponseWriter, r *http.Request) {
	user := r.PathValue("user")
	id := r.PathValue("id")
	split := strings.Split(id, ".")

	isDashboard := (r.URL.Query().Get("d") == "true")

	var imageData []byte
	var mimetype string
	err := utils.DB.QueryRow(
		context.Background(),
		`
		WITH updated AS (
			UPDATE images
			SET views = views + CASE WHEN $3 = false THEN 1 ELSE 0 END
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
		split[0], user, isDashboard,
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
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.Write(imageData)
}

type GetImagesResp struct {
	ID       string `json:"id"`
	Date     any    `json:"date"`
	Mimetype string `json:"mimetype"`
	Views    int    `json:"views"`
	URL      string `json:"url"`
}

func GetImages(w http.ResponseWriter, r *http.Request) {
	userID, _ := strconv.Atoi(r.Context().Value(utils.CtxUserID).(string))

	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		page = 0
	}

	const pageSize = 24
	offset := page * pageSize

	rows, err := utils.DB.Query(
		context.Background(),
		`
			SELECT id, date, mimetype, views 
			FROM images
			WHERE user_id = $1
			ORDER BY date DESC
			OFFSET ($2)
			LIMIT $3;
		`,
		userID, offset, pageSize+1,
	)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}
	defer rows.Close()

	images := make([]GetImagesResp, 0, pageSize+1)
	for rows.Next() {
		var img GetImagesResp
		var date time.Time

		if err := rows.Scan(&img.ID, &date, &img.Mimetype, &img.Views); err != nil {
			utils.InternalServerError(w, err)
			return
		}

		img.Date = date
		img.URL = fmt.Sprintf("%s/%s%s", utils.BaseURL, img.ID, mimetype.Lookup(img.Mimetype).Extension())
		images = append(images, img)
	}

	if err := rows.Err(); err != nil {
		utils.InternalServerError(w, err)
		return
	}

	hasNext := len(images) > pageSize
	hasPrev := page > 0

	if hasNext {
		images = images[:pageSize]
	}

	// utils.WriteJSONBody(w, utils.JSONResponse{Status: http.StatusOK, Data: })
	utils.WriteJSONBody(w, utils.JSONResponse{
		Status:   http.StatusOK,
		Data:     images,
		Page:     page,
		PageSize: pageSize,
		HasNext:  hasNext,
		HasPrev:  hasPrev,
	})
}

func DeleteImage(w http.ResponseWriter, r *http.Request) {
	userID := utils.GetUserID(r)
	imageID := r.URL.Query().Get("id")

	tag, err := utils.DB.Exec(
		context.Background(),
		`
			DELETE FROM images
			WHERE id = $1
			AND user_id = $2
		`,
		imageID, userID,
	)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}
	if tag.RowsAffected() == 1 {
		utils.WriteJSONBody(w, utils.JSONResponse{Status: http.StatusOK})
		return
	}
	utils.WriteJSONBody(w, utils.JSONResponse{Status: http.StatusNotFound})
}
