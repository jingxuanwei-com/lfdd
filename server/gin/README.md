# Gin 模块

Gin 模块负责初始化 Gin 框架、中间件、路由和 web 服务。

# 使用方法

gin.Init() 初始化中间件（Recovery + Logger）
gin.Start() 启动 web 服务（从配置读取 server_ip + server_port）
gin.Restart() 热切换服务（停掉旧服务，从配置重新读取 ip 和端口启动）
gin.GetAddr() 获取当前监听地址（ip:port）


