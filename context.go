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
	"strconv"
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
	IP             string //方便外部其他方法设置IP
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
func (c *Context) HandlerName() string {
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

// GetIP get client ip address
func (c *Context) GetIP() (ip string) {
	addr := c.Req.RemoteAddr
	str := strings.Split(addr, ":")
	if len(str) > 1 {
		ip = str[0]
	}
	c.IP = ip
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
	if c.engine.RunMode == devMode {
		encoder.SetIndent("", "  ")
	}
	if err := encoder.Encode(data); err != nil {
		c.Fail(http.StatusServiceUnavailable, err.Error())
	}
}

// JSON response successful json format
func (c *Context) JSON(v interface{}) {
	c.ServerJSON(http.StatusOK, v)
}

// ServerXML response xml
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
	//未设置 AutoRender时，不渲染模板
	if !c.engine.AutoRender {
		c.ServerString(404, string(default404Body))
		return
	}
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

// Query Query
func (c *Context) Query(key string) string {
	return c.Req.URL.Query().Get(key)
}

// Form
func (c *Context) Form(key string) string {
	return c.Req.FormValue(key)
}

//input
func (c *Context) input() url.Values {
	if c.Req.Form == nil {
		c.Req.ParseForm()
	}
	return c.Req.Form
}

//formValue formValue
func (c *Context) formValue(key string) string {
	if v := c.Form(key); v != "" {
		return v
	}
	if c.Req.Form == nil {
		c.Req.ParseForm()
	}
	return c.Req.Form.Get(key)
}

//GetString 按key返回字串值，可以设置default值
func (c *Context) GetString(key string, def ...string) string {
	if v := c.formValue(key); v != "" {
		return v
	}
	if len(def) > 0 {
		return def[0]
	}
	return ""
}

//GetStrings GetStrings
func (c *Context) GetStrings(key string, def ...[]string) []string {
	var defaultDef []string
	if len(def) > 0 {
		defaultDef = def[0]
	}

	if v := c.input(); v == nil {
		return defaultDef
	} else if kv := v[key]; len(kv) > 0 {
		return kv
	}
	return defaultDef
}

//GetInt
func (c *Context) GetInt(key string, def ...int) (int, error) {
	v := c.formValue(key)
	if len(v) == 0 && len(def) > 0 {
		return def[0], nil
	}
	return strconv.Atoi(v)
}

//GetInt8 GetInt8
func (c *Context) GetInt8(key string, def ...int8) (int8, error) {
	v := c.formValue(key)
	if len(v) == 0 && len(def) > 0 {
		return def[0], nil
	}
	i64, err := strconv.ParseInt(v, 10, 8)
	return int8(i64), err
}

//GetUint8 GetUint8
func (c *Context) GetUint8(key string, def ...uint8) (uint8, error) {
	v := c.formValue(key)
	if len(v) == 0 && len(def) > 0 {
		return def[0], nil
	}
	i64, err := strconv.ParseUint(v, 10, 8)
	return uint8(i64), err
}

//GetInt16 GetInt16
func (c *Context) GetInt16(key string, def ...int16) (int16, error) {
	v := c.formValue(key)
	if len(v) == 0 && len(def) > 0 {
		return def[0], nil
	}
	i64, err := strconv.ParseInt(v, 10, 16)
	return int16(i64), err
}

//GetUint8 GetUint8
func (c *Context) GetUint16(key string, def ...uint16) (uint16, error) {
	v := c.formValue(key)
	if len(v) == 0 && len(def) > 0 {
		return def[0], nil
	}
	i64, err := strconv.ParseUint(v, 10, 16)
	return uint16(i64), err
}

//GetInt32 GetInt32
func (c *Context) GetInt32(key string, def ...int32) (int32, error) {
	v := c.formValue(key)
	if len(v) == 0 && len(def) > 0 {
		return def[0], nil
	}
	i64, err := strconv.ParseInt(v, 10, 32)
	return int32(i64), err
}

//GetUint32 GetUint32
func (c *Context) GetUint32(key string, def ...uint32) (uint32, error) {
	v := c.formValue(key)
	if len(v) == 0 && len(def) > 0 {
		return def[0], nil
	}
	i64, err := strconv.ParseUint(v, 10, 32)
	return uint32(i64), err
}

//GetInt64 GetInt64
func (c *Context) GetInt64(key string, def ...int64) (int64, error) {
	v := c.formValue(key)
	if len(v) == 0 && len(def) > 0 {
		return def[0], nil
	}
	return strconv.ParseInt(v, 10, 64)
}

//GetUint64 GetUint64
func (c *Context) GetUint64(key string, def ...uint64) (uint64, error) {
	v := c.formValue(key)
	if len(v) == 0 && len(def) > 0 {
		return def[0], nil
	}
	i64, err := strconv.ParseUint(v, 10, 64)
	return uint64(i64), err
}

//GetInt64 GetInt64
func (c *Context) GetFloat64(key string, def ...float64) (float64, error) {
	v := c.formValue(key)
	if len(v) == 0 && len(def) > 0 {
		return def[0], nil
	}
	return strconv.ParseFloat(v, 64)
}

//GetBool GetBool
func (c *Context) GetBool(key string, def ...bool) (bool, error) {
	v := c.formValue(key)
	if len(v) == 0 && len(def) > 0 {
		return def[0], nil
	}
	return strconv.ParseBool(v)
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
