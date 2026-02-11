package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/O-Nikitin/Social/internal/store"
	"github.com/go-chi/chi/v5"
)

type userKey string

const userCtx userKey = "user"

// TODO temp
type FollowUser struct {
	CurrentUserID int64 `json:"current_user_id"`
}

func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromCtx(r)

	if err := app.jsonResponse(w, http.StatusOK, user); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) followUserHandler(w http.ResponseWriter, r *http.Request) {
	followerUser := getUserFromCtx(r)
	//We have ID of user that we want to follow.
	//Byt we do not have current user ID which we will fetch
	//when implement autentification
	//TODO temp
	var curUserID FollowUser
	if err := readJSON(w, r, &curUserID); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	if followerUser == nil {
		log.Println("followerUser nil")
	}

	err := app.store.Followers.Follow(r.Context(), followerUser.ID, curUserID.CurrentUserID)
	if err != nil {
		switch err {
		case store.ErrConflict:
			app.conflictResponse(w, r, err)
			return
		default:
			app.internalServerError(w, r, err)
			return
		}

	}
	if err := app.jsonResponse(w, http.StatusNoContent, followerUser); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	followerUser := getUserFromCtx(r)

	//We have ID of user that we want to follow.
	//Byt we do not have current user ID which we will fetch
	//when implement autentification
	//TODO temp
	var curUserID FollowUser
	if err := readJSON(w, r, &curUserID); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	err := app.store.Followers.Unfollow(r.Context(), followerUser.ID, curUserID.CurrentUserID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := app.jsonResponse(w, http.StatusNoContent, followerUser); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) userContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paramID := chi.URLParam(r, "userID")
		userID, err := strconv.ParseInt(paramID, 10, 64)
		if err != nil {
			app.badRequestResponse(w, r, err)
			return
		}

		user, err := app.store.Users.GetByID(r.Context(), userID)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrNotFound):
				app.notFoundResponse(w, r, err)
			default:
				app.internalServerError(w, r, err)
			}
			return
		}
		log.Println(user)
		ctx := context.WithValue(r.Context(), userCtx, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getUserFromCtx(r *http.Request) *store.User {
	user, _ := r.Context().Value(userCtx).(*store.User)
	return user
}
