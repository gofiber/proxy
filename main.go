// 🚀 Fiber is an Express inspired web framework written in Go with 💖
// 📌 API Documentation: https://fiber.wiki
// 📝 Github Repository: https://github.com/gofiber/fiber

package proxy

import (
	"sync"

	"github.com/gofiber/fiber"
	"github.com/valyala/fasthttp"
)

// Config defines the config for BasicAuth middleware
type Config struct {
	// Filter defines a function to skip middleware.
	// Optional. Default: nil
	Filter func(*fiber.Ctx) bool

	// Backend is the upstream host
	Backend string

	// fasthttp.HostClient pool
	pool *sync.Pool
}

// New returns a new reverse proxy handler
func New(config ...Config) func(*fiber.Ctx) {
	// Init config
	var cfg Config
	if len(config) > 0 {
		cfg = config[0]
	}
	// if cfg.Users == nil {
	// 	cfg.Users = map[string]string{}
	// }
	cfg.pool = &sync.Pool{
		New: func() interface{} {
			return new(fasthttp.HostClient)
		},
	}

	// Return middleware handler
	return func(c *fiber.Ctx) {
		// Filter request to skip middleware
		if cfg.Filter != nil && cfg.Filter(c) {
			c.Next()
			return
		}

		// Get new hostclient and set some settings
		client := cfg.acquire()
		// Release hostclient when we are done
		defer cfg.release(client)
		client.Addr = cfg.Backend

		// Prepare request
		req := &c.Fasthttp.Request
		resp := &c.Fasthttp.Response

		// Strip unneeded headers from request
		req.Header.Del("Connection")
		// Alter other request params before sending them to upstream host

		// Make request to upstream host and handle erropr
		if err := client.Do(req, resp); err != nil {
			c.SendStatus(503) // Service Unavailble
			return
		}

		// Remove unneeded headers from response
		resp.Header.Del("Connection")
		// Alter other response data if needed
	}
}

// Acquire Ctx from pool
func (config *Config) acquire() *fasthttp.HostClient {
	return config.pool.Get().(*fasthttp.HostClient)
}

// Return Ctx to pool
func (config *Config) release(hc *fasthttp.HostClient) {
	hc.Addr = ""
	config.pool.Put(hc)
}
