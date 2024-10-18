// Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package gin

import (
	"errors"
	"github.com/Superdanda/hade/framework"
	"io"
	"log"
	"math"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Superdanda/hade/framework/gin/binding"
	"github.com/Superdanda/hade/framework/gin/render"
	"github.com/gin-contrib/sse"
)

// Content-Type MIME of the most common data formats.
const (
	MIMEJSON              = binding.MIMEJSON
	MIMEHTML              = binding.MIMEHTML
	MIMEXML               = binding.MIMEXML
	MIMEXML2              = binding.MIMEXML2
	MIMEPlain             = binding.MIMEPlain
	MIMEPOSTForm          = binding.MIMEPOSTForm
	MIMEMultipartPOSTForm = binding.MIMEMultipartPOSTForm
	MIMEYAML              = binding.MIMEYAML
	MIMETOML              = binding.MIMETOML
)

// BodyBytesKey indicates a default body bytes key.
const BodyBytesKey = "_gin-gonic/gin/bodybyteskey"

// ContextKey is the key that a Context returns itself for.
const ContextKey = "_gin-gonic/gin/contextkey"

type ContextKeyType int

const ContextRequestKey ContextKeyType = 0

// abortIndex represents a typical value used in abort functions.
const abortIndex int8 = math.MaxInt8 >> 1

// Context is the most important part of gin. It allows us to pass variables between middleware,
// manage the flow, validate the JSON of a request and render a JSON response for example.
type Context struct {
	//新增 Hade框架的 容器
	container framework.Container

	writermem responseWriter
	Request   *http.Request
	Writer    ResponseWriter

	Params   Params
	handlers HandlersChain
	index    int8
	fullPath string

	engine       *Engine
	params       *Params
	skippedNodes *[]skippedNode

	// This mutex protects Keys map.
	mu sync.RWMutex

	// Keys is a key/value pair exclusively for the context of each request.
	Keys map[string]any

	// Errors is a list of errors attached to all the handlers/middlewares who used this context.
	Errors errorMsgs

	// Accepted defines a list of manually accepted formats for content negotiation.
	Accepted []string

	// queryCache caches the query result from c.Request.URL.Query().
	queryCache url.Values

	// formCache caches c.Request.PostForm, which contains the parsed form data from POST, PATCH,
	// or PUT body parameters.
	formCache url.Values

	// SameSite allows a server to define a cookie attribute making it impossible for
	// the browser to send this cookie along with cross-site requests.
	sameSite http.SameSite
}

/************************************/
/********** CONTEXT CREATION ********/
/************************************/

func (ctx *Context) reset() {
	ctx.Writer = &ctx.writermem
	ctx.Params = ctx.Params[:0]
	ctx.handlers = nil
	ctx.index = -1

	ctx.fullPath = ""
	ctx.Keys = nil
	ctx.Errors = ctx.Errors[:0]
	ctx.Accepted = nil
	ctx.queryCache = nil
	ctx.formCache = nil
	ctx.sameSite = 0
	*ctx.params = (*ctx.params)[:0]
	*ctx.skippedNodes = (*ctx.skippedNodes)[:0]
}

// Copy returns a copy of the current context that can be safely used outside the request's scope.
// This has to be used when the context has to be passed to a goroutine.
func (ctx *Context) Copy() *Context {
	cp := Context{
		writermem: ctx.writermem,
		Request:   ctx.Request,
		engine:    ctx.engine,
	}

	cp.writermem.ResponseWriter = nil
	cp.Writer = &cp.writermem
	cp.index = abortIndex
	cp.handlers = nil
	cp.fullPath = ctx.fullPath

	cKeys := ctx.Keys
	cp.Keys = make(map[string]any, len(cKeys))
	ctx.mu.RLock()
	for k, v := range cKeys {
		cp.Keys[k] = v
	}
	ctx.mu.RUnlock()

	cParams := ctx.Params
	cp.Params = make([]Param, len(cParams))
	copy(cp.Params, cParams)

	return &cp
}

// HandlerName returns the main handler's name. For example if the handler is "handleGetUsers()",
// this function will return "main.handleGetUsers".
func (ctx *Context) HandlerName() string {
	return nameOfFunction(ctx.handlers.Last())
}

// HandlerNames returns a list of all registered handlers for this context in descending order,
// following the semantics of HandlerName()
func (ctx *Context) HandlerNames() []string {
	hn := make([]string, 0, len(ctx.handlers))
	for _, val := range ctx.handlers {
		hn = append(hn, nameOfFunction(val))
	}
	return hn
}

// Handler returns the main handler.
func (ctx *Context) Handler() HandlerFunc {
	return ctx.handlers.Last()
}

// FullPath returns a matched route full path. For not found routes
// returns an empty string.
//
//	router.GET("/user/:id", func(c *gin.Context) {
//	    c.FullPath() == "/user/:id" // true
//	})
func (ctx *Context) FullPath() string {
	return ctx.fullPath
}

/************************************/
/*********** FLOW CONTROL ***********/
/************************************/

// Next should be used only inside middleware.
// It executes the pending handlers in the chain inside the calling handler.
// See example in GitHub.
func (ctx *Context) Next() {
	ctx.index++
	for ctx.index < int8(len(ctx.handlers)) {
		ctx.handlers[ctx.index](ctx)
		ctx.index++
	}
}

