package handler

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/gofiber/fiber/v2"
	"github.com/miekg/dns"

	"github.com/zhufengme/Cloudflare-DNS-Manager/internal/service"
)

type ZoneHandler struct{}

func NewZoneHandler() *ZoneHandler {
	return &ZoneHandler{}
}

// ListZones 域名列表页面
func (h *ZoneHandler) ListZones(c *fiber.Ctx) error {
	email := c.Locals("cloudflare_email").(string)
	apiKey := c.Locals("user_api_key").(string)

	// 创建 Cloudflare 服务
	cfService, err := service.NewCloudflareService(email, apiKey)
	if err != nil {
		return c.Status(500).SendString("Failed to create Cloudflare service")
	}

	// 获取页码
	page := 1
	if p := c.Query("page"); p != "" {
		page, _ = strconv.Atoi(p)
	}

	// 获取域名列表
	zones, resultInfo, err := cfService.ListZones(context.Background(), page)
	if err != nil {
		return c.Status(500).SendString("Failed to fetch zones: " + err.Error())
	}

	return c.Render("zone/list", fiber.Map{
		"PageTitle":  "域名列表",
		"ShowNav":    true,
		"AppTitle":   "Cloudflare DNS Manager",
		"CurrentPage": "域名列表",
		"Zones":      zones,
		"ResultInfo": resultInfo,
		"Page":       page,
	})
}

// ShowAddZone 显示添加域名页面
func (h *ZoneHandler) ShowAddZone(c *fiber.Ctx) error {
	return c.Render("zone/add", fiber.Map{
		"PageTitle":   "添加域名",
		"ShowNav":     true,
		"AppTitle":    "Cloudflare DNS Manager",
		"CurrentPage": "添加域名",
	})
}

// AddZone 添加域名
func (h *ZoneHandler) AddZone(c *fiber.Ctx) error {
	email := c.Locals("cloudflare_email").(string)
	apiKey := c.Locals("user_api_key").(string)
	zoneName := c.FormValue("zone_name")

	if zoneName == "" {
		return c.Render("zone/add", fiber.Map{
			"PageTitle":   "添加域名",
			"ShowNav":     true,
			"AppTitle":    "Cloudflare DNS Manager",
			"CurrentPage": "添加域名",
			"Error":       "域名不能为空",
		})
	}

	// 创建 Cloudflare 服务
	cfService, err := service.NewCloudflareService(email, apiKey)
	if err != nil {
		return c.Render("zone/add", fiber.Map{
			"PageTitle":   "添加域名",
			"ShowNav":     true,
			"AppTitle":    "Cloudflare DNS Manager",
			"CurrentPage": "添加域名",
			"Error":       "Failed to create Cloudflare service",
		})
	}

	// 添加域名
	zone, err := cfService.CreateZone(context.Background(), zoneName)
	if err != nil {
		return c.Render("zone/add", fiber.Map{
			"PageTitle":   "添加域名",
			"ShowNav":     true,
			"AppTitle":    "Cloudflare DNS Manager",
			"CurrentPage": "添加域名",
			"Error":       "Failed to add zone: " + err.Error(),
		})
	}

	return c.Render("zone/add", fiber.Map{
		"PageTitle":   "添加域名",
		"ShowNav":     true,
		"AppTitle":    "Cloudflare DNS Manager",
		"CurrentPage": "添加域名",
		"Success":     "域名添加成功！",
		"Zone":        zone,
	})
}

