// ⚡️ Fiber is an Express inspired web framework written in Go with ☕️
// 🤖 Github Repository: https://github.com/gofiber/fiber
// 📌 API Documentation: https://docs.gofiber.io

package fiber

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	utils "github.com/gofiber/utils"
	fasthttp "github.com/valyala/fasthttp"
)

func testStatus200(t *testing.T, app *App, url string, method string) {
	req := httptest.NewRequest(method, url, nil)

	resp, err := app.Test(req)
	utils.AssertEqual(t, nil, err, "app.Test(req)")
	utils.AssertEqual(t, 200, resp.StatusCode, "Status code")
}

func Test_App_Routes(t *testing.T) {
	app := New()
	h := func(c *Ctx) {}
	app.Get("/Get", h)
	app.Head("/Head", h)
	app.Post("/post", h)
	utils.AssertEqual(t, 3, len(app.Routes()))
}

func Test_App_ServerErrorHandler_SmallReadBuffer(t *testing.T) {
	expectedError := regexp.MustCompile(
		`error when reading request headers: small read buffer\. Increase ReadBufferSize\. Buffer size=4096, contents: "GET / HTTP/1.1\\r\\nHost: example\.com\\r\\nVery-Long-Header: -+`,
	)
	app := New()

	app.Get("/", func(c *Ctx) {
		panic(errors.New("should never called"))
	})

	request := httptest.NewRequest("GET", "/", nil)
	logHeaderSlice := make([]string, 5000, 5000)
	request.Header.Set("Very-Long-Header", strings.Join(logHeaderSlice, "-"))
	_, err := app.Test(request)

	if err == nil {
		t.Error("Expect an error at app.Test(request)")
	}

	utils.AssertEqual(
		t,
		true,
		expectedError.MatchString(err.Error()),
		fmt.Sprintf("Has: %s, expected pattern: %s", err.Error(), expectedError.String()),
	)
}

