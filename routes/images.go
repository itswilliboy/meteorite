package routes

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
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

	file, header, err := r.FormFile("file")
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

	mtype := mimetype.Detect(data)
	category := strings.Split(mtype.String(), "/")[0]
	userID := utils.GetUserID(r)

	var filename *string
	if header != nil && header.Filename != "" {
		name := header.Filename
		filename = &name
	}

	var width, height, durationMs, bitrate, sampleRate, channels *int
	var codec *string
	var framerate *float64
	var coverArt []byte

	switch category {
	case "image":
		if cfg, _, err := image.DecodeConfig(bytes.NewReader(data)); err == nil {
			w, h := cfg.Width, cfg.Height
			width, height = &w, &h
		}
	case "video", "audio":
		if meta, err := utils.ProbeMedia(data, mtype.Extension()); err == nil {
			if meta.Width > 0 {
				width = &meta.Width
			}
			if meta.Height > 0 {
				height = &meta.Height
			}
			if meta.DurationMs > 0 {
				durationMs = &meta.DurationMs
			}
			if meta.Bitrate > 0 {
				bitrate = &meta.Bitrate
			}
			if meta.Codec != "" {
				codec = &meta.Codec
			}
			if meta.Framerate > 0 {
				framerate = &meta.Framerate
			}
			if meta.SampleRate > 0 {
				sampleRate = &meta.SampleRate
			}
			if meta.Channels > 0 {
				channels = &meta.Channels
			}
		} else {
			log.Printf("Error probing media: %v\n", err)
		}

		if category == "audio" {
			if art, err := utils.ExtractCoverArt(data, mtype.Extension()); err == nil {
				coverArt = art
			}
		}
	}

	_, err = utils.DB.Exec(
		r.Context(),
		`
			INSERT INTO media
				(id, data, mimetype, user_id, filename, width, height, duration_ms, bitrate, codec, framerate, sample_rate, channels, cover_art)
			VALUES
				($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		`,
		id, data, mtype.String(), userID, filename, width, height, durationMs, bitrate, codec, framerate, sampleRate, channels, coverArt,
	)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}

	URL := fmt.Sprintf(`%s/%s%s`,
		utils.BaseURL,
		id,
		mtype.Extension(),
	)

	respJSON, err := json.Marshal(&imageUploadResponse{URL})
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(respJSON)
}

// /{id} --redirect--> /{user}/{id}

func ImageRedirect(w http.ResponseWriter, r *http.Request) {
	filename := r.PathValue("id")

	// EeZDFWheuD.png
	imageID := strings.Split(filename, ".")[0]

	var userID int
	err := utils.DB.QueryRow(r.Context(), "SELECT user_id FROM media WHERE id = $1", imageID).Scan(&userID)
	if err != nil {
		utils.WriteCodeError(w, http.StatusNotFound)
		return
	}
	user, err := utils.GetUserByID(userID)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}

	redirectURL := fmt.Sprintf("/%s/%s", user.Name, filename)
	queryParams := r.URL.Query()

	if len(queryParams) > 0 {
		redirectURL += "?" + queryParams.Encode()
	}

	// cross-origin fetch() (eg. the dashboard's text preview) needs this on
	// the redirect response itself, not just the final destination
	w.Header().Set("Access-Control-Allow-Origin", "*")
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