// ShowZone 显示域名详情（DNS 记录管理）
func (h *ZoneHandler) ShowZone(c *fiber.Ctx) error {
	email := c.Locals("cloudflare_email").(string)
	apiKey := c.Locals("user_api_key").(string)
	zoneID := c.Query("zoneid")
	domain := c.Query("domain")

	if zoneID == "" || domain == "" {
		return c.Redirect("/zones")
	}

	// 创建 Cloudflare 服务
	cfService, err := service.NewCloudflareService(email, apiKey)
	if err != nil {
		return c.Status(500).SendString("Failed to create Cloudflare service")
	}

	// 获取页码
	page := 1
	if p := c.Query("page"); p != "" {
		page, _ = strconv.Atoi(p)
	}

	// 获取 DNS 记录
	rc := cloudflare.ZoneIdentifier(zoneID)
	records, resultInfo, err := cfService.ListDNSRecords(context.Background(), rc, cloudflare.ListDNSRecordsParams{
		ResultInfo: cloudflare.ResultInfo{
			Page:    page,
			PerPage: 20,
		},
	})
	if err != nil {
		return c.Status(500).SendString("Failed to fetch DNS records: " + err.Error())
	}

	// 获取 Zone 信息（包含 NS 记录和类型）
	zone, err := cfService.GetZone(context.Background(), zoneID)
	var nsRecords []string
	var zoneType string
	if err == nil {
		if zone.NameServers != nil {
			nsRecords = zone.NameServers
		}
		zoneType = zone.Type
	}

	// 只有 CNAME Setup (partial) 模式才查询 Anycast IP
	// Full Setup 模式通过 NS 接入，不需要 CNAME 设置
	var anycastIPs map[string]interface{}
	if zoneType == "partial" && len(records) > 0 {
		firstRecord := records[0].Name
		anycastIPs = queryAnycastIPs(firstRecord)
	}

	return c.Render("zone/manage", fiber.Map{
		"PageTitle":   domain,
		"ShowNav":     true,
		"AppTitle":    "Cloudflare DNS Manager",
		"CurrentPage": domain,
		"ZoneID":      zoneID,
		"Domain":      domain,
		"Records":     records,
		"ResultInfo":  resultInfo,
		"Page":        page,
		"NSRecords":   nsRecords,
		"AnycastIPs":  anycastIPs,
		"ZoneType":    zoneType, // 传递 zone type 到模板
	})
}

// queryAnycastIPs 查询 Anycast IP 地址
// 通过 DNS 查询 hostname.cdn.cloudflare.net 的 A/AAAA 记录
// 如果能查询到 IP，说明是 CNAME Setup 模式（合作伙伴模式）
// 如果查询不到，说明是 Full Setup 模式，不显示 CNAME 设置
func queryAnycastIPs(hostname string) map[string]interface{} {
	result := make(map[string]interface{})

	// 构建查询目标：hostname.cdn.cloudflare.net
	target := hostname + ".cdn.cloudflare.net"

	// 使用 Cloudflare DNS 服务器查询
	nameservers := []string{"173.245.59.31:53", "[2400:cb00:2049:1::adf5:3b1f]:53"}

	// 查询 IPv4 (A 记录)
	ipv4List := queryDNS(target, "A", nameservers)

	// 查询 IPv6 (AAAA 记录)
	ipv6List := queryDNS(target, "AAAA", nameservers)

	// 只有当至少有 2 个 IPv4 或 2 个 IPv6 地址时，才认为是 CNAME 模式
	if len(ipv4List) >= 2 {
		result["IPv4"] = ipv4List
	}
	if len(ipv6List) >= 2 {
		result["IPv6"] = ipv6List
	}

	return result
}

// queryDNS 执行 DNS 查询
func queryDNS(target, recordType string, nameservers []string) []string {
	var results []string

	c := new(dns.Client)
	c.Timeout = 2 * time.Second

	var qtype uint16
	if recordType == "A" {
		qtype = dns.TypeA
	} else if recordType == "AAAA" {
		qtype = dns.TypeAAAA
	} else {
		return results
	}

	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(target), qtype)

	// 尝试每个 nameserver
	for _, ns := range nameservers {
		r, _, err := c.Exchange(m, ns)
		if err != nil {
			continue
		}

		// 解析响应
		if r != nil && r.Rcode == dns.RcodeSuccess {
			for _, ans := range r.Answer {
				if recordType == "A" {
					if a, ok := ans.(*dns.A); ok {
						results = append(results, a.A.String())
					}
				} else if recordType == "AAAA" {
					if aaaa, ok := ans.(*dns.AAAA); ok {
						results = append(results, aaaa.AAAA.String())
					}
				}
			}

			// 如果找到了结果就返回
			if len(results) > 0 {
				break
			}
		}
	}

	return results
}

