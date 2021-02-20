package fiber

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"io/ioutil"
	"mime/multipart"
	"net"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2/utils"
	"github.com/valyala/fasthttp/fasthttputil"
)

func Test_Client_Invalid_URL(t *testing.T) {
	t.Parallel()

	ln := fasthttputil.NewInmemoryListener()

	app := New(Config{DisableStartupMessage: true})

	app.Get("/", func(c *Ctx) error {
		return c.SendString(c.Hostname())
	})

	go app.Listener(ln) //nolint:errcheck

	a := Get("http://example.com\r\n\r\nGET /\r\n\r\n")

	a.HostClient.Dial = func(addr string) (net.Conn, error) { return ln.Dial() }

	_, body, errs := a.String()

	utils.AssertEqual(t, "", body)
	utils.AssertEqual(t, 1, len(errs))
	utils.AssertEqual(t, "missing required Host header in request", errs[0].Error())
}

func Test_Client_Unsupported_Protocol(t *testing.T) {
	t.Parallel()

	a := Get("ftp://example.com")

	_, body, errs := a.String()

	utils.AssertEqual(t, "", body)
	utils.AssertEqual(t, 1, len(errs))
	utils.AssertEqual(t, `unsupported protocol "ftp". http and https are supported`,
		errs[0].Error())
}

func Test_Client_Get(t *testing.T) {
	t.Parallel()

	ln := fasthttputil.NewInmemoryListener()

	app := New(Config{DisableStartupMessage: true})

	app.Get("/", func(c *Ctx) error {
		return c.SendString(c.Hostname())
	})

	go app.Listener(ln) //nolint:errcheck

	for i := 0; i < 5; i++ {
		a := Get("http://example.com")

		a.HostClient.Dial = func(addr string) (net.Conn, error) { return ln.Dial() }

		code, body, errs := a.String()

		utils.AssertEqual(t, StatusOK, code)
		utils.AssertEqual(t, "example.com", body)
		utils.AssertEqual(t, 0, len(errs))
	}
}

func Test_Client_Head(t *testing.T) {
	t.Parallel()

	ln := fasthttputil.NewInmemoryListener()

	app := New(Config{DisableStartupMessage: true})

	app.Get("/", func(c *Ctx) error {
		return c.SendString(c.Hostname())
	})

	go app.Listener(ln) //nolint:errcheck

	for i := 0; i < 5; i++ {
		a := Head("http://example.com")

		a.HostClient.Dial = func(addr string) (net.Conn, error) { return ln.Dial() }

		code, body, errs := a.String()

		utils.AssertEqual(t, StatusOK, code)
		utils.AssertEqual(t, "", body)
		utils.AssertEqual(t, 0, len(errs))
	}
}

func Test_Client_Post(t *testing.T) {
	t.Parallel()

	ln := fasthttputil.NewInmemoryListener()

	app := New(Config{DisableStartupMessage: true})

	app.Post("/", func(c *Ctx) error {
		return c.SendString(c.Hostname())
	})

	go app.Listener(ln) //nolint:errcheck

	for i := 0; i < 5; i++ {
		args := AcquireArgs()

		args.Set("foo", "bar")

		a := Post("http://example.com").
			Form(args)

		a.HostClient.Dial = func(addr string) (net.Conn, error) { return ln.Dial() }

		code, body, errs := a.String()

		utils.AssertEqual(t, StatusOK, code)
		utils.AssertEqual(t, "example.com", body)
		utils.AssertEqual(t, 0, len(errs))

		ReleaseArgs(args)
	}
}

func Test_Client_Put(t *testing.T) {
	t.Parallel()

	ln := fasthttputil.NewInmemoryListener()

	app := New(Config{DisableStartupMessage: true})

	app.Put("/", func(c *Ctx) error {
		return c.SendString(c.FormValue("foo"))
	})

	go app.Listener(ln) //nolint:errcheck

	for i := 0; i < 5; i++ {
		args := AcquireArgs()

		args.Set("foo", "bar")

		a := Put("http://example.com").
			Form(args)

		a.HostClient.Dial = func(addr string) (net.Conn, error) { return ln.Dial() }

		code, body, errs := a.String()

		utils.AssertEqual(t, StatusOK, code)
		utils.AssertEqual(t, "bar", body)
		utils.AssertEqual(t, 0, len(errs))

		ReleaseArgs(args)
	}
}

