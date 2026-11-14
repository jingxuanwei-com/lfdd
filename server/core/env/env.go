package env

import (
	"log"

	"github.com/1szt/go-pkg/env"
)

// Start 初始化配置文件并补全缺失的配置项（不覆盖已有值）。
// 配置文件默认路径为 data/.env，可用环境变量覆盖任意项。
func Start() {
	env.Path("data/.env")

	env.Add("LFDD_PORT", "9081", "端口")
	env.Add("LFDD_SERVER_IP", "", "服务器 IP")

	env.Add("LFDD_DEBUG", "false", "调试模式（true 时 gin 输出路由注册等调试日志，false 静默）")

	env.Add("LFDD_LOG_PATH", "data/log/lfdd.log", "日志文件路径（空则使用默认 data/log/lfdd.log）")

	env.Add("LFDD_SERVER_HTTPS", "true", "是否启用 HTTPS（true/false）")
	env.Add("LFDD_SERVER_HTTPS_AUTO", "true", "是否自签名证书（true/false）")
	env.Add("LFDD_SERVER_CERT", "", "HTTPS 证书路径")
	env.Add("LFDD_SERVER_KEY", "", "HTTPS 私钥路径")
	env.Add("LFDD_SERVER_INSTALL", "true", "是否已完成初始化安装（true/false）")

	env.Add("LFDD_DB_TYPE", "sqlite", "数据库类型（sqlite/pgsql/mysql）")
	env.Add("LFDD_DB_PATH", "data/db/lfdd.db", "SQLite 数据库文件路径")
	env.Add("LFDD_DB_HOST", "", "数据库主机地址")
	env.Add("LFDD_DB_PORT", "", "数据库端口")
	env.Add("LFDD_DB_NAME", "", "数据库名")
	env.Add("LFDD_DB_USER", "", "数据库用户名")
	env.Add("LFDD_DB_PASSWORD", "", "数据库密码")

	env.Add("LFDD_UPLOAD_PATH", "data/upload", "图片上传存储目录")

	env.Add("LFDD_AI_BASE_URL", "", "大模型 API 地址（OpenAI 兼容，空则用 https://api.deepseek.com/v1）")
	env.Add("LFDD_AI_API_KEY", "", "大模型 API Key（空则禁用 AI 点单匹配）")
	env.Add("LFDD_AI_MODEL", "", "大模型名称（空则用 deepseek-chat）")

	log.Println("✅ [Env] 配置文件 初始化完成")
}
