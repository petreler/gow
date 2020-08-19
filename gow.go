package gow

import (
	"fmt"
	"github.com/gkzy/gow/render"
	"html/template"
	"net/http"
	"path"
	"strings"
	"sync"
)

var (
	default404Body = []byte("404 page not found")
	default405Body = []byte("405 method not allowed")
)

//HandlerFunc handler func
type HandlerFunc func(*Context)

type HandlersChain []HandlerFunc

//Last last handler
func (c HandlersChain) Last() HandlerFunc {
	if length := len(c); length > 0 {
		return c[length-1]
	}
	return nil
}

// RouteInfo represents a request route's specification which contains method and path and its handler.
type RouteInfo struct {
	Method      string
	Path        string
	Handler     string
	HandlerFunc HandlerFunc
}

// RoutesInfo defines a RouteInfo array.
type RoutesInfo []RouteInfo

const (
	defaultMode            = "dev"
	devMode                = "dev"
	prodMode               = "prod"
	defaultViews           = "views"
	defaultStatic          = "static"
	defaultMultipartMemory = 32 << 20
)

type Engine struct {
	AppName string
	RunMode string
	AppPath string //程序的运行地址
	*RouterGroup

	//template
	HTMLRender render.Render
	FuncMap    template.FuncMap
	delims     render.Delims
	AutoRender bool //是否渲染模板

	HandleMethodNotAllowed bool

	UseRawPath            bool
	UnescapePathValues    bool
	RemoveExtraSlash      bool
	RedirectTrailingSlash bool
	RedirectFixedPath     bool
	MaxMultipartMemory    int64

	//views template directory
	viewsPath   string
	staticPath  string
	httpAddr    string
	trees       methodTrees
	allNoRoute  HandlersChain
	allNoMethod HandlersChain
	noRoute     HandlersChain
	noMethod    HandlersChain
	pool        sync.Pool

	// session switch
	SessionOn bool
}

func New() *Engine {
	engine := &Engine{
		RouterGroup: &RouterGroup{
			Handlers: nil,
			basePath: "/",
			root:     true,
		},
		FuncMap:                template.FuncMap{},
		delims:                 render.Delims{Left: "{{", Right: "}}"},
		AutoRender:             false,
		RedirectTrailingSlash:  true,
		RedirectFixedPath:      false,
		HandleMethodNotAllowed: false,
		AppName:                "gow",
		UseRawPath:             false,
		RemoveExtraSlash:       false,
		UnescapePathValues:     true,
		MaxMultipartMemory:     defaultMultipartMemory,
		trees:                  make(methodTrees, 0, 9),
		viewsPath:              defaultViews,
		staticPath:             defaultStatic,
		httpAddr:               ":8080", //default http Addr
		RunMode:                defaultMode,
		AppPath:                getCurrentDirectory(),
	}
	engine.RouterGroup.engine = engine
	engine.pool.New = func() interface{} {
		ctx := &Context{engine: engine}
		return ctx
	}

	return engine
}

// Default get default engine
//	use Recovery()
//	use Logger()
func Default() *Engine {
	engine := New()
	engine.Use(Recovery())
	engine.Use(Logger())
	return engine
}

//Use use middleware
func (engine *Engine) Use(middleware ...HandlerFunc) {
	engine.RouterGroup.Use(middleware...)
	engine.engine.rebuild404Handlers()
	engine.engine.rebuild405Handlers()
}

// ServeHTTP implement the http.handler interface
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := engine.pool.Get().(*Context)
	c.responseWriter.reset(w)
	c.Req = req
	c.reset()
	c.Data = make(map[interface{}]interface{},0)
	engine.handleHTTPRequest(c)
	engine.pool.Put(c)
}

