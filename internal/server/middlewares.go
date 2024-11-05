package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
)

const (
	USER_KEY_CTX key = "key_for_auth"
)

type key string

type Middleware func(http.HandlerFunc) http.HandlerFunc

func Chain(f http.HandlerFunc, middlewares ...Middleware) http.HandlerFunc {
	for _, m := range middlewares {
		f = m(f)
	}
	return f
}

func Logging() Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			slog.Info(r.RemoteAddr, r.Method, r.RequestURI)
			next(w, r)
		}
	}
}

func Auth() Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {

			auth := r.Header.Get("Authorization")
			var res Response
			var login LoginData

			// Split the "Bearer" prefix and the token
			tokenParts := strings.SplitN(auth, " ", 2)
			if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
				w.WriteHeader(http.StatusUnauthorized)
				res = Response{Message: "Invalid Authorization header format"}
				json.NewEncoder(w).Encode(res)
				return
			}

			login, err := parseToken(tokenParts[1])
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				res = Response{Message: "Invalid token"}
				json.NewEncoder(w).Encode(res)
				return
			}

			ctx := context.WithValue(r.Context(), USER_KEY_CTX, login)

			// res = Response{Message: "Authorized...", Data: login}

			// json.NewEncoder(os.Stdout).Encode(res)
			r = r.WithContext(ctx)
			next(w, r)

		}

	}
}

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		w.Header().Set("Access-Control-Allow-Origin", origin)
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST")

			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-CSRF-Token, Authorization")
			return
		} else {
			next.ServeHTTP(w, r)
		}

	})
}

func wrapMiddleware(mux *http.ServeMux, middleware func(http.Handler) http.Handler) http.Handler {
	return middleware(mux)
}
