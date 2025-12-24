# Cloudflare DNS Manager

> 更快、更易用的 Cloudflare 管理面板

专为解决国内访问慢、操作繁琐、CNAME 接入配置难等问题而生的轻量级 Cloudflare 管理面板。

## 核心特性

- 🚀 **告别访问慢**：国内服务器部署，秒开管理面板，无需翻墙
- 🎯 **支持 CNAME Setup**：官方Dashboard不支持的合作伙伴模式配置
- ⚡ **操作更高效**：配置模板、一键清除缓存、实时搜索过滤
- 🔒 **功能更完整**：SSL 证书管理、DNSSEC、性能优化、安全设置
- 💎 **100% 开源**：MIT 许可证，自由部署，完全免费，无使用限制
- 📦 **单文件部署**：编译成单个可执行文件，无需依赖，开箱即用

## 功能列表

### DNS 记录管理
- ✅ 完整的 CRUD 操作（增删改查）
- ✅ 支持所有记录类型：A/AAAA/CNAME/MX/TXT/NS/SRV/CAA
- ✅ 一键开启/关闭 CDN 代理（CloudFlare Proxy）
- ✅ 实时搜索和过滤（按类型、代理状态）
- ✅ DNS 记录统计面板

### SSL 证书管理
- ✅ 查看边缘证书详情（有效期、状态）
- ✅ 创建免费 15 年回源证书（Origin CA Certificate）
- ✅ 一键下载 PEM 格式证书
- ✅ 证书撤销和续期提醒
- ✅ 自定义上传证书查看

### Zone 设置管理
- ✅ 开发模式一键切换（临时绕过缓存）
- ✅ SSL/TLS 加密模式选择（Off/Flexible/Full/Strict）
- ✅ 性能优化开关：Auto Minify、Brotli、HTTP/2、HTTP/3、Rocket Loader
- ✅ 安全级别调整（CAPTCHA 阈值）
- ✅ 浏览器缓存 TTL 设置
- ✅ Always Online 模式
- ✅ TLS 最低版本设置

### 缓存管理
- ✅ 清除所有缓存
- ✅ 按 URL 清除（支持批量）
- ✅ 按主机名清除
- ✅ 按前缀清除（企业版）
- ✅ 按标签清除（企业版）

### 配置预设模板
一键应用最佳实践配置：
- 🎨 WordPress 优化
- 📄 静态网站优化
- 🔌 API 服务优化
- 🛒 电商网站优化
- 🔧 开发环境

### 安全功能
- ✅ DNSSEC 管理
- ✅ SSL 验证信息查看
- ✅ 安全级别动态调整
- ✅ 删除域名（带严格意图确认）

### 现代化 UI
- ✅ 响应式设计，支持移动端
- ✅ HTMX 无刷新交互
- ✅ 实时消息提醒
- ✅ Bootstrap 5 美观界面

## 快速开始

### 环境要求

- Go 1.21+ （编译时）
- Cloudflare 账号和 Global API Key

### 编译

\`\`\`bash
# 克隆项目
git clone https://github.com/yourusername/Cloudflare-DNS-Manager.git
cd Cloudflare-DNS-Manager

# 编译
CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/cf-dns-manager
\`\`\`

### 运行

\`\`\`bash
# 直接运行
./bin/cf-dns-manager

# 后台运行
nohup ./bin/cf-dns-manager > app.log 2>&1 &

# 访问 http://localhost:8080
\`\`\`

## 安全声明

⚠️ **重要提示**：
- API Key 仅保存在服务器内存中（不写入磁盘）
- 浏览器关闭后会话自动清除
- 不会永久保存到数据库或日志文件
- **强烈建议自行部署**，不要使用他人提供的公共服务

## 技术栈

- **后端**：Go + Fiber v2
- **前端**：Bootstrap 5 + HTMX
- **API**：cloudflare-go

## 许可证

MIT License

---

⭐ 如果这个项目对您有帮助，请给个 Star！