func Test_Client_Patch(t *testing.T) {
	t.Parallel()

	ln := fasthttputil.NewInmemoryListener()

	app := New(Config{DisableStartupMessage: true})

	app.Patch("/", func(c *Ctx) error {
		return c.SendString(c.FormValue("foo"))
	})

	go app.Listener(ln) //nolint:errcheck

	for i := 0; i < 5; i++ {
		args := AcquireArgs()

		args.Set("foo", "bar")

		a := Patch("http://example.com").
			Form(args)

		a.HostClient.Dial = func(addr string) (net.Conn, error) { return ln.Dial() }

		code, body, errs := a.String()

		utils.AssertEqual(t, StatusOK, code)
		utils.AssertEqual(t, "bar", body)
		utils.AssertEqual(t, 0, len(errs))

		ReleaseArgs(args)
	}
}

func Test_Client_Delete(t *testing.T) {
	t.Parallel()

	ln := fasthttputil.NewInmemoryListener()

	app := New(Config{DisableStartupMessage: true})

	app.Delete("/", func(c *Ctx) error {
		return c.Status(StatusNoContent).
			SendString("deleted")
	})

	go app.Listener(ln) //nolint:errcheck

	for i := 0; i < 5; i++ {
		args := AcquireArgs()

		a := Delete("http://example.com")

		a.HostClient.Dial = func(addr string) (net.Conn, error) { return ln.Dial() }

		code, body, errs := a.String()

		utils.AssertEqual(t, StatusNoContent, code)
		utils.AssertEqual(t, "", body)
		utils.AssertEqual(t, 0, len(errs))

		ReleaseArgs(args)
	}
}

func Test_Client_UserAgent(t *testing.T) {
	t.Parallel()

	ln := fasthttputil.NewInmemoryListener()

	app := New(Config{DisableStartupMessage: true})

	app.Get("/", func(c *Ctx) error {
		return c.Send(c.Request().Header.UserAgent())
	})

	go app.Listener(ln) //nolint:errcheck

	t.Run("default", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			a := Get("http://example.com")

			a.HostClient.Dial = func(addr string) (net.Conn, error) { return ln.Dial() }

			code, body, errs := a.String()

			utils.AssertEqual(t, StatusOK, code)
			utils.AssertEqual(t, defaultUserAgent, body)
			utils.AssertEqual(t, 0, len(errs))
		}
	})

	t.Run("custom", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			c := AcquireClient()
			c.UserAgent = "ua"

			a := c.Get("http://example.com")

			a.HostClient.Dial = func(addr string) (net.Conn, error) { return ln.Dial() }

			code, body, errs := a.String()

			utils.AssertEqual(t, StatusOK, code)
			utils.AssertEqual(t, "ua", body)
			utils.AssertEqual(t, 0, len(errs))
			ReleaseClient(c)
		}
	})
}

func Test_Client_Agent_Headers(t *testing.T) {
	handler := func(c *Ctx) error {
		c.Request().Header.VisitAll(func(key, value []byte) {
			if k := string(key); k == "K1" || k == "K2" {
				_, _ = c.Write(key)
				_, _ = c.Write(value)
			}
		})
		return nil
	}

	wrapAgent := func(a *Agent) {
		a.Set("k1", "v1").
			Add("k1", "v11").
			Set("k2", "v2")
	}

	testAgent(t, handler, wrapAgent, "K1v1K1v11K2v2")
}

