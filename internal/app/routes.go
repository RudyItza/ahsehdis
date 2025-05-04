package app

import "net/http"

// Routes defines the main application routes and middleware stack.
func (app *Application) Routes() http.Handler {
	// Create a new HTTP request multiplexer.
	mux := http.NewServeMux()

	// Serve static files (CSS, JS, images) from the /ui/static directory.
	// Adds Cache-Control headers to encourage long-term caching.
	fileServer := http.FileServer(http.Dir("./ui/static"))
	mux.Handle("/static/", http.StripPrefix("/static", app.cacheControl(fileServer)))

	// Public routes (Accessible without login)
	mux.HandleFunc("/", app.HomeHandler)
	mux.HandleFunc("/login", app.LoginForm)
	mux.HandleFunc("/login/submit", app.LoginHandler)
	mux.HandleFunc("/signup", app.SignupForm)
	mux.HandleFunc("/signup/submit", app.SignupHandler)

	// Protected Routes (Require user authentication)
	mux.Handle("/stories", app.RequireAuthentication(http.HandlerFunc(app.ViewStoriesHandler)))
	mux.Handle("/story/submit", app.RequireAuthentication(http.HandlerFunc(app.SubmitStoryForm)))
	mux.Handle("/story/create", app.RequireAuthentication(http.HandlerFunc(app.SubmitStoryHandler)))
	mux.Handle("/story/edit", app.RequireAuthentication(http.HandlerFunc(app.EditStoryForm)))
	mux.Handle("/story/update", app.RequireAuthentication(http.HandlerFunc(app.EditStoryHandler)))
	mux.Handle("/story/delete", app.RequireAuthentication(http.HandlerFunc(app.DeleteStoryHandler)))
	mux.Handle("/logout", app.RequireAuthentication(http.HandlerFunc(app.LogoutHandler)))

	// -------- Middleware Stack --------
	// Wrap the entire mux with a chain of middleware for:
	// - Panic recovery
	// - Secure HTTP headers
	// - Request logging
	// - HTTPS enforcement
	// - Flash message support
	// - User authentication context loading
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

// cacheControl is a middleware that sets a long-term cache policy for static assets.
func (app *Application) cacheControl(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Instructs the browser to cache static resources for 1 year.
		w.Header().Set("Cache-Control", "public, max-age=31536000")
		next.ServeHTTP(w, r)
	})
}
