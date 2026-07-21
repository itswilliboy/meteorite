package main

import (
	"img/routes"
	"img/utils"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

// Handler for frontend
func app() http.Handler {
	r := chi.NewRouter()

	r.Handle("/assets/*", http.StripPrefix("/app/assets/", http.FileServer(http.Dir("./assets"))))
	r.Handle("/*", http.StripPrefix("/app/", http.FileServer(http.Dir("./public"))))

	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path != "/app" {
			path = strings.Replace(path, "/app", "", 1)
		}

		filePath := filepath.Join("./public", path)

		if _, err := os.Stat(filePath); os.IsNotExist(err) || path == "/" {
			http.ServeFile(w, r, "index.html")
			return
		}

		http.ServeFile(w, r, filePath)
	})

	return r
}

func main() {
	utils.DB = utils.GetDBConnectionPool()
	defer utils.DB.Close()

	utils.RunDBMigrations()

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) { io.WriteString(w, "Pong!") })

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/app", http.StatusTemporaryRedirect)
	})

	r.Mount("/app", app())
	r.Route("/api", func(r chi.Router) {
		r.Post("/register", routes.RegisterUser)
		r.Post("/login", routes.LoginUser)

		r.Group(func(r chi.Router) {
			r.Use(DashAuthMiddleware)

			r.Post("/logout", routes.LogoutUser)
			r.Post("/reset-token", routes.ResetToken)

			r.Get("/ping", routes.DashboardPing)
			r.Get("/stats", routes.DashboardStatistics)

			r.Get("/get-images", routes.GetImages)
			r.Post("/upload", routes.ImageUpload)
			r.Post("/delete-image", routes.DeleteImage)
		})

		r.With(DashAuthAdminMiddleware).Get("/admin-stats", routes.AdminStatistics)
	})

	// Authenticated routes
	r.Group(func(r chi.Router) {
		r.Use(AuthMiddleware)

		r.Post("/upload", routes.ImageUpload)
		r.Post("/set-password", routes.ChangePassword)
	})

	// Wildcards
	r.Get("/{user}/{id}", routes.ImageGet)
	r.Get("/{id}", routes.ImageRedirect)

	log.Println("Listening and serving on port 3000")
	err := http.ListenAndServe(":3000", r)
	utils.CheckError(err)
}