// IsAborted returns true if the current context was aborted.
func (ctx *Context) IsAborted() bool {
	return ctx.index >= abortIndex
}

// Abort prevents pending handlers from being called. Note that this will not stop the current handler.
// Let's say you have an authorization middleware that validates that the current request is authorized.
// If the authorization fails (ex: the password does not match), call Abort to ensure the remaining handlers
// for this request are not called.
func (ctx *Context) Abort() {
	ctx.index = abortIndex
}

// AbortWithStatus calls `Abort()` and writes the headers with the specified status code.
// For example, a failed attempt to authenticate a request could use: context.AbortWithStatus(401).
func (ctx *Context) AbortWithStatus(code int) {
	ctx.Status(code)
	ctx.Writer.WriteHeaderNow()
	ctx.Abort()
}

// AbortWithStatusJSON calls `Abort()` and then `JSON` internally.
// This method stops the chain, writes the status code and return a JSON body.
// It also sets the Content-Type as "application/json".
func (ctx *Context) AbortWithStatusJSON(code int, jsonObj any) {
	ctx.Abort()
	ctx.JSON(code, jsonObj)
}

// AbortWithError calls `AbortWithStatus()` and `Error()` internally.
// This method stops the chain, writes the status code and pushes the specified error to `c.Errors`.
// See Context.Error() for more details.
func (ctx *Context) AbortWithError(code int, err error) *Error {
	ctx.AbortWithStatus(code)
	return ctx.Error(err)
}

/************************************/
/********* ERROR MANAGEMENT *********/
/************************************/

// Error attaches an error to the current context. The error is pushed to a list of errors.
// It's a good idea to call Error for each error that occurred during the resolution of a request.
// A middleware can be used to collect all the errors and push them to a database together,
// print a log, or append it in the HTTP response.
// Error will panic if err is nil.
func (ctx *Context) Error(err error) *Error {
	if err == nil {
		panic("err is nil")
	}

	var parsedError *Error
	ok := errors.As(err, &parsedError)
	if !ok {
		parsedError = &Error{
			Err:  err,
			Type: ErrorTypePrivate,
		}
	}

	ctx.Errors = append(ctx.Errors, parsedError)
	return parsedError
}

/************************************/
/******** METADATA MANAGEMENT********/
/************************************/

// Set is used to store a new key/value pair exclusively for this context.
// It also lazy initializes  c.Keys if it was not used previously.
func (ctx *Context) Set(key string, value any) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	if ctx.Keys == nil {
		ctx.Keys = make(map[string]any)
	}

	ctx.Keys[key] = value
}

// Get returns the value for the given key, ie: (value, true).
// If the value does not exist it returns (nil, false)
func (ctx *Context) Get(key string) (value any, exists bool) {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	value, exists = ctx.Keys[key]
	return
}

// MustGet returns the value for the given key if it exists, otherwise it panics.
func (ctx *Context) MustGet(key string) any {
	if value, exists := ctx.Get(key); exists {
		return value
	}
	panic("Key \"" + key + "\" does not exist")
}

// GetString returns the value associated with the key as a string.
func (ctx *Context) GetString(key string) (s string) {
	if val, ok := ctx.Get(key); ok && val != nil {
		s, _ = val.(string)
	}
	return
}

// GetBool returns the value associated with the key as a boolean.
func (ctx *Context) GetBool(key string) (b bool) {
	if val, ok := ctx.Get(key); ok && val != nil {
		b, _ = val.(bool)
	}
	return
}

// GetInt returns the value associated with the key as an integer.
func (ctx *Context) GetInt(key string) (i int) {
	if val, ok := ctx.Get(key); ok && val != nil {
		i, _ = val.(int)
	}
	return
}

// GetInt64 returns the value associated with the key as an integer.
func (ctx *Context) GetInt64(key string) (i64 int64) {
	if val, ok := ctx.Get(key); ok && val != nil {
		i64, _ = val.(int64)
	}
	return
}

// GetUint returns the value associated with the key as an unsigned integer.
func (ctx *Context) GetUint(key string) (ui uint) {
	if val, ok := ctx.Get(key); ok && val != nil {
		ui, _ = val.(uint)
	}
	return
}

// GetUint64 returns the value associated with the key as an unsigned integer.
func (ctx *Context) GetUint64(key string) (ui64 uint64) {
	if val, ok := ctx.Get(key); ok && val != nil {
		ui64, _ = val.(uint64)
	}
	return
}

// GetFloat64 returns the value associated with the key as a float64.
func (ctx *Context) GetFloat64(key string) (f64 float64) {
	if val, ok := ctx.Get(key); ok && val != nil {
		f64, _ = val.(float64)
	}
	return
}

// GetTime returns the value associated with the key as time.
func (ctx *Context) GetTime(key string) (t time.Time) {
	if val, ok := ctx.Get(key); ok && val != nil {
		t, _ = val.(time.Time)
	}
	return
}

// GetDuration returns the value associated with the key as a duration.
func (ctx *Context) GetDuration(key string) (d time.Duration) {
	if val, ok := ctx.Get(key); ok && val != nil {
		d, _ = val.(time.Duration)
	}
	return
}

