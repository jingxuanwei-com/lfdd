# GORM 模块

GORM 模块负责数据库连接管理，支持 SQLite、MySQL、PostgreSQL。

# 使用方法

gorm.DB 全局数据库实例

gorm.Start() 启动时检查初始化状态，未初始化则跳过连接
gorm.Stop() 关闭数据库连接
gorm.Restart() 强制重连（关闭旧连接 → 重新读取配置 → 连接）
gorm.Reload() 智能重连（配置无变化且已连接则跳过，否则重连）
gorm.Status() 获取数据库状态

# DBStatus 结构体

Type - 数据库类型（sqlite/mysql/postgresql）
Host - 主机地址
Port - 端口
Name - 数据库名
Status - 连接状态（connected/disconnected）