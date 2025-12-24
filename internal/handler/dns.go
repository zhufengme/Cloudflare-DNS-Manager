package handler

import (
	"context"
	"strconv"

	"github.com/cloudflare/cloudflare-go"
	"github.com/gofiber/fiber/v2"

	"github.com/yourusername/cloudflare-cname-go/internal/service"
)

type DNSHandler struct{}

func NewDNSHandler() *DNSHandler {
	return &DNSHandler{}
}

// ShowAddRecord 显示添加记录页面
func (h *DNSHandler) ShowAddRecord(c *fiber.Ctx) error {
	zoneID := c.Query("zoneid")
	domain := c.Query("domain")

	return c.Render("dns/add", fiber.Map{
		"PageTitle":   "添加 DNS 记录",
		"ShowNav":     true,
		"AppTitle":    "Cloudflare DNS Manager",
		"CurrentPage": "添加记录 - " + domain,
		"ZoneID":      zoneID,
		"Domain":      domain,
	})
}

// AddRecord 添加 DNS 记录
func (h *DNSHandler) AddRecord(c *fiber.Ctx) error {
	email := c.Locals("cloudflare_email").(string)
	apiKey := c.Locals("user_api_key").(string)
	zoneID := c.FormValue("zoneid")
	domain := c.FormValue("domain")

	// 创建 Cloudflare 服务
	cfService, err := service.NewCloudflareService(email, apiKey)
	if err != nil {
		return c.Status(500).SendString("Failed to create Cloudflare service")
	}

	// 构建 DNS 记录参数
	recordType := c.FormValue("type")
	name := c.FormValue("name")
	content := c.FormValue("content")
	ttl, _ := strconv.Atoi(c.FormValue("ttl"))
	proxied := c.FormValue("proxied") == "true"

	params := cloudflare.CreateDNSRecordParams{
		Type:    recordType,
		Name:    name,
		Content: content,
		TTL:     ttl,
	}

	// 只有 A, AAAA, CNAME 可以启用代理
	if recordType == "A" || recordType == "AAAA" || recordType == "CNAME" {
		val := proxied
		params.Proxied = &val
	}

	// 处理 MX 记录优先级
	if recordType == "MX" {
		priority, _ := strconv.Atoi(c.FormValue("priority"))
		prio := uint16(priority)
		params.Priority = &prio
	}

	// 处理 CAA 记录
	if recordType == "CAA" {
		params.Data = map[string]interface{}{
			"tag":   c.FormValue("data_tag"),
			"value": c.FormValue("data_value"),
			"flags": 0,
		}
	}

	// 处理 SRV 记录
	if recordType == "SRV" {
		port, _ := strconv.Atoi(c.FormValue("srv_port"))
		priority, _ := strconv.Atoi(c.FormValue("srv_priority"))
		weight, _ := strconv.Atoi(c.FormValue("srv_weight"))

		params.Data = map[string]interface{}{
			"service":  c.FormValue("srv_service"),
			"proto":    c.FormValue("srv_proto"),
			"name":     name,
			"port":     port,
			"priority": priority,
			"weight":   weight,
			"target":   c.FormValue("srv_target"),
		}
	}

	// 创建记录
	rc := cloudflare.ZoneIdentifier(zoneID)
	_, err = cfService.CreateDNSRecord(context.Background(), rc, params)
	if err != nil {
		return c.Render("dns/add", fiber.Map{
			"PageTitle":   "添加 DNS 记录",
			"ShowNav":     true,
			"AppTitle":    "Cloudflare DNS Manager",
			"CurrentPage": "添加记录 - " + domain,
			"ZoneID":      zoneID,
			"Domain":      domain,
			"Error":       "Failed to add record: " + err.Error(),
		})
	}

	return c.Redirect("/zone?zoneid=" + zoneID + "&domain=" + domain)
}

// ShowEditRecord 显示编辑记录页面
func (h *DNSHandler) ShowEditRecord(c *fiber.Ctx) error {
	email := c.Locals("cloudflare_email").(string)
	apiKey := c.Locals("user_api_key").(string)
	zoneID := c.Query("zoneid")
	domain := c.Query("domain")
	recordID := c.Query("recordid")

	// 创建 Cloudflare 服务
	cfService, err := service.NewCloudflareService(email, apiKey)
	if err != nil {
		return c.Status(500).SendString("Failed to create Cloudflare service")
	}

	// 获取记录详情
	rc := cloudflare.ZoneIdentifier(zoneID)
	record, err := cfService.GetDNSRecord(context.Background(), rc, recordID)
	if err != nil {
		return c.Status(500).SendString("Failed to fetch record: " + err.Error())
	}

	return c.Render("dns/edit", fiber.Map{
		"PageTitle":   "编辑 DNS 记录",
		"ShowNav":     true,
		"AppTitle":    "Cloudflare DNS Manager",
		"CurrentPage": "编辑记录 - " + domain,
		"ZoneID":      zoneID,
		"Domain":      domain,
		"Record":      record,
	})
}