// GetStringSlice returns the value associated with the key as a slice of strings.
func (ctx *Context) GetStringSlice(key string) (ss []string) {
	if val, ok := ctx.Get(key); ok && val != nil {
		ss, _ = val.([]string)
	}
	return
}

// GetStringMap returns the value associated with the key as a map of interfaces.
func (ctx *Context) GetStringMap(key string) (sm map[string]any) {
	if val, ok := ctx.Get(key); ok && val != nil {
		sm, _ = val.(map[string]any)
	}
	return
}

// GetStringMapString returns the value associated with the key as a map of strings.
func (ctx *Context) GetStringMapString(key string) (sms map[string]string) {
	if val, ok := ctx.Get(key); ok && val != nil {
		sms, _ = val.(map[string]string)
	}
	return
}

// GetStringMapStringSlice returns the value associated with the key as a map to a slice of strings.
func (ctx *Context) GetStringMapStringSlice(key string) (smss map[string][]string) {
	if val, ok := ctx.Get(key); ok && val != nil {
		smss, _ = val.(map[string][]string)
	}
	return
}

/************************************/
/************ INPUT DATA ************/
/************************************/

// Param returns the value of the URL param.
// It is a shortcut for c.Params.ByName(key)
//
//	router.GET("/user/:id", func(c *gin.Context) {
//	    // a GET request to /user/john
//	    id := c.Param("id") // id == "john"
//	    // a GET request to /user/john/
//	    id := c.Param("id") // id == "/john/"
//	})
func (ctx *Context) Param(key string) string {
	return ctx.Params.ByName(key)
}

// AddParam adds param to context and
// replaces path param key with given value for e2e testing purposes
// Example Route: "/user/:id"
// AddParam("id", 1)
// Result: "/user/1"
func (ctx *Context) AddParam(key, value string) {
	ctx.Params = append(ctx.Params, Param{Key: key, Value: value})
}

// Query returns the keyed url query value if it exists,
// otherwise it returns an empty string `("")`.
// It is shortcut for `c.Request.URL.Query().Get(key)`
//
//	    GET /path?id=1234&name=Manu&value=
//		   c.Query("id") == "1234"
//		   c.Query("name") == "Manu"
//		   c.Query("value") == ""
//		   c.Query("wtf") == ""
func (ctx *Context) Query(key string) (value string) {
	value, _ = ctx.GetQuery(key)
	return
}

// DefaultQuery returns the keyed url query value if it exists,
// otherwise it returns the specified defaultValue string.
// See: Query() and GetQuery() for further information.
//
//	GET /?name=Manu&lastname=
//	c.DefaultQuery("name", "unknown") == "Manu"
//	c.DefaultQuery("id", "none") == "none"
//	c.DefaultQuery("lastname", "none") == ""
func (ctx *Context) DefaultQuery(key, defaultValue string) string {
	if value, ok := ctx.GetQuery(key); ok {
		return value
	}
	return defaultValue
}

// GetQuery is like Query(), it returns the keyed url query value
// if it exists `(value, true)` (even when the value is an empty string),
// otherwise it returns `("", false)`.
// It is shortcut for `c.Request.URL.Query().Get(key)`
//
//	GET /?name=Manu&lastname=
//	("Manu", true) == c.GetQuery("name")
//	("", false) == c.GetQuery("id")
//	("", true) == c.GetQuery("lastname")
func (ctx *Context) GetQuery(key string) (string, bool) {
	if values, ok := ctx.GetQueryArray(key); ok {
		return values[0], ok
	}
	return "", false
}

// QueryArray returns a slice of strings for a given query key.
// The length of the slice depends on the number of params with the given key.
func (ctx *Context) QueryArray(key string) (values []string) {
	values, _ = ctx.GetQueryArray(key)
	return
}

func (ctx *Context) initQueryCache() {
	if ctx.queryCache == nil {
		if ctx.Request != nil {
			ctx.queryCache = ctx.Request.URL.Query()
		} else {
			ctx.queryCache = url.Values{}
		}
	}
}

// GetQueryArray returns a slice of strings for a given query key, plus
// a boolean value whether at least one value exists for the given key.
func (ctx *Context) GetQueryArray(key string) (values []string, ok bool) {
	ctx.initQueryCache()
	values, ok = ctx.queryCache[key]
	return
}

// QueryMap returns a map for a given query key.
func (ctx *Context) QueryMap(key string) (dicts map[string]string) {
	dicts, _ = ctx.GetQueryMap(key)
	return
}

// GetQueryMap returns a map for a given query key, plus a boolean value
// whether at least one value exists for the given key.
func (ctx *Context) GetQueryMap(key string) (map[string]string, bool) {
	ctx.initQueryCache()
	return ctx.get(ctx.queryCache, key)
}

// PostForm returns the specified key from a POST urlencoded form or multipart form
// when it exists, otherwise it returns an empty string `("")`.
func (ctx *Context) PostForm(key string) (value string) {
	value, _ = ctx.GetPostForm(key)
	return
}

// DefaultPostForm returns the specified key from a POST urlencoded form or multipart form
// when it exists, otherwise it returns the specified defaultValue string.
// See: PostForm() and GetPostForm() for further information.
func (ctx *Context) DefaultPostForm(key, defaultValue string) string {
	if value, ok := ctx.GetPostForm(key); ok {
		return value
	}
	return defaultValue
}

