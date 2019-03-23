package loong

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

var (
	ErrUnauthorized = errors.New("jwtauth: token is unauthorized")
	ErrExpired      = errors.New("jwtauth: token is expired")
	ErrNoTokenFound = errors.New("jwtauth: no token found")
)

type JWTAuth struct {
	signKey   interface{}
	verifyKey interface{}
	signer    jwt.SigningMethod
	parser    *jwt.Parser
}

// NewJWTAuth creates a JWTAuth authenticator instance that provides middleware handlers
// and encoding/decoding functions for JWT signing.
func NewJWTAuth(alg string, signKey interface{}, verifyKey interface{}) *JWTAuth {
	return NewJWTAuthWithParser(alg, &jwt.Parser{}, signKey, verifyKey)
}

// NewJWTAuthWithParser is the same as New, except it supports custom parser settings
// introduced in jwt-go/v2.4.0.
//
// We explicitly toggle `SkipClaimsValidation` in the `jwt-go` parser so that
// we can control when the claims are validated - in our case, by the Verifier
// http middleware handler.
func NewJWTAuthWithParser(alg string, parser *jwt.Parser, signKey interface{}, verifyKey interface{}) *JWTAuth {
	parser.SkipClaimsValidation = true
	return &JWTAuth{
		signKey:   signKey,
		verifyKey: verifyKey,
		signer:    jwt.GetSigningMethod(alg),
		parser:    parser,
	}
}

// JWTVerifier http middleware handler will verify a JWT string from a http request.
//
// JWTVerifier will search for a JWT token in a http request, in the order:
//   1. 'jwt' URI query parameter
//   2. 'Authorization: BEARER T' request header
//   3. Cookie 'jwt' value
//
// The first JWT string that is found as a query parameter, authorization header
// or cookie header is then decoded by the `jwt-go` library and a *jwt.Token
// object is set on the request context. In the case of a signature decoding error
// the Verifier will also set the error on the request context.
//
// The Verifier always calls the next http handler in sequence, which can either
// be the generic `jwtauth.Authenticator` middleware or your own custom handler
// which checks the request context jwt token and error to prepare a custom
// http response.
func JWTVerifier(ja *JWTAuth, ssoAuth func(r *http.Request) bool) func(HandlerFunc) HandlerFunc {
	if ssoAuth == nil {
		ssoAuth = func(r *http.Request) bool {
			return false
		}
	}
	return func(next HandlerFunc) HandlerFunc {
		return JWTVerify(ja, ssoAuth, JWTTokenFromQuery, JWTTokenFromHeader)(next)
	}
}

func renderError(w http.ResponseWriter, errText string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code":  statusCode,
		"error": errText,
	})
}

func JWTVerify(ja *JWTAuth, ssoAuth func(r *http.Request) bool, findTokenFns ...func(r *http.Request) string) func(HandlerFunc) HandlerFunc {
	return func(next HandlerFunc) HandlerFunc {
		hfn := func(ctx *Context) error {
			var tokenStr string

			// Extract token string from the request by calling token find functions in
			// the order they where provided. Further extraction stops if a function
			// returns a non-empty string.
			for _, fn := range findTokenFns {
				tokenStr = fn(ctx.Request())
				if tokenStr != "" {
					break
				}
			}
			if tokenStr == "" {
				if ssoAuth(ctx.Request()) {
					// SSO is authenticated, pass it through
					return next(ctx)
				}
				return ctx.ReturnError(ErrNoTokenFound, http.StatusUnauthorized)
			}

			// Verify the token
			token, err := ja.Decode(tokenStr)
			if err != nil {
				switch err.Error() {
				case "token is expired":
					err = ErrExpired
				}
				return ctx.ReturnError(err, http.StatusUnauthorized)
			}

			if token == nil || !token.Valid || token.Method != ja.signer {
				return ctx.ReturnError(ErrUnauthorized, http.StatusUnauthorized)
			}

			if token.Claims == nil {
				return ctx.ReturnError(ErrUnauthorized, http.StatusUnauthorized)
			}

			if err = token.Claims.Valid(); err != nil {
				return ctx.ReturnError(err, http.StatusUnauthorized)
			}

			if token == nil || !token.Valid {
				return ctx.ReturnError(ErrUnauthorized, http.StatusUnauthorized)
			}

			ctx.StdContext = ContextWithToken(ctx.StdContext, token)

			// Token is authenticated, pass it through
			return next(ctx)
		}
		return HandlerFunc(hfn)
	}
}

func (ja *JWTAuth) Encode(claims *jwt.StandardClaims) (t *jwt.Token, tokenString string, err error) {
	t = jwt.New(ja.signer)
	t.Claims = claims
	tokenString, err = t.SignedString(ja.signKey)
	t.Raw = tokenString
	return
}

func (ja *JWTAuth) Decode(tokenString string) (*jwt.Token, error) {
	return ja.parser.ParseWithClaims(tokenString, &jwt.StandardClaims{}, ja.keyFunc)
}

func (ja *JWTAuth) Signer() jwt.SigningMethod {
	return ja.signer
}

func (ja *JWTAuth) keyFunc(t *jwt.Token) (interface{}, error) {
	if ja.verifyKey != nil {
		return ja.verifyKey, nil
	}
	return ja.signKey, nil
}

// JWTTokenFromCookie tries to retreive the token string from a cookie named
// "jwt".
func JWTTokenFromCookie(r *http.Request) string {
	cookie, err := r.Cookie("jwt")
	if err != nil {
		return ""
	}
	return cookie.Value
}

// JWTTokenFromHeader tries to retreive the token string from the
// "Authorization" reqeust header: "Authorization: BEARER T".
func JWTTokenFromHeader(r *http.Request) string {
	// Get token from authorization header.
	bearer := r.Header.Get("Authorization")
	if len(bearer) > 7 && strings.ToUpper(bearer[0:6]) == "BEARER" {
		return bearer[7:]
	}
	return ""
}

// JWTTokenFromQuery tries to retreive the token string from the "token" URI
// query parameter.
func JWTTokenFromQuery(r *http.Request) string {
	// Get token from query param named "token".
	return r.URL.Query().Get("token")
}

// contextKey is a value for use with context.WithValue. It's used as
// a pointer so it fits in an interface{} without allocation. This technique
// for defining context keys was copied from Go 1.7's new use of context in net/http.
type contextKey struct {
	name string
}

func (k *contextKey) String() string {
	return "jwtauth context value " + k.name
}
