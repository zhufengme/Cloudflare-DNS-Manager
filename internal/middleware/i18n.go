package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/zhufengme/Cloudflare-DNS-Manager/internal/i18n"
)

// I18n 中间件：检测语言并注入 Localizer
func I18n(c *fiber.Ctx) error {
	// 从 Accept-Language header 检测语言
	acceptLang := c.Get("Accept-Language")
	lang := detectLanguage(acceptLang)

	// 获取对应的 Localizer
	localizer := i18n.GetLocalizer(lang)

	// 注入到 Locals，供后续 handler 使用
	c.Locals("lang", lang)
	c.Locals("localizer", localizer)

	// 创建翻译函数并注入到视图变量
	c.Locals("T", func(messageID string) string {
		return i18n.T(localizer, messageID)
	})

	return c.Next()
}

// detectLanguage 从 Accept-Language header 检测语言
func detectLanguage(acceptLang string) string {
	// 默认英文
	if acceptLang == "" {
		return "en"
	}

	// 解析 Accept-Language (简化版本)
	// 例如: "zh-CN,zh;q=0.9,en;q=0.8"
	parts := strings.Split(acceptLang, ",")
	if len(parts) > 0 {
		firstLang := strings.TrimSpace(parts[0])
		// 提取主语言代码
		if strings.HasPrefix(strings.ToLower(firstLang), "zh") {
			return "zh"
		}
		if strings.HasPrefix(strings.ToLower(firstLang), "en") {
			return "en"
		}
	}

	return "en" // 默认英文
}
