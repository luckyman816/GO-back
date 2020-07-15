package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber"
	"github.com/gofiber/utils"
	"github.com/valyala/bytebufferpool"
	"github.com/valyala/fasthttp"
)

// go test -run Test_Middleware_Logger
func Test_Middleware_Logger(t *testing.T) {
	format := "${ip}-${ips}-${url}-${host}-${method}-${path}-${protocol}-${route}-${referer}-${ua}-${status}-${body}-${error}-${bytesSent}-${bytesReceived}-${header:header}-${query:query}-${form:form}-${cookie:cookie}-${black}-${red}-${green}-${yellow}-${blue}-${magenta}-${cyan}-${white}-${resetColor}"
	expect := "0.0.0.0--/test?query=query-example.com-POST-/test-http-/test-ref-ua-500-form=form-error-5-9-header-query-form-cookie-\u001b[90m-\u001b[91m-\u001b[92m-\u001b[93m-\u001b[94m-\u001b[95m-\u001b[96m-\u001b[97m-\u001b[0m"

	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	app := fiber.New()

	app.Use(Logger(LoggerConfig{
		Format: format,
		Output: buf,
	}))

	app.Post("/test", func(ctx *fiber.Ctx) {
		ctx.Next(errors.New("error"))
	})

	req := httptest.NewRequest("POST", "/test?query=query", strings.NewReader("form=form"))
	req.Header.Set("header", "header")
	req.Header.Set("Cookie", "cookie=cookie")
	req.Header.Set("User-Agent", "ua")
	req.Header.Set("Referer", "ref")
	req.Header.Set("Content-type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req)
	utils.AssertEqual(t, nil, err, "app.Test(req)")
	utils.AssertEqual(t, fiber.StatusInternalServerError, resp.StatusCode, "Status code")
	utils.AssertEqual(t, expect, buf.String())

}

func Test_Middleware_Logger_WithDefaultFormat(t *testing.T) {
	// fake output
	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	LoggerConfigDefault.Output = buf
	config := LoggerConfigDefault
	config.Output = nil

	app := fiber.New(&fiber.Settings{DisableStartupMessage: true})
	app.Use(Logger(config))

	app.Get("/", func(ctx *fiber.Ctx) {
		ctx.SendStatus(fiber.StatusOK)
	})

	_, err := app.Test(httptest.NewRequest(http.MethodGet, "/", nil))
	utils.AssertEqual(t, nil, err, "app.Test(req)")
	res := buf.String()
	utils.AssertEqual(t, 48, len(res), fmt.Sprintf("Has length: %v, expected: %v, raw: %s", len(res), 48, res))
}

// go test -run Test_Middleware_Logger_Skip
func Test_Middleware_Logger_Skip(t *testing.T) {
	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	LoggerConfigDefault.Output = buf

	app := fiber.New()

	app.Use(Logger(func(_ *fiber.Ctx) bool {
		return true
	}))

	app.Get("/", func(_ *fiber.Ctx) {})

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/", nil))
	utils.AssertEqual(t, nil, err, "app.Test(req)")
	utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode, "Status code")
	utils.AssertEqual(t, 0, buf.Len(), "buf.Len()")
}

// go test -run Test_Middleware_Logger_Options_And_WithConfig
func Test_Middleware_Logger_Options_And_WithConfig(t *testing.T) {
	t.Parallel()

	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	LoggerConfigDefault.Output = buf

	loggers := []fiber.Handler{
		Logger(buf),
		Logger("15:04:05"),
		Logger("${time} ${method} ${path} - ${ip} - ${status} - ${latency}\n"),
		Logger(LoggerConfig{Output: buf}),
	}

	for i, logger := range loggers {
		buf.Reset()

		app := fiber.New()

		app.Use(logger)

		app.Get("/", func(_ *fiber.Ctx) {})

		resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/", nil))
		utils.AssertEqual(t, nil, err, "app.Test(req)")
		utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode, "Status code")
		res := buf.String()

		if i == 0 {
			utils.AssertEqual(t, 48, len(res), fmt.Sprintf("Has length: %v, expected: %v, raw: %s", len(res), 48, res))
		} else if i == 1 {
			utils.AssertEqual(t, 37, len(res), fmt.Sprintf("Has length: %v, expected: %v, raw: %s", len(res), 37, res))
		} else if i == 2 {
			utils.AssertEqual(t, 51, len(res), fmt.Sprintf("Has length: %v, expected: %v, raw: %s", len(res), 51, res))
		} else if i == 3 {
			utils.AssertEqual(t, 48, len(res), fmt.Sprintf("Has length: %v, expected: %v, raw: %s", len(res), 48, res))
		}
	}
}

// go test -v -run=^$ -bench=Benchmark_Middleware_Logger -benchmem -count=4
func Benchmark_Middleware_Logger(b *testing.B) {

	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	app := fiber.New()
	app.Use(Logger(LoggerConfig{
		Output: buf,
	}))

	app.Get("/", func(c *fiber.Ctx) {})
	handler := app.Handler()

	c := &fasthttp.RequestCtx{}
	c.Request.SetRequestURI("/")

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		handler(c)
	}
}