func Test_Client_Agent_Connection_Close(t *testing.T) {
	handler := func(c *Ctx) error {
		if c.Request().Header.ConnectionClose() {
			return c.SendString("close")
		}
		return c.SendString("not close")
	}

	wrapAgent := func(a *Agent) {
		a.ConnectionClose()
	}

	testAgent(t, handler, wrapAgent, "close")
}

func Test_Client_Agent_UserAgent(t *testing.T) {
	handler := func(c *Ctx) error {
		return c.Send(c.Request().Header.UserAgent())
	}

	wrapAgent := func(a *Agent) {
		a.UserAgent("ua")
	}

	testAgent(t, handler, wrapAgent, "ua")
}

func Test_Client_Agent_Cookie(t *testing.T) {
	handler := func(c *Ctx) error {
		return c.SendString(
			c.Cookies("k1") + c.Cookies("k2") + c.Cookies("k3") + c.Cookies("k4"))
	}

	wrapAgent := func(a *Agent) {
		a.Cookie("k1", "v1").
			Cookie("k2", "v2").
			Cookies("k3", "v3", "k4", "v4")
	}

	testAgent(t, handler, wrapAgent, "v1v2v3v4")
}

func Test_Client_Agent_Referer(t *testing.T) {
	handler := func(c *Ctx) error {
		return c.Send(c.Request().Header.Referer())
	}

	wrapAgent := func(a *Agent) {
		a.Referer("http://referer.com")
	}

	testAgent(t, handler, wrapAgent, "http://referer.com")
}

func Test_Client_Agent_ContentType(t *testing.T) {
	handler := func(c *Ctx) error {
		return c.Send(c.Request().Header.ContentType())
	}

	wrapAgent := func(a *Agent) {
		a.ContentType("custom-type")
	}

	testAgent(t, handler, wrapAgent, "custom-type")
}

func Test_Client_Agent_Specific_Host(t *testing.T) {
	t.Parallel()

	ln := fasthttputil.NewInmemoryListener()

	app := New(Config{DisableStartupMessage: true})

	app.Get("/", func(c *Ctx) error {
		return c.SendString(c.Hostname())
	})

	go app.Listener(ln) //nolint:errcheck

	a := Get("http://1.1.1.1:8080").
		Host("example.com")

	utils.AssertEqual(t, "1.1.1.1:8080", a.HostClient.Addr)

	a.HostClient.Dial = func(addr string) (net.Conn, error) { return ln.Dial() }

	code, body, errs := a.String()

	utils.AssertEqual(t, StatusOK, code)
	utils.AssertEqual(t, "example.com", body)
	utils.AssertEqual(t, 0, len(errs))
}

func Test_Client_Agent_QueryString(t *testing.T) {
	handler := func(c *Ctx) error {
		return c.Send(c.Request().URI().QueryString())
	}

	wrapAgent := func(a *Agent) {
		a.QueryString("foo=bar&bar=baz")
	}

	testAgent(t, handler, wrapAgent, "foo=bar&bar=baz")
}

func Test_Client_Agent_BasicAuth(t *testing.T) {
	handler := func(c *Ctx) error {
		// Get authorization header
		auth := c.Get(HeaderAuthorization)
		// Decode the header contents
		raw, err := base64.StdEncoding.DecodeString(auth[6:])
		utils.AssertEqual(t, nil, err)

		return c.Send(raw)
	}

	wrapAgent := func(a *Agent) {
		a.BasicAuth("foo", "bar")
	}

	testAgent(t, handler, wrapAgent, "foo:bar")
}

func Test_Client_Agent_BodyString(t *testing.T) {
	handler := func(c *Ctx) error {
		return c.Send(c.Request().Body())
	}

	wrapAgent := func(a *Agent) {
		a.BodyString("foo=bar&bar=baz")
	}

	testAgent(t, handler, wrapAgent, "foo=bar&bar=baz")
}

func Test_Client_Agent_BodyStream(t *testing.T) {
	handler := func(c *Ctx) error {
		return c.Send(c.Request().Body())
	}

	wrapAgent := func(a *Agent) {
		a.BodyStream(strings.NewReader("body stream"), -1)
	}

	testAgent(t, handler, wrapAgent, "body stream")
}

