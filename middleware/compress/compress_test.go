package compress

import (
	"errors"
	"io/ioutil"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
)

var filedata []byte

func init() {
	dat, err := ioutil.ReadFile("../../.github/README.md")
	if err != nil {
		panic(err)
	}
	filedata = dat
}

// go test -run Test_Compress
func Test_Compress_Gzip(t *testing.T) {
	app := fiber.New()

	app.Use(New(Config{Level: 10}))

	app.Get("/", func(c *fiber.Ctx) error {
		c.Set(fiber.HeaderContentType, fiber.MIMETextPlainCharsetUTF8)
		return c.Send(filedata)
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	resp, err := app.Test(req)
	utils.AssertEqual(t, nil, err, "app.Test(req)")
	utils.AssertEqual(t, 200, resp.StatusCode, "Status code")
	utils.AssertEqual(t, "gzip", resp.Header.Get(fiber.HeaderContentEncoding))

	// Validate the file size is shrinked
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		utils.AssertEqual(t, nil, err)
	}
	utils.AssertEqual(t, true, len(body) < len(filedata))
}

func Test_Compress_Deflate(t *testing.T) {
	app := fiber.New()

	app.Use(New(Config{Level: LevelBestSpeed}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Send(filedata)
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "deflate")

	resp, err := app.Test(req)
	utils.AssertEqual(t, nil, err, "app.Test(req)")
	utils.AssertEqual(t, 200, resp.StatusCode, "Status code")
	utils.AssertEqual(t, "deflate", resp.Header.Get(fiber.HeaderContentEncoding))

	// Validate the file size is shrinked
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		utils.AssertEqual(t, nil, err)
	}
	utils.AssertEqual(t, true, len(body) < len(filedata))
}

func Test_Compress_Brotli(t *testing.T) {
	app := fiber.New()

	app.Use(New(Config{Level: LevelBestCompression}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Send(filedata)
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "br")

	resp, err := app.Test(req)
	utils.AssertEqual(t, nil, err, "app.Test(req)")
	utils.AssertEqual(t, 200, resp.StatusCode, "Status code")
	utils.AssertEqual(t, "br", resp.Header.Get(fiber.HeaderContentEncoding))

	// Validate the file size is shrinked
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		utils.AssertEqual(t, nil, err)
	}
	utils.AssertEqual(t, true, len(body) < len(filedata))
}

func Test_Compress_Disabled(t *testing.T) {
	app := fiber.New()

	app.Use(New(Config{Level: LevelDisabled}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Send(filedata)
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "br")

	resp, err := app.Test(req)
	utils.AssertEqual(t, nil, err, "app.Test(req)")
	utils.AssertEqual(t, 200, resp.StatusCode, "Status code")
	utils.AssertEqual(t, "", resp.Header.Get(fiber.HeaderContentEncoding))

	// Validate the file size is not shrinked
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		utils.AssertEqual(t, nil, err)
	}
	utils.AssertEqual(t, true, len(body) == len(filedata))
}

func Test_Compress_Next_Error(t *testing.T) {
	app := fiber.New()

	app.Use(New())

	app.Get("/", func(c *fiber.Ctx) error {
		return errors.New("next error")
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	resp, err := app.Test(req)
	utils.AssertEqual(t, nil, err, "app.Test(req)")
	utils.AssertEqual(t, 500, resp.StatusCode, "Status code")
	utils.AssertEqual(t, "", resp.Header.Get(fiber.HeaderContentEncoding))

	body, err := ioutil.ReadAll(resp.Body)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, "next error", string(body))
}

// go test -run Test_Compress_Next
func Test_Compress_Next(t *testing.T) {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	app.Use(New(Config{
		Next: func(_ *fiber.Ctx) bool {
			return true
		},
	}))

	resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusNotFound, resp.StatusCode)
}
