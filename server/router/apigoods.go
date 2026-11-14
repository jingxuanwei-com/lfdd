package router

import (
	"server/core/web"
	"server/handler"
	"server/middleware"
)

// apigoods 商品接口：侧边菜单 + 商品页面 + 图片直出免登录（OptionalGuard），
// 新建/修改/删除/顺序管理/图片上传须登录（AuthGuard）
func apiGoodsRouter() {
	apiGoodsHandler := &handler.ApiGoodsHandler{}
	group := web.Engine.Group("/api/goods")

	// 商品页（免登录）
	group.GET("/sidebar", middleware.OptionalGuard(), apiGoodsHandler.Sidebar)
	group.GET("/page", middleware.OptionalGuard(), apiGoodsHandler.Page)

	// 商品管理（登录）
	group.POST("/create", middleware.AuthGuard(), apiGoodsHandler.Create)
	group.POST("/update", middleware.AuthGuard(), apiGoodsHandler.Update)
	group.POST("/delete", middleware.AuthGuard(), apiGoodsHandler.Delete)
	group.POST("/sort", middleware.AuthGuard(), apiGoodsHandler.Sort)
	group.POST("/upload", middleware.AuthGuard(), apiGoodsHandler.Upload)

	// 图片直出（免登录）：GET /api/upload/{hash}{ext}（goods.image 存的就是这个路径）
	web.Engine.GET("/api/upload/:name", apiGoodsHandler.Image)
}
