// 🚀 Fiber is an Express inspired web framework written in Go with 💖
// 📌 API Documentation: https://fiber.wiki
// 📝 Github Repository: https://github.com/gofiber/fiber

package proxy

import (
	"sync"

	"github.com/gofiber/fiber"
	"github.com/valyala/fasthttp"
)

// *fasthttp.HostClient pool
var poolHostClient = sync.Pool{
	New: func() interface{} {
		return new(fasthttp.HostClient)
	},
}

// Acquire *fasthttp.HostClient from pool
func acquireHostClient() *fasthttp.HostClient {
	return poolHostClient.Get().(*fasthttp.HostClient)
}

// Return *fasthttp.HostClient to pool
func releaseHostClient(hc *fasthttp.HostClient) {
	hc.Addr = ""
	poolHostClient.Put(hc)
}

// New returns a new reverse proxy handler
func New(target string) func(*fiber.Ctx) {
	// Must provide a target
	if target == "" {
		panic("You must provide a backend server <host>:<port>")
	}
	// Return middleware handler
	return func(c *fiber.Ctx) {
		// Get new hostclient from pool and release when done
		client := acquireHostClient()
		defer releaseHostClient(client)
		// Adjust hostclient settings
		client.Addr = target
		// Prepare request
		req := &c.Fasthttp.Request
		resp := &c.Fasthttp.Response
		// Alter request and remove unneeded headers from request
		req.Header.Del("Connection")
		// Make request to upstream host and handle erropr
		if err := client.Do(req, resp); err != nil {
			c.SendStatus(503) // Service Unavailble
			return
		}
		// Alter request and remove unneeded headers from response
		resp.Header.Del("Connection")
	}
}

// Forward proxies the Ctx to the target
func Forward(c *fiber.Ctx, target string) error {
	// Must provide a target
	if target == "" {
		panic("You must provide a backend server <host>:<port>")
	}
	// Get new hostclient from pool and release when done
	client := acquireHostClient()
	defer releaseHostClient(client)
	// Adjust hostclient settings
	client.Addr = target
	// Prepare request
	req := &c.Fasthttp.Request
	resp := &c.Fasthttp.Response
	// Alter request and remove unneeded headers from request
	req.Header.Del("Connection")
	// Make request to upstream host and handle erropr
	if err := client.Do(req, resp); err != nil {
		c.SendStatus(503) // 503 Service Unavailble
		return err
	}
	// Alter request and remove unneeded headers from response
	resp.Header.Del("Connection")
	// Return without errors
	return nil
}
