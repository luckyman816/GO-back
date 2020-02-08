// 🚀 Fiber is an Express.js inspired web framework written in Go with 💖
// 📌 Please open an issue if you got suggestions or found a bug!
// 🖥 Links: https://github.com/gofiber/fiber, https://fiber.wiki

// 🦸 Not all heroes wear capes, thank you to some amazing people
// 💖 @valyala, @erikdubbelboer, @savsgio, @julienschmidt, @koddr

package fiber

// Connect establishes a tunnel to the server
// identified by the target resource.
func (r *Fiber) Connect(args ...interface{}) {
	r.register("CONNECT", args...)
}

// Put replaces all current representations
// of the target resource with the request payload.
func (r *Fiber) Put(args ...interface{}) {
	r.register("PUT", args...)
}

// Post is used to submit an entity to the specified resource,
// often causing a change in state or side effects on the server.
func (r *Fiber) Post(args ...interface{}) {
	r.register("POST", args...)
}

// Delete deletes the specified resource.
func (r *Fiber) Delete(args ...interface{}) {
	r.register("DELETE", args...)
}

// Head asks for a response identical to that of a GET request,
// but without the response body.
func (r *Fiber) Head(args ...interface{}) {
	r.register("HEAD", args...)
}

// Patch is used to apply partial modifications to a resource.
func (r *Fiber) Patch(args ...interface{}) {
	r.register("PATCH", args...)
}

// Options is used to describe the communication options
// for the target resource.
func (r *Fiber) Options(args ...interface{}) {
	r.register("OPTIONS", args...)
}

// Trace performs a message loop-back test
// along the path to the target resource.
func (r *Fiber) Trace(args ...interface{}) {
	r.register("TRACE", args...)
}

// Get requests a representation of the specified resource.
// Requests using GET should only retrieve data.
func (r *Fiber) Get(args ...interface{}) {
	r.register("GET", args...)
}

// All matches any HTTP method
func (r *Fiber) All(args ...interface{}) {
	r.register("ALL", args...)
}

// Use only matches the starting path
func (r *Fiber) Use(args ...interface{}) {
	r.register("MIDWARE", args...)
}
