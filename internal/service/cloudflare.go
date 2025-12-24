package service

import (
	"context"
	"fmt"

	"github.com/cloudflare/cloudflare-go"
)

type CloudflareService struct {
	API   *cloudflare.API
	Email string
}

func NewCloudflareService(email, apiKey string) (*CloudflareService, error) {
	api, err := cloudflare.New(apiKey, email)
	if err != nil {
		return nil, fmt.Errorf("failed to create Cloudflare API client: %w", err)
	}

	return &CloudflareService{
		API:   api,
		Email: email,
	}, nil
}

// VerifyCredentials 验证凭证是否有效
func (s *CloudflareService) VerifyCredentials(ctx context.Context) error {
	_, err := s.API.UserDetails(ctx)
	return err
}

// ListZones 获取域名列表
func (s *CloudflareService) ListZones(ctx context.Context, page int) ([]cloudflare.Zone, *cloudflare.ResultInfo, error) {
	zones, err := s.API.ListZones(ctx)
	if err != nil {
		return nil, nil, err
	}

	// 简单分页处理
	perPage := 20
	start := (page - 1) * perPage
	end := start + perPage

	resultInfo := &cloudflare.ResultInfo{
		Page:       page,
		PerPage:    perPage,
		TotalPages: (len(zones) + perPage - 1) / perPage,
		Count:      len(zones),
		Total:      len(zones),
	}

	if start >= len(zones) {
		return []cloudflare.Zone{}, resultInfo, nil
	}
	if end > len(zones) {
		end = len(zones)
	}

	return zones[start:end], resultInfo, nil
}

// GetZone 获取单个域名信息
func (s *CloudflareService) GetZone(ctx context.Context, zoneID string) (cloudflare.Zone, error) {
	zone, err := s.API.ZoneDetails(ctx, zoneID)
	return zone, err
}

// CreateZone 添加域名
func (s *CloudflareService) CreateZone(ctx context.Context, name string) (cloudflare.Zone, error) {
	zone, err := s.API.CreateZone(ctx, name, false, cloudflare.Account{}, "full")
	return zone, err
}

// DeleteZone 删除域名
func (s *CloudflareService) DeleteZone(ctx context.Context, zoneID string) error {
	_, err := s.API.DeleteZone(ctx, zoneID)
	return err
}

// ListDNSRecords 获取 DNS 记录列表
func (s *CloudflareService) ListDNSRecords(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListDNSRecordsParams) ([]cloudflare.DNSRecord, *cloudflare.ResultInfo, error) {
	records, result, err := s.API.ListDNSRecords(ctx, rc, params)
	return records, result, err
}

// GetDNSRecord 获取单条 DNS 记录
func (s *CloudflareService) GetDNSRecord(ctx context.Context, rc *cloudflare.ResourceContainer, recordID string) (cloudflare.DNSRecord, error) {
	record, err := s.API.GetDNSRecord(ctx, rc, recordID)
	return record, err
}

// CreateDNSRecord 创建 DNS 记录
func (s *CloudflareService) CreateDNSRecord(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateDNSRecordParams) (cloudflare.DNSRecord, error) {
	record, err := s.API.CreateDNSRecord(ctx, rc, params)
	return record, err
}

// UpdateDNSRecord 更新 DNS 记录
func (s *CloudflareService) UpdateDNSRecord(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateDNSRecordParams) (cloudflare.DNSRecord, error) {
	record, err := s.API.UpdateDNSRecord(ctx, rc, params)
	return record, err
}

// DeleteDNSRecord 删除 DNS 记录
func (s *CloudflareService) DeleteDNSRecord(ctx context.Context, rc *cloudflare.ResourceContainer, recordID string) error {
	return s.API.DeleteDNSRecord(ctx, rc, recordID)
}

// GetSSLVerification 获取 SSL 验证信息
func (s *CloudflareService) GetSSLVerification(ctx context.Context, zoneID string) ([]cloudflare.SSLValidationRecord, error) {
	// 使用原始 API 调用
	// 注意：cloudflare-go SDK 可能没有直接的 SSL 验证方法，需要使用原始请求
	return nil, fmt.Errorf("not implemented yet")
}