// SetAppConfig 统一配置入口
// 可使用此方法统一配置，也可以使用其他方法单独设置
func (engine *Engine) SetAppConfig(app *AppConfig) {
	if app != nil {
		debugPrint("[%s] Load the configuration using the SetAppConfig method", app.AppName)
		engine.AppName = app.AppName
		engine.RunMode = app.RunMode
		engine.viewsPath = app.Views
		engine.delims = render.Delims{Left: app.TemplateLeft, Right: app.TemplateRight}
		engine.AutoRender = app.AutoRender
		engine.httpAddr = app.HttpAddr
	}
}

// Run
func (engine *Engine) Run(addr ...string) (err error) {
	defer func() {
		debugPrintError(err)
	}()

	if engine.AutoRender {
		//builder template
		err = render.AddViewPath(engine.viewsPath)
	}

	address := engine.resolveAddress(addr)

	if engine.RunMode == devMode {
		fmt.Println(logo)
		debugPrint("package: %s", pkg)
		debugPrint("website: %s", site)
	}

	debugPrint("[%s] [%s] Listening and serving HTTP on %s", engine.AppName, engine.RunMode, address)
	err = http.ListenAndServe(address, engine)
	return
}

// RunTLS
func (engine *Engine) RunTLS(certFile, keyFile string, addr ...string) (err error) {
	defer func() {
		debugPrintError(err)
	}()

	if engine.AutoRender {
		//builder template
		err = render.AddViewPath(engine.viewsPath)
	}
	address := engine.resolveAddress(addr)
	if engine.RunMode == devMode {
		fmt.Println(logo)
		debugPrint("package: %s", pkg)
		debugPrint("website: %s", site)
	}
	debugPrint("[%s] [%s] Listening and serving HTTP on %s", engine.AppName, engine.RunMode, address)
	err = http.ListenAndServeTLS(address, certFile, keyFile, engine)
	return
}

// SetSessionOn SetSessionOn
func (engine *Engine) SetSessionOn(on bool) {
	engine.SessionOn = on
}

// NoRoute adds handlers for NoRoute. It return a 404 code by default.
func (engine *Engine) NoRoute(handlers ...HandlerFunc) {
	engine.noRoute = handlers
	engine.rebuild404Handlers()
}

// NoMethod sets the handlers called when...
// TODO:
func (engine *Engine) NoMethod(handlers ...HandlerFunc) {
	engine.noMethod = handlers
	engine.rebuild405Handlers()
}

func (engine *Engine) rebuild404Handlers() {
	engine.allNoRoute = engine.combineHandlers(engine.noRoute)
}

func (engine *Engine) rebuild405Handlers() {
	engine.allNoMethod = engine.combineHandlers(engine.noMethod)
}

// AddFuncMap add fn func to template func map
func (engine *Engine) AddFuncMap(key string, fn interface{}) {
	engine.FuncMap[key] = fn
}

// Delims set Delims
func (engine *Engine) Delims(left, right string) {
	engine.delims = render.Delims{Left: left, Right: right}
}

// SetView set views path
// 模板目录为 views 时，可不用设置此值
func (engine *Engine) SetView(path ...string) {
	dir := defaultViews
	if len(path) > 0 {
		dir = path[0]
	}
	engine.viewsPath = dir
}

// RoutesMap get all router map
func (engine *Engine) RouterMap() (routes RoutesInfo) {
	for _, tree := range engine.trees {
		routes = iterate("", tree.method, routes, tree.root)
	}
	return routes
}

// ==========================private func=======================

func iterate(path, method string, routes RoutesInfo, root *node) RoutesInfo {
	path += root.path
	if len(root.handlers) > 0 {
		handlerFunc := root.handlers.Last()
		routes = append(routes, RouteInfo{
			Method:      method,
			Path:        path,
			Handler:     nameOfFunction(handlerFunc),
			HandlerFunc: handlerFunc,
		})
	}
	for _, child := range root.children {
		routes = iterate(path, method, routes, child)
	}
	return routes
}

