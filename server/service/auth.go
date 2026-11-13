package service

import (
	"errors"
	"server/core/db"
	"server/model"
)

// ================= 登录 =================
type LoginInput struct {
	Account  string `json:"account"`
	Password string `json:"password"`
	IP       string `json:"ip"`
	Device   string `json:"device"`
}
type LoginOutput struct {
	Token string `json:"token"`
}

// Login 登录
func Login(i LoginInput) (o LoginOutput, err error) {
	return o, nil
}

// ================= 注册 =================
type RegisterInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	IP       string `json:"ip"`
	Device   string `json:"device"`
}

type RegisterOutput struct {
	Token string `json:"token"`
}

// Register 注册
func Register(i RegisterInput) (o RegisterOutput, err error) {
	// 检测是否存在相同邮箱的用户
	if err := db.DB.Where("email = ?", i.Email).First(&model.User{}).Error; err == nil {
		return o, errors.New("邮箱已存在")
	}

	return o, nil
}

// ================= 退出登录 =================
type LogoutInput struct {
	Token string `json:"token"`
}

type LogoutOutput struct {
}

// Logout 退出登录
func Logout(in LogoutInput) (o LogoutOutput, err error) {
	return o, nil
}

// ================= 刷新token =================
type RefreshInput struct {
	Token string `json:"token"`
}

type RefreshOutput struct {
	Token string `json:"token"`
}

// Refresh 刷新token
func Refresh(in RefreshInput) (o RefreshOutput, err error) {
	return o, nil
}
