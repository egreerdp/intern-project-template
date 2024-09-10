package api

import (
	"net/http"

	"github.com/egreerdp/intern-project-template/db"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

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

	r.Get("/{id}", h.HandleGetUser)

	return r
}

func (h Handler) HandleGetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	w.WriteHeader(200)
	w.Write([]byte("Request Successful, id:" + id))
}
