package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/Chandan185/Societal/internal/store"
	"github.com/go-chi/chi/v5"
)

type PostKey string

const PostCtx PostKey = "post"

type CreatePostPayload struct {
	Title   string   `json:"title" validate:"required,max=100"`
	Content string   `json:"content" validate:"required,max=1000"`
	Tags    []string `json:"tags"`
}

type UpdatePostPayload struct {
	Title   *string `json:"title" validate:"omitempty,max=100"`
	Content *string `json:"content" validate:"omitempty,max=1000"`
}

// createPost godoc
//
//	@Summary		Creates post
//	@Description	Create a new post
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			post	body		CreatePostPayload	true	"Post payload"
//	@Success		200		{object}	store.Post
//	@Failure		400		{object}	error	"invalid payload"
//	@Failure		500		{object}	error	"internal server error"
//	@Security		ApiKeyAuth
//	@Router			/posts [post]
func (app *application) createPostHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreatePostPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.statusBadRequest(w, r, err)
		return
	}

	if err := Validator.Struct(payload); err != nil {
		app.statusBadRequest(w, r, err)
		return
	}

	post := &store.Post{
		Title:   payload.Title,
		Content: payload.Content,
		USERID:  1,
		Tags:    payload.Tags,
	}

	ctx := r.Context()
	if err := app.store.Posts.Create(ctx, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// getPost godoc
//
//	@Summary		get post
//	@Description	get a post by ID
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			postID	path		int	true	"Post ID"
//	@Success		200		{object}	store.Post
//	@Failure		404		{object}	error	"post not found"
//	@Failure		500		{object}	error	"internal server error"
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

// deletePost godoc
//
//	@Summary		deletes post
//	@Description	delete a post by ID
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			postID	path		int		true	"Post ID"
//	@Success		204		{string}	string	"Post deleted"
//	@Failure		404		{object}	error	"post not found"
//	@Failure		500		{object}	error	"internal server error"
//	@Security		ApiKeyAuth
//	@Router			/posts/{postID} [delete]
func (app *application) deletePostHandler(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "postID")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	ctx := r.Context()
	if err = app.store.Posts.Delete(ctx, id); err != nil {
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

// updatePost godoc
//
//	@Summary		update post
//	@Description	update a post by ID
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			postID	path		int	true	"Post ID"
//	@Success		200		{object}	store.Post
//	@Failure		404		{object}	error	"post not found"
//	@Failure		500		{object}	error	"internal server error"
//	@Security		ApiKeyAuth
//	@Router			/posts/{postID} [patch]
func (app *application) updatePostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)

	var payload UpdatePostPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.statusBadRequest(w, r, err)
		return
	}

	if err := Validator.Struct(payload); err != nil {
		app.statusBadRequest(w, r, err)
		return
	}

	if payload.Title != nil {
		post.Title = *payload.Title
	}
	if payload.Content != nil {
		post.Content = *payload.Content
	}

	if err := app.store.Posts.Update(r.Context(), post); err != nil {
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
		idParam := chi.URLParam(r, "postID")
		id, err := strconv.ParseInt(idParam, 10, 64)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}
		ctx := r.Context()
		post, err := app.store.Posts.GetByID(ctx, id)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrNotFound):
				app.notFoundResponse(w, r, err)
			default:
				app.internalServerError(w, r, err)
			}
			return
		}
		ctx = context.WithValue(ctx, PostCtx, post)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getPostFromCtx(r *http.Request) *store.Post {
	post, _ := r.Context().Value(PostCtx).(*store.Post)
	return post
}
