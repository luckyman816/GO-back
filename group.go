// 🚀 Fiber is an Express inspired web framework written in Go with 💖
// 📌 API Documentation: https://fiber.wiki
// 📝 Github Repository: https://github.com/gofiber/fiber

package fiber

import (
	"log"
	"net/http"
	"reflect"
)

// Group ...
type Group struct {
	prefix string
	app    *App
}

// Group : https://fiber.wiki/application#group
func (grp *Group) Group(prefix string, handlers ...func(*Ctx)) *Group {
	prefix = groupPaths(grp.prefix, prefix)
	if len(handlers) > 0 {
		grp.app.registerMethod("USE", prefix, handlers...)
	}
	return &Group{
		prefix: prefix,
		app:    grp.app,
	}
}

// Static : https://fiber.wiki/application#static
func (grp *Group) Static(prefix, root string, config ...Static) *Group {
	prefix = groupPaths(grp.prefix, prefix)
	grp.app.registerStatic(prefix, root, config...)
	return grp
}

// Use only match requests starting with the specified prefix
// It's optional to provide a prefix, default: "/"
// Example: Use("/product", handler)
// will match 	/product
// will match 	/product/cool
// will match 	/product/foo
//
// https://fiber.wiki/application#http-methods
func (grp *Group) Use(args ...interface{}) *Group {
	var path = ""
	var handlers []func(*Ctx)
	for i := 0; i < len(args); i++ {
		switch arg := args[i].(type) {
		case string:
			path = arg
		case func(*Ctx):
			handlers = append(handlers, arg)
		default:
			log.Fatalf("Invalid handler: %v", reflect.TypeOf(arg))
		}
	}
	path = groupPaths(grp.prefix, path)
	grp.app.registerMethod("USE", path, handlers...)
	return grp
}

// Connect : https://fiber.wiki/application#http-methods
func (grp *Group) Connect(path string, handlers ...func(*Ctx)) *Group {
	path = groupPaths(grp.prefix, path)
	grp.app.registerMethod(http.MethodConnect, path, handlers...)
	return grp
}

// Put : https://fiber.wiki/application#http-methods
func (grp *Group) Put(path string, handlers ...func(*Ctx)) *Group {
	path = groupPaths(grp.prefix, path)
	grp.app.registerMethod(http.MethodPut, path, handlers...)
	return grp
}

// Post : https://fiber.wiki/application#http-methods
func (grp *Group) Post(path string, handlers ...func(*Ctx)) *Group {
	path = groupPaths(grp.prefix, path)
	grp.app.registerMethod(http.MethodPost, path, handlers...)
	return grp
}

// Delete : https://fiber.wiki/application#http-methods
func (grp *Group) Delete(path string, handlers ...func(*Ctx)) *Group {
	path = groupPaths(grp.prefix, path)
	grp.app.registerMethod(http.MethodDelete, path, handlers...)
	return grp
}

// Head : https://fiber.wiki/application#http-methods
func (grp *Group) Head(path string, handlers ...func(*Ctx)) *Group {
	path = groupPaths(grp.prefix, path)
	grp.app.registerMethod(http.MethodHead, path, handlers...)
	return grp
}

// Patch : https://fiber.wiki/application#http-methods
func (grp *Group) Patch(path string, handlers ...func(*Ctx)) *Group {
	path = groupPaths(grp.prefix, path)
	grp.app.registerMethod(http.MethodPatch, path, handlers...)
	return grp
}

// Options : https://fiber.wiki/application#http-methods
func (grp *Group) Options(path string, handlers ...func(*Ctx)) *Group {
	path = groupPaths(grp.prefix, path)
	grp.app.registerMethod(http.MethodOptions, path, handlers...)
	return grp
}

// Trace : https://fiber.wiki/application#http-methods
func (grp *Group) Trace(path string, handlers ...func(*Ctx)) *Group {
	path = groupPaths(grp.prefix, path)
	grp.app.registerMethod(http.MethodTrace, path, handlers...)
	return grp
}

// Get : https://fiber.wiki/application#http-methods
func (grp *Group) Get(path string, handlers ...func(*Ctx)) *Group {
	path = groupPaths(grp.prefix, path)
	grp.app.registerMethod(http.MethodGet, path, handlers...)
	return grp
}

// All matches all HTTP methods and complete paths
// Example: All("/product", handler)
// will match 	/product
// won't match 	/product/cool   <-- important
// won't match 	/product/foo    <-- important
//
// https://fiber.wiki/application#http-methods
func (grp *Group) All(path string, handlers ...func(*Ctx)) *Group {
	path = groupPaths(grp.prefix, path)
	grp.app.registerMethod("ALL", path, handlers...)
	return grp
}

// This function is deprecated since v1.8.2!
// Please us github.com/gofiber/websocket
func (grp *Group) WebSocket(path string, handle func(*Ctx)) *Group {
	log.Println("Warning: app.WebSocket() is deprecated since v1.8.2, please use github.com/gofiber/websocket instead.")
	path = groupPaths(grp.prefix, path)
	grp.app.registerWebSocket(http.MethodGet, path, handle)
	return grp
}

// This function is deprecated since v1.8.2!
// Please us github.com/gofiber/recover
func (grp *Group) Recover(handler func(*Ctx)) {
	log.Println("Warning: Recover() is deprecated since v1.8.2, please use github.com/gofiber/recover instead.")
	grp.app.recover = handler
}
