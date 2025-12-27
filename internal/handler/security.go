package handler

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/zhufengme/Cloudflare-DNS-Manager/internal/middleware"
	"github.com/zhufengme/Cloudflare-DNS-Manager/internal/service"
)

type SecurityHandler struct{}

func NewSecurityHandler() *SecurityHandler {
	return &SecurityHandler{}
}

// ShowSecurity 显示安全设置页面
func (h *SecurityHandler) ShowSecurity(c *fiber.Ctx) error {
	zoneID := c.Query("zoneid")
	domain := c.Query("domain")

	if zoneID == "" || domain == "" {
		return c.Status(400).SendString("Missing zoneid or domain parameter")
	}

	// 获取凭证
	email := c.Locals("cloudflare_email").(string)
	apiKey := c.Locals("user_api_key").(string)

	cfService, err := service.NewCloudflareService(email, apiKey)
	if err != nil {
		return c.Render("security/index", fiber.Map{
			"Error":  "Failed to initialize Cloudflare service",
			"ZoneID": zoneID,
			"Domain": domain,
		})
	}

	// 获取 SSL 验证信息
	sslRecords, _ := cfService.GetSSLVerification(context.Background(), zoneID)

	// 获取 DNSSEC 状态
	dnssec, _ := cfService.GetDNSSEC(context.Background(), zoneID)

	return c.Render("security/index", fiber.Map{
		"ZoneID":     zoneID,
		"Domain":     domain,
		"SSLRecords": sslRecords,
		"DNSSEC":     dnssec,
	})
}

// ToggleDNSSEC 切换 DNSSEC 状态
func (h *SecurityHandler) ToggleDNSSEC(c *fiber.Ctx) error {
	zoneID := c.FormValue("zoneid")
	domain := c.FormValue("domain")
	action := c.FormValue("action") // "enable" or "disable"

	if zoneID == "" || domain == "" {
		return c.Redirect(fmt.Sprintf("/security?zoneid=%s&domain=%s&error=missing_params", zoneID, domain))
	}

	// 获取凭证
	sess, _ := middleware.Store.Get(c)
	email := sess.Get("cloudflare_email").(string)
	apiKey := sess.Get("user_api_key").(string)

	cfService, err := service.NewCloudflareService(email, apiKey)
	if err != nil {
		return c.Redirect(fmt.Sprintf("/security?zoneid=%s&domain=%s&error=service_init_failed", zoneID, domain))
	}

	// 更新 DNSSEC 状态
	status := "disabled"
	if action == "enable" {
		status = "active"
	}

	_, err = cfService.UpdateDNSSEC(context.Background(), zoneID, status)
	if err != nil {
		return c.Redirect(fmt.Sprintf("/security?zoneid=%s&domain=%s&error=dnssec_update_failed", zoneID, domain))
	}

	return c.Redirect(fmt.Sprintf("/security?zoneid=%s&domain=%s&success=dnssec_updated", zoneID, domain))
}
