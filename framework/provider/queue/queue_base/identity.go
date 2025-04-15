package queue_base

import "time"

// AuthIdentity 是统一的用户身份接口，供 MQContext 使用
type AuthIdentity interface {
	// 基础信息
	GetTenantKey() string    // 租户标识（唯一标识一个租户）
	GetEmployeeNo() string   // 员工编号
	GetUserId() int64        // 用户 ID
	GetUserName() string     // 用户姓名
	GetPhone() string        // 手机号
	GetPlatformMark() string // 平台标识（如 Web、App、Feishu 等）
	GetPlatformCode() string // 平台类型编码（从 AuthPlatForm 提取）

	// 鉴权状态
	IsAuthenticated() bool // 是否已通过身份认证
	IsTokenExpired() bool  // Token 是否已过期

	// 身份与权限
	GetUserRoles() []string                 // 角色标识列表
	GetCompanyCodes() []string              // 可操作公司代码
	GetOrgPermissions() map[string][]string // 组织维度权限信息

	// 登录信息
	GetLoginTime() int64       // 登录时间（时间戳）
	GetToken() string          // 登录 Token
	GetTokenExpiration() int64 // Token 过期时间
}

type AuthSession struct {
	Token           string       `json:"token"`           // 用户登录的 Token
	TokenExpiration int64        `json:"tokenExpiration"` // Token 过期时间（毫秒时间戳）
	TenantKey       string       `json:"tenantKey"`
	PlatformMark    string       `json:"platformMark"`
	AuthPlatForm    string       `json:"authPlatForm"` // 你可以定义枚举类型
	UserInfo        AuthUserInfo `json:"userInfo"`     // 用户基本信息
	LoginTime       int64        `json:"loginTime"`    // 登录时间（毫秒时间戳）
	AuthIdentity
}

// 实现 queue_base.AuthIdentity 接口
func (a *AuthSession) GetTenantKey() string {
	return a.TenantKey
}

func (a *AuthSession) GetEmployeeNo() string {
	if a.UserInfo.EmployeeNo != "" {
		return a.UserInfo.EmployeeNo
	}
	return ""
}

func (a *AuthSession) GetPlatformMark() string {
	return a.PlatformMark
}

func (a *AuthSession) IsAuthenticated() bool {
	return a.Token != "" && !a.IsExpired()
}

func (a *AuthSession) GetUserId() int64 {
	return a.UserInfo.UserId
}

func (a *AuthSession) GetUserName() string {
	return a.UserInfo.Name
}

func (a *AuthSession) GetPhone() string {
	return a.UserInfo.Phone
}

func (a *AuthSession) GetPlatformCode() string {
	return a.AuthPlatForm // 或者你可以解析成枚举平台代码
}

func (a *AuthSession) IsTokenExpired() bool {
	return a.IsExpired()
}

func (a *AuthSession) GetUserRoles() []string {
	return a.UserInfo.UserIdentitys
}

func (a *AuthSession) GetCompanyCodes() []string {
	return a.UserInfo.CompanyCodes
}

func (a *AuthSession) GetOrgPermissions() map[string][]string {
	return a.UserInfo.IdentitiesWithOrg
}

func (a *AuthSession) GetLoginTime() int64 {
	return a.LoginTime
}

func (a *AuthSession) GetToken() string {
	return a.Token
}

func (a *AuthSession) GetTokenExpiration() int64 {
	return a.TokenExpiration
}

func (a *AuthSession) IsExpired() bool {
	return time.Now().UnixMilli() > a.TokenExpiration
}

type AuthUserInfo struct {
	UserId            int64               `json:"userId"` // 对应 Long
	EmployeeNo        string              `json:"employeeNo"`
	Name              string              `json:"name"`
	Phone             string              `json:"phone"`
	UserIdentitys     []string            `json:"userIdentitys"`
	IdentitiesWithOrg map[string][]string `json:"identitiesWithOrg"` // 组织维度权限
	CompanyCodes      []string            `json:"companyCodes"`
}

type AuthPlatForm struct {
	Platform string `json:"platform"`
}

func NewAuthPlatForm(platform string) AuthPlatForm {
	return AuthPlatForm{
		Platform: platform,
	}
}
