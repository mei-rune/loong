package loong

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/runner-mei/loong/util"
)

var (
	ErrUnauthorized       = &util.Error{Code: http.StatusUnauthorized, Message: "auth: token is unauthorized"}
	ErrTokenExpired       = &util.Error{Code: http.StatusUnauthorized, Message: "auth: token is expired"}
	ErrTokenNotFound      = &util.Error{Code: http.StatusUnauthorized, Message: "auth: no token found"}
	ErrUserNotFound       = &util.Error{Code: http.StatusForbidden, Message: "auth: user isnot exists"}
	ErrInvalidCredentials = &util.Error{Code: http.StatusForbidden, Message: "auth: invalid credentials"}
)

type AuthValidateFunc func(ctx context.Context, req *http.Request) (context.Context, error)

func HTTPAuth(validateFns ...AuthValidateFunc) func(HandlerFunc) HandlerFunc {
	return func(next HandlerFunc) HandlerFunc {
		hfn := func(ctx *Context) error {
			for _, fn := range validateFns {
				stdctx, err := fn(ctx.StdContext, ctx.Request())
				if err == nil {
					ctx.StdContext = stdctx
					return next(ctx)
				}

				if err != ErrTokenNotFound {
					return err
				}
			}
			return ctx.ReturnError(ErrTokenNotFound, http.StatusUnauthorized)
		}
		return HandlerFunc(hfn)
	}
}

func RawHTTPAuth(returnError func(http.ResponseWriter, string, int), validateFns ...AuthValidateFunc) func(http.Handler) http.HandlerFunc {
	if returnError == nil {
		returnError = func(w http.ResponseWriter, err string, statusCode int) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(statusCode)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"code":  statusCode,
				"error": err,
			})
		}
	}

	return func(next http.Handler) http.HandlerFunc {
		hfn := func(w http.ResponseWriter, r *http.Request) {
			for _, fn := range validateFns {
				_, err := fn(context.Background(), r)
				if err == nil {
					next.ServeHTTP(w, r)
					return
				}

				if err != ErrTokenNotFound {
					returnError(w, err.Error(), http.StatusUnauthorized)
					return
				}
			}
			returnError(w, ErrTokenNotFound.Error(), http.StatusUnauthorized)
		}
		return http.HandlerFunc(hfn)
	}
}
