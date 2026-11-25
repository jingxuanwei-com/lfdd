package service

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"server/core/db"
	"server/model"

	"gorm.io/gorm"
)

// ------------------------ 访问控制（管理页维护） ------------------------
// 开启限制后，除 /admin 外所有页面都要求 URL 带有效令牌（?access=<token>）；
// 令牌存 access_tokens 表，可设有效期（不设则永久有效）。

// CfgAccessAllowDirect 是否允许直接访问："1" 允许（默认）/ "0" 需令牌
const CfgAccessAllowDirect = "access_allow_direct"

// AccessTokenItem 令牌列表项（管理页展示）
type AccessTokenItem struct {
	ID        int64      `json:"id,string"`
	Token     string     `json:"token"`
	ExpiresAt *time.Time `json:"expires_at"`
	CreatedAt time.Time  `json:"created_at"`
}

type GetAccessSettingsInput struct {
}
type GetAccessSettingsOutput struct {
	AllowDirect bool              `json:"allow_direct"`
	Tokens      []AccessTokenItem `json:"tokens"`
}

// GetAccessSettings 读取访问控制配置与令牌列表
func GetAccessSettings(_ GetAccessSettingsInput) (o GetAccessSettingsOutput, err error) {
	o.AllowDirect = getCfg(CfgAccessAllowDirect, "1") == "1"
	o.Tokens = []AccessTokenItem{}
	if db.DB == nil {
		return o, nil
	}

	var rows []model.AccessToken
	if err = db.DB.Order("created_at DESC").Find(&rows).Error; err != nil {
		return o, err
	}
	for _, r := range rows {
		o.Tokens = append(o.Tokens, AccessTokenItem{
			ID:        r.ID,
			Token:     r.Token,
			ExpiresAt: r.ExpiresAt,
			CreatedAt: r.CreatedAt,
		})
	}
	return o, nil
}

type SaveAccessSettingsInput struct {
	AllowDirect bool `json:"allow_direct"`
}
type SaveAccessSettingsOutput struct {
}

// SaveAccessSettings 保存「是否允许直接访问」，立即生效
func SaveAccessSettings(i SaveAccessSettingsInput) (o SaveAccessSettingsOutput, err error) {
	if db.DB == nil {
		return o, errors.New("数据库未连接")
	}
	v := "0"
	if i.AllowDirect {
		v = "1"
	}
	if err = setCfg(CfgAccessAllowDirect, v, "允许直接访问（关闭后需带访问令牌）"); err != nil {
		return o, err
	}
	return o, nil
}

type CreateAccessTokenInput struct {
	ExpireHours int `json:"expire_hours"` // 有效小时数，0 表示永久
}
type CreateAccessTokenOutput struct {
	ID        int64      `json:"id,string"`
	Token     string     `json:"token"`
	ExpiresAt *time.Time `json:"expires_at"`
}

// CreateAccessToken 生成一个随机访问令牌（48 字节 → 64 位 base64url），可设置有效小时数
func CreateAccessToken(i CreateAccessTokenInput) (o CreateAccessTokenOutput, err error) {
	if db.DB == nil {
		return o, errors.New("数据库未连接")
	}
	if i.ExpireHours < 0 {
		return o, errors.New("有效期不能为负数")
	}

	var b [48]byte
	if _, err = rand.Read(b[:]); err != nil {
		return o, err
	}
	o.Token = base64.RawURLEncoding.EncodeToString(b[:])

	var exp *time.Time
	if i.ExpireHours > 0 {
		t := time.Now().Add(time.Duration(i.ExpireHours) * time.Hour)
		exp = &t
	}

	row := &model.AccessToken{Token: o.Token, ExpiresAt: exp}
	if err = db.DB.Create(row).Error; err != nil {
		return o, err
	}
	o.ID = row.ID
	o.ExpiresAt = exp
	return o, nil
}

type DeleteAccessTokenInput struct {
	ID int64 `json:"id,string"`
}
type DeleteAccessTokenOutput struct {
}

// DeleteAccessToken 删除令牌（立即失效）
func DeleteAccessToken(i DeleteAccessTokenInput) (o DeleteAccessTokenOutput, err error) {
	if db.DB == nil {
		return o, errors.New("数据库未连接")
	}
	if i.ID == 0 {
		return o, errors.New("令牌 ID 不能为空")
	}
	if err = db.DB.Delete(&model.AccessToken{}, "id = ?", i.ID).Error; err != nil {
		return o, err
	}
	return o, nil
}

type VerifyAccessTokenInput struct {
	Token string `json:"token"`
	// Authorization 原始请求头：携带有效登录会话的用户免令牌访问（游客身份不算）
	Authorization string `json:"-"`
}
type VerifyAccessTokenOutput struct {
	Allowed bool `json:"allowed"`
}

// bearerFromHeader 解析 Authorization: Bearer <token>，没有则返回空串
func bearerFromHeader(h string) string {
	if !strings.HasPrefix(h, "Bearer ") {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
}

// VerifyAccessToken 校验访问资格：允许直接访问时一律放行；
// 已登录用户（有效登录会话，游客不算）免令牌放行；
// 其余情况令牌必须存在且未过期。
func VerifyAccessToken(i VerifyAccessTokenInput) (o VerifyAccessTokenOutput, err error) {
	if getCfg(CfgAccessAllowDirect, "1") == "1" {
		o.Allowed = true
		return o, nil
	}
	if db.DB == nil {
		return o, errors.New("数据库未连接")
	}

	// 登录用户免令牌：解析 Bearer 后校验会话（游客 token 不以会话表校验，故排除）
	if auth := bearerFromHeader(i.Authorization); auth != "" && !strings.HasPrefix(auth, model.GuestTokenPrefix) {
		if _, err = GetSession(GetSessionInput{Token: auth}); err == nil {
			o.Allowed = true
			return o, nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return o, err
		}
	}

	token := strings.TrimSpace(i.Token)
	if token == "" {
		return o, nil
	}

	var row model.AccessToken
	if err = db.DB.First(&row, "token = ?", token).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return o, nil
		}
		return o, err
	}

	o.Allowed = row.ExpiresAt == nil || row.ExpiresAt.After(time.Now())
	return o, nil
}
