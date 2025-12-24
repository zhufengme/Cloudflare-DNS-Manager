package handler

import (
	"context"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/cloudflare-cname-go/internal/service"
)

type SettingsHandler struct{}

func NewSettingsHandler() *SettingsHandler {
	return &SettingsHandler{}
}

// ShowSettings 显示 Zone 设置页面
func (h *SettingsHandler) ShowSettings(c *fiber.Ctx) error {
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
		return c.Status(500).SendString("Failed to initialize Cloudflare service")
	}

	// 获取所有 Zone 设置
	settings, err := cfService.GetZoneSettings(context.Background(), zoneID)
	if err != nil {
		return c.Render("settings/index", fiber.Map{
			"Error":  "Failed to fetch zone settings: " + err.Error(),
			"ZoneID": zoneID,
			"Domain": domain,
		})
	}

	// 将设置转换为 map 方便模板使用
	settingsMap := make(map[string]interface{})
	for _, setting := range settings {
		settingsMap[setting.ID] = setting.Value
	}

	return c.Render("settings/index", fiber.Map{
		"ZoneID":   zoneID,
		"Domain":   domain,
		"Settings": settingsMap,
	})
}

// ToggleDevelopmentMode 切换开发模式
func (h *SettingsHandler) ToggleDevelopmentMode(c *fiber.Ctx) error {
	zoneID := c.Query("zoneid")
	currentValue := c.FormValue("current")

	if zoneID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Missing zoneid"})
	}

	email := c.Locals("cloudflare_email").(string)
	apiKey := c.Locals("user_api_key").(string)

	cfService, err := service.NewCloudflareService(email, apiKey)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create service"})
	}

	// 切换值
	newValue := "off"
	if currentValue == "off" {
		newValue = "on"
	}

	err = cfService.UpdateZoneSetting(context.Background(), zoneID, "development_mode", newValue)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"value":   newValue,
		"message": "开发模式已" + map[string]string{"on": "启用", "off": "禁用"}[newValue],
	})
}

// UpdateSetting 更新单个设置
func (h *SettingsHandler) UpdateSetting(c *fiber.Ctx) error {
	zoneID := c.Query("zoneid")
	settingID := c.Params("setting")
	value := c.FormValue("value")

	if zoneID == "" || settingID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Missing parameters"})
	}

	email := c.Locals("cloudflare_email").(string)
	apiKey := c.Locals("user_api_key").(string)

	cfService, err := service.NewCloudflareService(email, apiKey)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create service"})
	}

	// 根据设置类型转换值
	var settingValue interface{} = value

	// 对于特殊的设置，需要转换为对象
	if settingID == "minify" {
		// minify 需要是一个对象 {css: "on", html: "on", js: "on"}
		minifyValue := make(map[string]string)
		if strings.Contains(value, "css") {
			minifyValue["css"] = "on"
		} else {
			minifyValue["css"] = "off"
		}
		if strings.Contains(value, "html") {
			minifyValue["html"] = "on"
		} else {
			minifyValue["html"] = "off"
		}
		if strings.Contains(value, "js") {
			minifyValue["js"] = "on"
		} else {
			minifyValue["js"] = "off"
		}
		settingValue = minifyValue
	}

	err = cfService.UpdateZoneSetting(context.Background(), zoneID, settingID, settingValue)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "设置已更新",
	})
}

// PurgeCache 清除缓存
func (h *SettingsHandler) PurgeCache(c *fiber.Ctx) error {
	zoneID := c.Query("zoneid")
	purgeType := c.FormValue("type") // "all", "urls", "hosts", "prefixes", "tags"
	content := c.FormValue("content") // 统一的内容字段

	if zoneID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Missing zoneid"})
	}

	email := c.Locals("cloudflare_email").(string)
	apiKey := c.Locals("user_api_key").(string)

	cfService, err := service.NewCloudflareService(email, apiKey)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create service"})
	}

	// 清除所有缓存
	if purgeType == "all" {
		err = cfService.PurgeAllCache(context.Background(), zoneID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{
			"success": true,
			"message": "所有缓存已清除",
		})
	}

	// 其他清除方式需要内容
	if content == "" {
		return c.Status(400).JSON(fiber.Map{"error": "请输入要清除的内容"})
	}

	// 按行分割内容
	lines := strings.Split(content, "\n")
	var cleanedItems []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			cleanedItems = append(cleanedItems, trimmed)
		}
	}

	if len(cleanedItems) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "没有有效的内容"})
	}

	var message string

	switch purgeType {
	case "urls":
		err = cfService.PurgeCacheByURLs(context.Background(), zoneID, cleanedItems)
		message = fmt.Sprintf("已清除 %d 个 URL 的缓存", len(cleanedItems))
	case "hosts":
		err = cfService.PurgeCacheByHosts(context.Background(), zoneID, cleanedItems)
		message = fmt.Sprintf("已清除 %d 个主机名的缓存", len(cleanedItems))
	case "prefixes":
		err = cfService.PurgeCacheByPrefixes(context.Background(), zoneID, cleanedItems)
		message = fmt.Sprintf("已清除 %d 个前缀的缓存", len(cleanedItems))
	case "tags":
		err = cfService.PurgeCacheByTags(context.Background(), zoneID, cleanedItems)
		message = fmt.Sprintf("已清除 %d 个 Tag 的缓存", len(cleanedItems))
	default:
		return c.Status(400).JSON(fiber.Map{"error": "Invalid purge type"})
	}

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": message,
	})
}

// ApplyPreset 应用配置预设
func (h *SettingsHandler) ApplyPreset(c *fiber.Ctx) error {
	zoneID := c.Query("zoneid")
	presetName := c.FormValue("preset")

	if zoneID == "" || presetName == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Missing parameters"})
	}

	email := c.Locals("cloudflare_email").(string)
	apiKey := c.Locals("user_api_key").(string)

	cfService, err := service.NewCloudflareService(email, apiKey)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create service"})
	}

	// 应用预设
	err = cfService.ApplyPreset(context.Background(), zoneID, presetName)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// 获取预设信息
	preset, _ := service.GetPresetInfo(presetName)

	return c.JSON(fiber.Map{
		"success": true,
		"message": fmt.Sprintf("已应用「%s」配置模板", preset.Name),
	})
}
