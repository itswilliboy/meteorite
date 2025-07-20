package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"img/utils"

	"github.com/gabriel-vasile/mimetype"
)

type imageUploadResponse struct {
	Url string `json:"url"`
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
	userId, _ := r.Context().Value("userId").(int)
	_, err = utils.DB.Exec(context.Background(), "INSERT INTO images (id, image_data, mimetype, user_id) VALUES ($1, $2, $3, $4)", id, data, mimetype.String(), userId)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}

	Url := fmt.Sprintf(`%s/%s%s`,
		utils.BASE_URL,
		id,
		mimetype.Extension(),
	)

	json, _ := json.Marshal(&imageUploadResponse{Url})
	w.Write(json)
}

// /{id} --redirect--> /{user}/{id}
func Test(w http.ResponseWriter, r *http.Request) {
	filename := r.PathValue("id")

	// EeZDFWheuD.png
	imageId := strings.Split(filename, ".")[0]

	var userId int
	err := utils.DB.QueryRow(context.Background(), "SELECT user_id FROM images WHERE id = $1", imageId).Scan(&userId)
	if err != nil {
		utils.WriteCodeError(w, http.StatusNotFound)
		return
	}
	user := utils.GetUserByID(userId)

	http.Redirect(w, r, fmt.Sprintf("/%s/%s", user.Name, filename), http.StatusTemporaryRedirect)

}

func ImageGet(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	split := strings.Split(id, ".")

	var imageData []byte
	var mimetype string
	err := utils.DB.QueryRow(context.Background(), "SELECT image_data, mimetype FROM images WHERE id = $1", split[0]).Scan(&imageData, &mimetype)
	if err != nil {
		utils.WriteCodeError(w, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", mimetype)
	w.Write(imageData)
}