func Test_Client_Agent_Custom_Request_And_Response(t *testing.T) {
	t.Parallel()

	ln := fasthttputil.NewInmemoryListener()

	app := New(Config{DisableStartupMessage: true})

	app.Get("/", func(c *Ctx) error {
		return c.SendString("custom")
	})

	go app.Listener(ln) //nolint:errcheck

	for i := 0; i < 5; i++ {
		a := AcquireAgent()
		req := AcquireRequest()
		resp := AcquireResponse()

		req.Header.SetMethod(MethodGet)
		req.SetRequestURI("http://example.com")
		a.Request(req)

		utils.AssertEqual(t, nil, a.Parse())

		a.HostClient.Dial = func(addr string) (net.Conn, error) { return ln.Dial() }

		code, body, errs := a.String(resp)

		utils.AssertEqual(t, StatusOK, code)
		utils.AssertEqual(t, "custom", body)
		utils.AssertEqual(t, "custom", string(resp.Body()))
		utils.AssertEqual(t, 0, len(errs))

		ReleaseRequest(req)
		ReleaseResponse(resp)
	}
}

func Test_Client_Agent_Json(t *testing.T) {
	handler := func(c *Ctx) error {
		utils.AssertEqual(t, MIMEApplicationJSON, string(c.Request().Header.ContentType()))

		return c.Send(c.Request().Body())
	}

	wrapAgent := func(a *Agent) {
		a.JSON(data{Success: true})
	}

	testAgent(t, handler, wrapAgent, `{"success":true}`)
}

func Test_Client_Agent_Json_Error(t *testing.T) {
	a := Get("http://example.com").
		JSON(complex(1, 1))

	_, body, errs := a.String()

	utils.AssertEqual(t, "", body)
	utils.AssertEqual(t, 1, len(errs))
	utils.AssertEqual(t, "json: unsupported type: complex128", errs[0].Error())
}

func Test_Client_Agent_XML(t *testing.T) {
	handler := func(c *Ctx) error {
		utils.AssertEqual(t, MIMEApplicationXML, string(c.Request().Header.ContentType()))

		return c.Send(c.Request().Body())
	}

	wrapAgent := func(a *Agent) {
		a.XML(data{Success: true})
	}

	testAgent(t, handler, wrapAgent, "<data><success>true</success></data>")
}

func Test_Client_Agent_XML_Error(t *testing.T) {
	a := Get("http://example.com").
		XML(complex(1, 1))

	_, body, errs := a.String()

	utils.AssertEqual(t, "", body)
	utils.AssertEqual(t, 1, len(errs))
	utils.AssertEqual(t, "xml: unsupported type: complex128", errs[0].Error())
}

func Test_Client_Agent_Form(t *testing.T) {
	handler := func(c *Ctx) error {
		utils.AssertEqual(t, MIMEApplicationForm, string(c.Request().Header.ContentType()))

		return c.Send(c.Request().Body())
	}

	args := AcquireArgs()

	args.Set("foo", "bar")

	wrapAgent := func(a *Agent) {
		a.Form(args)
	}

	testAgent(t, handler, wrapAgent, "foo=bar")

	ReleaseArgs(args)
}

func Test_Client_Agent_Multipart(t *testing.T) {
	t.Parallel()

	ln := fasthttputil.NewInmemoryListener()

	app := New(Config{DisableStartupMessage: true})

	app.Post("/", func(c *Ctx) error {
		utils.AssertEqual(t, "multipart/form-data; boundary=myBoundary", c.Get(HeaderContentType))

		mf, err := c.MultipartForm()
		utils.AssertEqual(t, nil, err)
		utils.AssertEqual(t, "bar", mf.Value["foo"][0])

		return c.Send(c.Request().Body())
	})

	go app.Listener(ln) //nolint:errcheck

	args := AcquireArgs()

	args.Set("foo", "bar")

	a := Post("http://example.com").
		Boundary("myBoundary").
		MultipartForm(args)

	a.HostClient.Dial = func(addr string) (net.Conn, error) { return ln.Dial() }

	code, body, errs := a.String()

	utils.AssertEqual(t, StatusOK, code)
	utils.AssertEqual(t, "--myBoundary\r\nContent-Disposition: form-data; name=\"foo\"\r\n\r\nbar\r\n--myBoundary--\r\n", body)
	utils.AssertEqual(t, 0, len(errs))
	ReleaseArgs(args)
}

