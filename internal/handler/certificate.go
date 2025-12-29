package handler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/zhufengme/Cloudflare-DNS-Manager/internal/service"
)

type CertificateHandler struct{}

func NewCertificateHandler() *CertificateHandler {
	return &CertificateHandler{}
}

// ShowCertificates 显示证书管理页面
func (h *CertificateHandler) ShowCertificates(c *fiber.Ctx) error {
	email := c.Locals("cloudflare_email").(string)
	apiKey := c.Locals("user_api_key").(string)
	zoneID := c.Query("zoneid")
	domain := c.Query("domain")
	tab := c.Query("tab", "edge") // 默认显示边缘证书

	if zoneID == "" || domain == "" {
		return c.Redirect("/zones")
	}

	cfService, err := service.NewCloudflareService(email, apiKey)
	if err != nil {
		return c.Status(500).SendString("Failed to create Cloudflare service")
	}

	data := fiber.Map{
		"ZoneID": zoneID,
		"Domain": domain,
		"Tab":    tab,
	}

	// 根据选择的标签页加载对应数据
	switch tab {
	case "edge":
		// 获取边缘证书
		edgeCerts, err := cfService.ListEdgeCertificates(context.Background(), zoneID)
		if err != nil {
			data["Error"] = "Failed to fetch edge certificates: " + err.Error()
		} else {
			data["EdgeCertificates"] = edgeCerts
		}

	case "origin":
		// 获取回源证书
		originCerts, err := cfService.ListOriginCertificates(context.Background(), zoneID)
		if err != nil {
			data["Error"] = "Failed to fetch origin certificates: " + err.Error()
		} else {
			data["OriginCertificates"] = originCerts
		}

	case "custom":
		// 获取自定义证书
		customCerts, err := cfService.ListCustomSSLCertificates(context.Background(), zoneID)
		if err != nil {
			data["Error"] = "Failed to fetch custom certificates: " + err.Error()
		} else {
			data["CustomCertificates"] = customCerts
		}
	}

	return c.Render("certificate/index", data)
}

// DownloadOriginCertificate 下载回源证书
func (h *CertificateHandler) DownloadOriginCertificate(c *fiber.Ctx) error {
	email := c.Locals("cloudflare_email").(string)
	apiKey := c.Locals("user_api_key").(string)
	certID := c.Params("id")

	if certID == "" {
		return c.Status(400).SendString("Missing certificate ID")
	}

	cfService, err := service.NewCloudflareService(email, apiKey)
	if err != nil {
		return c.Status(500).SendString("Failed to create service")
	}

	// 获取证书详情
	cert, err := cfService.GetOriginCertificate(context.Background(), certID)
	if err != nil {
		return c.Status(500).SendString("Failed to fetch certificate: " + err.Error())
	}

	// 设置下载响应头
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=origin-cert-%s.pem", certID[:8]))
	c.Set("Content-Type", "application/x-pem-file")

	return c.SendString(cert.Certificate)
}

// CreateOriginCertificate 创建回源证书
func (h *CertificateHandler) CreateOriginCertificate(c *fiber.Ctx) error {
	email := c.Locals("cloudflare_email").(string)
	apiKey := c.Locals("user_api_key").(string)

	// 解析表单数据
	hostnamesStr := c.FormValue("hostnames")
	requestType := c.FormValue("request_type")
	validityDays := c.FormValue("requested_validity")

	if hostnamesStr == "" || requestType == "" || validityDays == "" {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "缺少必填字段：域名、证书类型或有效期",
		})
	}

	// 解析主机名列表
	hostnames := strings.Split(hostnamesStr, ",")
	var cleanedHostnames []string
	for _, host := range hostnames {
		trimmed := strings.TrimSpace(host)
		if trimmed != "" {
			cleanedHostnames = append(cleanedHostnames, trimmed)
		}
	}

	if len(cleanedHostnames) == 0 {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "至少需要一个域名",
		})
	}

	// 转换有效期为整数
	var validity int
	_, err := fmt.Sscanf(validityDays, "%d", &validity)
	if err != nil || validity <= 0 {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "无效的有效期参数",
		})
	}

	cfService, err := service.NewCloudflareService(email, apiKey)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "创建 Cloudflare 服务失败",
		})
	}

	// 创建证书
	cert, err := cfService.CreateOriginCertificate(context.Background(), cleanedHostnames, requestType, validity)
	if err != nil {
		// 详细的错误信息
		errorMsg := err.Error()
		if strings.Contains(errorMsg, "1007") {
			errorMsg = "无效的域名格式。请确保域名正确且属于当前账户管理的区域。"
		}
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   fmt.Sprintf("创建证书失败: %s", errorMsg),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "回源证书创建成功",
		"cert_id": cert.ID,
	})
}

// RevokeOriginCertificate 撤销回源证书
func (h *CertificateHandler) RevokeOriginCertificate(c *fiber.Ctx) error {
	email := c.Locals("cloudflare_email").(string)
	apiKey := c.Locals("user_api_key").(string)
	certID := c.Params("id")

	if certID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Missing certificate ID"})
	}

	cfService, err := service.NewCloudflareService(email, apiKey)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create service"})
	}

	err = cfService.RevokeOriginCertificate(context.Background(), certID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "证书已撤销",
	})
}

// GetEdgeCertificateDetails 获取边缘证书详情（HTMX）
func (h *CertificateHandler) GetEdgeCertificateDetails(c *fiber.Ctx) error {
	email := c.Locals("cloudflare_email").(string)
	apiKey := c.Locals("user_api_key").(string)
	zoneID := c.Query("zoneid")
	certID := c.Params("id")

	if zoneID == "" || certID == "" {
		return c.Status(400).SendString("Missing parameters")
	}

	cfService, err := service.NewCloudflareService(email, apiKey)
	if err != nil {
		return c.Status(500).SendString("Failed to create service")
	}

	cert, err := cfService.GetEdgeCertificate(context.Background(), zoneID, certID)
	if err != nil {
		return c.Status(500).SendString("Failed to fetch certificate: " + err.Error())
	}

	return c.Render("certificate/partials/edge-details", fiber.Map{
		"Certificate": cert,
	}, "")
}

// Helper functions

// daysUntil 计算距离某个日期还有多少天
func daysUntil(t time.Time) int {
	duration := time.Until(t)
	return int(duration.Hours() / 24)
}