// AnalyticsData 分析数据结构
type AnalyticsData struct {
	Requests   []DataPoint `json:"requests"`
	Bandwidth  []DataPoint `json:"bandwidth"`
	PageViews  []DataPoint `json:"pageviews"`
	Uniques    []DataPoint `json:"uniques"`
	Threats    []DataPoint `json:"threats"`
}

// DataPoint 数据点
type DataPoint struct {
	Timestamp string  `json:"timestamp"`
	Value     float64 `json:"value"`
}

// GetAnalytics 获取分析数据
func (s *CloudflareService) GetAnalytics(ctx context.Context, zoneID string, since, until int) (*AnalyticsData, error) {
	// TODO: 使用 Cloudflare Analytics API 获取实际数据
	// 目前返回模拟数据用于演示
	analytics := &AnalyticsData{
		Requests: []DataPoint{
			{Timestamp: "00:00", Value: 1200},
			{Timestamp: "04:00", Value: 800},
			{Timestamp: "08:00", Value: 2400},
			{Timestamp: "12:00", Value: 3600},
			{Timestamp: "16:00", Value: 2800},
			{Timestamp: "20:00", Value: 1600},
		},
		Bandwidth: []DataPoint{
			{Timestamp: "00:00", Value: 52.4},
			{Timestamp: "04:00", Value: 35.2},
			{Timestamp: "08:00", Value: 98.6},
			{Timestamp: "12:00", Value: 142.8},
			{Timestamp: "16:00", Value: 112.3},
			{Timestamp: "20:00", Value: 68.5},
		},
		PageViews: []DataPoint{
			{Timestamp: "00:00", Value: 450},
			{Timestamp: "04:00", Value: 320},
			{Timestamp: "08:00", Value: 890},
			{Timestamp: "12:00", Value: 1250},
			{Timestamp: "16:00", Value: 960},
			{Timestamp: "20:00", Value: 580},
		},
		Uniques: []DataPoint{
			{Timestamp: "00:00", Value: 320},
			{Timestamp: "04:00", Value: 240},
			{Timestamp: "08:00", Value: 650},
			{Timestamp: "12:00", Value: 890},
			{Timestamp: "16:00", Value: 720},
			{Timestamp: "20:00", Value: 420},
		},
		Threats: []DataPoint{
			{Timestamp: "00:00", Value: 12},
			{Timestamp: "04:00", Value: 8},
			{Timestamp: "08:00", Value: 24},
			{Timestamp: "12:00", Value: 35},
			{Timestamp: "16:00", Value: 28},
			{Timestamp: "20:00", Value: 16},
		},
	}

	return analytics, nil
}

// DNSSECDetails 结构体
type DNSSECDetails struct {
	Status       string `json:"status"`
	DNSSECPresent bool  `json:"dnssec_present"`
	DS           string `json:"ds"`
	DNSKey       string `json:"dnskey"`
}

// GetDNSSEC 获取 DNSSEC 状态
func (s *CloudflareService) GetDNSSEC(ctx context.Context, zoneID string) (*DNSSECDetails, error) {
	// TODO: 使用 Cloudflare API 获取 DNSSEC 详情
	// 目前返回占位符数据
	return &DNSSECDetails{
		Status:       "disabled",
		DNSSECPresent: false,
	}, nil
}

// UpdateDNSSEC 更新 DNSSEC 状态
func (s *CloudflareService) UpdateDNSSEC(ctx context.Context, zoneID string, status string) (*DNSSECDetails, error) {
	// TODO: 使用 Cloudflare API 更新 DNSSEC 状态
	// 目前返回占位符数据
	return &DNSSECDetails{
		Status:       status,
		DNSSECPresent: status == "active",
	}, nil
}

// GetZoneSettings 获取所有 Zone 设置
func (s *CloudflareService) GetZoneSettings(ctx context.Context, zoneID string) ([]cloudflare.ZoneSetting, error) {
	settings, err := s.API.ZoneSettings(ctx, zoneID)
	if err != nil {
		return nil, err
	}
	return settings.Result, nil
}

