package store

import (
	"context"
	"database/sql"
	"errors"
)

type User struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"-"`
	CreatedAt string `json:"created_at"`
}

type UserStore struct {
	db *sql.DB
}

func (u *UserStore) Create(ctx context.Context,
	user *User) error {
	if u.db == nil {
		return errors.New("nil db in UserStore")
	}
	const query = `
		   INSERT INTO users (username, password, email)
		   VALUES ($1, $2, $3) RETURNING id, created_at 
		   `

	err := u.db.QueryRowContext(
		ctx,
		query,
		user.Username,
		user.Password,
		user.Email,
	).Scan(
		&user.ID,
		&user.CreatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

func (u *UserStore) GetByID(ctx context.Context, userID int64) (*User, error) {
	if u.db == nil {
		return nil, errors.New("nil db in UserStore")
	}

	const query = `
        SELECT
			id,
            email,
			username,
			password,
			created_at
        FROM users
        WHERE id = $1;
		`
	var user User
	err := u.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.Password,
		&user.CreatedAt,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}
