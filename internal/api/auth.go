package api

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
)

// GetUserIdFromContext extrai o user ID do contexto independentemente do tipo de autenticação
func (a *Api) GetUserIdFromContext(r *http.Request) (uuid.UUID, error) {
	userIdStr, ok := r.Context().Value(AuthUserIdKey).(string)
	if !ok {
		return uuid.Nil, errors.New("unauthorized: user not found in context")
	}

	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		return uuid.Nil, errors.New("unauthorized: invalid user id format")
	}

	return userId, nil
}
