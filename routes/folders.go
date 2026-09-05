package routes

import (
	"img/utils"
	"net/http"
	"slices"
	"strings"
	"time"
)

type Folder struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	ParentID  *string   `json:"parent_id"`
	CreatedAt time.Time `json:"created_at"`
}

type BreadcrumbEntry struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type getFoldersResp struct {
	Folders    []Folder          `json:"folders"`
	Breadcrumb []BreadcrumbEntry `json:"breadcrumb"`
}

func GetFolders(w http.ResponseWriter, r *http.Request) error {
	userID := utils.GetUserID(r)
	parentID := r.URL.Query().Get("parent_id")

	var parentIDArg *string
	if parentID != "" {
		parentIDArg = &parentID
	}

	rows, err := utils.DB.Query(
		r.Context(),
		`
			SELECT id, name, parent_id, created_at FROM folders
			WHERE user_id = $1 AND parent_id IS NOT DISTINCT FROM $2
			ORDER BY name ASC
		`,
		userID, parentIDArg,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	folders := make([]Folder, 0)
	for rows.Next() {
		var f Folder
		if err := rows.Scan(&f.ID, &f.Name, &f.ParentID, &f.CreatedAt); err != nil {
			return err
		}
		folders = append(folders, f)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	breadcrumb := make([]BreadcrumbEntry, 0)
	if parentID != "" {
		bcRows, err := utils.DB.Query(
			r.Context(),
			`
				WITH RECURSIVE path AS (
					SELECT id, name, parent_id FROM folders WHERE id = $1 AND user_id = $2
					UNION ALL
					SELECT f.id, f.name, f.parent_id FROM folders f JOIN path p ON f.id = p.parent_id WHERE f.user_id = $2
				)
				SELECT id, name FROM path
			`,
			parentID, userID,
		)
		if err != nil {
			return err
		}
		defer bcRows.Close()

		reversed := make([]BreadcrumbEntry, 0)
		for bcRows.Next() {
			var e BreadcrumbEntry
			if err := bcRows.Scan(&e.ID, &e.Name); err != nil {
				return err
			}
			reversed = append(reversed, e)
		}
		if err := bcRows.Err(); err != nil {
			return err
		}

		for _, r := range slices.Backward(reversed) {
			breadcrumb = append(breadcrumb, r)
		}
	}

	utils.WriteJSONBody(w, utils.JSONResponse{
		Status: http.StatusOK,
		Data:   getFoldersResp{Folders: folders, Breadcrumb: breadcrumb},
	})
	return nil
}

type createFolderReceive struct {
	Name     string  `json:"name"`
	ParentID *string `json:"parent_id"`
}

func CreateFolder(w http.ResponseWriter, r *http.Request) error {
	userID := utils.GetUserID(r)

	payload, err := utils.ReadJSONBody[*createFolderReceive](w, r.Body, 1<<10)
	if err != nil {
		return err
	}

	name := strings.TrimSpace(payload.Name)
	if name == "" {
		return utils.NewHTTPError(http.StatusBadRequest, "Folder name cannot be empty.")
	}

	if payload.ParentID != nil {
		exists, err := utils.FolderExists(r.Context(), *payload.ParentID, userID)
		if err != nil {
			return err
		}
		if !exists {
			return utils.NewHTTPError(http.StatusNotFound, "Parent folder not found.")
		}
	}

	id, err := utils.GetID(10, false)
	if err != nil {
		return err
	}

	var f Folder
	if err := utils.DB.QueryRow(
		r.Context(),
		`
			INSERT INTO folders (id, user_id, parent_id, name)
			VALUES ($1, $2, $3, $4)
			RETURNING id, name, parent_id, created_at
		`,
		id, userID, payload.ParentID, name,
	).Scan(&f.ID, &f.Name, &f.ParentID, &f.CreatedAt); err != nil {
		return err
	}

	utils.WriteJSONBody(w, utils.JSONResponse{Status: http.StatusOK, Data: f})
	return nil
}

type renameFolderReceive struct {
	Name string `json:"name"`
}

func RenameFolder(w http.ResponseWriter, r *http.Request) error {
	userID := utils.GetUserID(r)
	folderID := r.PathValue("id")

	payload, err := utils.ReadJSONBody[*renameFolderReceive](w, r.Body, 1<<10)
	if err != nil {
		return err
	}

	name := strings.TrimSpace(payload.Name)
	if name == "" {
		return utils.NewHTTPError(http.StatusBadRequest, "Folder name cannot be empty.")
	}

	var f Folder
	err = utils.DB.QueryRow(
		r.Context(),
		`
			UPDATE folders SET name = $1
			WHERE id = $2 AND user_id = $3
			RETURNING id, name, parent_id, created_at
		`,
		name, folderID, userID,
	).Scan(&f.ID, &f.Name, &f.ParentID, &f.CreatedAt)
	if err != nil {
		return utils.NotFoundIfNoRows(err, "Folder not found.")
	}

	utils.WriteJSONBody(w, utils.JSONResponse{Status: http.StatusOK, Data: f})
	return nil
}

type moveFolderReceive struct {
	ParentID *string `json:"parent_id"`
}

func MoveFolder(w http.ResponseWriter, r *http.Request) error {
	userID := utils.GetUserID(r)
	folderID := r.PathValue("id")

	payload, err := utils.ReadJSONBody[*moveFolderReceive](w, r.Body, 1<<10)
	if err != nil {
		return err
	}

	if payload.ParentID != nil {
		if *payload.ParentID == folderID {
			return utils.NewHTTPError(http.StatusBadRequest, "Cannot move a folder into itself.")
		}

		exists, err := utils.FolderExists(r.Context(), *payload.ParentID, userID)
		if err != nil {
			return err
		}
		if !exists {
			return utils.NewHTTPError(http.StatusNotFound, "Parent folder not found.")
		}

		var wouldCycle bool
		if err := utils.DB.QueryRow(
			r.Context(),
			`
				WITH RECURSIVE ancestors AS (
					SELECT id, parent_id FROM folders WHERE id = $1 AND user_id = $3
					UNION ALL
					SELECT f.id, f.parent_id FROM folders f JOIN ancestors a ON f.id = a.parent_id WHERE f.user_id = $3
				)
				SELECT EXISTS (SELECT 1 FROM ancestors WHERE id = $2)
			`,
			*payload.ParentID, folderID, userID,
		).Scan(&wouldCycle); err != nil {
			return err
		}
		if wouldCycle {
			return utils.NewHTTPError(http.StatusBadRequest, "Cannot move a folder into one of its own subfolders.")
		}
	}

	var f Folder
	err = utils.DB.QueryRow(
		r.Context(),
		`
			UPDATE folders SET parent_id = $1
			WHERE id = $2 AND user_id = $3
			RETURNING id, name, parent_id, created_at
		`,
		payload.ParentID, folderID, userID,
	).Scan(&f.ID, &f.Name, &f.ParentID, &f.CreatedAt)
	if err != nil {
		return utils.NotFoundIfNoRows(err, "Folder not found.")
	}

	utils.WriteJSONBody(w, utils.JSONResponse{Status: http.StatusOK, Data: f})
	return nil
}

func DeleteFolder(w http.ResponseWriter, r *http.Request) error {
	userID := utils.GetUserID(r)
	folderID := r.PathValue("id")

	rows, err := utils.DB.Query(
		r.Context(),
		`
			WITH RECURSIVE subtree AS (
				SELECT id FROM folders WHERE id = $1 AND user_id = $2
				UNION ALL
				SELECT f.id FROM folders f JOIN subtree s ON f.parent_id = s.id WHERE f.user_id = $2
			),
			removed_media AS (
				DELETE FROM media
				WHERE folder_id IN (SELECT id FROM subtree) AND user_id = $2
				RETURNING id, has_cover, COALESCE(size, 0)::bigint * COALESCE(views, 0) AS bandwidth
			),
			bumped AS (
				UPDATE users
				SET bandwidth = bandwidth + COALESCE((SELECT SUM(bandwidth) FROM removed_media), 0)
				WHERE id = $2 AND EXISTS (SELECT 1 FROM removed_media)
			),
			removed_folders AS (
				DELETE FROM folders
				WHERE id IN (SELECT id FROM subtree) AND user_id = $2
				-- forces this to run after removed_media, so the cascade doesn't beat it to the rows
				AND (SELECT count(*) FROM removed_media) IS NOT NULL
				RETURNING id
			)
			SELECT id, has_cover FROM removed_media
		`,
		folderID, userID,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	type removedMedia struct {
		id       string
		hasCover bool
	}

	removed := make([]removedMedia, 0)
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

	for _, rm := range removed {
		utils.DeleteMediaObjects(r.Context(), rm.id, rm.hasCover)
	}

	utils.WriteJSONBody(w, utils.JSONResponse{Status: http.StatusOK})
	return nil
}
