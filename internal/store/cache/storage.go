package cache

import (
	"context"

	"github.com/O-Nikitin/Social/internal/store"
	"github.com/redis/go-redis/v9"
)

//go:generate mockgen -source=./storage.go -destination=../../../cmd/api/mock/store/Mock_Cache.go -package=mock_storage -mock_names Users=MockUserCache Users
type Users interface {
	Get(context.Context, int64) (*store.User, error)
	Set(context.Context, *store.User) error
}

type Storage struct {
	//TODO add for posts also
	Users Users
}

func NewStorage(rdb *redis.Client) Storage {
	return Storage{
		Users: &UserStore{rdb: rdb},
	}
}
