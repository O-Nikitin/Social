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
	User      User      `json:"user"`
}

type PostWithMetadata struct {
	Post
	CommentsCount int `json:"comments_count"`
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

func (p *PostStore) GetUserFeed(ctx context.Context, userID int64, fq PaginatedFeedQuery) ([]PostWithMetadata, error) {
	if p.db == nil {
		return nil, errors.New("nil db in PostStore")
	}

	query := `
		SELECT
			p.id, p.user_id, p.title, p.content, p.created_at, p.version, p.tags,
			u.username,
			COUNT(c.id) AS comments_count
		FROM posts p
		LEFT JOIN comments c ON c.post_id = p.id
		LEFT JOIN users u ON p.user_id = u.id
		JOIN followers f ON f.follower_id = p.user_id OR p.user_id = $1
		WHERE
			f.user_id = $1 AND
			(p.title ILIKE '%' || $4 || '%' OR p.content ILIKE '%' || $4 || '%') AND
			(p.tags @> $5 OR $5 IS NULL)
		GROUP BY p.id, u.username
		ORDER BY p.created_at ` + fq.Sort + `
		LIMIT $2 OFFSET $3
	`

	rows, err := p.db.QueryContext(
		ctx, query, userID, fq.Limit,
		fq.Offset, fq.Search, pq.Array(fq.Tags))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feed []PostWithMetadata
	for rows.Next() {
		var p PostWithMetadata
		err := rows.Scan(
			&p.ID,
			&p.UserID,
			&p.Title,
			&p.Content,
			&p.CreatedAt,
			&p.Version,
			pq.Array(&p.Tags),
			&p.User.Username,
			&p.CommentsCount,
		)
		if err != nil {
			return nil, err
		}

		feed = append(feed, p)
	}

	return feed, nil
}
