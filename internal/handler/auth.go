package handler

import (
	"context"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/gofiber/fiber/v2"

	"github.com/zhufengme/Cloudflare-DNS-Manager/internal/middleware"
)

type AuthHandler struct {
	RateLimiter *middleware.RateLimiter
}

func NewAuthHandler(rateLimiter *middleware.RateLimiter) *AuthHandler {
	return &AuthHandler{
		RateLimiter: rateLimiter,
	}
}

func (h *AuthHandler) ShowLogin(c *fiber.Ctx) error {
	// 重定向到首页，使用首页的登录表单
	return c.Redirect("/")
}

func (h *AuthHandler) PostLogin(c *fiber.Ctx) error {
	email := c.FormValue("cloudflare_email")
	apiKey := c.FormValue("cloudflare_api")
	remember := c.FormValue("remember") == "on"

	// 限流检查
	if !h.RateLimiter.CheckAndIncrement(email) {
		return c.Render("home/index", fiber.Map{
			"Error": "登录失败次数过多，请一小时后再试",
		})
	}

	// 验证 API Key
	api, err := cloudflare.New(apiKey, email)
	if err != nil {
		return c.Render("home/index", fiber.Map{
			"Error": "无效的凭证",
		})
	}

	// 验证凭证有效性
	_, err = api.UserDetails(context.Background())
	if err != nil {
		return c.Render("home/index", fiber.Map{
			"Error": "无效的凭证或 API Key",
		})
	}

	// 创建会话
	sess, _ := middleware.Store.Get(c)
	sess.Set("cloudflare_email", email)
	sess.Set("user_api_key", apiKey)

	// 设置会话和 Cookie 过期时间
	if remember {
		// 记住我：365 天
		sess.SetExpiry(365 * 24 * time.Hour)

		// 设置 Cookie 过期时间为 365 天
		c.Cookie(&fiber.Cookie{
			Name:     "session_id",
			Value:    sess.ID(),
			MaxAge:   365 * 24 * 60 * 60, // 365 天（秒）
			HTTPOnly: true,
			SameSite: "Lax",
		})
	} else {
		// 不记住：使用会话 Cookie（浏览器关闭后过期）
		sess.SetExpiry(24 * time.Hour) // Session 本身 24 小时过期
	}

	if err := sess.Save(); err != nil {
		return c.Render("home/index", fiber.Map{
			"Error": "会话保存失败",
		})
	}

	return c.Redirect("/zones")
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	sess, _ := middleware.Store.Get(c)
	if err := sess.Destroy(); err != nil {
		return err
	}
	return c.Redirect("/")
}