// GetPostForm is like PostForm(key). It returns the specified key from a POST urlencoded
// form or multipart form when it exists `(value, true)` (even when the value is an empty string),
// otherwise it returns ("", false).
// For example, during a PATCH request to update the user's email:
//
//	    email=mail@example.com  -->  ("mail@example.com", true) := GetPostForm("email") // set email to "mail@example.com"
//		   email=                  -->  ("", true) := GetPostForm("email") // set email to ""
//	                            -->  ("", false) := GetPostForm("email") // do nothing with email
func (ctx *Context) GetPostForm(key string) (string, bool) {
	if values, ok := ctx.GetPostFormArray(key); ok {
		return values[0], ok
	}
	return "", false
}

// PostFormArray returns a slice of strings for a given form key.
// The length of the slice depends on the number of params with the given key.
func (ctx *Context) PostFormArray(key string) (values []string) {
	values, _ = ctx.GetPostFormArray(key)
	return
}

func (ctx *Context) initFormCache() {
	if ctx.formCache == nil {
		ctx.formCache = make(url.Values)
		req := ctx.Request
		if err := req.ParseMultipartForm(ctx.engine.MaxMultipartMemory); err != nil {
			if !errors.Is(err, http.ErrNotMultipart) {
				debugPrint("error on parse multipart form array: %v", err)
			}
		}
		ctx.formCache = req.PostForm
	}
}

// GetPostFormArray returns a slice of strings for a given form key, plus
// a boolean value whether at least one value exists for the given key.
func (ctx *Context) GetPostFormArray(key string) (values []string, ok bool) {
	ctx.initFormCache()
	values, ok = ctx.formCache[key]
	return
}

// PostFormMap returns a map for a given form key.
func (ctx *Context) PostFormMap(key string) (dicts map[string]string) {
	dicts, _ = ctx.GetPostFormMap(key)
	return
}

// GetPostFormMap returns a map for a given form key, plus a boolean value
// whether at least one value exists for the given key.
func (ctx *Context) GetPostFormMap(key string) (map[string]string, bool) {
	ctx.initFormCache()
	return ctx.get(ctx.formCache, key)
}

// get is an internal method and returns a map which satisfies conditions.
func (ctx *Context) get(m map[string][]string, key string) (map[string]string, bool) {
	dicts := make(map[string]string)
	exist := false
	for k, v := range m {
		if i := strings.IndexByte(k, '['); i >= 1 && k[0:i] == key {
			if j := strings.IndexByte(k[i+1:], ']'); j >= 1 {
				exist = true
				dicts[k[i+1:][:j]] = v[0]
			}
		}
	}
	return dicts, exist
}

// FormFile returns the first file for the provided form key.
func (ctx *Context) FormFile(name string) (*multipart.FileHeader, error) {
	if ctx.Request.MultipartForm == nil {
		if err := ctx.Request.ParseMultipartForm(ctx.engine.MaxMultipartMemory); err != nil {
			return nil, err
		}
	}
	f, fh, err := ctx.Request.FormFile(name)
	if err != nil {
		return nil, err
	}
	f.Close()
	return fh, err
}

// MultipartForm is the parsed multipart form, including file uploads.
func (ctx *Context) MultipartForm() (*multipart.Form, error) {
	err := ctx.Request.ParseMultipartForm(ctx.engine.MaxMultipartMemory)
	return ctx.Request.MultipartForm, err
}

// SaveUploadedFile uploads the form file to specific dst.
func (ctx *Context) SaveUploadedFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	if err = os.MkdirAll(filepath.Dir(dst), 0750); err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}

// Bind checks the Method and Content-Type to select a binding engine automatically,
// Depending on the "Content-Type" header different bindings are used, for example:
//
//	"application/json" --> JSON binding
//	"application/xml"  --> XML binding
//
// It parses the request's body as JSON if Content-Type == "application/json" using JSON or XML as a JSON input.
// It decodes the json payload into the struct specified as a pointer.
// It writes a 400 error and sets Content-Type header "text/plain" in the response if input is not valid.
func (ctx *Context) Bind(obj any) error {
	b := binding.Default(ctx.Request.Method, ctx.ContentType())
	return ctx.MustBindWith(obj, b)
}

// BindJSON is a shortcut for c.MustBindWith(obj, binding.JSON).
func (ctx *Context) BindJSON(obj any) error {
	return ctx.MustBindWith(obj, binding.JSON)
}

// BindXML is a shortcut for c.MustBindWith(obj, binding.BindXML).
func (ctx *Context) BindXML(obj any) error {
	return ctx.MustBindWith(obj, binding.XML)
}

// BindQuery is a shortcut for c.MustBindWith(obj, binding.Query).
func (ctx *Context) BindQuery(obj any) error {
	return ctx.MustBindWith(obj, binding.Query)
}

// BindYAML is a shortcut for c.MustBindWith(obj, binding.YAML).
func (ctx *Context) BindYAML(obj any) error {
	return ctx.MustBindWith(obj, binding.YAML)
}

// BindTOML is a shortcut for c.MustBindWith(obj, binding.TOML).
func (ctx *Context) BindTOML(obj any) error {
	return ctx.MustBindWith(obj, binding.TOML)
}

