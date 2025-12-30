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
		return handleAuthFailure(c)
	}

	email := sess.Get("cloudflare_email")
	apiKey := sess.Get("user_api_key")

	if email == nil || apiKey == nil {
		return handleAuthFailure(c)
	}

	// 注入到上下文
	c.Locals("cloudflare_email", email.(string))
	c.Locals("user_api_key", apiKey.(string))

	return c.Next()
}

// handleAuthFailure 处理认证失败 - 根据请求类型返回不同响应
func handleAuthFailure(c *fiber.Ctx) error {
	// 检查是否为 API 请求
	path := c.Path()
	if len(path) >= 4 && path[:4] == "/api" {
		// API 请求返回 JSON
		return c.Status(401).JSON(fiber.Map{
			"success": false,
			"error":   "未登录或会话已过期，请重新登录",
		})
	}
	// 页面请求重定向到登录页
	return c.Redirect("/login")
}
