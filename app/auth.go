package app

import (
	"context"
	"fmt"
	"go-contacts/models"
	u "go-contacts/utils"
	"net/http"
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

var JwtAuthentication = func(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		noAuth := []string{"/api/user/new", "/api/user/login"} // List of endpoints allowed without auth
		requestPath := r.URL.Path                              // current request path

		// check if request does not need authentication, serve if allowed
		for _, value := range noAuth {

			if value == requestPath {
				next.ServeHTTP(w, r)
				return
			}
		}

		response := make(map[string]interface{})
		tokenHeader := r.Header.Get("Authorization") // grabs the token from the header

		if tokenHeader == "" { // handle token missing to return 401 Unauthorized
			response = u.Message(false, "missing auth token")
			w.WriteHeader(http.StatusUnauthorized)
			w.Header().Add("Content-Type", "application/json")
			w.Header().Add("WWW-Authenticate", "Newauth realm=\"apps\", type=1, title=\"Login to \"\"apps\"\", Basic realm=\"simple\"")
			u.Respond(w, response)
			return
		}

		tokenSplit := strings.Split(tokenHeader, " ") // The token comes in format `Bearer {token-body}`, we check if the retrieved token matched this requirement
		if len(tokenSplit) != 2 {
			response = u.Message(false, "Invalid/Malformed auth token")
			w.WriteHeader(http.StatusUnauthorized)
			w.Header().Add("Content-Type", "application/json")
			w.Header().Add("WWW-Authenticate", "Unrecognized token format")
			u.Respond(w, response)
			return
		}

		tokenValue := tokenSplit[1] // Take the bearer value of the token
		tk := &models.Token{}

		token, err := jwt.ParseWithClaims(tokenValue, tk, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("token_password")), nil
		})

		if err != nil { // Malformed token, returns with 401
			response = u.Message(false, "Invalid/Malformed auth token")
			w.WriteHeader(http.StatusUnauthorized)
			w.Header().Add("Content-Type", "application/json")
			w.Header().Add("WWW-Authenticate", "Unrecognized token format")
			u.Respond(w, response)
			return
		}

		if !token.Valid { // Token itself is invalid and many not be signed by this server
			response = u.Message(false, "Token is not valid")
			w.WriteHeader(http.StatusForbidden)
			w.Header().Add("Content-Type", "application/json")
			u.Respond(w, response)
			return
		}

		// In any other case, all is well and do the next thing
		fmt.Sprintf("User %", tk.UserId) // Useful for monitoring
		ctx := context.WithValue(r.Context(), "user", tk.UserId)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r) // bubble up the middleware chain
	})
}
