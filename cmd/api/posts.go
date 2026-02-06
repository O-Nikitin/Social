package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/O-Nikitin/Social/internal/store"
	"github.com/go-chi/chi/v5"
)

// This struct is needed because store.Post contains internal info
// like "id", "created_at"  and we do not want to let user override it
type CreatePostPayload struct {
	Content string   `json:"content" validate:"required,gt=0, max=100"`
	Title   string   `json:"title" validate:"required,min=1,max=1000"`
	Tags    []string `json:"tags" validate:"omitempty,dive,min=1,max=20"`
}

func (app *application) createPostHandler(w http.ResponseWriter, r *http.Request) {
	var post CreatePostPayload
	if err := readJSON(w, r, &post); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(&post); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	DBpost := &store.Post{
		//TODO change after Auth
		UserID:  int64(1),
		Title:   post.Title,
		Tags:    post.Tags,
		Content: post.Content,
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := app.store.Posts.Create(ctx, DBpost); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := writeJSON(w, http.StatusCreated, DBpost); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) getPostHandler(w http.ResponseWriter, r *http.Request) {
	paramID := chi.URLParam(r, "postID")

	postID, err := strconv.ParseInt(paramID, 10, 64)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	post, err := app.store.Posts.GetByID(ctx, postID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	comments, err := app.store.Comments.GetByPostID(ctx, postID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	post.Comments = comments

	if err := writeJSON(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
