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

### 获取 Cloudflare API Key

1. 登录 [Cloudflare Dashboard](https://dash.cloudflare.com/)
2. 点击右上角头像 → My Profile
3. 选择 API Tokens
4. 在 API Keys 部分找到 Global API Key
5. 点击 View 查看您的 API Key

### 编译

```bash
# 克隆项目
git clone https://github.com/zhufengme/Cloudflare-CNAME-Setup.git
cd Cloudflare-CNAME-Setup

# 编译（Linux/macOS）
CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/cf-dns-manager

# 编译（Windows）
CGO_ENABLED=0 GOOS=windows go build -ldflags="-s -w" -o bin/cf-dns-manager.exe
```

### 运行

```bash
# 使用默认配置运行（会在当前目录查找 config.yaml）
./bin/cf-dns-manager

# 指定配置文件路径
./bin/cf-dns-manager -config /path/to/your/config.yaml

# 使用不同配置文件名
./bin/cf-dns-manager -config production.yaml

# 后台运行
nohup ./bin/cf-dns-manager -config config.yaml > app.log 2>&1 &

# 查看帮助
./bin/cf-dns-manager -h

# 默认访问地址: http://localhost:8080
```

**注意**：
- 如果不提供配置文件或配置文件不存在，程序将使用默认配置运行
- 默认监听地址：`0.0.0.0:8080`
- 可以不创建配置文件，直接使用默认值

## 配置说明

### 配置文件格式

创建 `config.yaml`（可选）：

```yaml
server:
  host: 0.0.0.0              # 监听地址（0.0.0.0 表示监听所有网卡）
  port: 8080                 # 监听端口
  page_title: "Cloudflare DNS Manager"  # 页面标题
  debug: false               # 调试模式（true 时显示详细日志）

session:
  expire: 3600               # 普通会话过期时间（秒）
                             # 默认 3600 秒 = 1 小时
  remember_expire: 31536000  # "记住我" 会话过期时间（秒）
                             # 默认 31536000 秒 = 365 天

rate_limit:
  max_attempts: 5            # 登录失败最大尝试次数
  window: 60                 # 限流时间窗口（分钟）
                             # 超过 max_attempts 次失败后，需等待 window 分钟

cache:
  dns_ttl: 172800            # DNS 记录缓存时间（秒）
                             # 默认 172800 秒 = 48 小时
```

### 配置参数详解

#### 服务器配置 (server)

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `host` | string | `0.0.0.0` | 监听地址。`0.0.0.0` 表示监听所有网卡，可以通过任意 IP 访问 |
| `port` | int | `8080` | 监听端口。确保端口未被占用 |
| `page_title` | string | `Cloudflare DNS Manager` | 页面标题，显示在浏览器标签页 |
| `debug` | bool | `false` | 是否开启调试模式。开启后会输出详细的 HTTP 请求日志 |

**监听地址示例**：
- `0.0.0.0` - 监听所有网卡，可通过任意 IP 访问（推荐用于服务器）
- `127.0.0.1` - 仅本机访问，其他机器无法连接（推荐用于本地开发）
- `192.168.1.100` - 仅通过指定 IP 访问

#### 会话配置 (session)

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `expire` | int | `3600` | 普通会话过期时间（秒）。用户未勾选"记住我"时使用 |
| `remember_expire` | int | `31536000` | "记住我" 会话过期时间（秒）。用户勾选"记住我"时使用 |

**会话时间换算**：
- 1 小时 = `3600`
- 1 天 = `86400`
- 7 天 = `604800`
- 30 天 = `2592000`
- 365 天 = `31536000`

#### 限流配置 (rate_limit)

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `max_attempts` | int | `5` | 登录失败最大尝试次数 |
| `window` | int | `60` | 限流时间窗口（分钟） |

**限流机制**：
- 用户在 `window` 分钟内登录失败超过 `max_attempts` 次
- 该邮箱将被锁定 `window` 分钟
- 防止暴力破解攻击

#### 缓存配置 (cache)

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `dns_ttl` | int | `172800` | DNS 记录缓存时间（秒）。用于减少 API 调用次数 |

### 命令行参数

```bash
./bin/cf-dns-manager -h
```

**可用参数**：

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `-config` | string | `config.yaml` | 配置文件路径 |

**使用示例**：

```bash
# 使用默认配置文件（当前目录的 config.yaml）
./bin/cf-dns-manager

# 指定配置文件路径
./bin/cf-dns-manager -config /etc/cf-dns-manager/config.yaml

# 使用当前目录的其他配置文件
./bin/cf-dns-manager -config production.yaml

# 使用绝对路径
./bin/cf-dns-manager -config /home/user/configs/cf.yaml

# 使用相对路径
./bin/cf-dns-manager -config ../configs/config.yaml
```

### 默认值说明

如果不提供配置文件或配置文件加载失败，程序将使用以下默认值：

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  page_title: "Cloudflare DNS Manager"
  debug: false

session:
  expire: 3600           # 1 小时
  remember_expire: 31536000  # 365 天

rate_limit:
  max_attempts: 5
  window: 60             # 60 分钟

cache:
  dns_ttl: 172800        # 48 小时
```

这意味着您可以**直接运行程序而不创建配置文件**，程序会使用默认配置。

### 配置文件示例

项目提供了 `config.yaml.example` 示例文件：

```bash
# 复制示例文件
cp config.yaml.example config.yaml

# 编辑配置
vi config.yaml

# 使用自定义配置运行
./bin/cf-dns-manager -config config.yaml
```

## 部署建议

### Systemd 服务（推荐）

创建 `/etc/systemd/system/cf-dns-manager.service`：

```ini
[Unit]
Description=Cloudflare DNS Manager
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/Cloudflare-DNS-Manager
ExecStart=/opt/Cloudflare-DNS-Manager/bin/cf-dns-manager -config /opt/Cloudflare-DNS-Manager/config.yaml
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

管理服务：

```bash
# 重新加载 systemd
sudo systemctl daemon-reload

# 启动服务
sudo systemctl start cf-dns-manager

# 设置开机自启
sudo systemctl enable cf-dns-manager

# 查看状态
sudo systemctl status cf-dns-manager

# 查看日志
sudo journalctl -u cf-dns-manager -f

# 停止服务
sudo systemctl stop cf-dns-manager

# 重启服务
sudo systemctl restart cf-dns-manager
```

### Nginx 反向代理

```nginx
server {
    listen 80;
    server_name dns.example.com;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### Docker 部署（可选）

虽然本项目已优化为单文件部署，但如果您仍需要 Docker：

```dockerfile
FROM alpine:latest
WORKDIR /app
COPY bin/cf-dns-manager /app/
COPY config.yaml /app/
EXPOSE 8080
CMD ["/app/cf-dns-manager", "-config", "/app/config.yaml"]
```

```bash
# 构建镜像
docker build -t cf-dns-manager .

# 运行容器
docker run -d -p 8080:8080 \
  -v /path/to/config.yaml:/app/config.yaml \
  --name cf-dns-manager \
  cf-dns-manager
```

## 安全声明

⚠️ **重要提示**：

- **API Key 存储**：仅保存在服务器内存中（不写入磁盘）
- **会话管理**：浏览器关闭后会话自动清除（未勾选"记住我"）
- **数据安全**：不会永久保存到数据库或日志文件
- **风险提示**：提供 Global API Key 意味着授予完整的账户操作权限
- **部署建议**：
  - ✅ **强烈建议自行部署**，不要使用他人提供的公共服务
  - ✅ 使用 HTTPS（配合 Nginx + Let's Encrypt）
  - ✅ 配置防火墙规则，仅允许可信 IP 访问
  - ✅ 定期更新程序和依赖
  - ⚠️ 如需生产使用，建议使用 Cloudflare API Token（权限更细粒度，需修改代码）

## 常见问题

### 1. 如何修改监听端口？

创建 `config.yaml`：

```yaml
server:
  port: 3000  # 修改为您想要的端口
```

或使用环境变量（需代码支持）。

### 2. 配置文件在哪里？

默认情况下，程序会在**当前工作目录**查找 `config.yaml`。

您可以通过 `-config` 参数指定任意位置：

```bash
./bin/cf-dns-manager -config /etc/cf-dns-manager/config.yaml
```

### 3. 配置文件是必需的吗？

**不是必需的**。如果不提供配置文件，程序使用默认值运行。

### 4. 如何启用调试日志？

在 `config.yaml` 中设置：

```yaml
server:
  debug: true
```

重启程序后，会在控制台输出详细的 HTTP 请求日志。

### 5. 会话为什么总是过期？

检查 `config.yaml` 中的 `session.expire` 设置：

```yaml
session:
  expire: 86400  # 设置为 24 小时（86400 秒）
```

如果需要长期保持登录，勾选登录页面的"记住我"选项。

### 6. 如何限制访问 IP？

使用防火墙或 Nginx：

```nginx
# 仅允许特定 IP 访问
location / {
    allow 192.168.1.0/24;
    deny all;
    proxy_pass http://127.0.0.1:8080;
}
```

### 7. 支持 HTTPS 吗？

程序本身不直接支持 HTTPS，建议通过 Nginx 反向代理实现：

```bash
# 使用 Certbot 获取免费 SSL 证书
sudo certbot --nginx -d dns.example.com
```

## 技术栈

- **后端**：Go 1.21+ + Fiber v2.52.10
- **前端**：Bootstrap 5 + HTMX + Alpine.js
- **API 客户端**：cloudflare-go v0.116.0
- **模板引擎**：Go html/template（嵌入式）
- **会话管理**：fiber/storage/memory

## 项目结构

```
Cloudflare-DNS-Manager/
├── bin/                    # 编译输出目录
│   └── cf-dns-manager      # 可执行文件
├── internal/               # 内部包
│   ├── config/             # 配置加载
│   ├── handler/            # HTTP 处理器
│   ├── middleware/         # 中间件
│   ├── service/            # 业务逻辑
│   └── i18n/               # 国际化
├── web/                    # 前端资源（嵌入式）
│   ├── static/             # CSS/JS/图片
│   ├── templates/          # HTML 模板
│   └── locales/            # 语言文件
├── main.go                 # 入口文件
├── config.yaml.example     # 配置示例
└── README.md               # 项目文档
```

## 贡献指南

欢迎提交 Issue 和 Pull Request！

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 许可证

本项目采用 [MIT 许可证](LICENSE)。

## 鸣谢

- [Cloudflare](https://www.cloudflare.com/) - CDN 和安全服务提供商
- [cloudflare-go](https://github.com/cloudflare/cloudflare-go) - Cloudflare API Go 客户端
- [Fiber](https://gofiber.io/) - 高性能 Go Web 框架
- [HTMX](https://htmx.org/) - 现代化无刷新交互库
- [Bootstrap](https://getbootstrap.com/) - 响应式 UI 框架

## 免责声明

本项目与 Cloudflare, Inc. 无关联。Cloudflare 是 Cloudflare, Inc. 的注册商标。

---

⭐ 如果这个项目对您有帮助，请给个 Star！
