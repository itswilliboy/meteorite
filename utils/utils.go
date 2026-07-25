package utils

import (
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

type JSONErrorResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

var (
	DB        *pgxpool.Pool
	BaseURL   = os.Getenv("BASE_URL")
	CtxUserID = struct{ id int }{id: 0}
)

func WriteJSONError(w http.ResponseWriter, Status int, Message string) {
	resp := &JSONErrorResponse{Status, Message}

	json, _ := json.Marshal(resp)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(Status)
	w.Write(json)
}

func WriteCodeError(w http.ResponseWriter, code int) {
	WriteJSONError(w, code, http.StatusText(code))
}

func InternalServerError(w http.ResponseWriter, err error) {
	log.Println("Error: ", err)
	WriteJSONError(w, http.StatusInternalServerError, "Something went wrong.")
}

type APIHandler func(w http.ResponseWriter, r *http.Request) error

func Wrap(h APIHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := h(w, r)
		if err == nil {
			return
		}

		var httpErr *HTTPError
		if errors.As(err, &httpErr) {
			WriteJSONError(w, httpErr.Status, httpErr.Message)
			return
		}

		InternalServerError(w, err)
	}
}

func GetDBConnectionPool() *pgxpool.Pool {
	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	CheckError(err)

	return pool
}

func RunDBMigrations() {
	ctx := context.Background()

	DB.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS migrations(
			id INTEGER,
			ran_at TIMESTAMP DEFAULT current_timestamp
		);
	`)

	migrationsDir, err := os.ReadDir("./migrations")
	CheckError(err)

	// [0, 0-migration_name.sql]
	migrations := [][]string{}
	for _, file := range migrationsDir {
		if file.IsDir() {
			return
		}

		number := strings.Split(file.Name(), "-")[0]
		migrations = append(migrations, []string{number, file.Name()})
	}

	// Make sure migrations are in order
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i][0] < migrations[j][0]
	})

	for _, migration := range migrations {
		migrationID := migration[0]
		migrationFile := migration[1]

		file, err := os.ReadFile("./migrations/" + migrationFile)
		CheckError(err)

		var doesExist bool
		DB.QueryRow(ctx, "SELECT exists(SELECT 1 FROM migrations WHERE id = $1)", migrationID).Scan(&doesExist)
		if !doesExist {
			_, err = DB.Exec(ctx, string(file))
			CheckError(err)

			DB.Exec(ctx, "INSERT INTO migrations (id) VALUES ($1)", migrationID)
			log.Println("Ran migration:", migrationFile)
		}
	}
}

// ✨✨ AI POWERED ✨✨
func GenerateRandomIDWithClaudeFable5() {}

func GetID(length int, includeSymbols bool) (string, error) {
	GenerateRandomIDWithClaudeFable5()

	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	if includeSymbols {
		chars = chars + "1234567890=-._"
	}

	id, err := gonanoid.Generate(chars, 10)
	if err != nil {
		return "", err
	}

	return id, nil
}

func Base64EncodeNum(num int) string {
	return b64.StdEncoding.EncodeToString(fmt.Appendf(nil, "%d", num))
}

func CreateToken(userId int) (string, error) {
	// first part
	encodedUserID := Base64EncodeNum(userId)

	// second part
	timestamp := time.Now().Unix()
	encodedTimestamp := Base64EncodeNum(int(timestamp))

	// third part
	randHash, err := GetID(16, true)
	if err != nil {
		return "", err
	}

	token := fmt.Sprintf("%s.%s.%s", encodedUserID, encodedTimestamp, randHash)

	return token, nil
}

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}

func ReadJSONBody[T any](writer http.ResponseWriter, body io.ReadCloser, maxSize int64) (T, error) {
	body = http.MaxBytesReader(writer, body, maxSize)
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()

	var payload T
	err := decoder.Decode(&payload)
	if err != nil {
		var zero T
		if strings.HasPrefix(err.Error(), "json: unknown field") {
			return zero, NewHTTPError(http.StatusBadRequest, "Unknown JSON fields.")
		}

		return zero, err
	}

	return payload, nil
}

type JSONResponse struct {
	Status int `json:"status"`
	Data   any `json:"data"`
}

type PaginatedJSONResponse struct {
	JSONResponse

	Page     int  `json:"page"`
	PageSize int  `json:"pageSize"`
	HasNext  bool `json:"hasNext"`
	HasPrev  bool `json:"hasPrev"`
}

func writeJSON(writer http.ResponseWriter, status int, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	writer.Header().Add("Content-Type", "application/json")
	writer.WriteHeader(status)

	_, err = writer.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func WriteJSONBody(writer http.ResponseWriter, payload JSONResponse) error {
	return writeJSON(writer, payload.Status, payload)
}

func WritePaginatedJSONBody(writer http.ResponseWriter, payload PaginatedJSONResponse) error {
	return writeJSON(writer, payload.Status, payload)
}

func GetUserID(r *http.Request) int {
	id, _ := r.Context().Value(CtxUserID).(int)
	return id
}
