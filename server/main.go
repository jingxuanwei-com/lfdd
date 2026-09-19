package main

import (
	"hxzl/api"
	"hxzl/gin"
	"hxzl/gininit"
	"hxzl/home"
	"hxzl/motd"
)

func main() {

	motd.Run()

	// config.Run()

	gininit.Run()

	home.Run()

	api.Run()

	gin.Run()
}
