package main

import (
	"net/http"
	"strconv"

	"github.com/O-Nikitin/Social/internal/store"
	"github.com/go-chi/chi/v5"
)

type CreateCommentPayload struct {
	UserID  int64   `json:"userID" validate:"required,gte=1"`
	Content *string `json:"content" validate:"required,max=1000"`
}

// CreateComment godoc
//
//	@Summary		Create a comment
//	@Description	Create a new comment
//	@Tags			comments
//	@Accept			json
//	@Produce		json
//	@Param			body	body		main.CreateCommentPayload	true	"Comment data"
//	@Success		201		{object}	main.envelopeSuccess{data=store.Comment}
//	@Failure		400		{object}	main.envelopeErr	"User payload missing"
//	@Failure		500		{object}	main.envelopeErr
//	@Security		ApiKeyAuth
//	@Router			/posts/{postID}/comments [post]
func (app *application) createCommentHandler(w http.ResponseWriter, r *http.Request) {
	paramID := chi.URLParam(r, "postID")
	//TODO check req in swagger after auth added
	postID, err := strconv.ParseInt(paramID, 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	var comment CreateCommentPayload
	if err := readJSON(w, r, &comment); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(&comment); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	DBcomment := &store.Comment{
		PostID:  postID,
		UserID:  comment.UserID,
		Content: *comment.Content,
	}

	if err := app.store.Comments.Create(r.Context(), DBcomment); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, DBcomment); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
