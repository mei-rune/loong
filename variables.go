package loong

import (
	"context"
	"net"
	"net/http"
	"strings"
)

type userKey struct{}

func (*userKey) String() string {
	return "loong-user-key"
}

var (
	UserKey = &userKey{}

	ContextWithUserHook func(ctx context.Context, u interface{}) context.Context
	UserFromContextHook func(ctx context.Context) interface{}
)

func ContextWithUser(ctx context.Context, u interface{}) context.Context {
	if ContextWithUserHook != nil {
		return ContextWithUserHook(ctx, u)
	}
	return context.WithValue(ctx, UserKey, u)
}

func UserFromContext(ctx context.Context) interface{} {
	if UserFromContextHook != nil {
		return UserFromContextHook(ctx)
	}
	return ctx.Value(UserKey)
}

type tokenKey struct{}

func (*tokenKey) String() string {
	return "loong-token-key"
}

var TokenKey = &tokenKey{}

func ContextWithToken(ctx context.Context, u interface{}) context.Context {
	return context.WithValue(ctx, TokenKey, u)
}

func TokenFromContext(ctx context.Context) interface{} {
	return ctx.Value(TokenKey)
}

type sessionKey struct{}

func (*sessionKey) String() string {
	return "loong-session-key"
}

var SessionKey = &sessionKey{}

func ContextWithSession(ctx context.Context, s interface{}) context.Context {
	return context.WithValue(ctx, SessionKey, s)
}

func SessionFromContext(ctx context.Context) interface{} {
	return ctx.Value(SessionKey)
}

type requestKey struct{}

func (*requestKey) String() string {
	return "loong-session-key"
}

var RequestKey = &requestKey{}

func ContextWithRequest(ctx context.Context, s *http.Request) context.Context {
	return context.WithValue(ctx, RequestKey, s)
}

func RequestFromContext(ctx context.Context) *http.Request {
	o := ctx.Value(RequestKey)
	if o == nil {
		return nil
	}
	req, _ := o.(*http.Request)
	return req
}

func RealIP(req *http.Request) string {
	ra := req.RemoteAddr
	if ip := req.Header.Get(HeaderXForwardedFor); ip != "" {
		ra = ip
	} else if ip := req.Header.Get(HeaderXRealIP); ip != "" {
		ra = ip
	} else {
		ra, _, _ = net.SplitHostPort(ra)
	}
	return ra
}

func IsConsumeJSON(r *http.Request) bool {
	accept := r.Header.Get(HeaderAccept)
	contentType := r.Header.Get(HeaderContentType)
	return strings.Contains(contentType, MIMEApplicationJSON) &&
		strings.Contains(accept, MIMEApplicationJSON)
}
