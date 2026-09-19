package gin

import (
	"context"
	"hxzl/config"
	"log"
	"net/http"
	"time"

	g "github.com/gin-gonic/gin"
)

var Router = g.New()
var server *http.Server

// Init 初始化中间件（在模块注册路由前调用）
func Init() {
	Router.Use(g.Recovery())
	Router.Use(g.Logger())
}

// Start 启动 web 服务（从配置读取 ip 和端口）
func Start() {
	ip := config.Get("server_ip").Value
	port := config.Get("server_port").Value
	if port == "" {
		port = "9081"
	}
	start(ip + ":" + port)
}

// start 在指定地址启动服务
func start(addr string) {
	server = &http.Server{
		Addr:         addr,
		Handler:      Router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("🌐 [Gin] 访问 http://%s", addr)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("❌ [Gin] 致命错误：%v", err)
	}
}

// Restart 停掉当前服务，从配置重新读取 ip 和端口重启
func Restart() {
	ip := config.Get("server_ip").Value
	port := config.Get("server_port").Value
	if port == "" {
		port = "9081"
	}
	addr := ip + ":" + port

	if server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("❌ [Gin] 停止旧服务失败: %v", err)
		} else {
			log.Printf("🔄 [Gin] 已停止旧端口服务")
		}
	}

	// 在新地址启动（开一个 goroutine 避免阻塞）
	go start(addr)
	log.Printf("✅ [Gin] 服务已切换到 %s", addr)
}

// GetAddr 获取当前监听地址（ip:port）
func GetAddr() string {
	if server != nil {
		return server.Addr
	}
	return ""
}
