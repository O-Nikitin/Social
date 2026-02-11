package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

type Follower struct {
	UserID     int64  `json:"user_id"`
	FollowerID string `json:"follower_id"`
	CreatedAt  string `json:"created_at"`
}

type FollowersStore struct {
	db *sql.DB
}

func (f *FollowersStore) Follow(ctx context.Context, followerID int64, currentUserID int64) error {
	if f.db == nil {
		return errors.New("nil db in FollowersStore")
	}
	const query = `
	INSERT INTO followers (user_id, follower_id)
	VALUES ($1, $2)
	`

	_, err := f.db.ExecContext(ctx, query, currentUserID,
		followerID)
	if err != nil {
		// 23505 = unique violation (already following)
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return ErrConflict
		}
		return err
	}
	return nil
}

func (f *FollowersStore) Unfollow(ctx context.Context, followerID int64, currentUserID int64) error {

	if f.db == nil {
		return errors.New("nil db in FollowersStore")
	}
	const query = `
	DELETE FROM followers
	WHERE user_id = $1 AND follower_id = $2
	`

	res, err := f.db.ExecContext(ctx, query, currentUserID, followerID)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	//TODO Maybe it is not needed if client interface designed well
	// and it is not possible to click unfollow on user that current
	// is not following at the moment
	if rows == 0 {
		return ErrNotFound // or ErrConflict depending on semantics
	}

	return nil
}
