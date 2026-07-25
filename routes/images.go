package routes

import (
	"bytes"
	"encoding/json"
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
)

type imageUploadResponse struct {
	URL string `json:"url"`
}

func ImageUpload(w http.ResponseWriter, r *http.Request) error {
	r.ParseMultipartForm(100 << 20)

	file, header, err := r.FormFile("file")
	if err != nil {
		log.Printf("Error while retrieving file: %v\n", err)
		utils.WriteCodeError(w, http.StatusBadRequest)
		return nil
	}
	defer file.Close()

	id, err := utils.GetID(10, false)
	if err != nil {
		return err
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return err
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

	if err := utils.PutObject(r.Context(), utils.MediaBucket, id, data, mtype.String()); err != nil {
		return err
	}
	if coverArt != nil {
		coverType := mimetype.Detect(coverArt).String()
		if err := utils.PutObject(r.Context(), utils.CoversBucket, id, coverArt, coverType); err != nil {
			return err
		}
	}

	_, err = utils.DB.Exec(
		r.Context(),
		`
			INSERT INTO media
				(id, size, mimetype, user_id, filename, width, height, duration_ms, bitrate, codec, framerate, sample_rate, channels, has_cover)
			VALUES
				($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		`,
		id, len(data), mtype.String(), userID, filename, width, height, durationMs, bitrate, codec, framerate, sampleRate, channels, coverArt != nil,
	)
	if err != nil {
		return err
	}

	URL := fmt.Sprintf(`%s/%s%s`,
		utils.BaseURL,
		id,
		mtype.Extension(),
	)

	respJSON, err := json.Marshal(&imageUploadResponse{URL})
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(respJSON)
	return nil
}

// /{id} --redirect--> /{user}/{id}

func ImageRedirect(w http.ResponseWriter, r *http.Request) error {
	filename := r.PathValue("id")

	// EeZDFWheuD.png
	imageID := strings.Split(filename, ".")[0]

	var userID int
	err := utils.DB.QueryRow(r.Context(), "SELECT user_id FROM media WHERE id = $1", imageID).Scan(&userID)
	if err != nil {
		utils.WriteCodeError(w, http.StatusNotFound)
		return nil
	}
	user, err := utils.GetUserByID(userID)
	if err != nil {
		return err
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
	return nil
}

func ImageGet(w http.ResponseWriter, r *http.Request) error {
	user := r.PathValue("user")
	id := r.PathValue("id")
	split := strings.Split(id, ".")

	// this route is unauthenticated and public,
	// so allowing cross-origin reads (needed for dash text
	// preview, which fetches the body via JS) doesn't expose anything new
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// sandbox to prevent potential malice
	w.Header().Set("Content-Security-Policy", "sandbox allow-scripts")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	isDashboard := (r.URL.Query().Get("d") == "true")
	wantCover := (r.URL.Query().Get("cover") == "true")
	wantDownload := (r.URL.Query().Get("download") == "true")

	if wantCover {
		var hasCover bool
		err := utils.DB.QueryRow(
			r.Context(),
			`
			SELECT m.has_cover
			FROM media m
			JOIN users ON m.user_id = users.id
				WHERE m.id = $1 AND users.name = $2;
			`,
			split[0], user,
		).Scan(&hasCover)
		if err != nil || !hasCover {
			utils.WriteCodeError(w, http.StatusNotFound)
			return nil
		}

		coverArt, err := utils.GetObject(r.Context(), utils.CoversBucket, split[0])
		if err != nil {
			if utils.IsNotFound(err) {
				utils.WriteCodeError(w, http.StatusNotFound)
				return nil
			}
			return err
		}

		width := r.URL.Query().Get("width")
		if width == "" {
			width = "512"
		}

		resized, contentType, err := utils.ResizeAndEncode(coverArt, width)
		if err != nil {
			return err
		}

		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Cache-Control", "public, max-age=86400")
		http.ServeContent(w, r, "", time.Time{}, bytes.NewReader(resized))
		return nil
	}

	var mimeType string
	var filename *string
	err := utils.DB.QueryRow(
		r.Context(),
		`
		WITH updated AS (
			UPDATE media
			SET views = views + CASE WHEN $3 = false THEN 1 ELSE 0 END
			WHERE id = $1
			RETURNING mimetype, filename, user_id
		)
		SELECT
			u.mimetype,
			u.filename
		FROM updated u
		JOIN users ON u.user_id = users.id
			WHERE users.name = $2;
		`,
		split[0], user, isDashboard,
	).Scan(&mimeType, &filename)
	if err != nil {
		return utils.NotFoundIfNoRows(err, http.StatusText(http.StatusNotFound))
	}

	imageData, err := utils.GetObject(r.Context(), utils.MediaBucket, split[0])
	if err != nil {
		if utils.IsNotFound(err) {
			utils.WriteCodeError(w, http.StatusNotFound)
			return nil
		}
		return err
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
			return err
		}

		w.Header().Set("Content-Type", contentType)
		imageData = resized
	} else {
		w.Header().Set("Content-Type", mimeType)
	}

	w.Header().Set("Cache-Control", "public, max-age=3600")
	http.ServeContent(w, r, "", time.Time{}, bytes.NewReader(imageData))
	return nil
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

var imageSortColumns = map[string]string{
	"date_desc":  "date DESC",
	"date_asc":   "date ASC",
	"size_desc":  "size DESC",
	"size_asc":   "size ASC",
	"views_desc": "views DESC",
	"views_asc":  "views ASC",
	"name_asc":   "COALESCE(filename, id) ASC",
	"name_desc":  "COALESCE(filename, id) DESC",
}

func GetImages(w http.ResponseWriter, r *http.Request) error {
	userID := utils.GetUserID(r)

	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 0 {
		page = 0
	}

	orderBy, ok := imageSortColumns[r.URL.Query().Get("sort")]
	if !ok {
		orderBy = imageSortColumns["date_desc"]
	}

	const pageSize = 24
	offset := page * pageSize

	rows, err := utils.DB.Query(
		r.Context(),
		fmt.Sprintf(
			`
				SELECT
					id, date, mimetype, views, filename, size,
					width, height, duration_ms, bitrate, codec, framerate, sample_rate, channels,
					has_cover
				FROM media
				WHERE user_id = $1
				ORDER BY %s
				OFFSET ($2)
				LIMIT $3;
			`,
			orderBy,
		),
		userID, offset, pageSize+1,
	)
	if err != nil {
		return err
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
			return err
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
		return err
	}

	hasNext := len(images) > pageSize
	hasPrev := page > 0

	if hasNext {
		images = images[:pageSize]
	}

	utils.WritePaginatedJSONBody(w, utils.PaginatedJSONResponse{
		JSONResponse: utils.JSONResponse{Status: http.StatusOK, Data: images},
		Page:         page,
		PageSize:     pageSize,
		HasNext:      hasNext,
		HasPrev:      hasPrev,
	})
	return nil
}

func DeleteImage(w http.ResponseWriter, r *http.Request) error {
	userID := utils.GetUserID(r)
	imageID := r.URL.Query().Get("id")

	var deleted, hasCover bool
	err := utils.DB.QueryRow(
		r.Context(),
		`
			WITH removed AS (
				DELETE FROM media
				WHERE id = $1 AND user_id = $2
				RETURNING COALESCE(size, 0)::bigint * COALESCE(views, 0) AS bandwidth, has_cover
			),
			bumped AS (
				UPDATE users
				SET bandwidth = bandwidth + COALESCE((SELECT SUM(bandwidth) FROM removed), 0)
				WHERE id = $2 AND EXISTS (SELECT 1 FROM removed)
			)
			SELECT EXISTS (SELECT 1 FROM removed), COALESCE((SELECT has_cover FROM removed), false)
		`,
		imageID, userID,
	).Scan(&deleted, &hasCover)
	if err != nil {
		return err
	}

	if deleted {
		if err := utils.DeleteObject(r.Context(), utils.MediaBucket, imageID); err != nil {
			log.Printf("Error deleting media object %q: %v\n", imageID, err)
		}
		if hasCover {
			if err := utils.DeleteObject(r.Context(), utils.CoversBucket, imageID); err != nil {
				log.Printf("Error deleting cover object %q: %v\n", imageID, err)
			}
		}

		utils.WriteJSONBody(w, utils.JSONResponse{Status: http.StatusOK})
		return nil
	}
	utils.WriteJSONBody(w, utils.JSONResponse{Status: http.StatusNotFound})
	return nil
}
