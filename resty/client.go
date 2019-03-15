package resty

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

var BufferPool sync.Pool

func init() {
	BufferPool.New = func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, 1024))
	}
}

type HTTPError struct {
	err      error
	httpCode int
}

func (e *HTTPError) Error() string {
	return e.err.Error()
}

func (e *HTTPError) HTTPCode() int {
	return e.httpCode
}

func WithHTTPCode(code int, err error) *HTTPError {
	return &HTTPError{err: err, httpCode: code}
}

type ResponseFunc func(req *http.Request, resp *http.Response) error

type Proxy struct {
	Client        *http.Client
	JSONUseNumber bool
	u             url.URL
	queryParams   url.Values
	headers       url.Values
}

func (proxy *Proxy) Release(request *Request) {}

func (proxy *Proxy) New(urlStr string) *Request {
	r := &Request{
		proxy:         proxy,
		jsonUseNumber: proxy.JSONUseNumber,
		u:             proxy.u,
		queryParams:   url.Values{},
		headers:       url.Values{},
	}

	for key, values := range proxy.queryParams {
		r.queryParams[key] = values
	}
	for key, values := range proxy.headers {
		r.headers[key] = values
	}

	if urlStr != "" {
		u, err := url.Parse(urlStr)
		if err != nil {
			panic(err)
		}

		if u.Scheme != "" {
			r.u = *u
			r.queryParams = u.Query()
		} else {
			r.u.Path = Join(r.u.Path, u.Path)
			for key, values := range u.Query() {
				r.queryParams[key] = values
			}
		}
	}

	return r
}

type Request struct {
	proxy         *Proxy
	jsonUseNumber bool
	u             url.URL
	queryParams   url.Values
	headers       url.Values
	requestBody   interface{}
	exceptedCode  int
	responseBody  interface{}
}

func (r *Request) JSONUseNumber() *Request {
	r.jsonUseNumber = true
	return r
}

