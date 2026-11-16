package main

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/DarkoKlisuric/godebugbar"
	ginadapter "github.com/DarkoKlisuric/godebugbar/adapter/gin"
	gormadapter "github.com/DarkoKlisuric/godebugbar/adapter/gorm"
)

var page = template.Must(template.New("page").Parse(`<!doctype html>
<html>
<head><title>{{.Title}}</title><style>body{font-family:system-ui,sans-serif;margin:40px;color:#222}a{margin-right:16px}</style></head>
<body>
<nav><a href="/">N+1 page</a><a href="/preload">Preload page</a><a href="/_debugbar">Profiler</a></nav>
<h1>{{.Title}}</h1>
<ul>{{range .Groups}}<li>{{.Name}}: {{len .Members}} members</li>{{end}}</ul>
</body>
</html>`))

func main() {
	if err := run(); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	dsn := envOr("DATABASE_URL", "postgres://godebugbar:godebugbar@localhost:5433/godebugbar?sslmode=disable")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("connect to postgres (run docker compose up -d): %w", err)
	}
	if err := seed(db); err != nil {
		return err
	}

	router := gin.New()
	router.Use(gin.Recovery())
	godebugbar.New(
		godebugbar.WithEnabled(true),
		godebugbar.WithHTMLInjection(),
		godebugbar.WithExcludedPaths("/favicon.ico"),
		ginadapter.Attach(router),
		gormadapter.Attach(db),
	)

	router.GET("/", func(c *gin.Context) {
		var groups []group
		if err := db.WithContext(c).Order("name").Find(&groups).Error; err != nil {
			_ = c.Error(err)
		}
		for i := range groups {
			if err := db.WithContext(c).Where("group_id = ?", groups[i].ID).Find(&groups[i].Members).Error; err != nil {
				_ = c.Error(err)
			}
		}
		render(c, "Groups (N+1)", groups)
	})
	router.GET("/preload", func(c *gin.Context) {
		var groups []group
		if err := db.WithContext(c).Preload("Members").Order("name").Find(&groups).Error; err != nil {
			_ = c.Error(err)
		}
		render(c, "Groups (Preload)", groups)
	})

	addr := envOr("ADDR", ":8080")
	server := &http.Server{Addr: addr, Handler: router, ReadHeaderTimeout: 5 * time.Second}
	slog.Info("open http://localhost" + addr + " and http://localhost" + addr + "/_debugbar")

	return server.ListenAndServe()
}

func seed(db *gorm.DB) error {
	if err := db.AutoMigrate(&group{}, &member{}); err != nil {
		return err
	}
	var count int64
	if err := db.Model(&group{}).Count(&count).Error; err != nil || count > 0 {
		return err
	}
	for _, name := range []string{"Alpha", "Bravo", "Charlie", "Delta", "Echo", "Foxtrot"} {
		members := make([]member, 5)
		for i := range members {
			members[i].Name = fmt.Sprintf("%s %d", name, i+1)
		}
		if err := db.Create(&group{Name: name, Members: members}).Error; err != nil {
			return err
		}
	}
	return nil
}

func render(c *gin.Context, title string, groups []group) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := page.Execute(c.Writer, map[string]any{"Title": title, "Groups": groups}); err != nil {
		_ = c.Error(err)
	}
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
