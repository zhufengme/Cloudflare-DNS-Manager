package handler

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/zhufengme/Cloudflare-DNS-Manager/internal/service"
)

type AnalyticsHandler struct{}

func NewAnalyticsHandler() *AnalyticsHandler {
	return &AnalyticsHandler{}
}

// ShowAnalytics 显示分析统计页面
func (h *AnalyticsHandler) ShowAnalytics(c *fiber.Ctx) error {
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
		return c.Render("analytics/index", fiber.Map{
			"Error":  "Failed to initialize Cloudflare service",
			"ZoneID": zoneID,
			"Domain": domain,
		})
	}

	// 获取过去24小时的数据
	now := time.Now().Unix()
	since := now - 86400 // 24小时前

	analytics, err := cfService.GetAnalytics(context.Background(), zoneID, int(since), int(now))
	if err != nil {
		return c.Render("analytics/index", fiber.Map{
			"Error":  "Failed to fetch analytics data",
			"ZoneID": zoneID,
			"Domain": domain,
		})
	}

	// 将数据转换为 JSON 供 Chart.js 使用
	analyticsJSON, _ := json.Marshal(analytics)

	return c.Render("analytics/index", fiber.Map{
		"ZoneID":        zoneID,
		"Domain":        domain,
		"Analytics":     analytics,
		"AnalyticsJSON": string(analyticsJSON),
	})
}