func Test_App_ErrorHandler(t *testing.T) {
	app := New()

	app.Get("/", func(c *Ctx) {
		c.Next(errors.New("hi, i'm an error"))
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
	utils.AssertEqual(t, nil, err, "app.Test(req)")
	utils.AssertEqual(t, 500, resp.StatusCode, "Status code")

	body, err := ioutil.ReadAll(resp.Body)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, "hi, i'm an error", string(body))

}

func Test_App_ErrorHandler_Custom(t *testing.T) {
	app := New(&Settings{
		ErrorHandler: func(ctx *Ctx, err error) {
			ctx.Status(200).SendString("hi, i'm an custom error")
		},
	})

	app.Get("/", func(c *Ctx) {
		c.Next(errors.New("hi, i'm an error"))
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
	utils.AssertEqual(t, nil, err, "app.Test(req)")
	utils.AssertEqual(t, 200, resp.StatusCode, "Status code")

	body, err := ioutil.ReadAll(resp.Body)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, "hi, i'm an custom error", string(body))
}

func Test_App_Nested_Params(t *testing.T) {
	app := New()

	app.Get("/test", func(c *Ctx) {
		c.Status(400).Send("Should move on")
	})
	app.Get("/test/:param", func(c *Ctx) {
		c.Status(400).Send("Should move on")
	})
	app.Get("/test/:param/test", func(c *Ctx) {
		c.Status(400).Send("Should move on")
	})
	app.Get("/test/:param/test/:param2", func(c *Ctx) {
		c.Status(200).Send("Good job")
	})

	req := httptest.NewRequest("GET", "/test/john/test/doe", nil)
	resp, err := app.Test(req)

	utils.AssertEqual(t, nil, err, "app.Test(req)")
	utils.AssertEqual(t, 200, resp.StatusCode, "Status code")
}

func Test_App_Use_Params(t *testing.T) {
	app := New()

	app.Use("/prefix/:param", func(c *Ctx) {
		utils.AssertEqual(t, "john", c.Params("param"))
	})

	app.Use("/:param/*", func(c *Ctx) {
		utils.AssertEqual(t, "john", c.Params("param"))
		utils.AssertEqual(t, "doe", c.Params("*"))
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/prefix/john", nil))
	utils.AssertEqual(t, nil, err, "app.Test(req)")
	utils.AssertEqual(t, 200, resp.StatusCode, "Status code")

	resp, err = app.Test(httptest.NewRequest("GET", "/john/doe", nil))
	utils.AssertEqual(t, nil, err, "app.Test(req)")
	utils.AssertEqual(t, 200, resp.StatusCode, "Status code")
}

func Test_App_Use_Params_Group(t *testing.T) {
	app := New()

	group := app.Group("/prefix/:param/*")
	group.Use("/", func(c *Ctx) {
		c.Next()
	})
	group.Get("/test", func(c *Ctx) {
		utils.AssertEqual(t, "john", c.Params("param"))
		utils.AssertEqual(t, "doe", c.Params("*"))
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/prefix/john/doe/test", nil))
	utils.AssertEqual(t, nil, err, "app.Test(req)")
	utils.AssertEqual(t, 200, resp.StatusCode, "Status code")
}

func Test_App_Chaining(t *testing.T) {
	n := func(c *Ctx) {
		c.Next()
	}
	app := New()
	app.Use("/john", n, n, n, n, func(c *Ctx) {
		c.Status(202)
	})

	req := httptest.NewRequest("POST", "/john", nil)

	resp, err := app.Test(req)
	utils.AssertEqual(t, nil, err, "app.Test(req)")
	utils.AssertEqual(t, 202, resp.StatusCode, "Status code")

	app.Get("/test", n, n, n, n, func(c *Ctx) {
		c.Status(203)
	})

	req = httptest.NewRequest("GET", "/test", nil)

	resp, err = app.Test(req)
	utils.AssertEqual(t, nil, err, "app.Test(req)")
	utils.AssertEqual(t, 203, resp.StatusCode, "Status code")

}

func Test_App_Order(t *testing.T) {
	app := New()

	app.Get("/test", func(c *Ctx) {
		c.Write("1")
		c.Next()
	})

	app.All("/test", func(c *Ctx) {
		c.Write("2")
		c.Next()
	})

	app.Use(func(c *Ctx) {
		c.Write("3")
	})

	req := httptest.NewRequest("GET", "/test", nil)

	resp, err := app.Test(req)
	utils.AssertEqual(t, nil, err, "app.Test(req)")
	utils.AssertEqual(t, 200, resp.StatusCode, "Status code")

	body, err := ioutil.ReadAll(resp.Body)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, "123", string(body))
}
func Test_App_Methods(t *testing.T) {
	var dummyHandler = func(c *Ctx) {}

	app := New()

	app.Connect("/:john?/:doe?", dummyHandler)
	testStatus200(t, app, "/john/doe", "CONNECT")

	app.Put("/:john?/:doe?", dummyHandler)
	testStatus200(t, app, "/john/doe", "PUT")

	app.Post("/:john?/:doe?", dummyHandler)
	testStatus200(t, app, "/john/doe", "POST")

	app.Delete("/:john?/:doe?", dummyHandler)
	testStatus200(t, app, "/john/doe", "DELETE")

	app.Head("/:john?/:doe?", dummyHandler)
	testStatus200(t, app, "/john/doe", "HEAD")

	app.Patch("/:john?/:doe?", dummyHandler)
	testStatus200(t, app, "/john/doe", "PATCH")

	app.Options("/:john?/:doe?", dummyHandler)
	testStatus200(t, app, "/john/doe", "OPTIONS")

	app.Trace("/:john?/:doe?", dummyHandler)
	testStatus200(t, app, "/john/doe", "TRACE")

	app.Get("/:john?/:doe?", dummyHandler)
	testStatus200(t, app, "/john/doe", "GET")

	app.All("/:john?/:doe?", dummyHandler)
	testStatus200(t, app, "/john/doe", "POST")

	app.Use("/:john?/:doe?", dummyHandler)
	testStatus200(t, app, "/john/doe", "GET")

}

func Test_App_New(t *testing.T) {
	app := New()
	app.Get("/", func(*Ctx) {

	})

	appConfig := New(&Settings{
		Immutable: true,
	})
	appConfig.Get("/", func(*Ctx) {

	})
}

func Test_App_Shutdown(t *testing.T) {
	app := New(&Settings{
		DisableStartupMessage: true,
	})
	_ = app.Shutdown()
}

// go test -run Test_App_Static_Index_Default
func Test_App_Static_Index_Default(t *testing.T) {
	app := New()

	app.Static("/prefix", "./.github/workflows")
	app.Static("/", "./.github")

	resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
	utils.AssertEqual(t, nil, err, "app.Test(req)")
	utils.AssertEqual(t, 200, resp.StatusCode, "Status code")
	utils.AssertEqual(t, false, resp.Header.Get("Content-Length") == "")
	utils.AssertEqual(t, "text/html; charset=utf-8", resp.Header.Get("Content-Type"))

	body, err := ioutil.ReadAll(resp.Body)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, true, strings.Contains(string(body), "Hello, World!"))

}

// go test -run Test_App_Static_Index
func Test_App_Static_Direct(t *testing.T) {
	app := New()

	app.Static("/", "./.github")

	resp, err := app.Test(httptest.NewRequest("GET", "/index.html", nil))
	utils.AssertEqual(t, nil, err, "app.Test(req)")
	utils.AssertEqual(t, 200, resp.StatusCode, "Status code")
	utils.AssertEqual(t, false, resp.Header.Get("Content-Length") == "")
	utils.AssertEqual(t, "text/html; charset=utf-8", resp.Header.Get("Content-Type"))

	body, err := ioutil.ReadAll(resp.Body)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, true, strings.Contains(string(body), "Hello, World!"))

	resp, err = app.Test(httptest.NewRequest("GET", "/FUNDING.yml", nil))
	utils.AssertEqual(t, nil, err, "app.Test(req)")
	utils.AssertEqual(t, 200, resp.StatusCode, "Status code")
	utils.AssertEqual(t, false, resp.Header.Get("Content-Length") == "")
	utils.AssertEqual(t, "text/plain; charset=utf-8", resp.Header.Get("Content-Type"))

	body, err = ioutil.ReadAll(resp.Body)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, true, strings.Contains(string(body), "buymeacoffee"))
}
func Test_App_Static_Group(t *testing.T) {
	app := New()

	grp := app.Group("/v1", func(c *Ctx) {
		c.Set("Test-Header", "123")
		c.Next()
	})

	grp.Static("/v2", "./.github/FUNDING.yml")

	req := httptest.NewRequest("GET", "/v1/v2", nil)
	resp, err := app.Test(req)
	utils.AssertEqual(t, nil, err, "app.Test(req)")
	utils.AssertEqual(t, 200, resp.StatusCode, "Status code")
	utils.AssertEqual(t, false, resp.Header.Get("Content-Length") == "")
	utils.AssertEqual(t, "text/plain; charset=utf-8", resp.Header.Get("Content-Type"))
	utils.AssertEqual(t, "123", resp.Header.Get("Test-Header"))

	grp = app.Group("/v2")
	grp.Static("/v3*", "./.github/FUNDING.yml")

	req = httptest.NewRequest("GET", "/v2/v3/john/doe", nil)
	resp, err = app.Test(req)
	utils.AssertEqual(t, nil, err, "app.Test(req)")
	utils.AssertEqual(t, 200, resp.StatusCode, "Status code")
	utils.AssertEqual(t, false, resp.Header.Get("Content-Length") == "")
	utils.AssertEqual(t, "text/plain; charset=utf-8", resp.Header.Get("Content-Type"))

}

func Test_App_Static_Wildcard(t *testing.T) {
	app := New()

	app.Static("*", "./.github/FUNDING.yml")

	req := httptest.NewRequest("GET", "/yesyes/john/doe", nil)
	resp, err := app.Test(req)
	utils.AssertEqual(t, nil, err, "app.Test(req)")
	utils.AssertEqual(t, 200, resp.StatusCode, "Status code")
	utils.AssertEqual(t, false, resp.Header.Get("Content-Length") == "")
	utils.AssertEqual(t, "text/plain; charset=utf-8", resp.Header.Get("Content-Type"))

	body, err := ioutil.ReadAll(resp.Body)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, true, strings.Contains(string(body), "buymeacoffee"))

}

func Test_App_Static_Prefix_Wildcard(t *testing.T) {
	app := New()

	app.Static("/test/*", "./.github/FUNDING.yml")

	req := httptest.NewRequest("GET", "/test/john/doe", nil)
	resp, err := app.Test(req)
	utils.AssertEqual(t, nil, err, "app.Test(req)")
	utils.AssertEqual(t, 200, resp.StatusCode, "Status code")
	utils.AssertEqual(t, false, resp.Header.Get("Content-Length") == "")
	utils.AssertEqual(t, "text/plain; charset=utf-8", resp.Header.Get("Content-Type"))

	app.Static("/my/nameisjohn*", "./.github/FUNDING.yml")

	resp, err = app.Test(httptest.NewRequest("GET", "/my/nameisjohn/no/its/not", nil))
	utils.AssertEqual(t, nil, err, "app.Test(req)")
	utils.AssertEqual(t, 200, resp.StatusCode, "Status code")
	utils.AssertEqual(t, false, resp.Header.Get("Content-Length") == "")
	utils.AssertEqual(t, "text/plain; charset=utf-8", resp.Header.Get("Content-Type"))

	body, err := ioutil.ReadAll(resp.Body)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, true, strings.Contains(string(body), "buymeacoffee"))
}