func (engine *Engine) handleHTTPRequest(c *Context) {
	httpMethod := c.Req.Method
	rPath := c.Req.URL.Path
	unescape := false
	if engine.UseRawPath && len(c.Req.URL.RawPath) > 0 {
		rPath = c.Req.URL.RawPath
		unescape = engine.UnescapePathValues
	}

	if engine.RemoveExtraSlash {
		rPath = cleanPath(rPath)
	}

	// Find root of the tree for the given HTTP method
	t := engine.trees
	for i, tl := 0, len(t); i < tl; i++ {
		if t[i].method != httpMethod {
			continue
		}
		root := t[i].root
		// Find route in tree
		value := root.getValue(rPath, c.Params, unescape)
		if value.handlers != nil {
			c.handlers = value.handlers
			c.Params = value.params
			c.fullPath = strings.ToLower(value.fullPath)
			c.Method = httpMethod
			c.Path = rPath
			c.Next()
			c.responseWriter.WriteHeaderNow()
			return
		}
		if httpMethod != "CONNECT" && rPath != "/" {
			if value.tsr && engine.RedirectTrailingSlash {
				redirectTrailingSlash(c)
				return
			}
			if engine.RedirectFixedPath && redirectFixedPath(c, root, engine.RedirectFixedPath) {
				return
			}
		}
		break
	}

	if engine.HandleMethodNotAllowed {
		for _, tree := range engine.trees {
			if tree.method == httpMethod {
				continue
			}
			if value := tree.root.getValue(rPath, nil, unescape); value.handlers != nil {
				c.handlers = engine.allNoMethod
				serveError(c, http.StatusMethodNotAllowed, default405Body)
				return
			}
		}
	}
	c.handlers = engine.allNoRoute
	serveError(c, http.StatusNotFound, default404Body)
}

func redirectTrailingSlash(c *Context) {
	req := c.Req
	p := req.URL.Path
	if prefix := path.Clean(c.Req.Header.Get("X-Forwarded-Prefix")); prefix != "." {
		p = prefix + "/" + req.URL.Path
	}
	req.URL.Path = p + "/"
	if length := len(p); length > 1 && p[length-1] == '/' {
		req.URL.Path = p[:length-1]
	}
	redirectRequest(c)
}

func redirectFixedPath(c *Context, root *node, trailingSlash bool) bool {
	req := c.Req
	rPath := req.URL.Path

	if fixedPath, ok := root.findCaseInsensitivePath(cleanPath(rPath), trailingSlash); ok {
		req.URL.Path = BytesToString(fixedPath)
		redirectRequest(c)
		return true
	}
	return false
}

func redirectRequest(c *Context) {
	req := c.Req
	rPath := req.URL.Path
	rURL := req.URL.String()

	code := http.StatusMovedPermanently // Permanent redirect, request with GET method
	if req.Method != http.MethodGet {
		code = http.StatusTemporaryRedirect
	}
	debugPrint("redirecting request %d: %s --> %s", code, rPath, rURL)
	http.Redirect(c.Writer, req, rURL, code)
	c.Writer.WriteHeaderNow()
}

var mimePlain = []string{"text/plain"}

func serveError(c *Context, code int, defaultMessage []byte) {
	c.responseWriter.status = code
	c.Next()
	if c.responseWriter.Written() {
		return
	}
	if c.responseWriter.Status() == code {
		c.responseWriter.Header()["Content-Type"] = mimePlain
		_, err := c.Writer.Write(defaultMessage)
		if err != nil {
			debugPrint("cannot write message to writer during serve error: %v", err)
		}
		return
	}
	c.responseWriter.WriteHeaderNow()
}

func (engine *Engine) reuseContext(ctx *Context) {
	engine.pool.Put(ctx)
}

// resolveAddress
func (engine *Engine) resolveAddress(addr []string) string {
	switch len(addr) {
	case 0:
		return engine.httpAddr
	case 1:
		return addr[0]
	default:
		panic("too many parameters")
	}
}
