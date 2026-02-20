package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int64    `json:"id"`
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	Password  password `json:"-"`
	CreatedAt string   `json:"created_at"`
	IsActive  bool     `json:"is_active"`
}

type password struct {
	text *string //TODO do we need to store pass in memory?
	hash []byte
}

func (p *password) Set(text string) error {
	hash, err := bcrypt.GenerateFromPassword(
		[]byte(text), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	p.text = &text
	p.hash = hash

	return nil
}

func (p *password) Matches(text string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(text))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

type UserStore struct {
	db *sql.DB
}

func (u *UserStore) Create(ctx context.Context, tx *sql.Tx,
	user *User) error {
	if u.db == nil {
		return errors.New("nil db in UserStore")
	}
	const query = `
		   INSERT INTO users (username, password, email)
		   VALUES ($1, $2, $3) RETURNING id, created_at 
		   `

	err := tx.QueryRowContext(
		ctx,
		query,
		user.Username,
		user.Password.hash,
		user.Email,
	).Scan(
		&user.ID,
		&user.CreatedAt,
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" { // 23505 is Unique Violation
				switch pqErr.Constraint {
				case "users_email_key":
					return ErrDuplicateEmail
				case "users_username_key":
					return ErrDuplicateUsername
				}
			}
		}
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

func (u *UserStore) CreateAndInvite(
	ctx context.Context,
	user *User,
	token string,
	invitationExp time.Duration) error {

	return withTx(u.db, ctx, func(tx *sql.Tx) error {
		if err := u.Create(ctx, tx, user); err != nil {
			return err
		}
		if err := u.createUserInvitation(ctx, tx, token, invitationExp, user.ID); err != nil {
			return err
		}
		return nil
	})
}

func (u *UserStore) Activate(
	ctx context.Context, token string) error {
	return withTx(u.db, ctx, func(tx *sql.Tx) error {
		//as we have several steps we use transaction here
		//1. find user by token
		user, err := u.getUserFromUnvitation(
			ctx, tx, token)
		if err != nil {
			return err
		}
		//2. update the user
		user.IsActive = true
		if err := u.update(ctx, tx, user); err != nil {
			return err
		}
		//3. clean the invitations
		if err := u.deleteUserInvitations(ctx, tx, user.ID); err != nil {
			return err
		}

		return nil
	})
}

func (u *UserStore) Delete(ctx context.Context, userID int64) error {
	return withTx(u.db, ctx, func(tx *sql.Tx) error {
		if err := u.delete(ctx, tx, userID); err != nil {
			fmt.Println("ERROR! ", err.Error())
			return err
		}

		if err := u.deleteUserInvitations(ctx, tx, userID); err != nil {
			fmt.Println("ERROR!! ", err.Error())
			return err
		}

		return nil
	})
}

func (u *UserStore) delete(ctx context.Context, tx *sql.Tx, id int64) error {
	query := `DELETE FROM users WHERE id = $1`

	_, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserStore) deleteUserInvitations(ctx context.Context, tx *sql.Tx, userID int64) error {
	query := `DELETE FROM user_invitations WHERE user_id = $1`

	_, err := tx.ExecContext(ctx, query, userID)
	if err != nil {
		return err
	}

	return nil
}

func (u *UserStore) getUserFromUnvitation(
	ctx context.Context, tx *sql.Tx,
	token string) (User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.created_at, u.is_active
		FROM users u
		JOIN user_invitations ui ON u.id = ui.user_id
		WHERE ui.token = $1 AND ui.expiry > $2
		`
	hash := sha256.Sum256([]byte(token))
	hashToken := hex.EncodeToString(hash[:])
	user := User{}

	err := tx.QueryRowContext(ctx, query, hashToken, time.Now()).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.CreatedAt,
		&user.IsActive,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return User{}, ErrNotFound
		default:
			return User{}, err
		}
	}
	return user, nil
}

func (u *UserStore) createUserInvitation(ctx context.Context, tx *sql.Tx, token string, exp time.Duration, userID int64) error {
	query := `INSERT INTO user_invitations (token, user_id, expiry) VALUES ($1, $2, $3)`

	_, err := tx.ExecContext(ctx, query, token, userID, time.Now().Add(exp))
	if err != nil {
		return err
	}

	return nil
}

func (u *UserStore) update(
	ctx context.Context, tx *sql.Tx, user User) error {

	query := `UPDATE users SET username = $1, email = $2, is_active = $3 WHERE id = $4`

	_, err := tx.ExecContext(ctx, query,
		user.Username,
		user.Email,
		user.IsActive,
		user.ID)
	if err != nil {
		return err
	}

	return nil
}
