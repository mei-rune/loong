package loong

import (
	"context"
	"encoding/csv"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/runner-mei/csvutil"
	"github.com/runner-mei/errors"
	"github.com/runner-mei/log"
	echoSwagger "github.com/swaggo/echo-swagger"
)

const MyContextKey = "my-context-key"

var ErrUnsupportedMediaType = echo.ErrUnsupportedMediaType
var ErrNotFound = echo.ErrNotFound
var ErrBadArgument = errors.BadArgument
var WithHTTPCode = errors.WithHTTPCode
var Wrap = errors.Wrap
var ToHTTPError = errors.ToApplicationError
var ToError = errors.ToApplicationError

type Error = errors.Error
type HTTPError = errors.HTTPError

type Context struct {
	echo.Context
	StdContext context.Context

	CtxLogger       log.Logger
	WrapOkResult    func(c *Context, code int, i interface{}) interface{}
	WrapErrorResult func(c *Context, code int, err error) interface{}
	LogArray        []string
}

func (c *Context) QueryParamArray(name string) []string {
	results, ok := c.QueryParams()[name]
	if !ok && !strings.HasSuffix(name, "[]") {
		results = c.QueryParams()[name+"[]"]
	}
	return results
}

func (c *Context) ReturnResult(code int, i interface{}) error {
	if c.WrapOkResult != nil {
		i = c.WrapOkResult(c, code, i)
	}
	return c.JSON(code, i)
}

func (c *Context) ReturnCreatedResult(i interface{}) error {
	return c.ReturnResult(http.StatusCreated, i)
}

func (c *Context) ReturnUpdatedResult(i interface{}) error {
	return c.ReturnResult(http.StatusOK, i)
}

func (c *Context) ReturnDeletedResult(i interface{}) error {
	return c.ReturnResult(http.StatusOK, i)
}

func marshalTime(t time.Time) ([]byte, error) {
	return t.AppendFormat(nil, "2006-01-02 15:04:05Z07:00"), nil
}

