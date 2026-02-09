package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

type Post struct {
	ID        int64     `json:"id"`
	Content   string    `json:"content"`
	Title     string    `json:"title"`
	UserID    int64     `json:"user_id"`
	Tags      []string  `json:"tags"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
	Version   int       `json:"version"`
	Comments  []Comment `json:"comments"`
}

type PostStore struct {
	db *sql.DB
}

func (p *PostStore) Create(
	ctx context.Context, post *Post) error {
	if p.db == nil {
		return errors.New("nil db in PostStore")
	}
	const query = `
	   INSERT INTO posts (content, title, user_id, tags)
	   VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at 
	   `

	err := p.db.QueryRowContext(
		ctx,
		query,
		post.Content,
		post.Title,
		post.UserID,
		pq.Array(post.Tags),
	).Scan(
		&post.ID,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

func (p *PostStore) GetByID(ctx context.Context, postID int64) (*Post, error) {
	if p.db == nil {
		return nil, errors.New("nil db in PostStore")
	}

	const query = `
        SELECT
            id,
            content,
            title,
            user_id,
            tags,
            created_at,
            updated_at,
			version
        FROM posts
        WHERE id = $1;
		`
	var post Post
	err := p.db.QueryRowContext(ctx, query, postID).Scan(
		&post.ID,
		&post.Content,
		&post.Title,
		&post.UserID,
		pq.Array(&post.Tags),
		&post.CreatedAt,
		&post.UpdatedAt,
		&post.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return &post, nil
}

func (p *PostStore) DeleteByID(ctx context.Context, postID int64) error {
	if p.db == nil {
		return errors.New("nil db in PostStore")
	}

	const query = `
       DELETE FROM posts
       WHERE id = $1
    `

	res, err := p.db.ExecContext(ctx, query, postID)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

func (p *PostStore) UpdateByID(ctx context.Context, post *Post) error {
	if p.db == nil {
		return errors.New("nil db in PostStore")
	}
	//Here we have "version" field. It helps with concurent updates.
	//For example two req readed post with same id 10 and version 1. First will write because version = $4(1)
	//Then first chenge version to 2. So when second try to execute SQL it just not find the row with id 10 and version 1
	//because version was updated to 2 by first req. Error "sql.ErrNoRows" will be returned from DB
	const query = `
       UPDATE posts
       SET title = $1, content = $2, version = version + 1
	   WHERE id = $3 AND version = $4
	   RETURNING version
    `

	err := p.db.QueryRowContext(
		ctx, query,
		post.Title,
		post.Content,
		post.ID,
		post.Version,
	).Scan(&post.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrNotFound
		default:
			return err
		}
	}
	return nil
}