func Test_Client_Agent_Multipart_SendFiles(t *testing.T) {
	t.Parallel()

	ln := fasthttputil.NewInmemoryListener()

	app := New(Config{DisableStartupMessage: true})

	app.Post("/", func(c *Ctx) error {
		utils.AssertEqual(t, "multipart/form-data; boundary=myBoundary", c.Get(HeaderContentType))

		mf, err := c.MultipartForm()
		utils.AssertEqual(t, nil, err)

		fh1 := mf.File["field1"][0]
		utils.AssertEqual(t, fh1.Filename, "name")
		buf := make([]byte, fh1.Size, fh1.Size)
		f, err := fh1.Open()
		utils.AssertEqual(t, nil, err)
		defer func() { _ = f.Close() }()
		_, err = f.Read(buf)
		utils.AssertEqual(t, nil, err)
		utils.AssertEqual(t, "form file", string(buf))

		checkFormFile(t, mf.File["index"][0], ".github/testdata/index.html")
		checkFormFile(t, mf.File["file3"][0], ".github/testdata/index.tmpl")

		return c.SendString("multipart form files")
	})

	go app.Listener(ln) //nolint:errcheck

	for i := 0; i < 5; i++ {
		ff := AcquireFormFile()
		ff.Fieldname = "field1"
		ff.Name = "name"
		ff.Content = []byte("form file")

		a := Post("http://example.com").
			Boundary("myBoundary").
			FileData(ff).
			SendFiles(".github/testdata/index.html", "index", ".github/testdata/index.tmpl").
			MultipartForm(nil)

		a.HostClient.Dial = func(addr string) (net.Conn, error) { return ln.Dial() }

		code, body, errs := a.String()

		utils.AssertEqual(t, StatusOK, code)
		utils.AssertEqual(t, "multipart form files", body)
		utils.AssertEqual(t, 0, len(errs))

		ReleaseFormFile(ff)
	}
}

func checkFormFile(t *testing.T, fh *multipart.FileHeader, filename string) {
	basename := filepath.Base(filename)
	utils.AssertEqual(t, fh.Filename, basename)

	b1, err := ioutil.ReadFile(filename)
	utils.AssertEqual(t, nil, err)

	b2 := make([]byte, fh.Size, fh.Size)
	f, err := fh.Open()
	utils.AssertEqual(t, nil, err)
	defer func() { _ = f.Close() }()
	_, err = f.Read(b2)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, b1, b2)
}

func Test_Client_Agent_Multipart_Random_Boundary(t *testing.T) {
	t.Parallel()

	a := Post("http://example.com").
		MultipartForm(nil)

	reg := regexp.MustCompile(`multipart/form-data; boundary=\w{30}`)

	utils.AssertEqual(t, true, reg.Match(a.req.Header.Peek(HeaderContentType)))
}

func Test_Client_Agent_Multipart_Invalid_Boundary(t *testing.T) {
	t.Parallel()

	a := Post("http://example.com").
		Boundary("*").
		MultipartForm(nil)

	utils.AssertEqual(t, 1, len(a.errs))
	utils.AssertEqual(t, "mime: invalid boundary character", a.errs[0].Error())
}

func Test_Client_Agent_SendFile_Error(t *testing.T) {
	t.Parallel()

	a := Post("http://example.com").
		SendFile("", "")

	utils.AssertEqual(t, 1, len(a.errs))
	utils.AssertEqual(t, "open : no such file or directory", a.errs[0].Error())
}

