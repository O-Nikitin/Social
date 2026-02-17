package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/O-Nikitin/Social/internal/store"
	"github.com/go-chi/chi/v5"
)

type postKey string

const postCtx postKey = "post"

// This struct is needed because store.Post contains internal info
// like "id", "created_at"  and we do not want to let user override it
type CreatePostPayload struct {
	Content string   `json:"content" validate:"required,gt=0,max=100"`
	Title   string   `json:"title" validate:"required,min=1,max=1000"`
	Tags    []string `json:"tags" validate:"omitempty,dive,min=1,max=20"`
}

type UpdatePostPayload struct {
	Title   *string `json:"title" validate:"omitempty,max=100"`
	Content *string `json:"content" validate:"omitempty,max=1000"`
}

// CreatePost godoc
//
//	@Summary		Create post
//	@Description	Create new post
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			body	body		main.CreatePostPayload	true	"Post data"
//	@Success		201		{object}	main.envelopeSuccess.{data=store.Post}
//	@Failure		400		{object}	main.envelopeErr	"User payload missing"
//	@Failure		500		{object}	main.envelopeErr
//	@Security		ApiKeyAuth
//	@Router			/posts [post]
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
		UserID:  int64(88),
		Title:   post.Title,
		Tags:    post.Tags,
		Content: post.Content,
	}

	if err := app.store.Posts.Create(r.Context(), DBpost); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, DBpost); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// GetPost godoc
//
//	@Summary		Get user post
//	@Description	Get Post by ID
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			postID	path		int	true	"postID"
//	@Success		200		{object}	main.envelopeSuccess{data=store.Post}
//	@Failure		400		{object}	main.envelopeErr	"User payload missing"
//	@Failure		404		{object}	main.envelopeErr
//	@Failure		500		{object}	main.envelopeErr
//	@Security		ApiKeyAuth
//	@Router			/posts/{postID} [get]
func (app *application) getPostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)

	comments, err := app.store.Comments.GetByPostID(r.Context(), post.ID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	post.Comments = comments

	if err := app.jsonResponse(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// DeletePost godoc
//
//	@Summary		Delete post
//	@Description	Delete existing post
//	@Tags			posts
//	@Param			postID	path	int	true	"postID"
//	@Success		204		"Post deleted successfully"
//	@Failure		404		{object}	main.envelopeErr
//	@Failure		500		{object}	main.envelopeErr
//	@Security		ApiKeyAuth
//	@Router			/posts/{postID} [delete]
func (app *application) deletePostHandler(w http.ResponseWriter, r *http.Request) {
	paramID := chi.URLParam(r, "postID")

	postID, err := strconv.ParseInt(paramID, 10, 64)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	err = app.store.Posts.DeleteByID(r.Context(), postID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdatePost godoc
//
//	@Summary		Update post
//	@Description	Update existing post
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			body	body		main.UpdatePostPayload	true	"post data"
//	@Success		200		{object}	main.envelopeSuccess{data=store.Post}
//	@Failure		400		{object}	main.envelopeErr	"User payload missing"
//	@Failure		404		{object}	main.envelopeErr
//	@Failure		500		{object}	main.envelopeErr
//	@Security		ApiKeyAuth
//	@Router			/posts/{postID} [patch]
func (app *application) updatePostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)

	var payload UpdatePostPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
	}

	if err := Validate.Struct(&payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if payload.Content != nil {
		post.Content = *payload.Content
	}

	if payload.Title != nil {
		post.Title = *payload.Title
	}

	if err := app.store.Posts.UpdateByID(r.Context(), post); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) postsContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paramID := chi.URLParam(r, "postID")

		postID, err := strconv.ParseInt(paramID, 10, 64)
		if err != nil {
			app.badRequestResponse(w, r, err)
			return
		}

		post, err := app.store.Posts.GetByID(r.Context(), postID)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrNotFound):
				app.notFoundResponse(w, r, err)
			default:
				app.internalServerError(w, r, err)
			}
			return
		}
		ctx := context.WithValue(r.Context(), postCtx, post)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getPostFromCtx(r *http.Request) *store.Post {
	post, _ := r.Context().Value(postCtx).(*store.Post)
	return post
}
