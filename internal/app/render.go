package app

import (
	"html/template"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gorilla/csrf"
)

// Define a map of custom template functions that can be used in templates.
var templateFunctions = template.FuncMap{
	// Format a time.Time value into a human-readable string.
	"humanDate": func(t time.Time) string {
		return t.Format("02 Jan 2006 at 15:04")
	},
	// Truncate a string to a maximum length, adding "..." if it exceeds.
	"truncate": func(s string, maxLength int) string {
		if len(s) <= maxLength {
			return s
		}
		return s[:maxLength] + "..."
	},
	// Add two integers.
	"add": func(a, b int) int {
		return a + b
	},
	// Subtract the second integer from the first.
	"subtract": func(a, b int) int {
		return a - b
	},
}

// Render renders an HTML template and writes it to the response writer.
func (app *Application) Render(w http.ResponseWriter, r *http.Request, name string, data map[string]interface{}) {
	// If no data map is provided, initialize an empty one.
	if data == nil {
		data = make(map[string]interface{})
	}

	// Inject the CSRF protection field into the template data
	data[csrf.TemplateTag] = csrf.TemplateField(r)
	// Indicate whether a user is currently authenticated.

	data["IsAuthenticated"] = app.ContextGetUser(r) != nil
	// If there are flash messages in the request context, add them to the data.
	if flashes, ok := r.Context().Value("flashes").([]interface{}); ok {
		data["Flashes"] = flashes
	}

	// Parse the base layout and the specified page template file, applying custom functions.
	ts, err := template.New(name).Funcs(templateFunctions).ParseFiles(
		filepath.Join("ui", "html", "base.layout.tmpl"),
		filepath.Join("ui", "html", name),
	)

	if err != nil {
		// If parsing fails, log the error and send a 500 response.
		app.ServerError(w, err)
		return
	}

	// Execute the base layout template with the provided data.// Execute the base layout template with the provided data.
	err = ts.ExecuteTemplate(w, "base", data)
	// If template execution fails, log the error and send a 500 response.
	if err != nil {
		app.ServerError(w, err)
	}
}