func ImageGet(w http.ResponseWriter, r *http.Request) {
	user := r.PathValue("user")
	id := r.PathValue("id")
	split := strings.Split(id, ".")

	// this route is unauthenticated and public,
	// so allowing cross-origin reads (needed for dash text
	// preview, which fetches the body via JS) doesn't expose anything new
	w.Header().Set("Access-Control-Allow-Origin", "*")

	isDashboard := (r.URL.Query().Get("d") == "true")
	wantCover := (r.URL.Query().Get("cover") == "true")
	wantDownload := (r.URL.Query().Get("download") == "true")

	if wantCover {
		var coverArt []byte
		err := utils.DB.QueryRow(
			r.Context(),
			`
			SELECT m.cover_art
			FROM media m
			JOIN users ON m.user_id = users.id
				WHERE m.id = $1 AND users.name = $2;
			`,
			split[0], user,
		).Scan(&coverArt)
		if err != nil || len(coverArt) == 0 {
			utils.WriteCodeError(w, http.StatusNotFound)
			return
		}

		width := r.URL.Query().Get("width")
		if width == "" {
			width = "512"
		}

		resized, contentType, err := utils.ResizeAndEncode(coverArt, width)
		if err != nil {
			utils.InternalServerError(w, err)
			return
		}

		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Cache-Control", "public, max-age=86400")
		http.ServeContent(w, r, "", time.Time{}, bytes.NewReader(resized))
		return
	}

	var imageData []byte
	var mimeType string
	var filename *string
	err := utils.DB.QueryRow(
		r.Context(),
		`
		WITH updated AS (
			UPDATE media
			SET views = views + CASE WHEN $3 = false THEN 1 ELSE 0 END
			WHERE id = $1
			RETURNING data, mimetype, filename, user_id
		)
		SELECT
			u.data,
			u.mimetype,
			u.filename
		FROM updated u
		JOIN users ON u.user_id = users.id
			WHERE users.name = $2;
		`,
		split[0], user, isDashboard,
	).Scan(&imageData, &mimeType, &filename)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			utils.WriteCodeError(w, http.StatusNotFound)
			return
		}

		utils.InternalServerError(w, err)
		return
	}

	if wantDownload {
		name := id
		if filename != nil && *filename != "" {
			name = *filename
		}
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", name))
	}

	width := r.URL.Query().Get("width")

	if width != "" {
		resized, contentType, err := utils.ResizeAndEncode(imageData, width)
		if err != nil {
			utils.InternalServerError(w, err)
			return
		}

		w.Header().Set("Content-Type", contentType)
		imageData = resized
	} else {
		w.Header().Set("Content-Type", mimeType)
	}

	w.Header().Set("Cache-Control", "public, max-age=3600")
	http.ServeContent(w, r, "", time.Time{}, bytes.NewReader(imageData))
}

type GetImagesResp struct {
	ID         string   `json:"id"`
	Date       any      `json:"date"`
	Mimetype   string   `json:"mimetype"`
	Views      int      `json:"views"`
	URL        string   `json:"url"`
	Filename   *string  `json:"filename"`
	Size       int64    `json:"size"`
	Width      *int     `json:"width"`
	Height     *int     `json:"height"`
	DurationMs *int     `json:"duration_ms"`
	Bitrate    *int     `json:"bitrate"`
	Codec      *string  `json:"codec"`
	Framerate  *float64 `json:"framerate"`
	SampleRate *int     `json:"sample_rate"`
	Channels   *int     `json:"channels"`
	HasCover   bool     `json:"has_cover"`
}

func GetImages(w http.ResponseWriter, r *http.Request) {
	userID := utils.GetUserID(r)

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
				id, date, mimetype, views, filename, octet_length(data) AS size,
				width, height, duration_ms, bitrate, codec, framerate, sample_rate, channels,
				(cover_art IS NOT NULL) AS has_cover
			FROM media
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

		if err := rows.Scan(
			&img.ID, &date, &img.Mimetype, &img.Views, &img.Filename, &img.Size,
			&img.Width, &img.Height, &img.DurationMs, &img.Bitrate, &img.Codec, &img.Framerate, &img.SampleRate, &img.Channels,
			&img.HasCover,
		); err != nil {
			utils.InternalServerError(w, err)
			return
		}

		img.Date = date

		extension := ""
		if m := mimetype.Lookup(img.Mimetype); m != nil {
			extension = m.Extension()
		}
		img.URL = fmt.Sprintf("%s/%s%s", utils.BaseURL, img.ID, extension)
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
		r.Context(),
		`
			DELETE FROM media
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