func (r *Request) SetHeader(key, value string) *Request {
	r.headers.Set(key, value)
	return r
}
func (r *Request) SetParam(key, value string) *Request {
	r.queryParams.Set(key, value)
	return r
}
func (r *Request) AddParam(key, value string) *Request {
	r.queryParams.Add(key, value)
	return r
}
func (r *Request) SetBody(body interface{}) *Request {
	r.requestBody = body
	return r
}
func (r *Request) Result(body interface{}) *Request {
	r.requestBody = body
	return r
}
func (r *Request) ExceptedCode(code int) *Request {
	r.exceptedCode = code
	return r
}
func (r *Request) GET(ctx context.Context) error {
	return r.invoke(ctx, "GET")
}
func (r *Request) POST(ctx context.Context) error {
	return r.invoke(ctx, "POST")
}
func (r *Request) PUT(ctx context.Context) error {
	return r.invoke(ctx, "PUT")
}
func (r *Request) CONNECT(ctx context.Context) error {
	return r.invoke(ctx, "CONNECT")
}
func (r *Request) DELETE(ctx context.Context) error {
	return r.invoke(ctx, "DELETE")
}
func (r *Request) HEAD(ctx context.Context) error {
	return r.invoke(ctx, "HEAD")
}
func (r *Request) OPTIONS(ctx context.Context) error {
	return r.invoke(ctx, "OPTIONS")
}
func (r *Request) PATCH(ctx context.Context) error {
	return r.invoke(ctx, "PATCH")
}
func (r *Request) TRACE(ctx context.Context) error {
	return r.invoke(ctx, "TRACE")
}
func (r *Request) Do(ctx context.Context, method string) error {
	return r.invoke(ctx, method)
}
func (r *Request) invoke(ctx context.Context, method string) error {
	var req *http.Request

	var body io.Reader
	if body != nil {
		switch value := r.requestBody.(type) {
		case []byte:
			body = bytes.NewReader(value)
		case string:
			body = strings.NewReader(value)
		case io.Reader:
			body = value
		default:
			buffer := BufferPool.Get().(*bytes.Buffer)
			e := json.NewEncoder(buffer).Encode(body)
			if nil != e {
				return WithHTTPCode(http.StatusBadRequest, e)
			}
			body = buffer
			defer func() {
				buffer.Reset()
				BufferPool.Put(buffer)
			}()
		}
	}

	r.u.RawQuery = r.queryParams.Encode()
	urlStr := r.u.String()
	req, e := http.NewRequest(method, urlStr, body)
	if e != nil {
		return WithHTTPCode(http.StatusBadRequest, e)
	}
	if ctx != nil {
		req = req.WithContext(ctx)
	}

	for key, values := range r.headers {
		req.Header[key] = values
	}

	resp, e := r.proxy.Client.Do(req)
	if nil != e {
		return WithHTTPCode(http.StatusServiceUnavailable, e)
	}

	isOK := false
	if r.exceptedCode == 0 {
		if resp.StatusCode >= http.StatusOK && resp.StatusCode <= 299 {
			isOK = true
		}
	} else if resp.StatusCode == r.exceptedCode {
		isOK = true
	}

	if !isOK {
		var respBody []byte

		if nil != resp.Body {
			respBody, e = ioutil.ReadAll(resp.Body)
			resp.Body.Close()

			if nil != e {
				panic(e.Error())
			}
		}

		if 0 == len(respBody) {
			return WithHTTPCode(resp.StatusCode, errors.New("request '"+urlStr+"' fail: "+resp.Status+": read_error"))
		}
		return WithHTTPCode(resp.StatusCode, errors.New("request '"+urlStr+"' fail: "+resp.Status+": "+string(respBody)))
	}

	// Install closing the request body (if any)
	defer func() {
		if nil != resp.Body {
			io.Copy(ioutil.Discard, resp.Body)
			resp.Body.Close()
		}
	}()

	if nil == r.responseBody {
		return nil
	}

	switch response := r.responseBody.(type) {
	case ResponseFunc:
		return response(req, resp)
	case io.Writer:
		_, e = io.Copy(response, resp.Body)
		return e
	case *string:
		var sb strings.Builder
		if _, e = io.Copy(&sb, resp.Body); nil != e {
			return e
		}
		*response = sb.String()
		return nil
	case *[]byte:
		buffer := bytes.NewBuffer(make([]byte, 0, 1024))
		if _, e = io.Copy(buffer, resp.Body); nil != e {
			return e
		}
		*response = buffer.Bytes()
		return nil
	default:
		if r.jsonUseNumber {
			decoder := json.NewDecoder(resp.Body)
			decoder.UseNumber()
			return decoder.Decode(response)
		}

		buffer := BufferPool.Get().(*bytes.Buffer)
		_, e = io.Copy(buffer, resp.Body)
		if nil != e {
			buffer.Reset()
			BufferPool.Put(buffer)
			return e
		}

		e = json.Unmarshal(buffer.Bytes(), response)
		buffer.Reset()
		BufferPool.Put(buffer)
		return e
	}
}

// Join 拼接 url
func Join(paths ...string) string {
	switch len(paths) {
	case 0:
		return ""
	case 1:
		return paths[0]
	default:
		return JoinWith(paths[0], paths[1:])
	}
}

// JoinWith 拼接 url
func JoinWith(base string, paths []string) string {
	var buf bytes.Buffer
	buf.WriteString(base)

	lastSplash := strings.HasSuffix(base, "/")
	for _, pa := range paths {
		if 0 == len(pa) {
			continue
		}

		if lastSplash {
			if '/' == pa[0] {
				buf.WriteString(pa[1:])
			} else {
				buf.WriteString(pa)
			}
		} else {
			if '/' != pa[0] {
				buf.WriteString("/")
			}
			buf.WriteString(pa)
		}

		lastSplash = strings.HasSuffix(pa, "/")
	}
	return buf.String()
}

// func NewRequest(proxy Proxy, url string) Request {
//  return nil
// }

// func ReleaseRequest(proxy Proxy, request Request) {
// }

func NewRequest(proxy *Proxy, urlStr string) *Request {
	return proxy.New(urlStr)
}

func ReleaseRequest(proxy *Proxy, r *Request) {
	proxy.Release(r)
}
