package main

import (
	"fmt"
	"img/routes"
	"img/utils"
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

func index(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Bye, world.")
}

func uid(w http.ResponseWriter, r *http.Request) {
	id, _ := r.Context().Value("userId").(int)
	user := utils.GetUserByID(id)

	io.WriteString(w, fmt.Sprintf("User ID: %d, username: %s", user.Id, user.Name))
}

func main() {
	utils.DB = utils.GetDBConnectionPool()
	utils.RunDBMigrations()
	defer utils.DB.Close()

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// Public routes
	r.Post("/register", routes.RegisterUser)
	r.Get("/", index)

	// Authenticated routes
	r.With(AuthMiddleware).Get("/uid", uid)
	r.With(AuthMiddleware).Post("/upload", routes.ImageUpload)

	// Wildcard
	r.Get("/{user}/{id}", routes.ImageGet)
	r.Get("/{id}", routes.Test)

	log.Println("Listening and serving on port 3000")
	err := http.ListenAndServe(":3000", r)
	utils.CheckError(err)
}
