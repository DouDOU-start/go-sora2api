package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed all:dist
var webDist embed.FS

// ServeWebUI 注册前端静态文件服务（SPA 支持）
func ServeWebUI(r *gin.Engine) {
	distFS, err := fs.Sub(webDist, "dist")
	if err != nil {
		log.Printf("[web] 未嵌入前端文件，跳过静态文件服务")
		return
	}

	// 检查是否有 index.html（判断前端是否已构建）
	if _, err := distFS.Open("index.html"); err != nil {
		log.Printf("[web] dist/index.html 不存在，跳过静态文件服务")
		return
	}

	fileServer := http.FileServer(http.FS(distFS))
	log.Printf("[web] 已启用前端静态文件服务")

	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// API 路径不做 fallback
		if strings.HasPrefix(path, "/v1/") || strings.HasPrefix(path, "/admin/") || path == "/health" {
			c.JSON(404, gin.H{"error": "not found"})
			return
		}

		// 尝试提供静态文件
		if len(path) > 1 {
			if f, err := distFS.Open(path[1:]); err == nil {
				_ = f.Close()
				fileServer.ServeHTTP(c.Writer, c.Request)
				return
			}
		}

		// SPA fallback: 返回 index.html
		c.Request.URL.Path = "/"
		fileServer.ServeHTTP(c.Writer, c.Request)
	})
}
