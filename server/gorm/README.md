# GORM 模块

GORM 模块负责数据库连接，支持 SQLite、MySQL、PostgreSQL。

# 使用方法

gorm.Start() 启动时检查是否已初始化，未初始化则跳过连接
gorm.Connect() 手动连接数据库（安装完成后调用)
gorm.DB 全局数据库实例