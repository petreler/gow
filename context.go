package gow

import (
	"encoding/json"
	"encoding/xml"
	"github.com/gkzy/gow/render"
	"io"
	"io/ioutil"
	"math"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
)

//Context gow context
type Context struct {
	Writer         ResponseWriter
	responseWriter responseWriter
	Req            *http.Request
	handlers       HandlersChain
	Keys           map[string]interface{}
	Path           string
	Method         string
	Params         Params
	StatusCode     int

	index    int8
	engine   *Engine
	fullPath string
}

const (
	abortIndex = math.MaxInt8 / 2
)

func (c *Context) reset() {
	c.Writer = &c.responseWriter
	c.Params = c.Params[0:0]
	c.handlers = nil
	c.index = -1
	c.Keys = nil
}

func (c *Context) Next() {
	c.index++
	s := int8(len(c.handlers))
	for ; c.index < s; c.index++ {
		c.handlers[c.index](c)
	}
}

//HandlerName get handler name
func (c *Context) HandlerName() string{
	return nameOfFunction(c.handlers.Last())
}

//Abort abort http response
func (c *Context) AbortCode(statusCode int) {
	c.Writer.WriteHeader(statusCode)
	c.index = abortIndex
}


//Abort abort http response
func (c *Context) Abort() {
	c.AbortCode(http.StatusNonAuthoritativeInfo)
}

//Fail fail http response
func (c *Context) Fail(statusCode int, err string) {
	//TODO:日志记录
	c.AbortCode(statusCode)
	_, _ = c.Writer.Write([]byte(err))
}

// IP get client ip address
func (c *Context) IP() (ip string) {
	addr := c.Req.RemoteAddr
	str := strings.Split(addr, ":")
	if len(str) > 1 {
		ip = str[0]
	}
	return
}

//SetKey
func (c *Context) SetKey(key string, v interface{}) {
	if c.Keys == nil {
		c.Keys = make(map[string]interface{})
	}
	c.Keys[key] = v
}

//GetKey
func (c *Context) GetKey(key string) interface{} {
	var (
		ok bool
		v  interface{}
	)
	if c.Keys != nil {
		v, ok = c.Keys[key]
	} else {
		v, ok = nil, false
	}
	if !ok || v == nil {
		//TODO: record logs:
		//key %s does not exist
	}
	return v
}

// SetHeader set http response header
func (c *Context) SetHeader(key string, v string) {
	if v == "" {
		c.Writer.Header().Del(key)
	}
	c.Writer.Header().Set(key, v)
}

// GetHeader get http request header value by key
func (c *Context) GetHeader(key string) string {
	return c.Req.Header.Get(key)
}

//Status set http status code
func (c *Context) Status(statusCode int) {
	c.StatusCode = statusCode
	c.Writer.WriteHeader(statusCode)
}

// WriteString response text
func (c *Context) ServerString(statusCode int, msg string) {
	c.Writer.Header().Set("Content-Type", "text/plain;charset=utf-8")
	c.Status(statusCode)
	_, _ = c.Writer.Write([]byte(msg))
}

//String response text
func (c *Context) String(msg string) {
	c.ServerString(http.StatusOK, msg)
}

// ServerJSON response json format
func (c *Context) ServerJSON(statusCode int, data interface{}) {
	if statusCode < 0 {
		statusCode = http.StatusOK
	}
	c.SetHeader("Content-Type", "application/json; charset=utf-8")
	c.Status(statusCode)
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(data); err != nil {
		c.Fail(http.StatusServiceUnavailable, err.Error())
	}
}

// JSON response successful json format
func (c *Context) JSON(v interface{}) {
	c.ServerJSON(http.StatusOK, v)
}

