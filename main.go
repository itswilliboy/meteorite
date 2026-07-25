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
	utils.InitStorage()

	if err := utils.InitWebAuthn(); err != nil {
		utils.CheckError(err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) { io.WriteString(w, "Pong!") })

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/app", http.StatusTemporaryRedirect)
	})

	r.Mount("/app", app())
	r.Route("/api", func(r chi.Router) {
		r.Post("/register", utils.Wrap(routes.RegisterUser))
		r.Post("/login", utils.Wrap(routes.LoginUser))

		r.Post("/webauthn/login/begin", utils.Wrap(routes.WebAuthnLoginBegin))
		r.Post("/webauthn/login/finish", utils.Wrap(routes.WebAuthnLoginFinish))

		r.Group(func(r chi.Router) {
			r.Use(DashAuthMiddleware)

			r.Post("/logout", utils.Wrap(routes.LogoutUser))
			r.Post("/reset-token", utils.Wrap(routes.ResetToken))

			r.Get("/ping", utils.Wrap(routes.DashboardPing))
			r.Get("/stats", utils.Wrap(routes.DashboardStatistics))
			r.Get("/stats/timeseries", utils.Wrap(routes.DashboardTimeseries))

			r.Get("/get-images", utils.Wrap(routes.GetImages))
			r.Post("/upload", utils.Wrap(routes.ImageUpload))
			r.Post("/delete-image", utils.Wrap(routes.DeleteImage))

			r.Post("/change-password", utils.Wrap(routes.ChangePassword))

			r.Post("/webauthn/register/begin", utils.Wrap(routes.WebAuthnRegisterBegin))
			r.Post("/webauthn/register/finish", utils.Wrap(routes.WebAuthnRegisterFinish))
			r.Get("/webauthn/credentials", utils.Wrap(routes.WebAuthnListCredentials))
			r.Post("/webauthn/credentials/{id}/delete", utils.Wrap(routes.WebAuthnDeleteCredential))
		})

		r.Route("/admin", func(r chi.Router) {
			r.Use(DashAuthAdminMiddleware)

			r.Get("/stats", utils.Wrap(routes.AdminStatistics))
			r.Get("/users", utils.Wrap(routes.AdminListUsers))
			r.Post("/users", utils.Wrap(routes.AdminCreateUser))
			r.Get("/users/{id}", utils.Wrap(routes.AdminGetUser))
			r.Post("/users/{id}/enabled", utils.Wrap(routes.AdminSetUserEnabled))
			r.Post("/users/{id}/admin", utils.Wrap(routes.AdminSetUserAdmin))
		})
	})

	// Authenticated routes
	r.Group(func(r chi.Router) {
		r.Use(AuthMiddleware)

		r.Post("/upload", utils.Wrap(routes.ImageUpload))
	})

	// Wildcards
	r.Get("/{user}/{id}", utils.Wrap(routes.ImageGet))
	r.Get("/{id}", utils.Wrap(routes.ImageRedirect))

	log.Println("Listening and serving on port 3000")
	err := http.ListenAndServe(":3000", r)
	utils.CheckError(err)
}
