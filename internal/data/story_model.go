package data

import (
	"database/sql"
	"errors"
)

type StoryModel struct {
	DB *sql.DB
}

func (m *StoryModel) Insert(story *Story) error {
	query := `
		INSERT INTO stories (title, content, user_id)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at`

	return m.DB.QueryRow(query, story.Title, story.Content, story.UserID).Scan(
		&story.ID,
		&story.CreatedAt,
		&story.UpdatedAt,
	)
}

func (m *StoryModel) Get(id int) (*Story, error) {
	query := `
		SELECT stories.id, stories.title, stories.content, stories.user_id,
			   stories.created_at, stories.updated_at, users.email
		FROM stories
		INNER JOIN users ON stories.user_id = users.id
		WHERE stories.id = $1`

	var story Story
	err := m.DB.QueryRow(query, id).Scan(
		&story.ID,
		&story.Title,
		&story.Content,
		&story.UserID,
		&story.CreatedAt,
		&story.UpdatedAt,
		&story.UserEmail,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return &story, nil
}

func (m *StoryModel) GetLatest(limit int) ([]*Story, error) {
	query := `
		SELECT stories.id, stories.title, LEFT(stories.content, 100) as excerpt,
			   stories.created_at, stories.updated_at, users.email
		FROM stories
		INNER JOIN users ON stories.user_id = users.id
		ORDER BY stories.created_at DESC
		LIMIT $1`

	rows, err := m.DB.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stories []*Story
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

func (m *StoryModel) GetAllPaginated(page, pageSize int) ([]*Story, error) {
	offset := (page - 1) * pageSize

	query := `
		SELECT stories.id, stories.title, LEFT(stories.content, 100) as excerpt,
			   stories.created_at, stories.updated_at, users.email
		FROM stories
		INNER JOIN users ON stories.user_id = users.id
		ORDER BY stories.created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := m.DB.Query(query, pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stories []*Story
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

func (m *StoryModel) GetTotalCount() (int, error) {
	var count int
	err := m.DB.QueryRow("SELECT COUNT(*) FROM stories").Scan(&count)
	return count, err
}

func (m *StoryModel) Update(story *Story) error {
	query := `
		UPDATE stories
		SET title = $1, content = $2, updated_at = NOW()
		WHERE id = $3 AND user_id = $4
		RETURNING updated_at`

	err := m.DB.QueryRow(query,
		story.Title,
		story.Content,
		story.ID,
		story.UserID,
	).Scan(&story.UpdatedAt)

	return err
}

func (m *StoryModel) Delete(id int, userID int) error {
	query := `
		DELETE FROM stories
		WHERE id = $1 AND user_id = $2`

	result, err := m.DB.Exec(query, id, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}