// BindHeader is a shortcut for c.MustBindWith(obj, binding.Header).
func (ctx *Context) BindHeader(obj any) error {
	return ctx.MustBindWith(obj, binding.Header)
}

// BindUri binds the passed struct pointer using binding.Uri.
// It will abort the request with HTTP 400 if any error occurs.
func (ctx *Context) BindUri(obj any) error {
	if err := ctx.ShouldBindUri(obj); err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err).SetType(ErrorTypeBind) //nolint: errcheck
		return err
	}
	return nil
}

// MustBindWith binds the passed struct pointer using the specified binding engine.
// It will abort the request with HTTP 400 if any error occurs.
// See the binding package.
func (ctx *Context) MustBindWith(obj any, b binding.Binding) error {
	if err := ctx.ShouldBindWith(obj, b); err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err).SetType(ErrorTypeBind) //nolint: errcheck
		return err
	}
	return nil
}

// ShouldBind checks the Method and Content-Type to select a binding engine automatically,
// Depending on the "Content-Type" header different bindings are used, for example:
//
//	"application/json" --> JSON binding
//	"application/xml"  --> XML binding
//
// It parses the request's body as JSON if Content-Type == "application/json" using JSON or XML as a JSON input.
// It decodes the json payload into the struct specified as a pointer.
// Like c.Bind() but this method does not set the response status code to 400 or abort if input is not valid.
func (ctx *Context) ShouldBind(obj any) error {
	b := binding.Default(ctx.Request.Method, ctx.ContentType())
	return ctx.ShouldBindWith(obj, b)
}

// ShouldBindJSON is a shortcut for c.ShouldBindWith(obj, binding.JSON).
func (ctx *Context) ShouldBindJSON(obj any) error {
	return ctx.ShouldBindWith(obj, binding.JSON)
}

// ShouldBindXML is a shortcut for c.ShouldBindWith(obj, binding.XML).
func (ctx *Context) ShouldBindXML(obj any) error {
	return ctx.ShouldBindWith(obj, binding.XML)
}

// ShouldBindQuery is a shortcut for c.ShouldBindWith(obj, binding.Query).
func (ctx *Context) ShouldBindQuery(obj any) error {
	return ctx.ShouldBindWith(obj, binding.Query)
}

// ShouldBindYAML is a shortcut for c.ShouldBindWith(obj, binding.YAML).
func (ctx *Context) ShouldBindYAML(obj any) error {
	return ctx.ShouldBindWith(obj, binding.YAML)
}

// ShouldBindTOML is a shortcut for c.ShouldBindWith(obj, binding.TOML).
func (ctx *Context) ShouldBindTOML(obj any) error {
	return ctx.ShouldBindWith(obj, binding.TOML)
}

// ShouldBindHeader is a shortcut for c.ShouldBindWith(obj, binding.Header).
func (ctx *Context) ShouldBindHeader(obj any) error {
	return ctx.ShouldBindWith(obj, binding.Header)
}

// ShouldBindUri binds the passed struct pointer using the specified binding engine.
func (ctx *Context) ShouldBindUri(obj any) error {
	m := make(map[string][]string, len(ctx.Params))
	for _, v := range ctx.Params {
		m[v.Key] = []string{v.Value}
	}
	return binding.Uri.BindUri(m, obj)
}

// ShouldBindWith binds the passed struct pointer using the specified binding engine.
// See the binding package.
func (ctx *Context) ShouldBindWith(obj any, b binding.Binding) error {
	return b.Bind(ctx.Request, obj)
}

// ShouldBindBodyWith is similar with ShouldBindWith, but it stores the request
// body into the context, and reuse when it is called again.
//
// NOTE: This method reads the body before binding. So you should use
// ShouldBindWith for better performance if you need to call only once.
func (ctx *Context) ShouldBindBodyWith(obj any, bb binding.BindingBody) (err error) {
	var body []byte
	if cb, ok := ctx.Get(BodyBytesKey); ok {
		if cbb, ok := cb.([]byte); ok {
			body = cbb
		}
	}
	if body == nil {
		body, err = io.ReadAll(ctx.Request.Body)
		if err != nil {
			return err
		}
		ctx.Set(BodyBytesKey, body)
	}
	return bb.BindBody(body, obj)
}

// ShouldBindBodyWithJSON is a shortcut for c.ShouldBindBodyWith(obj, binding.JSON).
func (ctx *Context) ShouldBindBodyWithJSON(obj any) error {
	return ctx.ShouldBindBodyWith(obj, binding.JSON)
}

// ShouldBindBodyWithXML is a shortcut for c.ShouldBindBodyWith(obj, binding.XML).
func (ctx *Context) ShouldBindBodyWithXML(obj any) error {
	return ctx.ShouldBindBodyWith(obj, binding.XML)
}

// ShouldBindBodyWithYAML is a shortcut for c.ShouldBindBodyWith(obj, binding.YAML).
func (ctx *Context) ShouldBindBodyWithYAML(obj any) error {
	return ctx.ShouldBindBodyWith(obj, binding.YAML)
}

// ShouldBindBodyWithTOML is a shortcut for c.ShouldBindBodyWith(obj, binding.TOML).
func (ctx *Context) ShouldBindBodyWithTOML(obj any) error {
	return ctx.ShouldBindBodyWith(obj, binding.TOML)
}

