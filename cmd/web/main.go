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
	//// Parse command-line flags for configuration settings configuration
	addr := flag.String("addr", ":4000", "HTTP network address")
	dsn := flag.String("dsn", "postgres://ahsehdis:rudy@localhost/ahsehdis?sslmode=disable", "PostgreSQL DSN")
	sessionKey := flag.String("session-key", "Zs6yBsEyTRu/Hw5x/tw2tSmR1VJEeCPKCdV88WU0gR8=", "Session encryption key")
	csrfKey := flag.String("csrf-key", "hD6VrOk/pCu8F7DWGNBHvbShSXZDC8W+jc4z/XBuwIY=", "CSRF encryption key")
	flag.Parse()

	//Set up custom loggers for info and error messages
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// Initialize database connection using the DSN provided
	dbConn, err := db.InitDBWithDSN(*dsn)
	if err != nil {
		errorLog.Fatal(err) // Exit if the database connection fails
	}
	defer dbConn.Close()

	// Create a new session store with encryption key
	sessionStore := sessions.NewCookieStore([]byte(*sessionKey))
	sessionStore.Options = &sessions.Options{
		HttpOnly: true,
		Path:     "/",
		MaxAge:   86400 * 7,
		SameSite: http.SameSiteLaxMode,
		Secure:   false, // Set to true in production with HTTPS
	}

	// Initialize the application struct with all dependencies
	app := &app.Application{
		ErrorLog:     errorLog,
		InfoLog:      infoLog,
		DB:           dbConn,
		SessionStore: sessionStore,
		UserModel:    &data.UserModel{DB: dbConn},
		StoryModel:   &data.StoryModel{DB: dbConn},
		CSRFKey:      []byte(*csrfKey),
	}

	// Set up CSRF protection middleware
	csrfMiddleware := csrf.Protect(
		app.CSRFKey,
		csrf.Secure(false), // Set to true in production with HTTPS
		csrf.Path("/"),
		csrf.SameSite(csrf.SameSiteLaxMode), // Protects against some types of CSRF attacks
		csrf.HttpOnly(true),
		csrf.FieldName("csrf_token"),
		csrf.ErrorHandler(http.HandlerFunc(app.InvalidCSRFHandler)),  // Custom handler when CSRF fails
	)

	// Configure TLS settings for secure communication
	tlsConfig := &tls.Config{
		MinVersion:       tls.VersionTLS12,
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	// Configure and create the HTTP server
	srv := &http.Server{
		Addr:         *addr,
		ErrorLog:     errorLog,
		Handler:      csrfMiddleware(app.Routes()),  // Routes wrapped in CSRF protection
		TLSConfig:    tlsConfig,
		IdleTimeout:  time.Minute,      // Max idle time before closing a connection
		ReadTimeout:  5 * time.Second,    // Max time to read the request
		WriteTimeout: 10 * time.Second, // Max time to write the response
	}
	// Start HTTPS server with TLS certificate and key
	infoLog.Printf("Starting server on %s", *addr)
	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")  // Paths to TLS cert and private key
	errorLog.Fatal(err) // Log any server errors and terminate
}
