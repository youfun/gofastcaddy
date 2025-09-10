package tls

import (
	"github.com/youfun/fastcaddy/internal/api"
	"github.com/youfun/fastcaddy/internal/config"
	"github.com/youfun/fastcaddy/pkg/types"
)

// 常量定义 - TLS 自动化配置路径
const AutomationPath = "/apps/tls/automation"

// Manager TLS 配置管理器 - 处理 SSL/TLS 相关配置
type Manager struct {
	client        *api.Client
	configManager *config.Manager
}

// NewManager 创建新的 TLS 管理器
func NewManager() *Manager {
	return &Manager{
		client:        api.NewClient(),
		configManager: config.NewManager(),
	}
}

// GetACMEConfig 获取 ACME 配置 - 对应 Python 的 get_acme_config(token) 函数
// 创建用于 Cloudflare DNS 挑战的 ACME 配置
func GetACMEConfig(token string) map[string]interface{} {
	provider := map[string]interface{}{
		"name":      "cloudflare",
		"api_token": token,
	}

	challenges := map[string]interface{}{
		"dns": map[string]interface{}{
			"provider": provider,
		},
	}

	return map[string]interface{}{
		"module":     "acme",
		"challenges": challenges,
	}
}

// AddTLSInternalConfig 添加内部 TLS 配置 - 对应 Python 的 add_tls_internal_config() 函数
// 为本地开发环境配置内部证书颁发者
func (m *Manager) AddTLSInternalConfig() error {
	// 检查自动化路径是否已存在
	if m.client.HasPath(AutomationPath) {
		return nil // 已存在，无需重复配置
	}

	// 创建空的根配置
	if err := m.client.PutConfig(map[string]interface{}{}, "/", "POST"); err != nil {
		return err
	}

	// 初始化自动化路径
	if err := m.configManager.InitPath(AutomationPath, 0); err != nil {
		return err
	}

	// 创建内部证书颁发者策略
	policies := []map[string]interface{}{
		{
			"issuers": []map[string]interface{}{
				{
					"module": "internal",
				},
			},
		},
	}

	// 设置策略配置
	policiesPath := AutomationPath + "/policies"
	return m.client.PutConfig(policies, policiesPath, "POST")
}

// AddACMEConfig 添加 ACME 配置 - 对应 Python 的 add_acme_config(cf_token) 函数  
// 为生产环境配置 ACME 证书颁发者（使用 Cloudflare）
func (m *Manager) AddACMEConfig(cfToken string) error {
	// 检查自动化路径是否已存在
	if m.client.HasPath(AutomationPath) {
		return nil // 已存在，无需重复配置
	}

	// 创建空的根配置
	if err := m.client.PutConfig(map[string]interface{}{}, "/", "POST"); err != nil {
		return err
	}

	// 初始化自动化路径
	if err := m.configManager.InitPath(AutomationPath, 0); err != nil {
		return err
	}

	// 创建 ACME 配置
	acmeConfig := GetACMEConfig(cfToken)
	issuers := []map[string]interface{}{acmeConfig}

	// 创建 ACME 策略
	policies := []map[string]interface{}{
		{
			"issuers": issuers,
		},
	}

	// 设置策略配置
	policiesPath := AutomationPath + "/policies"
	return m.client.PutConfig(policies, policiesPath, "POST")
}

// SetupPKITrust 配置 PKI 证书颁发机构信任 - 对应 Python 的 setup_pki_trust(install_trust) 函数
// 设置是否将内部 CA 证书安装到系统信任存储
func (m *Manager) SetupPKITrust(installTrust *bool) error {
	// 如果 installTrust 为 nil，不进行任何操作
	if installTrust == nil {
		return nil
	}

	// PKI 证书颁发机构路径
	pkiPath := "/apps/pki/certificate_authorities/local"

	// 初始化 PKI 路径，跳过第一级 (apps)
	if err := m.configManager.InitPath(pkiPath, 1); err != nil {
		return err
	}

	// 创建 PKI 配置
	pkiConfig := types.PKIConfig{
		InstallTrust: *installTrust,
	}

	// 设置 PKI 配置
	return m.client.PutConfig(pkiConfig, pkiPath, "POST")
}