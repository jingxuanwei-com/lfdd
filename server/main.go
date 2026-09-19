package main

import (
	"hxzl/api"
	"hxzl/config"
	"hxzl/gin"
	"hxzl/gininit"
	"hxzl/gorm"
	"hxzl/home"
	"hxzl/motd"
)

func main() {

	motd.Run()

	config.Run()

	gininit.Run()

	gorm.Run()

	home.Run()

	api.Run()

	gin.Run()
}
