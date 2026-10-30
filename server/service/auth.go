// Package service 业务层：业务规则与数据编排，只依赖 model/db，不依赖 gin/HTTP。
// 被 handler 与 middleware 共用的逻辑放这里；仅单模块私有的逻辑留在模块内。
package service

import (
	"errors"
	"hxzl/core/db"
	"hxzl/model"
	"math/rand/v2"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	Username string
	Email    string
	Password string
}

// Register 注册新用户：查重 → 密码哈希 → 写库
func (a *Auth) Register() (*model.User, error) {
	var count int64
	db.DB.Model(&model.User{}).
		Where("username = ? OR email = ?", a.Username, a.Email).
		Count(&count)
	if count > 0 {
		return nil, errors.New("用户名或邮箱已存在")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(a.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		ID:       NewID(),
		Username: a.Username,
		Email:    a.Email,
		Password: string(hash),
	}
	if err := db.DB.Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

// NewID 生成 18 位数字 ID：毫秒时间戳(13位) + 随机数(5位)，不依赖数据库自增。
// 多实例部署时如出现冲突，再升级为雪花算法。
func NewID() int64 {
	return time.Now().UnixMilli()*100_000 + int64(rand.IntN(100_000))
}
