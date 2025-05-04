package data

import (
	"database/sql"
	"errors"
)

// StoryModel is a struct that holds a reference to the database connection (DB).
// It contains methods for interacting with the 'stories' table in the database.
type StoryModel struct {
	DB *sql.DB
}

// Insert inserts a new story into the 'stories' table and returns the story's details.
func (m *StoryModel) Insert(story *Story) error {
	// The query to insert a new story into the 'stories' table
	query := `
		INSERT INTO stories (title, content, user_id)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at`
	// Executes the query and returns the inserted story's ID, created_at, and updated_at values
	return m.DB.QueryRow(query, story.Title, story.Content, story.UserID).Scan(
		&story.ID,
		&story.CreatedAt,
		&story.UpdatedAt,
	)
}

// Get retrieves a story by its ID from the database and returns the story's details.
func (m *StoryModel) Get(id int) (*Story, error) {
	// The query to retrieve a story by its ID
	query := `
		SELECT stories.id, stories.title, stories.content, stories.user_id,
			   stories.created_at, stories.updated_at, users.email
		FROM stories
		INNER JOIN users ON stories.user_id = users.id
		WHERE stories.id = $1`

	var story Story
	// Executes the query to fetch the story details and scan the results into the 'story' struct
	err := m.DB.QueryRow(query, id).Scan(
		&story.ID,
		&story.Title,
		&story.Content,
		&story.UserID,
		&story.CreatedAt,
		&story.UpdatedAt,
		&story.UserEmail,
	)
	// If no rows are returned, return a custom error (ErrRecordNotFound)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return &story, nil
}

// GetLatest retrieves the latest 'limit' number of stories from the database.
func (m *StoryModel) GetLatest(limit int) ([]*Story, error) {
	// The query to retrieve the latest stories, ordered by creation date (descending).
	query := `
		SELECT stories.id, stories.title, LEFT(stories.content, 500) as excerpt,
			   stories.created_at, stories.updated_at, users.email
		FROM stories
		INNER JOIN users ON stories.user_id = users.id
		ORDER BY stories.created_at DESC
		LIMIT $1`
	// Executes the query to fetch the latest stories
	rows, err := m.DB.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stories []*Story
	// Iterate through the returned rows and append the story details to the stories slice
	for rows.Next() {
		var story Story
		err := rows.Scan(
			&story.ID,
			&story.Title,
			&story.Content,
			&story.CreatedAt,
			&story.UpdatedAt,
			&story.UserEmail,
		)
		if err != nil {
			return nil, err
		}
		stories = append(stories, &story)
	}

	return stories, nil
}
// GetAllPaginated retrieves all stories in a paginated manner based on the page number and page size.
func (m *StoryModel) GetAllPaginated(page, pageSize int) ([]*Story, error) {
	offset := (page - 1) * pageSize
	// The query to retrieve paginated stories, ordered by creation date (descending)
	query := `
		SELECT stories.id, stories.title, LEFT(stories.content, 500) as excerpt,
			   stories.created_at, stories.updated_at, users.email
		FROM stories
		INNER JOIN users ON stories.user_id = users.id
		ORDER BY stories.created_at DESC
		LIMIT $1 OFFSET $2`
// Executes the query to fetch the paginated stories
	rows, err := m.DB.Query(query, pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stories []*Story
	// Iterate through the returned rows and append the story details to the stories slice
	for rows.Next() {
		var story Story
		err := rows.Scan(
			&story.ID,
			&story.Title,
			&story.Content,
			&story.CreatedAt,
			&story.UpdatedAt,
			&story.UserEmail,
		)
		if err != nil {
			return nil, err
		}
		stories = append(stories, &story)
	}

	return stories, nil
}
// GetTotalCount retrieves the total number of stories in the 'stories' table.
func (m *StoryModel) GetTotalCount() (int, error) {
	var count int
		// The query to count the number of stories in the 'stories' table
	err := m.DB.QueryRow("SELECT COUNT(*) FROM stories").Scan(&count)
	return count, err
}
// Update updates a story's title and content in the 'stories' table.
func (m *StoryModel) Update(story *Story) error {
	// The query to update the story's title, content, and updated_at timestamp
	query := `
		UPDATE stories
		SET title = $1, content = $2, updated_at = NOW()
		WHERE id = $3 AND user_id = $4
		RETURNING updated_at`
		// Executes the query to update the story and retrieve the updated timestamp
	err := m.DB.QueryRow(query,
		story.Title,
		story.Content,
		story.ID,
		story.UserID,
	).Scan(&story.UpdatedAt)

	return err
}
// Delete deletes a story by its ID if it belongs to the given user ID.
func (m *StoryModel) Delete(id int, userID int) error {
	// The query to delete a story based on its ID and the user ID
	query := `
		DELETE FROM stories
		WHERE id = $1 AND user_id = $2`
// Executes the query to delete the story
	result, err := m.DB.Exec(query, id, userID)
	if err != nil {
		return err
	}
// Check how many rows were affected by the delete query
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
// If no rows were affected, the story was not found
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}
