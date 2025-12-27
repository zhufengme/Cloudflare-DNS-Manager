package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html/v2"

	"github.com/zhufengme/Cloudflare-DNS-Manager/internal/config"
	"github.com/zhufengme/Cloudflare-DNS-Manager/internal/handler"
	"github.com/zhufengme/Cloudflare-DNS-Manager/internal/i18n"
	"github.com/zhufengme/Cloudflare-DNS-Manager/internal/middleware"
)

//go:embed web
var webFS embed.FS

func main() {
	// 命令行参数
	configFile := flag.String("config", "config.yaml", "配置文件路径")
	flag.Parse()

	// 加载配置
	cfg, err := config.Load(*configFile)
	if err != nil {
		log.Printf("Warning: Failed to load %s, using defaults: %v", *configFile, err)
		cfg = &config.Config{}
		cfg.Server.Host = "0.0.0.0"
		cfg.Server.Port = 8080
		cfg.Server.PageTitle = "Cloudflare DNS Manager"
		cfg.Session.Expire = 3600
		cfg.Session.RememberExpire = 31536000
		cfg.RateLimit.MaxAttempts = 5
		cfg.RateLimit.Window = 60
		cfg.Cache.DNSTTL = 172800
	}

	// 初始化会话存储
	middleware.InitSession(time.Duration(cfg.Session.Expire) * time.Second)

	// 初始化 i18n
	if err := i18n.Init(webFS); err != nil {
		log.Fatalf("Failed to initialize i18n: %v", err)
	}

	// 创建模板引擎
	templateFS, _ := fs.Sub(webFS, "web/templates")
	engine := html.NewFileSystem(http.FS(templateFS), ".html")
	engine.Reload(true) // 开发模式重载
	// 不设置全局布局，每个模板独立定义
	engine.AddFunc("add", func(a, b int) int { return a + b })
	engine.AddFunc("sub", func(a, b int) int { return a - b })
	engine.AddFunc("default", func(defaultVal interface{}, val interface{}) interface{} {
		if val == nil {
			return defaultVal
		}
		return val
	})
	engine.AddFunc("dict", func(values ...interface{}) (map[string]interface{}, error) {
		if len(values)%2 != 0 {
			return nil, fmt.Errorf("invalid dict call")
		}
		dict := make(map[string]interface{}, len(values)/2)
		for i := 0; i < len(values); i += 2 {
			key, ok := values[i].(string)
			if !ok {
				return nil, fmt.Errorf("dict keys must be strings")
			}
			dict[key] = values[i+1]
		}
		return dict, nil
	})
	engine.AddFunc("daysUntil", func(t time.Time) int {
		duration := time.Until(t)
		return int(duration.Hours() / 24)
	})

	// 创建 Fiber 应用
	app := fiber.New(fiber.Config{
		Views:             engine,
		PassLocalsToViews: true,
		ServerHeader:      "Cloudflare DNS Manager",
		AppName:           cfg.Server.PageTitle,
	})

	// 中间件
	app.Use(recover.New())
	if cfg.Server.Debug {
		app.Use(logger.New())
	}
	app.Use(middleware.I18n) // i18n 中间件

	// 静态文件服务（从嵌入资源）
	staticFS, _ := fs.Sub(webFS, "web/static")
	app.Use("/static", filesystem.New(filesystem.Config{
		Root:   http.FS(staticFS),
		Browse: false,
	}))

	// 路由
	setupRoutes(app, cfg)

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Starting server on %s", addr)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func setupRoutes(app *fiber.App, cfg *config.Config) {
	// 创建限流器
	rateLimiter := middleware.NewRateLimiter(
		cfg.RateLimit.MaxAttempts,
		time.Duration(cfg.RateLimit.Window)*time.Minute,
	)

	// 创建 handler
	homeHandler := handler.NewHomeHandler()
	authHandler := handler.NewAuthHandler(rateLimiter)
	zoneHandler := handler.NewZoneHandler()
	dnsHandler := handler.NewDNSHandler()
	securityHandler := handler.NewSecurityHandler()
	settingsHandler := handler.NewSettingsHandler()
	certificateHandler := handler.NewCertificateHandler()
	// analyticsHandler := handler.NewAnalyticsHandler() // Analytics 功能已移除

	// 首页 - 显示 landing page 或跳转到域名列表
	app.Get("/", homeHandler.ShowHome)

	// 公开路由 - 登录/登出
	app.Get("/login", authHandler.ShowLogin)
	app.Post("/login", authHandler.PostLogin)
	app.Get("/logout", authHandler.Logout)

	// 受保护的路由
	protected := app.Group("/", middleware.AuthRequired)

	// 域名管理路由
	protected.Get("/zones", zoneHandler.ListZones)
	protected.Get("/zone/add", zoneHandler.ShowAddZone)
	protected.Post("/zone/add", zoneHandler.AddZone)
	protected.Get("/zone", zoneHandler.ShowZone)
	protected.Post("/api/zone/delete", zoneHandler.DeleteZone)

	// DNS 记录管理路由
	protected.Get("/dns/add", dnsHandler.ShowAddRecord)
	protected.Post("/dns/add", dnsHandler.AddRecord)
	protected.Get("/dns/edit", dnsHandler.ShowEditRecord)
	protected.Post("/dns/edit", dnsHandler.EditRecord)
	protected.Get("/dns/delete", dnsHandler.DeleteRecord)

	// HTMX API 端点
	protected.Post("/api/dns/:id/toggle-proxy", dnsHandler.ToggleProxy)
	protected.Get("/api/dns/search", zoneHandler.SearchDNSRecords)
	protected.Get("/api/dns/stats", zoneHandler.GetDNSStats)

	// 安全功能路由
	protected.Get("/security", securityHandler.ShowSecurity)
	protected.Post("/security/dnssec", securityHandler.ToggleDNSSEC)

	// Zone 设置路由
	protected.Get("/settings", settingsHandler.ShowSettings)
	protected.Post("/api/settings/development_mode/toggle", settingsHandler.ToggleDevelopmentMode)
	protected.Post("/api/settings/:setting/update", settingsHandler.UpdateSetting)
	protected.Post("/api/cache/purge", settingsHandler.PurgeCache)
	protected.Post("/api/settings/preset/apply", settingsHandler.ApplyPreset)

	// SSL 证书管理路由
	protected.Get("/certificates", certificateHandler.ShowCertificates)
	protected.Get("/api/certificates/edge/:id/details", certificateHandler.GetEdgeCertificateDetails)
	protected.Get("/api/certificates/origin/:id/download", certificateHandler.DownloadOriginCertificate)
	protected.Post("/api/certificates/origin/create", certificateHandler.CreateOriginCertificate)
	protected.Post("/api/certificates/origin/:id/revoke", certificateHandler.RevokeOriginCertificate)

	// 统计分析路由 - 已移除（Cloudflare Analytics API 实现复杂，需要 GraphQL）
	// protected.Get("/analytics", analyticsHandler.ShowAnalytics)

	// 健康检查
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"time":   time.Now().Unix(),
		})
	})
}
