package store

import (
	"context"
	"database/sql"
	"errors"
)

type Comment struct {
	ID        int64  `json:"id"`
	PostID    int64  `json:"post_id"`
	UserID    int64  `json:"user_id"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	User      User   `json:"user"`
}
type CommentStore struct {
	db *sql.DB
}

func (c *CommentStore) Create(ctx context.Context, cm *Comment) error {
	query := `
	INSERT INTO comments (post_id, user_id, content)
	VALUES ($1, $2, $3)
	RETURNING id, created_at
	`

	err := c.db.QueryRowContext(
		ctx,
		query,
		cm.PostID,
		cm.UserID,
		cm.Content,
	).Scan(
		&cm.ID,
		&cm.CreatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

func (c *CommentStore) GetByPostID(ctx context.Context, postID int64) ([]Comment, error) {
	if c.db == nil {
		return nil, errors.New("nil db in CommentsStore")
	}

	query := `
		select c.id, c.post_id, c.user_id, c.content, c.created_at, users.username, users.id from comments c
		join users on users.id = c.user_id
		where c.post_id = $1
		order by c.created_at DESC;
		`

	rows, err := c.db.QueryContext(ctx, query, postID)
	if err != nil {
		return nil, err
	}
	comments := []Comment{}
	for rows.Next() {
		var c Comment
		c.User = User{}
		err := rows.Scan(
			&c.ID, &c.PostID,
			&c.UserID, &c.Content,
			&c.CreatedAt,
			&c.User.Username, &c.User.ID)
		if err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}

	return comments, nil
}
