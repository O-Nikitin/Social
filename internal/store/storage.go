package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrNotFound          = errors.New("record not found")
	ErrConflict          = errors.New("resource already exists")
	ErrDuplicateEmail    = errors.New("email already exists")
	ErrDuplicateUsername = errors.New("username already exists")
)

//go:generate mockgen -source=./storage.go -destination=../../cmd/api/mock/store/Mock_Storage.go -package=mock_storage Posts,Users,Comments,Followers,Roles

type Posts interface {
	Create(context.Context, *Post) error
	GetByID(context.Context, int64) (*Post, error)
	DeleteByID(context.Context, int64) error
	UpdateByID(context.Context, *Post) error
	GetUserFeed(context.Context, int64, PaginatedFeedQuery) ([]PostWithMetadata, error)
}

type Users interface {
	Create(context.Context, *sql.Tx, *User) error
	GetByID(context.Context, int64) (*User, error)
	GetByEmail(context.Context, string) (*User, error)
	CreateAndInvite(context.Context, *User, string, time.Duration) error
	Activate(context.Context, string) error
	Delete(context.Context, int64) error
}

type Comments interface {
	Create(context.Context, *Comment) error
	GetByPostID(context.Context, int64) ([]Comment, error)
}

type Followers interface {
	Follow(context.Context, int64, int64) error
	Unfollow(context.Context, int64, int64) error
}

type Roles interface {
	GetByName(context.Context, string) (*Role, error)
}

type Storage struct {
	Posts     Posts
	Users     Users
	Comments  Comments
	Followers Followers
	Roles     Roles
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Posts:     &PostStore{db: db},
		Users:     &UserStore{db: db},
		Comments:  &CommentStore{db: db},
		Followers: &FollowersStore{db: db},
		Roles:     &RoleStore{db: db},
	}
}

func withTx(
	db *sql.DB,
	ctx context.Context,
	fn func(*sql.Tx) error) error {

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
