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

	"github.com/runner-mei/loong/util"
)

var ErrBadArgument = util.ErrBadArgument
var WithHTTPCode = util.WithHTTPCode
var Wrap = util.Wrap

var BufferPool sync.Pool

func init() {
	BufferPool.New = func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, 1024))
	}
}

type HTTPError = util.HTTPError

type ResponseFunc func(req *http.Request, resp *http.Response) error

func New(urlStr string) (*Proxy, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	var queryParams url.Values
	for key, values := range u.Query() {
		queryParams[key] = values
	}
	u.RawQuery = ""

	return &Proxy{
		u:           *u,
		queryParams: queryParams,
	}, nil
}

type Proxy struct {
	Client        *http.Client
	jsonUseNumber bool
	u             url.URL
	queryParams   url.Values
	headers       url.Values
}

func (px *Proxy) JSONUseNumber() *Proxy {
	px.jsonUseNumber = true
	return px
}
func (px *Proxy) SetHeader(key, value string) *Proxy {
	px.headers.Set(key, value)
	return px
}
func (px *Proxy) AddHeader(key, value string) *Proxy {
	px.headers.Add(key, value)
	return px
}
func (px *Proxy) SetParam(key, value string) *Proxy {
	px.queryParams.Set(key, value)
	return px
}
func (px *Proxy) AddParam(key, value string) *Proxy {
	px.queryParams.Add(key, value)
	return px
}

func (proxy *Proxy) Release(request *Request) {}

func (proxy *Proxy) New(urlStr string) *Request {
	r := &Request{
		proxy:         proxy,
		jsonUseNumber: proxy.jsonUseNumber,
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
		} else {
			r.u.Path = Join(r.u.Path, u.Path)
		}

		for key, values := range u.Query() {
			r.queryParams[key] = values
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
func (r *Request) AddHeader(key, value string) *Request {
	r.headers.Add(key, value)
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
	r.responseBody = body
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
	if r.requestBody != nil && method != "GET" {
		switch value := r.requestBody.(type) {
		case []byte:
			body = bytes.NewReader(value)
		case string:
			body = strings.NewReader(value)
		case io.Reader:
			body = value
		default:
			buffer := BufferPool.Get().(*bytes.Buffer)
			e := json.NewEncoder(buffer).Encode(r.requestBody)
			if e != nil {
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

	client := r.proxy.Client
	if client == nil {
		client = http.DefaultClient
	}
	resp, e := client.Do(req)
	if e != nil {
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
		var responseBody string

		if nil != resp.Body {
			respBody, e := ioutil.ReadAll(resp.Body)
			resp.Body.Close()

			if e != nil {
				responseBody = string(respBody) + "\r\n*************** "
				responseBody += e.Error()
				responseBody += "***************"
			} else {
				responseBody = string(respBody)
			}
		}

		if len(responseBody) == 0 {
			return WithHTTPCode(resp.StatusCode, errors.New("request '"+urlStr+"' fail: "+resp.Status+": read_error"))
		}
		return WithHTTPCode(resp.StatusCode, errors.New("request '"+urlStr+"' fail: "+resp.Status+": "+responseBody))
	}

	// Install closing the request body (if any)
	defer func() {
		if nil != resp.Body {
			io.Copy(ioutil.Discard, resp.Body)
			resp.Body.Close()
		}
	}()

	if r.responseBody == nil {
		return nil
	}

	switch response := r.responseBody.(type) {
	case ResponseFunc:
		return response(req, resp)
	case io.Writer:
		_, e = io.Copy(response, resp.Body)
		return WithHTTPCode(11, e)
	case *string:
		var sb strings.Builder
		if _, e = io.Copy(&sb, resp.Body); e != nil {
			return WithHTTPCode(11, Wrap(e, "request '"+method+"' is ok and read response fail"))
		}
		*response = sb.String()
		return nil
	case *[]byte:
		buffer := bytes.NewBuffer(make([]byte, 0, 1024))
		if _, e = io.Copy(buffer, resp.Body); e != nil {
			return WithHTTPCode(11, Wrap(e, "request '"+method+"' is ok and read response fail"))
		}
		*response = buffer.Bytes()
		return nil
	default:
		if r.jsonUseNumber {
			decoder := json.NewDecoder(resp.Body)
			decoder.UseNumber()
			e = decoder.Decode(response)
			if e != nil {
				return WithHTTPCode(12, Wrap(e, "request '"+method+"' is ok and read response fail"))
			}
			return nil
		}

		buffer := BufferPool.Get().(*bytes.Buffer)
		_, e = io.Copy(buffer, resp.Body)
		if e != nil {
			buffer.Reset()
			BufferPool.Put(buffer)
			return WithHTTPCode(11, Wrap(e, "request '"+method+"' is ok and read response fail"))
		}

		e = json.Unmarshal(buffer.Bytes(), response)
		buffer.Reset()
		BufferPool.Put(buffer)
		if e != nil {
			return WithHTTPCode(12, Wrap(e, "request '"+method+"' is ok and read response fail"))
		}
		return nil
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
