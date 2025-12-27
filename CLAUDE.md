# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Cloudflare DNS Manager is a lightweight web panel for managing Cloudflare DNS and zone settings, designed for faster access from China. It's built as a single-file Go binary with embedded web assets.

**Key Features:**
- DNS record management (all record types)
- SSL/TLS certificate management
- Zone settings with configuration presets (WordPress, static site, API, e-commerce, development)
- Cache management
- DNSSEC support
- Session-based authentication using Cloudflare Global API Key

## Build and Run Commands

### Building

```bash
# Build for Linux/macOS (creates static binary)
CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/cf-dns-manager

# Build for Windows
CGO_ENABLED=0 GOOS=windows go build -ldflags="-s -w" -o bin/cf-dns-manager.exe

# Build for other platforms
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o bin/cf-dns-manager-darwin-arm64
```

### Running

```bash
# Run with default config (looks for config.yaml in current directory)
./bin/cf-dns-manager

# Run with specific config file
./bin/cf-dns-manager -config /path/to/config.yaml

# Run in background
nohup ./bin/cf-dns-manager -config config.yaml > app.log 2>&1 &
```

**Default server:** `http://localhost:8080` (or configured port)

### Development

```bash
# Get dependencies
go mod download

# Run directly (for development)
go run main.go -config config.yaml

# Format code
go fmt ./...

# Check for issues
go vet ./...
```

## Architecture

### Request Flow

1. **Entry Point** (`main.go`):
   - Loads config from YAML or uses defaults
   - Initializes embedded file system for web assets (`//go:embed web`)
   - Sets up Fiber web framework with HTML template engine
   - Registers middleware (recovery, logger, i18n, session)
   - Defines routes and starts server

2. **Authentication Flow**:
   - User provides Cloudflare email + Global API Key via login form
   - `handler.AuthHandler` validates credentials using Cloudflare API
   - On success, credentials are stored in **memory-only session** (never disk)
   - `middleware.AuthRequired` injects credentials into request context for protected routes
   - Rate limiting prevents brute force (configurable max attempts/window)

3. **Session Management**:
   - Uses `gofiber/storage/memory` - sessions are **ephemeral** (lost on restart)
   - Credentials stored: `cloudflare_email` and `user_api_key`
   - Session expiry: configurable via `session.expire` and `session.remember_expire`
   - Protected routes extract credentials from context to create per-request `CloudflareService`

4. **Cloudflare API Interaction**:
   - Each request creates a new `service.CloudflareService` instance with user's API key
   - Wraps `cloudflare-go` SDK methods
   - Handles zone operations, DNS CRUD, settings, certificates, cache purging

5. **Frontend Architecture**:
   - **Templates**: Go html/template (in `web/templates/`)
   - **UI Framework**: Bootstrap 5
   - **Interactivity**: HTMX for dynamic content updates without full page reloads
   - **I18n**: Support for English and Chinese (`web/locales/`)

### Key Architectural Patterns

**Embedded Assets:**
- All web assets (templates, static files, locales) are embedded using `//go:embed web`
- Results in single binary with no external dependencies
- Templates: `web/templates/*.html`
- Static: `web/static/` (CSS, JS, images)
- Locales: `web/locales/{en,zh}.yaml`

**Handler Pattern:**
- Each major feature has a dedicated handler in `internal/handler/`:
  - `auth.go` - Login/logout
  - `zone.go` - Zone listing and management
  - `dns.go` - DNS record CRUD
  - `settings.go` - Zone settings, cache purging, presets
  - `certificate.go` - SSL certificate management
  - `security.go` - DNSSEC
  - `home.go` - Landing page

**Service Layer:**
- `internal/service/cloudflare.go` - Wraps cloudflare-go SDK, provides domain methods
- `internal/service/presets.go` - Configuration presets (WordPress, static, API, etc.)

**Middleware:**
- `middleware.AuthRequired` - Session validation and credential injection
- `middleware.RateLimiter` - Login attempt limiting
- `middleware.I18n` - Language detection and setup

