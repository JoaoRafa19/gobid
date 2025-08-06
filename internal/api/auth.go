package api

import (
	"github.com/JoaoRafa19/gobid/internal/jsonutils"
	"github.com/gorilla/csrf"
	"net/http"
)

func (a *Api) HandleCSRFToken(w http.ResponseWriter, r *http.Request) {
	token := csrf.Token(r)
	_ = jsonutils.EncodeJson(w, r, http.StatusOK, map[string]any{
		"csrfToken": token,
	})
}

func (a *Api) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !a.Sessions.Exists(r.Context(), "authUserId") {
			_ = jsonutils.EncodeJson(w, r, http.StatusUnauthorized, map[string]any{
				"message": "must be logged in",
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}
