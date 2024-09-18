package api

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/egreerdp/intern-project-template/db"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

var logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{AddSource: true}))

type Handler struct {
	db db.UserStore
}

func NewHandler(db db.UserStore) *Handler {
	return &Handler{
		db: db,
	}
}

func (h Handler) MountRoutes() http.Handler {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	// r.Use(mymiddleware.JWTAuth)

	// API routes
	r.Get("/api/v1/{id}", h.HandleGetUser)
	r.Post("/api/v1/", h.HandleCreateUser)
	r.Put("/api/v1/{id}", h.HandleUpdateUser)
	r.Delete("/api/v1/{id}", h.HandleDeleteUser)

	// Serve index.html at the root "/"
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./api/views/index.html")
	})

	// Serve static files under "/static/*"
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("./api/public"))))

	r.Get("/api/v1/users", h.HandleGetUsers)
	r.Delete("/api/v1/users", h.HandleDeleteUserView)

	return r
}

func (h Handler) HandleGetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	user, err := h.db.GetUser(id)
	if err != nil {
		render.Status(r, 404)
		render.JSON(w, r, map[string]any{"msg": "user not found"})
		return
	}
	logger.Info("Got user", "id", user.ID)
	render.JSON(w, r, user)
}

func (h Handler) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	var user db.User
	b, err := io.ReadAll(r.Body)
	if err != nil {
		render.Status(r, 400)
		render.JSON(w, r, map[string]any{"msg": "malformed request data"})
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(b, &user)
	if err != nil {
		render.Status(r, 400)
		render.JSON(w, r, map[string]any{"msg": "malformed request data"})
		return
	}

	if user.Name == "" {
		render.Status(r, 400)
		render.JSON(w, r, map[string]any{"msg": "user must have a name"})
		return
	}

	userId, err := h.db.CreateUser(&user)
	if err != nil {
		render.Status(r, 500)
		render.JSON(w, r, map[string]any{"msg": "could not create user"})
		return
	}

	logger.Info("Created user", "id", userId)
	render.JSON(w, r, map[string]any{"user_id": userId})
}

func (h Handler) HandleUpdateUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var user db.User
	b, err := io.ReadAll(r.Body)
	if err != nil {
		render.Status(r, 400)
		render.JSON(w, r, map[string]any{"msg": "malformed request data"})
		return
	}
	defer r.Body.Close()

	intID, err := strconv.Atoi(id)
	if err != nil {
		render.Status(r, 400)
		render.JSON(w, r, map[string]any{"msg": "id must be a valid intger"})
	}
	user.ID = uint(intID)

	err = json.Unmarshal(b, &user)
	if err != nil {
		render.Status(r, 400)
		render.JSON(w, r, map[string]any{"msg": "malformed request data"})
		return
	}

	if user.Name == "" {
		render.Status(r, 400)
		render.JSON(w, r, map[string]any{"msg": "user must have a name"})
		return
	}

	userId, err := h.db.UpdateUser(&user)
	if err != nil {
		render.Status(r, 500)
		render.JSON(w, r, map[string]any{"msg": "could not update user"})
		return
	}

	logger.Info("Updated user", "id", userId)
	render.JSON(w, r, map[string]any{"user_id": userId})
}

func (h Handler) HandleDeleteUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userId, err := strconv.Atoi(id)
	if err != nil {
		render.Status(r, 400)
		render.JSON(w, r, map[string]any{"msg": "id must be a valid integer"})
		return
	}

	err = h.db.DeleteUser(userId)
	if err != nil {
		render.Status(r, 500)
		render.JSON(w, r, map[string]any{"msg": "could not delete user"})
		return
	}

	logger.Info("Deleted user", "id", userId)
	render.JSON(w, r, map[string]any{"user_id": userId})
}

func (h Handler) HandleGetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.db.GetUsers()
	if err != nil {
		http.Error(w, "Unable to fetch users", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("./api/public/table.html")
	if err != nil {
		http.Error(w, "Unable to parse template", http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, users)
	if err != nil {
		http.Error(w, "Unable to execute template", http.StatusInternalServerError)
		return
	}

	// Set the response content type and write the buffer to the response
	w.Header().Set("Content-Type", "text/html")
	w.Write(buf.Bytes())
}

func (h Handler) HandleDeleteUserView(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	id, err := strconv.Atoi(userID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	err = h.db.DeleteUser(id)
	if err != nil {
		http.Error(w, "Unable to delete user", http.StatusInternalServerError)
		return
	}

	// Fetch the updated user list and render the table
	users, err := h.db.GetUsers()
	if err != nil {
		http.Error(w, "Unable to fetch users", http.StatusInternalServerError)
		return
	}

	// Render the updated table
	tmpl, err := template.ParseFiles("./api/public/table.html")
	if err != nil {
		http.Error(w, "Unable to parse template", http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, users)
	if err != nil {
		http.Error(w, "Unable to execute template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write(buf.Bytes())
}
