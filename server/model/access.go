package model

import "time"

// AccessToken 访问令牌表：开启访问限制后，页面需带上 ?access=<token> 才能访问。
// 令牌由服务端随机生成，可设置有效期；ExpiresAt 为 nil 表示永久有效。
type AccessToken struct {
	// 自增主键，仅系统内部使用
	ID int64 `json:"id,string" gorm:"primaryKey;autoIncrement"`

	// 访问令牌（随机长串），全局唯一
	Token string `json:"token" gorm:"size:128;not null;uniqueIndex"`

	// 过期时间，nil 表示永久有效
	ExpiresAt *time.Time `json:"expires_at"`

	// 创建时间，由 gorm 自动维护
	CreatedAt time.Time `json:"created_at"`
}

// AccessToken 表结构
func (AccessToken) TableName() string {
	return "access_tokens"
}
