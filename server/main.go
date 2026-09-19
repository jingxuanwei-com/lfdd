package main

import (
	"hxzl/api"
	"hxzl/config"
	"hxzl/gin"
	"hxzl/gorm"
	"hxzl/home"
	"hxzl/motd"
)

func main() {
	gin.Init()
	
	motd.Start()

	config.Start()

	gorm.Start()

	home.Start()

	api.Start()

	gin.Start()
}