func Test_App_Static_Prefix(t *testing.T) {
	app := New()
	app.Static("/john", "./.github")

	req := httptest.NewRequest("GET", "/john/stale.yml", nil)
	resp, err := app.Test(req)
	utils.AssertEqual(t, nil, err, "app.Test(req)")
	utils.AssertEqual(t, 200, resp.StatusCode, "Status code")
	utils.AssertEqual(t, false, resp.Header.Get("Content-Length") == "")
	utils.AssertEqual(t, "text/plain; charset=utf-8", resp.Header.Get("Content-Type"))

	app.Static("/prefix", "./.github/workflows")

	req = httptest.NewRequest("GET", "/prefix/test.yml", nil)
	resp, err = app.Test(req)
	utils.AssertEqual(t, nil, err, "app.Test(req)")
	utils.AssertEqual(t, 200, resp.StatusCode, "Status code")
	utils.AssertEqual(t, false, resp.Header.Get("Content-Length") == "")
	utils.AssertEqual(t, "text/plain; charset=utf-8", resp.Header.Get("Content-Type"))

	app.Static("/single", "./.github/workflows/test.yml")

	req = httptest.NewRequest("GET", "/single", nil)
	resp, err = app.Test(req)
	utils.AssertEqual(t, nil, err, "app.Test(req)")
	utils.AssertEqual(t, 200, resp.StatusCode, "Status code")
	utils.AssertEqual(t, false, resp.Header.Get("Content-Length") == "")
	utils.AssertEqual(t, "text/plain; charset=utf-8", resp.Header.Get("Content-Type"))
}

