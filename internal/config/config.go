package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Host      string `yaml:"host"`
		Port      int    `yaml:"port"`
		Debug     bool   `yaml:"debug"`
		PageTitle string `yaml:"page_title"`
	} `yaml:"server"`

	Session struct {
		Expire         int `yaml:"expire"`
		RememberExpire int `yaml:"remember_expire"`
	} `yaml:"session"`

	RateLimit struct {
		MaxAttempts int `yaml:"max_attempts"`
		Window      int `yaml:"window"`
	} `yaml:"ratelimit"`

	Cache struct {
		DNSTTL int `yaml:"dns_ttl"`
	} `yaml:"cache"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// 设置默认值
	if cfg.Server.Host == "" {
		cfg.Server.Host = "0.0.0.0"
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Server.PageTitle == "" {
		cfg.Server.PageTitle = "Cloudflare DNS Manager"
	}
	if cfg.Session.Expire == 0 {
		cfg.Session.Expire = 3600
	}
	if cfg.Session.RememberExpire == 0 {
		cfg.Session.RememberExpire = 31536000
	}
	if cfg.RateLimit.MaxAttempts == 0 {
		cfg.RateLimit.MaxAttempts = 5
	}
	if cfg.RateLimit.Window == 0 {
		cfg.RateLimit.Window = 60
	}
	if cfg.Cache.DNSTTL == 0 {
		cfg.Cache.DNSTTL = 172800
	}

	return &cfg, nil
}
