package gow

import (
	"github.com/gkzy/gow/session"
)

var (
	cookieName     = "gow_session_id"
	sessionManager *session.Manager
	sessionID      string
)

// InitSession   init gow session
//	before using session,please call this function
func InitSession() {
	sessionManager = session.NewSessionManager(cookieName, 3600)
}

// Session session middleware
//		r := gow.Default()
//		r.Use(gow.Session())
func Session() HandlerFunc {
	return func(c *Context) {
		if sessionManager == nil {
			panic("please call gow.InitSession()")
		}
		sessionID = sessionManager.Start(c.Writer, c.Req)
		sessionManager.Extension(c.Writer, c.Req)
		c.Next()
	}
}

//getSession getSession
func getSession(key interface{}) interface{} {
	v, ok := sessionManager.Get(sessionID, key)
	if ok {
		return v
	}
	return nil
}

//setSession setSession
func setSession(key, value interface{}) {
	sessionManager.Set(sessionID, key, value)
}

//deleteSession deleteSession
func deleteSession(key interface{}) {
	sessionManager.Delete(sessionID, key)
}
