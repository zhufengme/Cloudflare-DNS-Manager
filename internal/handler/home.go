package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/zhufengme/Cloudflare-DNS-Manager/internal/middleware"
)

type HomeHandler struct{}

func NewHomeHandler() *HomeHandler {
	return &HomeHandler{}
}

// ShowHome 显示首页
func (h *HomeHandler) ShowHome(c *fiber.Ctx) error {
	// 检查用户是否已登录
	sess, _ := middleware.Store.Get(c)
	email := sess.Get("cloudflare_email")

	// 如果已登录，直接跳转到域名列表
	if email != nil {
		return c.Redirect("/zones")
	}

	// 未登录，显示首页
	return c.Render("home/index", fiber.Map{})
}
