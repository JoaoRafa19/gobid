package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/JoaoRafa19/gobid/internal/jsonutils"
	"github.com/JoaoRafa19/gobid/internal/services"
	"github.com/JoaoRafa19/gobid/internal/usecase/user"
	"github.com/golang-jwt/jwt/v5"
)

// handleSignupUser godoc
// @Summary      Sign up a new user
// @Description  Creates a new user account.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        user  body      user.CreateUserRequest  true  "User signup information"
// @Success      201   {object}  map[string]any
// @Failure      422   {object}  map[string]any
// @Failure      500   {object}  map[string]any
// @Router       /api/v2/users/signup [post]
// @Router       /api/v1/users/signup [post]
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

// handleLoginUser godoc
// @Summary      Logs in a user
// @Description  Authenticates a user and provides a session token.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        user  body      user.LoginUserRequest  true  "User login credentials"
// @Success      200   {object}  map[string]string
// @Failure      400   {object}  map[string]any
// @Failure      500   {object}  map[string]any
// @Router       /users/login [post]
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

	// Create JWT Claims
	claims := jwt.MapClaims{
		"sub": id,                                    // Subject (user id)
		"iat": time.Now().Unix(),                     // Issued At
		"exp": time.Now().Add(time.Hour * 24).Unix(), // Expiration Time
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token with secret
	tss, err := token.SignedString([]byte(a.JwtSecret))
	if err != nil {
		_ = jsonutils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"error": "failed to sign token",
		})
		return
	}

	_ = jsonutils.EncodeJson(w, r, http.StatusOK, map[string]any{
		"token": tss,
	})
}
