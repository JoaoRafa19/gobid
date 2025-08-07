package api

import (
	"errors"
	"github.com/JoaoRafa19/gobid/internal/jsonutils"
	"github.com/JoaoRafa19/gobid/internal/services"
	"github.com/JoaoRafa19/gobid/internal/usecase/user"
	"net/http"
)

func (a *Api) handleSignupUser(w http.ResponseWriter, r *http.Request) {
	data, problems, err := jsonutils.DecodeValidJson[user.CreateUserRequest](r)
	if err != nil {
		_ = jsonutils.EncodeJson(w, r, http.StatusUnprocessableEntity, problems)
		return
	}

	id, err := a.UserService.CreateUser(
		r.Context(),
		data.UserName,
		data.Email,
		data.Password,
		data.Bio,
	)

	if err != nil {
		if errors.Is(err, services.ErrDuplicatedEmailOrUsername) {
			_ = jsonutils.EncodeJson(w, r, http.StatusUnprocessableEntity, map[string]any{
				"error": "email or username already in use",
			})
			return
		}

	}

	_ = jsonutils.EncodeJson(w, r, http.StatusCreated, map[string]any{
		"id": id,
	})

}

func (a *Api) handleLoginUser(w http.ResponseWriter, r *http.Request) {
	data, problems, err := jsonutils.DecodeValidJson[user.LoginUserRequest](r)
	if err != nil {
		_ = jsonutils.EncodeJson(w, r, http.StatusBadRequest, problems)
		return
	}

	id, err := a.UserService.AuthenticateUser(r.Context(), data.Email, data.Password)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			_ = jsonutils.EncodeJson(w, r, http.StatusBadRequest, map[string]any{
				"error": err.Error(),
			})
			return
		}
		_ = jsonutils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"error": "unexpected error occurred",
		})
		return
	}

	err = a.Sessions.RenewToken(r.Context())
	if err != nil {
		_ = jsonutils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"error": "unexpected error occurred",
		})
		return
	}

	a.Sessions.Put(r.Context(), "authUserId", id)

	_ = jsonutils.EncodeJson(w, r, http.StatusOK, map[string]any{
		"message": "successfully logged in",
	})
}

func (a *Api) handleLogOutUser(w http.ResponseWriter, r *http.Request) {
	err := a.Sessions.RenewToken(r.Context())
	if err != nil {
		_ = jsonutils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"error": "unexpected error occurred",
		})
		return
	}

	a.Sessions.Remove(r.Context(), "authUserId")
	_ = jsonutils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
		"message": "logged out successfully",
	})
}
