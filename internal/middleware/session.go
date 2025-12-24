package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/memory/v2"
)

var Store *session.Store

func InitSession(expiration time.Duration) {
	Store = session.New(session.Config{
		Storage:    memory.New(),
		Expiration: expiration,
		KeyLookup:  "cookie:session_id",
		CookieHTTPOnly: true,
		CookieSameSite: "Lax",
	})
}

// AuthRequired 认证中间件
func AuthRequired(c *fiber.Ctx) error {
	sess, err := Store.Get(c)
	if err != nil {
		return c.Redirect("/login")
	}

	email := sess.Get("cloudflare_email")
	apiKey := sess.Get("user_api_key")

	if email == nil || apiKey == nil {
		return c.Redirect("/login")
	}

	// 注入到上下文
	c.Locals("cloudflare_email", email.(string))
	c.Locals("user_api_key", apiKey.(string))

	return c.Next()
}
