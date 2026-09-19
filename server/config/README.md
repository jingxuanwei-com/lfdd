# 数据库表结构

Name - 数据名称
Value - 数据值
User - 数据修改最新用户
CreatedAt - 数据创建时间
UpdatedAt - 数据更新时间

# 配置文件字段


server_ip - 服务ip地址
server_port - 服务器端口
server_install - 是否初始化

db_type - 数据库类型
db_path - 数据库路径
db_host - 数据库主机
db_port - 数据库端口
db_name - 数据库名称
db_user - 数据库用户名
db_password - 数据库密码

# 使用方法

读取配置
Get("Name")获得配置 返回结构体

写入配置
Set("Name", "Value", "User")三个参数不能为空 不存在Name则创建 存在修改