// ClientIP implements one best effort algorithm to return the real client IP.
// It calls c.RemoteIP() under the hood, to check if the remote IP is a trusted proxy or not.
// If it is it will then try to parse the headers defined in Engine.RemoteIPHeaders (defaulting to [X-Forwarded-For, X-Real-Ip]).
// If the headers are not syntactically valid OR the remote IP does not correspond to a trusted proxy,
// the remote IP (coming from Request.RemoteAddr) is returned.
func (ctx *Context) ClientIP() string {
	// Check if we're running on a trusted platform, continue running backwards if error
	if ctx.engine.TrustedPlatform != "" {
		// Developers can define their own header of Trusted Platform or use predefined constants
		if addr := ctx.requestHeader(ctx.engine.TrustedPlatform); addr != "" {
			return addr
		}
	}

	// Legacy "AppEngine" flag
	if ctx.engine.AppEngine {
		log.Println(`The AppEngine flag is going to be deprecated. Please check issues #2723 and #2739 and use 'TrustedPlatform: gin.PlatformGoogleAppEngine' instead.`)
		if addr := ctx.requestHeader("X-Appengine-Remote-Addr"); addr != "" {
			return addr
		}
	}

	// It also checks if the remoteIP is a trusted proxy or not.
	// In order to perform this validation, it will see if the IP is contained within at least one of the CIDR blocks
	// defined by Engine.SetTrustedProxies()
	remoteIP := net.ParseIP(ctx.RemoteIP())
	if remoteIP == nil {
		return ""
	}
	trusted := ctx.engine.isTrustedProxy(remoteIP)

	if trusted && ctx.engine.ForwardedByClientIP && ctx.engine.RemoteIPHeaders != nil {
		for _, headerName := range ctx.engine.RemoteIPHeaders {
			ip, valid := ctx.engine.validateHeader(ctx.requestHeader(headerName))
			if valid {
				return ip
			}
		}
	}
	return remoteIP.String()
}

// RemoteIP parses the IP from Request.RemoteAddr, normalizes and returns the IP (without the port).
func (ctx *Context) RemoteIP() string {
	ip, _, err := net.SplitHostPort(strings.TrimSpace(ctx.Request.RemoteAddr))
	if err != nil {
		return ""
	}
	return ip
}

// ContentType returns the Content-Type header of the request.
func (ctx *Context) ContentType() string {
	return filterFlags(ctx.requestHeader("Content-Type"))
}

// IsWebsocket returns true if the request headers indicate that a websocket
// handshake is being initiated by the client.
func (ctx *Context) IsWebsocket() bool {
	if strings.Contains(strings.ToLower(ctx.requestHeader("Connection")), "upgrade") &&
		strings.EqualFold(ctx.requestHeader("Upgrade"), "websocket") {
		return true
	}
	return false
}

func (ctx *Context) requestHeader(key string) string {
	return ctx.Request.Header.Get(key)
}

/************************************/
/******** RESPONSE RENDERING ********/
/************************************/

// bodyAllowedForStatus is a copy of http.bodyAllowedForStatus non-exported function.
func bodyAllowedForStatus(status int) bool {
	switch {
	case status >= 100 && status <= 199:
		return false
	case status == http.StatusNoContent:
		return false
	case status == http.StatusNotModified:
		return false
	}
	return true
}

// Status sets the HTTP response code.
func (ctx *Context) Status(code int) {
	ctx.Writer.WriteHeader(code)
}

// Header is an intelligent shortcut for c.Writer.Header().Set(key, value).
// It writes a header in the response.
// If value == "", this method removes the header `c.Writer.Header().Del(key)`
func (ctx *Context) Header(key, value string) {
	if value == "" {
		ctx.Writer.Header().Del(key)
		return
	}
	ctx.Writer.Header().Set(key, value)
}

// GetHeader returns value from request headers.
func (ctx *Context) GetHeader(key string) string {
	return ctx.requestHeader(key)
}

// GetRawData returns stream data.
func (ctx *Context) GetRawData() ([]byte, error) {
	if ctx.Request.Body == nil {
		return nil, errors.New("cannot read nil body")
	}
	return io.ReadAll(ctx.Request.Body)
}

// SetSameSite with cookie
func (ctx *Context) SetSameSite(samesite http.SameSite) {
	ctx.sameSite = samesite
}

// SetCookie adds a Set-Cookie header to the ResponseWriter's headers.
// The provided cookie must have a valid Name. Invalid cookies may be
// silently dropped.
func (ctx *Context) SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	if path == "" {
		path = "/"
	}
	http.SetCookie(ctx.Writer, &http.Cookie{
		Name:     name,
		Value:    url.QueryEscape(value),
		MaxAge:   maxAge,
		Path:     path,
		Domain:   domain,
		SameSite: ctx.sameSite,
		Secure:   secure,
		HttpOnly: httpOnly,
	})
}

// Cookie returns the named cookie provided in the request or
// ErrNoCookie if not found. And return the named cookie is unescaped.
// If multiple cookies match the given name, only one cookie will
// be returned.
func (ctx *Context) Cookie(name string) (string, error) {
	cookie, err := ctx.Request.Cookie(name)
	if err != nil {
		return "", err
	}
	val, _ := url.QueryUnescape(cookie.Value)
	return val, nil
}

