package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func New(addr string, maxOpenConns,
	maxIdleConns int, maxIdleTime string) (*sql.DB, error) {
	fmt.Println(addr)
	db, err := sql.Open("pgx", addr)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)

	duration, err := time.ParseDuration(maxIdleTime)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(time.Duration(duration))
	ctx, cancel := context.WithTimeout(
		context.Background(), time.Second*5)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}
