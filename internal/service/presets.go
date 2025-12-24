package service

import (
	"context"

	"github.com/cloudflare/cloudflare-go"
)

// ConfigPreset 定义配置预设
type ConfigPreset struct {
	Name        string
	Description string
	Settings    []cloudflare.ZoneSetting
}

// Presets 预定义的配置模板
var Presets = map[string]ConfigPreset{
	"wordpress": {
		Name:        "WordPress 优化",
		Description: "适合 WordPress 网站的推荐配置",
		Settings: []cloudflare.ZoneSetting{
			{ID: "ssl", Value: "full"},
			{ID: "security_level", Value: "medium"},
			{ID: "browser_cache_ttl", Value: 14400}, // 4 hours
			{ID: "minify", Value: map[string]string{"css": "on", "html": "on", "js": "on"}},
			{ID: "brotli", Value: "on"},
			{ID: "http2", Value: "on"},
			{ID: "http3", Value: "on"},
			{ID: "rocket_loader", Value: "on"},
			{ID: "always_online", Value: "on"},
		},
	},
	"static": {
		Name:        "静态网站优化",
		Description: "适合静态网站（HTML/CSS/JS），最大化缓存",
		Settings: []cloudflare.ZoneSetting{
			{ID: "ssl", Value: "full"},
			{ID: "security_level", Value: "low"},
			{ID: "browser_cache_ttl", Value: 31536000}, // 1 year
			{ID: "minify", Value: map[string]string{"css": "on", "html": "on", "js": "on"}},
			{ID: "brotli", Value: "on"},
			{ID: "http2", Value: "on"},
			{ID: "http3", Value: "on"},
			{ID: "always_online", Value: "on"},
			{ID: "rocket_loader", Value: "off"}, // 静态站点不需要
		},
	},
	"api": {
		Name:        "API 服务优化",
		Description: "适合 API 服务，禁用缓存，强化安全",
		Settings: []cloudflare.ZoneSetting{
			{ID: "ssl", Value: "full_strict"},
			{ID: "security_level", Value: "high"},
			{ID: "browser_cache_ttl", Value: 0}, // 不缓存
			{ID: "brotli", Value: "on"},
			{ID: "http2", Value: "on"},
			{ID: "http3", Value: "on"},
			{ID: "minify", Value: map[string]string{"css": "off", "html": "off", "js": "off"}},
			{ID: "rocket_loader", Value: "off"},
			{ID: "always_online", Value: "off"},
		},
	},
	"ecommerce": {
		Name:        "电商网站优化",
		Description: "适合电商网站，平衡性能和安全",
		Settings: []cloudflare.ZoneSetting{
			{ID: "ssl", Value: "full_strict"},
			{ID: "security_level", Value: "high"},
			{ID: "browser_cache_ttl", Value: 7200}, // 2 hours
			{ID: "minify", Value: map[string]string{"css": "on", "html": "on", "js": "on"}},
			{ID: "brotli", Value: "on"},
			{ID: "http2", Value: "on"},
			{ID: "http3", Value: "on"},
			{ID: "rocket_loader", Value: "off"}, // 电商站点可能有冲突
			{ID: "always_online", Value: "on"},
		},
	},
	"development": {
		Name:        "开发测试环境",
		Description: "适合开发环境，禁用缓存和优化",
		Settings: []cloudflare.ZoneSetting{
			{ID: "ssl", Value: "flexible"},
			{ID: "security_level", Value: "essentially_off"},
			{ID: "browser_cache_ttl", Value: 0},
			{ID: "minify", Value: map[string]string{"css": "off", "html": "off", "js": "off"}},
			{ID: "brotli", Value: "off"},
			{ID: "http2", Value: "on"},
			{ID: "http3", Value: "off"},
			{ID: "rocket_loader", Value: "off"},
			{ID: "always_online", Value: "off"},
			{ID: "development_mode", Value: "on"},
		},
	},
}

// ApplyPreset 应用预设配置
func (s *CloudflareService) ApplyPreset(ctx context.Context, zoneID, presetName string) error {
	preset, ok := Presets[presetName]
	if !ok {
		return cloudflare.ErrMissingZoneID // 返回一个合适的错误
	}

	// 批量更新设置
	_, err := s.API.UpdateZoneSettings(ctx, zoneID, preset.Settings)
	return err
}

// GetPresetInfo 获取预设信息（用于前端显示）
func GetPresetInfo(presetName string) (ConfigPreset, bool) {
	preset, ok := Presets[presetName]
	return preset, ok
}

// GetAllPresets 获取所有预设列表
func GetAllPresets() map[string]ConfigPreset {
	return Presets
}
