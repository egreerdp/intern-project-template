package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
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

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(mymiddleware.JWTAuth)

		r.Get("/{id}", h.HandleGetUser)
		r.Post("/", h.HandleCreateUser)
		r.Put("/{id}", h.HandleUpdateUser)
		r.Delete("/{id}", h.HandleDeleteUser)
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("./api/views/index.html"))

		var buf bytes.Buffer
		err := tmpl.Execute(&buf, nil)
		if err != nil {
			logger.Error("Could not execute template", "err", err)
			return
		}

		render.HTML(w, r, buf.String())
	})

	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("./api/components"))))

	r.Get("/users", h.HandleGetUsers)
	r.Delete("/users/{id}", h.HandleDeleteUserView)
	r.Get("/users/modal", h.HandleNewUserForm)
	r.Post("/users", h.HandleCreateUserView)
	r.Get("/cancel-modal", h.HandleCancel)
	r.Get("/users/{id}/edit", h.HandleEditUserModal)
	r.Post("/users/{id}", h.HandleUpdateUserView)

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

	if user.Email == "" {
		render.Status(r, 400)
		render.JSON(w, r, map[string]any{"msg": "user must have an email"})
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
		render.JSON(w, r, map[string]any{"msg": "id must be a valid integer"})
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

	if user.Email == "" {
		render.Status(r, 400)
		render.JSON(w, r, map[string]any{"msg": "user must have an email"})
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

//**************************************************************
//********************* FRONT END HANDLERS *********************
//**************************************************************

func (h Handler) HandleGetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.db.GetUsers()
	if err != nil {
		render.HTML(w, r, fmt.Sprintf("<p>%s</p>", err.Error()))
		return
	}

	tmpl, err := template.ParseFiles("./api/components/table.html")
	if err != nil {
		render.HTML(w, r, fmt.Sprintf("<p>%s</p>", err.Error()))
		return
	}

	err = tmpl.Execute(w, users)
	if err != nil {
		render.HTML(w, r, fmt.Sprintf("<p>%s</p>", err.Error()))
		return
	}
}

func (h Handler) HandleDeleteUserView(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	id, err := strconv.Atoi(userID)
	if err != nil {
		render.HTML(w, r, fmt.Sprintf("<p>%s</p>", err.Error()))
		return
	}

	err = h.db.DeleteUser(id)
	if err != nil {
		render.HTML(w, r, fmt.Sprintf("<p>%s</p>", err.Error()))
		return
	}

	users, err := h.db.GetUsers()
	if err != nil {
		render.HTML(w, r, fmt.Sprintf("<p>%s</p>", err.Error()))
		return
	}

	tmpl, err := template.ParseFiles("./api/components/table.html")
	if err != nil {
		render.HTML(w, r, fmt.Sprintf("<p>%s</p>", err.Error()))
		return
	}

	err = tmpl.Execute(w, users)
	if err != nil {
		render.HTML(w, r, fmt.Sprintf("<p>%s</p>", err.Error()))
		return
	}
}

func (h Handler) HandleNewUserForm(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("./api/components/modal_form.html")
	if err != nil {
		render.HTML(w, r, fmt.Sprintf("<p>%s</p>", err.Error()))
		return
	}
	tmpl.Execute(w, nil)
}

func (h Handler) HandleCreateUserView(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		render.HTML(w, r, fmt.Sprintf("<p>%s</p>", err.Error()))
		return
	}

	user := db.User{
		Name:  r.FormValue("name"),
		Email: r.FormValue("email"),
	}

	_, err = h.db.CreateUser(&user)
	if err != nil {
		render.HTML(w, r, fmt.Sprintf("<p>%s</p>", err.Error()))
		return
	}

	users, err := h.db.GetUsers()
	if err != nil {
		render.HTML(w, r, fmt.Sprintf("<p>%s</p>", err.Error()))
		return
	}

	tmpl, err := template.ParseFiles("./api/components/table.html")
	if err != nil {
		render.HTML(w, r, fmt.Sprintf("<p>%s</p>", err.Error()))
		return
	}

	tmpl.Execute(w, users)
}

func (h Handler) HandleCancel(w http.ResponseWriter, r *http.Request) {
	render.HTML(w, r, `<div id="modal"></div>`)
}

func (h Handler) HandleEditUserModal(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	user, err := h.db.GetUser(id)
	if err != nil {
		return
	}

	tmpl, err := template.ParseFiles("./api/components/update_user_modal.html")
	if err != nil {
		render.HTML(w, r, fmt.Sprintf("<p>%s</p>", err.Error()))
		return
	}

	tmpl.Execute(w, user)
}

func (h Handler) HandleUpdateUserView(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	name := r.FormValue("name")
	email := r.FormValue("email")

	user := db.User{
		Name:  name,
		Email: email,
	}

	intID, err := strconv.Atoi(id)
	if err != nil {
		render.Status(r, 400)
		render.HTML(w, r, err.Error())
	}
	user.ID = uint(intID)

	if user.Name == "" {
		render.Status(r, 400)
		render.HTML(w, r, err.Error())
		return
	}

	if user.Email == "" {
		render.Status(r, 400)
		render.HTML(w, r, err.Error())
		return
	}

	_, err = h.db.UpdateUser(&user)
	if err != nil {
		render.Status(r, 500)
		render.HTML(w, r, err.Error())
		return
	}

	tmpl, err := template.ParseFiles("./api/components/table.html")
	if err != nil {
		render.Status(r, 500)
		render.HTML(w, r, err.Error())
		return
	}

	users, err := h.db.GetUsers()
	if err != nil {
		render.Status(r, 500)
		render.HTML(w, r, err.Error())
		return
	}

	tmpl.Execute(w, users)
}
