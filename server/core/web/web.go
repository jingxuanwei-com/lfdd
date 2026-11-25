// Package web web 框架初始化模块：持有全局 Engine 并注册全局中间件，必须在模块注册路由前调用。
package web

import (
	"log"
	"server/middleware"

	"github.com/1szt/go-pkg/env"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

// Engine 全局路由引擎，各业务模块通过它注册自己的路由
var Engine *gin.Engine

// Start 注册全局中间件（Logger + Recovery + 初始化检查）
// 必须在所有业务模块注册路由之前调用（gin 中间件对之后注册的路由才生效）
func Start() {
	// 非调试模式切换 gin 为 release，屏蔽路由注册等调试日志刷屏
	if env.Get("LFDD_DEBUG") != "true" {
		gin.SetMode(gin.ReleaseMode)
	}
	Engine = gin.New()

	Engine.Use(gin.Logger())
	Engine.Use(gin.Recovery())
	// 跨域要放在鉴权之前：预检 OPTIONS 请求不带身份，需直接放行
	Engine.Use(middleware.Cors())
	// 响应压缩：JSON 文本压缩后传输；图片/视频本身已压缩，直出路径排除
	Engine.Use(gzip.Gzip(gzip.DefaultCompression, gzip.WithExcludedPaths([]string{"/api/upload/"})))
	Engine.Use(middleware.InitGuard())
	// 访问令牌拦截：开启访问限制后，未授权请求一律返回伪装的 404
	Engine.Use(middleware.AccessGuard())
	log.Println("✅ [Web]核心初始化 加载完成！")
}
