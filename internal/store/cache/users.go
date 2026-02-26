package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/O-Nikitin/Social/internal/store"
	"github.com/redis/go-redis/v9"
)

const UserExpTime = time.Hour

type UserStore struct {
	rdb *redis.Client
}

func (u *UserStore) Get(ctx context.Context, userID int64) (*store.User, error) {
	if u.rdb == nil {
		return nil, errors.New("Redis cache disabled in config")
	}
	cacheKey := fmt.Sprintf("user-%d", userID)
	data, err := u.rdb.Get(ctx, cacheKey).Result()
	if err == redis.Nil { //Key not exists
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	var user store.User
	if data != "" {
		if err := json.Unmarshal([]byte(data), &user); err != nil {
			return nil, err
		}
		return &user, nil
	}
	return nil, nil

}
func (u *UserStore) Set(ctx context.Context, user *store.User) error {
	if u.rdb == nil {
		return errors.New("Redis cache disabled in config")
	}
	cacheKey := fmt.Sprintf("user-%d", user.ID)

	json, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return u.rdb.Set(ctx, cacheKey, json, UserExpTime).Err()
}
