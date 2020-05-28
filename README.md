# Proxy

![Release](https://img.shields.io/github/release/gofiber/proxy.svg)
[![Discord](https://img.shields.io/badge/discord-join%20channel-7289DA)](https://gofiber.io/discord)
![Test](https://github.com/gofiber/proxy/workflows/Test/badge.svg)
![Security](https://github.com/gofiber/proxy/workflows/Security/badge.svg)
![Linter](https://github.com/gofiber/proxy/workflows/Linter/badge.svg)

Simple reverse proxy

### Install
```
go get -u github.com/gofiber/fiber
go get -u github.com/gofiber/proxy
```

### Signature
```go
proxy.New(target string) func(*fiber.Ctx)
proxy.Forward(c *fiber.Ctx, target string) error
```

### Functions
| Name | Signature | Description
| :--- | :--- | :---
| New | `New(config ...Config) func(*fiber.Ctx)` | Returns a middleware that proxies the request
| Handler | `Handler(target string) func(*fiber.Ctx)` | Returns a handler that proxies the request
| Forward | `func Forward(c *fiber.Ctx, target string) error` | A function that proxies the requests

### Example
```go
package main

import (
	"github.com/gofiber/fiber"
	"github.com/gofiber/proxy"
)

func main() {
	go proxy()       // Reverse proxy running on port 3000
	go backend3001() // Backend dummy running on port 3001
	go backend3002() // Backend dummy running on port 3002
}

func proxy() {
	app := fiber.New()

	app.Use("/proxy", proxy.New(proxy.Config{
		DownstreamHosts: []string{
			"127.0.0.1:3001",
			"127.0.0.1:3002",
		},
		Rules: map[string]string{
			"/proxy": "/",
		},
		UpstreamMethods: []string{"GET"},
	}))

	app.Get("/3001", func(ctx *fiber.Ctx) {
		// Alter request
		ctx.Set("X-Forwarded-For", "3001")
		// Forward request using proxy mw function
		if err := proxy.Forward(ctx, "127.0.0.1:3001"); err != nil {
			ctx.SendStatus(503)
			return
		}
		// Alter response
		ctx.Set("X-Forwarded-By", "3001")
	})

	app.Get("/3002", proxy.Handler("127.0.0.1:3002")) // handler

	app.Listen(3000)
}

func backend3001() {
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) {
		c.Send("Hello from the backend server running on port 3001")
	})
	app.Listen(3001)
}

func backend3002() {
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) {
		c.Send("Hello from the backend server running on port 3002")
	})
	app.Listen(3002)
}

```
### Test
```curl
curl http://localhost:3000/3001
curl http://localhost:3000/3002
```