package app

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/RudyItza/ahsehdis/internal/data"
)

func (app *Application) HomeHandler(w http.ResponseWriter, r *http.Request) {
	stories, err := app.StoryModel.GetLatest(10)
	if err != nil {
		app.ServerError(w, err)
		return
	}

	data := map[string]interface{}{
		"Stories": stories,
	}

	app.Render(w, r, "home.tmpl", data)
}

func (app *Application) LoginForm(w http.ResponseWriter, r *http.Request) {
	app.Render(w, r, "login.tmpl", nil)
}

func (app *Application) LoginHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.ServerError(w, err)
		return
	}

	email := r.PostForm.Get("email")
	password := r.PostForm.Get("password")

	user, err := app.UserModel.GetByEmail(email)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			app.Render(w, r, "login.tmpl", map[string]interface{}{
				"Error": "Invalid credentials",
			})
			return
		}
		app.ServerError(w, err)
		return
	}

	err = user.MatchesPassword(password)
	if err != nil {
		app.Render(w, r, "login.tmpl", map[string]interface{}{
			"Error": "Invalid credentials",
		})
		return
	}

	session, err := app.SessionStore.Get(r, SessionName)
	if err != nil {
		app.ServerError(w, err)
		return
	}

	session.Values[SessionUserKey] = user.ID
	if err := session.Save(r, w); err != nil {
		app.ServerError(w, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *Application) SignupForm(w http.ResponseWriter, r *http.Request) {
	app.Render(w, r, "signup.tmpl", nil)
}

func (app *Application) SignupHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.ServerError(w, err)
		return
	}

	email := r.PostForm.Get("email")
	password := r.PostForm.Get("password")

	v := NewValidator()
	v.Check(NotBlank(email), "email", "Email is required")
	v.Check(ValidateEmail(email), "email", "Invalid email format")
	v.Check(NotBlank(password), "password", "Password is required")
	v.Check(len(password) >= 8, "password", "Password must be at least 8 characters")

	if !v.Valid() {
		app.Render(w, r, "signup.tmpl", map[string]interface{}{
			"Errors": v.Errors,
			"Email":  email,
		})
		return
	}

	user := &data.User{Email: email}
	err = user.SetPassword(password)
	if err != nil {
		app.ServerError(w, err)
		return
	}

	err = app.UserModel.Insert(user)
	if err != nil {
		if errors.Is(err, data.ErrDuplicateEmail) {
			v.Errors["email"] = "Email already in use"
			app.Render(w, r, "signup.tmpl", map[string]interface{}{
				"Errors": v.Errors,
				"Email":  email,
			})
			return
		}
		app.ServerError(w, err)
		return
	}

	session, err := app.SessionStore.New(r, SessionName)
	if err != nil {
		app.ServerError(w, err)
		return
	}

	session.Values[SessionUserKey] = user.ID
	if err := session.Save(r, w); err != nil {
		app.ServerError(w, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *Application) SubmitStoryForm(w http.ResponseWriter, r *http.Request) {
	app.Render(w, r, "submit_story.tmpl", nil)
}

func (app *Application) SubmitStoryHandler(w http.ResponseWriter, r *http.Request) {
	user := app.ContextGetUser(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	err := r.ParseForm()
	if err != nil {
		app.ServerError(w, err)
		return
	}

	title := r.PostForm.Get("title")
	content := r.PostForm.Get("content")

	v := NewValidator()
	v.Check(NotBlank(title), "title", "Title is required")
	v.Check(len(title) <= 100, "title", "Title too long (max 100 chars)")
	v.Check(NotBlank(content), "content", "Content is required")

	if !v.Valid() {
		app.Render(w, r, "submit_story.tmpl", map[string]interface{}{
			"Errors":  v.Errors,
			"Title":   title,
			"Content": content,
		})
		return
	}

	story := &data.Story{
		Title:   title,
		Content: content,
		UserID:  user.ID,
	}

	err = app.StoryModel.Insert(story)
	if err != nil {
		app.ServerError(w, err)
		return
	}

	session, err := app.SessionStore.Get(r, SessionName)
	if err != nil {
		app.ServerError(w, err)
		return
	}

	session.AddFlash("Story created successfully!")
	if err := session.Save(r, w); err != nil {
		app.ServerError(w, err)
		return
	}

	http.Redirect(w, r, "/stories", http.StatusSeeOther)
}

func (app *Application) ViewStoriesHandler(w http.ResponseWriter, r *http.Request) {
	const storiesPerPage = 10

	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1
	}

	stories, err := app.StoryModel.GetAllPaginated(page, storiesPerPage)
	if err != nil {
		app.ServerError(w, err)
		return
	}

	totalStories, err := app.StoryModel.GetTotalCount()
	if err != nil {
		app.ServerError(w, err)
		return
	}

	totalPages := (totalStories + storiesPerPage - 1) / storiesPerPage

	data := map[string]interface{}{
		"Stories": stories,
		"Pagination": struct {
			Current int
			Total   int
			HasNext bool
			HasPrev bool
		}{
			Current: page,
			Total:   totalPages,
			HasNext: page < totalPages,
			HasPrev: page > 1,
		},
	}

	app.Render(w, r, "view_stories.tmpl", data)
}

func (app *Application) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, err := app.SessionStore.Get(r, SessionName)
	if err != nil {
		app.ServerError(w, err)
		return
	}

	delete(session.Values, SessionUserKey)
	if err := session.Save(r, w); err != nil {
		app.ServerError(w, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *Application) EditStoryForm(w http.ResponseWriter, r *http.Request) {
	user := app.ContextGetUser(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id < 1 {
		app.ClientError(w, http.StatusBadRequest)
		return
	}

	story, err := app.StoryModel.Get(id)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			app.ClientError(w, http.StatusNotFound)
		} else {
			app.ServerError(w, err)
		}
		return
	}

	if story.UserID != user.ID {
		app.ClientError(w, http.StatusForbidden)
		return
	}

	app.Render(w, r, "edit_story.tmpl", map[string]interface{}{
		"Story": story,
	})
}

func (app *Application) EditStoryHandler(w http.ResponseWriter, r *http.Request) {
	user := app.ContextGetUser(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	err := r.ParseForm()
	if err != nil {
		app.ServerError(w, err)
		return
	}

	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil || id < 1 {
		app.ClientError(w, http.StatusBadRequest)
		return
	}

	existingStory, err := app.StoryModel.Get(id)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			app.ClientError(w, http.StatusNotFound)
		} else {
			app.ServerError(w, err)
		}
		return
	}

	if existingStory.UserID != user.ID {
		app.ClientError(w, http.StatusForbidden)
		return
	}

	title := r.FormValue("title")
	content := r.FormValue("content")

	v := NewValidator()
	v.Check(NotBlank(title), "title", "Title is required")
	v.Check(len(title) <= 100, "title", "Title too long (max 100 chars)")
	v.Check(NotBlank(content), "content", "Content is required")

	if !v.Valid() {
		app.Render(w, r, "edit_story.tmpl", map[string]interface{}{
			"Story":  existingStory,
			"Errors": v.Errors,
		})
		return
	}

	story := &data.Story{
		ID:      id,
		Title:   title,
		Content: content,
		UserID:  user.ID,
	}

	err = app.StoryModel.Update(story)
	if err != nil {
		app.ServerError(w, err)
		return
	}

	session, err := app.SessionStore.Get(r, SessionName)
	if err != nil {
		app.ServerError(w, err)
		return
	}

	session.AddFlash("Story updated successfully!")
	if err := session.Save(r, w); err != nil {
		app.ServerError(w, err)
		return
	}

	http.Redirect(w, r, "/stories", http.StatusSeeOther)
}

func (app *Application) DeleteStoryHandler(w http.ResponseWriter, r *http.Request) {
	user := app.ContextGetUser(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id < 1 {
		app.ClientError(w, http.StatusBadRequest)
		return
	}

	err = app.StoryModel.Delete(id, user.ID)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			app.ClientError(w, http.StatusNotFound)
		} else {
			app.ServerError(w, err)
		}
		return
	}

	session, err := app.SessionStore.Get(r, SessionName)
	if err != nil {
		app.ServerError(w, err)
		return
	}

	session.AddFlash("Story deleted successfully!")
	if err := session.Save(r, w); err != nil {
		app.ServerError(w, err)
		return
	}

	http.Redirect(w, r, "/stories", http.StatusSeeOther)
}