func (c *Context) ServerXML(statusCode int, data interface{}) {
	if statusCode < 0 {
		statusCode = http.StatusOK
	}
	c.SetHeader("Content-Type", "application/xml; charset=utf-8")
	c.Status(statusCode)
	encoder := xml.NewEncoder(c.Writer)
	if err := encoder.Encode(data); err != nil {
		c.Fail(http.StatusServiceUnavailable, err.Error())
	}
}

//XML XML
func (c *Context) XML(data interface{}) {
	c.ServerXML(http.StatusOK, data)
}

// ServerHTML ServerHTML
func (c *Context) ServerHTML(statusCode int, name string, data interface{}) {
	c.SetHeader("Content-Type", "text/html; charset=utf-8")
	c.Status(statusCode)
	//if dev mode reBuilder template
	if c.engine.RunMode == devMode {
		render.BuildTemplate(c.engine.viewsPath, c.engine.FuncMap, c.engine.delims)
	}
	c.engine.HTMLRender = render.HTMLRender{}.Instance(name, data)
	err := c.engine.HTMLRender.Render(c.Writer)
	if err != nil {
		c.Fail(http.StatusServiceUnavailable, err.Error())
	}
}

//HTML
func (c *Context) HTML(name string, data interface{}) {
	c.ServerHTML(http.StatusOK, name, data)
}

//DecodeJSONBody
func (c *Context) DecodeJSONBody(v interface{}) error {
	decoder := json.NewDecoder(c.Req.Body)
	if err := decoder.Decode(&v); err != nil {
		return err
	}
	return nil
}

// RequestBody request body
func (c *Context) RequestBody() []byte {
	if c.Req.Body == nil {
		return []byte{}
	}
	var b []byte
	b, _ = ioutil.ReadAll(c.Req.Body)
	return b
}

//Param Param
func (c *Context) Param(name string) string {
	return c.Params.ByName(name)
}

//UserAgent
func (c *Context) UserAgent() string {
	return c.GetHeader("User-Agent")
}

//Query Query
func (c *Context) Query(key string) string {
	return c.Req.URL.Query().Get(key)
}

// Form
func (c *Context) Form(key string) string {
	return c.Req.FormValue(key)
}

// Redirect http redirect
// like : 301 302 ...
func (c *Context) Redirect(statusCode int, url string) {
	http.Redirect(c.Writer, c.Req, url, statusCode)
}

//Download
func (c *Context) Download(data []byte) {
	c.SetHeader("Content-Type", "application/octet-stream; charset=UTF-8")
	c.Status(200)
	c.Writer.Write(data)
}

//GetCookie get request cookie
func (c *Context) GetCookie(key string) string {
	ck, err := c.Req.Cookie(key)
	if err != nil {
		return ""
	}
	val, _ := url.QueryUnescape(ck.Value)
	return val
}

//SetCookie
func (c *Context) SetCookie(key, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	if path == "" {
		path = "/"
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     key,
		Value:    url.QueryEscape(value),
		MaxAge:   maxAge,
		Path:     path,
		Domain:   domain,
		SameSite: http.SameSiteDefaultMode,
		Secure:   secure,
		HttpOnly: httpOnly,
	})
}

//File read file to http body stream
func (c *Context) File(filePath string) {
	c.Status(http.StatusOK)
	http.ServeFile(c.Writer, c.Req, filePath)
}

//GetFile get single from form
func (c *Context) GetFile(key string) (multipart.File, *multipart.FileHeader, error) {
	return c.Req.FormFile(key)
}

//GetFiles get []file from form
func (c *Context) GetFiles(key string) ([]*multipart.FileHeader, error) {
	if files, ok := c.Req.MultipartForm.File[key]; ok {
		return files, nil
	}
	return nil, http.ErrMissingFile
}

// SaveToFile saves uploaded file to new path.
// on server
func (c *Context) SaveToFile(fromFile, toFile string) error {
	file, _, err := c.Req.FormFile(fromFile)
	if err != nil {
		return err
	}
	defer file.Close()
	f, err := os.OpenFile(toFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	io.Copy(f, file)
	return nil
}