**Configuration:**
- `internal/config/config.go` - YAML config loader with defaults
- Supports: server settings, session expiry, rate limiting, cache TTL
- Graceful degradation: works without config file using hardcoded defaults

### Important Constraints

**Security Model:**
- API keys are **never persisted to disk** - only in memory sessions
- Sessions are volatile (cleared on server restart)
- Uses Cloudflare Global API Key (full account access)
- No database - stateless except for in-memory sessions

**No Database:**
- All data fetched from Cloudflare API on demand
- DNS records cached in memory with configurable TTL (`cache.dns_ttl`)
- Session store is in-memory only

**Embedded Resources:**
- Cannot modify templates/static files at runtime
- Must rebuild binary to update web assets
- Templates use Go html/template syntax, not Jinja/etc.

### Configuration Presets

Five built-in zone configuration templates (`internal/service/presets.go`):
1. **wordpress** - Optimized for WordPress (minify, Brotli, Rocket Loader)
2. **static** - Max caching for static sites (1-year browser cache)
3. **api** - API services (no cache, strict SSL, high security)
4. **ecommerce** - E-commerce balance (2hr cache, strict SSL)
5. **development** - Dev mode (no cache, minimal security, flexible SSL)

Applied via `settingsHandler.ApplyPreset()` - batch updates multiple zone settings.

## Common Development Tasks

### Adding a New Route

1. Define route in `main.go` `setupRoutes()`:
   ```go
   protected.Get("/new-feature", newHandler.ShowFeature)
   protected.Post("/api/new-feature", newHandler.DoFeature)
   ```

2. Create handler in `internal/handler/`:
   ```go
   func (h *NewHandler) ShowFeature(c *fiber.Ctx) error {
       // Extract credentials from context
       email := c.Locals("cloudflare_email").(string)
       apiKey := c.Locals("user_api_key").(string)

       // Create Cloudflare service
       cfService, _ := service.NewCloudflareService(email, apiKey)

       // Render template
       return c.Render("new-feature", fiber.Map{...})
   }
   ```

3. Create template in `web/templates/`:
   ```html
   {{template "header" .}}
   <!-- Your content -->
   {{template "footer" .}}
   ```

### Working with Templates

- Templates use Go's `html/template` package
- Base templates: `header.html`, `footer.html` in `web/templates/`
- Custom template functions (defined in `main.go`):
  - `add`, `sub` - Math operations
  - `default` - Default value if nil
  - `dict` - Create map from key-value pairs
  - `daysUntil` - Days until a time.Time
- Pass data via `fiber.Map{}`
- HTMX attributes for dynamic updates (e.g., `hx-get`, `hx-post`, `hx-target`)

### Adding Cloudflare API Methods

1. Add method to `internal/service/cloudflare.go`:
   ```go
   func (s *CloudflareService) NewAPIMethod(ctx context.Context, zoneID string) error {
       // Use s.API to call cloudflare-go SDK
       return s.API.SomeMethod(ctx, zoneID, params)
   }
   ```

2. Use in handler:
   ```go
   err := cfService.NewAPIMethod(c.Context(), zoneID)
   ```

### Internationalization

- Language files: `web/locales/{en,zh}.yaml`
- Middleware sets `lang` based on `Accept-Language` header or `?lang=` param
- Access in templates via `{{.Lang.KeyName}}`
- Structure uses nested keys:
  ```yaml
  common:
    save: Save
    cancel: Cancel
  dns:
    add_record: Add DNS Record
  ```

### Modifying Configuration

Add new config fields:

1. Update `internal/config/config.go` struct:
   ```go
   type Config struct {
       NewSection struct {
           NewField string `yaml:"new_field"`
       } `yaml:"newsection"`
   }
   ```

2. Set default in `Load()` function
3. Use in `main.go`: `cfg.NewSection.NewField`

## Code Style Notes

- Use Go conventions (gofmt, camelCase)
- Handlers return `error` (Fiber standard)
- Context passed as `*fiber.Ctx`
- Cloudflare operations use `context.Context` (typically `c.Context()`)
- Session access: `middleware.Store.Get(c)`
- Error messages rendered via `error.html` template or HTMX partial responses