// go test -run Test_App_Mixed_Routes_WithSameLen
func Test_App_Mixed_Routes_WithSameLen(t *testing.T) {
	app := New()

	// middleware
	app.Use(func(ctx *Ctx) {
		ctx.Set("TestHeader", "TestValue")
		ctx.Next()
	})
	// routes with the same length
	app.Static("/tesbar", "./.github")
	app.Get("/foobar", func(ctx *Ctx) {
		ctx.Send("FOO_BAR")
		ctx.Type("html")
	})

	// match get route
	req := httptest.NewRequest("GET", "/foobar", nil)
	resp, err := app.Test(req)
	utils.AssertEqual(t, nil, err, "app.Test(req)")
	utils.AssertEqual(t, 200, resp.StatusCode, "Status code")
	utils.AssertEqual(t, false, resp.Header.Get("Content-Length") == "")
	utils.AssertEqual(t, "TestValue", resp.Header.Get("TestHeader"))
	utils.AssertEqual(t, "text/html", resp.Header.Get("Content-Type"))

	body, err := ioutil.ReadAll(resp.Body)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, "FOO_BAR", string(body))

	// match static route
	req = httptest.NewRequest("GET", "/tesbar", nil)
	resp, err = app.Test(req)
	utils.AssertEqual(t, nil, err, "app.Test(req)")
	utils.AssertEqual(t, 200, resp.StatusCode, "Status code")
	utils.AssertEqual(t, false, resp.Header.Get("Content-Length") == "")
	utils.AssertEqual(t, "TestValue", resp.Header.Get("TestHeader"))
	utils.AssertEqual(t, "text/html; charset=utf-8", resp.Header.Get("Content-Type"))

	body, err = ioutil.ReadAll(resp.Body)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, true, strings.Contains(string(body), "Hello, World!"), "Response: "+string(body))
	utils.AssertEqual(t, true, strings.HasPrefix(string(body), "<!DOCTYPE html>"), "Response: "+string(body))
}

