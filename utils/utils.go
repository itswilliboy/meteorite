package utils

import (
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

type JSONResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

var DB *pgxpool.Pool
var BASE_URL = os.Getenv("BASE_URL")

func WriteJSONError(w http.ResponseWriter, code int, message string) {
	resp := &JSONResponse{Status: code, Message: message}

	json, _ := json.Marshal(resp)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(json)
}

func WriteCodeError(w http.ResponseWriter, code int) {
	WriteJSONError(w, code, http.StatusText(code))
}

func InternalServerError(w http.ResponseWriter, err error) {
	log.Println("Error: ", err)
	WriteJSONError(w, http.StatusInternalServerError, "Something went wrong.")
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
			id INT,
			ran_at TIMESTAMP DEFAULT NOW()
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
		file, err := os.ReadFile("./migrations/" + migration[1])
		CheckError(err)

		var doesExist bool
		DB.QueryRow(ctx, "SELECT exists(SELECT 1 FROM migrations WHERE id = $1)", migration[0]).Scan(&doesExist)
		if doesExist {
			continue
		}

		_, err = DB.Exec(ctx, string(file))
		CheckError(err)

		DB.Exec(ctx, string(file))
		log.Println("Ran migration:", migration[1])
	}
}

// ✨✨ AI POWERED ✨✨
func GenerateRandomIDWithGPT4() {}

func GetID(length int, includeSymbols bool) (string, error) {
	GenerateRandomIDWithGPT4()

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

func CreateToken(userId int, createdAt time.Time) (string, error) {
	// first part
	encodedUserId := Base64EncodeNum(userId)

	// second part
	timestamp := createdAt.Unix()
	encodedTimestamp := Base64EncodeNum(int(timestamp))

	// third part
	randHash, err := GetID(16, true)
	if err != nil {
		return "", err
	}

	token := fmt.Sprintf("%s.%s.%s", encodedUserId, encodedTimestamp, randHash)

	return token, nil
}

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}