// Render writes the response headers and calls render.Render to render data.
func (ctx *Context) Render(code int, r render.Render) {
	ctx.Status(code)

	if !bodyAllowedForStatus(code) {
		r.WriteContentType(ctx.Writer)
		ctx.Writer.WriteHeaderNow()
		return
	}

	if err := r.Render(ctx.Writer); err != nil {
		// Pushing error to c.Errors
		_ = ctx.Error(err)
		ctx.Abort()
	}
}

// HTML renders the HTTP template specified by its file name.
// It also updates the HTTP code and sets the Content-Type as "text/html".
// See http://golang.org/doc/articles/wiki/
func (ctx *Context) HTML(code int, name string, obj any) {
	instance := ctx.engine.HTMLRender.Instance(name, obj)
	ctx.Render(code, instance)
}

// IndentedJSON serializes the given struct as pretty JSON (indented + endlines) into the response body.
// It also sets the Content-Type as "application/json".
// WARNING: we recommend using this only for development purposes since printing pretty JSON is
// more CPU and bandwidth consuming. Use Context.JSON() instead.
func (ctx *Context) IndentedJSON(code int, obj any) {
	ctx.Render(code, render.IndentedJSON{Data: obj})
}

// SecureJSON serializes the given struct as Secure JSON into the response body.
// Default prepends "while(1)," to response body if the given struct is array values.
// It also sets the Content-Type as "application/json".
func (ctx *Context) SecureJSON(code int, obj any) {
	ctx.Render(code, render.SecureJSON{Prefix: ctx.engine.secureJSONPrefix, Data: obj})
}

// JSONP serializes the given struct as JSON into the response body.
// It adds padding to response body to request data from a server residing in a different domain than the client.
// It also sets the Content-Type as "application/javascript".
func (ctx *Context) JSONP(code int, obj any) {
	callback := ctx.DefaultQuery("callback", "")
	if callback == "" {
		ctx.Render(code, render.JSON{Data: obj})
		return
	}
	ctx.Render(code, render.JsonpJSON{Callback: callback, Data: obj})
}

// JSON serializes the given struct as JSON into the response body.
// It also sets the Content-Type as "application/json".
func (ctx *Context) JSON(code int, obj any) {
	ctx.Render(code, render.JSON{Data: obj})
}

// AsciiJSON serializes the given struct as JSON into the response body with unicode to ASCII string.
// It also sets the Content-Type as "application/json".
func (ctx *Context) AsciiJSON(code int, obj any) {
	ctx.Render(code, render.AsciiJSON{Data: obj})
}

// PureJSON serializes the given struct as JSON into the response body.
// PureJSON, unlike JSON, does not replace special html characters with their unicode entities.
func (ctx *Context) PureJSON(code int, obj any) {
	ctx.Render(code, render.PureJSON{Data: obj})
}

// XML serializes the given struct as XML into the response body.
// It also sets the Content-Type as "application/xml".
func (ctx *Context) XML(code int, obj any) {
	ctx.Render(code, render.XML{Data: obj})
}

// YAML serializes the given struct as YAML into the response body.
func (ctx *Context) YAML(code int, obj any) {
	ctx.Render(code, render.YAML{Data: obj})
}

// TOML serializes the given struct as TOML into the response body.
func (ctx *Context) TOML(code int, obj any) {
	ctx.Render(code, render.TOML{Data: obj})
}

// ProtoBuf serializes the given struct as ProtoBuf into the response body.
func (ctx *Context) ProtoBuf(code int, obj any) {
	ctx.Render(code, render.ProtoBuf{Data: obj})
}

// String writes the given string into the response body.
func (ctx *Context) String(code int, format string, values ...any) {
	ctx.Render(code, render.String{Format: format, Data: values})
}

// Redirect returns an HTTP redirect to the specific location.
func (ctx *Context) Redirect(code int, location string) {
	ctx.Render(-1, render.Redirect{
		Code:     code,
		Location: location,
		Request:  ctx.Request,
	})
}

// Data writes some data into the body stream and updates the HTTP code.
func (ctx *Context) Data(code int, contentType string, data []byte) {
	ctx.Render(code, render.Data{
		ContentType: contentType,
		Data:        data,
	})
}

// DataFromReader writes the specified reader into the body stream and updates the HTTP code.
func (ctx *Context) DataFromReader(code int, contentLength int64, contentType string, reader io.Reader, extraHeaders map[string]string) {
	ctx.Render(code, render.Reader{
		Headers:       extraHeaders,
		ContentType:   contentType,
		ContentLength: contentLength,
		Reader:        reader,
	})
}

// File writes the specified file into the body stream in an efficient way.
func (ctx *Context) File(filepath string) {
	http.ServeFile(ctx.Writer, ctx.Request, filepath)
}

// FileFromFS writes the specified file from http.FileSystem into the body stream in an efficient way.
func (ctx *Context) FileFromFS(filepath string, fs http.FileSystem) {
	defer func(old string) {
		ctx.Request.URL.Path = old
	}(ctx.Request.URL.Path)

	ctx.Request.URL.Path = filepath

	http.FileServer(fs).ServeHTTP(ctx.Writer, ctx.Request)
}

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