// EditRecord 编辑 DNS 记录
func (h *DNSHandler) EditRecord(c *fiber.Ctx) error {
	email := c.Locals("cloudflare_email").(string)
	apiKey := c.Locals("user_api_key").(string)
	zoneID := c.FormValue("zoneid")
	domain := c.FormValue("domain")
	recordID := c.FormValue("recordid")

	// 创建 Cloudflare 服务
	cfService, err := service.NewCloudflareService(email, apiKey)
	if err != nil {
		return c.Status(500).SendString("Failed to create Cloudflare service")
	}

	// 获取原记录
	rc := cloudflare.ZoneIdentifier(zoneID)
	record, err := cfService.GetDNSRecord(context.Background(), rc, recordID)
	if err != nil {
		return c.Status(500).SendString("Failed to fetch record")
	}

	// 构建更新参数
	name := c.FormValue("name")
	content := c.FormValue("content")
	ttl, _ := strconv.Atoi(c.FormValue("ttl"))
	proxied := c.FormValue("proxied") == "true"

	params := cloudflare.UpdateDNSRecordParams{
		ID:      recordID,
		Type:    record.Type,
		Name:    name,
		Content: content,
		TTL:     ttl,
	}

	// 只有 A, AAAA, CNAME 可以启用代理
	if record.Type == "A" || record.Type == "AAAA" || record.Type == "CNAME" {
		val := proxied
		params.Proxied = &val
	}

	// 处理 MX 记录
	if record.Type == "MX" {
		priority, _ := strconv.Atoi(c.FormValue("priority"))
		prio := uint16(priority)
		params.Priority = &prio
	}

	// 更新记录
	_, err = cfService.UpdateDNSRecord(context.Background(), rc, params)
	if err != nil {
		return c.Render("dns/edit", fiber.Map{
			"PageTitle":   "编辑 DNS 记录",
			"ShowNav":     true,
			"AppTitle":    "Cloudflare DNS Manager",
			"CurrentPage": "编辑记录 - " + domain,
			"ZoneID":      zoneID,
			"Domain":      domain,
			"Record":      record,
			"Error":       "Failed to update record: " + err.Error(),
		})
	}

	return c.Redirect("/zone?zoneid=" + zoneID + "&domain=" + domain)
}

// DeleteRecord 删除 DNS 记录
func (h *DNSHandler) DeleteRecord(c *fiber.Ctx) error {
	email := c.Locals("cloudflare_email").(string)
	apiKey := c.Locals("user_api_key").(string)
	zoneID := c.Query("zoneid")
	domain := c.Query("domain")
	recordID := c.Query("delete")

	// 创建 Cloudflare 服务
	cfService, err := service.NewCloudflareService(email, apiKey)
	if err != nil {
		return c.Status(500).SendString("Failed to create Cloudflare service")
	}

	// 删除记录
	rc := cloudflare.ZoneIdentifier(zoneID)
	err = cfService.DeleteDNSRecord(context.Background(), rc, recordID)
	if err != nil {
		return c.SendString("Failed to delete record: " + err.Error())
	}

	return c.Redirect("/zone?zoneid=" + zoneID + "&domain=" + domain)
}

// ToggleProxy HTMX API：切换 CDN 代理
func (h *DNSHandler) ToggleProxy(c *fiber.Ctx) error {
	email := c.Locals("cloudflare_email").(string)
	apiKey := c.Locals("user_api_key").(string)
	zoneID := c.Query("zoneid")
	recordID := c.Params("id")

	// 创建 Cloudflare 服务
	cfService, err := service.NewCloudflareService(email, apiKey)
	if err != nil {
		return c.Status(500).SendString("Error")
	}

	// 获取当前记录
	rc := cloudflare.ZoneIdentifier(zoneID)
	record, err := cfService.GetDNSRecord(context.Background(), rc, recordID)
	if err != nil {
		return c.Status(500).SendString("Error")
	}

	// 切换代理状态
	newProxied := true
	if record.Proxied != nil {
		newProxied = !*record.Proxied
	}

	// 更新记录
	params := cloudflare.UpdateDNSRecordParams{
		ID:      recordID,
		Type:    record.Type,
		Name:    record.Name,
		Content: record.Content,
		TTL:     record.TTL,
		Proxied: &newProxied,
	}

	_, err = cfService.UpdateDNSRecord(context.Background(), rc, params)
	if err != nil {
		return c.Status(500).SendString("Error")
	}

	// 返回更新后的图标 HTML
	imgPath := "/static/images/cloud_off.png"
	height := "30"
	if newProxied {
		imgPath = "/static/images/cloud_on.png"
		height = "19"
	}

	return c.SendString(`<img src="` + imgPath + `" height="` + height + `" hx-post="/api/dns/` + recordID + `/toggle-proxy?zoneid=` + zoneID + `" hx-trigger="click" hx-swap="outerHTML" style="cursor:pointer;" />`)
}