func (c *Context) ReturnQueryResult(i interface{}) error {
	format := c.QueryParam("format")
	if format == "csv" {
		w := c.Response()
		w.Header().Set("Content-Type", "application/csv; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		csvWriter := csv.NewWriter(w)
		defer csvWriter.Flush()

		encoder := csvutil.NewEncoder(csvWriter)
		encoder.Register(marshalTime)
		encoder.Tag = "csv"
		return encoder.EncodeEx(i)
	}
	return c.ReturnResult(http.StatusOK, i)
}

func (c *Context) ReturnCountResult(i int64) error {
	format := c.QueryParam("format")
	if format == "csv" {
		w := c.Response()
		w.Header().Set("Content-Type", "application/csv; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "count\r\n")
		_, err := io.WriteString(w, strconv.FormatInt(i, 10))
		return err
	}

	return c.ReturnResult(http.StatusOK, i)
}

func (c *Context) ReturnError(err error, code ...int) error {
	var httpCode int
	if len(code) > 0 {
		httpCode = code[0]
	} else {
		httpCode = errors.HTTPCode(err, http.StatusInternalServerError)
	}

	if c.WrapErrorResult != nil {
		return c.JSON(httpCode, c.WrapErrorResult(c, httpCode, err))
	}

	return c.JSON(httpCode, ToHTTPError(err, httpCode))
}

var _ echo.Context = &Context{}

// HandlerFunc defines a function to serve HTTP requests.
type HandlerFunc func(*Context) error

// MiddlewareFunc defines a function to process middleware.
type MiddlewareFunc func(HandlerFunc) HandlerFunc

// Route contains a handler and information for matching against requests.
type Route = echo.Route

// Validator is the interface that wraps the Validate function.
type Validator = echo.Validator

// Renderer is the interface that wraps the Render function.
type Renderer interface {
	Render(io.Writer, string, interface{}, Context) error
}

// Map defines a generic map of type `map[string]interface{}`.
type Map map[string]interface{}

// MIME types
const (
	MIMEApplicationJSON                  = echo.MIMEApplicationJSON
	MIMEApplicationJSONCharsetUTF8       = echo.MIMEApplicationJSONCharsetUTF8
	MIMEApplicationJavaScript            = echo.MIMEApplicationJavaScript
	MIMEApplicationJavaScriptCharsetUTF8 = echo.MIMEApplicationJavaScriptCharsetUTF8
	MIMEApplicationXML                   = echo.MIMEApplicationXML
	MIMEApplicationXMLCharsetUTF8        = echo.MIMEApplicationXMLCharsetUTF8
	MIMETextXML                          = echo.MIMETextXML
	MIMETextXMLCharsetUTF8               = echo.MIMETextXMLCharsetUTF8
	MIMEApplicationForm                  = echo.MIMEApplicationForm
	MIMEApplicationProtobuf              = echo.MIMEApplicationProtobuf
	MIMEApplicationMsgpack               = echo.MIMEApplicationMsgpack
	MIMETextHTML                         = echo.MIMETextHTML
	MIMETextHTMLCharsetUTF8              = echo.MIMETextHTMLCharsetUTF8
	MIMETextPlain                        = echo.MIMETextPlain
	MIMETextPlainCharsetUTF8             = echo.MIMETextPlainCharsetUTF8
	MIMEMultipartForm                    = echo.MIMEMultipartForm
	MIMEOctetStream                      = echo.MIMEOctetStream
)

// Headers
const (
	HeaderAccept              = echo.HeaderAccept
	HeaderAcceptEncoding      = echo.HeaderAcceptEncoding
	HeaderAllow               = echo.HeaderAllow
	HeaderAuthorization       = echo.HeaderAuthorization
	HeaderContentDisposition  = echo.HeaderContentDisposition
	HeaderContentEncoding     = echo.HeaderContentEncoding
	HeaderContentLength       = echo.HeaderContentLength
	HeaderContentType         = echo.HeaderContentType
	HeaderCookie              = echo.HeaderCookie
	HeaderSetCookie           = echo.HeaderSetCookie
	HeaderIfModifiedSince     = echo.HeaderIfModifiedSince
	HeaderLastModified        = echo.HeaderLastModified
	HeaderLocation            = echo.HeaderLocation
	HeaderUpgrade             = echo.HeaderUpgrade
	HeaderVary                = echo.HeaderVary
	HeaderWWWAuthenticate     = echo.HeaderWWWAuthenticate
	HeaderXForwardedFor       = echo.HeaderXForwardedFor
	HeaderXForwardedProto     = echo.HeaderXForwardedProto
	HeaderXForwardedProtocol  = echo.HeaderXForwardedProtocol
	HeaderXForwardedSsl       = echo.HeaderXForwardedSsl
	HeaderXUrlScheme          = echo.HeaderXUrlScheme
	HeaderXHTTPMethodOverride = echo.HeaderXHTTPMethodOverride
	HeaderXRealIP             = echo.HeaderXRealIP
	HeaderXRequestID          = echo.HeaderXRequestID
	HeaderXRequestedWith      = echo.HeaderXRequestedWith
	HeaderServer              = echo.HeaderServer
	HeaderOrigin              = echo.HeaderOrigin

	// Access control
	HeaderAccessControlRequestMethod    = echo.HeaderAccessControlRequestMethod
	HeaderAccessControlRequestHeaders   = echo.HeaderAccessControlRequestHeaders
	HeaderAccessControlAllowOrigin      = echo.HeaderAccessControlAllowOrigin
	HeaderAccessControlAllowMethods     = echo.HeaderAccessControlAllowMethods
	HeaderAccessControlAllowHeaders     = echo.HeaderAccessControlAllowHeaders
	HeaderAccessControlAllowCredentials = echo.HeaderAccessControlAllowCredentials
	HeaderAccessControlExposeHeaders    = echo.HeaderAccessControlExposeHeaders
	HeaderAccessControlMaxAge           = echo.HeaderAccessControlMaxAge

	// Security
	HeaderStrictTransportSecurity         = echo.HeaderStrictTransportSecurity
	HeaderXContentTypeOptions             = echo.HeaderXContentTypeOptions
	HeaderXXSSProtection                  = echo.HeaderXXSSProtection
	HeaderXFrameOptions                   = echo.HeaderXFrameOptions
	HeaderContentSecurityPolicy           = echo.HeaderContentSecurityPolicy
	HeaderContentSecurityPolicyReportOnly = echo.HeaderContentSecurityPolicyReportOnly
	HeaderXCSRFToken                      = echo.HeaderXCSRFToken
)

type Party interface {
	Use(middleware ...MiddlewareFunc)

	// CONNECT implements `Echo#CONNECT()` for sub-routes within the Group.
	CONNECT(path string, h HandlerFunc, m ...MiddlewareFunc) *Route

	// DELETE implements `Echo#DELETE()` for sub-routes within the Group.
	DELETE(path string, h HandlerFunc, m ...MiddlewareFunc) *Route

	// GET implements `Echo#GET()` for sub-routes within the Group.
	GET(path string, h HandlerFunc, m ...MiddlewareFunc) *Route

	// HEAD implements `Echo#HEAD()` for sub-routes within the Group.
	HEAD(path string, h HandlerFunc, m ...MiddlewareFunc) *Route

	// OPTIONS implements `Echo#OPTIONS()` for sub-routes within the Group.
	OPTIONS(path string, h HandlerFunc, m ...MiddlewareFunc) *Route

	// PATCH implements `Echo#PATCH()` for sub-routes within the Group.
	PATCH(path string, h HandlerFunc, m ...MiddlewareFunc) *Route

	// POST implements `Echo#POST()` for sub-routes within the Group.
	POST(path string, h HandlerFunc, m ...MiddlewareFunc) *Route

	// PUT implements `Echo#PUT()` for sub-routes within the Group.
	PUT(path string, h HandlerFunc, m ...MiddlewareFunc) *Route

	// TRACE implements `Echo#TRACE()` for sub-routes within the Group.
	TRACE(path string, h HandlerFunc, m ...MiddlewareFunc) *Route

	// Any implements `Echo#Any()` for sub-routes within the Group.
	Any(path string, handler HandlerFunc, middleware ...MiddlewareFunc) []*Route

	// Match implements `Echo#Match()` for sub-routes within the Group.
	Match(methods []string, path string, handler HandlerFunc, middleware ...MiddlewareFunc) []*Route

	// Group creates a new sub-group with prefix and optional sub-group-level middleware.
	Group(prefix string, middleware ...MiddlewareFunc) Party

	// Static implements `Echo#Static()` for sub-routes within the Group.
	Static(prefix, root string)

	// File implements `Echo#File()` for sub-routes within the Group.
	File(path, file string)

	// Add implements `Echo#Add()` for sub-routes within the Group.
	Add(method, path string, handler HandlerFunc, middleware ...MiddlewareFunc) *Route
}

type Engine struct {
	Echo *echo.Echo

	Logger          log.Logger
	WrapOkResult    func(c *Context, code int, i interface{}) interface{}
	WrapErrorResult func(c *Context, code int, err error) interface{}

	noRoutes []struct {
		prefix  string
		handler HandlerFunc
	}
	anyNoRoutes []HandlerFunc
}

func (e *Engine) convertHandler(h HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		if actx, ok := ctx.(*Context); ok {
			return h(actx)
		}

		return h(ctx.Get(MyContextKey).(*Context))
	}
}

func (e *Engine) convertFromHandler(h echo.HandlerFunc) HandlerFunc {
	return func(ctx *Context) error {
		return h(ctx)
	}
}

func (e *Engine) convertFromPreHandler(h echo.HandlerFunc) HandlerFunc {
	return func(ctx *Context) error {
		return h(ctx.Context)
	}
}

func (e *Engine) convertMiddleware(middleware MiddlewareFunc) echo.MiddlewareFunc {
	return func(h echo.HandlerFunc) echo.HandlerFunc {
		return e.convertHandler(middleware(e.convertFromHandler(h)))
	}
}

func (e *Engine) convertPreMiddleware(middleware MiddlewareFunc) echo.MiddlewareFunc {
	return func(h echo.HandlerFunc) echo.HandlerFunc {
		return e.convertHandler(middleware(e.convertFromPreHandler(h)))
	}
}

func (e *Engine) convertMiddlewares(middlewares []MiddlewareFunc) []echo.MiddlewareFunc {
	funcs := make([]echo.MiddlewareFunc, len(middlewares))
	for idx := range middlewares {
		funcs[idx] = e.convertMiddleware(middlewares[idx])
	}
	return funcs
}

func (e *Engine) convertPreMiddlewares(middlewares []MiddlewareFunc) []echo.MiddlewareFunc {
	funcs := make([]echo.MiddlewareFunc, len(middlewares))
	for idx := range middlewares {
		funcs[idx] = e.convertPreMiddleware(middlewares[idx])
	}
	return funcs
}

// Pre adds middleware to the chain which is run before router.
func (e *Engine) Pre(middlewares ...MiddlewareFunc) {
	e.Echo.Pre(e.convertPreMiddlewares(middlewares)...)
}

// Use adds middleware to the chain which is run after router.
func (e *Engine) Use(middlewares ...MiddlewareFunc) {
	e.Echo.Use(e.convertMiddlewares(middlewares)...)
}

// CONNECT registers a new CONNECT route for a path with matching handler in the
// router with optional route-level middleware.
func (e *Engine) CONNECT(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return e.Echo.CONNECT(path, e.convertHandler(h), e.convertMiddlewares(m)...)
}

// DELETE registers a new DELETE route for a path with matching handler in the router
// with optional route-level middleware.
func (e *Engine) DELETE(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return e.Echo.DELETE(path, e.convertHandler(h), e.convertMiddlewares(m)...)
}

// GET registers a new GET route for a path with matching handler in the router
// with optional route-level middleware.
func (e *Engine) GET(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return e.Echo.GET(path, e.convertHandler(h), e.convertMiddlewares(m)...)
}

// HEAD registers a new HEAD route for a path with matching handler in the
// router with optional route-level middleware.
func (e *Engine) HEAD(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return e.Echo.HEAD(path, e.convertHandler(h), e.convertMiddlewares(m)...)
}

// OPTIONS registers a new OPTIONS route for a path with matching handler in the
// router with optional route-level middleware.
func (e *Engine) OPTIONS(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return e.Echo.OPTIONS(path, e.convertHandler(h), e.convertMiddlewares(m)...)
}

// PATCH registers a new PATCH route for a path with matching handler in the
// router with optional route-level middleware.
func (e *Engine) PATCH(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return e.Echo.PATCH(path, e.convertHandler(h), e.convertMiddlewares(m)...)
}

// POST registers a new POST route for a path with matching handler in the
// router with optional route-level middleware.
func (e *Engine) POST(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return e.Echo.POST(path, e.convertHandler(h), e.convertMiddlewares(m)...)
}

// PUT registers a new PUT route for a path with matching handler in the
// router with optional route-level middleware.
func (e *Engine) PUT(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return e.Echo.PUT(path, e.convertHandler(h), e.convertMiddlewares(m)...)
}

// TRACE registers a new TRACE route for a path with matching handler in the
// router with optional route-level middleware.
func (e *Engine) TRACE(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return e.Echo.TRACE(path, e.convertHandler(h), e.convertMiddlewares(m)...)
}

// Any registers a new route for all HTTP methods and path with matching handler
// in the router with optional route-level middleware.
func (e *Engine) Any(path string, handler HandlerFunc, m ...MiddlewareFunc) []*Route {
	return e.Echo.Any(path, e.convertHandler(handler), e.convertMiddlewares(m)...)
}

// Add registers a new route for an HTTP method and path with matching handler
// in the router with optional route-level middleware.
func (e *Engine) Add(method, path string, handler HandlerFunc, m ...MiddlewareFunc) *Route {
	return e.Echo.Add(method, path, e.convertHandler(handler), e.convertMiddlewares(m)...)
}

// Match implements `Echo#Match()` for sub-routes within the Group.
func (e *Engine) Match(methods []string, path string, handler HandlerFunc, m ...MiddlewareFunc) []*Route {
	return e.Echo.Match(methods, path, e.convertHandler(handler), e.convertMiddlewares(m)...)
}

// File registers a new route with path to serve a static file with optional route-level middleware.
func (e *Engine) File(path, file string) {
	e.Echo.File(path, file)
}

// Static implements `Echo#Static()` for sub-routes within the Group.
func (e *Engine) Static(prefix, root string) {
	e.Echo.Static(prefix, root)
}

func (e *Engine) Group(prefix string, m ...MiddlewareFunc) Party {
	g := e.Echo.Group(prefix, e.convertMiddlewares(m)...)
	return &Group{e, g}
}

func (e *Engine) NoRoute(prefix string, handler HandlerFunc, m ...MiddlewareFunc) {
	e.noRoutes = append(e.noRoutes, struct {
		prefix  string
		handler HandlerFunc
	}{
		prefix:  prefix,
		handler: handler,
	})
}

func (e *Engine) NoRouteAny(handler HandlerFunc) {
	e.anyNoRoutes = append(e.anyNoRoutes, handler)
}

type Group struct {
	engine *Engine
	group  *echo.Group
}

// Use adds middleware to the chain which is run after router.
func (e *Group) Use(middlewares ...MiddlewareFunc) {
	e.group.Use(e.engine.convertMiddlewares(middlewares)...)
}

// CONNECT registers a new CONNECT route for a path with matching handler in the
// router with optional route-level middleware.
func (g *Group) CONNECT(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return g.group.CONNECT(path, g.engine.convertHandler(h), g.engine.convertMiddlewares(m)...)
}

// DELETE registers a new DELETE route for a path with matching handler in the router
// with optional route-level middleware.
func (g *Group) DELETE(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return g.group.DELETE(path, g.engine.convertHandler(h), g.engine.convertMiddlewares(m)...)
}

// GET registers a new GET route for a path with matching handler in the router
// with optional route-level middleware.
func (g *Group) GET(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return g.group.GET(path, g.engine.convertHandler(h), g.engine.convertMiddlewares(m)...)
}

// HEAD registers a new HEAD route for a path with matching handler in the
// router with optional route-level middleware.
func (g *Group) HEAD(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return g.group.HEAD(path, g.engine.convertHandler(h), g.engine.convertMiddlewares(m)...)
}

// OPTIONS registers a new OPTIONS route for a path with matching handler in the
// router with optional route-level middleware.
func (g *Group) OPTIONS(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return g.group.OPTIONS(path, g.engine.convertHandler(h), g.engine.convertMiddlewares(m)...)
}

// PATCH registers a new PATCH route for a path with matching handler in the
// router with optional route-level middleware.
func (g *Group) PATCH(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return g.group.PATCH(path, g.engine.convertHandler(h), g.engine.convertMiddlewares(m)...)
}

// POST registers a new POST route for a path with matching handler in the
// router with optional route-level middleware.
func (g *Group) POST(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return g.group.POST(path, g.engine.convertHandler(h), g.engine.convertMiddlewares(m)...)
}

// PUT registers a new PUT route for a path with matching handler in the
// router with optional route-level middleware.
func (g *Group) PUT(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return g.group.PUT(path, g.engine.convertHandler(h), g.engine.convertMiddlewares(m)...)
}

// TRACE registers a new TRACE route for a path with matching handler in the
// router with optional route-level middleware.
func (g *Group) TRACE(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return g.group.TRACE(path, g.engine.convertHandler(h), g.engine.convertMiddlewares(m)...)
}

// Any registers a new route for all HTTP methods and path with matching handler
// in the router with optional route-level middleware.
func (g *Group) Any(path string, handler HandlerFunc, m ...MiddlewareFunc) []*Route {
	return g.group.Any(path, g.engine.convertHandler(handler), g.engine.convertMiddlewares(m)...)
}

// Add registers a new route for an HTTP method and path with matching handler
// in the router with optional route-level middleware.
func (g *Group) Add(method, path string, handler HandlerFunc, m ...MiddlewareFunc) *Route {
	return g.group.Add(method, path, g.engine.convertHandler(handler), g.engine.convertMiddlewares(m)...)
}

// Match implements `Echo#Match()` for sub-routes within the Group.
func (g *Group) Match(methods []string, path string, handler HandlerFunc, m ...MiddlewareFunc) []*Route {
	return g.group.Match(methods, path, g.engine.convertHandler(handler), g.engine.convertMiddlewares(m)...)
}

// File registers a new route with path to serve a static file with optional route-level middleware.
func (g *Group) File(path, file string) {
	g.group.File(path, file)
}

// Static implements `Echo#Static()` for sub-routes within the Group.
func (g *Group) Static(prefix, root string) {
	g.group.Static(prefix, root)
}

func (g *Group) Group(prefix string, m ...MiddlewareFunc) Party {
	sg := g.group.Group(prefix, g.engine.convertMiddlewares(m)...)
	return &Group{g.engine, sg}
}

func (engine *Engine) SetTracing(tracer opentracing.Tracer, componentName string, traceAll bool) *Engine {
	engine.Pre(Tracing(tracer, componentName, traceAll))
	return engine
}

func (engine *Engine) Routes() []*echo.Route {
	return engine.Echo.Routes()
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	engine.Echo.ServeHTTP(w, r)
}

func (engine *Engine) ServeHTTPWithContext(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	engine.Echo.ServeHTTP(w, r.WithContext(ctx))
}

func toContext(e *Engine, ctx echo.Context) *Context {
	req := ctx.Request()
	actx := &Context{
		Context:         ctx,
		StdContext:      req.Context(),
		WrapOkResult:    e.WrapOkResult,
		WrapErrorResult: e.WrapErrorResult,
	}
	if e.Logger != nil {
		actx.CtxLogger = e.Logger.With(log.String("http.method", req.Method), log.Stringer("http.url", req.URL))
		actx.StdContext = log.ContextWithLogger(actx.StdContext, actx.CtxLogger)
	}
	ctx.Set(MyContextKey, actx)
	return actx
}

func (engine *Engine) EnalbeSwaggerAt(prefix, instanceName string) {
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	if strings.HasSuffix(prefix, "/") {
		prefix = strings.TrimSuffix(prefix, "/")
	}

	handler := echoSwagger.EchoWrapHandler(echoSwagger.InstanceName(instanceName))

	mux := engine.Echo.Group(prefix, func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			rawRequestURI := c.Request().RequestURI
			index := strings.Index(rawRequestURI, prefix)
			if index >= 0 {
				c.Request().RequestURI = rawRequestURI[index:]
			}
			if c.Request().RequestURI == prefix ||
				c.Request().RequestURI == prefix+"/" {
				if strings.HasSuffix(rawRequestURI, "/") {
					return c.Redirect(http.StatusTemporaryRedirect, rawRequestURI+"index.html")
				} else {
					return c.Redirect(http.StatusTemporaryRedirect, rawRequestURI+"/index.html")
				}
			}
			return next(c)
		}
	})
	mux.Any("/*", handler)
}

