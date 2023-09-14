package loong

import (
	"context"
	"net/http"
	"strings"

	jwt "github.com/golang-jwt/jwt/v4"
)

type TokenFindFunc func(r *http.Request) string
type TokenCheckFunc func(ctx context.Context, req *http.Request, tokenStr string) (context.Context, error)

// TokenVerify http middleware handler will verify a Token string from a http request.
//
// TokenVerify will search for a token in a http request, in the order:
//  1. 'token' URI query parameter
//  2. 'Authorization: BEARER T' request header
//  3. Cookie 'token' value
func TokenVerify(findTokenFns []TokenFindFunc, checkTokenFns []TokenCheckFunc) AuthValidateFunc {
	return func(ctx context.Context, req *http.Request) (context.Context, error) {
		var tokenStr string

		// Extract token string from the request by calling token find functions in
		// the order they where provided. Further extraction stops if a function
		// returns a non-empty string.
		for _, fn := range findTokenFns {
			tokenStr = fn(req)
			if tokenStr != "" {
				break
			}
		}
		if tokenStr == "" {
			return nil, ErrTokenNotFound
		}

		for _, fn := range checkTokenFns {
			c, err := fn(ctx, req, tokenStr)
			if err == nil || err != ErrSkipped {
				return c, err
			}
		}

		return nil, ErrUnauthorized
	}
}

func JWTCheck(ja *JWTAuth) TokenCheckFunc {
	return func(ctx context.Context, req *http.Request, tokenStr string) (context.Context, error) {
		// Verify the token
		token, err := ja.Decode(tokenStr)
		if err != nil {
			if ve, ok := err.(*jwt.ValidationError); ok {
				if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
					// Token is either expired or not active yet
					err = ErrTokenExpired
				} else {
					err = WithHTTPCode(err, http.StatusUnauthorized)
				}
			} else if strings.HasPrefix(err.Error(), "token is expired") {
				err = ErrTokenExpired
			} else {
				err = WithHTTPCode(err, http.StatusUnauthorized)
			}
			return nil, err
		}

		if token == nil || !token.Valid || token.Method != ja.signer {
			return nil, ErrUnauthorized
		}

		if token.Claims == nil {
			return nil, ErrUnauthorized
		}

		if err = token.Claims.Valid(); err != nil {
			return nil, err
		}

		return ContextWithToken(ctx, token), nil
	}
}

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

// TokenFromCookie tries to retreive the token string from a cookie named
// "token".
func TokenFromCookie(r *http.Request) string {
	cookie, err := r.Cookie("token")
	if err != nil {
		return ""
	}
	return cookie.Value
}

// TokenFromHeader tries to retreive the token string from the
// "Authorization" reqeust header: "Authorization: BEARER T".
func TokenFromHeader(r *http.Request) string {
	// Get token from authorization header.
	bearer := r.Header.Get("Authorization")
	if len(bearer) > 7 && strings.ToUpper(bearer[0:6]) == "BEARER" {
		return bearer[7:]
	}
	return ""
}

// TokenFromQuery tries to retreive the token string from the "token" URI
// query parameter.
func TokenFromQuery(r *http.Request) string {
	// Get token from query param named "token".
	return r.URL.Query().Get("token")
}
