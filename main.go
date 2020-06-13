// 🚀 Fiber is an Express inspired web framework written in Go with 💖
// 📌 API Documentation: https://fiber.wiki
// 📝 Github Repository: https://github.com/gofiber/fiber

package proxy

import (
	"log"
	"regexp"
	"strconv"
	"strings"
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

// Config holds configuration for the middleware
type Config struct {
	// Targets is list of backend hosts used to proxy the request.
	// Backend hosts is selected by Round-Robin scheduling.
	// Required. Default: nil
	Targets []string

	// Methods is list of HTTP methods allowed for proxying.
	// Optional. Default: nil
	Methods []string

	// Filter defines a function to skip middleware.
	// Optional. Default: nil
	Filter func(*fiber.Ctx) bool

	// ErrorHandler is a function for handling unexpected errors.
	// Optional. Default: StatusServiceUnavailable
	ErrorHandler func(*fiber.Ctx, error)

	// Rules defines the URL path rewrite rules. The values captured in asterisk can be
	// retrieved by index e.g. $1, $2 and so on.
	// Optional. Default: nil
	// Example:
	// "/old":              "/new",
	// "/api/*":            "/$1",
	// "/js/*":             "/public/javascripts/$1",
	// "/users/*/orders/*": "/user/$1/order/$2",
	Rules map[string]string

	rulesRegex map[*regexp.Regexp]string
}

// New returns a new reverse proxy middleware
func New(config ...Config) func(*fiber.Ctx) {

	var cfg Config
	if len(config) > 0 {
		cfg = config[0]
	}

	if len(cfg.Targets) == 0 {
		log.Fatal("Fiber: Proxy middleware requires at least one backend server <host>:<port>")
	}

	if cfg.ErrorHandler == nil {
		cfg.ErrorHandler = func(c *fiber.Ctx, err error) {
			c.SendStatus(fiber.StatusServiceUnavailable)
			return
		}
	}

	cfg.rulesRegex = map[*regexp.Regexp]string{}
	for k, v := range cfg.Rules {
		k = strings.Replace(k, "*", "(.*)", -1)
		k = k + "$"
		cfg.rulesRegex[regexp.MustCompile(k)] = v
	}

	return func(c *fiber.Ctx) {

		if cfg.Filter != nil && cfg.Filter(c) {
			c.Next()
			return
		}

		for _, method := range cfg.Methods {
			if c.Method() != strings.ToUpper(method) {
				c.Next()
				return
			}
		}

		req := &c.Fasthttp.Request
		resp := &c.Fasthttp.Response

		for k, v := range cfg.rulesRegex {
			replacer := captureTokens(k, c.Path())
			if replacer != nil {
				req.SetRequestURI(replacer.Replace(v))
			}
		}

		target := strings.Join(cfg.Targets, ",")
		if err := proxy(req, resp, target); err != nil {
			cfg.ErrorHandler(c, err)
		}
		return
	}
}

// Handler returns a reverse proxy handler
func Handler(target string) func(*fiber.Ctx) {
	if target == "" {
		log.Fatal("Fiber: Proxy middleware requires backend server <host>:<port>")
	}

	return func(c *fiber.Ctx) {
		if err := Forward(c, target); err != nil {
			c.Next(err)
		}
	}
}

// Forward proxies the Ctx to the target
func Forward(c *fiber.Ctx, target string) error {
	// Must provide a target
	if target == "" {
		panic("You must provide a backend server <host>:<port>")
	}

	req := &c.Fasthttp.Request
	resp := &c.Fasthttp.Response

	if err := proxy(req, resp, target); err != nil {
		c.SendStatus(fiber.StatusServiceUnavailable)
		return err
	}
	return nil
}

func proxy(req *fasthttp.Request, resp *fasthttp.Response, target string) error {
	// Get new hostclient from pool and release when done
	client := acquireHostClient()
	defer releaseHostClient(client)
	// Adjust hostclient settings
	client.Addr = target
	// Alter request and remove unneeded headers from request
	req.Header.Del("Connection")
	// Make request to upstream host and handle error
	if err := client.Do(req, resp); err != nil {
		return err
	}
	// Alter request and remove unneeded headers from response
	resp.Header.Del("Connection")
	// Return without errors
	return nil
}

// https://github.com/labstack/echo/blob/master/middleware/rewrite.go
func captureTokens(pattern *regexp.Regexp, input string) *strings.Replacer {
	groups := pattern.FindAllStringSubmatch(input, -1)
	if groups == nil {
		return nil
	}
	values := groups[0][1:]
	replace := make([]string, 2*len(values))
	for i, v := range values {
		j := 2 * i
		replace[j] = "$" + strconv.Itoa(i+1)
		replace[j+1] = v
	}
	return strings.NewReplacer(replace...)
}