func Test_App_Group(t *testing.T) {
	var dummyHandler = func(c *Ctx) {}

	app := New()

	grp := app.Group("/test")
	grp.Get("/", dummyHandler)
	testStatus200(t, app, "/test", "GET")

	grp.Get("/:demo?", dummyHandler)
	testStatus200(t, app, "/test/john", "GET")

	grp.Connect("/CONNECT", dummyHandler)
	testStatus200(t, app, "/test/CONNECT", "CONNECT")

	grp.Put("/PUT", dummyHandler)
	testStatus200(t, app, "/test/PUT", "PUT")

	grp.Post("/POST", dummyHandler)
	testStatus200(t, app, "/test/POST", "POST")

	grp.Delete("/DELETE", dummyHandler)
	testStatus200(t, app, "/test/DELETE", "DELETE")

	grp.Head("/HEAD", dummyHandler)
	testStatus200(t, app, "/test/HEAD", "HEAD")

	grp.Patch("/PATCH", dummyHandler)
	testStatus200(t, app, "/test/PATCH", "PATCH")

	grp.Options("/OPTIONS", dummyHandler)
	testStatus200(t, app, "/test/OPTIONS", "OPTIONS")

	grp.Trace("/TRACE", dummyHandler)
	testStatus200(t, app, "/test/TRACE", "TRACE")

	grp.All("/ALL", dummyHandler)
	testStatus200(t, app, "/test/ALL", "POST")

	grp.Use("/USE", dummyHandler)
	testStatus200(t, app, "/test/USE/oke", "GET")

	api := grp.Group("/v1")
	api.Post("/", dummyHandler)

	resp, err := app.Test(httptest.NewRequest("POST", "/test/v1/", nil))
	utils.AssertEqual(t, nil, err, "app.Test(req)")
	utils.AssertEqual(t, 200, resp.StatusCode, "Status code")
	//utils.AssertEqual(t, "/test/v1", resp.Header.Get("Location"), "Location")

	api.Get("/users", dummyHandler)
	resp, err = app.Test(httptest.NewRequest("GET", "/test/v1/UsErS", nil))
	utils.AssertEqual(t, nil, err, "app.Test(req)")
	utils.AssertEqual(t, 200, resp.StatusCode, "Status code")
	//utils.AssertEqual(t, "/test/v1/users", resp.Header.Get("Location"), "Location")
}

func Test_App_Listen(t *testing.T) {
	app := New(&Settings{
		DisableStartupMessage: true,
	})
	go func() {
		time.Sleep(1000 * time.Millisecond)
		utils.AssertEqual(t, nil, app.Shutdown())
	}()

	utils.AssertEqual(t, nil, app.Listen(4003))

	go func() {
		time.Sleep(1000 * time.Millisecond)
		utils.AssertEqual(t, nil, app.Shutdown())
	}()

	utils.AssertEqual(t, nil, app.Listen("4010"))
}

func Test_App_Serve(t *testing.T) {
	app := New(&Settings{
		DisableStartupMessage: true,
		Prefork:               true,
	})
	ln, err := net.Listen("tcp4", ":4020")
	utils.AssertEqual(t, nil, err)

	go func() {
		time.Sleep(1000 * time.Millisecond)
		utils.AssertEqual(t, nil, app.Shutdown())
	}()

	utils.AssertEqual(t, nil, app.Serve(ln))
}

// go test -v -run=^$ -bench=Benchmark_App_ETag -benchmem -count=4
func Benchmark_App_ETag(b *testing.B) {
	app := New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(c)
	c.Send("Hello, World!")
	for n := 0; n < b.N; n++ {
		setETag(c, false)
	}
	utils.AssertEqual(b, `"13-1831710635"`, string(c.Fasthttp.Response.Header.Peek(HeaderETag)))
}

// go test -v -run=^$ -bench=Benchmark_App_ETag_Weak -benchmem -count=4
func Benchmark_App_ETag_Weak(b *testing.B) {
	app := New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(c)
	c.Send("Hello, World!")
	for n := 0; n < b.N; n++ {
		setETag(c, true)
	}
	utils.AssertEqual(b, `W/"13-1831710635"`, string(c.Fasthttp.Response.Header.Peek(HeaderETag)))
}
