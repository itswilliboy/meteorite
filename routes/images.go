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

type mediaMetadata struct {
	Width, Height, DurationMs, Bitrate, SampleRate, Channels *int
	Codec                                                    *string
	Framerate                                                *float64
	CoverArt                                                 []byte
}

func extractMediaMetadata(data []byte, category string, extension string) mediaMetadata {
	var meta mediaMetadata

	switch category {
	case "image":
		if cfg, _, err := image.DecodeConfig(bytes.NewReader(data)); err == nil {
			w, h := cfg.Width, cfg.Height
			meta.Width, meta.Height = &w, &h
		}
	case "video", "audio":
		if m, err := utils.ProbeMedia(data, extension); err == nil {
			if m.Width > 0 {
				meta.Width = &m.Width
			}
			if m.Height > 0 {
				meta.Height = &m.Height
			}
			if m.DurationMs > 0 {
				meta.DurationMs = &m.DurationMs
			}
			if m.Bitrate > 0 {
				meta.Bitrate = &m.Bitrate
			}
			if m.Codec != "" {
				meta.Codec = &m.Codec
			}
			if m.Framerate > 0 {
				meta.Framerate = &m.Framerate
			}
			if m.SampleRate > 0 {
				meta.SampleRate = &m.SampleRate
			}
			if m.Channels > 0 {
				meta.Channels = &m.Channels
			}
		} else {
			log.Printf("Error probing media: %v\n", err)
		}

		if category == "audio" {
			if art, err := utils.ExtractCoverArt(data, extension); err == nil {
				meta.CoverArt = art
			}
		}
	}

	return meta
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
	category, _, _ := strings.Cut(mtype.String(), "/")
	userID := utils.GetUserID(r)

	var folderID *string
	if fid := r.FormValue("folder_id"); fid != "" {
		exists, err := utils.FolderExists(r.Context(), fid, userID)
		if err != nil {
			return err
		}
		if !exists {
			return utils.NewHTTPError(http.StatusNotFound, "Folder not found.")
		}
		folderID = &fid
	}

	var filename *string
	if header != nil && header.Filename != "" {
		name := header.Filename
		filename = &name
	}

	meta := extractMediaMetadata(data, category, mtype.Extension())

	if err := utils.PutObject(r.Context(), utils.MediaBucket, id, data, mtype.String()); err != nil {
		return err
	}
	if meta.CoverArt != nil {
		coverType := mimetype.Detect(meta.CoverArt).String()
		if err := utils.PutObject(r.Context(), utils.CoversBucket, id, meta.CoverArt, coverType); err != nil {
			return err
		}
	}

	_, err = utils.DB.Exec(
		r.Context(),
		`
			INSERT INTO media
				(id, size, mimetype, user_id, filename, width, height, duration_ms, bitrate, codec, framerate, sample_rate, channels, has_cover, folder_id)
			VALUES
				($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		`,
		id, len(data), mtype.String(), userID, filename,
		meta.Width, meta.Height, meta.DurationMs, meta.Bitrate, meta.Codec, meta.Framerate, meta.SampleRate, meta.Channels,
		meta.CoverArt != nil, folderID,
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

type renameImageReceive struct {
	Filename string `json:"filename"`
}

type renameImageResp struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
}

func RenameImage(w http.ResponseWriter, r *http.Request) error {
	userID := utils.GetUserID(r)
	imageID := r.PathValue("id")

	payload, err := utils.ReadJSONBody[*renameImageReceive](w, r.Body, 1<<10)
	if err != nil {
		return err
	}

	filename := strings.TrimSpace(payload.Filename)
	if filename == "" {
		return utils.NewHTTPError(http.StatusBadRequest, "Filename cannot be empty.")
	}

	var resp renameImageResp
	err = utils.DB.QueryRow(
		r.Context(),
		`
			UPDATE media SET filename = $1
			WHERE id = $2 AND user_id = $3
			RETURNING id, filename
		`,
		filename, imageID, userID,
	).Scan(&resp.ID, &resp.Filename)
	if err != nil {
		return utils.NotFoundIfNoRows(err, "Image not found.")
	}

	utils.WriteJSONBody(w, utils.JSONResponse{Status: http.StatusOK, Data: resp})
	return nil
}

// /{id} --redirect--> /{user}/{id}

func ImageRedirect(w http.ResponseWriter, r *http.Request) error {
	filename := r.PathValue("id")

	// EeZDFWheuD.png
	imageID, _, _ := strings.Cut(filename, ".")

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
	mediaID, _, _ := strings.Cut(id, ".")

	// this route is unauthenticated and public,
	// so allowing cross-origin reads (needed for dash text
	// preview, which fetches the body via JS) doesn't expose anything new
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// sandbox to prevent potential malice
	w.Header().Set("Content-Security-Policy", "sandbox allow-scripts")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	if r.URL.Query().Get("cover") == "true" {
		return serveCoverArt(w, r, user, mediaID)
	}
	return serveMedia(w, r, user, id, mediaID)
}

func serveCoverArt(w http.ResponseWriter, r *http.Request, user, mediaID string) error {
	var hasCover bool
	err := utils.DB.QueryRow(
		r.Context(),
		`
		SELECT m.has_cover
		FROM media m
		JOIN users ON m.user_id = users.id
			WHERE m.id = $1 AND users.name = $2;
		`,
		mediaID, user,
	).Scan(&hasCover)
	if err != nil || !hasCover {
		utils.WriteCodeError(w, http.StatusNotFound)
		return nil
	}

	coverArt, err := utils.GetObject(r.Context(), utils.CoversBucket, mediaID)
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

func serveMedia(w http.ResponseWriter, r *http.Request, user, id, mediaID string) error {
	isDashboard := (r.URL.Query().Get("d") == "true")
	wantDownload := (r.URL.Query().Get("download") == "true")

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
		mediaID, user, isDashboard,
	).Scan(&mimeType, &filename)
	if err != nil {
		return utils.NotFoundIfNoRows(err, http.StatusText(http.StatusNotFound))
	}

	imageData, err := utils.GetObject(r.Context(), utils.MediaBucket, mediaID)
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

	folderID := r.URL.Query().Get("folder_id")
	var folderIDArg *string
	if folderID != "" {
		folderIDArg = &folderID
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
				WHERE user_id = $1 AND folder_id IS NOT DISTINCT FROM $4
				ORDER BY %s, id DESC
				OFFSET ($2)
				LIMIT $3;
			`,
			orderBy,
		),
		userID, offset, pageSize+1, folderIDArg,
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
		utils.DeleteMediaObjects(r.Context(), imageID, hasCover)

		utils.WriteJSONBody(w, utils.JSONResponse{Status: http.StatusOK})
		return nil
	}
	utils.WriteJSONBody(w, utils.JSONResponse{Status: http.StatusNotFound})
	return nil
}

type bulkDeleteReceive struct {
	IDs []string `json:"ids"`
}

func BulkDeleteImages(w http.ResponseWriter, r *http.Request) error {
	userID := utils.GetUserID(r)

	payload, err := utils.ReadJSONBody[*bulkDeleteReceive](w, r.Body, 1<<16)
	if err != nil {
		return err
	}

	if len(payload.IDs) == 0 {
		utils.WriteJSONError(w, http.StatusBadRequest, "No ids provided.")
		return nil
	}

	rows, err := utils.DB.Query(
		r.Context(),
		`
			WITH removed AS (
				DELETE FROM media
				WHERE id = ANY($1) AND user_id = $2
				RETURNING id, COALESCE(size, 0)::bigint * COALESCE(views, 0) AS bandwidth, has_cover
			),
			bumped AS (
				UPDATE users
				SET bandwidth = bandwidth + COALESCE((SELECT SUM(bandwidth) FROM removed), 0)
				WHERE id = $2 AND EXISTS (SELECT 1 FROM removed)
			)
			SELECT id, has_cover FROM removed
		`,
		payload.IDs, userID,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	type removedMedia struct {
		id       string
		hasCover bool
	}

	removed := make([]removedMedia, 0, len(payload.IDs))
	for rows.Next() {
		var rm removedMedia
		if err := rows.Scan(&rm.id, &rm.hasCover); err != nil {
			return err
		}
		removed = append(removed, rm)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	deletedIDs := make([]string, 0, len(removed))
	for _, rm := range removed {
		deletedIDs = append(deletedIDs, rm.id)
		utils.DeleteMediaObjects(r.Context(), rm.id, rm.hasCover)
	}

	utils.WriteJSONBody(w, utils.JSONResponse{Status: http.StatusOK, Data: deletedIDs})
	return nil
}

type bulkMoveReceive struct {
	IDs      []string `json:"ids"`
	FolderID *string  `json:"folder_id"`
}

func BulkMoveImages(w http.ResponseWriter, r *http.Request) error {
	userID := utils.GetUserID(r)

	payload, err := utils.ReadJSONBody[*bulkMoveReceive](w, r.Body, 1<<16)
	if err != nil {
		return err
	}

	if len(payload.IDs) == 0 {
		utils.WriteJSONError(w, http.StatusBadRequest, "No ids provided.")
		return nil
	}

	if payload.FolderID != nil {
		exists, err := utils.FolderExists(r.Context(), *payload.FolderID, userID)
		if err != nil {
			return err
		}
		if !exists {
			return utils.NewHTTPError(http.StatusNotFound, "Folder not found.")
		}
	}

	rows, err := utils.DB.Query(
		r.Context(),
		`
			UPDATE media
			SET folder_id = $1
			WHERE id = ANY($2) AND user_id = $3
			RETURNING id
		`,
		payload.FolderID, payload.IDs, userID,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	movedIDs := make([]string, 0, len(payload.IDs))
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return err
		}
		movedIDs = append(movedIDs, id)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	utils.WriteJSONBody(w, utils.JSONResponse{Status: http.StatusOK, Data: movedIDs})
	return nil
}
