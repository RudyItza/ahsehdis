package app

import "net/http"

func (app *Application) Routes() http.Handler {
	mux := http.NewServeMux()

	// Static files with cache control
	fileServer := http.FileServer(http.Dir("./ui/static"))
	mux.Handle("/static/", http.StripPrefix("/static", app.cacheControl(fileServer)))

	// Public routes
	mux.HandleFunc("/", app.HomeHandler)
	mux.HandleFunc("/login", app.LoginForm)
	mux.HandleFunc("/login/submit", app.LoginHandler)
	mux.HandleFunc("/signup", app.SignupForm)
	mux.HandleFunc("/signup/submit", app.SignupHandler)

	// Protected routes
	mux.Handle("/stories", app.RequireAuthentication(http.HandlerFunc(app.ViewStoriesHandler)))
	mux.Handle("/story/submit", app.RequireAuthentication(http.HandlerFunc(app.SubmitStoryForm)))
	mux.Handle("/story/create", app.RequireAuthentication(http.HandlerFunc(app.SubmitStoryHandler)))
	mux.Handle("/story/edit", app.RequireAuthentication(http.HandlerFunc(app.EditStoryForm)))
	mux.Handle("/story/update", app.RequireAuthentication(http.HandlerFunc(app.EditStoryHandler)))
	mux.Handle("/story/delete", app.RequireAuthentication(http.HandlerFunc(app.DeleteStoryHandler)))
	mux.Handle("/logout", app.RequireAuthentication(http.HandlerFunc(app.LogoutHandler)))

	// Apply middleware chain
	return app.RecoverPanic(
		app.SecureHeaders(
			app.LogRequest(
				app.EnforceHTTPS(
					app.FlashMessages(
						app.Authenticate(mux),
					),
				),
			),
		),
	)
}

// cacheControl adds Cache-Control headers to static files
func (app *Application) cacheControl(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=31536000")
		next.ServeHTTP(w, r)
	})
}
