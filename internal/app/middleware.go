package app

import (
	"context"
	"net/http"
	"time"
)

func (app *Application) LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		app.InfoLog.Printf("%s %s %s", r.Method, r.URL.Path, r.Proto)
		next.ServeHTTP(w, r)
		app.InfoLog.Printf("%s %s completed in %v", r.Method, r.URL.Path, time.Since(start))
	})
}

func (app *Application) RecoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.ServerError(w, err.(error))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (app *Application) SecureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}

func (app *Application) EnforceHTTPS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip for local development
		if r.Host == "localhost:4000" || r.Host == "127.0.0.1:4000" {
			next.ServeHTTP(w, r)
			return
		}
		
		if r.Header.Get("X-Forwarded-Proto") != "https" {
			httpsURL := "https://" + r.Host + r.URL.String()
			http.Redirect(w, r, httpsURL, http.StatusPermanentRedirect)
			return
		}
		
		// Set HSTS header for browsers
		w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		next.ServeHTTP(w, r)
	})
}

func (app *Application) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := app.SessionStore.Get(r, SessionName)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		userID, ok := session.Values[SessionUserKey].(int)
		if !ok {
			next.ServeHTTP(w, r)
			return
		}

		user, err := app.UserModel.GetByID(userID)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *Application) FlashMessages(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := app.SessionStore.Get(r, SessionName)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		if flashes := session.Flashes(); len(flashes) > 0 {
			ctx := context.WithValue(r.Context(), "flashes", flashes)
			r = r.WithContext(ctx)
			
			if err := session.Save(r, w); err != nil {
				app.ServerError(w, err)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (app *Application) RequireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.ContextGetUser(r) == nil {
			session, err := app.SessionStore.Get(r, SessionName)
			if err != nil {
				app.ServerError(w, err)
				return
			}

			session.AddFlash("Please login to access this page")
			if err := session.Save(r, w); err != nil {
				app.ServerError(w, err)
				return
			}

			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, post-check=0, pre-check=0")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		next.ServeHTTP(w, r)
	})
}