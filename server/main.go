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

	motd.Start()

	config.Start()

	gin.Init()

	gorm.Start()

	home.Start()

	api.Start()

	gin.Start()
}
