package api

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/egreerdp/intern-project-template/db"
	mymiddleware "github.com/egreerdp/intern-project-template/internal/middleware"
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

	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(mymiddleware.JWTAuth)

	r.Get("/{id}", h.HandleGetUser)
	r.Post("/", h.HandleCreateUser)
	r.Put("/{id}", h.HandleUpdateUser)
	r.Delete("/{id}", h.HandleDeleteUser)

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
