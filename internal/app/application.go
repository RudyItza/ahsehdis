package app

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/RudyItza/ahsehdis/internal/data"
	"github.com/gorilla/sessions"
)

// Application holds shared dependencies for the web application.
type Application struct {
	ErrorLog     *log.Logger
	InfoLog      *log.Logger
	DB           *sql.DB
	SessionStore *sessions.CookieStore
	UserModel    *data.UserModel
	StoryModel   *data.StoryModel
	CSRFKey      []byte // Key used for CSRF protection
}

const (
	SessionName    = "ahsehdis-session"    // Name of the session cookie
	SessionUserKey = "authenticatedUserID" // Key used to store/retrieve user ID in session
)

// InvalidCSRFHandler responds with 403 Forbidden for invalid or missing CSRF tokens
func (app *Application) InvalidCSRFHandler(w http.ResponseWriter, r *http.Request) {
	app.ClientError(w, http.StatusForbidden)
}

// ServerError logs server-side errors and returns a 500 Internal Server Error response
func (app *Application) ServerError(w http.ResponseWriter, err error) {
	app.ErrorLog.Output(2, err.Error()) // Log the error with call depth 2 (to report caller)
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
}

// ClientError sends a specific status code and its corresponding message to the client
func (app *Application) ClientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

// ContextGetUser extracts the authenticated user from the request context
func (app *Application) ContextGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value("user").(*data.User)
	if !ok {
		return nil
	}
	return user // Return the authenticated user
}
