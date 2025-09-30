package api

import (
	"errors"
	"net/http"

	"github.com/JoaoRafa19/gobid/internal/jsonutils"
	"github.com/JoaoRafa19/gobid/internal/services"
	"github.com/JoaoRafa19/gobid/internal/usecase/user"
)

// handleLoginUserV2 godoc
// @Summary      Logs in a user (v2 - Cookie Auth)
// @Description  Authenticates a user and creates a session cookie.
// @Tags         users-v2
// @Accept       json
// @Produce      json
// @Param        user  body      user.LoginUserRequest  true  "User login credentials"
// @Success      200   {object}  map[string]any
// @Failure      400   {object}  map[string]any
// @Failure      500   {object}  map[string]any
// @Router       /api/v2/users/login [post]
func (a *Api) handleLoginUserV2(w http.ResponseWriter, r *http.Request) {
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

	// Renovar token da sessão para segurança
	err = a.Sessions.RenewToken(r.Context())
	if err != nil {
		_ = jsonutils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"error": "unexpected error occurred",
		})
		return
	}

	// Armazenar user ID na sessão
	a.Sessions.Put(r.Context(), "authUserId", id)

	_ = jsonutils.EncodeJson(w, r, http.StatusOK, map[string]any{
		"message": "successfully logged in",
	})
}

// handleLogOutUserV2 godoc
// @Summary      Logs out a user (v2 - Cookie Auth)
// @Description  Logs out the currently authenticated user by destroying the session.
// @Tags         users-v2
// @Produce      json
// @Success      200  {object}  map[string]any
// @Failure      500  {object}  map[string]any
// @Router       /api/v2/users/logout [post]
func (a *Api) handleLogOutUserV2(w http.ResponseWriter, r *http.Request) {
	err := a.Sessions.RenewToken(r.Context())
	if err != nil {
		_ = jsonutils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"error": "unexpected error occurred",
		})
		return
	}

	a.Sessions.Remove(r.Context(), "authUserId")
	_ = jsonutils.EncodeJson(w, r, http.StatusOK, map[string]any{
		"message": "logged out successfully",
	})
}
