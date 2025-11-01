package main

import (
	"context"
	"net/http"
	"strconv"

	"github.com/Chandan185/Societal/internal/store"
	"github.com/go-chi/chi/v5"
)

type userKey string

var UserContextKey userKey = "user"

// GetUser godoc
//
//	@Summary		Fetches a user profile
//	@Description	Fetches a user profile by ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"User ID"
//	@Success		200	{object}	store.User
//	@Failure		400	{object}	error
//	@Failure		404	{object}	error
//	@Failure		500	{object}	error
//	@Router			/users/{id} [get]
func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromCtx(r)
	if err := app.jsonResponse(w, http.StatusOK, user); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

type FollowUser struct {
	UserID int64 `json:"user_id"`
}

// FollowUser godoc
//
//	@Summary		Follows a user
//	@Description	Follows a user profile by ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			userid	path		int		true	"User ID"
//	@Success		204		{string}	string	"User followed"
//	@Failure		404		{object}	error	"user payload invalid"
//	@Failure		400		{object}	error	"user not found"
//	@Security		ApiKeyAuth
//	@Router			/users/{userId}/follow [put]
func (app *application) followUserHandler(w http.ResponseWriter, r *http.Request) {
	followerUser := getUserFromCtx(r)

	var payload FollowUser
	if err := readJSON(w, r, &payload); err != nil {
		app.statusBadRequest(w, r, err)
		return
	}
	if err := app.store.Followers.Follow(r.Context(), followerUser.ID, payload.UserID); err != nil {
		switch err {
		case store.ErrConflict:
			app.conflictResponse(w, r, err)
			return
		default:
			app.internalServerError(w, r, err)
			return
		}
	}
	if err := app.jsonResponse(w, http.StatusNoContent, nil); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// FollowUser godoc
//
//	@Summary		unFollows a user
//	@Description	unFollows a user profile by ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			userid	path		int		true	"User ID"
//	@Success		204		{string}	string	"User unfollowed"
//	@Failure		404		{object}	error	"user payload invalid"
//	@Failure		400		{object}	error	"user not found"
//	@Security		ApiKeyAuth
//	@Router			/users/{userId}/unfollow [delete]
func (app *application) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	unfollowedUser := getUserFromCtx(r)

	var payload FollowUser
	if err := readJSON(w, r, &payload); err != nil {
		app.statusBadRequest(w, r, err)
		return
	}
	if err := app.store.Followers.Unfollow(r.Context(), unfollowedUser.ID, payload.UserID); err != nil {
		app.internalServerError(w, r, err)
		return
	}
	if err := app.jsonResponse(w, http.StatusNoContent, nil); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) userContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userId, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
		if err != nil {
			app.statusBadRequest(w, r, err)
			return
		}
		user, err := app.store.Users.GetByID(r.Context(), userId)
		if err != nil {
			switch err {
			case store.ErrNotFound:
				app.notFoundResponse(w, r, err)
				return
			default:
				app.internalServerError(w, r, err)
				return
			}
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getUserFromCtx(r *http.Request) *store.User {
	user, ok := r.Context().Value(UserContextKey).(*store.User)
	if !ok {
		return nil
	}
	return user
}
