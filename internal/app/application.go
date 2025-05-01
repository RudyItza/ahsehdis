package app

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/RudyItza/ahsehdis/internal/data"
	"github.com/gorilla/sessions"
)

type Application struct {
	ErrorLog     *log.Logger
	InfoLog      *log.Logger
	DB           *sql.DB
	SessionStore *sessions.CookieStore
	UserModel    *data.UserModel
	StoryModel   *data.StoryModel
	CSRFKey      []byte
}

const (
	SessionName    = "ahsehdis-session"
	SessionUserKey = "authenticatedUserID"
)

func (app *Application) InvalidCSRFHandler(w http.ResponseWriter, r *http.Request) {
	app.ClientError(w, http.StatusForbidden)
}

func (app *Application) ServerError(w http.ResponseWriter, err error) {
	app.ErrorLog.Output(2, err.Error())
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
}

func (app *Application) ClientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (app *Application) ContextGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value("user").(*data.User)
	if !ok {
		return nil
	}
	return user
}
