package gow

import (
	"bufio"
	"net"
	"net/http"
)

const (
	noWritten = -1
)

type beforeFunc func(ResponseWriter)

//ResponseWriter interface
type ResponseWriter interface {
	http.ResponseWriter
	http.Hijacker
	http.Flusher
	http.CloseNotifier
	Status() int
	Size() int
	Written() bool
	WriteHeaderNow()
	Before(func(ResponseWriter))
}

//writer write 实现
type responseWriter struct {
	http.ResponseWriter
	status      int
	size        int
	beforeFuncs []beforeFunc
}

func (c *responseWriter) Status() int {
	return c.status
}

func (c *responseWriter) Size() int {
	return c.size
}

func (c *responseWriter) Written() bool {
	return c.size != noWritten
}

func (c *responseWriter) WriteHeaderNow() {
	if !c.Written() {
		c.size = 0
		c.callBefore()
		c.ResponseWriter.WriteHeader(c.status)
	}
}
func (c *responseWriter) Before(before func(ResponseWriter)) {
	c.beforeFuncs = append(c.beforeFuncs, before)
}

func (c *responseWriter) Write(data []byte) (size int, err error) {
	c.WriteHeaderNow()
	size, err = c.ResponseWriter.Write(data)
	c.size += size
	return
}

func (c *responseWriter) WriteHeader(code int) {
	if code > 0 && c.status != code {
		if c.Written() {
			debugPrint("[WARNING] Headers were already written. Wanted to override status code %d with %d", c.status, code)
		}
		c.status = code
	}
}

func (c *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if c.size < 0 {
		c.size = 0
	}
	return c.ResponseWriter.(http.Hijacker).Hijack()
}

func (c *responseWriter) CloseNotify() <-chan bool {
	return c.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

func (c *responseWriter) Flush() {
	c.WriteHeaderNow()
	c.ResponseWriter.(http.Flusher).Flush()
}

func (c *responseWriter) callBefore() {
	for i := len(c.beforeFuncs) - 1; i >= 0; i-- {
		c.beforeFuncs[i](c)
	}
}

func (c *responseWriter) reset(writer http.ResponseWriter) {
	c.ResponseWriter = writer
	c.beforeFuncs = nil
	c.size = noWritten
	c.status = http.StatusOK
}