// FileAttachment writes the specified file into the body stream in an efficient way
// On the client side, the file will typically be downloaded with the given filename
func (ctx *Context) FileAttachment(filepath, filename string) {
	if isASCII(filename) {
		ctx.Writer.Header().Set("Content-Disposition", `attachment; filename="`+escapeQuotes(filename)+`"`)
	} else {
		ctx.Writer.Header().Set("Content-Disposition", `attachment; filename*=UTF-8''`+url.QueryEscape(filename))
	}
	http.ServeFile(ctx.Writer, ctx.Request, filepath)
}

// SSEvent writes a Server-Sent Event into the body stream.
func (ctx *Context) SSEvent(name string, message any) {
	ctx.Render(-1, sse.Event{
		Event: name,
		Data:  message,
	})
}

// Stream sends a streaming response and returns a boolean
// indicates "Is client disconnected in middle of stream"
func (ctx *Context) Stream(step func(w io.Writer) bool) bool {
	w := ctx.Writer
	clientGone := w.CloseNotify()
	for {
		select {
		case <-clientGone:
			return true
		default:
			keepOpen := step(w)
			w.Flush()
			if !keepOpen {
				return false
			}
		}
	}
}

/************************************/
/******** CONTENT NEGOTIATION *******/
/************************************/

// Negotiate contains all negotiations data.
type Negotiate struct {
	Offered  []string
	HTMLName string
	HTMLData any
	JSONData any
	XMLData  any
	YAMLData any
	Data     any
	TOMLData any
}

// Negotiate calls different Render according to acceptable Accept format.
func (ctx *Context) Negotiate(code int, config Negotiate) {
	switch ctx.NegotiateFormat(config.Offered...) {
	case binding.MIMEJSON:
		data := chooseData(config.JSONData, config.Data)
		ctx.JSON(code, data)

	case binding.MIMEHTML:
		data := chooseData(config.HTMLData, config.Data)
		ctx.HTML(code, config.HTMLName, data)

	case binding.MIMEXML:
		data := chooseData(config.XMLData, config.Data)
		ctx.XML(code, data)

	case binding.MIMEYAML:
		data := chooseData(config.YAMLData, config.Data)
		ctx.YAML(code, data)

	case binding.MIMETOML:
		data := chooseData(config.TOMLData, config.Data)
		ctx.TOML(code, data)

	default:
		ctx.AbortWithError(http.StatusNotAcceptable, errors.New("the accepted formats are not offered by the server")) //nolint: errcheck
	}
}

// NegotiateFormat returns an acceptable Accept format.
func (ctx *Context) NegotiateFormat(offered ...string) string {
	assert1(len(offered) > 0, "you must provide at least one offer")

	if ctx.Accepted == nil {
		ctx.Accepted = parseAccept(ctx.requestHeader("Accept"))
	}
	if len(ctx.Accepted) == 0 {
		return offered[0]
	}
	for _, accepted := range ctx.Accepted {
		for _, offer := range offered {
			// According to RFC 2616 and RFC 2396, non-ASCII characters are not allowed in headers,
			// therefore we can just iterate over the string without casting it into []rune
			i := 0
			for ; i < len(accepted) && i < len(offer); i++ {
				if accepted[i] == '*' || offer[i] == '*' {
					return offer
				}
				if accepted[i] != offer[i] {
					break
				}
			}
			if i == len(accepted) {
				return offer
			}
		}
	}
	return ""
}

// SetAccepted sets Accept header data.
func (ctx *Context) SetAccepted(formats ...string) {
	ctx.Accepted = formats
}

/************************************/
/***** GOLANG.ORG/X/NET/CONTEXT *****/
/************************************/

// hasRequestContext returns whether c.Request has Context and fallback.
func (ctx *Context) hasRequestContext() bool {
	hasFallback := ctx.engine != nil && ctx.engine.ContextWithFallback
	hasRequestContext := ctx.Request != nil && ctx.Request.Context() != nil
	return hasFallback && hasRequestContext
}

// Deadline returns that there is no deadline (ok==false) when c.Request has no Context.
func (ctx *Context) Deadline() (deadline time.Time, ok bool) {
	if !ctx.hasRequestContext() {
		return
	}
	return ctx.Request.Context().Deadline()
}

// Done returns nil (chan which will wait forever) when c.Request has no Context.
func (ctx *Context) Done() <-chan struct{} {
	if !ctx.hasRequestContext() {
		return nil
	}
	return ctx.Request.Context().Done()
}

// Err returns nil when c.Request has no Context.
func (ctx *Context) Err() error {
	if !ctx.hasRequestContext() {
		return nil
	}
	return ctx.Request.Context().Err()
}

// Value returns the value associated with this context for key, or nil
// if no value is associated with key. Successive calls to Value with
// the same key returns the same result.
func (ctx *Context) Value(key any) any {
	if key == ContextRequestKey {
		return ctx.Request
	}
	if key == ContextKey {
		return ctx
	}
	if keyAsString, ok := key.(string); ok {
		if val, exists := ctx.Get(keyAsString); exists {
			return val
		}
	}
	if !ctx.hasRequestContext() {
		return nil
	}
	return ctx.Request.Context().Value(key)
}
