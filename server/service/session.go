package service

import (
	"crypto/rand"
	"encoding/hex"
	"server/core/db"
	"server/model"
	"time"
)

// ------------------------ NewToken 生成一个新的 Token ------------------------
type NewTokenInput struct {
	UserID int64  `json:"user_id,string"`
	IP     string `json:"ip"`
	Device string `json:"device"`
}
type NewTokenOutput struct {
	Token string `json:"token"`
}

// NewToken 生成一个新的 Token 并持久化保存到数据库
func NewToken(i NewTokenInput) (o NewTokenOutput, err error) {
	// 1. 生成 32 字节（256位）的随机二进制数据
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return o, err
	}

	// 2. 转换为 16 进制字符串作为 Token
	tokenStr := hex.EncodeToString(bytes)

	// 3. 设置过期时间为 7 天
	expireAt := time.Now().Add(7 * 24 * time.Hour)

	// 4. 保存会话到数据库
	if err = db.DB.Create(&model.Session{
		Token:     tokenStr,
		UserID:    i.UserID,
		IP:        i.IP,
		Device:    i.Device,
		ExpiresAt: expireAt,
	}).Error; err != nil {
		return o, err
	}

	o.Token = tokenStr
	return o, nil
}

// ------------------------ RefreshToken 刷新 Token ------------------------
type RefreshTokenInput struct {
	Token string `json:"token"`
}
type RefreshTokenOutput struct {
	Token string `json:"token"`
}

func RefreshToken(i RefreshTokenInput) (o RefreshTokenOutput, err error) {
	// TODO: 实现滑动过期或生成新 Token 替换旧 Token 的逻辑
	return o, nil
}

// ------------------------ RevokeToken 吊销 Token ------------------------
type RevokeTokenInput struct {
	Token string `json:"token"`
}
type RevokeTokenOutput struct {
	// 保持结构体配对规范，留空以备未来扩展
}

func RevokeToken(i RevokeTokenInput) (o RevokeTokenOutput, err error) {
	// 对应登出/踢出：从数据库中软删除或物理删除该 Session
	if err = db.DB.Where("token = ?", i.Token).Delete(&model.Session{}).Error; err != nil {
		return o, err
	}
	return o, nil
}

// ------------------------ GetSessionByToken 查询会话 ------------------------

type GetSessionByTokenInput struct {
	Token string `json:"token"`
}
type GetSessionByTokenOutput struct {
	Session struct {
		Token string `json:"token"`
		UserID int64 `json:"user_id,string"`
		IP string `json:"ip"`
		Device string `json:"device"`
	} `json:"session"`
}

func GetSessionByToken(i GetSessionByTokenInput) (o GetSessionByTokenOutput, err error) {
	if err = db.DB.Where("token = ?", i.Token).First(&o.Session).Error; err != nil {
		return o, err
	}
	return o, nil
}

// ------------------------ GetSessionByUserID 查询用户的会话列表 ------------------------
type GetSessionByUserIDInput struct {
	UserID int64 `json:"user_id,string"`
}
type GetSessionByUserIDOutput struct {
	Sessions []struct {
		Token string `json:"token"`
		UserID int64 `json:"user_id,string"`
		IP string `json:"ip"`
		Device string `json:"device"`
	} `json:"sessions"`
}

func GetSessionByUserID(i GetSessionByUserIDInput) (o GetSessionByUserIDOutput, err error) {
	if err = db.DB.Where("user_id = ?", i.UserID).Find(&o.Sessions).Error; err != nil {
		return o, err
	}
	return o, nil
}
