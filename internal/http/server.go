package http

import (
	"context"
	"crypto/tls"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/uptrace/bun"
	"golang.org/x/time/rate"

	"github.com/fivemanage/lite/internal/http/internalapi"
	internalmiddleware "github.com/fivemanage/lite/internal/http/middleware"
	"github.com/fivemanage/lite/internal/http/publicapi"
	_validator "github.com/fivemanage/lite/internal/http/validator"
	"github.com/fivemanage/lite/internal/service/auth"
	"github.com/fivemanage/lite/internal/service/dataset"
	"github.com/fivemanage/lite/internal/service/file"
	"github.com/fivemanage/lite/internal/service/invite"
	"github.com/fivemanage/lite/internal/service/log"
	"github.com/fivemanage/lite/internal/service/member"
	"github.com/fivemanage/lite/internal/service/organization"
	"github.com/fivemanage/lite/internal/service/system"
	"github.com/fivemanage/lite/internal/service/token"
	"github.com/fivemanage/lite/pkg/otel"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4/middleware"
)

//go:embed dist
var webContent embed.FS

type Server struct {
	Engine *echo.Echo
}

// TODO: Add sentry for monitoring. There should be an opt-out option.
func NewServer(
	store *bun.DB,
	authService *auth.Service,
	tokenService *token.Service,
	fileService *file.Service,
	organizationService *organization.Service,
	memberService *member.Service,
	inviteService *invite.Service,
	logService *log.Service,
	datasetService *dataset.Service,
	systemService *system.Service,
) *echo.Echo {
	app := echo.New()
	app.Debug = os.Getenv("ENV") == "dev"

	// Behind Cloudflare + Traefik (Coolify) the public HTTPS connection terminates at
	// the proxy, so requests reach Echo as plain HTTP. Trust the forwarded headers so
	// the real client IP and public scheme are used. Only the configured proxy ranges
	// are trusted, so clients cannot spoof X-Forwarded-For to forge their IP.
	app.IPExtractor = echo.ExtractIPFromXFFHeader(trustedProxyOptions()...)
	app.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			if req.Header.Get(echo.HeaderXForwardedProto) == "https" {
				req.URL.Scheme = "https"
				if req.TLS == nil {
					// Mark the request as TLS so c.Scheme()/c.IsTLS() report "https",
					// which fixes Secure cookies and generated absolute URLs.
					req.TLS = &tls.ConnectionState{}
				}
			}
			return next(c)
		}
	})

	app.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOriginFunc: allowedOrigin,
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
			"X-CSRF-Token",
			"X-Fivemanage-Dataset",
		},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowCredentials: true,
	}))

	if otel.Enabled() {
		app.Use(otelecho.Middleware("lite-api"))
	}
	app.Use(middleware.Recover())
	app.Use(internalmiddleware.AppContext)
	// Rate limit API traffic only. The static SPA pulls many assets per page load,
	// which would otherwise trip a shared per-IP limit for legitimate users.
	app.Use(middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Store: middleware.NewRateLimiterMemoryStore(rateLimit()),
		Skipper: func(c echo.Context) bool {
			return !strings.HasPrefix(c.Request().URL.Path, "/api")
		},
	}))

	app.Validator = &_validator.CustomValidator{Validator: validator.New()}

	app.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Filesystem: getFileSystem("dist"),
		HTML5:      true,
		Skipper: func(c echo.Context) bool {
			return strings.HasPrefix(c.Request().URL.Path, "/api")
		},
	}))

	apiGroup := app.Group("/api")
	registerHealthRoutes(apiGroup, store)
	dashApi := apiGroup.Group("/dash")

	internalapi.Add(
		dashApi,
		authService,
		tokenService,
		organizationService,
		memberService,
		inviteService,
		fileService,
		datasetService,
		systemService,
	)
	publicapi.Add(apiGroup, fileService, tokenService, logService)

	return app
}

func getFileSystem(path string) http.FileSystem {
	fs, err := fs.Sub(webContent, path)
	if err != nil {
		panic(err)
	}

	slog.Info(fmt.Sprintf("serving static files from %s", path))

	return http.FS(fs)
}

// registerHealthRoutes wires liveness and readiness probes. /api/health reports
// process liveness; /api/ready additionally verifies the database is reachable.
func registerHealthRoutes(group *echo.Group, store *bun.DB) {
	group.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, echo.Map{"status": "ok"})
	})
	group.GET("/ready", func(c echo.Context) error {
		ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
		defer cancel()
		if store == nil {
			return c.JSON(http.StatusServiceUnavailable, echo.Map{"status": "unavailable", "check": "database"})
		}
		if err := store.PingContext(ctx); err != nil {
			slog.Warn("readiness check failed", slog.Any("error", err))
			return c.JSON(http.StatusServiceUnavailable, echo.Map{"status": "unavailable", "check": "database"})
		}
		return c.JSON(http.StatusOK, echo.Map{"status": "ready"})
	})
}

// rateLimit returns the per-IP request/second limit, configurable via
// API_RATE_LIMIT (default 20).
func rateLimit() rate.Limit {
	if v := os.Getenv("API_RATE_LIMIT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return rate.Limit(n)
		}
	}
	return rate.Limit(20)
}

// trustedProxyOptions builds the set of trusted-proxy ranges used when parsing
// X-Forwarded-For. Loopback, link-local and private ranges are trusted by
// default (the app runs behind an in-cluster proxy); additional public proxy
// CIDRs (e.g. Cloudflare) can be supplied via TRUSTED_PROXIES (comma-separated).
func trustedProxyOptions() []echo.TrustOption {
	opts := []echo.TrustOption{
		echo.TrustLoopback(true),
		echo.TrustLinkLocal(true),
		echo.TrustPrivateNet(true),
	}
	for _, cidr := range strings.Split(os.Getenv("TRUSTED_PROXIES"), ",") {
		cidr = strings.TrimSpace(cidr)
		if cidr == "" {
			continue
		}
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			slog.Warn("ignoring invalid TRUSTED_PROXIES entry", slog.String("cidr", cidr), slog.Any("error", err))
			continue
		}
		opts = append(opts, echo.TrustIPRange(ipNet))
	}
	return opts
}

func allowedOrigin(origin string) (bool, error) {
	if origin == "" {
		return true, nil
	}

	allowedOrigins := strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",")
	for _, allowed := range allowedOrigins {
		if strings.TrimSpace(allowed) == origin {
			return true, nil
		}
	}

	return false, nil
}
