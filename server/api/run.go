package api

import (
	web "hxzl/gin"
	"log"

	g "github.com/gin-gonic/gin"
)

func Run() {
	apiRoute := web.Router.Group("/api")
	apiRoute.GET("/", home)
	apiRoute.GET("/menu", GetMenu)

	log.Print("✅ [API] API模块 加载完成！")
}

func home(c *g.Context) {
	c.String(200, "API")
}
