package loong

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/runner-mei/errors"
)

var (
	ErrUnauthorized       = errors.NewHTTPError(http.StatusUnauthorized, "auth: token is unauthorized")
	ErrTokenExpired       = errors.NewHTTPError(http.StatusUnauthorized, "auth: token is expired")
	ErrTokenNotFound      = errors.NewHTTPError(http.StatusUnauthorized, "auth: no token found")
	ErrUserNotFound       = errors.NewHTTPError(http.StatusForbidden, "auth: user isnot exists")
	ErrInvalidCredentials = errors.NewHTTPError(http.StatusForbidden, "auth: invalid credentials")

	// 仅用于 token 找到，但不适用检验函数的时候
	ErrSkipped            = errors.NewHTTPError(http.StatusForbidden, "auth: has not check token")
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

				if !errors.Is(err, ErrTokenNotFound) {
					return ctx.ReturnError(err, http.StatusUnauthorized)
				}
			}
			return ctx.ReturnError(ErrTokenNotFound, http.StatusUnauthorized)
		}
		return HandlerFunc(hfn)
	}
}

func RawHTTPAuth(returnError func(http.ResponseWriter, *http.Request, string, int), validateFns ...AuthValidateFunc) func(ContextHandlerFunc) ContextHandlerFunc {
	if returnError == nil {
		returnError = func(w http.ResponseWriter, r *http.Request, err string, statusCode int) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(statusCode)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"code":  statusCode,
				"error": err,
			})
		}
	}

	return func(next ContextHandlerFunc) ContextHandlerFunc {
		hfn := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			for _, fn := range validateFns {
				nctx, err := fn(ctx, r)
				if err == nil {
					next(nctx, w, r)
					return
				}

				if err != ErrTokenNotFound {
					returnError(w, r, err.Error(), http.StatusUnauthorized)
					return
				}
			}
			returnError(w, r, ErrTokenNotFound.Error(), http.StatusUnauthorized)
		}
		return ContextHandlerFunc(hfn)
	}
}
