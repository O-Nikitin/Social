package main

import (
	"net/http"

	"github.com/O-Nikitin/Social/internal/store"
)

func (app *application) getUserFeedHandler(w http.ResponseWriter, r *http.Request) {
	// pagination, filters
	fq := store.PaginatedFeedQuery{ //default values if wasn't provided in URL
		Limit:  20,
		Offset: 0,
		Sort:   "desc",
	}

	fq, err := fq.Parse(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(fq); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	//TODO hardcoded userID
	feed, err := app.store.Posts.GetUserFeed(
		r.Context(), int64(7), fq)

	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, feed); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