// SearchDNSRecords 搜索和过滤 DNS 记录
func (h *ZoneHandler) SearchDNSRecords(c *fiber.Ctx) error {
	email := c.Locals("cloudflare_email").(string)
	apiKey := c.Locals("user_api_key").(string)
	zoneID := c.Query("zoneid")
	domain := c.Query("domain")     // 添加 domain 参数
	query := c.Query("query")       // 搜索关键词
	recordType := c.Query("type")   // 记录类型过滤
	proxied := c.Query("proxied")   // CDN 状态过滤

	if zoneID == "" {
		return c.Status(400).SendString("Missing zoneid")
	}

	// 创建 Cloudflare 服务
	cfService, err := service.NewCloudflareService(email, apiKey)
	if err != nil {
		return c.Status(500).SendString("Failed to create Cloudflare service")
	}

	// 构建查询参数
	params := cloudflare.ListDNSRecordsParams{
		ResultInfo: cloudflare.ResultInfo{
			PerPage: 100, // 搜索时返回更多结果
		},
	}

	// 添加类型过滤
	if recordType != "" {
		params.Type = recordType
	}

	// 添加 CDN 状态过滤
	if proxied == "true" {
		proxiedBool := true
		params.Proxied = &proxiedBool
	} else if proxied == "false" {
		proxiedBool := false
		params.Proxied = &proxiedBool
	}

	// 获取记录
	rc := cloudflare.ZoneIdentifier(zoneID)
	records, _, err := cfService.ListDNSRecords(context.Background(), rc, params)
	if err != nil {
		return c.Status(500).SendString("Failed to fetch DNS records: " + err.Error())
	}

	// 如果有搜索关键词，进行本地过滤
	if query != "" {
		query = strings.ToLower(query)
		var filtered []cloudflare.DNSRecord
		for _, record := range records {
			// 搜索名称或内容
			if strings.Contains(strings.ToLower(record.Name), query) ||
				strings.Contains(strings.ToLower(record.Content), query) {
				filtered = append(filtered, record)
			}
		}
		records = filtered
	}

	// 渲染记录列表片段
	return c.Render("zone/partials/records-table", fiber.Map{
		"Records": records,
		"ZoneID":  zoneID,
		"Domain":  domain,
	}, "")
}

// GetDNSStats 获取 DNS 记录统计信息
func (h *ZoneHandler) GetDNSStats(c *fiber.Ctx) error {
	email := c.Locals("cloudflare_email").(string)
	apiKey := c.Locals("user_api_key").(string)
	zoneID := c.Query("zoneid")

	if zoneID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Missing zoneid"})
	}

	// 创建 Cloudflare 服务
	cfService, err := service.NewCloudflareService(email, apiKey)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create service"})
	}

	// 获取所有记录
	rc := cloudflare.ZoneIdentifier(zoneID)
	records, _, err := cfService.ListDNSRecords(context.Background(), rc, cloudflare.ListDNSRecordsParams{
		ResultInfo: cloudflare.ResultInfo{
			PerPage: 1000,
		},
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// 统计数据
	stats := map[string]interface{}{
		"total":         len(records),
		"proxied_count": 0,
		"type_counts":   make(map[string]int),
	}

	typeCounts := make(map[string]int)
	proxiedCount := 0

	for _, record := range records {
		// 统计类型
		typeCounts[record.Type]++

		// 统计 CDN 启用数量
		if record.Proxied != nil && *record.Proxied {
			proxiedCount++
		}
	}

	stats["proxied_count"] = proxiedCount
	stats["type_counts"] = typeCounts

	// 渲染统计卡片
	return c.Render("zone/partials/stats-card", stats, "")
}

// DeleteZone 删除域名
func (h *ZoneHandler) DeleteZone(c *fiber.Ctx) error {
	zoneID := c.Query("zoneid")
	domain := c.FormValue("domain")

	if zoneID == "" || domain == "" {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "缺少必要参数",
		})
	}

	// 获取 Cloudflare Service
	email := c.Locals("cloudflare_email").(string)
	apiKey := c.Locals("user_api_key").(string)

	cfService, err := service.NewCloudflareService(email, apiKey)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "获取服务失败",
		})
	}

	// 验证域名是否匹配（双重验证）
	ctx := context.Background()
	zone, err := cfService.GetZone(ctx, zoneID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "获取域名信息失败",
		})
	}

	// 严格验证域名匹配
	if zone.Name != domain {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "域名不匹配，删除失败",
		})
	}

	// 执行删除
	err = cfService.DeleteZone(ctx, zoneID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "删除域名失败: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "域名已成功删除",
	})
}