// UpdateZoneSetting 更新单个设置
func (s *CloudflareService) UpdateZoneSetting(ctx context.Context, zoneID, settingID string, value interface{}) error {
	_, err := s.API.UpdateZoneSettings(ctx, zoneID, []cloudflare.ZoneSetting{
		{ID: settingID, Value: value},
	})
	return err
}

// PurgeAllCache 清除所有缓存
func (s *CloudflareService) PurgeAllCache(ctx context.Context, zoneID string) error {
	_, err := s.API.PurgeEverything(ctx, zoneID)
	return err
}

// PurgeCacheByURLs 按 URL 清除缓存
func (s *CloudflareService) PurgeCacheByURLs(ctx context.Context, zoneID string, urls []string) error {
	_, err := s.API.PurgeCache(ctx, zoneID, cloudflare.PurgeCacheRequest{
		Files: urls,
	})
	return err
}

// PurgeCacheByHosts 按主机名清除缓存
func (s *CloudflareService) PurgeCacheByHosts(ctx context.Context, zoneID string, hosts []string) error {
	_, err := s.API.PurgeCache(ctx, zoneID, cloudflare.PurgeCacheRequest{
		Hosts: hosts,
	})
	return err
}

// PurgeCacheByPrefixes 按前缀清除缓存
func (s *CloudflareService) PurgeCacheByPrefixes(ctx context.Context, zoneID string, prefixes []string) error {
	_, err := s.API.PurgeCache(ctx, zoneID, cloudflare.PurgeCacheRequest{
		Prefixes: prefixes,
	})
	return err
}

// PurgeCacheByTags 按 Cache-Tag 清除缓存（需要企业账户）
func (s *CloudflareService) PurgeCacheByTags(ctx context.Context, zoneID string, tags []string) error {
	_, err := s.API.PurgeCache(ctx, zoneID, cloudflare.PurgeCacheRequest{
		Tags: tags,
	})
	return err
}

// ============ 证书管理方法 ============

// ListEdgeCertificates 列出边缘证书（Certificate Packs）
func (s *CloudflareService) ListEdgeCertificates(ctx context.Context, zoneID string) ([]cloudflare.CertificatePack, error) {
	return s.API.ListCertificatePacks(ctx, zoneID)
}

// GetEdgeCertificate 获取单个边缘证书详情
func (s *CloudflareService) GetEdgeCertificate(ctx context.Context, zoneID, certID string) (cloudflare.CertificatePack, error) {
	return s.API.CertificatePack(ctx, zoneID, certID)
}

// ListOriginCertificates 列出回源证书
func (s *CloudflareService) ListOriginCertificates(ctx context.Context, zoneID string) ([]cloudflare.OriginCACertificate, error) {
	params := cloudflare.ListOriginCertificatesParams{
		ZoneID: zoneID,
	}
	return s.API.ListOriginCACertificates(ctx, params)
}

// GetOriginCertificate 获取单个回源证书
func (s *CloudflareService) GetOriginCertificate(ctx context.Context, certID string) (*cloudflare.OriginCACertificate, error) {
	return s.API.GetOriginCACertificate(ctx, certID)
}

// CreateOriginCertificate 创建回源证书
func (s *CloudflareService) CreateOriginCertificate(ctx context.Context, hostnames []string, requestType string, validityDays int) (*cloudflare.OriginCACertificate, error) {
	params := cloudflare.CreateOriginCertificateParams{
		Hostnames:       hostnames,
		RequestType:     requestType,
		RequestValidity: validityDays,
	}
	return s.API.CreateOriginCACertificate(ctx, params)
}

// RevokeOriginCertificate 撤销回源证书
func (s *CloudflareService) RevokeOriginCertificate(ctx context.Context, certID string) error {
	_, err := s.API.RevokeOriginCACertificate(ctx, certID)
	return err
}

// ListCustomSSLCertificates 列出自定义 SSL 证书
func (s *CloudflareService) ListCustomSSLCertificates(ctx context.Context, zoneID string) ([]cloudflare.ZoneCustomSSL, error) {
	return s.API.ListSSL(ctx, zoneID)
}