func Test_Client_Debug(t *testing.T) {
	handler := func(c *Ctx) error {
		return c.SendString("debug")
	}

	var output bytes.Buffer

	wrapAgent := func(a *Agent) {
		a.Debug(&output)
	}

	testAgent(t, handler, wrapAgent, "debug", 1)

	str := output.String()

	utils.AssertEqual(t, true, strings.Contains(str, "Connected to example.com(pipe)"))
	utils.AssertEqual(t, true, strings.Contains(str, "GET / HTTP/1.1"))
	utils.AssertEqual(t, true, strings.Contains(str, "User-Agent: fiber"))
	utils.AssertEqual(t, true, strings.Contains(str, "Host: example.com\r\n\r\n"))
	utils.AssertEqual(t, true, strings.Contains(str, "HTTP/1.1 200 OK"))
	utils.AssertEqual(t, true, strings.Contains(str, "Content-Type: text/plain; charset=utf-8\r\nContent-Length: 5\r\n\r\ndebug"))
}

func Test_Client_Agent_Timeout(t *testing.T) {
	t.Parallel()

	ln := fasthttputil.NewInmemoryListener()

	app := New(Config{DisableStartupMessage: true})

	app.Get("/", func(c *Ctx) error {
		time.Sleep(time.Millisecond * 200)
		return c.SendString("timeout")
	})

	go app.Listener(ln) //nolint:errcheck

	a := Get("http://example.com").
		Timeout(time.Millisecond * 100)

	a.HostClient.Dial = func(addr string) (net.Conn, error) { return ln.Dial() }

	_, body, errs := a.String()

	utils.AssertEqual(t, "", body)
	utils.AssertEqual(t, 1, len(errs))
	utils.AssertEqual(t, "timeout", errs[0].Error())
}

func Test_Client_Agent_Reuse(t *testing.T) {
	t.Parallel()

	ln := fasthttputil.NewInmemoryListener()

	app := New(Config{DisableStartupMessage: true})

	app.Get("/", func(c *Ctx) error {
		return c.SendString("reuse")
	})

	go app.Listener(ln) //nolint:errcheck

	a := Get("http://example.com").
		Reuse()

	a.HostClient.Dial = func(addr string) (net.Conn, error) { return ln.Dial() }

	code, body, errs := a.String()

	utils.AssertEqual(t, StatusOK, code)
	utils.AssertEqual(t, "reuse", body)
	utils.AssertEqual(t, 0, len(errs))

	code, body, errs = a.String()

	utils.AssertEqual(t, StatusOK, code)
	utils.AssertEqual(t, "reuse", body)
	utils.AssertEqual(t, 0, len(errs))
}

func Test_Client_Agent_TLS(t *testing.T) {
	t.Parallel()

	// Create tls certificate
	cer, err := tls.LoadX509KeyPair("./.github/testdata/ssl.pem", "./.github/testdata/ssl.key")
	utils.AssertEqual(t, nil, err)

	config := &tls.Config{
		Certificates: []tls.Certificate{cer},
	}

	ln, err := net.Listen(NetworkTCP4, "127.0.0.1:0")
	utils.AssertEqual(t, nil, err)

	ln = tls.NewListener(ln, config)

	app := New(Config{DisableStartupMessage: true})

	app.Get("/", func(c *Ctx) error {
		return c.SendString("tls")
	})

	go app.Listener(ln) //nolint:errcheck

	code, body, errs := Get("https://" + ln.Addr().String()).
		InsecureSkipVerify().
		TLSConfig(config).
		InsecureSkipVerify().
		String()

	utils.AssertEqual(t, 0, len(errs))
	utils.AssertEqual(t, StatusOK, code)
	utils.AssertEqual(t, "tls", body)
}

