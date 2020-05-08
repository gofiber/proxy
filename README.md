### Proxy
Basic auth middleware provides an HTTP basic authentication. It calls the next handler for valid credentials and `401 Unauthorized` for missing or invalid credentials.

### Install
```
go get -u github.com/gofiber/fiber
go get -u github.com/gofiber/proxy
```

### Signature
```go
proxy.New(config ...proxy.Config) func(*fiber.Ctx)
```

### Config
| Property | Type | Description | Default |
| :--- | :--- | :--- | :--- |
| Skip | `func(*Ctx) bool` | Defines a function to skip middleware | `nil` |
| Backend | `string` | Upstream host | `""` |

### Example
```go
package main

import (
	"github.com/gofiber/fiber"
	"github.com/gofiber/proxy"
)

func main() {
	go backend()

	app := fiber.New()

	app.Use(proxy.New(proxy.Config{
		Backend: "127.0.0.1:3001",
	}))

	app.Listen(3000)
}

func backend() {
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) {
		c.Send("Hello from the backend server running on port 3001")
	})

	app.Listen(3001)
}
```
### Test
```curl
curl http://localhost:3000
```