func New() *Engine {
	e := &Engine{
		Echo: echo.New(),
	}

	// 这里没有用 middleware.RemoveTrailingSlash() 是因为它会修改 req.RequestURI, 而我不希望被修改
	e.Echo.Pre(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			url := req.URL
			path := url.Path
			l := len(path) - 1
			if l > 0 && strings.HasSuffix(path, "/") {

				// // Redirect
				// if config.RedirectCode != 0 {
				// 	uri := req.RequestURI
				// 	questIdx := strings.IndexByte(uri, '?')
				// 	if questIdx > 0 && path[questIdx - 1] == '/' {
				// 		uri = req.RequestURI[:questIdx-1] + req.RequestURI[questIdx:]
				// 	}
				// 	return c.Redirect(http.StatusTemporaryRedirect, uri)
				// }

				path = path[:l]

				// Forward
				url.Path = path
			}
			return next(c)
		}
	})
	// e.Echo.Pre(middleware.AddTrailingSlash())
	e.Echo.Pre(middleware.MethodOverrideWithConfig(middleware.MethodOverrideConfig{
		Getter: func(c echo.Context) string {
			m := c.QueryParam("_method")
			if m != "" {
				return m
			}
			return c.Request().Header.Get("X-HTTP-Method-Override")

			// 不用用  c.FormValue("_method"), 因为调用它时会读 body, 然后
			// http.Handler 就读不到无效了

			// r := c.Request()

			// var buf = bytes.NewBuffer(make([]byte, 0, len(r.ContentLength)))
			// r.Body = &teeReader{r: r.Body, w: buf}

			// m = c.FormValue("_method")
			// if buf.Len() > 0 {
			// 	r.Body = buf
			// }
			// return m
		}}))

	e.Echo.Pre(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			toContext(e, ctx)
			return next(ctx)
		}
	})

	getContext := func(ctx echo.Context) *Context {
		if actx, ok := ctx.(*Context); ok {
			return actx
		}
		if o := ctx.Get(MyContextKey); o != nil {
			if actx, ok := o.(*Context); ok {
				return actx
			}
		}
		return toContext(e, ctx)
	}

	// Middleware
	e.Echo.Use(middleware.Logger())
	e.Echo.Use(middleware.Recover())

	e.Echo.HTTPErrorHandler = echo.HTTPErrorHandler(func(err error, c echo.Context) {
		if err == echo.ErrNotFound {
			if len(e.noRoutes) > 0 {
				pa := c.Request().URL.Path
				for idx := range e.noRoutes {
					if strings.HasPrefix(pa, e.noRoutes[idx].prefix) {
						err = e.noRoutes[idx].handler(getContext(c))
						if err == nil {
							return
						}
						break
					}
				}
			}

			if len(e.anyNoRoutes) > 0 {
				ctx := getContext(c)
				for _, noRoute := range e.anyNoRoutes {
					err = noRoute(ctx)
					if err == nil {
						return
					}
					if err != echo.ErrNotFound {
						break
					}
				}
			}
		}

		if e.Logger != nil {
			if err == echo.ErrNotFound {
				e.Logger.Warn("没有找到请求的处理函数",
					log.String("method", c.Request().Method),
					log.String("url", c.Request().RequestURI),
					log.String("path", c.Request().URL.Path),
					log.Error(err))

				c.JSON(http.StatusNotFound, &Result{
					Success: false,
					Error: ToHTTPError(errors.New("url '"+c.Request().RequestURI+"' isnot found"),
						http.StatusNotFound),
				})
				return
			} else {
				e.Logger.Warn("处理请求发生错误",
					log.String("method", c.Request().Method),
					log.String("url", c.Request().RequestURI),
					log.Error(err))
			}
		}

		e.Echo.DefaultHTTPErrorHandler(err, c)
	})

	docHandler := func(c echo.Context) error {
		return c.JSON(http.StatusOK, Result{Success: true, Data: e.Echo.Routes()})
	}
	doc := e.Echo.Group("/internal").Group("/routeinfo")
	doc.Any("*", docHandler)
	doc.GET("", docHandler)
	return e
}

func WrapHandler(handler http.Handler) HandlerFunc {
	return func(c *Context) error {
		handler.ServeHTTP(c.Response(), c.Request())
		return nil
	}
}

func WrapHandlerFunc(handler http.HandlerFunc) HandlerFunc {
	return func(c *Context) error {
		handler(c.Response(), c.Request())
		return nil
	}
}

type ContextHandlerFunc func(context.Context, http.ResponseWriter, *http.Request)

func WrapContextHandler(handler ContextHandlerFunc) HandlerFunc {
	return func(c *Context) error {
		handler(c.StdContext, c.Response(), c.Request())
		return nil
	}
}

type ContextHandler interface {
	ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request)
}
