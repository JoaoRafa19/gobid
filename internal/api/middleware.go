package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/JoaoRafa19/gobid/internal/jsonutils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// A custom key type to avoid context key collisions

type contextKey string

const AuthUserIdKey contextKey = "authUserId"

func (a *Api) JwtAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// Allow websocket to pass authentication to be handled by the query param
			if r.Header.Get("Upgrade") == "websocket" {
				next.ServeHTTP(w, r)
				return
			}
			_ = jsonutils.EncodeJson(w, r, http.StatusUnauthorized, map[string]any{"error": "authorization header required"})
			return
		}

		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			_ = jsonutils.EncodeJson(w, r, http.StatusUnauthorized, map[string]any{"error": "invalid authorization header format"})
			return
		}

		tokenString := headerParts[1]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(a.JwtSecret), nil
		})

		if err != nil {
			_ = jsonutils.EncodeJson(w, r, http.StatusUnauthorized, map[string]any{"error": "invalid token"})
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			userID, ok := claims["sub"].(string)
			if !ok {
				_ = jsonutils.EncodeJson(w, r, http.StatusUnauthorized, map[string]any{"error": "invalid user id in token"})
				return
			}

			// Add user ID to the context
			ctx := context.WithValue(r.Context(), AuthUserIdKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			_ = jsonutils.EncodeJson(w, r, http.StatusUnauthorized, map[string]any{"error": "invalid token"})
		}
	})
}

// WebSocketAuthMiddleware é modificado para procurar por um 'token' em vez de 'session'
func (a *Api) WebSocketAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Upgrade") == "websocket" {
			// Se o cabeçalho de autorização não estiver presente
			if r.Header.Get("Authorization") == "" {
				// Tenta obter o token de um parâmetro de consulta chamado "token"
				token := r.URL.Query().Get("token")
				if token != "" {
					// Adiciona o cabeçalho de autorização para que o JwtAuthMiddleware possa encontrá-lo.
					r.Header.Set("Authorization", "Bearer "+token)
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

// SessionAuthMiddleware é o middleware baseado em cookies para a API v2
func (a *Api) SessionAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !a.Sessions.Exists(r.Context(), "authUserId") {
			_ = jsonutils.EncodeJson(w, r, http.StatusUnauthorized, map[string]any{
				"message": "must be logged in",
			})
			return
		}

		// Obter o user ID da sessão
		userId, ok := a.Sessions.Get(r.Context(), "authUserId").(uuid.UUID)
		if !ok {
			_ = jsonutils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
				"error": "unexpected error occurred",
			})
			return
		}

		// Adicionar o user ID ao contexto
		ctx := context.WithValue(r.Context(), AuthUserIdKey, userId.String())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// WebSocketSessionAuthMiddleware é o middleware baseado em sessão para WebSockets na API v2
func (a *Api) WebSocketSessionAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Upgrade") == "websocket" {
			// Para WebSockets na v2, verificamos se há uma sessão válida
			if !a.Sessions.Exists(r.Context(), "authUserId") {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
