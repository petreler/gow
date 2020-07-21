package gow

import (
	"net/http"
	"path"
	"strings"
)

var (
	// HTTPMethod  http request method
	HTTPMethod = map[string]bool{
		"GET":     true,
		"POST":    true,
		"PUT":     true,
		"DELETE":  true,
		"PATCH":   true,
		"OPTIONS": true,
		"HEAD":    true,
		"TRACE":   true,
	}
)

// RouterGroup gow routerGroup
type RouterGroup struct {
	Handlers HandlersChain
	basePath string
	engine   *Engine
	root     bool
}

//Use use middleware
func (group *RouterGroup) Use(middleware ...HandlerFunc) {
	group.engine.rebuild404Handlers()
	group.engine.rebuild405Handlers()
	group.Handlers = append(group.Handlers, middleware...)
}

//Group router group
func (group *RouterGroup) Group(path string, handlers ...HandlerFunc) *RouterGroup {
	return &RouterGroup{
		Handlers: group.combineHandlers(handlers),
		basePath: group.calculateAbsolutePath(path),
		engine:   group.engine,
	}
}

//Handle Handle
func (group *RouterGroup) Handle(method, p string, handlers []HandlerFunc) {
	method = strings.ToUpper(method)
	//debugPrint("method:%s",method)
	if method != "*" && !HTTPMethod[method] {
		panic("not support http method: " + method)
	}
	absolutePath := group.calculateAbsolutePath(p)
	handlers = group.combineHandlers(handlers)
	//handler any method
	if method == "*" {
		for val := range HTTPMethod {
			group.engine.addRoute(val, absolutePath, handlers)
		}
	} else {
		group.engine.addRoute(method, absolutePath, handlers)
	}
}

//RouterMap
type RouterMap struct {
	Path        string
	Method      string
	HandlerName string
}

// Any
func (group *RouterGroup) Any(path string, handlers ...HandlerFunc) {
	group.Handle("*", path, handlers)
}

// HEAD
func (group *RouterGroup) HEAD(path string, handlers ...HandlerFunc) {
	group.Handle("HEAD", path, handlers)
}

// POST
func (group *RouterGroup) POST(path string, handlers ...HandlerFunc) {
	group.Handle("POST", path, handlers)
}

// GET
func (group *RouterGroup) GET(path string, handlers ...HandlerFunc) {
	group.Handle("GET", path, handlers)
}

// DELETE
func (group *RouterGroup) DELETE(path string, handlers ...HandlerFunc) {
	group.Handle("DELETE", path, handlers)
}

// PATCH
func (group *RouterGroup) PATCH(path string, handlers ...HandlerFunc) {
	group.Handle("PATCH", path, handlers)
}

// PUT
func (group *RouterGroup) PUT(path string, handlers ...HandlerFunc) {
	group.Handle("PUT", path, handlers)
}

// TRACE
func (group *RouterGroup) TRACE(path string, handlers ...HandlerFunc) {
	group.Handle("TRACE", path, handlers)
}

//StaticFile 实现了单个静态文件的路由
// router.StaticFile("favicon.ico","./static/favicon.ico")
func (group *RouterGroup) StaticFile(relativePath, filepath string) {
	if strings.Contains(relativePath, ":") || strings.Contains(relativePath, "*") {
		panic("URL parameters can not be used when serving a static file")
	}
	handler := func(c *Context) {
		c.File(filepath)
	}
	group.GET(relativePath, handler)
	group.HEAD(relativePath, handler)
}

//Static 静态资源文件
// use:
//		router.Static("/static","static")
func (group *RouterGroup) Static(relativePath, root string) {
	group.engine.staticPath = root
	group.StaticFS(relativePath, Dir(group.engine.staticPath, false))
}

//StaticFS
func (group *RouterGroup) StaticFS(relativePath string, fs http.FileSystem) {
	if strings.Contains(relativePath, ":") || strings.Contains(relativePath, "*") {
		panic("URL parameters can not be used when serving a static folder")
	}
	handler := group.createStaticHandler(relativePath, fs)
	urlPattern := path.Join(relativePath, "/*filepath")
	// Register GET and HEAD handlers
	group.GET(urlPattern, handler)
	group.HEAD(urlPattern, handler)
}

//================================private func=============================

func (group *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	absolutePath := group.calculateAbsolutePath(relativePath)
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))

	return func(c *Context) {
		if _, nolisting := fs.(*onlyfilesFS); nolisting {
			c.Writer.WriteHeader(http.StatusNotFound)
		}

		file := c.Param("filepath")
		// Check if file exists and/or if we have permission to access it
		f, err := fs.Open(file)
		if err != nil {
			c.Writer.WriteHeader(http.StatusNotFound)
			c.handlers = group.engine.noRoute
			// Reset index
			c.ServerString(404, string(default404Body))
			return
		}
		c.Status(200)
		f.Close()
		fileServer.ServeHTTP(c.Writer, c.Req)
	}
}

func (engine *Engine) addRoute(method, path string, handlers HandlersChain) {
	root := engine.trees.get(method)
	if root == nil {
		root = new(node)
		root.fullPath = "/"
		engine.trees = append(engine.trees, methodTree{method: method, root: root})
	}
	root.addRoute(path, handlers)
}

func (group *RouterGroup) combineHandlers(handlers HandlersChain) HandlersChain {
	finalSize := len(group.Handlers) + len(handlers)
	if finalSize >= int(abortIndex) {
		panic("too many handlers")
	}
	mergedHandlers := make(HandlersChain, finalSize)
	copy(mergedHandlers, group.Handlers)
	copy(mergedHandlers[len(group.Handlers):], handlers)
	return mergedHandlers
}

func (group *RouterGroup) calculateAbsolutePath(relativePath string) string {
	return joinPaths(group.basePath, relativePath)
}