func Test_Client_Agent_MaxRedirectsCount(t *testing.T) {
	t.Parallel()

	ln := fasthttputil.NewInmemoryListener()

	app := New(Config{DisableStartupMessage: true})

	app.Get("/", func(c *Ctx) error {
		if c.Request().URI().QueryArgs().Has("foo") {
			return c.Redirect("/foo")
		}
		return c.Redirect("/")
	})
	app.Get("/foo", func(c *Ctx) error {
		return c.SendString("redirect")
	})

	go app.Listener(ln) //nolint:errcheck

	t.Run("success", func(t *testing.T) {
		a := Get("http://example.com?foo").
			MaxRedirectsCount(1)

		a.HostClient.Dial = func(addr string) (net.Conn, error) { return ln.Dial() }

		code, body, errs := a.String()

		utils.AssertEqual(t, 200, code)
		utils.AssertEqual(t, "redirect", body)
		utils.AssertEqual(t, 0, len(errs))
	})

	t.Run("error", func(t *testing.T) {
		a := Get("http://example.com").
			MaxRedirectsCount(1)

		a.HostClient.Dial = func(addr string) (net.Conn, error) { return ln.Dial() }

		_, body, errs := a.String()

		utils.AssertEqual(t, "", body)
		utils.AssertEqual(t, 1, len(errs))
		utils.AssertEqual(t, "too many redirects detected when doing the request", errs[0].Error())
	})
}

func Test_Client_Agent_Struct(t *testing.T) {
	t.Parallel()

	ln := fasthttputil.NewInmemoryListener()

	app := New(Config{DisableStartupMessage: true})

	app.Get("/", func(c *Ctx) error {
		return c.JSON(data{true})
	})

	app.Get("/error", func(c *Ctx) error {
		return c.SendString(`{"success"`)
	})

	go app.Listener(ln) //nolint:errcheck

	t.Run("success", func(t *testing.T) {
		a := Get("http://example.com")

		a.HostClient.Dial = func(addr string) (net.Conn, error) { return ln.Dial() }

		var d data

		code, body, errs := a.Struct(&d)

		utils.AssertEqual(t, StatusOK, code)
		utils.AssertEqual(t, `{"success":true}`, string(body))
		utils.AssertEqual(t, 0, len(errs))
		utils.AssertEqual(t, true, d.Success)
	})

	t.Run("error", func(t *testing.T) {
		a := Get("http://example.com/error")

		a.HostClient.Dial = func(addr string) (net.Conn, error) { return ln.Dial() }

		var d data

		code, body, errs := a.Struct(&d)

		utils.AssertEqual(t, StatusOK, code)
		utils.AssertEqual(t, `{"success"`, string(body))
		utils.AssertEqual(t, 1, len(errs))
		utils.AssertEqual(t, "json: unexpected end of JSON input after object field key: ", errs[0].Error())
	})
}

func Test_Client_Agent_Parse(t *testing.T) {
	t.Parallel()

	a := Get("https://example.com:10443")

	utils.AssertEqual(t, nil, a.Parse())
}

func Test_AddMissingPort_TLS(t *testing.T) {
	addr := addMissingPort("example.com", true)
	utils.AssertEqual(t, "example.com:443", addr)
}

func testAgent(t *testing.T, handler Handler, wrapAgent func(agent *Agent), excepted string, count ...int) {
	t.Parallel()

	ln := fasthttputil.NewInmemoryListener()

	app := New(Config{DisableStartupMessage: true})

	app.Get("/", handler)

	go app.Listener(ln) //nolint:errcheck

	c := 1
	if len(count) > 0 {
		c = count[0]
	}

	for i := 0; i < c; i++ {
		a := Get("http://example.com")

		wrapAgent(a)

		a.HostClient.Dial = func(addr string) (net.Conn, error) { return ln.Dial() }

		code, body, errs := a.String()

		utils.AssertEqual(t, StatusOK, code)
		utils.AssertEqual(t, excepted, body)
		utils.AssertEqual(t, 0, len(errs))
	}
}

type data struct {
	Success bool `json:"success" xml:"success"`
}
