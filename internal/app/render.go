package app

import (
	"html/template"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gorilla/csrf"
)

var templateFunctions = template.FuncMap{
	"humanDate": func(t time.Time) string {
		return t.Format("02 Jan 2006 at 15:04")
	},
	"truncate": func(s string, maxLength int) string {
		if len(s) <= maxLength {
			return s
		}
		return s[:maxLength] + "..."
	},
	"add": func(a, b int) int {
		return a + b
	},
	"subtract": func(a, b int) int {
		return a - b
	},
}

func (app *Application) Render(w http.ResponseWriter, r *http.Request, name string, data map[string]interface{}) {
	if data == nil {
		data = make(map[string]interface{})
	}

	// Add default data
	data[csrf.TemplateTag] = csrf.TemplateField(r)
	data["IsAuthenticated"] = app.ContextGetUser(r) != nil

	// Add flash messages if they exist
	if flashes, ok := r.Context().Value("flashes").([]interface{}); ok {
		data["Flashes"] = flashes
	}

	// Create template set with custom functions
	ts, err := template.New(name).Funcs(templateFunctions).ParseFiles(
		filepath.Join("ui", "html", "base.layout.tmpl"),
		filepath.Join("ui", "html", name),
	)

	if err != nil {
		app.ServerError(w, err)
		return
	}

	// Execute template
	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		app.ServerError(w, err)
	}
}
