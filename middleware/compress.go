package middleware

import (
	"log"

	"github.com/gofiber/fiber"
	"github.com/valyala/fasthttp"
)

// CompressConfig defines the config for Compress middleware.
type CompressConfig struct {
	// Next defines a function to skip this middleware if returned true.
	Next func(ctx *fiber.Ctx) bool
	// Compression level for brotli, gzip and deflate
	Level int
}

// Compression levels
const (
	CompressLevelDisabled        = -1
	CompressLevelDefault         = 0
	CompressLevelBestSpeed       = 1
	CompressLevelBestCompression = 2
)

// CompressConfigDefault is the default config
var CompressConfigDefault = CompressConfig{
	Next:  nil,
	Level: CompressLevelDefault,
}

/*
Compress allows the following config arguments in any order:
	- Compress()
	- Compress(next func(*fiber.Ctx) bool)
	- Compress(level int)
*/
func Compress(options ...interface{}) fiber.Handler {
	// Create default config
	var config = CompressConfigDefault
	// Assert options if provided to adjust the config
	if len(options) > 0 {
		for i := range options {
			switch opt := options[i].(type) {
			case func(*fiber.Ctx) bool:
				config.Next = opt
			case int:
				config.Level = opt
			default:
				log.Fatal("Compress: the following option types are allowed: int")
			}
		}
	}
	// Return CompressWithConfig
	return CompressWithConfig(config)
}

// CompressWithConfig allows you to pass an CompressConfig
// It supports brotli, gzip and deflate compression
// The same order is used to check against the Accept-Encoding header
func CompressWithConfig(config CompressConfig) fiber.Handler {
	// Init middleware settings
	var compress fasthttp.RequestHandler
	switch config.Level {
	case -1: // Disabled
		compress = fasthttp.CompressHandlerBrotliLevel(func(c *fasthttp.RequestCtx) {}, fasthttp.CompressBrotliNoCompression, fasthttp.CompressNoCompression)
	case 1: // Best speed
		compress = fasthttp.CompressHandlerBrotliLevel(func(c *fasthttp.RequestCtx) {}, fasthttp.CompressBrotliBestSpeed, fasthttp.CompressBestSpeed)
	case 2: // Best compression
		compress = fasthttp.CompressHandlerBrotliLevel(func(c *fasthttp.RequestCtx) {}, fasthttp.CompressBrotliBestCompression, fasthttp.CompressBestCompression)
	default: // Default
		compress = fasthttp.CompressHandlerBrotliLevel(func(c *fasthttp.RequestCtx) {}, fasthttp.CompressBrotliDefaultCompression, fasthttp.CompressDefaultCompression)
	}
	// Return handler
	return func(c *fiber.Ctx) {
		// Don't execute the middleware if Next returns false
		if config.Next != nil && config.Next(c) {
			c.Next()
			return
		}
		// Middleware logic...
		c.Next()
		// Compress response
		compress(c.Fasthttp)
	}
}
