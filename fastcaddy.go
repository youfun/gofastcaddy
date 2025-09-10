package fastcaddy

import (
	"github.com/youfun/fastcaddy/internal/api"
	"github.com/youfun/fastcaddy/internal/config"
	"github.com/youfun/fastcaddy/internal/routes"
	"github.com/youfun/fastcaddy/internal/tls"
	"github.com/youfun/fastcaddy/internal/utils"
)

// FastCaddy 主要客户端 - 提供 Caddy 配置管理的统一接口
// 这是主要的入口点，整合了所有功能模块
type FastCaddy struct {
	API    *api.Client      // API 客户端
	Config *config.Manager  // 配置管理器  
	TLS    *tls.Manager     // TLS 管理器
	Routes *routes.Manager  // 路由管理器
}

// New 创建新的 FastCaddy 客户端实例
func New() *FastCaddy {
	return &FastCaddy{
		API:    api.NewClient(),
		Config: config.NewManager(),
		TLS:    tls.NewManager(),
		Routes: routes.NewManager(),
	}
}

// SetupCaddy 设置 Caddy 基本配置 - 对应 Python 的 setup_caddy 函数
// 这是初始化 Caddy 配置的主要函数，包括 SSL 配置和 HTTP 应用骨架
func (fc *FastCaddy) SetupCaddy(cfToken, serverName string, local bool, installTrust *bool) error {
	// 根据环境设置 TLS 配置
	if local {
		// 本地开发环境：使用内部证书
		if err := fc.TLS.AddTLSInternalConfig(); err != nil {
			return err
		}
	} else {
		// 生产环境：使用 ACME 证书（需要 Cloudflare 令牌）
		if cfToken == "" {
			cfToken = utils.GetCloudflareToken()
		}
		if cfToken != "" {
			if err := fc.TLS.AddACMEConfig(cfToken); err != nil {
				return err
			}
		}
	}

	// 设置 PKI 信任配置
	if err := fc.TLS.SetupPKITrust(installTrust); err != nil {
		return err
	}

	// 初始化路由配置
	if serverName == "" {
		serverName = "srv0" // 默认服务器名
	}
	return fc.Routes.InitRoutes(serverName, 1)
}

// AddReverseProxy 添加反向代理 - 便利方法
// 创建从指定主机到目标 URL 的反向代理路由
func (fc *FastCaddy) AddReverseProxy(fromHost, toURL string) error {
	return fc.Routes.AddReverseProxy(fromHost, toURL)
}

// AddWildcardRoute 添加通配符路由 - 便利方法
// 为指定域名创建通配符子域名路由
func (fc *FastCaddy) AddWildcardRoute(domain string) error {
	return fc.Routes.AddWildcardRoute(domain)
}

// AddSubReverseProxy 添加子域名反向代理 - 便利方法
// 为通配符域名下的特定子域名添加反向代理
func (fc *FastCaddy) AddSubReverseProxy(domain, subdomain string, ports interface{}, host string) error {
	return fc.Routes.AddSubReverseProxyWithPorts(domain, subdomain, ports, host)
}

// DeleteRoute 删除路由 - 便利方法
// 通过路由 ID 删除特定路由
func (fc *FastCaddy) DeleteRoute(id string) error {
	return fc.Routes.DeleteByID(id)
}

// HasID 检查 ID 是否存在 - 便利方法
func (fc *FastCaddy) HasID(id string) bool {
	return fc.API.HasID(id)
}

// HasPath 检查路径是否存在 - 便利方法
func (fc *FastCaddy) HasPath(path string) bool {
	return fc.API.HasPath(path)
}

// GetConfig 获取配置 - 便利方法
func (fc *FastCaddy) GetConfig(path string) (map[string]interface{}, error) {
	return fc.API.GetConfig(path)
}

// PutConfig 设置配置 - 便利方法
func (fc *FastCaddy) PutConfig(data interface{}, path, method string) error {
	return fc.API.PutConfig(data, path, method)
}