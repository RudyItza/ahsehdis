package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/RudyItza/ahsehdis/internal/app"
	"github.com/RudyItza/ahsehdis/internal/data"
	"github.com/RudyItza/ahsehdis/internal/db"
	"github.com/gorilla/csrf"
	"github.com/gorilla/sessions"
)

func main() {
	// Configuration
	addr := flag.String("addr", ":4000", "HTTP network address")
	dsn := flag.String("dsn", "postgres://ahsehdis:rudy@localhost/ahsehdis?sslmode=disable", "PostgreSQL DSN")
	sessionKey := flag.String("session-key", "Zs6yBsEyTRu/Hw5x/tw2tSmR1VJEeCPKCdV88WU0gR8=", "Session encryption key")
	csrfKey := flag.String("csrf-key", "hD6VrOk/pCu8F7DWGNBHvbShSXZDC8W+jc4z/XBuwIY=", "CSRF encryption key")
	flag.Parse()

	// Loggers
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// Database
	dbConn, err := db.InitDBWithDSN(*dsn)
	if err != nil {
		errorLog.Fatal(err)
	}
	defer dbConn.Close()

	// Session store
	sessionStore := sessions.NewCookieStore([]byte(*sessionKey))
	sessionStore.Options = &sessions.Options{
		HttpOnly: true,
		Path:     "/",
		MaxAge:   86400 * 7,
		SameSite: http.SameSiteLaxMode,
		Secure:   false, // Set to true in production with HTTPS
	}

	// Application
	app := &app.Application{
		ErrorLog:     errorLog,
		InfoLog:      infoLog,
		DB:           dbConn,
		SessionStore: sessionStore,
		UserModel:    &data.UserModel{DB: dbConn},
		StoryModel:   &data.StoryModel{DB: dbConn},
		CSRFKey:      []byte(*csrfKey),
	}

	// CSRF protection
	csrfMiddleware := csrf.Protect(
		app.CSRFKey,
		csrf.Secure(false), // Set to true in production with HTTPS
		csrf.Path("/"),
		csrf.SameSite(csrf.SameSiteLaxMode),
		csrf.HttpOnly(true),
		csrf.FieldName("csrf_token"),
		csrf.ErrorHandler(http.HandlerFunc(app.InvalidCSRFHandler)),
	)

	// TLS config
	tlsConfig := &tls.Config{
		MinVersion:       tls.VersionTLS12,
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	// Server
	srv := &http.Server{
		Addr:         *addr,
		ErrorLog:     errorLog,
		Handler:      csrfMiddleware(app.Routes()),
		TLSConfig:    tlsConfig,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	infoLog.Printf("Starting server on %s", *addr)
	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	errorLog.Fatal(err)
}
